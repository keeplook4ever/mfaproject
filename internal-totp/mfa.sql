-- 创建数据库 mfa
CREATE DATABASE mfa
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE mfa;

-- 用户 TOTP 种子表
CREATE TABLE IF NOT EXISTS mfa_totp_seeds (
                                              id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, -- ✅ 新增自增ID
                                              user_id BIGINT UNSIGNED NOT NULL UNIQUE,
                                              secret_base32 VARCHAR(64) NOT NULL,
    period_seconds INT NOT NULL DEFAULT 30,
    digits INT NOT NULL DEFAULT 6,
    algo VARCHAR(16) NOT NULL DEFAULT 'SHA1',
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户TOTP秘钥表';

-- 备份码表
CREATE TABLE mfa_backup_codes (
                                  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
                                  user_id        BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                                  code_hash      CHAR(60) NOT NULL COMMENT '一次性备份码哈希（bcrypt/argon2）',
                                  used           TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已使用',
                                  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                  used_at        TIMESTAMP NULL DEFAULT NULL COMMENT '使用时间',
                                  INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户TOTP备份码表';

-- 审计日志表
CREATE TABLE mfa_audit_logs (
                                id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
                                user_id       BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                                action        VARCHAR(32) NOT NULL COMMENT '操作: enroll/activate/verify/disable/rotate',
                                success       TINYINT(1) NOT NULL COMMENT '是否成功',
                                ip_address    VARCHAR(45) DEFAULT NULL COMMENT '请求来源IP',
                                user_agent    VARCHAR(255) DEFAULT NULL COMMENT 'UA或调用方信息',
                                created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发生时间',
                                INDEX idx_user (user_id),
                                INDEX idx_action_time (action, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MFA操作审计日志';
