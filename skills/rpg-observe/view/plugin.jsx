// RPG Observe card plugin — JIT loaded from ~/.mywant/custom-types/rpg-observe/view/plugin.jsx
// window.React and window.__mywant are provided by the host app.
const React = window.React;

// ── Layout constants ──────────────────────────────────────────────────────────
const BW = 54, BH = 21, BG = 27, PAD = 15;
const BY = 57;
const FCY = BY - 33;
const DVY = BY + BH + 15;
const QRAD = 36;

// ── Stick figure (you) ────────────────────────────────────────────────────────
function You({ cx, cy }) {
  const p = { stroke: '#58a6ff', strokeWidth: 1.8 };
  return React.createElement('g', null,
    React.createElement('circle', { cx, cy, r: 5.25, fill: 'none', ...p }),
    React.createElement('line', { x1: cx,     y1: cy+5.25, x2: cx,     y2: cy+18,   ...p }),
    React.createElement('line', { x1: cx-7.5, y1: cy+10.5, x2: cx+7.5, y2: cy+10.5, ...p }),
    React.createElement('line', { x1: cx,     y1: cy+18,   x2: cx-6,   y2: cy+27,   ...p }),
    React.createElement('line', { x1: cx,     y1: cy+18,   x2: cx+6,   y2: cy+27,   ...p }),
  );
}

// ── Key icon ──────────────────────────────────────────────────────────────────
function KeyIcon({ x, y, label, used }) {
  const color = used ? '#6e7681' : '#e3b341';
  const short = label.replace(/^key_/, '').slice(0, 6);
  return React.createElement('g', null,
    React.createElement('rect', { x, y, width: 28, height: 11, rx: 2,
      fill: used ? '#161b22' : '#2d1f00',
      stroke: color, strokeWidth: 0.8,
    }),
    React.createElement('text', { x: x + 14, y: y + 8, textAnchor: 'middle', fontSize: 6.5, fill: color }, short),
  );
}

// ── Chap corner ───────────────────────────────────────────────────────────────
function ChapCorner({ chapItems, usedKeys }) {
  const cx = 12, cy = 12, r = 5.25;
  return React.createElement('g', null,
    React.createElement('path', {
      d: `M ${QRAD} 0 A ${QRAD} ${QRAD} 0 0 1 0 ${QRAD} L 0 0 Z`,
      fill: '#0d2018', stroke: '#3fb950', strokeWidth: 1,
    }),
    React.createElement('circle', { cx, cy, r, fill: '#132d1a', stroke: '#3fb950', strokeWidth: 1.5 }),
    React.createElement('text', { x: cx, y: cy + r + 7, textAnchor: 'middle', fontSize: 6.5, fill: '#3fb950' }, 'chap'),
    ...(chapItems || []).map((key, i) =>
      React.createElement(KeyIcon, {
        key,
        x: QRAD + 6 + i * 32,
        y: 4,
        label: key,
        used: (usedKeys || []).includes(key),
      })
    ),
  );
}

// ── Door icon ─────────────────────────────────────────────────────────────────
function DoorIcon({ cx, cy, door }) {
  if (door.open) {
    return React.createElement('rect', {
      x: cx-3, y: cy-7.5, width: 6, height: 13.5, rx: 1.5,
      fill: '#238636', transform: `rotate(-55,${cx},${cy})`,
    });
  }
  return React.createElement('rect', {
    x: cx-3.75, y: cy-9, width: 7.5, height: 18, rx: 1.5,
    fill: door.locked ? '#3d1f1f' : '#2d333b',
    stroke: door.locked ? '#f85149' : '#6e7681',
    strokeWidth: 1.2,
  });
}

