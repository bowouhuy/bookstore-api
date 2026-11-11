package usecases

import "context"

// Liveness is a function to check liveness
func (u *healthUsecase) Liveness(ctx context.Context) (string, error) {
	if err := u.healthRepo.GetLiveness(ctx); err != nil {
		return "", err
	}

	return "Service is healthy", nil
}
