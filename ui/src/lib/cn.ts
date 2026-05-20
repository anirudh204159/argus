import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Combine class names intelligently.
 * Handles conditional classes (via clsx) and Tailwind conflicts (via twMerge).
 *
 * Example:
 *   cn("px-2 py-1", isActive && "bg-accent", "px-4") => "py-1 bg-accent px-4"
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}