package service

type AsynchronousRunner interface {
	Run(f func())
}

type Service struct {
	concurrencyRunner AsynchronousRunner
}

func NewService(concurrencyRunner AsynchronousRunner) Service {
	return Service{
		concurrencyRunner: concurrencyRunner,
	}
}
