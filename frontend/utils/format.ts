import dayjs from "dayjs";
import "dayjs/locale/ru";

/** Currency formatter: 1 234 567 → "1 234 567" */
export function formatNumber(n: number | null | undefined, digits = 0): string {
  if (n == null || isNaN(Number(n))) return "—";
  return new Intl.NumberFormat("ru-RU", {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(Number(n));
}

export function formatPrice(n: number | null | undefined, currency = ""): string {
  if (n == null) return "—";
  const v = formatNumber(n);
  return currency ? `${v} ${currency}` : v;
}

export function formatDate(d: string | Date | null | undefined, fmt = "DD.MM.YYYY"): string {
  if (!d) return "—";
  return dayjs(d).format(fmt);
}

export function formatDateTime(d: string | Date | null | undefined): string {
  if (!d) return "—";
  return dayjs(d).format("DD.MM.YYYY HH:mm");
}

export function formatBytes(b: number): string {
  if (!b) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let i = 0;
  let v = b;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(1)} ${units[i]}`;
}

/** Color helpers based on string hashing (used for avatars/colors). */
export function stringToColor(s: string): string {
  let hash = 0;
  for (let i = 0; i < s.length; i++) hash = s.charCodeAt(i) + ((hash << 5) - hash);
  const h = Math.abs(hash) % 360;
  return `hsl(${h}, 65%, 55%)`;
}

export function initials(name?: string): string {
  if (!name) return "?";
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((s) => s[0]?.toUpperCase() || "")
    .join("");
}
