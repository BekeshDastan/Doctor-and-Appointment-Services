package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrDoctorServiceUnavailable = errors.New("doctor service is currently unavailable")

type RESTDoctorClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewRESTDoctorClient(baseURL string) *RESTDoctorClient {
	return &RESTDoctorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *RESTDoctorClient) CheckDoctorExists(ctx context.Context, doctorID string) (bool, error) {
	url := fmt.Sprintf("%s/doctors/%s", c.baseURL, doctorID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, ErrDoctorServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("%w: unexpected status code %d", ErrDoctorServiceUnavailable, resp.StatusCode)
}
