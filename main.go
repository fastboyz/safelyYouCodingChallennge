package main

import (
	"net/http"
	"safelyYouCodingChallenge/api"
	"safelyYouCodingChallenge/services"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msgf("Starting server")

	deviceCommunicationService := services.NewDeviceCommunicationService()
	err := deviceCommunicationService.ReadCSV("./resources/devices.csv")
	if err != nil {
		log.Err(err).Msg("Failed to read CSV")
		return
	}

	server := api.NewServer(deviceCommunicationService)

	serverMux := http.NewServeMux()

	handler := api.HandlerFromMuxWithBaseURL(server, serverMux, "/api/v1")

	httServer := &http.Server{
		Handler: handler,
		Addr:    ":6733",
	}
	err = httServer.ListenAndServe()
	log.Err(err).Msg("")
}
