import { type ReactNode } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useAdmin } from '../contexts/AdminContext'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { ShieldAlert, ShieldOff, Shield } from 'lucide-react'

interface AdminGuardProps {
  children: ReactNode
}

export function AdminGuard({ children }: AdminGuardProps) {
  const { user, isLoading: authLoading } = useAuth()
  const { isAdminMode, isLoading: adminLoading } = useAdmin()

  if (authLoading || adminLoading) {
    return (
      <div className="flex items-center justify-center min-h-[50vh]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!user?.is_admin) {
    return (
      <div className="flex items-center justify-center min-h-[50vh]">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-destructive">
              <ShieldAlert className="h-5 w-5" />
              Access Denied
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">
              You do not have administrator privileges. Contact an existing admin to grant you access.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!isAdminMode) {
    return (
      <div className="flex items-center justify-center min-h-[50vh]">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ShieldOff className="h-5 w-5" />
              Admin Mode Required
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-muted-foreground">
              You need to enable admin mode to access this page.
            </p>
            <ol className="list-decimal list-inside text-sm text-muted-foreground space-y-1">
              <li>Click the <Shield className="inline h-3.5 w-3.5 mx-0.5" /> <strong>Enable Admin Mode</strong> button in the header</li>
              <li>Provide a reason for enabling admin mode</li>
              <li>Click <strong>Enable</strong> to activate</li>
            </ol>
          </CardContent>
        </Card>
      </div>
    )
  }

  return <>{children}</>
}
