package middleware

import (
	"net/http"
	"time"

	"internal-totp/internal/db"
	"internal-totp/internal/models"
	"internal-totp/internal/util"
)

// 包一层审计日志
func WithAudit(action string, next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		success := false
		defer func(start time.Time) {
			uid := util.ParseUserIDFromRequest(r)
			_ = db.DB.Create(&models.MFAAuditLog{
				UserID:    uid,
				Action:    action,
				Success:   success,
				IPAddress: util.ClientIP(r),
				UserAgent: r.UserAgent(),
			}).Error
		}(time.Now())

		ww := &wrapWriter{ResponseWriter: w, status: 200}
		next(ww, r)
		success = ww.status >= 200 && ww.status < 400
	}
}

type wrapWriter struct {
	http.ResponseWriter
	status int
}

func (w *wrapWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
