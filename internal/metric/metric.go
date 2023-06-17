package metric

import (
	"errors"
)

const (
	GaugeMetric   = "gauge"
	CounterMetric = "counter"
)

var (
	ErrWrongMetricType  = errors.New("wrong metric type")
	ErrWrongMetricName  = errors.New("wrong metric name")
	ErrWrongMetricValue = errors.New("wrong metric value")
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // тип метрики (gauge или counter)
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Создает структуру метрики.
//
// Если параметр mType не заполнен, тип метрики определяется по переданному значению value:
// float64 => gauge; int64 => counter.
func NewMetrics(mName, mType string, value interface{}) (*Metrics, error) {
	if mName == "" {
		return nil, ErrWrongMetricName
	}

	if value == nil {
		return nil, ErrWrongMetricValue
	}

	switch mType {
	case CounterMetric:
		v, ok := value.(int64)
		if !ok {
			return nil, ErrWrongMetricValue
		}

		return &Metrics{
			ID:    mName,
			MType: CounterMetric,
			Delta: &v,
		}, nil
	case GaugeMetric:
		vf, ok := value.(float64)
		if !ok {
			vi, ok := value.(int64)
			if !ok {
				return nil, ErrWrongMetricValue
			}

			vf = float64(vi)
		}

		return &Metrics{
			ID:    mName,
			MType: GaugeMetric,
			Value: &vf,
		}, nil
	case "":
		return newMetricsByValue(mName, value)
	default:
		return nil, ErrWrongMetricType
	}
}

func newMetricsByValue(mName string, value interface{}) (*Metrics, error) {
	switch v := value.(type) {
	case int64:
		return &Metrics{
			ID:    mName,
			MType: CounterMetric,
			Delta: &v,
		}, nil
	case float64:
		return &Metrics{
			ID:    mName,
			MType: GaugeMetric,
			Value: &v,
		}, nil
	default:
		return nil, ErrWrongMetricValue
	}
}

func (mtrc *Metrics) Validate() error {
	if mtrc.ID == "" {
		return ErrWrongMetricName
	}

	switch mtrc.MType {
	case CounterMetric:
		if mtrc.Delta == nil {
			return ErrWrongMetricValue
		}
	case GaugeMetric:
		if mtrc.Value == nil {
			return ErrWrongMetricValue
		}
	default:
		return ErrWrongMetricType
	}

	return nil
}
