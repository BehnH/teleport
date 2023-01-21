package dep

import (
    "net/http"
    "time"

    "github.com/pkg/errors"
)

const (
    GetDevicesPath = "devices"
    GetAllDevicesPath = "server/devices"
    SyncDevicesPath = "devices/sync"
)

type Device struct {
    SerialNumber string `json:"serial_number"`
    Model string `json:"model"`
    Description string `json:"description"`
    Color string `json:"color"`
    AssetTag string `json:"asset_tag"`
    ProfileStatus string `json:"profile_status"`
    ProfileUUID string `json:"profile_uuid"`
    ProfileAssignTime string `json:"profile_assign_time"`
    ProfileAssignDate string `json:"profile_assign_date"`
    ProfilePushTime string `json:"profile_push_time"`
    DeviceAssignedBy string `json:"device_assigned_by"`
    OS string `json:"os"`
    DeviceFamily string `json:"device_family"`
    // Below fields only exist in Sync requests
    OpType string    `json:"op_type,omitempty"`
    OpDate time.Time `json:"op_date,omitempty"`
    // Below field only exists in Device requests
    ResponseStatus string `json:"response_status,omitempty"`
}


type DeviceResponse struct {
    Devices      []Device  `json:"devices"`
    Cursor       string    `json:"cursor"`
    FetchedUntil time.Time `json:"fetched_until"`
    MoreToFollow bool      `json:"more_to_follow"`
}

type DeviceRequestOpt func(opts *DeviceRequestOpts) error

type DeviceRequestOpts struct {
    Cursor string `json:"cursor,omitempty"`
    Limit int `json:"limit,omitempty"`
}

// Cursor is an optional argument that can be passed on Sync calls as a means of pagination
func Cursor(cursor string) DeviceRequestOpt {
    return func(opts *DeviceRequestOpts) error {
        opts.Cursor = cursor
        return nil
    }
}

// Limit is an optional argument that can be passed on FetchDevices and SyncDevices calls
// to limit the number of results returned
func Limit(limit int) DeviceRequestOpt {
    return func(opts *DeviceRequestOpts) error {
        if limit > 1000 {
            return errors.New("limit cannot not be higher than 1000")
        }
        opts.Limit = limit
        return nil
    }
}

func (client *Client) GetAllDevices(opts ...DeviceRequestOpt) (*DeviceResponse, error) {
    request := &DeviceRequestOpts{}

    for _, option := range opts {
        if err := option(request); err != nil {
            return nil, err
        }
    }

    var res DeviceResponse
    req, err := client.CreateRequest(http.MethodPost, GetAllDevicesPath, request)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create GetAllDevices request")
    }

    err = client.do(req, &res)
    return &res, errors.Wrap(err, "Error running GetAllDevices request")
}

func (client *Client) SyncDevices(cursor string, opts ...DeviceRequestOpt) (*DeviceResponse, error)  {
    request := &DeviceRequestOpts{
        Cursor: cursor,
    }

    for _, option := range opts {
        if err := option(request); err != nil {
            return nil, err
        }
    }

    var res DeviceResponse
    req, err := client.CreateRequest(http.MethodPost, SyncDevicesPath, request)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create SyncDevices request")
    }

    err = client.do(req, &res)
    return &res, errors.Wrap(err, "Error running SyncDevices request")
}

type GetDevicesResponse struct {
    Devices map[string]Device `json:"devices"`
}

func (client *Client) GetDevices(serials ...string) (*GetDevicesResponse, error) {
    request := struct {
        Devices []string `json:"devices"`
    }{
        Devices: serials,
    }

    var res GetDevicesResponse
    req, err := client.CreateRequest(http.MethodPost, GetDevicesPath, request)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create GetDevices request")
    }

    err = client.do(req, &res)
    return &res, errors.Wrap(err, "Error running GetDevices request")
}