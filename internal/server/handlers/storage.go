package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type StorageController struct {
	storage storage.Storage
}

func newStorageHandlers(router *echo.Router, s storage.Storage) error {
	if router == nil {
		return errors.New("router is nil")
	}
	if s == nil {
		return errors.New("storage is nil")
	}

	c := &StorageController{storage: s}

	router.Add(http.MethodPost, "/update/:mtype/:mname/:value", c.updateHandler)
	router.Add(http.MethodGet, "/value/:mtype/:mname", c.getValueHandler)
	router.Add(http.MethodGet, "/", c.getAllHandler)

	return nil
}

func (c *StorageController) updateHandler(e echo.Context) error {
	mtype := e.Param("mtype")
	mname := e.Param("mname")
	value := e.Param("value")

	var (
		counterVal *int64
		gaugeVal   *float64
		err        error
	)

	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		// HTTP-интерфейс не знает про связь типа метрики и типа данных его значения.
		// Он знает только то, что нам должно прийти значение int64 или float64.
		// И так как метрика gauge (float64) может прийти целым значением и преобразоваться в int64,
		// то пишем полученное значение и в counterVal, и в gaugeVal.
		counterVal = new(int64)
		gaugeVal = new(float64)
		*counterVal = v
		*gaugeVal = float64(v)
	} else {
		gaugeVal = new(float64)
		*gaugeVal, err = strconv.ParseFloat(value, 64)
		if err != nil {
			log.Info(storage.ErrWrongMetricValue.Error())
			return e.NoContent(http.StatusBadRequest)
		}
	}

	err = c.storage.Update(&metric.Metrics{
		ID:    mname,
		MType: mtype,
		Delta: counterVal,
		Value: gaugeVal,
	})
	if err == storage.ErrWrongMetricName {
		log.Info(err.Error())
		return e.NoContent(http.StatusNotFound)
	}
	if err == storage.ErrWrongMetricType || err == storage.ErrWrongMetricValue {
		log.Info(err.Error())
		return e.NoContent(http.StatusBadRequest)
	}
	if err != nil {
		log.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	return e.NoContent(http.StatusOK)
}

func (c *StorageController) getValueHandler(e echo.Context) error {
	mtype := e.Param("mtype")
	mname := e.Param("mname")

	v, ok := c.storage.GetValue(mtype, mname)
	if !ok {
		return e.NoContent(http.StatusNotFound)
	}

	return e.String(http.StatusOK, fmt.Sprintf("%v", v))
}

func (c *StorageController) getAllHandler(e echo.Context) error {
	values := c.storage.GetAll()

	page := "<table><tr><th>Metric name</th><th>Metric type</th><th>Value</th></tr>%s</table>"
	var rows string
	for _, v := range values {
		var val interface{}
		if v.Delta != nil {
			val = *v.Delta
		} else {
			val = *v.Value
		}
		rows += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%v</td></tr>", v.ID, v.MType, val)
	}
	page = fmt.Sprintf(page, rows)

	return e.HTML(http.StatusOK, page)
}
