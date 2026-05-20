import { NavLink } from "react-router-dom";
import { Activity, Database, Webhook, AlertCircle, LayoutDashboard } from "lucide-react";
import { cn } from "../lib/cn";

const NAV_ITEMS = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard },
  { to: "/sources", label: "Sources", icon: Database },
  { to: "/subscriptions", label: "Subscriptions", icon: Webhook },
  { to: "/events", label: "Events", icon: Activity },
  { to: "/dlq", label: "Dead Letter Queue", icon: AlertCircle },
] as const;

export function Sidebar() {
  return (
    <aside className="w-60 bg-bg-surface border-r border-border flex flex-col">
      {/* Logo */}
      <div className="px-5 py-5 border-b border-border">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 bg-accent rounded-md flex items-center justify-center">
            <Activity className="w-4 h-4 text-white" strokeWidth={2.5} />
          </div>
          <span className="text-base font-semibold text-text">Argus</span>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-3 py-4 space-y-1">
        {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === "/"}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors",
                isActive
                  ? "bg-bg-hover text-text"
                  : "text-text-muted hover:text-text hover:bg-bg-hover"
              )
            }
          >
            <Icon className="w-4 h-4" />
            <span>{label}</span>
          </NavLink>
        ))}
      </nav>

      {/* Footer */}
      <div className="px-5 py-4 border-t border-border">
        <div className="text-xs text-text-subtle">
          v0.2.0 · <a href="https://github.com/anirudh204159/argus" target="_blank" rel="noopener" className="hover:text-text-muted">GitHub</a>
        </div>
      </div>
    </aside>
  );
}