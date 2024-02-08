package service

import "context"

type RespondRequest struct {
	OriginalMessage string
}

type RespondResponse struct {
	ResponseMessage string
}

func (s *Service) Respond(ctx context.Context, request RespondRequest) (RespondResponse, error) {
	return RespondResponse{}, nil
}
