import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './components/Layout';
import { Dashboard } from './components/pages/Dashboard';
import { Metrics } from './components/pages/Metrics';
import { Traces } from './components/pages/Traces';
import { Logs } from './components/pages/Logs';
import { Services } from './components/pages/Services';
import { Alerts } from './components/pages/Alerts';
import { QueryBuilder } from './components/pages/QueryBuilder';
import { Settings } from './components/pages/Settings';
import { ThemeProvider } from './components/ThemeProvider';

export default function App() {
  return (
    <ThemeProvider>
      <Router>
        <Layout>
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/metrics" element={<Metrics />} />
            <Route path="/traces" element={<Traces />} />
            <Route path="/logs" element={<Logs />} />
            <Route path="/services" element={<Services />} />
            <Route path="/alerts" element={<Alerts />} />
            <Route path="/query-builder" element={<QueryBuilder />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </Layout>
      </Router>
    </ThemeProvider>
  );
}

