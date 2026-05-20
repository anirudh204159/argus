import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "../lib/cn";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, id, ...props }, ref) => {
    return (
      <div className="space-y-1.5">
        {label && (
          <label htmlFor={id} className="block text-sm font-medium text-text">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={cn(
            "w-full px-3 py-2 bg-bg-surface border border-border rounded-md",
            "text-sm text-text placeholder:text-text-subtle",
            "focus:outline-none focus:border-accent focus:shadow-glow",
            "transition-colors",
            "disabled:opacity-50 disabled:cursor-not-allowed",
            error && "border-danger focus:border-danger focus:shadow-none",
            className
          )}
          {...props}
        />
        {error && <p className="text-xs text-danger">{error}</p>}
      </div>
    );
  }
);

Input.displayName = "Input";