// ── Main component ────────────────────────────────────────────────────────────
function RpgObserveSection({ want, isChild, isControl, isFocused }) {
  const fr    = want.state?.final_result;
  const scene = fr?.scene ?? want.state?.current?.scene;
  const mt    = isChild || (isControl && !isFocused) ? 'mt-2' : 'mt-3';

  if (!scene?.stage_id) {
    return React.createElement('div', {
      className: `${mt} rounded-lg bg-gray-900 text-gray-500 text-xs font-mono px-3 py-2`,
    }, '観測中…');
  }

  const { stage_id, title, nodes, edges, devices, next_goal, event_history, chap_items } = scene;

  const boxCX = (i) => PAD + i * (BW + BG) + BW / 2;
  const SW    = Math.max(120, nodes.length * (BW + BG) - BG + PAD * 2);
  const SH    = DVY + (devices.length ? 14 : 2);

  const edgeMap = new Map();
  edges.forEach(e => {
    edgeMap.set(`${e.from}|${e.to}`, e);
    edgeMap.set(`${e.to}|${e.from}`, e);
  });

  return React.createElement('div', {
    className: `${mt} rounded-lg overflow-hidden`,
    style: { background: '#0d1117' },
  },

    // title bar
    React.createElement('div', {
      className: 'flex items-center gap-1.5 px-3 py-1',
      style: { background: '#161b22', borderBottom: '1px solid #30363d' },
    },
      React.createElement('span', { className: 'w-2 h-2 rounded-full bg-red-500 opacity-80' }),
      React.createElement('span', { className: 'w-2 h-2 rounded-full bg-yellow-400 opacity-80' }),
      React.createElement('span', { className: 'w-2 h-2 rounded-full bg-green-500 opacity-80' }),
      React.createElement('span', { className: 'ml-2 text-xs font-mono', style: { color: '#8b949e' } },
        stage_id + (title ? ` · ${title}` : '')),
      React.createElement('span', { className: 'ml-auto flex items-center gap-1' },
        React.createElement('span', { className: 'w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse' }),
        React.createElement('span', { className: 'text-xs', style: { color: '#3fb950' } }, 'live'),
      ),
    ),

    // SVG scene
    React.createElement('svg', { viewBox: `0 0 ${SW} ${SH}`, width: SW, style: { display: 'block' } },

      // edges
      ...nodes.map((node, i) => {
        if (i === 0) return null;
        const prev = nodes[i - 1];
        const x1 = boxCX(i - 1) + BW / 2;
        const x2 = boxCX(i)     - BW / 2;
        const ly = BY + BH / 2;
        const mx = (x1 + x2) / 2;
        const edge = edgeMap.get(`${prev.id}|${node.id}`);
        return React.createElement('g', { key: `e${i}` },
          React.createElement('line', { x1, y1: ly, x2, y2: ly, stroke: '#30363d', strokeWidth: 1 }),
          edge?.door ? React.createElement(DoorIcon, { cx: mx, cy: ly, door: edge.door }) : null,
        );
      }),

      // nodes
      ...nodes.map((node, i) => {
        const cx    = boxCX(i);
        const short = node.label.length > 5 ? node.label.slice(0, 4) + '…' : node.label;
        return React.createElement('g', { key: node.id },
          React.createElement('rect', {
            x: cx - BW/2, y: BY, width: BW, height: BH, rx: 3,
            fill: node.has_you ? '#1a3a5c' : '#161b22',
            stroke: node.has_you ? '#388bfd' : '#30363d',
            strokeWidth: node.has_you ? 1.5 : 1,
          }),
          React.createElement('text', {
            x: cx, y: BY + BH/2 + 4, textAnchor: 'middle', fontSize: 9,
            fill: node.has_you ? '#79c0ff' : '#8b949e',
          }, short),
          node.has_you ? React.createElement('g', null,
            React.createElement(You, { cx, cy: FCY }),
            React.createElement('text', { x: cx, y: FCY - 7, textAnchor: 'middle', fontSize: 8, fill: '#388bfd' }, 'you'),
          ) : null,
        );
      }),

      // devices
      ...devices.map((dev, i) => {
        const dx    = PAD + i * 108;
        const short = dev.label.length > 8 ? dev.label.slice(0, 7) + '…' : dev.label;
        return React.createElement('g', { key: dev.id },
          React.createElement('circle', {
            cx: dx + 6, cy: DVY + 6, r: 6,
            fill: dev.on ? '#f59e0b' : '#21262d',
            stroke: dev.on ? '#fbbf24' : '#6e7681', strokeWidth: 1.2,
          }),
          React.createElement('text', { x: dx + 16, y: DVY + 10, fontSize: 9, fill: '#8b949e' }, short),
        );
      }),

      // chap corner + key inventory
      React.createElement(ChapCorner, {
        chapItems: chap_items,
        usedKeys: (event_history || [])
          .filter(ev => ev.action === 'open' && ev.actor === 'chap' && ev.result === 'rejected' && ev.args?.key)
          .map(ev => ev.args.key),
      }),
    ),

    // event history
    event_history && event_history.length > 0 ? React.createElement('div', {
      style: { borderTop: '1px solid #21262d' },
    },
      ...event_history.map((ev, i) => {
        const ok = ev.result === 'ok';
        const isLast = i === event_history.length - 1;
        return React.createElement('div', {
          key: i,
          className: 'px-3 py-0.5 text-xs font-mono flex items-baseline gap-1',
          style: { color: isLast ? '#cdd9e5' : '#6e7681', opacity: 0.4 + 0.6 * ((i + 1) / event_history.length) },
        },
          React.createElement('span', { style: { color: ok ? '#3fb950' : '#f85149', flexShrink: 0 } }, ok ? '✓' : '✗'),
          React.createElement('span', null, `${ev.actor} ${ev.action}`),
          ev.target ? React.createElement('span', { style: { color: '#79c0ff' } }, ev.target) : null,
          ev.reason ? React.createElement('span', { style: { color: '#6e7681' } }, `(${ev.reason})`) : null,
        );
      }),
    ) : null,

    // next goal
    next_goal ? React.createElement('div', {
      className: 'px-3 py-1 text-xs leading-snug',
      style: { color: '#cdd9e5', borderTop: '1px solid #21262d' },
    },
      React.createElement('span', { style: { color: '#fbbf24' } }, '▶ '),
      next_goal,
    ) : null,
  );
}

window.__mywant.registerPlugin({
  types: ['rpg_observe'],
  ContentSection: RpgObserveSection,
  hideFinalResult: true,
});
