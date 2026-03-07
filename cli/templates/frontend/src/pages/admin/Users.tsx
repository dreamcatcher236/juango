import { useState, useEffect, useCallback, useMemo } from 'react'
import { useAdmin } from '../../contexts/AdminContext'
import { useAuth } from '../../contexts/AuthContext'
import { useBreadcrumb } from '../../contexts/BreadcrumbContext'
import { getApiClient } from '../../lib/api'
import type { UserManagement, UserStats, UpdateUserRequest } from '../../lib/types'
import { AdminGuard } from '../../components/AdminGuard'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../../components/ui/alert-dialog'
import { Label } from '../../components/ui/label'
import { Textarea } from '../../components/ui/textarea'
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
import {
  Search, Users as UsersIcon, Shield, UserCog, UserX,
  ArrowUpDown, ArrowUp, ArrowDown, ChevronLeft, ChevronRight,
  UserPlus, UserMinus, Clock,
} from 'lucide-react'

type SortField = 'name' | 'email' | 'created_at' | 'last_login'
type SortDir = 'asc' | 'desc'

export default function Users() {
  const { refreshAuth } = useAuth()
  const { isAdminMode } = useAdmin()
  const { setItems } = useBreadcrumb()

  const [users, setUsers] = useState<UserManagement[]>([])
  const [stats, setStats] = useState<UserStats | null>(null)
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)

  // Sorting
  const [sortField, setSortField] = useState<SortField>('created_at')
  const [sortDir, setSortDir] = useState<SortDir>('desc')

  // Pagination
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(25)

  // Impersonation dialog
  const [impersonateUser, setImpersonateUser] = useState<UserManagement | null>(null)
  const [impersonateReason, setImpersonateReason] = useState('')

  // Edit dialog
  const [editUser, setEditUser] = useState<UserManagement | null>(null)
  const [editIsAdmin, setEditIsAdmin] = useState(false)

  // Deactivation confirmation
  const [deactivateUser, setDeactivateUser] = useState<UserManagement | null>(null)

  useEffect(() => {
    setItems([
      { label: 'Home', href: '/' },
      { label: 'Administration', href: '/admin/users' },
      { label: 'Users' },
    ])
  }, [setItems])

  const loadData = useCallback(async () => {
    if (!isAdminMode) return
    setIsLoading(true)
    try {
      const api = getApiClient()
      const [usersData, statsData] = await Promise.all([
        api.getUsers(search || undefined),
        api.getUserStats(),
      ])
      setUsers(usersData)
      setStats(statsData)
    } catch {
      toast.error('Failed to load users')
    } finally {
      setIsLoading(false)
    }
  }, [isAdminMode, search])

  useEffect(() => {
    loadData()
  }, [loadData])

  useEffect(() => {
    setPage(0)
  }, [search, pageSize])

  const sortedUsers = useMemo(() => {
    const sorted = [...users].sort((a, b) => {
      let cmp = 0
      switch (sortField) {
        case 'name':
          cmp = (a.display_name || a.name).localeCompare(b.display_name || b.name)
          break
        case 'email':
          cmp = a.email.localeCompare(b.email)
          break
        case 'created_at':
          cmp = new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
          break
        case 'last_login':
          cmp = (a.last_login ? new Date(a.last_login).getTime() : 0) - (b.last_login ? new Date(b.last_login).getTime() : 0)
          break
      }
      return sortDir === 'asc' ? cmp : -cmp
    })
    return sorted
  }, [users, sortField, sortDir])

  const paginatedUsers = useMemo(() => {
    const start = page * pageSize
    return sortedUsers.slice(start, start + pageSize)
  }, [sortedUsers, page, pageSize])

  const totalPages = Math.ceil(sortedUsers.length / pageSize)

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDir('asc')
    }
  }

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return <ArrowUpDown className="h-3 w-3 ml-1 opacity-40" />
    return sortDir === 'asc'
      ? <ArrowUp className="h-3 w-3 ml-1" />
      : <ArrowDown className="h-3 w-3 ml-1" />
  }

  const handleImpersonate = async () => {
    if (!impersonateUser || !impersonateReason.trim()) return
    try {
      await getApiClient().startImpersonation(impersonateUser.id, impersonateReason.trim())
      toast.success(`Now impersonating ${impersonateUser.name}`)
      setImpersonateUser(null)
      setImpersonateReason('')
      await refreshAuth()
    } catch {
      toast.error('Failed to start impersonation')
    }
  }

  const handleUpdateUser = async () => {
    if (!editUser) return
    try {
      const req: UpdateUserRequest = { is_admin: editIsAdmin }
      await getApiClient().updateUser(editUser.id, req)
      toast.success('User updated')
      setEditUser(null)
      loadData()
    } catch {
      toast.error('Failed to update user')
    }
  }

  const handleDeactivate = async () => {
    if (!deactivateUser) return
    try {
      await getApiClient().updateUser(deactivateUser.id, { active: false })
      toast.success(`User ${deactivateUser.email} deactivated`)
      setDeactivateUser(null)
      loadData()
    } catch {
      toast.error('Failed to deactivate user')
    }
  }

  return (
    <AdminGuard>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold">Users</h1>
        </div>

        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Users</CardTitle>
                <UsersIcon className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.total_users}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Admins</CardTitle>
                <Shield className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.admin_users}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Active (24h)</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.active_users_24h}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Active (7d)</CardTitle>
                <UsersIcon className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.active_users_7d}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">New (7d)</CardTitle>
                <UserPlus className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.new_users_7d}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Inactive</CardTitle>
                <UserMinus className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.inactive_users}</div>
              </CardContent>
            </Card>
          </div>
        )}

        <div className="flex items-center gap-2">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search users..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8"
            />
          </div>
        </div>

        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>
                  <button className="flex items-center" onClick={() => toggleSort('name')}>
                    User <SortIcon field="name" />
                  </button>
                </TableHead>
                <TableHead>
                  <button className="flex items-center" onClick={() => toggleSort('email')}>
                    Email <SortIcon field="email" />
                  </button>
                </TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>
                  <button className="flex items-center" onClick={() => toggleSort('last_login')}>
                    Last Login <SortIcon field="last_login" />
                  </button>
                </TableHead>
                <TableHead>
                  <button className="flex items-center" onClick={() => toggleSort('created_at')}>
                    Created <SortIcon field="created_at" />
                  </button>
                </TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                    Loading...
                  </TableCell>
                </TableRow>
              ) : paginatedUsers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                    No users found
                  </TableCell>
                </TableRow>
              ) : (
                paginatedUsers.map((u) => {
                  const isActive = !u.modified_at || !('deleted_at' in u) || !(u as Record<string, unknown>).deleted_at
                  return (
                    <TableRow key={u.id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {u.profile_pic_url ? (
                            <img src={u.profile_pic_url} alt="" className="h-6 w-6 rounded-full" />
                          ) : (
                            <div className="h-6 w-6 rounded-full bg-muted flex items-center justify-center text-xs">
                              {(u.display_name || u.name)?.[0]?.toUpperCase()}
                            </div>
                          )}
                          <span className="font-medium">{u.display_name || u.name}</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-muted-foreground">{u.email}</TableCell>
                      <TableCell>
                        {u.is_admin ? (
                          <Badge variant="default">Admin</Badge>
                        ) : (
                          <Badge variant="secondary">User</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        {isActive ? (
                          <Badge variant="outline" className="border-green-500 text-green-700 dark:text-green-400">Active</Badge>
                        ) : (
                          <Badge variant="outline" className="border-red-500 text-red-700 dark:text-red-400">Inactive</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {u.last_login ? format(new Date(u.last_login), 'PP') : 'Never'}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {format(new Date(u.created_at), 'PP')}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              setEditUser(u)
                              setEditIsAdmin(u.is_admin)
                            }}
                            title="Edit user"
                          >
                            <UserCog className="h-4 w-4" />
                          </Button>
                          {!u.is_admin && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setImpersonateUser(u)}
                              title="Impersonate user"
                            >
                              <UserX className="h-4 w-4" />
                            </Button>
                          )}
                          {isActive && !u.is_admin && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setDeactivateUser(u)}
                              title="Deactivate user"
                              className="text-destructive hover:text-destructive"
                            >
                              <UserMinus className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>

        {/* Pagination */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>Rows per page:</span>
            <Select value={pageSize.toString()} onValueChange={(v) => setPageSize(Number(v))}>
              <SelectTrigger className="w-16 h-8">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="10">10</SelectItem>
                <SelectItem value="25">25</SelectItem>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
              </SelectContent>
            </Select>
            <span>
              {sortedUsers.length > 0
                ? `${page * pageSize + 1}-${Math.min((page + 1) * pageSize, sortedUsers.length)} of ${sortedUsers.length}`
                : '0 results'}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page === 0}
              onClick={() => setPage(p => p - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= totalPages - 1}
              onClick={() => setPage(p => p + 1)}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Impersonation Dialog */}
        <Dialog open={!!impersonateUser} onOpenChange={(open) => !open && setImpersonateUser(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Impersonate User</DialogTitle>
              <DialogDescription>
                You will see the application as <strong>{impersonateUser?.display_name || impersonateUser?.name}</strong> ({impersonateUser?.email}).
                This action is logged.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2">
              <Label htmlFor="impersonate-reason">Reason</Label>
              <Textarea
                id="impersonate-reason"
                placeholder="e.g., Investigating support ticket #1234"
                value={impersonateReason}
                onChange={(e) => setImpersonateReason(e.target.value)}
              />
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setImpersonateUser(null)}>Cancel</Button>
              <Button onClick={handleImpersonate} disabled={!impersonateReason.trim()}>
                Start Impersonation
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Edit User Dialog */}
        <Dialog open={!!editUser} onOpenChange={(open) => !open && setEditUser(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit User</DialogTitle>
              <DialogDescription>
                Update settings for {editUser?.display_name || editUser?.name}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <Label>Administrator</Label>
                <Button
                  variant={editIsAdmin ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setEditIsAdmin(!editIsAdmin)}
                >
                  {editIsAdmin ? 'Admin' : 'User'}
                </Button>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setEditUser(null)}>Cancel</Button>
              <Button onClick={handleUpdateUser}>Save</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Deactivation Confirmation */}
        <AlertDialog open={!!deactivateUser} onOpenChange={(open) => !open && setDeactivateUser(null)}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Deactivate User</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to deactivate <strong>{deactivateUser?.display_name || deactivateUser?.name}</strong> ({deactivateUser?.email})?
                They will no longer be able to access the application.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={handleDeactivate} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                Deactivate
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </AdminGuard>
  )
}
