package glytos

import "context"

// SipTrunksService manages BYO SIP trunks: connect a carrier directly, with no
// third party in between.
//
// A trunk registers with the carrier using credentials it issued you. Numbers
// are attached to a registered trunk through PhoneNumbersService.ImportNumber
// with SipTrunkUUID set.
type SipTrunksService struct{ client *Client }

// SipTrunkCreateParams are the fields for SipTrunksService.Create. Username and
// Password are required; give a Preset and the server, port and transport are
// filled in from it, otherwise set them yourself.
type SipTrunkCreateParams struct {
	// Username is the account or line id the carrier issued (required).
	Username string
	// Password is stored encrypted and is never returned (required).
	Password string
	// Name is a label for the trunk.
	Name string
	// Preset names a known carrier, so the connection details are filled in.
	Preset string
	// SipServer is the carrier host, for example "sip.example.com".
	SipServer string
	// SipPort defaults to the transport's usual port when left at zero.
	SipPort int
	// Transport is udp, tcp or tls.
	Transport string
	// FromDomain overrides the domain in outbound From headers.
	FromDomain string
	// OutboundProxy routes signalling through a different host.
	OutboundProxy string
	// CallerID is the number presented on outbound calls.
	CallerID string
	// CarrierNetworks are the source networks calls may arrive from.
	CarrierNetworks []string
	// UseRegistration is true when the carrier expects the trunk to register
	// rather than accept calls from a fixed address. Zero-valued fields are not
	// sent, so use Extra to send false explicitly.
	UseRegistration bool
	// ExpiresSeconds is the registration lifetime.
	ExpiresSeconds int
	// Extra carries any field not spelled out above, and overrides the ones that
	// are, so a new server-side option needs no SDK release.
	Extra map[string]any
}

// Presets returns carriers whose connection settings are already known, so only
// the login has to be supplied.
func (s *SipTrunksService) Presets(ctx context.Context) ([]SipPreset, error) {
	var out []SipPreset
	err := s.client.do(ctx, "GET", "/telephony/sip-trunks/presets", nil, nil, &out)
	return out, err
}

// List returns your trunks and their registration state.
func (s *SipTrunksService) List(ctx context.Context) ([]SipTrunk, error) {
	var out []SipTrunk
	err := s.client.do(ctx, "GET", "/telephony/sip-trunks", nil, nil, &out)
	return out, err
}

// Create registers a trunk with its carrier.
func (s *SipTrunksService) Create(ctx context.Context, params SipTrunkCreateParams) (*SipTrunk, error) {
	body := map[string]any{"username": params.Username, "password": params.Password}
	if params.Name != "" {
		body["name"] = params.Name
	}
	if params.Preset != "" {
		body["preset"] = params.Preset
	}
	if params.SipServer != "" {
		body["sip_server"] = params.SipServer
	}
	if params.SipPort != 0 {
		body["sip_port"] = params.SipPort
	}
	if params.Transport != "" {
		body["transport"] = params.Transport
	}
	if params.FromDomain != "" {
		body["from_domain"] = params.FromDomain
	}
	if params.OutboundProxy != "" {
		body["outbound_proxy"] = params.OutboundProxy
	}
	if params.CallerID != "" {
		body["caller_id"] = params.CallerID
	}
	if params.CarrierNetworks != nil {
		body["carrier_networks"] = params.CarrierNetworks
	}
	if params.UseRegistration {
		body["use_registration"] = true
	}
	if params.ExpiresSeconds != 0 {
		body["expires_seconds"] = params.ExpiresSeconds
	}
	for k, v := range params.Extra {
		body[k] = v
	}
	var out SipTrunk
	err := s.client.do(ctx, "POST", "/telephony/sip-trunks", body, nil, &out)
	return &out, err
}

// Update changes a trunk. Only the fields present in body are changed.
func (s *SipTrunksService) Update(ctx context.Context, trunkUUID string, body map[string]any) (*SipTrunk, error) {
	var out SipTrunk
	err := s.client.do(ctx, "PATCH", "/telephony/sip-trunks/"+esc(trunkUUID), body, nil, &out)
	return &out, err
}

// Delete removes a trunk. Numbers attached to it stop receiving calls.
func (s *SipTrunksService) Delete(ctx context.Context, trunkUUID string) error {
	return s.client.do(ctx, "DELETE", "/telephony/sip-trunks/"+esc(trunkUUID), nil, nil, nil)
}

// Test re-checks the trunk against its carrier now, rather than waiting for the
// next reconcile. Reachable in the result separates "the carrier refused these
// credentials" from "nobody answered"; only the first is worth changing the
// password over.
func (s *SipTrunksService) Test(ctx context.Context, trunkUUID string) (*SipTrunkTest, error) {
	var out SipTrunkTest
	err := s.client.do(ctx, "POST", "/telephony/sip-trunks/"+esc(trunkUUID)+"/test", nil, nil, &out)
	return &out, err
}
