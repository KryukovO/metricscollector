package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type StorageController struct {
	storage storage.Storage
	l       *log.Logger
}

func setStorageHandlers(router *echo.Router, s storage.Storage, l *log.Logger) error {
	if router == nil {
		return errors.New("router is nil")
	}
	if s == nil {
		return errors.New("storage is nil")
	}

	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	c := &StorageController{storage: s, l: lg}

	router.Add(http.MethodPost, "/update/:mtype/:mname/:value", c.updateHandler)
	router.Add(http.MethodPost, "/update/", c.updateJSONHandler)
	router.Add(http.MethodGet, "/value/:mtype/:mname", c.getValueHandler)
	router.Add(http.MethodPost, "/value/", c.getValueJSONHandler)
	router.Add(http.MethodGet, "/", c.getAllHandler)
	router.Add(http.MethodGet, "/ping", c.pingHandler)

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
		counterVal = &v
		gaugeVal = new(float64)
		*gaugeVal = float64(v)
	} else {
		gaugeVal = new(float64)
		*gaugeVal, err = strconv.ParseFloat(value, 64)
		if err != nil {
			c.l.Info(storage.ErrWrongMetricValue.Error())
			return e.NoContent(http.StatusBadRequest)
		}
	}

	err = c.storage.Update(e.Request().Context(), &metric.Metrics{
		ID:    mname,
		MType: mtype,
		Delta: counterVal,
		Value: gaugeVal,
	})
	if err == storage.ErrWrongMetricName {
		c.l.Info(err.Error())
		return e.NoContent(http.StatusNotFound)
	}
	if err == storage.ErrWrongMetricType || err == storage.ErrWrongMetricValue {
		c.l.Info(err.Error())
		return e.NoContent(http.StatusBadRequest)
	}
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	return e.NoContent(http.StatusOK)
}

func (c *StorageController) updateJSONHandler(e echo.Context) error {
	body, err := io.ReadAll(e.Request().Body)
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	var mtrc metric.Metrics
	err = json.Unmarshal(body, &mtrc)
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	err = c.storage.Update(e.Request().Context(), &mtrc)
	if err == storage.ErrWrongMetricName {
		c.l.Info(err.Error())
		return e.NoContent(http.StatusNotFound)
	}
	if err == storage.ErrWrongMetricType || err == storage.ErrWrongMetricValue {
		c.l.Info(err.Error())
		return e.NoContent(http.StatusBadRequest)
	}
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	return e.JSON(http.StatusOK, &mtrc)
}

func (c *StorageController) getValueHandler(e echo.Context) error {
	mtype := e.Param("mtype")
	mname := e.Param("mname")

	v, err := c.storage.GetValue(e.Request().Context(), mtype, mname)
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}
	if v == nil {
		return e.NoContent(http.StatusNotFound)
	}

	if v.Delta != nil {
		return e.String(http.StatusOK, fmt.Sprintf("%d", *v.Delta))
	}

	return e.String(http.StatusOK, strconv.FormatFloat(*v.Value, 'f', -1, 64))
}

func (c *StorageController) getValueJSONHandler(e echo.Context) error {
	body, err := io.ReadAll(e.Request().Body)
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	var mtrc metric.Metrics
	err = json.Unmarshal(body, &mtrc)
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	v, err := c.storage.GetValue(e.Request().Context(), mtrc.MType, mtrc.ID)
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}
	if v == nil {
		return e.NoContent(http.StatusNotFound)
	}

	return e.JSON(http.StatusOK, v)
}

func (c *StorageController) getAllHandler(e echo.Context) error {
	values, err := c.storage.GetAll(e.Request().Context())
	if err != nil {
		c.l.Infof("something went wrong: %s", err.Error())
		return e.NoContent(http.StatusInternalServerError)
	}

	builder := strings.Builder{}

	builder.WriteString("<table><tr><th>Metric name</th><th>Metric type</th><th>Value</th></tr>")
	for _, v := range values {
		if v.Delta != nil {
			builder.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td></tr>", v.ID, v.MType, *v.Delta))
		} else {
			builder.WriteString(
				fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", v.ID, v.MType, strconv.FormatFloat(*v.Value, 'f', -1, 64)),
			)
		}
	}
	builder.WriteString("</table>")

	return e.HTML(http.StatusOK, builder.String())
}

func (c *StorageController) pingHandler(e echo.Context) error {
	if c.storage.Ping() {
		return e.NoContent(http.StatusOK)
	}
	return e.NoContent(http.StatusInternalServerError)
}
