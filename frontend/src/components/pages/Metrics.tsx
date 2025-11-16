import { useEffect, useRef } from 'react';
import { BarChart3 } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function Metrics() {
  const appJsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Placeholder for app.js integration
    if (appJsContainerRef.current && window.appJs?.initMetrics) {
      window.appJs.initMetrics(appJsContainerRef.current);
    }
  }, []);

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-blue-100 dark:bg-blue-900/30">
            <BarChart3 className="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Metrics</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Monitor and analyze application metrics
            </p>
          </div>
        </div>

        {/* Content Card */}
        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Metrics Explorer</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              Query and visualize your metrics data
            </CardDescription>
          </CardHeader>
          <CardContent>
            {/* App.js Integration Point */}
            <div 
              ref={appJsContainerRef} 
              id="app-js-metrics-container"
              className="min-h-[400px]"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

