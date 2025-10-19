package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

type deviceCommunicationService interface {
	RecordHeartbeat(deviceID string, request HeartbeatRequest)
	RecordStat(deviceID string, request UploadStatsRequest)
	GetDeviceStats(deviceID string) GetDeviceStatsResponse
}

type serverImpl struct {
	deviceCommunicationService deviceCommunicationService
}

func NewServer(dcs deviceCommunicationService) ServerInterface {
	return &serverImpl{
		deviceCommunicationService: dcs,
	}
}

// PostDevicesDeviceIdHeartbeat (POST /devices/{device_id}/heartbeat)
func (s *serverImpl) PostDevicesDeviceIdHeartbeat(w http.ResponseWriter, r *http.Request, deviceId DeviceIDPathParam) {
	var req HeartbeatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResponse := Error{Msg: "error decoding body"}
		log.Err(err).Msg("error decoding body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errResponse)
		return
	}
	s.deviceCommunicationService.RecordHeartbeat(deviceId, req)
	w.WriteHeader(http.StatusNoContent)
	_ = json.NewEncoder(w).Encode(struct{}{})
}

// PostDevicesDeviceIdStats (POST /devices/{device_id}/stats)
func (s *serverImpl) PostDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId DeviceIDPathParam) {
	var req UploadStatsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResponse := Error{Msg: "error decoding body"}
		log.Err(err).Msg("error decoding body")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errResponse)
		return
	}
	s.deviceCommunicationService.RecordStat(deviceId, req)
	w.WriteHeader(http.StatusNoContent)
	_ = json.NewEncoder(w).Encode(struct{}{})
}

// GetDevicesDeviceIdStats (GET /devices/{device_id}/stats)
func (s *serverImpl) GetDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceId DeviceIDPathParam) {
	res := s.deviceCommunicationService.GetDeviceStats(deviceId)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}
