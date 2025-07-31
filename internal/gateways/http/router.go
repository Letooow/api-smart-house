package http

import (
	"errors"
	"homework/internal/domain"
	"homework/internal/repository/event/postgres"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/strfmt"
	model "homework/api/generated"
)

const contentTypeErrorMessage = "Content-Type must be 'application/json'"

func setupRouter(r *gin.Engine, uc UseCases, ws *WebSocketHandler) {
	r.HandleMethodNotAllowed = true
	r.Use(checkMediaTypeMiddleWare)

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	setEvents(r, uc)
	setSensors(r, uc)
	setUsers(r, uc)

	r.GET("/sensors/:id/events", getLastEventBySensor(uc, ws))
}

func setEvents(r *gin.Engine, uc UseCases) {
	r.POST("/events", receiveEventToSensor(uc))
	r.OPTIONS("/events", setHeaderOptions("POST,OPTIONS"))
}

func setSensors(r *gin.Engine, uc UseCases) {
	r.POST("/sensors", postSensor(uc))
	r.GET("/sensors", getSensors(uc, false))
	r.HEAD("/sensors", getSensors(uc, true))
	r.OPTIONS("/sensors", setHeaderOptions("GET,POST,OPTIONS,HEAD"))

	r.GET("/sensors/:id", getSensorByID(uc, false))
	r.HEAD("/sensors/:id", getSensorByID(uc, true))
	r.GET("/sensors/:id/history", getSensorHistory(uc))
	r.OPTIONS("/sensors/:id", setHeaderOptions("GET,OPTIONS,HEAD"))
}

func setUsers(r *gin.Engine, uc UseCases) {
	r.POST("/users", postUser(uc))
	r.OPTIONS("/users", setHeaderOptions("GET,POST,OPTIONS,HEAD"))

	r.GET("/users/:id/sensors", getUserSensors(uc, false))
	r.HEAD("/users/:id/sensors", getUserSensors(uc, true))
	r.POST("/users/:id/sensors", postSensorToUser(uc))
	r.OPTIONS("users/:id/sensors", setHeaderOptions("POST,GET,OPTIONS,HEAD"))
}

func getLastEventBySensor(uc UseCases, ws *WebSocketHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if _, err = uc.Sensor.GetSensorByID(c, id); err != nil {
			setError(c, http.StatusNotFound, err.Error())
			return
		}

		if err = ws.Handle(c, id); err != nil {
			setError(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.Status(http.StatusSwitchingProtocols)
	}
}

func checkMediaTypeMiddleWare(c *gin.Context) {
	connectionHeader := c.GetHeader("Connection")
	upgrade := c.GetHeader("Upgrade")

	if (strings.Contains(connectionHeader, "Upgrade") && upgrade == "websocket") || c.FullPath() == "" {
		c.Next()
		return
	}

	switch c.Request.Method {
	case "GET", "HEAD":
		if c.GetHeader("Accept") != "application/json" {
			c.AbortWithStatusJSON(http.StatusNotAcceptable, errors.New(contentTypeErrorMessage))
			return
		}
	case "POST":
		if c.ContentType() != "application/json" {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, errors.New(contentTypeErrorMessage))
			return
		}
	}
	c.Next()
}

func postSensor(uc UseCases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sensorToCreate model.SensorToCreate
		if err := c.ShouldBindJSON(&sensorToCreate); err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}

		if err := sensorToCreate.Validate(strfmt.Default); err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}

		sensor := domain.Sensor{
			Description:  *sensorToCreate.Description,
			IsActive:     *sensorToCreate.IsActive,
			SerialNumber: *sensorToCreate.SerialNumber,
			Type:         domain.SensorType(*sensorToCreate.Type),
		}

		registeredSensor, err := uc.Sensor.RegisterSensor(c.Request.Context(), &sensor)
		if err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, registeredSensor)
	}
}

func receiveEventToSensor(uc UseCases) func(c *gin.Context) {
	return func(c *gin.Context) {
		var event model.SensorEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}

		if err := event.Validate(strfmt.Default); err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}

		domainEvent := domain.Event{
			Payload:            *event.Payload,
			SensorSerialNumber: *event.SensorSerialNumber,
		}

		err := uc.Event.ReceiveEvent(c.Request.Context(), &domainEvent)
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
		c.JSON(http.StatusCreated, &domainEvent)
	}
}

