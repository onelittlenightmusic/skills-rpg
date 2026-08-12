// RPG Try Keys – line-art door + key visualiser
const React = window.React;

// The door and the keys come from the shared kit, so the door on this card and
// the door on the door card are the same door.
import { Door, Key, PALETTE as P } from '/api/v1/plugins/rpg-artkit.js';

// ── Main component ────────────────────────────────────────────────────────────
function RpgTryKeysSection({ want, isChild, isControl, isFocused }) {
  const state   = want.state?.current || {};
  const scene   = state.scene  || {};
  const tried   = state.tried  || [];
  const ok      = state.ok;
  const target  = state.target || scene.target || '';
  const summary = state.summary || '';
  const isDone  = ok || !!state.error;
  const allKeys = (scene.all_keys?.length > 0) ? scene.all_keys : tried;

  // Parse currently-trying key from live summary ("trying key_xxx...")
  const tryMatch  = !isDone && summary.match(/trying (key_\w+)/);
  const currentKey = tryMatch ? tryMatch[1] : null;

  // Build per-key status
  const failedSet  = new Set(ok ? tried.slice(0, -1) : tried);
  const successKey = ok ? tried[tried.length - 1] : null;

  function keyStatus(k) {
    if (k === successKey)  return 'success';
    if (failedSet.has(k))  return 'failed';
    if (k === currentKey)  return 'trying';
    return 'idle';
  }

  // Door state
  const doorOpen   = !!ok;
  const doorLocked = !ok && !state.error;

  // Human-readable door label
  const doorLabel = target
    .replace(/_/g, ' ')
    .replace(/\b\w/g, c => c.toUpperCase())
    || 'Door';

  // Layout constants
  const DOOR_W = 50, DOOR_H = 90;
  const KEY_ROW_H = 22;     // px per key row
  const KEY_SVG_W = 72;     // width of key line-art SVG
  const LABEL_GAP = 8;
  const LABEL_W   = 60;     // estimated label width
  const COL_GAP   = 18;     // gap between door and key column
  const nKeys = allKeys.length;

  const sceneH = Math.max(DOOR_H, nKeys * KEY_ROW_H);
  const sceneW = DOOR_W + COL_GAP + KEY_SVG_W + LABEL_GAP + LABEL_W;
  const svgH   = sceneH + 14;  // +14 for door label below

  const doorY = (sceneH - DOOR_H) / 2;
  const keysY = (sceneH - nKeys * KEY_ROW_H) / 2;

  const labelColorOf = {
    idle:    P.goldDim,
    trying:  P.warn,
    failed:  '#7a3535',
    success: P.ok,
  };

  const statusColor = ok ? P.ok : state.error ? P.bad : P.textDim;
  const statusText  = ok ? '✓ OPENED' : state.error ? '✗ FAILED' : '⟳ TRYING KEYS';

  return window.__mywant.createCardLayout({
    className: 'rounded-lg overflow-hidden',
    style: { background:P.bg, border:`1px solid ${P.border}` },
    top: React.createElement('div', {
      className: 'flex items-center px-3 py-1.5',
      style: { background:P.bgHeader, borderBottom:`1px solid ${P.border}` },
    },
      React.createElement('span', {
        style: { fontFamily:P.mono, fontSize:11, color:statusColor },
      }, statusText),
      target && React.createElement('span', {
        style: { marginLeft:'auto', fontFamily:P.mono, fontSize:10, color:P.textFaint },
      }, doorLabel),
    ),
    content: React.createElement('div', { style: { height:'100%', overflow:'auto', display:'flex', justifyContent:'center', alignItems:'center', padding:'10px 12px' } },
      React.createElement('svg', {
        width: sceneW,
        height: svgH,
        style: { display:'block' },
      },

        // Door
        React.createElement('g', { transform: `translate(0, ${doorY})` },
          React.createElement(Door, { open:doorOpen, locked:doorLocked }),
          React.createElement('text', {
            x: DOOR_W / 2, y: DOOR_H + 11,
            textAnchor:'middle', fontSize:8,
            fontFamily:P.mono,
            fill: doorOpen ? P.ok : doorLocked ? P.bad : P.textFaint,
          }, doorLabel),
        ),

        // Keys
        React.createElement('g', { transform: `translate(${DOOR_W + COL_GAP}, ${keysY})` },
          allKeys.map((key, i) => {
            const st    = keyStatus(key);
            const short = key.replace(/^key_/, '');
            const lc    = labelColorOf[st] || labelColorOf.idle;
            return React.createElement('g', {
              key,
              transform: `translate(0, ${i * KEY_ROW_H + (KEY_ROW_H - 14) / 2})`,
            },
              React.createElement(Key, { status:st }),
              React.createElement('text', {
                x: KEY_SVG_W + LABEL_GAP,
                y: 12,
                fill: lc,
                fontSize: 10,
                fontFamily: P.mono,
              }, short),
            );
          }),
          allKeys.length === 0 && React.createElement('text', {
            x:0, y:18, fill:P.muted, fontSize:11, fontFamily:P.mono,
          }, '…'),
        ),
      ),
    ),
  });
}

window.__mywant.registerPlugin({
  types: ['rpg_try_keys'],
  ContentSection: RpgTryKeysSection,
  hideFinalResult: true,
});
