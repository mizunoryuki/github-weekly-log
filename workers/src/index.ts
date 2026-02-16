import { Hono } from 'hono'
import type { Context } from 'hono'

type Env = {
  Bindings: {
    CF_API_TOKEN: string
    CF_ACCOUNT_ID: string
    D1_DATABASE_NAME: string
    ALLOWED_ORIGINS: string
  }
}

type D1ApiResponse = {
  success: boolean
  errors?: Array<{ code: number; message: string }>
  result?: Array<{ results?: unknown[] }>
}

const app = new Hono<Env>()

const CF_API_BASE = 'https://api.cloudflare.com/client/v4'

const parseAllowedOrigins = (value: string) =>
  new Set(
    value
      .split(',')
      .map((origin) => origin.trim())
      .filter(Boolean)
  )

app.use('/api/*', async (c, next) => {
  const allowedOrigins = parseAllowedOrigins(c.env.ALLOWED_ORIGINS)
  const origin = c.req.header('Origin')
  if (origin && allowedOrigins.has(origin)) {
    c.header('Access-Control-Allow-Origin', origin)
    c.header('Access-Control-Allow-Credentials', 'true')
    c.header('Access-Control-Allow-Headers', 'Content-Type')
    c.header('Access-Control-Allow-Methods', 'GET,OPTIONS')
    c.header('Vary', 'Origin')
  }

  if (c.req.method === 'OPTIONS') {
    return c.body(null, 204)
  }

  await next()
})

const getDatabaseId = async (c: Context<Env>) => {
  const url = new URL(
    `${CF_API_BASE}/accounts/${c.env.CF_ACCOUNT_ID}/d1/database`
  )
  url.searchParams.set('name', c.env.D1_DATABASE_NAME)

  const response = await fetch(url.toString(), {
    headers: {
      Authorization: `Bearer ${c.env.CF_API_TOKEN}`
    }
  })

  const data = (await response.json()) as {
    success: boolean
    errors?: Array<{ code: number; message: string }>
    result?: Array<{ uuid: string }>
  }

  if (!response.ok || !data.success || !data.result || data.result.length === 0) {
    const message = data.errors?.[0]?.message ?? 'Database not found'
    throw new Error(`D1 list failed: ${response.status} ${message}`)
  }

  return data.result[0].uuid
}

const d1Query = async (c: Context<Env>, sql: string) => {
  const databaseId = await getDatabaseId(c)
  const response = await fetch(
    `${CF_API_BASE}/accounts/${c.env.CF_ACCOUNT_ID}/d1/database/${databaseId}/query`,
    {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${c.env.CF_API_TOKEN}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ sql })
    }
  )

  const data = (await response.json()) as D1ApiResponse
  if (!response.ok || !data.success) {
    const message = data.errors?.[0]?.message ?? 'D1 query failed'
    throw new Error(`D1 query failed: ${response.status} ${message}`)
  }

  return data.result?.[0]?.results ?? []
}

app.get('/', (c) => c.text('OK'))

app.get('/api/weekly', async (c) => {
  try {
    const rows = await d1Query(
      c,
      'SELECT id, start_date, end_date, total_commits, active_days, created_at FROM weekly_stats ORDER BY start_date DESC'
    )
    return c.json({ items: rows })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unexpected error'
    return c.json({ error: message }, 500)
  }
})

app.get('/api/weekly/:id', async (c) => {
  const id = Number(c.req.param('id'))
  if (!Number.isFinite(id) || id <= 0) {
    return c.json({ error: 'Invalid id' }, 400)
  }

  try {
    const summary = (await d1Query(
      c,
      `SELECT total_commits, active_days FROM weekly_stats WHERE id = ${id}`
    ))[0] as { total_commits: number; active_days: number } | undefined

    if (!summary) {
      return c.json({ error: 'Not found' }, 404)
    }

    const daily = (await d1Query(
      c,
      `SELECT date, commits FROM daily_commits WHERE weekly_stats_id = ${id} ORDER BY date ASC`
    )) as Array<{ date: string; commits: number }>

    const hourly = (await d1Query(
      c,
      `SELECT hour, commits FROM hourly_activity WHERE weekly_stats_id = ${id} ORDER BY hour ASC`
    )) as Array<{ hour: number; commits: number }>

    const repos = (await d1Query(
      c,
      `SELECT repo_name, commits, bar_width FROM repo_details WHERE weekly_stats_id = ${id} ORDER BY commits DESC`
    )) as Array<{ repo_name: string; commits: number; bar_width: number }>

    const languages = (await d1Query(
      c,
      `SELECT language, commits, is_main FROM language_commits WHERE weekly_stats_id = ${id} ORDER BY commits DESC`
    )) as Array<{ language: string; commits: number; is_main: number }>

    return c.json({
      summary: {
        total_commits: summary.total_commits,
        active_days: summary.active_days
      },
      daily: {
        labels: daily.map((row) => row.date),
        data: daily.map((row) => row.commits)
      },
      hourly: {
        labels: hourly.map((row) => String(row.hour)),
        data: hourly.map((row) => row.commits)
      },
      repos: {
        labels: repos.map((row) => row.repo_name),
        data: repos.map((row) => row.commits),
        barWidth: repos.map((row) => row.bar_width)
      },
      languages: {
        labels: languages.map((row) => row.language),
        data: languages.map((row) => row.commits),
        isMain: languages.map((row) => Boolean(row.is_main))
      }
    })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unexpected error'
    return c.json({ error: message }, 500)
  }
})

export default app
