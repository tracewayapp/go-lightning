const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.jsx',
  defaultShowCopyCode: true,
  flexsearch: {
    codeblocks: true
  }
})

module.exports = withNextra({
  output: 'export',
  images: {
    unoptimized: true
  },
  basePath: '/go-lightning',
  assetPrefix: '/go-lightning/',
  trailingSlash: true
})
