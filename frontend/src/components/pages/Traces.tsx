import { useEffect, useRef } from 'react';
import { GitBranch } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function Traces() {
  const appJsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (appJsContainerRef.current && window.appJs?.initTraces) {
      window.appJs.initTraces(appJsContainerRef.current);
    }
  }, []);

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-purple-100 dark:bg-purple-900/30">
            <GitBranch className="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Traces</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Explore distributed traces
            </p>
          </div>
        </div>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Trace Explorer</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              View and analyze trace data
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div 
              ref={appJsContainerRef} 
              id="app-js-traces-container"
              className="min-h-[400px]"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

