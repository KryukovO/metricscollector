package storage

import "errors"

var (
	ErrWrongMetricType  = errors.New("wrong metric type")
	ErrWrongMetricValue = errors.New("wrong metric value")
)

type Storage interface {
	Update(mtype, mname string, value interface{}) error
}
