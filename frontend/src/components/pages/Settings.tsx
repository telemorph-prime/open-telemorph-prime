import { Settings as SettingsIcon, Moon, Sun } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Label } from '../ui/label';
import { Switch } from '../ui/switch';
import { useTheme } from '../ThemeProvider';

export function Settings() {
  const { theme, toggleTheme } = useTheme();

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-slate-200 dark:bg-slate-800">
            <SettingsIcon className="w-6 h-6 text-slate-600 dark:text-slate-400" />
          </div>
          <div>
            <h1 className="text-slate-900 dark:text-slate-100">Settings</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Configure your application preferences
            </p>
          </div>
        </div>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Appearance</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              Customize the visual appearance of the application
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                {theme === 'dark' ? (
                  <Moon className="w-5 h-5 text-slate-600 dark:text-slate-400" />
                ) : (
                  <Sun className="w-5 h-5 text-slate-600 dark:text-slate-400" />
                )}
                <div>
                  <Label htmlFor="theme-toggle" className="text-slate-900 dark:text-slate-100">
                    Dark Mode
                  </Label>
                  <p className="text-slate-600 dark:text-slate-400">
                    Toggle between light and dark theme
                  </p>
                </div>
              </div>
              <Switch
                id="theme-toggle"
                checked={theme === 'dark'}
                onCheckedChange={toggleTheme}
              />
            </div>
          </CardContent>
        </Card>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Data & Storage</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              Manage data retention and storage settings
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-slate-600 dark:text-slate-400">
              Data storage settings are managed by the backend server.
            </p>
          </CardContent>
        </Card>

        <Card className="border-slate-200 dark:border-slate-800">
          <CardHeader>
            <CardTitle className="text-slate-900 dark:text-slate-100">Integrations</CardTitle>
            <CardDescription className="text-slate-600 dark:text-slate-400">
              Configure external integrations and data sources
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-slate-600 dark:text-slate-400">
              Integration settings will be available in future updates.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

