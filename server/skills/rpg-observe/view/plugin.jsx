// RPG Observe card plugin — JIT loaded from ~/.mywant/custom-types/rpg-observe/view/plugin.jsx
// window.React and window.__mywant are provided by the host app.
const React = window.React;

// Figures and glyphs come from the shared kit — this card is a map of the same
// world the other rpg cards each show one corner of.
import { You, ChapHead, KeyGlyph, DoorGlyph, PALETTE as P } from '/api/v1/plugins/rpg-artkit.js';

// ── Layout constants ──────────────────────────────────────────────────────────
const BW = 54, BH = 21, BG = 27, PAD = 15;
const BY = 57;
const FCY = BY - 33;
const DVY = BY + BH + 15;
const QRAD = 36;

// ── Chap corner ───────────────────────────────────────────────────────────────
// chap sits in the top-left quadrant with the keys he is carrying laid out
// beside him — the inventory belongs to him, so it is drawn where he is.
function ChapCorner({ chapItems, usedKeys }) {
  const cx = 12, cy = 12, r = 5.25;
  return React.createElement('g', null,
    React.createElement('path', {
      d: `M ${QRAD} 0 A ${QRAD} ${QRAD} 0 0 1 0 ${QRAD} L 0 0 Z`,
      fill: '#0d2018', stroke: P.ok, strokeWidth: 1,
    }),
    React.createElement(ChapHead, { cx, cy, r }),
    React.createElement('text', { x: cx, y: cy + r + 7, textAnchor: 'middle', fontSize: 6.5, fill: P.ok }, 'chap'),
    ...(chapItems || []).map((key, i) =>
      React.createElement(KeyGlyph, {
        key,
        x: QRAD + 6 + i * 32,
        y: 4,
        label: key,
        used: (usedKeys || []).includes(key),
      })
    ),
  );
}

