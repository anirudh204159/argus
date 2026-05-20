import { useState, type FormEvent } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, Pencil, Database } from "lucide-react";
import { Header } from "../components/Header";
import { Button } from "../components/Button";
import { Input } from "../components/Input";
import { Modal } from "../components/Modal";
import { Badge } from "../components/Badge";
import { listSources, createSource, updateSource, deleteSource } from "../api/sources";
import { getErrorMessage } from "../api/client";
import type { Source, SourceCreate } from "../types/api";

const EMPTY_FORM: SourceCreate = {
  name: "",
  host: "",
  port: 3306,
  database_name: "",
  replication_user: "",
  replication_password: "",
};

export function Sources() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<Source | null>(null);
  const [deleting, setDeleting] = useState<Source | null>(null);

  const { data: sources = [], isLoading } = useQuery({
    queryKey: ["sources"],
    queryFn: listSources,
  });

  const createMutation = useMutation({
    mutationFn: createSource,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sources"] });
      setCreateOpen(false);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: Partial<SourceCreate> }) =>
      updateSource(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sources"] });
      setEditing(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteSource,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sources"] });
      setDeleting(null);
    },
  });

  return (
    <>
      <Header
        title="Sources"
        description="MySQL databases that Argus replicates from"
        actions={
          <Button size="md" onClick={() => setCreateOpen(true)}>
            <Plus className="w-4 h-4" />
            New source
          </Button>
        }
      />

      <div className="flex-1 overflow-y-auto p-8">
        {isLoading ? (
          <div className="text-sm text-text-muted">Loading sources...</div>
        ) : sources.length === 0 ? (
          <EmptyState onCreate={() => setCreateOpen(true)} />
        ) : (
          <div className="bg-bg-surface border border-border rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-bg-hover/50 border-b border-border">
                <tr>
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wide">Name</th>
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wide">Host</th>
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wide">Database</th>
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wide">Status</th>
                  <th className="text-right px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wide w-24">Actions</th>
                </tr>
              </thead>
              <tbody>
                {sources.map((source) => (
                  <tr key={source.id} className="border-b border-border last:border-0 hover:bg-bg-hover/30">
                    <td className="px-4 py-3 text-sm font-medium text-text">{source.name}</td>
                    <td className="px-4 py-3 text-sm font-mono text-text-muted">
                      {source.host}:{source.port}
                    </td>
                    <td className="px-4 py-3 text-sm font-mono text-text-muted">{source.database_name}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={source.status} />
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="inline-flex items-center gap-1">
                        <Button variant="ghost" size="sm" onClick={() => setEditing(source)} aria-label="Edit">
                          <Pencil className="w-3.5 h-3.5" />
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => setDeleting(source)} aria-label="Delete">
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create Modal */}
      <SourceFormModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title="New source"
        description="Add a MySQL database to replicate from"
        submitLabel="Create source"
        loading={createMutation.isPending}
        error={createMutation.error ? getErrorMessage(createMutation.error) : null}
        onSubmit={(values) => createMutation.mutate(values)}
      />

      {/* Edit Modal */}
      {editing && (
        <SourceFormModal
          open={!!editing}
          onClose={() => setEditing(null)}
          title="Edit source"
          description={`Update ${editing.name}`}
          submitLabel="Save changes"
          initial={editing}
          loading={updateMutation.isPending}
          error={updateMutation.error ? getErrorMessage(updateMutation.error) : null}
          onSubmit={(values) => updateMutation.mutate({ id: editing.id, payload: values })}
        />
      )}

      {/* Delete Confirm Modal */}
      <Modal
        open={!!deleting}
        onClose={() => setDeleting(null)}
        title="Delete source?"
        description="This cannot be undone. All subscriptions on this source will also be deleted."
      >
        {deleting && (
          <div className="space-y-4">
            <div className="px-3 py-2 bg-bg-hover border border-border rounded-md">
              <div className="text-sm font-medium text-text">{deleting.name}</div>
              <div className="text-xs text-text-muted font-mono mt-0.5">
                {deleting.host}:{deleting.port} / {deleting.database_name}
              </div>
            </div>
            {deleteMutation.error && (
              <div className="px-3 py-2 bg-danger-subtle border border-danger/30 rounded-md text-sm text-danger">
                {getErrorMessage(deleteMutation.error)}
              </div>
            )}
            <div className="flex gap-2 justify-end">
              <Button variant="secondary" onClick={() => setDeleting(null)}>Cancel</Button>
              <Button
                variant="danger"
                loading={deleteMutation.isPending}
                onClick={() => deleteMutation.mutate(deleting.id)}
              >
                Delete source
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}

// ---------- Subcomponents ----------

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="w-12 h-12 bg-bg-surface border border-border rounded-lg flex items-center justify-center mb-4">
        <Database className="w-5 h-5 text-text-muted" />
      </div>
      <h3 className="text-base font-semibold text-text mb-1">No sources yet</h3>
      <p className="text-sm text-text-muted mb-5 max-w-xs">
        Connect your first MySQL database to start capturing change events.
      </p>
      <Button onClick={onCreate}>
        <Plus className="w-4 h-4" />
        Add source
      </Button>
    </div>
  );
}

function StatusBadge({ status }: { status: Source["status"] }) {
  if (status === "connected") return <Badge variant="success">● Connected</Badge>;
  if (status === "error") return <Badge variant="danger">● Error</Badge>;
  return <Badge variant="muted">○ Disconnected</Badge>;
}

interface SourceFormModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  submitLabel: string;
  initial?: Source;
  loading: boolean;
  error: string | null;
  onSubmit: (values: SourceCreate) => void;
}

function SourceFormModal({
  open, onClose, title, description, submitLabel, initial, loading, error, onSubmit,
}: SourceFormModalProps) {
  const [values, setValues] = useState<SourceCreate>(
    initial
      ? { ...EMPTY_FORM, ...initial, replication_password: "" }
      : EMPTY_FORM
  );

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    // For edit, skip empty password (means "don't change")
    const payload = initial && !values.replication_password
      ? { ...values, replication_password: undefined as unknown as string }
      : values;
    // Strip out the password field if empty on edit
    if (initial && !values.replication_password) {
      const { replication_password: _, ...rest } = payload;
      onSubmit(rest as SourceCreate);
    } else {
      onSubmit(payload);
    }
  }

  function update<K extends keyof SourceCreate>(key: K, value: SourceCreate[K]) {
    setValues((v) => ({ ...v, [key]: value }));
  }

  return (
    <Modal open={open} onClose={onClose} title={title} description={description} size="lg">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Name"
          required
          value={values.name}
          onChange={(e) => update("name", e.target.value)}
          placeholder="production-orders"
        />
        <div className="grid grid-cols-3 gap-3">
          <div className="col-span-2">
            <Input
              label="Host"
              required
              value={values.host}
              onChange={(e) => update("host", e.target.value)}
              placeholder="db.example.com"
            />
          </div>
          <Input
            label="Port"
            type="number"
            required
            value={values.port}
            onChange={(e) => update("port", parseInt(e.target.value) || 3306)}
          />
        </div>
        <Input
          label="Database"
          required
          value={values.database_name}
          onChange={(e) => update("database_name", e.target.value)}
          placeholder="my_database"
        />
        <div className="grid grid-cols-2 gap-3">
          <Input
            label="Replication user"
            required
            value={values.replication_user}
            onChange={(e) => update("replication_user", e.target.value)}
            placeholder="argus_user"
          />
          <Input
            label={initial ? "Password (leave blank to keep)" : "Password"}
            type="password"
            required={!initial}
            value={values.replication_password}
            onChange={(e) => update("replication_password", e.target.value)}
            placeholder="••••••••"
          />
        </div>

        {error && (
          <div className="px-3 py-2 bg-danger-subtle border border-danger/30 rounded-md text-sm text-danger">
            {error}
          </div>
        )}

        <div className="flex gap-2 justify-end pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>Cancel</Button>
          <Button type="submit" loading={loading}>{submitLabel}</Button>
        </div>
      </form>
    </Modal>
  );
}