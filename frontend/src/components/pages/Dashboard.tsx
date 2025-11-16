import { useEffect, useRef } from 'react';
import { BarChart3, GitBranch, FileText, RefreshCw } from 'lucide-react';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function Dashboard() {
  const appJsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Placeholder for app.js integration
    // The container div can be targeted by app.js for rendering
    if (appJsContainerRef.current && window.appJs?.initDashboard) {
      window.appJs.initDashboard(appJsContainerRef.current);
    }
  }, []);

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Dashboard</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Overview of your observability data
            </p>
          </div>
          <Button>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>

        {/* Welcome Section */}
        <Card className="border-slate-200 dark:border-slate-800 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">
              Welcome to Open-Telemorph-Prime
            </CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              A simplified observability platform for home users and developers.
            </CardDescription>
          </CardHeader>
        </Card>

        {/* Feature Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card className="border-slate-200 dark:border-slate-800">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/30">
                  <BarChart3 className="w-6 h-6 text-blue-600 dark:text-blue-400" />
                </div>
                <CardTitle className="text-slate-900 dark:text-slate-100">Metrics</CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              <CardDescription className="text-slate-600 dark:text-slate-400">
                View and query your application metrics with PromQL-like syntax.
              </CardDescription>
            </CardContent>
          </Card>

          <Card className="border-slate-200 dark:border-slate-800">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-purple-100 dark:bg-purple-900/30">
                  <GitBranch className="w-6 h-6 text-purple-600 dark:text-purple-400" />
                </div>
                <CardTitle className="text-slate-900 dark:text-slate-100">Traces</CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              <CardDescription className="text-slate-600 dark:text-slate-400">
                Trace requests across your distributed systems and identify bottlenecks.
              </CardDescription>
            </CardContent>
          </Card>

          <Card className="border-slate-200 dark:border-slate-800">
            <CardHeader>
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-green-100 dark:bg-green-900/30">
                  <FileText className="w-6 h-6 text-green-600 dark:text-green-400" />
                </div>
                <CardTitle className="text-slate-900 dark:text-slate-100">Logs</CardTitle>
              </div>
            </CardHeader>
            <CardContent>
              <CardDescription className="text-slate-600 dark:text-slate-400">
                Search and filter through your application logs with powerful queries.
              </CardDescription>
            </CardContent>
          </Card>
        </div>

        {/* Quick Start */}
        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Quick Start</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              Send telemetry data to your Open-Telemorph-Prime instance:
            </CardDescription>
          </CardHeader>
          <CardContent>
            <pre className="bg-slate-900 dark:bg-slate-950 text-slate-100 p-4 rounded-lg overflow-x-auto">
              <code>{`curl -X POST http://localhost:8080/v1/traces \\
  -H "Content-Type: application/json" \\
  -d '{"resourceSpans": [...]}'`}</code>
            </pre>
          </CardContent>
        </Card>

        {/* App.js Integration Point */}
        <div 
          ref={appJsContainerRef} 
          id="app-js-dashboard-container"
          className="min-h-[200px]"
        />
      </div>
    </div>
  );
}

// Extend window interface for TypeScript
declare global {
  interface Window {
    appJs?: {
      initDashboard?: (container: HTMLElement) => void;
      initMetrics?: (container: HTMLElement) => void;
      initTraces?: (container: HTMLElement) => void;
      initLogs?: (container: HTMLElement) => void;
      initServices?: (container: HTMLElement) => void;
      initAlerts?: (container: HTMLElement) => void;
      initQueryBuilder?: (container: HTMLElement) => void;
    };
  }
}

