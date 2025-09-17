import React, { useMemo, useState } from "react";

export default function OriginalApp() {
    // ====== 可按需修改 ======
    const [userId, setUserId] = useState(1001);
    const apiBase = useMemo(() => "http://localhost:8080", []);
    // ======================

    const [qrDataUrl, setQrDataUrl] = useState(null); // data:image/png;base64,...
    const [secret, setSecret] = useState("");
    const [otp, setOtp] = useState("");
    const [status, setStatus] = useState("");
    const [loading, setLoading] = useState(false);

    // 绑定：从 /enroll 获取二维码（兼容后端两种返回形式）
    const enroll = async () => {
        if (!userId) return;
        setLoading(true);
        setStatus("正在申请二维码…");
        setQrDataUrl(null);
        setSecret("");

        try {
            const url = `${apiBase}/enroll?user_id=${encodeURIComponent(userId)}`;
            const res = await fetch(url);

            // 根据 Content-Type 判断后端返回形式
            const ctype = res.headers.get("content-type") || "";
            if (ctype.includes("application/json")) {
                // 约定 JSON 返回 { secret, qrcode_base64, provision_uri, ... }
                const data = await res.json();
                if (data.qrcode_base64) {
                    setQrDataUrl(`data:image/png;base64,${data.qrcode_base64}`);
                }
                if (data.secret) setSecret(data.secret);
            } else if (ctype.includes("image/png")) {
                // 直接返回 PNG 二进制
                const blob = await res.blob();
                const objectUrl = URL.createObjectURL(blob);
                setQrDataUrl(objectUrl);
            } else {
                throw new Error(`不支持的返回类型: ${ctype}`);
            }

            setStatus("请使用 Google Authenticator 扫码，然后输入 6 位验证码点击激活。");
        } catch (err) {
            console.error(err);
            setStatus(`绑定失败：${err.message}`);
        } finally {
            setLoading(false);
        }
    };

    // 激活：/activate
    const activate = async () => {
        if (!userId || !otp) return;
        setLoading(true);
        setStatus("正在激活…");
        try {
            const res = await fetch(`${apiBase}/activate`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ user_id: Number(userId), code: otp.trim() }),
            });
            const data = await res.json();
            if (data.ok) {
                setStatus("激活成功 ✅");
            } else {
                setStatus("激活失败 ❌，验证码不正确");
            }
        } catch (err) {
            console.error(err);
            setStatus(`激活请求失败：${err.message}`);
        } finally {
            setLoading(false);
        }
    };

    // 验证：/verify
    const verify = async () => {
        if (!userId || !otp) return;
        setLoading(true);
        setStatus("正在验证…");
        try {
            const res = await fetch(`${apiBase}/verify`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ user_id: Number(userId), code: otp.trim() }),
            });
            const data = await res.json();
            setStatus(data.ok ? "验证成功 ✅" : "验证失败 ❌");
        } catch (err) {
            console.error(err);
            setStatus(`验证请求失败：${err.message}`);
        } finally {
            setLoading(false);
        }
    };

    // 解绑/注销：/disable
    const disable = async () => {
        if (!userId) return;
        setLoading(true);
        setStatus("正在解绑/注销…");
        try {
            const res = await fetch(`${apiBase}/disable`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ user_id: Number(userId) }),
            });
            const data = await res.json().catch(() => ({}));
            if (res.ok && (data.ok === undefined || data.ok === true)) {
                setStatus("解绑成功 ✅");
                setQrDataUrl(null);
                setSecret("");
                setOtp("");
            } else {
                setStatus("解绑失败 ❌");
            }
        } catch (err) {
            console.error(err);
            setStatus(`解绑请求失败：${err.message}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
            <div className="w-full max-w-md bg-white rounded-2xl shadow p-6 space-y-5">
                <h1 className="text-2xl font-bold">公司内部 TOTP 管理</h1>

                {/* 用户ID输入 */}
                <div className="space-y-2">
                    <label className="text-sm text-gray-600">用户 ID</label>
                    <input
                        type="number"
                        className="w-full px-3 py-2 border rounded-md"
                        value={userId}
                        onChange={(e) => setUserId(e.target.value)}
                        placeholder="如 1001"
                    />
                </div>

                {/* 操作按钮 */}
                <div className="flex flex-wrap gap-2">
                    <button
                        onClick={enroll}
                        disabled={loading}
                        className="px-4 py-2 rounded-md text-white bg-blue-600 disabled:opacity-60"
                    >
                        绑定（生成二维码）
                    </button>

                    <button
                        onClick={activate}
                        disabled={loading || !otp}
                        className="px-4 py-2 rounded-md text-white bg-green-600 disabled:opacity-60"
                    >
                        激活
                    </button>

                    <button
                        onClick={verify}
                        disabled={loading || !otp}
                        className="px-4 py-2 rounded-md text-white bg-yellow-600 disabled:opacity-60"
                    >
                        验证
                    </button>

                    <button
                        onClick={disable}
                        disabled={loading}
                        className="px-4 py-2 rounded-md text-white bg-red-600 disabled:opacity-60"
                    >
                        解绑/注销
                    </button>
                </div>

                {/* OTP 输入框 */}
                <div className="space-y-2">
                    <label className="text-sm text-gray-600">一次性验证码（6位）</label>
                    <input
                        inputMode="numeric"
                        pattern="\d{6}"
                        maxLength={6}
                        className="w-full px-3 py-2 border rounded-md tracking-widest text-center font-mono"
                        value={otp}
                        onChange={(e) => setOtp(e.target.value.replace(/\D/g, "").slice(0, 6))}
                        placeholder="123456"
                    />
                </div>

                {/* 二维码与秘钥显示 */}
                {qrDataUrl && (
                    <div className="space-y-2">
                        <img
                            src={qrDataUrl}
                            alt="TOTP QR Code"
                            className="w-56 h-56 mx-auto border rounded-lg"
                        />
                        {secret && (
                            <p className="text-xs text-gray-500 break-words">
                                秘钥（备份用）：<span className="font-mono">{secret}</span>
                            </p>
                        )}
                    </div>
                )}

                {/* 状态提示 */}
                {status && (
                    <div className="text-center text-sm text-gray-800 bg-gray-100 rounded-md p-3">
                        {status}
                    </div>
                )}
            </div>
        </div>
    );
}
