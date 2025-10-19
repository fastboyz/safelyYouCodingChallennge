package services

import (
	"os"
	"path/filepath"
	"safelyYouCodingChallenge/api"
	"safelyYouCodingChallenge/models"
	"testing"
	"time"
)

func TestNewDeviceCommunicationService_and_type(t *testing.T) {
	svc := NewDeviceCommunicationService()
	if svc == nil {
		t.Fatalf("NewDeviceCommunicationService returned nil")
	}
	if _, ok := svc.(*deviceCommunicationService); !ok {
		t.Fatalf("expected *deviceCommunicationService, got %T", svc)
	}
}

func TestReadCSV(t *testing.T) {
	tmpDir := t.TempDir()

	validCSV := "device_id\ndevice-1\ndevice-2\n"
	validPath := filepath.Join(tmpDir, "valid.csv")
	if err := os.WriteFile(validPath, []byte(validCSV), 0o644); err != nil {
		t.Fatalf("failed to write temp csv: %v", err)
	}

	tests := []struct {
		name          string
		path          string
		expectErr     bool
		expectDevices int
	}{
		{
			name:      "file not found",
			path:      filepath.Join(tmpDir, "no-such-file.csv"),
			expectErr: true,
		},
		{
			name:          "valid csv",
			path:          validPath,
			expectErr:     false,
			expectDevices: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &deviceCommunicationService{}
			err := s.ReadCSV(tt.path)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error for path %s but got nil", tt.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(s.devices) != tt.expectDevices {
				t.Fatalf("expected %d devices, got %d", tt.expectDevices, len(s.devices))
			}
		})
	}
}

func Test_createDevices(t *testing.T) {
	tests := []struct {
		name string
		data [][]string
		want int
	}{
		{
			name: "header only",
			data: [][]string{{"device_id"}},
			want: 0,
		},
		{
			name: "multiple devices",
			data: [][]string{
				{"device_id"},
				{"a"},
				{"b"},
				{"c"},
			},
			want: 3,
		},
		{
			name: "duplicate ids",
			data: [][]string{
				{"device_id"},
				{"x"},
				{"x"},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createDevices(tt.data)
			if len(got) != tt.want {
				t.Fatalf("expected %d devices, got %d", tt.want, len(got))
			}
		})
	}
}

func TestRecordHeartbeat_and_RecordStat_and_GetDeviceStats(t *testing.T) {
	now := time.Now()

	t.Run("RecordHeartbeat creates device when missing and appends", func(t *testing.T) {
		s := &deviceCommunicationService{
			devices: make(map[string]*models.Device),
		}
		// call using the concrete signature expected by the service
		s.RecordHeartbeat("dev-1", api.HeartbeatRequest{SentAt: now})
		// The service method appends a heartbeat; verify device created and heartbeat count 1
		if s.devices == nil {
			t.Fatalf("devices map nil after RecordHeartbeat")
		}
		dev, ok := s.devices["dev-1"]
		if !ok {
			t.Fatalf("device not created")
		}
		if len(dev.Heartbeats) != 1 {
			t.Fatalf("expected 1 heartbeat, got %d", len(dev.Heartbeats))
		}
	})

	t.Run("RecordStat creates device when missing and appends", func(t *testing.T) {
		s := &deviceCommunicationService{
			devices: make(map[string]*models.Device),
		}
		s.RecordStat("dev-2", api.UploadStatsRequest{
			SentAt:     now,
			UploadTime: 1000,
		})
		if s.devices == nil {
			t.Fatalf("devices map nil after RecordStat")
		}
		dev, ok := s.devices["dev-2"]
		if !ok {
			t.Fatalf("device not created")
		}
		if len(dev.Stats) != 1 {
			t.Fatalf("expected 1 stat, got %d", len(dev.Stats))
		}
	})

	t.Run("GetDeviceStats returns zero values for missing device", func(t *testing.T) {
		s := &deviceCommunicationService{
			devices: make(map[string]*models.Device),
		}
		got := s.GetDeviceStats("nope")
		if got.AvgUploadTime != "0s" || got.Uptime != 0 {
			t.Fatalf("expected zero values, got %+v", got)
		}
	})

	t.Run("GetDeviceStats computes average upload and uptime", func(t *testing.T) {
		s := &deviceCommunicationService{
			devices: make(map[string]*models.Device),
		}
		dev := &models.Device{DeviceID: "dev-3"}
		dev.Stats = []*models.Stat{
			{UploadTime: 1000000, SentAt: now},
			{UploadTime: 1000000, SentAt: now.Add(time.Second)},
		}

		dev.StatCount = len(dev.Stats)
		var sum int
		for _, st := range dev.Stats {
			sum += st.UploadTime
		}
		dev.SumUploadTime = sum

		dev.Heartbeats = []*models.Heartbeat{
			{SentAt: now},
			{SentAt: now.Add(2 * time.Minute)},
			{SentAt: now.Add(4 * time.Minute)},
		}
		// Initialize heartbeat metadata used by GetDeviceStats
		dev.HeartbeatCount = len(dev.Heartbeats)
		dev.FirstHeartbeat = dev.Heartbeats[0].SentAt
		dev.LastHeartbeat = dev.Heartbeats[len(dev.Heartbeats)-1].SentAt

		s.devices["dev-3"] = dev

		got := s.GetDeviceStats("dev-3")

		if got.AvgUploadTime != "1ms" {
			t.Fatalf("expected avg upload '1ms', got %q", got.AvgUploadTime)
		}
		if got.Uptime != 75.0 {
			t.Fatalf("expected uptime 75.0, got %v", got.Uptime)
		}
	})

	t.Run("GetDeviceStats with single heartbeat yields uptime 0", func(t *testing.T) {
		s := &deviceCommunicationService{
			devices: make(map[string]*models.Device),
		}
		dev := &models.Device{DeviceID: "dev-4"}
		dev.Heartbeats = []*models.Heartbeat{{SentAt: now}}
		s.devices["dev-4"] = dev

		got := s.GetDeviceStats("dev-4")
		if got.Uptime != 0 {
			t.Fatalf("expected uptime 0 for single heartbeat, got %v", got.Uptime)
		}
	})
}
