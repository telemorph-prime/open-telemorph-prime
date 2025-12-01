# Frontend

React-based frontend for Open-Telemorph-Prime.

## Development

```bash
# Install dependencies
npm install

# Start development server (with hot reload)
npm run dev

# Build for production
npm run build
```

The build output goes to `dist/` which is embedded into the Go binary during the backend build process.

## Structure

- `public/` - Public static files (index.html, favicon, etc.)
- `src/` - React source code
  - `components/` - React components
    - `pages/` - Page components
    - `ui/` - UI component library (shadcn/ui style)
  - `assets/` - Static assets (images, etc.)
  - `styles/` - Global styles
- `package.json` - Node.js dependencies
- `vite.config.ts` - Vite configuration
- `tailwind.config.js` - Tailwind CSS configuration


