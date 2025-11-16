import { useEffect, useRef } from 'react';
import { Search } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function QueryBuilder() {
  const appJsContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (appJsContainerRef.current && window.appJs?.initQueryBuilder) {
      window.appJs.initQueryBuilder(appJsContainerRef.current);
    }
  }, []);

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-indigo-100 dark:bg-indigo-900/30">
            <Search className="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
          </div>
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Query Builder</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Build and execute queries
            </p>
          </div>
        </div>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Query Editor</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              Create and execute PromQL, LogQL, and TraceQL queries
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div 
              ref={appJsContainerRef} 
              id="app-js-query-builder-container"
              className="min-h-[400px]"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

