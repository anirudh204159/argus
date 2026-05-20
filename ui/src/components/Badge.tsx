import { cn } from "../lib/cn";

interface BadgeProps {
  variant?: "default" | "success" | "warning" | "danger" | "muted";
  children: React.ReactNode;
  className?: string;
}

const VARIANTS = {
  default: "bg-accent-subtle text-accent border-accent/30",
  success: "bg-success-subtle text-success border-success/30",
  warning: "bg-warning-subtle text-warning border-warning/30",
  danger: "bg-danger-subtle text-danger border-danger/30",
  muted: "bg-bg-hover text-text-muted border-border",
} as const;

export function Badge({ variant = "default", children, className }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs font-medium border",
        VARIANTS[variant],
        className
      )}
    >
      {children}
    </span>
  );
}