// ── Main component ────────────────────────────────────────────────────────────
function RpgObserveSection({ want, isChild, isControl, isFocused, isExpanded }) {
  const fr    = want.state?.final_result;
  const scene = fr?.scene ?? want.state?.current?.scene;

  if (!scene?.stage_id) {
    return window.__mywant.createCardLayout({
      className: 'rounded-lg bg-gray-900',
      centerContent: true,
      content: React.createElement('span', {
        className: 'text-gray-500 text-xs font-mono',
      }, '観測中…'),
    });
  }

  const { stage_id, title, nodes, edges, devices, next_goal, event_history, chap_items } = scene;

  const boxCX = (i) => PAD + i * (BW + BG) + BW / 2;
  const maxDeviceTextEnd = devices.length > 0
    ? Math.max(...devices.map((d, i) => PAD + i * 108 + 16 + d.label.length * 5.5))
    : 0;
  const SW    = Math.max(120, nodes.length * (BW + BG) - BG + PAD * 2, Math.ceil(maxDeviceTextEnd));
  const SH    = DVY + (devices.length ? 14 : 2);

  const edgeMap = new Map();
  edges.forEach(e => {
    edgeMap.set(`${e.from}|${e.to}`, e);
    edgeMap.set(`${e.to}|${e.from}`, e);
  });

  // Bottom: event history + next goal (expanded only)
  const bottomParts = [];
  if (isExpanded && event_history && event_history.length > 0) {
    const lastAction = [...event_history].reverse().find(e =>
      e.action !== 'observe' && (e.narration?.conversations?.length > 0 || e.narration?.on_success)
    );
    const lastObserve = [...event_history].reverse().find(e =>
      e.action === 'observe' && e.narration?.conversations?.length > 0
    );
    const ev = lastAction || lastObserve || event_history[event_history.length - 1];
    const ok = ev.result === 'ok';
    const narr = ev.narration;
    const narratext = narr ? (ok ? (narr.on_success || narr.lore) : narr.lore) : null;
    bottomParts.push(
      React.createElement('div', { key: 'hist', style: { borderTop: `1px solid ${P.borderSoft}` } },
        React.createElement('div', {
          className: 'px-3 py-0.5 text-xs font-mono flex items-baseline gap-1',
          style: { color: P.text },
        },
          React.createElement('span', { style: { color: ok ? P.ok : P.bad, flexShrink: 0 } }, ok ? '✓' : '✗'),
          React.createElement('span', null, `${ev.actor} ${ev.action}`),
          ev.target ? React.createElement('span', { style: { color: '#79c0ff' } }, ev.target) : null,
          ev.reason ? React.createElement('span', { style: { color: P.textFaint } }, `(${ev.reason})`) : null,
        ),
        narr?.conversations?.length > 0 ? React.createElement('div', {
          className: 'px-3 pb-2 flex flex-col gap-0.5',
        },
          ...narr.conversations.map((line, j) => {
            const speakerColor = line.speaker === 'you' ? P.info
              : line.speaker === 'chap' ? P.ok
              : P.textDim;
            return React.createElement('div', { key: j, className: 'flex gap-1.5 text-xs leading-snug' },
              React.createElement('span', {
                style: { color: speakerColor, flexShrink: 0, fontWeight: 'bold', minWidth: 28 },
              }, line.speaker),
              React.createElement('span', { style: { color: P.text } }, line.text),
            );
          }),
        ) : null,
        narratext ? React.createElement('div', {
          className: 'px-4 pb-1 text-xs leading-snug whitespace-pre-wrap',
          style: { color: P.textDim },
        }, narratext.trim()) : null,
      )
    );
  }
  if (isExpanded && next_goal) {
    bottomParts.push(
      React.createElement('div', {
        key: 'goal',
        className: 'px-3 py-1 text-xs leading-snug',
        style: { color: P.text, borderTop: `1px solid ${P.borderSoft}` },
      },
        React.createElement('span', { style: { color: P.amber } }, '▶ '),
        next_goal,
      )
    );
  }
  const bottomEl = bottomParts.length > 0
    ? React.createElement('div', null, ...bottomParts)
    : null;

  return window.__mywant.createCardLayout({
    className: 'rounded-lg overflow-hidden',
    style: { background: P.bg },
    top: React.createElement('div', {
      className: 'px-3 py-1',
      style: { background: P.bgHeader, borderBottom: `1px solid ${P.border}` },
    },
      next_goal
        ? React.createElement('div', {
            className: 'text-[11px] font-mono font-semibold leading-tight',
            style: { color: P.ok },
          }, next_goal)
        : null,
      React.createElement('div', {
        className: 'text-[10px] font-mono leading-tight',
        style: { color: P.textDim },
      }, stage_id + (title ? ` · ${title}` : '')),
    ),
    content: React.createElement('div', { style: { height: '100%', overflow: 'hidden' } },
      React.createElement('svg', { viewBox: `0 0 ${SW} ${SH}`, width: '100%', height: '100%', preserveAspectRatio: 'xMidYMid meet', style: { display: 'block' } },

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
            React.createElement('line', { x1, y1: ly, x2, y2: ly, stroke: P.border, strokeWidth: 1 }),
            edge?.door ? React.createElement('g', null,
              React.createElement(DoorGlyph, { cx: mx, cy: ly, open: edge.door.open, locked: edge.door.locked }),
              React.createElement('text', {
                x: mx, y: ly + 19,
                textAnchor: 'middle', fontSize: 7,
                fontFamily: P.mono,
                fill: edge.door.open ? P.ok : edge.door.locked ? P.bad : P.textFaint,
              }, edge.door.id),
            ) : null,
          );
        }),

        // nodes
        ...nodes.map((node, i) => {
          const cx    = boxCX(i);
          const short = node.label;
          return React.createElement('g', { key: node.id },
            React.createElement('rect', {
              x: cx - BW/2, y: BY, width: BW, height: BH, rx: 3,
              fill: node.has_you ? '#1a3a5c' : P.bgHeader,
              stroke: node.has_you ? P.infoDeep : P.border,
              strokeWidth: node.has_you ? 1.5 : 1,
            }),
            React.createElement('text', {
              x: cx, y: BY + BH/2 + 4, textAnchor: 'middle', fontSize: 9,
              fill: node.has_you ? '#79c0ff' : P.textDim,
            }, short),
            node.has_you ? React.createElement('g', null,
              React.createElement(You, { cx, cy: FCY }),
              React.createElement('text', { x: cx, y: FCY - 7, textAnchor: 'middle', fontSize: 8, fill: P.infoDeep }, 'you'),
            ) : null,
          );
        }),

        // devices
        ...devices.map((dev, i) => {
          const dx    = PAD + i * 108;
          const short = dev.label;
          return React.createElement('g', { key: dev.id },
            React.createElement('circle', {
              cx: dx + 6, cy: DVY + 6, r: 6,
              fill: dev.on ? '#f59e0b' : P.borderSoft,
              stroke: dev.on ? P.amber : P.textFaint, strokeWidth: 1.2,
            }),
            React.createElement('text', { x: dx + 16, y: DVY + 10, fontSize: 9, fill: P.textDim }, short),
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
    ),
    bottom: bottomEl,
  });
}

window.__mywant.registerPlugin({
  types: ['rpg_observe'],
  ContentSection: RpgObserveSection,
  hideFinalResult: true,
});
