package util

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/pquerna/otp"
)

// helpers

func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func HTTPError(w http.ResponseWriter, code int, msg string) {
	http.Error(w, msg, code)
}

func ParseUserIDFromRequest(r *http.Request) uint64 {
	// query
	if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
		if v, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			return v
		}
	}
	// body (best-effort)
	var body struct {
		UserID uint64 `json:"user_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	return body.UserID
}

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	hostPort := strings.Split(r.RemoteAddr, ":")
	return hostPort[0]
}

func ToDigits(n uint8) otp.Digits {
	switch n {
	case 8:
		return otp.DigitsEight
	default:
		return otp.DigitsSix
	}
}

func ToAlgo(a string) otp.Algorithm {
	switch strings.ToUpper(a) {
	case "SHA256":
		return otp.AlgorithmSHA256
	case "SHA512":
		return otp.AlgorithmSHA512
	default:
		return otp.AlgorithmSHA1
	}
}
