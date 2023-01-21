package dep

import (
    "github.com/pkg/errors"
    "net/http"
)

const (
    ActivationLockPath = "device/activationlock"
)

type ActivationLockRequest struct {
    Device string `json:"device"`
    EscrowKey string `json:"escrow_key"`
    LostMessage string `json:"lost_message"`
}

type ActivationLockResponse struct {
    SerialNumber string `json:"serial_number"`
    ResponseStatus string `json:"response_status"`
}

func (client *Client) LockDevice(opts ActivationLockRequest) (*ActivationLockResponse, error) {
    var res ActivationLockResponse

    req, err := client.CreateRequest(http.MethodPost, ActivationLockPath, opts)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create LockDevice request")
    }

    err = client.do(req, &res)
    return &res, errors.Wrap(err, "Error running LockDevice request")
}