# go-lightning Documentation

Nextra-based documentation for go-lightning library.

## Development

### Prerequisites

- Node.js 18+
- npm or yarn

### Install Dependencies

```bash
npm install
```

### Run Development Server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build for Production

```bash
npm run build
```

### Export Static Site

```bash
npm run export
```

This generates static HTML in the `out/` directory.

## Deployment

### GitHub Pages

#### Option 1: Manual Deployment

1. Build and export:
   ```bash
   npm run export
   ```

2. Copy `out/` contents to `gh-pages` branch:
   ```bash
   git checkout --orphan gh-pages
   cp -r out/* .
   git add .
   git commit -m "Deploy documentation"
   git push origin gh-pages
   ```

3. In GitHub repository settings:
   - Go to Settings → Pages
   - Set source to `gh-pages` branch
   - Save

#### Option 2: GitHub Actions (Recommended)

Create `.github/workflows/deploy-docs.yml`:

```yaml
name: Deploy Documentation

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-node@v3
        with:
          node-version: 18

      - name: Install dependencies
        run: cd docs && npm install

      - name: Build and export
        run: cd docs && npm run export

      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs/out
```

#### Base Path for Subdirectory

If hosting at `https://user.github.io/go-lightning/`:

```bash
BASE_PATH=/go-lightning npm run export
```

### Netlify

1. Connect repository to Netlify
2. Build command: `cd docs && npm run export`
3. Publish directory: `docs/out`

### Vercel

1. Connect repository to Vercel
2. Root directory: `docs`
3. Build command: `npm run export`
4. Output directory: `out`

## Project Structure

```
docs/
├── package.json          # Dependencies
├── next.config.js        # Next.js configuration
├── theme.config.jsx      # Nextra theme (dark mode)
├── tsconfig.json         # TypeScript config
├── .gitignore           # Git ignore rules
└── pages/               # Documentation pages
    ├── index.mdx        # Landing page
    ├── getting-started/ # Getting Started section
    ├── guides/          # In-depth guides
    ├── api-reference/   # API documentation
    ├── examples/        # Code examples
    ├── troubleshooting/ # Common issues
    └── contributing/    # Development guide
```

## Writing Documentation

### MDX Files

All pages are written in MDX (Markdown + JSX):

```mdx
# Page Title

Regular markdown content.

## Code Example

\`\`\`go
func main() {
    fmt.Println("Hello")
}
\`\`\`

## See Also

- [Related Page](/path/to/page)
```

### Navigation

Update `_meta.json` files to control sidebar navigation:

```json
{
  "index": "Home",
  "getting-started": "Getting Started",
  "guides": "Guides"
}
```

## Features

- ✅ **Dark mode by default**
- ✅ **Full-text search** (flexsearch)
- ✅ **Syntax highlighting** for Go and SQL
- ✅ **Mobile responsive**
- ✅ **Static export** for GitHub Pages
- ✅ **Auto-generated navigation**
- ✅ **Table of contents** on each page

## Troubleshooting

### "Module not found"

```bash
rm -rf node_modules package-lock.json
npm install
```

### Build fails

Check Node.js version:
```bash
node --version  # Should be 18+
```

### Search not working

Rebuild the site:
```bash
rm -rf .next out
npm run build
```

## License

MIT
