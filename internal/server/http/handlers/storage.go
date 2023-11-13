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

var (
	// ErrControllerIsNil возвращается MapStorageHandlers, если контроллер не инициализирован.
	ErrControllerIsNil = errors.New("controller is nil")
	// ErrRouterIsNil возвращается MapStorageHandlers, если маршрутизатор echo не инициализирован.
	ErrRouterIsNil = errors.New("router is nil")
)

// StorageController представляет собой контроллер для хранилища.
type StorageController struct {
	storage storage.Storage
	l       *log.Logger
}

// NewStorageController создаёт новый контроллер хранилища.
func NewStorageController(s storage.Storage, l *log.Logger) (*StorageController, error) {
	if s == nil {
		return nil, ErrStorageIsNil
	}

	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	return &StorageController{storage: s, l: lg}, nil
}

// MapStorageHandlers выполняет маппинг маршрутов и обработчиков в маршрутизатор echo.
func MapStorageHandlers(router *echo.Router, c *StorageController) error {
	if router == nil {
		return ErrRouterIsNil
	}

	if c == nil {
		return ErrControllerIsNil
	}

	router.Add(http.MethodPost, "/update/:mtype/:mname/:value", c.updateHandler)
	router.Add(http.MethodPost, "/update/", c.updateJSONHandler)
	router.Add(http.MethodPost, "/updates/", c.updatesHandler)
	router.Add(http.MethodGet, "/value/:mtype/:mname", c.getValueHandler)
	router.Add(http.MethodPost, "/value/", c.getValueJSONHandler)
	router.Add(http.MethodGet, "/", c.getAllHandler)
	router.Add(http.MethodGet, "/ping", c.pingHandler)

	return nil
}

// updateHandler представляет собой обработчик запроса на обновление единственной метрики.
// Параметры метрики передаются через URL.
func (c *StorageController) updateHandler(e echo.Context) error {
	uuid := e.Get("uuid")

	mType := e.Param("mtype")
	if mType == "" {
		c.l.Debugf("[%s] %s", uuid, metric.ErrWrongMetricType)

		return e.NoContent(http.StatusBadRequest)
	}

	mName := e.Param("mname")

	value := e.Param("value")

	var (
		val interface{}
		err error
	)

	if counterVal, parseIntErr := strconv.ParseInt(value, 10, 64); parseIntErr == nil {
		val = counterVal
	} else {
		gaugeVal, parseFloatErr := strconv.ParseFloat(value, 64)
		if parseFloatErr != nil {
			c.l.Debugf("[%s] %s", uuid, metric.ErrWrongMetricValue.Error())

			return e.NoContent(http.StatusBadRequest)
		}
		val = gaugeVal
	}

	mtrc, err := metric.NewMetrics(mName, metric.MetricType(mType), val)
	if errors.Is(err, metric.ErrWrongMetricName) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusNotFound)
	}

	if errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusBadRequest)
	}

	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	err = c.storage.Update(e.Request().Context(), &mtrc)
	if errors.Is(err, metric.ErrWrongMetricName) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusNotFound)
	}

	if errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusBadRequest)
	}

	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	return e.NoContent(http.StatusOK)
}

// updateJSONHandler представляет собой обработчик запроса на обновление единственной метрики.
// Параметры метрики передаются в формате JSON в теле http-запроса.
func (c *StorageController) updateJSONHandler(e echo.Context) error {
	uuid := e.Get("uuid")

	body, err := io.ReadAll(e.Request().Body)
	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	var mtrc metric.Metrics
	err = json.Unmarshal(body, &mtrc)

	var jsonErr *json.UnmarshalTypeError
	if err != nil {
		if errors.As(err, &jsonErr) {
			c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return e.NoContent(http.StatusBadRequest)
		}

		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	err = c.storage.Update(e.Request().Context(), &mtrc)
	if errors.Is(err, metric.ErrWrongMetricName) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusNotFound)
	}

	if errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusBadRequest)
	}

	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	return e.JSON(http.StatusOK, &mtrc)
}

