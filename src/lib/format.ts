import { isCallError } from "mygo-runtime";

export function errorMessage(err: unknown): string {
  if (isCallError(err)) return err.message;
  if (err instanceof Error) return err.message;
  return String(err);
}

/** HH:MM:SS since an RFC 3339 time. */
export function duration(since: string | null, now: number): string {
  if (!since) return "00:00:00";
  const t = Date.parse(since);
  if (Number.isNaN(t)) return "00:00:00";
  const s = Math.max(0, Math.floor((now - t) / 1000));
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(Math.floor(s / 3600))}:${pad(Math.floor((s % 3600) / 60))}:${pad(s % 60)}`;
}

export function clock(time: string): string {
  const d = new Date(time);
  if (Number.isNaN(d.getTime())) return "--:--";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

/** Compact Chinese number: 12873 → 1.3万. */
export function compact(n: number): string {
  if (n >= 1e8) return `${(n / 1e8).toFixed(1).replace(/\.0$/, "")}亿`;
  if (n >= 1e4) return `${(n / 1e4).toFixed(1).replace(/\.0$/, "")}万`;
  return n.toLocaleString("zh-CN");
}

/** Counts characters as Bilibili does: by code point. */
export function charCount(s: string): number {
  return [...s].length;
}

export function yuan(gold: number): string {
  const v = gold / 1000;
  return `¥${v >= 10 ? v.toFixed(0) : v.toFixed(1).replace(/\.0$/, "")}`;
}

/** Masks the key parameter of a stream code, keeping its shape. */
export function maskKey(code: string): string {
  return code.replace(/(key=)([^&]+)/, (_, k: string, v: string) => k + "•".repeat(Math.min(v.length, 16)));
}
