package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"

	"internal-totp/internal/config"
	"internal-totp/internal/db"
	"internal-totp/internal/models"
	"internal-totp/internal/util"
)

// DTO
type EnrollResponse struct {
	Secret       string `json:"secret"`
	ProvisionURI string `json:"provision_uri"`
	QRCodeBase64 string `json:"qrcode_base64"`
	Digits       int    `json:"digits"`
	Period       int    `json:"period"`
	Algorithm    string `json:"algorithm"`
}
type VerifyRequest struct {
	UserID uint64 `json:"user_id"`
	Code   string `json:"code"`
}
type VerifyResponse struct {
	Ok bool `json:"ok"`
}

// GET /enroll?user_id=1001
func HandleEnroll(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			util.HTTPError(w, http.StatusBadRequest, "user_id required")
			return
		}
		uid, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			util.HTTPError(w, http.StatusBadRequest, "invalid user_id")
			return
		}

		// 生成 TOTP Key
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      cfg.Issuer,
			AccountName: fmt.Sprintf("user%d@%s", uid, strings.ToLower(cfg.Issuer)),
			Period:      30,
			Digits:      otp.DigitsSix,
			Algorithm:   otp.AlgorithmSHA1,
		})
		if err != nil {
			util.HTTPError(w, http.StatusInternalServerError, "failed to generate key")
			return
		}

		// 二维码 PNG -> base64
		qrBytes, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
		if err != nil {
			util.HTTPError(w, http.StatusInternalServerError, "failed to generate qrcode")
			return
		}
		qrBase64 := base64.StdEncoding.EncodeToString(qrBytes)

		// upsert mfa_totp_seeds
		seed := models.MFATOTPSeed{
			UserID:        uid,
			SecretBase32:  key.Secret(), // 生产建议改为密文存储
			Status:        "pending",
			Algo:          "SHA1",
			Digits:        6,
			PeriodSeconds: 30,
		}
		err = db.DB.Transaction(func(tx *gorm.DB) error {
			var existing models.MFATOTPSeed
			if err := tx.Where("user_id = ?", uid).First(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return tx.Create(&seed).Error
				}
				return err
			}
			existing.SecretBase32 = seed.SecretBase32
			existing.Status = "pending"
			existing.Algo = seed.Algo
			existing.Digits = seed.Digits
			existing.PeriodSeconds = seed.PeriodSeconds
			return tx.Save(&existing).Error
		})
		if err != nil {
			util.HTTPError(w, http.StatusInternalServerError, "db upsert error")
			return
		}

		resp := EnrollResponse{
			Secret:       key.Secret(),
			ProvisionURI: key.URL(),
			QRCodeBase64: qrBase64,
			Digits:       6,
			Period:       30,
			Algorithm:    "SHA1",
		}
		util.WriteJSON(w, resp)
	}
}

// POST /activate { "user_id":1001, "code":"123456" }
func HandleActivate(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 || req.Code == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var m models.MFATOTPSeed
	if err := db.DB.Where("user_id = ?", req.UserID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	ok, _ := totp.ValidateCustom(req.Code, m.SecretBase32, time.Now(),
		totp.ValidateOpts{
			Period:    uint(m.PeriodSeconds),
			Skew:      1,
			Digits:    util.ToDigits(uint8(m.Digits)),
			Algorithm: util.ToAlgo(m.Algo),
		})

	if ok && m.Status == "pending" {
		_ = db.DB.Model(&models.MFATOTPSeed{}).Where("user_id = ?", req.UserID).
			Update("status", "active").Error
	}

	util.WriteJSON(w, VerifyResponse{Ok: ok})
}

// POST /verify { "user_id":1001, "code":"654321" }
func HandleVerify(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.HTTPError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == 0 || req.Code == "" {
		util.HTTPError(w, http.StatusBadRequest, "user_id and code required")
		return
	}

	var m models.MFATOTPSeed
	if err := db.DB.Where("user_id = ?", req.UserID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.HTTPError(w, http.StatusNotFound, "user not found")
			return
		}
		util.HTTPError(w, http.StatusInternalServerError, "db error")
		return
	}

	ok, _ := totp.ValidateCustom(req.Code, m.SecretBase32, time.Now(),
		totp.ValidateOpts{
			Period:    uint(m.PeriodSeconds),
			Skew:      1,
			Digits:    util.ToDigits(uint8(m.Digits)),
			Algorithm: util.ToAlgo(m.Algo),
		})

	// 可选：首次激活
	if ok && m.Status == "pending" {
		_ = db.DB.Model(&models.MFATOTPSeed{}).Where("user_id = ?", req.UserID).
			Update("status", "active").Error
	}

	util.WriteJSON(w, VerifyResponse{Ok: ok})
}

// POST /disable { "user_id":1001 }
func HandleDisable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uint64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	result := db.DB.Model(&models.MFATOTPSeed{}).
		Where("user_id = ?", req.UserID).
		Update("status", "disabled")

	if result.Error != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	_ = db.DB.Create(&models.MFAAuditLog{
		UserID:    req.UserID,
		Action:    "disable",
		Success:   true,
		IPAddress: util.ClientIP(r),
		UserAgent: r.UserAgent(),
	}).Error

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}
