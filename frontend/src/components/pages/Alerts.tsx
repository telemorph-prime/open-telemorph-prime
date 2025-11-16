import { useEffect, useRef } from 'react';
import { Bell } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function Alerts() {
  const appJsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (appJsContainerRef.current && window.appJs?.initAlerts) {
      window.appJs.initAlerts(appJsContainerRef.current);
    }
  }, []);

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-red-100 dark:bg-red-900/30">
            <Bell className="w-6 h-6 text-red-600 dark:text-red-400" />
          </div>
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Alerts</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Manage and monitor alerts
            </p>
          </div>
        </div>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Alert Management</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              View and manage alert rules
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div 
              ref={appJsContainerRef} 
              id="app-js-alerts-container"
              className="min-h-[400px]"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

