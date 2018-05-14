package common


// Status is the status of the health check.
type Status int

const (
	// OK is the status for a successful health check.
	OK Status = iota
	// KO is the status for an unsuccessful health check.
	KO
	// Degraded is the status for a degraded service, e.g. the service still works, but the metrics DB is KO.
	Degraded
	// Deactivated is the status for a service that is deactivated, e.g. we can disable error tracking, instrumenting, tracing,...
	Deactivated
)

func (s Status) String() string {
	var names = []string{"OK", "KO", "Degraded", "Deactivated"}

	if s < OK || s > Deactivated {
		return "Unknown"
	}

	return names[s]
}

// Report contains the result of one health test.
type Report struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

// err return the string error that will be in the health report
func err(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}