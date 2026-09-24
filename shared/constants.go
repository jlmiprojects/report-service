package utils

// NATS subjects this service answers (handlers/handler.go). The BluHive app
// requests REPORTS_FIND_ALL and REPORTS_PARAM_OPTIONS
// (bluhive/ui/services/report.go), so keep the two in step.
const (
	REPORTS_FIND_ALL      = "blueasset.reports.findall"
	REPORTS_SCHEDULE_ADD  = "blueasset.reports.schedule.add"
	REPORTS_SCHEDULE_FIND = "blueasset.reports.schedule.find"
	REPORTS_PARAM_OPTIONS = "blueasset.reports.param.options"
)
