import { type ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type Variant = "default" | "ghost" | "outline";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
};

export function Button({ className, variant = "default", ...props }: ButtonProps) {
  return (
    <button
      className={cn(
        "inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition",
        variant === "default" && "bg-[var(--primary)] text-[var(--primary-foreground)] hover:opacity-90",
        variant === "ghost" && "hover:bg-slate-100",
        variant === "outline" && "border border-[var(--border)] bg-white hover:bg-slate-50",
        className
      )}
      {...props}
    />
  );
}
