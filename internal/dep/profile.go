package dep

import (
    "github.com/pkg/errors"
    "net/http"
)

const (
    GetProfilePath = "profile"
    CreateProfilePath = "profile"
    AssignProfilePath = "profile/devices"
    RemoveProfilePath = "profile/devices"
)

type Profile struct {
    ProfileName string `json:"profile_name"`
    ServerUrl string `json:"url"`
    AllowPairing bool `json:"allow_pairing"`
    Supervised bool `json:"is_supervised"`
    MultiUser bool `json:"is_multi_user"`
    Mandatory bool `json:"is_mandatory,omitempty"`
    AwaitDeviceConfiguration bool `json:"await_device_configuration"`
    Removable bool `json:"is_mdm_removable"`
    AutoAdvanceSetup bool `json:"auto_advance_setup,omitempty"`
    SupportPhoneNumber string `json:"support_phone_number,omitempty"`
    SupportEmailAddress string `json:"support_email_address,omitempty"`
    OrganizationMagic []string `json:"organization_magic,omitempty"`
    SupervisingHostCerts []string `json:"supervising_host_certs,omitempty"`
    SkipSetupItems []string `json:"skip_setup_items,omitempty"`
    Department string `json:"department,omitempty"`
    Devices []string `json:"devices"`

    // tvOS Only
    Language string `json:"language,omitempty"`
    Region string `json:"region,omitempty"`
}

type ProfileResponse struct {
    ProfileUUID string `json:"profile_uuid"`
    Devices map[string]string
}

func (client *Client) GetProfile(uuid string) (*Profile, error) {
    var res Profile
    req, err := client.CreateRequest(http.MethodGet, GetProfilePath, nil)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create GetProfile request")
    }

    profileQuery := req.URL.Query()
    profileQuery.Add("profile_uuid", uuid)
    req.URL.RawQuery = profileQuery.Encode()

    err = client.do(req, &res)
    return &res, errors.Wrap(err, "Error running GetProfile request")
}

func (client *Client) CreateProfile(opts ...Profile) (*ProfileResponse, error) {
    var res ProfileResponse
    req, err := client.CreateRequest(http.MethodPost, CreateProfilePath, opts)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create GetProfile request")
    }

    err = client.do(req, &res)
    return &res, errors.Wrap(err, "Error running CreateProfile request")
}

func (client *Client) AssignProfile(uuid string, devices ...string) (*ProfileResponse, error)  {
    var res ProfileResponse
    var body = struct {
        Devices []string `json:"devices"`
    }{
        Devices: devices,
    }

    req, err := client.CreateRequest(http.MethodPost, AssignProfilePath, body)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create AssignProfile request")
    }

    err = client.do(req, &res)
    return &res, errors.Wrap(err, "Error running AssignProfile request")
}

func (client *Client) RemoveProfile(devices ...string) (map[string]string, error) {
    var res struct {
        Devices map[string]string `json:"devices"`
    }
    var body = struct {
        Devices []string `json:"devices"`
    }{
        Devices: devices,
    }

    req, err := client.CreateRequest(http.MethodDelete, RemoveProfilePath, body)
    if err != nil {
        return nil, errors.Wrap(err, "Failed to create RemoveProfile request")
    }

    err = client.do(req, &res)
    return res.Devices, errors.Wrap(err, "Error running RemoveProfile request")
}