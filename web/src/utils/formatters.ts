import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

dayjs.extend(relativeTime);

/**
 * Format a number as currency
 */
export function formatCurrency(
  amount: number,
  currency = "USD",
  locale = "en-US",
): string {
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount);
}

/**
 * Format a number with compact notation (e.g., 1.2M, 45K)
 */
export function formatCompactNumber(value: number): string {
  return new Intl.NumberFormat("en-US", {
    notation: "compact",
    compactDisplay: "short",
    maximumFractionDigits: 1,
  }).format(value);
}

/**
 * Format a number as a percentage
 */
export function formatPercentage(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

/**
 * Format a date string
 */
export function formatDate(
  date: string | Date,
  format = "MMM D, YYYY",
): string {
  return dayjs(date).format(format);
}

/**
 * Format a date as relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(date: string | Date): string {
  return dayjs(date).fromNow();
}

/**
 * Format a date with time
 */
export function formatDateTime(date: string | Date): string {
  return dayjs(date).format("MMM D, YYYY h:mm A");
}

/**
 * Get initials from a name
 */
export function getInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

/**
 * Truncate text with ellipsis
 */
export function truncate(text: string, maxLength = 50): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + "...";
}

/**
 * Generate a random avatar color based on string
 */
export function getAvatarColor(name: string): string {
  const colors = [
    "#7c3aed",
    "#6d28d9",
    "#8b5cf6",
    "#ec4899",
    "#f59e0b",
    "#10b981",
    "#3b82f6",
    "#ef4444",
  ];
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return colors[Math.abs(hash) % colors.length];
}

/**
 * Format deviation percentage with sign
 */
export function formatDeviation(value: number): string {
  const sign = value >= 0 ? "+" : "";
  return `${sign}${value.toFixed(1)}%`;
}

/**
 * Get severity color
 */
export function getSeverityColor(severity: string): string {
  const colors: Record<string, string> = {
    critical: "#ef4444",
    high: "#f97316",
    medium: "#f59e0b",
    low: "#3b82f6",
    info: "#94a3b8",
  };
  return colors[severity] || "#94a3b8";
}
