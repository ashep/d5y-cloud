// Author:  Oleksandr Shepetko
// Email:   a@shepetko.com
// License: MIT

package httpcli

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Client struct {
	cli *http.Client
}

func New() *Client {
	return &Client{
		cli: &http.Client{},
	}
}

func (c *Client) GetJSON(u string, dst interface{}) error {
	resp, err := c.cli.Get(u) //nolint:noctx // ok
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
