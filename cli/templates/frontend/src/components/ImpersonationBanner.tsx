import { useAuth } from '../contexts/AuthContext'
import { getApiClient } from '../lib/api'
import { Button } from './ui/button'
import { UserX, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'

export function ImpersonationBanner() {
  const { session, refreshAuth } = useAuth()

  if (!session?.impersonation?.enabled) return null

  const { target_user_email, target_user_name, reason } = session.impersonation

  const handleStop = async () => {
    try {
      await getApiClient().stopImpersonation()
      toast.success('Impersonation stopped')
      await refreshAuth()
    } catch {
      toast.error('Failed to stop impersonation')
    }
  }

  return (
    <div className="bg-yellow-500 text-yellow-950 px-4 py-2 flex items-center justify-between text-sm">
      <div className="flex items-center gap-3">
        <UserX className="h-4 w-4 flex-shrink-0" />
        <div className="flex flex-col sm:flex-row sm:items-center sm:gap-2">
          <span>
            Impersonating <strong>{target_user_name || target_user_email}</strong>
          </span>
          {reason && (
            <span className="text-yellow-800 text-xs">
              Reason: {reason}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1 text-yellow-800 text-xs">
          <AlertTriangle className="h-3 w-3" />
          <span>All actions are being logged</span>
        </div>
      </div>
      <Button
        variant="outline"
        size="sm"
        onClick={handleStop}
        className="bg-yellow-600 border-yellow-700 text-yellow-50 hover:bg-yellow-700 hover:text-yellow-50 h-7"
      >
        Stop Impersonation
      </Button>
    </div>
  )
}
