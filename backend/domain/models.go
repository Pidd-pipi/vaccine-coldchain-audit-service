package domain

type Batch struct {
	ID           string  `json:"id"`
	Vaccine      string  `json:"vaccine"`
	Clinic       string  `json:"clinic"`
	Status       string  `json:"status"`
	TemperatureC float64 `json:"temperature_c"`
	ReceivedAt   string  `json:"received_at"`
}
type StatusChange struct {
	Status string `json:"status"`
}
