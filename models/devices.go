package models

import "time"

type Heartbeat struct {
	SentAt time.Time
}

type Stat struct {
	SentAt     time.Time
	UploadTime int
}
type Device struct {
	DeviceID       string `json:"device_id"`
	Heartbeats     []*Heartbeat
	Stats          []*Stat
	SumUploadTime  int
	StatCount      int
	HeartbeatCount int
	FirstHeartbeat time.Time
	LastHeartbeat  time.Time
}

func (device *Device) GetId() string {
	if device != nil {
		return device.DeviceID
	}
	return ""
}

func (device *Device) GetStats() []*Stat {
	if device != nil {
		return device.Stats
	}
	return nil
}

func (device *Device) GetHeartbeats() []*Heartbeat {
	if device != nil {
		return device.Heartbeats
	}
	return nil
}

func (heartbeat *Heartbeat) GetSentAt() time.Time {
	if heartbeat != nil {
		return heartbeat.SentAt
	}
	return time.Unix(0, 0)
}

func (stas *Stat) GetUploadTime() int {
	if stas != nil {
		return stas.UploadTime
	}
	return 0
}

func (stas *Stat) GetSentAt() time.Time {
	if stas != nil {
		return stas.SentAt
	}
	return time.Unix(0, 0)
}
