package xpath

type Logger interface {
	Printf(string, ...interface{})
}

type nopLoggerImpl struct{}

func newNopLogger() *nopLoggerImpl {
	return &nopLoggerImpl{}
}

func (*nopLoggerImpl) Printf(format string, v ...interface{}) {
	//noop
}
