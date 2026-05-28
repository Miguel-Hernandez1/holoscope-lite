'use strict';

const REFRESH_MS = 5000;

function $(id) { return document.getElementById(id); }

function fmtTime(iso) {
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function fmtLatency(ms) {
  return ms >= 1000 ? (ms / 1000).toFixed(2) + 's' : ms.toFixed(1) + 'ms';
}

function statusClass(code) { return code >= 500 ? 'status-err' : 'status-ok'; }
function methodClass(method) { return `badge-method ${method}`; }

// --- Header stats ---
function renderStats(data) {
  const reqs   = data.total_requests ?? 0;
  const errs   = data.total_errors   ?? 0;
  const pct    = data.error_rate_pct ?? 0;
  const avg    = data.avg_latency_ms ?? 0;

  $('stat-reqs').textContent      = reqs;
  $('stat-errors').textContent    = errs;
  $('stat-error-pct').textContent = reqs > 0 ? `(${pct.toFixed(1)}%)` : '';
  $('stat-avg').textContent       = fmtLatency(avg);
}

// --- Latency ranking (CSS bars, no Chart.js) ---
function renderLatencyRanking(data) {
  const container = $('latency-ranking');
  const endpoints = (data.endpoints ?? [])
    .filter(e => e.request_count > 0)
    .sort((a, b) => b.avg_latency_ms - a.avg_latency_ms);

  container.replaceChildren();

  if (endpoints.length === 0) {
    const el = document.createElement('div');
    el.className = 'lat-empty';
    el.textContent = 'no data yet';
    container.appendChild(el);
    return;
  }

  const maxMs = endpoints[0].avg_latency_ms;

  for (const ep of endpoints) {
    const pct = maxMs > 0 ? Math.max((ep.avg_latency_ms / maxMs) * 100, 2) : 2;

    const row     = document.createElement('div');
    row.className = 'lat-row';

    const path    = document.createElement('div');
    path.className = 'lat-path';
    path.textContent = ep.path;
    path.title       = ep.path;

    const barWrap = document.createElement('div');
    barWrap.className = 'lat-bar-wrap';
    const bar = document.createElement('div');
    bar.className     = 'lat-bar';
    bar.style.width   = pct + '%';
    barWrap.appendChild(bar);

    const value    = document.createElement('div');
    value.className = 'lat-value';
    value.textContent = fmtLatency(ep.avg_latency_ms);

    row.append(path, barWrap, value);
    container.appendChild(row);
  }
}

// --- Traces table ---
async function fetchMetrics() {
  const res = await fetch('/observability/metrics');
  return res.json();
}

async function fetchTraces() {
  const res = await fetch('/observability/traces');
  return res.json();
}

function renderTraces(traces) {
  const tbody = $('traces-body');
  tbody.replaceChildren();

  if (!traces || traces.length === 0) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan   = 6;
    td.className = 'empty-state';
    td.textContent = 'no traces yet — hit some endpoints to populate';
    tr.appendChild(td);
    tbody.appendChild(tr);
    return;
  }

  for (const t of traces) {
    const tr = document.createElement('tr');

    const tdID = document.createElement('td');
    tdID.className   = 'trace-id';
    tdID.textContent = t.trace_id.slice(0, 13);

    const tdMethod = document.createElement('td');
    const badge    = document.createElement('span');
    badge.className   = methodClass(t.method);
    badge.textContent = t.method;
    tdMethod.appendChild(badge);

    const tdPath = document.createElement('td');
    tdPath.className   = 'trace-path';
    tdPath.textContent = t.path;

    const tdStatus = document.createElement('td');
    tdStatus.className   = statusClass(t.status_code);
    tdStatus.textContent = t.status_code;

    const tdLatency = document.createElement('td');
    tdLatency.className   = 'trace-latency';
    tdLatency.textContent = fmtLatency(t.latency_ms);

    const tdTime = document.createElement('td');
    tdTime.className   = 'trace-time';
    tdTime.textContent = fmtTime(t.timestamp);

    tr.append(tdID, tdMethod, tdPath, tdStatus, tdLatency, tdTime);
    tr.addEventListener('click', () => openModal(t.trace_id));
    tbody.appendChild(tr);
  }
}

// --- Trace detail modal ---
async function openModal(traceID) {
  const res  = await fetch(`/observability/traces/${traceID}`);
  const data = await res.json();
  $('modal-trace-id').textContent = traceID;
  $('modal-body-pre').textContent = JSON.stringify(data, null, 2);
  $('trace-modal').showModal();
}

function closeModal() {
  $('trace-modal').close();
}

// --- Refresh loop ---
async function fetchAndRender() {
  try {
    const [metrics, traces] = await Promise.all([fetchMetrics(), fetchTraces()]);
    renderStats(metrics);
    renderLatencyRanking(metrics);
    renderTraces(traces);
    $('last-updated').textContent = 'refreshed ' + new Date().toLocaleTimeString();
  } catch (err) {
    console.error('refresh failed:', err);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  fetchAndRender();
  setInterval(fetchAndRender, REFRESH_MS);

  $('modal-close-btn').addEventListener('click', closeModal);
  $('trace-modal').addEventListener('click', e => {
    if (e.target === $('trace-modal')) closeModal();
  });
});
