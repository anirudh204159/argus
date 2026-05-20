import { Database, Webhook, Activity, AlertCircle } from "lucide-react";
import { Header } from "../components/Header";

interface StatCardProps {
  label: string;
  value: string | number;
  icon: React.ComponentType<{ className?: string }>;
  trend?: string;
}

function StatCard({ label, value, icon: Icon, trend }: StatCardProps) {
  return (
    <div className="bg-bg-surface border border-border rounded-lg p-5">
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm text-text-muted">{label}</span>
        <Icon className="w-4 h-4 text-text-subtle" />
      </div>
      <div className="text-2xl font-semibold text-text">{value}</div>
      {trend && <div className="text-xs text-text-subtle mt-1">{trend}</div>}
    </div>
  );
}

export function Dashboard() {
  return (
    <>
      <Header
        title="Dashboard"
        description="Overview of your CDC pipelines"
      />
      <div className="flex-1 overflow-y-auto p-8">
        <div className="max-w-6xl">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <StatCard label="Sources" value="—" icon={Database} trend="Connect a source to begin" />
            <StatCard label="Subscriptions" value="—" icon={Webhook} trend="No active subscriptions" />
            <StatCard label="Events processed" value="—" icon={Activity} trend="Last 24 hours" />
            <StatCard label="Failed deliveries" value="—" icon={AlertCircle} trend="In DLQ" />
          </div>

          <div className="bg-bg-surface border border-border rounded-lg p-6">
            <h2 className="text-base font-semibold text-text mb-1">Getting started</h2>
            <p className="text-sm text-text-muted mb-4">
              Argus captures row-level changes from MySQL and delivers them to your webhooks in real time.
            </p>
            <ol className="space-y-2.5 text-sm text-text">
              <li className="flex gap-3">
                <span className="flex-shrink-0 w-5 h-5 bg-accent-subtle text-accent rounded-full flex items-center justify-center text-xs font-medium">1</span>
                <span>Add a <span className="text-text">MySQL source</span> with binlog enabled.</span>
              </li>
              <li className="flex gap-3">
                <span className="flex-shrink-0 w-5 h-5 bg-accent-subtle text-accent rounded-full flex items-center justify-center text-xs font-medium">2</span>
                <span>Create a <span className="text-text">subscription</span> pointing at your webhook URL.</span>
              </li>
              <li className="flex gap-3">
                <span className="flex-shrink-0 w-5 h-5 bg-accent-subtle text-accent rounded-full flex items-center justify-center text-xs font-medium">3</span>
                <span>Watch events flow in the <span className="text-text">Events</span> view as they happen.</span>
              </li>
            </ol>
          </div>
        </div>
      </div>
    </>
  );
}