import { useState, useEffect, useCallback } from 'react'
import { useAdmin } from '../../contexts/AdminContext'
import { useBreadcrumb } from '../../contexts/BreadcrumbContext'
import { getApiClient } from '../../lib/api'
import type { AuditLog, AuditLogFilters, AuditLogStats } from '../../lib/types'
import { AdminGuard } from '../../components/AdminGuard'
import { AuditLogDetailModal } from '../../components/AuditLogDetailModal'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../../components/ui/table'
import { Input } from '../../components/ui/input'
import { Button } from '../../components/ui/button'
import { Badge } from '../../components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../../components/ui/select'
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card'
import { toast } from 'sonner'
import { format } from 'date-fns'
import { Search, FileText, ChevronLeft, ChevronRight, Download } from 'lucide-react'

const PAGE_SIZE = 50

function actionBadgeVariant(action: string): "default" | "secondary" | "destructive" | "outline" {
  if (action.includes('delete') || action.includes('remove') || action.includes('revoked') || action.includes('deactivated')) return 'destructive'
  if (action.includes('create') || action.includes('granted') || action.includes('bootstrap')) return 'default'
  if (action.includes('update') || action.includes('enabled') || action.includes('logged_in')) return 'secondary'
  return 'outline'
}

function extractResourceName(log: AuditLog): string {
  if (!log.changes) return ''
  const changes = log.changes as Record<string, unknown>
  if (changes.email) return String(changes.email)
  if (changes.target_user_email) return String(changes.target_user_email)
  if (changes.admin_email) return String(changes.admin_email)
  return ''
}

export default function AuditLogs() {
  const { isAdminMode } = useAdmin()
  const { setItems } = useBreadcrumb()

  const [logs, setLogs] = useState<AuditLog[]>([])
  const [stats, setStats] = useState<AuditLogStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null)

  // Filters
  const [actionFilter, setActionFilter] = useState<string>('')
  const [resourceTypeFilter, setResourceTypeFilter] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')
  const [offset, setOffset] = useState(0)

  useEffect(() => {
    setItems([
      { label: 'Home', href: '/' },
      { label: 'Administration', href: '/admin/users' },
      { label: 'Audit Logs' },
    ])
  }, [setItems])

  const loadData = useCallback(async () => {
    if (!isAdminMode) return
    setIsLoading(true)
    try {
      const api = getApiClient()
      const filters: AuditLogFilters = {
        limit: PAGE_SIZE,
        offset,
      }
      if (actionFilter) filters.action = actionFilter
      if (resourceTypeFilter) filters.resource_type = resourceTypeFilter
      if (searchQuery) filters.search = searchQuery

      const [logsData, statsData] = await Promise.all([
        api.getAuditLogs(filters),
        api.getAuditLogStats(),
      ])
      setLogs(logsData)
      setStats(statsData)
    } catch {
      toast.error('Failed to load audit logs')
    } finally {
      setIsLoading(false)
    }
  }, [isAdminMode, actionFilter, resourceTypeFilter, searchQuery, offset])

  useEffect(() => {
    loadData()
  }, [loadData])

  useEffect(() => {
    setOffset(0)
  }, [actionFilter, resourceTypeFilter, searchQuery])

  const handleExportCSV = () => {
    if (logs.length === 0) return

    const headers = ['ID', 'Timestamp', 'Actor', 'Action', 'Resource Type', 'Resource ID', 'IP Address']
    const rows = logs.map(log => [
      log.id,
      log.timestamp,
      log.actor_name || log.actor_user_id || 'System',
      log.action,
      log.resource_type,
      log.resource_id,
      log.ip_address || '',
    ])

    const csv = [headers, ...rows]
      .map(row => row.map(v => `"${String(v).replace(/"/g, '""')}"`).join(','))
      .join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-logs-${format(new Date(), 'yyyy-MM-dd')}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <AdminGuard>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold">Audit Logs</h1>
          <Button variant="outline" size="sm" onClick={handleExportCSV} disabled={logs.length === 0}>
            <Download className="h-4 w-4 mr-1" />
            Export CSV
          </Button>
        </div>

        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Entries</CardTitle>
                <FileText className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.total_entries}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Today</CardTitle>
                <FileText className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.entries_today}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Unique Actions</CardTitle>
                <FileText className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.unique_actions}</div>
              </CardContent>
            </Card>
          </div>
        )}

        <div className="flex items-center gap-2 flex-wrap">
          <div className="relative max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search logs..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 w-64"
            />
          </div>
          <Select value={actionFilter} onValueChange={(v) => setActionFilter(v === 'all' ? '' : v)}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="All actions" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All actions</SelectItem>
              {stats?.top_actions?.map((a) => (
                <SelectItem key={a} value={a}>{a}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={resourceTypeFilter} onValueChange={(v) => setResourceTypeFilter(v === 'all' ? '' : v)}>
            <SelectTrigger className="w-40">
              <SelectValue placeholder="All resources" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All resources</SelectItem>
              {stats?.top_resource_types?.map((r) => (
                <SelectItem key={r} value={r}>{r}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Time</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Resource ID</TableHead>
                <TableHead>IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                    Loading...
                  </TableCell>
                </TableRow>
              ) : logs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                    No audit logs found
                  </TableCell>
                </TableRow>
              ) : (
                logs.map((log) => {
                  const resourceName = extractResourceName(log)
                  return (
                    <TableRow
                      key={log.id}
                      className="cursor-pointer"
                      onClick={() => setSelectedLog(log)}
                    >
                      <TableCell className="text-muted-foreground text-sm whitespace-nowrap">
                        {format(new Date(log.timestamp), 'PP p')}
                      </TableCell>
                      <TableCell className="font-medium">
                        {log.actor_name || log.actor_user_id || 'System'}
                      </TableCell>
                      <TableCell>
                        <Badge variant={actionBadgeVariant(log.action)}>{log.action}</Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground">{log.resource_type}</TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground max-w-[200px] truncate" title={log.resource_id}>
                        {resourceName || log.resource_id}
                      </TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {log.ip_address || '-'}
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>

        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {logs.length > 0 ? offset + 1 : 0}–{offset + logs.length}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={offset === 0}
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            >
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={logs.length < PAGE_SIZE}
              onClick={() => setOffset(offset + PAGE_SIZE)}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>

        <AuditLogDetailModal
          log={selectedLog}
          open={!!selectedLog}
          onOpenChange={(open) => !open && setSelectedLog(null)}
        />
      </div>
    </AdminGuard>
  )
}
