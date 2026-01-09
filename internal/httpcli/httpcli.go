package httpcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type config struct {
	Authorization string
}

type Option func(*config)

func WithAuthorization(token string) Option {
	return func(c *config) {
		c.Authorization = token
	}
}

type Client struct {
	cli *http.Client
	cfg config
}

func New(opts ...Option) *Client {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Client{
		cli: &http.Client{},
		cfg: cfg,
	}
}

func (c *Client) GetJSON(u string, dst interface{}) error {
	req, err := http.NewRequest(http.MethodGet, u, http.NoBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	if c.cfg.Authorization != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Authorization)
	}

	resp, err := c.cli.Do(req) //nolint:noctx // ok
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response: " + resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("failed to read the response: " + err.Error())
	}

	err = json.Unmarshal(bytes, dst)
	if err != nil {
		return errors.New("failed to unmarshal the response: " + err.Error())
	}

	return nil
}
