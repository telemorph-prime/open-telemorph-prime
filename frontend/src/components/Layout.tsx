import { NavLink } from 'react-router-dom';
import { 
  LayoutDashboard, 
  BarChart3, 
  GitBranch, 
  FileText, 
  Server, 
  Bell, 
  Search,
  Settings as SettingsIcon
} from 'lucide-react';
import logo from 'figma:asset/ef5a899a28876d2f687ea79fba14eb4f5cd1ce63.png';
import { ImageWithFallback } from './figma/ImageWithFallback';

const navItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/metrics', icon: BarChart3, label: 'Metrics' },
  { to: '/traces', icon: GitBranch, label: 'Traces' },
  { to: '/logs', icon: FileText, label: 'Logs' },
  { to: '/services', icon: Server, label: 'Services' },
  { to: '/alerts', icon: Bell, label: 'Alerts' },
  { to: '/query-builder', icon: Search, label: 'Query Builder' },
  { to: '/settings', icon: SettingsIcon, label: 'Settings' },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-screen bg-slate-50 dark:bg-slate-950">
      {/* Sidebar */}
      <aside className="w-64 bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800 flex flex-col">
        {/* Logo */}
        <div className="p-6 border-b border-slate-200 dark:border-slate-800">
          <ImageWithFallback 
            src={logo} 
            alt="Telemorph Prime" 
            className="w-32 h-auto"
          />
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 overflow-y-auto">
          <ul className="space-y-1">
            {navItems.map((item) => (
              <li key={item.to}>
                <NavLink
                  to={item.to}
                  className={({ isActive }) =>
                    `flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                      isActive
                        ? 'bg-blue-50 dark:bg-blue-950 text-blue-600 dark:text-blue-400'
                        : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
                    }`
                  }
                >
                  <item.icon className="w-5 h-5" />
                  <span>{item.label}</span>
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-slate-200 dark:border-slate-800">
          <div className="text-xs text-slate-500 dark:text-slate-400">
            Phase 4 Frontend<br />v1.0.1
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto">
        {children}
      </main>
    </div>
  );
}

