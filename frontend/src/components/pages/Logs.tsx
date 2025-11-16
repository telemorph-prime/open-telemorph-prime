import { useEffect, useRef } from 'react';
import { FileText } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function Logs() {
  const appJsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (appJsContainerRef.current && window.appJs?.initLogs) {
      window.appJs.initLogs(appJsContainerRef.current);
    }
  }, []);

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-green-100 dark:bg-green-900/30">
            <FileText className="w-6 h-6 text-green-600 dark:text-green-400" />
          </div>
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Logs</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Search and filter application logs
            </p>
          </div>
        </div>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Log Viewer</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              Browse and search log entries
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div 
              ref={appJsContainerRef} 
              id="app-js-logs-container"
              className="min-h-[400px]"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

