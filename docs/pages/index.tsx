import Head from 'next/head'
import Link from 'next/link'
import s from '../styles/home.module.css'

const features = [
  {
    title: 'SQL-First',
    desc: 'Write actual SQL queries. No DSLs, no query builders, no magic. Full control over every database operation.',
  },
  {
    title: 'Cached Queries',
    desc: 'INSERT and UPDATE statements are pre-computed at registration. Zero runtime overhead for query generation.',
  },
  {
    title: 'Zero Code Generation',
    desc: 'No build steps. No generated files to maintain. Import the library and start building immediately.',
  },
  {
    title: 'Lightweight',
    desc: 'Only two dependencies: database/sql and uuid. Small binary footprint and fast compilation times.',
  },
  {
    title: 'Type Safe',
    desc: 'Go generics provide compile-time type safety for all your database operations.',
  },
  {
    title: 'Flexible Mapping',
    desc: 'Map query results to any struct. Use DTOs for JOINs, aggregations, and projections.',
  },
]

function CodeExample({ styles }: { styles: typeof s }) {
  return (
    <pre>
      <span className={styles.kw}>type</span>{' '}
      <span className={styles.ty}>User</span>{' '}
      <span className={styles.kw}>struct</span>
      {' {\n'}
      {'    Id        '}
      <span className={styles.ty}>int</span>
      {'\n'}
      {'    FirstName '}
      <span className={styles.ty}>string</span>
      {'\n'}
      {'    LastName  '}
      <span className={styles.ty}>string</span>
      {'\n'}
      {'    Email     '}
      <span className={styles.ty}>string</span>
      {'\n}\n\n'}
      <span className={styles.kw}>func</span>
      {' main() {\n'}
      {'    '}
      <span className={styles.cm}>{'// Register once at startup'}</span>
      {'\n'}
      {'    lit.'}
      <span className={styles.fn}>RegisterModel</span>
      {'['}
      <span className={styles.ty}>User</span>
      {'](lit.PostgreSQL)\n\n'}
      {'    db, _ := sql.'}
      <span className={styles.fn}>Open</span>
      {'('}
      <span className={styles.st}>&quot;postgres&quot;</span>
      {', connStr)\n\n'}
      {'    '}
      <span className={styles.cm}>{'// Create'}</span>
      {'\n'}
      {'    id, _ := lit.'}
      <span className={styles.fn}>Insert</span>
      {'(db, &user)\n\n'}
      {'    '}
      <span className={styles.cm}>{'// Read'}</span>
      {'\n'}
      {'    users, _ := lit.'}
      <span className={styles.fn}>Select</span>
      {'['}
      <span className={styles.ty}>User</span>
      {'](db, '}
      <span className={styles.st}>&quot;SELECT * FROM users&quot;</span>
      {')\n\n'}
      {'    '}
      <span className={styles.cm}>{'// Update'}</span>
      {'\n'}
      {'    lit.'}
      <span className={styles.fn}>Update</span>
      {'(db, &user, '}
      <span className={styles.st}>&quot;id = $1&quot;</span>
      {', user.Id)\n\n'}
      {'    '}
      <span className={styles.cm}>{'// Delete'}</span>
      {'\n'}
      {'    lit.'}
      <span className={styles.fn}>Delete</span>
      {'(db, '}
      <span className={styles.st}>&quot;DELETE FROM users WHERE id = $1&quot;</span>
      {', id)\n'}
      {'}'}
    </pre>
  )
}

export default function HomePage() {
  return (
    <>
      <Head>
        <title>lit - Lightweight Go Database Library</title>
        <meta
          name="description"
          content="A lightweight Go library that eliminates database boilerplate. Write real SQL with type safety, cached queries, and zero code generation."
        />
      </Head>

      <div className={s.page}>
        <div className={s.bgEffect} />

        <div className={s.content}>
          <nav className={s.nav}>
            <span className={s.logo}>lit</span>
            <div className={s.navLinks}>
              <Link href="/getting-started/installation" className={s.navLink}>
                Docs
              </Link>
              <a
                href="https://github.com/tracewayapp/lit"
                target="_blank"
                rel="noopener noreferrer"
                className={s.navLink}
              >
                GitHub
              </a>
            </div>
          </nav>

          <section className={s.hero}>
            <a href="https://github.com/tracewayapp/lit" target="_blank" rel="noopener noreferrer" className={s.badge}>Open Source Go ORM</a>
            <h1 className={s.title}>lit</h1>
            <p className={s.subtitle}>
              A lightweight Go library that eliminates database boilerplate.
              Write real SQL with type safety, cached queries, and zero code
              generation.
            </p>
            <div className={s.heroActions}>
              <div className={s.buttons}>
                <Link
                  href="/getting-started/installation"
                  className={s.primaryBtn}
                >
                  Get Started
                </Link>
                <a
                  href="https://github.com/tracewayapp/lit"
                  target="_blank"
                  rel="noopener noreferrer"
                  className={s.secondaryBtn}
                >
                  GitHub
                </a>
              </div>
              <code className={s.installCmd}>
                go get github.com/tracewayapp/lit/v2
              </code>
            </div>
          </section>

          <section className={s.features}>
            <p className={s.sectionLabel}>Features</p>
            <div className={s.grid}>
              {features.map((f) => (
                <div key={f.title} className={s.card}>
                  <h3 className={s.cardTitle}>{f.title}</h3>
                  <p className={s.cardDesc}>{f.desc}</p>
                </div>
              ))}
            </div>
          </section>

          <section className={s.codeSection}>
            <h2 className={s.codeTitle}>Simple by design</h2>
            <p className={s.codeSubtitle}>
              Define a struct. Register it. Query.
            </p>
            <div className={s.codeBlock}>
              <div className={s.codeHeader}>
                <span className={s.dot} />
                <span className={s.dot} />
                <span className={s.dot} />
                <span className={s.fileName}>main.go</span>
              </div>
              <div className={s.codeBody}>
                <CodeExample styles={s} />
              </div>
            </div>
          </section>

          <section className={s.traceway}>
            <a
              href="https://tracewayapp.com"
              target="_blank"
              rel="noopener noreferrer"
              className={s.tracewayLink}
            >
              <img
                src="/traceway-logo-white.svg"
                alt="Traceway"
                className={s.tracewayLogo}
              />
            </a>
            <p className={s.tracewayText}>
              Built and maintained by the{' '}
              <a
                href="https://tracewayapp.com"
                target="_blank"
                rel="noopener noreferrer"
                className={s.tracewayInline}
              >
                Traceway
              </a>{' '}
              team.
            </p>
            <p className={s.tracewayDesc}>
              See lit in action &mdash; Traceway uses lit to power its
              observability platform for Go applications.
            </p>
            <a
              href="https://tracewayapp.com"
              target="_blank"
              rel="noopener noreferrer"
              className={s.secondaryBtn}
            >
              Visit Traceway
            </a>
          </section>

          <footer className={s.footer}>
            MIT {new Date().getFullYear()} &copy;{' '}
            <a
              href="https://github.com/tracewayapp"
              target="_blank"
              rel="noopener noreferrer"
              className={s.footerLink}
            >
              Traceway
            </a>
          </footer>
        </div>
      </div>
    </>
  )
}
