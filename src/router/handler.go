package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mister-mr-matrix/turnify/src/config"
)

const (
	CloudflareTurnAPIPattern = "https://rtc.live.cloudflare.com/v1/turn/keys/%s/credentials/generate"
)

type CloudflareRequest struct {
	TTL int `json:"ttl"`
}

type CloudflareResponse struct {
	IceServers struct {
		URLs       []string `json:"urls"`
		Username   string   `json:"username"`
		Credential string   `json:"credential"`
	} `json:"iceServers"`
}

type MatrixTurnResponse struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	TTL      int      `json:"ttl"`
	URIs     []string `json:"uris"`
}

func handleRequest(cfg config.Config, w http.ResponseWriter, r *http.Request) {
	slog.Debug("Handling request", "url_path", r.URL.Path)

	if !strings.HasSuffix(r.URL.Path, "/voip/turnServer") {
		slog.Error("Probably wrong reverse proxy settings", "url_path", r.URL.Path)
		http.Error(w, "Bad Gateway: Probably wrong reverse proxy settings, since you ended up on Turnify", http.StatusBadGateway)
		return
	}

	target := fmt.Sprintf("%s%s", cfg.MatrixHomeserverUrl, r.URL.Path)
	slog.Debug("Created target url for Matrix homeserver", "target", target)

	proxyReq, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create a new request for target %s: %s", target, err))
	}

	slog.Debug("Copying all headers")
	for key, values := range r.Header {
		if strings.ToUpper(key) == "HOST" {
			continue
		}

		if len(values) == 0 {
			panic(fmt.Sprintf("(r.Header) Empty values array in headers for key %s", key))
		}

		for _, v := range values {
			proxyReq.Header.Add(key, v)
		}
	}

	// TODO: Is it needed?
	// slog.Debug("Copying all query params")
	// for key, values := range r.URL.Query() {
	// 	if len(values) == 0 {
	// 		panic(fmt.Sprintf("(r.URL.Query()) Empty values array in query parameters for key %s", key))
	// 	}

	// 	for _, v := range values {
	// 		// TODO: Modifying a copy
	// 		proxyReq.URL.Query().Add(key, v)
	// 	}
	// }

	slog.Debug("Proxying request")
	client := http.Client{Timeout: 10 * time.Second}
	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		slog.Error("Couldn't reach the Matrix homeserver", "err", err, "target", target)
		http.Error(w, "Couldn't reach the Matrix homeserver", http.StatusBadGateway)
		return
	}
	slog.Debug("Got response from Matrix homeserver", "status_code", proxyResp.StatusCode)

	if proxyResp.StatusCode != http.StatusOK {
		slog.Debug("Copying response headers for proxyResp")
		for key, values := range proxyResp.Header {
			if strings.ToUpper(key) == "TRANSFER-ENCODING" {
				continue
			}

			if len(values) == 0 {
				panic(fmt.Sprintf("(proxyResp > proxyResp.Header) Empty values array in headers for key %s", key))
			}

			for _, v := range values {
				w.Header().Add(key, v)
			}
		}

		slog.Debug("Copying response status code")
		w.WriteHeader(proxyResp.StatusCode)

		slog.Debug("Copying response body")
		_, err := io.Copy(w, proxyResp.Body)
		if err != nil {
			slog.Error("Failed to copy proxy response body", "err", err)
			http.Error(w, "Failed to copy proxy response body", http.StatusInternalServerError)
			return
		}

		return
	}

	cfTarget := fmt.Sprintf(CloudflareTurnAPIPattern, cfg.CFTurnTokenID)
	slog.Debug("Created target url for Cloudflare API", "target", cfTarget)

	cfBody, err := json.Marshal(CloudflareRequest{cfg.TurnCredentialTTLSeconds})
	if err != nil {
		panic(fmt.Sprintf("Couldn't marshal body for CF request: %s", err))
	}

	cfReq, err := http.NewRequest(http.MethodPost, cfTarget, bytes.NewReader(cfBody))
	if err != nil {
		panic(fmt.Sprintf("Couldn't create a new request for CF target %s: %s", target, err))
	}

	cfReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.CFTurnApiToken))
	cfReq.Header.Set("Content-Type", "application/json")

	slog.Debug("Sending API request to Cloudflare")
	cfClient := http.Client{Timeout: 10 * time.Second}
	cfResp, err := cfClient.Do(cfReq)
	if err != nil {
		slog.Error("Couldn't reach Cloudflare API", "err", err, "target", target)
		http.Error(w, "Couldn't reach Cloudflare API", http.StatusBadGateway)
		return
	}
	defer cfResp.Body.Close()

	cfRespBody, err := io.ReadAll(cfResp.Body)
	if err != nil {
		slog.Error("Couldn't read all body of response", "err", err, "target", target)
		http.Error(w, "Couldn't read all body of response", http.StatusInternalServerError)
		return
	}

	var cfRespJSON CloudflareResponse
	if err := json.Unmarshal(cfRespBody, &cfRespJSON); err != nil {
		slog.Error("Couldn't unmarshal the response body to JSON", "err", err, "body", string(cfRespBody))
		http.Error(w, "Couldn't unmarshal the response body to JSON", http.StatusInternalServerError)
		return
	}

	turnifyResp, err := json.Marshal(MatrixTurnResponse{
		cfRespJSON.IceServers.Username,
		cfRespJSON.IceServers.Credential,
		cfg.TurnCredentialTTLSeconds,
		cfRespJSON.IceServers.URLs,
	})
	if err != nil {
		panic(fmt.Sprintf("Couldn't marshal body for final response %s: %s", string(turnifyResp), err))
	}

	slog.Debug("Copying response headers for Turnify response")
	for key, values := range proxyResp.Header {
		if strings.ToUpper(key) == "CONTENT-LENGTH" || strings.ToUpper(key) ==  "CONTENT-ENCODING" || strings.ToUpper(key) == "TRANSFER-ENCODING" {
			continue
		}

		if len(values) == 0 {
			panic(fmt.Sprintf("(turnifyResp > proxyResp.Header) Empty values array in headers for key %s", key))
		}

		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	slog.Debug("Setting last headers for final response")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(turnifyResp)))

	slog.Debug("Writing status code 200 OK and body for the final response")
	w.WriteHeader(http.StatusOK)
	w.Write(turnifyResp)

	slog.Debug("Done handling the request")
}
