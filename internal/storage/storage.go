package storage

import "errors"

var (
	ErrWrongMetricType  = errors.New("wrong metric type")
	ErrWrongMetricName  = errors.New("wrong metric name")
	ErrWrongMetricValue = errors.New("wrong metric value")
)

type Storage interface {
	GetAll() map[string]interface{}
	GetValue(mtype string, mname string) (interface{}, bool)
	Update(mtype, mname string, value interface{}) error
}
