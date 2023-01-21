package dep

import (
    "github.com/pkg/errors"
    "net/http"
)

type RequestLimit struct {
    Default int `json:"default"`
    Maximum int `json:"maximum"`
}

type URL struct {
    URI string `json:"uri"`
    HttpMethod string `json:"http_method"`
    Limit RequestLimit `json:"limit"`
}

// Definition for the MDM server's "account" according to Apple's
// MDM server account spec
// https://developer.apple.com/business/documentation/MDM-Protocol-Reference.pdf - Page 102
type Account struct {
    ServerName string `json:"server_name"`
    ServerUUID string `json:"server_uuid"`
    AdminId string `json:"admin_id"`
    FacilitatorId string `json:"facilitator_id"`
    OrgId string `json:"org_id"`
    OrgIdHash string `json:"org_id_hash"`
    OrgName string `json:"org_name"`
    OrgEmail string `json:"org_email"`
    OrgPhone string `json:"org_phone"`
    OrgType string `json:"org_type"`
    OrgVersion string `json:"org_version"`
    URLs []URL `json:"urls"`
}

func (client *Client) GetAccount() (*Account, error) {
    var acct Account
    req, err := client.CreateRequest(http.MethodGet, "account", nil)

    if err != nil {
        return nil, errors.Wrap(err, "Account request creation failed")
    }

    err = client.do(req, acct)
    return &acct, errors.Wrap(err, "Get account request failed")
}