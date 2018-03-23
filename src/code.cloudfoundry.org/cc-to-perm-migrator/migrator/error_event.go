package migrator

type ErrorEvent struct {
	Cause      error
	GUID       string
	EntityType string
}

func (e *ErrorEvent) Error() string {
	return e.Cause.Error()
}
