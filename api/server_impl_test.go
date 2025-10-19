package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockDCS implements deviceCommunicationService for tests.
type mockDCS struct {
	recordHeartbeatFn func(deviceID string, request HeartbeatRequest)
	recordStatFn      func(deviceID string, request UploadStatsRequest)
	getDeviceStatsFn  func(deviceID string) GetDeviceStatsResponse
}

func (m *mockDCS) RecordHeartbeat(deviceID string, request HeartbeatRequest) {
	if m.recordHeartbeatFn != nil {
		m.recordHeartbeatFn(deviceID, request)
	}
}

func (m *mockDCS) RecordStat(deviceID string, request UploadStatsRequest) {
	if m.recordStatFn != nil {
		m.recordStatFn(deviceID, request)
	}
}

func (m *mockDCS) GetDeviceStats(deviceID string) GetDeviceStatsResponse {
	if m.getDeviceStatsFn != nil {
		return m.getDeviceStatsFn(deviceID)
	}
	return GetDeviceStatsResponse{}
}

func TestNewServer_concreteType(t *testing.T) {
	s := NewServer(&mockDCS{})
	if s == nil {
		t.Fatalf("NewServer returned nil")
	}
	// ensure concrete type is serverImpl
	if _, ok := s.(*serverImpl); !ok {
		t.Fatalf("NewServer did not return *serverImpl, got %T", s)
	}
}

func Test_serverImpl_PostDevicesDeviceIdHeartbeat(t *testing.T) {
	tests := []struct {
		name             string
		body             string
		dcs              *mockDCS
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:             "decode error",
			body:             "not-json",
			dcs:              &mockDCS{},
			wantStatus:       400,
			wantBodyContains: "error decoding body",
		},
		{
			name: "success",
			body: "{}",
			dcs:  &mockDCS{},
			// on success the handler writes an empty JSON object
			wantStatus:       204,
			wantBodyContains: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/devices/device-1/heartbeat", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			s := &serverImpl{deviceCommunicationService: tt.dcs}
			s.PostDevicesDeviceIdHeartbeat(rr, req, DeviceIDPathParam("device-1"))

			if rr.Code != tt.wantStatus {
				t.Fatalf("status: got %d want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if !strings.Contains(rr.Body.String(), tt.wantBodyContains) {
				t.Fatalf("body does not contain expected substring. got: %s want substring: %s", rr.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

func Test_serverImpl_PostDevicesDeviceIdStats(t *testing.T) {
	tests := []struct {
		name             string
		body             string
		dcs              *mockDCS
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:             "decode error",
			body:             "not-json",
			dcs:              &mockDCS{},
			wantStatus:       400,
			wantBodyContains: "error decoding body",
		},
		{
			name:             "success",
			body:             "{}",
			dcs:              &mockDCS{},
			wantStatus:       204,
			wantBodyContains: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/devices/device-1/stats", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			s := &serverImpl{deviceCommunicationService: tt.dcs}
			s.PostDevicesDeviceIdStats(rr, req, "device-1")

			if rr.Code != tt.wantStatus {
				t.Fatalf("status: got %d want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if !strings.Contains(rr.Body.String(), tt.wantBodyContains) {
				t.Fatalf("body does not contain expected substring. got: %s want substring: %s", rr.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

func Test_serverImpl_GetDevicesDeviceIdStats(t *testing.T) {
	expected := GetDeviceStatsResponse{}
	// Try to populate expected with something JSON-encodable if fields exist:
	_ = json.Unmarshal([]byte("{}"), &expected)

	dcs := &mockDCS{
		getDeviceStatsFn: func(deviceID string) GetDeviceStatsResponse {
			return expected
		},
	}

	req := httptest.NewRequest("GET", "/devices/device-1/stats", nil)
	rr := httptest.NewRecorder()

	s := &serverImpl{deviceCommunicationService: dcs}
	s.GetDevicesDeviceIdStats(rr, req, "device-1")

	if rr.Code != 200 {
		t.Fatalf("expected status 200 got %d; body: %s", rr.Code, rr.Body.String())
	}

	// Ensure the body is valid JSON matching the returned value (at least decodable)
	var got GetDeviceStatsResponse
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("response body is not valid JSON for GetDeviceStatsResponse: %v; body: %s", err, rr.Body.String())
	}
}
