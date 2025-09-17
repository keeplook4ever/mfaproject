import React, { useMemo, useState } from "react";

export default function App() {
    const [userId, setUserId] = useState(1001);
    const apiBase = useMemo(() => "http://localhost:8080", []);
    const [qrDataUrl, setQrDataUrl] = useState(null);
    const [secret, setSecret] = useState("");
    const [otp, setOtp] = useState("");
    const [status, setStatus] = useState("");
    const [loading, setLoading] = useState(false);

    const enroll = async () => {
        setStatus("正在申请二维码…");
        setLoading(true);
        setQrDataUrl(null);
        setSecret("");

        try {
            const res = await fetch(`${apiBase}/enroll?user_id=${encodeURIComponent(userId)}`, {
                method: "GET",
                headers: { "Accept": "application/json, image/png" },
            });

            const ct = res.headers.get("content-type") || "";
            if (ct.includes("application/json")) {
                const data = await res.json();
                console.log("[enroll][json]", data);
                if (data.qrcode_base64) setQrDataUrl(`data:image/png;base64,${data.qrcode_base64}`);
                if (data.secret) setSecret(data.secret);
                setStatus("二维码已生成，请扫码并输入 6 位验证码后点击“激活”。");
            } else if (ct.includes("image/png")) {
                const blob = await res.blob();
                const url = URL.createObjectURL(blob);
                setQrDataUrl(url);
                setStatus("二维码已生成，请扫码并输入 6 位验证码后点击“激活”。");
            } else {
                const text = await res.text();
                throw new Error(`返回类型不支持：${ct}，响应：${text}`);
            }
        } catch (e) {
            console.error("[enroll][error]", e);
            setStatus(`绑定失败：${e.message}`);
        } finally {
            setLoading(false);
        }
    };

    const activate = async () => {
        if (!otp) { setStatus("请输入 6 位验证码再激活"); return; }
        setLoading(true);
        setStatus("正在激活…");
        try {
            const body = { user_id: Number(userId), code: otp.trim() };
            console.log("[activate][req]", body);
            const res = await fetch(`${apiBase}/activate`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const data = await safeJson(res);
            console.log("[activate][res]", res.status, data);
            if (res.ok && data?.ok) {
                setStatus("激活成功 ✅");
            } else {
                setStatus(`激活失败 ❌（HTTP ${res.status}）`);
            }
        } catch (e) {
            console.error("[activate][error]", e);
            setStatus(`激活请求异常：${e.message}`);
        } finally {
            setLoading(false);
        }
    };

    const verify = async () => {
        if (!otp) { setStatus("请输入 6 位验证码再验证"); return; }
        setLoading(true);
        setStatus("正在验证…");
        try {
            const body = { user_id: Number(userId), code: otp.trim() };
            console.log("[verify][req]", body);
            const res = await fetch(`${apiBase}/verify`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const data = await safeJson(res);
            console.log("[verify][res]", res.status, data);
            if (res.ok && data?.ok === true) {
                setStatus("验证成功 ✅");
            } else if (res.ok && data?.ok === false) {
                setStatus("验证失败 ❌（验证码不正确或过期）");
            } else {
                setStatus(`验证失败 ❌（HTTP ${res.status}）`);
            }
        } catch (e) {
            console.error("[verify][error]", e);
            setStatus(`验证请求异常：${e.message}`);
        } finally {
            setLoading(false);
        }
    };

    const disable = async () => {
        setLoading(true);
        setStatus("正在解绑/注销…");
        try {
            const body = { user_id: Number(userId) };
            console.log("[disable][req]", body);
            const res = await fetch(`${apiBase}/disable`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const data = await safeJson(res);
            console.log("[disable][res]", res.status, data);
            if (res.ok && (data?.ok === undefined || data?.ok === true)) {
                setStatus("解绑成功 ✅");
                setQrDataUrl(null);
                setSecret("");
                setOtp("");
            } else {
                setStatus(`解绑失败 ❌（HTTP ${res.status}）`);
            }
        } catch (e) {
            console.error("[disable][error]", e);
            setStatus(`解绑请求异常：${e.message}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
            <div className="w-full max-w-md bg-white rounded-2xl shadow p-6 space-y-5">
                <h1 className="text-2xl font-bold">公司内部 TOTP 管理</h1>

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

                <div className="flex flex-wrap gap-2">
                    <button onClick={enroll}  disabled={loading} className="px-4 py-2 rounded-md text-white bg-blue-600 disabled:opacity-60">绑定（二维码）</button>
                    <button onClick={activate} disabled={loading} className="px-4 py-2 rounded-md text-white bg-green-600 disabled:opacity-60">激活</button>
                    <button onClick={verify}   disabled={loading} className="px-4 py-2 rounded-md text-white bg-yellow-600 disabled:opacity-60">验证</button>
                    <button onClick={disable}  disabled={loading} className="px-4 py-2 rounded-md text-white bg-red-600 disabled:opacity-60">解绑/注销</button>
                </div>

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

                {qrDataUrl && (
                    <div className="space-y-2">
                        <img src={qrDataUrl} alt="TOTP QR" className="w-56 h-56 mx-auto border rounded-lg" />
                        {secret && <p className="text-xs text-gray-500 break-words">秘钥：<span className="font-mono">{secret}</span></p>}
                    </div>
                )}

                {status && <div className="text-center text-sm text-gray-800 bg-gray-100 rounded-md p-3">{status}</div>}
            </div>
        </div>
    );
}

// 响应可能不是 JSON（例如 204 / 文本），安全解析
async function safeJson(res) {
    const ct = res.headers.get("content-type") || "";
    if (ct.includes("application/json")) return res.json();
    const text = await res.text();
    try { return JSON.parse(text) } catch { return { raw: text } }
}
