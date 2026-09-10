package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"blueassetgroup.com/reports-service/model"
	utils "blueassetgroup.com/reports-service/shared"
	"github.com/nats-io/nats.go/micro"
)

// lookupRequestTimeout is the default TTL for a lookup's mongo/nats call when
// its DataActions.TTL is unset. Short relative to Generic's 120s default for
// a full report run — a lookup blocks rendering the run dialog.
const lookupRequestTimeout = 30 * time.Second

// ResolveLookup renders a "lookup" parameter's DataAction against query (the
// same {{.field}} templating Generic's mongo/nats data actions use) and runs
// it through the same engine, then maps the resulting rows to Option via
// LabelField/ValueField.
func (rh ReportHandler) ResolveLookup(report *model.Report, lm *model.LookupMetadata, query map[string]any) ([]model.Option, error) {
	action := lm.DataActions

	if err := rh.renderActionRequest(report, &action, query); err != nil {
		return nil, fmt.Errorf("failed to render lookup request: %w", err)
	}

	var rows []any

	switch action.Type {
	case model.DATA_ACTION_MONGO:
		response, err := rh.conns.RunMongoDataAction(&action)
		if err != nil {
			return nil, fmt.Errorf("failed to run mongo lookup: %w", err)
		}
		rows, _ = response["data"].([]any)
	default: // NATS request/reply (type "nats" or unset)
		ttl := lookupRequestTimeout
		if action.TTL != "" {
			if d, err := time.ParseDuration(action.TTL); err == nil {
				ttl = d
			}
		}
		response, err := rh.serviceCall.Generic(action.Action, ttl, action.Request)
		if err != nil {
			return nil, fmt.Errorf("failed to run nats lookup: %w", err)
		}
		if m, ok := response.(map[string]any); ok {
			if data, ok := m["data"].([]any); ok {
				rows = data
			} else {
				rows = []any{m}
			}
		} else if a, ok := response.([]any); ok {
			rows = a
		}
	}

	options := make([]model.Option, 0, len(rows))
	for _, r := range rows {
		row, ok := r.(map[string]any)
		if !ok {
			continue
		}
		value, ok := row[lm.ValueField]
		if !ok {
			continue
		}
		name := row[lm.LabelField]
		options = append(options, model.Option{
			Name:  fmt.Sprintf("%v", name),
			Value: fmt.Sprintf("%v", value),
		})
	}

	return options, nil
}

// ParamOptions serves the REPORTS_PARAM_OPTIONS NATS endpoint: it resolves
// one report parameter's "lookup" options for the broker-portal run dialog to
// render before the user ever sees the form.
func (handler Handler) ParamOptions(req micro.Request) {

	logger := handler.logger.With("name", "ParamOptions")

	err := func() error {

		request := new(model.ParamOptionsRequest)

		if err := json.Unmarshal(req.Data(), request); err != nil {
			logger.Error("Failed to unmarshal req", "error", err)
			return req.RespondJSON(utils.MakeResult(http.StatusBadRequest, "Failed to unmarshal request", err))
		}

		report, err := handler.dataStore.FindByName(request.Report)
		if err != nil {
			logger.Error("Failed to find report", "error", err, "name", request.Report)
			return req.RespondJSON(utils.MakeResult(http.StatusNotFound, "Failed to find report", err))
		}

		var param *model.ReportParams
		for i := range report.Parameters {
			if report.Parameters[i].Name == request.Parameter {
				param = &report.Parameters[i]
				break
			}
		}

		if param == nil || param.Type != model.PARAM_TYPE_LOOKUP {
			err := fmt.Errorf("parameter %q is not a lookup parameter on report %q", request.Parameter, request.Report)
			logger.Error("Invalid lookup parameter", "error", err)
			return req.RespondJSON(utils.MakeResult(http.StatusBadRequest, "Invalid lookup parameter", err))
		}

		lm, err := model.ParseLookupMetadata(param.Metadata)
		if err != nil {
			logger.Error("Failed to parse lookup metadata", "error", err)
			return req.RespondJSON(utils.MakeResult(http.StatusBadRequest, "Failed to parse lookup metadata", err))
		}

		options, err := handler.reportHandler.ResolveLookup(report, lm, request.Query)
		if err != nil {
			logger.Error("Failed to resolve lookup", "error", err)
			return req.RespondJSON(utils.MakeResult(http.StatusInternalServerError, "Failed to resolve lookup", err))
		}

		return req.RespondJSON(model.ParamOptionsResult{
			Result:  utils.MakeResult(http.StatusOK, "OK", nil),
			Options: options,
			Display: DefaultIfZero(lm.Display, "select"),
		})

	}()

	if err != nil {
		logger.Error("Failed to reply", "error", err)
	}
}
