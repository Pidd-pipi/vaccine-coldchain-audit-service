package store

import (
	"errors"
	"example.com/vaccine-coldchain-audit-service/domain"
	"sync"
)

var ErrNotFound = errors.New("batch not found")

type Store struct {
	mu    sync.RWMutex
	items []domain.Batch
}

func New() *Store {
	return &Store{items: []domain.Batch{{ID: "vb-401", Vaccine: "流感疫苗", Clinic: "东城接种门诊", Status: "cold", TemperatureC: 4.2, ReceivedAt: "2026-08-20T08:30:00Z"}, {ID: "vb-402", Vaccine: "儿童联合疫苗", Clinic: "南湖社区门诊", Status: "quarantine", TemperatureC: 9.8, ReceivedAt: "2026-08-20T09:10:00Z"}}}
}
func (s *Store) List() []domain.Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Batch, len(s.items))
	copy(out, s.items)
	return out
}
func (s *Store) UpdateStatus(id, v string) (domain.Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Status = v
			return s.items[i], nil
		}
	}
	return domain.Batch{}, ErrNotFound
}
