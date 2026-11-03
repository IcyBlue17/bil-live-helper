import dayjs from "dayjs";
import "dayjs/locale/zh-cn";
import duration from "dayjs/plugin/duration";
import relativeTime from "dayjs/plugin/relativeTime";
import { isCallError } from "mygo-runtime";

dayjs.extend(duration);
dayjs.extend(relativeTime);
dayjs.locale("zh-cn");

export { dayjs };

export function errorMessage(err: unknown): string {
  if (isCallError(err)) return err.message;
  if (err instanceof Error) return err.message;
  return String(err);
}

/** HH:MM:SS since an RFC 3339 time. */
export function elapsed(since: string | null, now: number): string {
  if (!since) return "00:00:00";
  const t = dayjs(since);
  if (!t.isValid()) return "00:00:00";
  const d = dayjs.duration(Math.max(0, now - t.valueOf()));
  return `${String(Math.floor(d.asHours())).padStart(2, "0")}:${d.format("mm:ss")}`;
}

export function clock(time: string): string {
  const t = dayjs(time);
  return t.isValid() ? t.format("HH:mm:ss") : "--:--";
}

/** Compact Chinese number: 12873 → 1.3万. */
export function compact(n: number): string {
  if (n >= 1e8) return `${(n / 1e8).toFixed(1).replace(/\.0$/, "")}亿`;
  if (n >= 1e4) return `${(n / 1e4).toFixed(1).replace(/\.0$/, "")}万`;
  return n.toLocaleString("zh-CN");
}

/** Intl options for NumberFlow, matching compact(). */
export const compactFormat: Intl.NumberFormatOptions = { notation: "compact", maximumFractionDigits: 1 };

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

export const guardName = { 0: "", 1: "总督", 2: "提督", 3: "舰长" } as const;

export const spaceURL = (uid: number) => `https://space.bilibili.com/${uid}`;
