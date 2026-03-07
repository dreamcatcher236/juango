import { useState } from 'react'
import { useAdmin } from '../contexts/AdminContext'
import { Button } from './ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './ui/dialog'
import { Label } from './ui/label'
import { Textarea } from './ui/textarea'
import { Alert, AlertDescription } from './ui/alert'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from './ui/tooltip'
import { Shield, ShieldOff, Clock, AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'
import { formatDistanceToNow } from 'date-fns'

export function AdminModeToggle() {
  const { isAdminMode, adminModeState, enableAdminMode, disableAdminMode, isLoading } = useAdmin()
  const [showReasonDialog, setShowReasonDialog] = useState(false)
  const [reason, setReason] = useState('')

  const handleToggle = async () => {
    if (isAdminMode) {
      try {
        await disableAdminMode()
        toast.success('Admin mode disabled')
      } catch {
        toast.error('Failed to disable admin mode')
      }
    } else {
      setShowReasonDialog(true)
    }
  }

  const handleEnable = async () => {
    if (!reason.trim()) return
    try {
      await enableAdminMode(reason.trim())
      toast.success('Admin mode enabled')
      setShowReasonDialog(false)
      setReason('')
    } catch {
      toast.error('Failed to enable admin mode')
    }
  }

  const duration = adminModeState?.since
    ? formatDistanceToNow(new Date(adminModeState.since), { addSuffix: false })
    : null

  const toggleButton = (
    <Button
      variant={isAdminMode ? 'destructive' : 'outline'}
      size="sm"
      onClick={handleToggle}
      disabled={isLoading}
      className="gap-1.5"
    >
      {isAdminMode ? (
        <>
          <ShieldOff className="h-4 w-4" />
          <span>Admin Mode</span>
          <span className="flex items-center gap-1 text-xs opacity-80">
            <Clock className="h-3 w-3" />
            {duration}
          </span>
        </>
      ) : (
        <>
          <Shield className="h-4 w-4" />
          Enable Admin Mode
        </>
      )}
    </Button>
  )

  return (
    <>
      {isAdminMode && adminModeState ? (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              {toggleButton}
            </TooltipTrigger>
            <TooltipContent side="bottom" className="max-w-xs">
              <div className="space-y-1 text-xs">
                <p><span className="font-semibold">Reason:</span> {adminModeState.reason}</p>
                <p><span className="font-semibold">Since:</span> {new Date(adminModeState.since).toLocaleString()}</p>
                <p><span className="font-semibold">Duration:</span> {duration}</p>
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      ) : (
        toggleButton
      )}

      <Dialog open={showReasonDialog} onOpenChange={setShowReasonDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Enable Admin Mode</DialogTitle>
            <DialogDescription>
              Provide a reason for enabling admin mode. This will be logged for audit purposes.
            </DialogDescription>
          </DialogHeader>
          <Alert className="border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-950">
            <AlertTriangle className="h-4 w-4 text-amber-600" />
            <AlertDescription className="text-amber-800 dark:text-amber-200">
              Admin mode grants elevated privileges. All actions will be audited. The session will expire after the configured timeout.
            </AlertDescription>
          </Alert>
          <div className="space-y-2">
            <Label htmlFor="reason">Reason</Label>
            <Textarea
              id="reason"
              placeholder="e.g., Investigating user report #1234"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowReasonDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleEnable} disabled={!reason.trim() || isLoading}>
              Enable
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
