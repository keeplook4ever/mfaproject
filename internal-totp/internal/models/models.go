package models

import "time"

// mfa_totp_seeds
type MFATOTPSeed struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"uniqueIndex;not null"      json:"user_id"`
	SecretBase32  string    `gorm:"size:64;not null"          json:"secret_base32"`
	PeriodSeconds int       `gorm:"default:30"                json:"period_seconds"`
	Digits        int       `gorm:"default:6"                 json:"digits"`
	Algo          string    `gorm:"size:16;default:SHA1"      json:"algo"`
	Status        string    `gorm:"size:16;default:pending"   json:"status"` // pending/active/disabled
	CreatedAt     time.Time `gorm:"autoCreateTime"            json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"            json:"updated_at"`
}

func (MFATOTPSeed) TableName() string { return "mfa_totp_seeds" }

// mfa_backup_codes
type MFABackupCode struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    uint64     `gorm:"index:idx_user;not null;column:user_id"`
	CodeHash  string     `gorm:"size:100;not null;column:code_hash"`
	Used      bool       `gorm:"not null;default:false;column:used"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UsedAt    *time.Time `gorm:"column:used_at"`
}

func (MFABackupCode) TableName() string { return "mfa_backup_codes" }

// mfa_audit_logs
type MFAAuditLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    uint64    `gorm:"index:idx_user;not null;column:user_id"`
	Action    string    `gorm:"size:32;not null;column:action"` // enroll/activate/verify/disable/rotate
	Success   bool      `gorm:"not null;column:success"`
	IPAddress string    `gorm:"size:45;column:ip_address"`
	UserAgent string    `gorm:"size:255;column:user_agent"`
	CreatedAt time.Time `gorm:"index:idx_action_time;autoCreateTime;column:created_at"`
}

func (MFAAuditLog) TableName() string { return "mfa_audit_logs" }
