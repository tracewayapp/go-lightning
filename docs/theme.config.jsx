export default {
  logo: <span style={{ fontWeight: 700, fontSize: '1.2rem' }}>⚡ go-lightning</span>,
  project: {
    link: 'https://github.com/tracewayapp/go-lightning'
  },
  docsRepositoryBase: 'https://github.com/tracewayapp/go-lightning/tree/main/docs',
  useNextSeoProps() {
    return {
      titleTemplate: '%s – go-lightning'
    }
  },
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta property="og:title" content="go-lightning" />
      <meta property="og:description" content="Lightweight Go library for simplified database operations" />
    </>
  ),
  primaryHue: 205,
  darkMode: true,
  nextThemes: {
    defaultTheme: 'dark'
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true
  },
  toc: {
    backToTop: true
  },
  editLink: {
    text: 'Edit this page on GitHub →'
  },
  feedback: {
    content: 'Question? Give us feedback →',
    labels: 'feedback'
  },
  footer: {
    text: (
      <span>
        MIT {new Date().getFullYear()} ©{' '}
        <a href="https://github.com/tracewayapp" target="_blank">
          Traceway
        </a>
      </span>
    )
  }
}
