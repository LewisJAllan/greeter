package service

import "context"

type RespondRequest struct {
	originalMessage string
}

type RespondResponse struct {
	responseMessage string
}

func (s *Service) Respond(ctx context.Context, request RespondRequest) (RespondRequest, error) {
	return RespondRequest{}, nil
}