func postUser(uc UseCases) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userToCreate model.UserToCreate

		if err := c.ShouldBindJSON(&userToCreate); err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}

		err := userToCreate.Validate(strfmt.Default)
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
		user := domain.User{
			Name: *userToCreate.Name,
		}

		newUser, err := uc.User.RegisterUser(c.Request.Context(), &user)
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
		c.JSON(http.StatusOK, newUser)
	}
}

func postSensorToUser(uc UseCases) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
		_, err = uc.User.GetUserSensors(c.Request.Context(), int64(id))
		if err != nil {
			setError(c, http.StatusNotFound, err.Error())
			return
		}

		var sensor model.SensorToUserBinding
		if err = c.ShouldBindJSON(&sensor); err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}
		if err = sensor.Validate(strfmt.Default); err != nil {
			setError(c, http.StatusUnprocessableEntity, "Sensor ID must be greater than zero")
			return
		}

		err = uc.User.AttachSensorToUser(c.Request.Context(), int64(id), *sensor.SensorID)
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
		c.JSON(http.StatusCreated, sensor)
	}
}

func getSensorHistory(uc UseCases) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}
		startDate := c.DefaultQuery("start_date", "")
		endDate := c.DefaultQuery("end_date", "")
		if startDate == "" || endDate == "" {
			setError(c, http.StatusBadRequest, "Start date or end date required")
			return
		}
		_, err = uc.Sensor.GetSensorByID(c.Request.Context(), id)
		if err != nil {
			setError(c, http.StatusNotFound, err.Error())
			return
		}

		layout := "2006-01-02T15:04:05"
		start, err := time.Parse(layout, startDate)
		if err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}
		end, err := time.Parse(layout, endDate)
		if err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}
		events, err := uc.Event.GetEventsBySensorIDWithDate(c.Request.Context(), id, start, end)
		if errors.Is(err, postgres.ErrEventNotFound) {
			c.JSON(http.StatusOK, gin.H{})
			return
		} else if err != nil {
			setError(c, http.StatusBadRequest, err.Error())
			return
		}

		c.JSON(http.StatusOK, events)
	}
}

func getUserSensors(uc UseCases, head bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sensors := getErrorOfUserOfSensorID(c, uc, c.GetHeader("Accept"))
		if sensors == nil {
			return
		}
		if head {
			c.Header("Content-Length", strconv.Itoa(len(sensors)))
		}
		c.JSON(http.StatusOK, sensors)
	}
}

func setHeaderOptions(methods string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Allow", methods)
		c.Status(http.StatusNoContent)
	}
}

func getSensorByID(uc UseCases, head bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			setError(c, http.StatusUnprocessableEntity, err.Error())
			return
		}

		sensor, err := uc.Sensor.GetSensorByID(c.Request.Context(), int64(id))
		if err != nil {
			setError(c, http.StatusNotFound, err.Error())
			return
		}
		if head {
			c.Header("Content-Length", strconv.Itoa(len(sensor.SerialNumber)))
		}
		c.JSON(http.StatusOK, sensor)
	}
}

func getSensors(uc UseCases, head bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sensors, _ := uc.Sensor.GetSensors(c.Request.Context())
		if head {
			c.Header("Content-Length", strconv.Itoa(len(sensors)))
		}
		c.JSON(http.StatusOK, sensors)
	}
}

func getErrorOfUserOfSensorID(c *gin.Context, uc UseCases, header string) []domain.Sensor {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		setError(c, http.StatusUnprocessableEntity, err.Error())
		return nil
	}
	if header != "application/json" {
		setError(c, http.StatusNotAcceptable, contentTypeErrorMessage)
		return nil
	}

	sensors, err := uc.User.GetUserSensors(c.Request.Context(), int64(id))
	if err != nil {
		setError(c, http.StatusNotFound, err.Error())
		return nil
	}
	return sensors
}

func setError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, model.Error{Reason: &message})
}
