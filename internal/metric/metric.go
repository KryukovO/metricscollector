package metric

import "errors"

const (
	GaugeMetric   = "gauge"
	CounterMetric = "counter"
)

var ErrWrongMetricValue = errors.New("wrong metric value")

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMetrics(mname string, val interface{}) (*Metrics, error) {
	switch v := val.(type) {
	case int64:
		return &Metrics{
			ID:    mname,
			MType: CounterMetric,
			Delta: &v,
		}, nil
	case float64:
		return &Metrics{
			ID:    mname,
			MType: GaugeMetric,
			Value: &v,
		}, nil
	default:
		return nil, ErrWrongMetricValue
	}
}
