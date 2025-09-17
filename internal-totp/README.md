# 创建数据库mfa和数据表mfa_totp_seeds

sql在mfa.sql里


# 创建应用数据库账号
CREATE USER 'otp'@'127.0.0.1' IDENTIFIED BY '%%%%%';
GRANT ALL PRIVILEGES ON mfa.* TO 'otp'@'127.0.0.1';
FLUSH PRIVILEGES;

# 编译go代码并开启服务
配置 goland run kind: Package