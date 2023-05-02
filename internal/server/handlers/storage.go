package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/labstack/echo"
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

	var v interface{}
	var err error

	v, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		v, err = strconv.ParseFloat(value, 64)
		if err != nil {
			log.Println(storage.ErrWrongMetricValue.Error())
			return e.NoContent(http.StatusBadRequest)
		}
	}

	err = c.storage.Update(mtype, mname, v)
	if err == storage.ErrWrongMetricName {
		log.Println(err.Error())
		return e.NoContent(http.StatusNotFound)
	}
	if err == storage.ErrWrongMetricType || err == storage.ErrWrongMetricValue {
		log.Println(err.Error())
		return e.NoContent(http.StatusBadRequest)
	}
	if err != nil {
		log.Printf("something went wrong: %s\n", err.Error())
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

	page := "<table><tr><th>Metric name</th><th>Value</th></tr>%s</table>"
	var rows string
	for key, v := range values {
		rows += fmt.Sprintf("<tr><td>%s</td><td>%v</td></tr>", key, v)
	}
	page = fmt.Sprintf(page, rows)

	return e.HTML(http.StatusOK, page)
}
