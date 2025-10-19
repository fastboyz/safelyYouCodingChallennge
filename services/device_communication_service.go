package services

import (
	"encoding/csv"
	"os"
	"safelyYouCodingChallenge/api"
	"safelyYouCodingChallenge/models"
	"time"

	"github.com/rs/zerolog/log"
)

type DeviceCommunicationService interface {
	ReadCSV(path string) error
	RecordHeartbeat(deviceID string, request api.HeartbeatRequest)
	RecordStat(deviceID string, request api.UploadStatsRequest)
	GetDeviceStats(deviceID string) api.GetDeviceStatsResponse
}

type deviceCommunicationService struct {
	devices map[string]*models.Device
}

func NewDeviceCommunicationService() DeviceCommunicationService {
	return &deviceCommunicationService{
		devices: make(map[string]*models.Device),
	}
}

func (d *deviceCommunicationService) ReadCSV(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	reader := csv.NewReader(f)
	data, err := reader.ReadAll()
	if err != nil {
		return err
	}
	d.devices = createDevices(data)

	return nil
}

func (d *deviceCommunicationService) RecordHeartbeat(deviceID string, request api.HeartbeatRequest) {
	device, exists := d.devices[deviceID]
	if !exists {
		log.Warn().Str("device", deviceID).Msg("device not found - Adding it to the records")
		device = &models.Device{
			DeviceID: deviceID,
		}
		d.devices[deviceID] = device
	}
	heartbeat := &models.Heartbeat{
		SentAt: request.SentAt,
	}
	device.Heartbeats = append(device.Heartbeats, heartbeat)

	device.HeartbeatCount++
	if device.FirstHeartbeat.IsZero() || request.SentAt.Before(device.FirstHeartbeat) {
		device.FirstHeartbeat = request.SentAt
	}
	if device.LastHeartbeat.IsZero() || request.SentAt.After(device.LastHeartbeat) {
		device.LastHeartbeat = request.SentAt
	}
}

func (d *deviceCommunicationService) RecordStat(deviceID string, request api.UploadStatsRequest) {
	device, exists := d.devices[deviceID]
	if !exists {
		log.Warn().Str("device", deviceID).Msg("device not found - Adding it to the records")
		device = &models.Device{
			DeviceID: deviceID,
		}
		d.devices[deviceID] = device
	}
	stat := &models.Stat{
		SentAt:     request.SentAt,
		UploadTime: request.UploadTime,
	}
	device.Stats = append(device.Stats, stat)
	device.StatCount++
	device.SumUploadTime += request.UploadTime
}

func (d *deviceCommunicationService) GetDeviceStats(deviceID string) api.GetDeviceStatsResponse {
	device, exists := d.devices[deviceID]
	if !exists {
		log.Warn().Str("device", deviceID).Msg("device not found - Adding it to the records")
		return api.GetDeviceStatsResponse{
			AvgUploadTime: "0s",
			Uptime:        0,
		}
	}

	avgUploadTime := 0
	if len(device.Stats) > 0 {
		avgUploadTime = device.SumUploadTime / device.StatCount
	}

	uptime := 0.0
	if device.HeartbeatCount > 0 {
		minutesBetween := device.LastHeartbeat.Sub(device.FirstHeartbeat).Minutes()
		if minutesBetween > 0 {
			uptime = (float64(device.HeartbeatCount) / minutesBetween) * 100
		}
	}

	return api.GetDeviceStatsResponse{
		AvgUploadTime: (time.Duration(avgUploadTime) * time.Nanosecond).String(),
		Uptime:        uptime,
	}
}

func createDevices(data [][]string) map[string]*models.Device {
	devices := make(map[string]*models.Device)
	if data == nil || len(data) == 0 {
		return devices
	}
	for _, line := range data[1:] {
		device := &models.Device{
			DeviceID: line[0],
		}
		log.Info().Str("device", device.DeviceID).Msg("Adding device from CSV")
		devices[device.DeviceID] = device
	}
	return devices
}
