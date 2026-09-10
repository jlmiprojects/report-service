package handlers

import (
	"fmt"
	"time"

	"blueassetgroup.com/reports-service/model"
	"blueassetgroup.com/reports-service/reportapi"
)

// ResolveLookup runs a "lookup" parameter's script (same reportapi.Context a
// report Script gets) and returns the options it resolves.
func (rh ReportHandler) ResolveLookup(report *model.Report, lm *model.LookupMetadata, query map[string]any) ([]model.Option, error) {
	flatQuery := make(map[string]string, len(query))
	flatQueryAll := make(map[string][]string, len(query))
	for k, v := range query {
		switch val := v.(type) {
		case []any:
			all := make([]string, len(val))
			for i, e := range val {
				all[i] = fmt.Sprintf("%v", e)
			}
			flatQueryAll[k] = all
			if len(all) > 0 {
				flatQuery[k] = all[len(all)-1]
			}
		default:
			s := fmt.Sprintf("%v", v)
			flatQuery[k] = s
			flatQueryAll[k] = []string{s}
		}
	}

	ctx := reportapi.Context{Query: flatQuery, QueryAll: flatQueryAll, ReportName: report.Name, Now: time.Now()}

	opts, err := rh.scripts.RunOptions(lm.Script, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve lookup: %w", err)
	}

	out := make([]model.Option, len(opts))
	for i, o := range opts {
		out[i] = model.Option{Name: o.Name, Value: o.Value}
	}
	return out, nil
}

/*

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
*/
