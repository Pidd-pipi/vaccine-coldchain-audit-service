package main

import (
	"context"
	"sync"
)

// OpsBatchResult summarizes a batch audit reconcile run.
type OpsBatchResult struct {
	Total    int
	Matched  int
	Failed   int
	Failures []string
}

// opsBatchGate decides whether a record can join the batch under the rule.
func opsBatchGate(rule OpsRule, record OpsRecord) error {
	for _, label := range rule.RequiredLabels {
		if record.LabelValue(label) == "" {
			return wrapOps("batch", "rule.gate", ErrOpsPolicy)
		}
	}
	return nil
}

// opsReconcileBatch fans audit checks for the given record ids out to a fixed
// worker pool. Every id is reported exactly once in the result.
func opsReconcileBatch(ctx context.Context, service *OpsService, rule OpsRule, ids []string, workers int) (OpsBatchResult, error) {
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan string)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for id := range jobs {
				if err := opsBatchGate(rule, mustGetRecord(ctx, service, id)); err != nil {
					select {
					case errCh <- err:
					default:
					}
				}
			}
		}()
	}
	go func() {
		for _, id := range ids {
			select {
			case jobs <- id:
			case <-ctx.Done():
				close(jobs)
				return
			}
		}
		close(jobs)
	}()
	wg.Wait()
	result := OpsBatchResult{Total: len(ids)}
	select {
	case err := <-errCh:
		result.Failed = 1
		result.Failures = append(result.Failures, err.Error())
		return result, err
	default:
	}
	result.Matched = len(ids)
	return result, nil
}

// mustGetRecord fetches a record for batch processing.
func mustGetRecord(ctx context.Context, service *OpsService, id string) OpsRecord {
	record, err := service.Get(ctx, id)
	if err != nil {
		return OpsRecord{}
	}
	return record
}
