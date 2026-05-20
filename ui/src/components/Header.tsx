import { LogOut, User } from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { Button } from "./Button";

interface HeaderProps {
  title: string;
  description?: string;
  actions?: React.ReactNode;
}

export function Header({ title, description, actions }: HeaderProps) {
  const { logout } = useAuth();

  return (
    <header className="border-b border-border bg-bg">
      <div className="px-8 py-5 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text">{title}</h1>
          {description && (
            <p className="text-sm text-text-muted mt-0.5">{description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          {actions}
          <div className="w-px h-6 bg-border mx-2" />
          <div className="flex items-center gap-2 px-2.5 py-1.5 text-sm text-text-muted">
            <User className="w-4 h-4" />
          </div>
          <Button variant="ghost" size="sm" onClick={logout}>
            <LogOut className="w-4 h-4" />
            <span>Sign out</span>
          </Button>
        </div>
      </div>
    </header>
  );
}