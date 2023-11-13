// Package metric содержит описание метрики.
package metric

import (
	"errors"

	pb "github.com/KryukovO/metricscollector/api/serverpb"
)

var (
	// ErrWrongMetricType возвращается, если метрика содержит некорректный тип.
	ErrWrongMetricType = errors.New("wrong metric type")
	// ErrWrongMetricType возвращается, если метрика содержит некорректное имя.
	ErrWrongMetricName = errors.New("wrong metric name")
	// ErrWrongMetricType возвращается, если тип значения метрики не соответствует её типу.
	ErrWrongMetricValue = errors.New("wrong metric value")
)

// MetricType - тип метрики.
type MetricType string

const (
	GaugeMetric   MetricType = "gauge"
	CounterMetric MetricType = "counter"
)

// MapGRPCToMetricType - маппинг типа метрики в запросе/ответе gRPC в MetricType.
var MapGRPCToMetricType = map[pb.MetricType]MetricType{
	pb.MetricType_COUNTER: CounterMetric,
	pb.MetricType_GAUGE:   GaugeMetric,
}

// MapMetricTypeToGRPC - маппинг MetricType в тип метрики в запросе/ответе gRPC.
var MapMetricTypeToGRPC = map[MetricType]pb.MetricType{
	CounterMetric: pb.MetricType_COUNTER,
	GaugeMetric:   pb.MetricType_GAUGE,
}

// Metrics описывает структуру метрики.
type Metrics struct {
	ID    string     `json:"id"`              // Имя метрики
	MType MetricType `json:"type"`            // Тип метрики (gauge или counter)
	Delta *int64     `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64   `json:"value,omitempty"` // Значение метрики в случае передачи gauge
}

// NewMetrics создает структуру метрики.
//
// Если параметр mType не заполнен, тип метрики определяется по переданному значению value:
// float64 => gauge; int64 => counter.
func NewMetrics(mName string, mType MetricType, value interface{}) (Metrics, error) {
	if mName == "" {
		return Metrics{}, ErrWrongMetricName
	}

	if value == nil {
		return Metrics{}, ErrWrongMetricValue
	}

	switch mType {
	case CounterMetric:
		v, ok := value.(int64)
		if !ok {
			return Metrics{}, ErrWrongMetricValue
		}

		return Metrics{
			ID:    mName,
			MType: CounterMetric,
			Delta: &v,
		}, nil
	case GaugeMetric:
		vf, ok := value.(float64)
		if !ok {
			vi, ok := value.(int64)
			if !ok {
				return Metrics{}, ErrWrongMetricValue
			}

			vf = float64(vi)
		}

		return Metrics{
			ID:    mName,
			MType: GaugeMetric,
			Value: &vf,
		}, nil
	case "":
		return newMetricsByValue(mName, value)
	default:
		return Metrics{}, ErrWrongMetricType
	}
}

func newMetricsByValue(mName string, value interface{}) (Metrics, error) {
	switch v := value.(type) {
	case int64:
		return Metrics{
			ID:    mName,
			MType: CounterMetric,
			Delta: &v,
		}, nil
	case float64:
		return Metrics{
			ID:    mName,
			MType: GaugeMetric,
			Value: &v,
		}, nil
	default:
		return Metrics{}, ErrWrongMetricValue
	}
}

// Validate осуществляет валидацию метрики.
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