// updatesHandler представляет собой обработчик запроса на обновление набора метрик.
// Параметры метрик передаются в формате JSON в теле HTTP-запроса.
func (c *StorageController) updatesHandler(e echo.Context) error {
	uuid := e.Get("uuid")

	body, err := io.ReadAll(e.Request().Body)
	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	var mtrcs []metric.Metrics
	err = json.Unmarshal(body, &mtrcs)

	var jsonErr *json.UnmarshalTypeError
	if err != nil {
		if errors.As(err, &jsonErr) {
			c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return e.NoContent(http.StatusBadRequest)
		}

		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	err = c.storage.UpdateMany(e.Request().Context(), mtrcs)
	if errors.Is(err, metric.ErrWrongMetricName) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusNotFound)
	}

	if errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
		c.l.Debugf("[%s] %s", uuid, err.Error())

		return e.NoContent(http.StatusBadRequest)
	}

	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	return e.NoContent(http.StatusOK)
}

// getValueHandler представляет собой обработчик запроса на получение параметров единственной метрики.
// Параметры запрашиваемой метрики передаются через URL.
func (c *StorageController) getValueHandler(e echo.Context) error {
	uuid := e.Get("uuid")

	v, err := c.storage.GetValue(e.Request().Context(), metric.MetricType(e.Param("mtype")), e.Param("mname"))
	if errors.Is(err, metric.ErrWrongMetricType) {
		return e.NoContent(http.StatusNotFound)
	}

	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	if v.ID == "" {
		return e.NoContent(http.StatusNotFound)
	}

	if v.Delta != nil {
		return e.String(http.StatusOK, strconv.FormatInt(*v.Delta, 10))
	}

	return e.String(http.StatusOK, strconv.FormatFloat(*v.Value, 'f', -1, 64))
}

// getValueJSONHandler представляет собой обработчик запроса на получение параметров единственной метрики.
// Параметры запрашиваемой метрики передаются в формате JSON в теле HTTP-запроса.
func (c *StorageController) getValueJSONHandler(e echo.Context) error {
	uuid := e.Get("uuid")

	body, err := io.ReadAll(e.Request().Body)
	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	var mtrc metric.Metrics

	err = json.Unmarshal(body, &mtrc)
	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	v, err := c.storage.GetValue(e.Request().Context(), mtrc.MType, mtrc.ID)
	if errors.Is(err, metric.ErrWrongMetricType) {
		return e.NoContent(http.StatusNotFound)
	}

	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	if v.ID == "" {
		return e.NoContent(http.StatusNotFound)
	}

	return e.JSON(http.StatusOK, v)
}

// getAllHandler представляет собой обработчик запроса списка всех метрик из хранилища.
// Результат возвращается в формате HTML в виде таблицы: Metric name | Metric type | Value.
func (c *StorageController) getAllHandler(e echo.Context) error {
	uuid := e.Get("uuid")

	values, err := c.storage.GetAll(e.Request().Context())
	if err != nil {
		c.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return e.NoContent(http.StatusInternalServerError)
	}

	builder := strings.Builder{}

	builder.WriteString("<table><tr><th>Metric name</th><th>Metric type</th><th>Value</th></tr>")

	for _, v := range values {
		if v.Delta != nil {
			builder.WriteString(
				fmt.Sprintf(
					"<tr><td>%s</td><td>%s</td><td>%d</td></tr>",
					v.ID, v.MType, *v.Delta,
				),
			)
		} else {
			builder.WriteString(
				fmt.Sprintf(
					"<tr><td>%s</td><td>%s</td><td>%s</td></tr>",
					v.ID, v.MType, strconv.FormatFloat(*v.Value, 'f', -1, 64),
				),
			)
		}
	}

	builder.WriteString("</table>")

	return e.HTML(http.StatusOK, builder.String())
}

// pingHandler представляет собой обработчик запроса на проверку доступности хранилища.
func (c *StorageController) pingHandler(e echo.Context) error {
	if c.storage.Ping(e.Request().Context()) {
		return e.NoContent(http.StatusOK)
	}

	return e.NoContent(http.StatusInternalServerError)
}
