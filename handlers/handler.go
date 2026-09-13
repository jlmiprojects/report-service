package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"blueassetgroup.com/reports-service/model"
	"blueassetgroup.com/reports-service/repository"
	utils "blueassetgroup.com/reports-service/shared"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Handler struct {
	nc            *nats.Conn
	config        *utils.Config
	service       micro.Service
	logger        *slog.Logger
	dataStore     *repository.MongoRepository
	reportHandler *ReportHandler
}

func NewHandler(version string, nc *nats.Conn, config *utils.Config, r *repository.MongoRepository, reportHandler *ReportHandler) (*Handler, error) {

	var err error

	handler := new(Handler)

	handler.nc = nc
	handler.config = config
	handler.dataStore = r
	handler.reportHandler = reportHandler

	// Setup ROOT logger for devices
	handler.logger = slog.With("name", "reportshandler")

	handler.service, err = micro.AddService(nc, micro.Config{
		Name:        "Reportservice",
		Version:     version,
		Description: "Reports Management Service",
	})

	if err != nil {
		slog.Error("Failed to add Service", "error", err)
	}

	err = handler.service.AddEndpoint("FindAll", micro.HandlerFunc(handler.FindAll), micro.WithEndpointSubject(utils.REPORTS_FIND_ALL))

	if err != nil {
		return nil, err
	}

	err = handler.service.AddEndpoint("ScheduleAdd", micro.HandlerFunc(handler.ScheduleAdd), micro.WithEndpointSubject(utils.REPORTS_SCHEDULE_ADD))

	if err != nil {
		return nil, err
	}

	err = handler.service.AddEndpoint("ScheduleFind", micro.HandlerFunc(handler.ScheduleFind), micro.WithEndpointSubject(utils.REPORTS_SCHEDULE_FIND))
	if err != nil {
		return nil, err
	}

	err = handler.service.AddEndpoint("ParamOptions", micro.HandlerFunc(handler.ParamOptions), micro.WithEndpointSubject(utils.REPORTS_PARAM_OPTIONS))
	if err != nil {
		return nil, err
	}

	return handler, nil

}

func (handler Handler) FindAll(req micro.Request) {

	logger := handler.logger.With("name", "FindAll")

	err := func() error {

		request := new(model.FindAllReportsRequest)

		err := json.Unmarshal(req.Data(), request)

		if err != nil {
			logger.Error("Failed to unmarshal req", "error", err)
			return req.RespondJSON(utils.MakeResult(http.StatusBadRequest, "Failed to unmarshal request", err))
		}

		result, err := handler.dataStore.FindAll(request.Management)

		if err != nil {
			logger.Error("Failed to findall", "error", err)
			return req.RespondJSON(utils.MakeResult(http.StatusBadRequest, "Failed to find all", err))
		}

		return req.RespondJSON(model.FindAllReportsResult{Reports: result})

	}()

	if err != nil {
		logger.Error("Failed to reply", "error", err)
	}

}

func (handler Handler) ScheduleAdd(req micro.Request) {

	logger := handler.logger.With("name", "ScheduleAdd")

	err := func() error {

		logger.Info("New Request Recevied")

		r := new(model.ReportSchedule)
		err := r.FromJSON(req.Data())
		if err != nil {
			return req.RespondJSON(utils.Result{
				Message: "Failed to process payload :" + err.Error(),
				Error:   err.Error()})
		}

		if e := r.Validate(); e != nil {
			return req.RespondJSON(utils.Result{
				StatusCode: http.StatusBadRequest,
				Message:    "Validation Failed",
				Errors:     e})
		}

		id, err := handler.dataStore.AddReportSchedule(r)

		if err != nil {
			return req.RespondJSON(utils.Result{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to add schedule ",
				Error:      err.Error()})
		}

		return req.RespondJSON(model.AddReportScheduleResponse{
			Result: utils.MakeResult(200, "Added", nil),
			ID:     id,
		})
	}()

	if err != nil {
		logger.Error("Failed to reply", "error", err)
	}

}

func (handler Handler) ScheduleFind(req micro.Request) {

	logger := handler.logger.With("name", "ScheduleFind")

	err := func() error {

		r := new(model.FindReportScheduleRequest)

		err := r.FromJSON(req.Data())
		if err != nil {
			logger.Error("Failed to parse", "error", err)
			return req.RespondJSON(utils.Result{
				Message: "Failed to process payload :" + err.Error(),
				Error:   err.Error()})
		}

		logger.Info("New Request Recevied", "request", r)

		if e := r.Validate(); e != nil {
			logger.Error("Failed to validate", "error", err)
			return req.RespondJSON(utils.Result{
				StatusCode: http.StatusBadRequest,
				Message:    "Validation Failed",
				Errors:     e})
		}

		schedule, err := handler.dataStore.FindReportSchedule(r.ProfileId, r.ReportId)

		if err != nil {
			logger.Error("Failed to add", "error", err)
			return req.RespondJSON(utils.Result{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to find report Schedule ",
				Error:      err.Error()})
		}

		return req.RespondJSON(model.FindReportScheduleResponse{
			Result:   utils.MakeResult(200, "OK ", nil),
			Schedule: schedule,
		})
	}()

	if err != nil {
		logger.Error("Failed to reply", "error", err)
	}

}
