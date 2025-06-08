package service

import (
	"context"
	"fmt"
	"time"

	"github.com/LewisJAllan/application-helper/zaphelper"
)

type RespondRequest struct {
	OriginalMessage string
}

type RespondResponse struct {
	ResponseMessage string
}

func (s *Service) Respond(ctx context.Context, request RespondRequest) (RespondResponse, error) {
	zaphelper.Info(ctx, "starting response")
	s.concurrencyRunner.Run(func() {
		time.Sleep(1 * time.Second)
		fmt.Print("hello")
	})
	return RespondResponse{
		ResponseMessage: "Hello World",
	}, nil
}
