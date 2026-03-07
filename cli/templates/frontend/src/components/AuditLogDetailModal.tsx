import type { AuditLog } from '../lib/types'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from './ui/dialog'
import { Badge } from './ui/badge'
import { Separator } from './ui/separator'
import { format } from 'date-fns'
import { Clock, Globe, UserX, FileText } from 'lucide-react'

interface AuditLogDetailModalProps {
  log: AuditLog | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function actionBadgeVariant(action: string): "default" | "secondary" | "destructive" | "outline" {
  if (action.includes('delete') || action.includes('remove') || action.includes('revoked') || action.includes('deactivated')) return 'destructive'
  if (action.includes('create') || action.includes('granted') || action.includes('bootstrap')) return 'default'
  if (action.includes('update') || action.includes('enabled')) return 'secondary'
  return 'outline'
}

export function AuditLogDetailModal({ log, open, onOpenChange }: AuditLogDetailModalProps) {
  if (!log) return null

  const isImpersonationEvent = log.action.includes('impersonation') ||
    (log.changes && '_impersonation' in log.changes)

  const impersonationData = log.changes?._impersonation as Record<string, string> | undefined

  // Extract non-impersonation changes
  const displayChanges = log.changes
    ? Object.fromEntries(Object.entries(log.changes).filter(([k]) => k !== '_impersonation'))
    : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            Audit Log #{log.id}
            <Badge variant={actionBadgeVariant(log.action)}>{log.action}</Badge>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Basic Info */}
          <div>
            <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground mb-2">
              <FileText className="h-4 w-4" />
              Basic Info
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-muted-foreground">Timestamp</p>
                <p className="font-medium">{format(new Date(log.timestamp), 'PPpp')}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Actor</p>
                <p className="font-medium">{log.actor_name || log.actor_user_id || 'System'}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Resource Type</p>
                <p className="font-medium">{log.resource_type}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Resource ID</p>
                <p className="font-medium font-mono text-xs">{log.resource_id}</p>
              </div>
            </div>
          </div>

          {/* Network Info */}
          {(log.ip_address || log.user_agent) && (
            <>
              <Separator />
              <div>
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground mb-2">
                  <Globe className="h-4 w-4" />
                  Network Info
                </div>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  {log.ip_address && (
                    <div>
                      <p className="text-muted-foreground">IP Address</p>
                      <p className="font-medium font-mono text-xs">{log.ip_address}</p>
                    </div>
                  )}
                  {log.user_agent && (
                    <div className="col-span-2">
                      <p className="text-muted-foreground">User Agent</p>
                      <p className="font-medium text-xs truncate">{log.user_agent}</p>
                    </div>
                  )}
                </div>
              </div>
            </>
          )}

          {/* Impersonation Details */}
          {(isImpersonationEvent || impersonationData) && (
            <>
              <Separator />
              <div className="rounded-md border border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950 p-3">
                <div className="flex items-center gap-2 text-sm font-medium text-amber-800 dark:text-amber-200 mb-2">
                  <UserX className="h-4 w-4" />
                  Impersonation Details
                </div>
                {impersonationData && (
                  <div className="grid grid-cols-2 gap-2 text-sm text-amber-900 dark:text-amber-100">
                    <div>
                      <p className="text-amber-700 dark:text-amber-300 text-xs">Impersonated User</p>
                      <p className="font-medium text-xs">{impersonationData.impersonated_user_email}</p>
                    </div>
                    <div>
                      <p className="text-amber-700 dark:text-amber-300 text-xs">Performed By Admin</p>
                      <p className="font-medium font-mono text-xs">{impersonationData.performed_by_admin_id}</p>
                    </div>
                  </div>
                )}
              </div>
            </>
          )}

          {/* Changes */}
          {displayChanges && Object.keys(displayChanges).length > 0 && (
            <>
              <Separator />
              <div>
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground mb-2">
                  <Clock className="h-4 w-4" />
                  Changes
                </div>
                <pre className="bg-muted rounded-md p-3 text-xs overflow-auto max-h-64">
                  {JSON.stringify(displayChanges, null, 2)}
                </pre>
              </div>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
