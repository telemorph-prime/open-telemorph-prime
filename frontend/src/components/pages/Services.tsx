import { useEffect, useRef } from 'react';
import { Server } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function Services() {
  const appJsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (appJsContainerRef.current && window.appJs?.initServices) {
      window.appJs.initServices(appJsContainerRef.current);
    }
  }, []);

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-orange-100 dark:bg-orange-900/30">
            <Server className="w-6 h-6 text-orange-600 dark:text-orange-400" />
          </div>
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Services</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Monitor service health and dependencies
            </p>
          </div>
        </div>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Service Overview</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              View all services and their status
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div 
              ref={appJsContainerRef} 
              id="app-js-services-container"
              className="min-h-[400px]"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

