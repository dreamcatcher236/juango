import { useState, useEffect } from 'react'
import { useBreadcrumb } from '../../contexts/BreadcrumbContext'
import { getApiClient } from '../../lib/api'
import type { TaskResponse } from '../../lib/types'
import { AdminGuard } from '../../components/AdminGuard'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { Label } from '../../components/ui/label'
import { Alert, AlertDescription, AlertTitle } from '../../components/ui/alert'
import { CheckCircle2, XCircle, Loader2, Play, Info } from 'lucide-react'

function TasksContent() {
  const { setItems } = useBreadcrumb()
  const [message, setMessage] = useState('Hello from the admin panel!')
  const [count, setCount] = useState(5)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TaskResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setItems([
      { label: 'Admin', href: '/' },
      { label: 'Tasks' },
    ])
    return () => setItems([])
  }, [setItems])

  const handleEnqueue = async () => {
    setLoading(true)
    setResult(null)
    setError(null)

    try {
      const api = getApiClient()
      const response = await api.enqueueDummyTask(message, count)
      setResult(response)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to enqueue task')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Tasks</h1>
        <p className="text-muted-foreground">Manage and test background tasks</p>
      </div>

      {/* Test Task Card */}
      <div className="rounded-lg border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">Test Task</h2>
        <p className="text-sm text-muted-foreground mb-4">
          Enqueue a dummy task to test the background worker. The task will log messages
          for the specified number of seconds.
        </p>

        <div className="space-y-4 max-w-md">
          <div className="space-y-2">
            <Label htmlFor="message">Message</Label>
            <Input
              id="message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Enter a test message"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="count">Duration (seconds): {count}</Label>
            <Input
              id="count"
              type="range"
              min={1}
              max={60}
              value={count}
              onChange={(e) => setCount(parseInt(e.target.value))}
              className="cursor-pointer"
            />
            <p className="text-xs text-muted-foreground">
              The worker will process this task for {count} second{count !== 1 ? 's' : ''}
            </p>
          </div>

          <Button onClick={handleEnqueue} disabled={loading || !message.trim()}>
            {loading ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Play className="mr-2 h-4 w-4" />
            )}
            Enqueue Task
          </Button>
        </div>

        {result && (
          <Alert className="mt-4 max-w-md border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950">
            <CheckCircle2 className="h-4 w-4 text-green-600" />
            <AlertTitle>Task Enqueued</AlertTitle>
            <AlertDescription>
              <div className="text-sm space-y-1 mt-1">
                <div><span className="font-medium">Task ID:</span> {result.task_id}</div>
                <div><span className="font-medium">Queue:</span> {result.queue}</div>
                <div><span className="font-medium">Message:</span> {result.message}</div>
                <div><span className="font-medium">Duration:</span> {result.count}s</div>
              </div>
            </AlertDescription>
          </Alert>
        )}

        {error && (
          <Alert className="mt-4 max-w-md" variant="destructive">
            <XCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </div>

      {/* Info Card */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-start gap-3">
          <Info className="h-5 w-5 text-blue-500 mt-0.5" />
          <div>
            <h3 className="font-semibold">Background Worker</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Tasks are processed by a background worker that connects to Redis.
              Make sure the worker is running:
            </p>
            <pre className="mt-2 rounded bg-muted p-3 text-sm font-mono">
              ./app worker -c config.yml
            </pre>
            <p className="text-sm text-muted-foreground mt-2">
              Or use Docker Compose which starts both the server and worker:
            </p>
            <pre className="mt-2 rounded bg-muted p-3 text-sm font-mono">
              docker compose up
            </pre>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function Tasks() {
  return (
    <AdminGuard>
      <TasksContent />
    </AdminGuard>
  )
}
