// RPG Try Keys – line-art door + key visualiser
const React = window.React;

// ── Door (50×90 SVG group) ────────────────────────────────────────────────────
function Door({ open, locked }) {
  const W = 50, H = 90;

  if (open) {
    return React.createElement('g', null,
      // outer frame
      React.createElement('rect', { x:0, y:0, width:W, height:H, rx:3,
        fill:'none', stroke:'#30363d', strokeWidth:1.5 }),
      // door panel swung open — perspective trapezoid
      React.createElement('path', {
        d:`M 2 2 L 18 9 L 18 ${H-9} L 2 ${H-2} Z`,
        fill:'none', stroke:'#3fb950', strokeWidth:2,
      }),
      // opening (dark interior)
      React.createElement('rect', { x:18, y:2, width:W-20, height:H-4, fill:'#030a04' }),
      // dashed border on opening
      React.createElement('rect', { x:18, y:2, width:W-20, height:H-4,
        fill:'none', stroke:'#238636', strokeWidth:0.7, strokeDasharray:'3 3', opacity:0.5 }),
      // glow dot in opening
      React.createElement('circle', { cx:34, cy:H/2, r:3, fill:'#3fb950', opacity:0.75 }),
    );
  }

  const sc = locked ? '#f85149' : '#6e7681';

  return React.createElement('g', null,
    // door panel
    React.createElement('rect', { x:0, y:0, width:W, height:H, rx:3,
      fill: locked ? '#180505' : '#0d1117', stroke:sc, strokeWidth:2 }),
    // inner panel border
    React.createElement('rect', { x:6, y:6, width:W-12, height:H-12, rx:2,
      fill:'none', stroke:sc, strokeWidth:0.6, opacity:0.3 }),
    // keyhole — circle
    React.createElement('circle', { cx:W/2, cy:H*0.42, r:5.5,
      fill:'none', stroke:sc, strokeWidth:1.5 }),
    // keyhole — downward slot
    React.createElement('path', {
      d:`M ${W/2-3.5} ${H*0.42+4} L ${W/2} ${H*0.42+16} L ${W/2+3.5} ${H*0.42+4}`,
      fill:'none', stroke:sc, strokeWidth:1.5, strokeLinejoin:'round',
    }),
    // door knob
    React.createElement('circle', { cx:W-11, cy:H/2, r:4,
      fill:'none', stroke:sc, strokeWidth:1.5 }),
    // padlock body (locked only)
    locked && React.createElement('g', null,
      React.createElement('rect', { x:W/2-7, y:H-22, width:14, height:10, rx:2,
        fill:'none', stroke:'#f85149', strokeWidth:1.2 }),
      React.createElement('path', {
        d:`M ${W/2-4} ${H-22} L ${W/2-4} ${H-27} Q ${W/2} ${H-31} ${W/2+4} ${H-27} L ${W/2+4} ${H-22}`,
        fill:'none', stroke:'#f85149', strokeWidth:1.2,
      }),
    ),
  );
}

// ── Key (line-art, fits in 72×18 px) ─────────────────────────────────────────
//  status: 'idle' | 'trying' | 'failed' | 'success'
function Key({ status }) {
  const strokeOf = {
    idle:    '#c9971a',
    trying:  '#f5e000',
    failed:  '#7a3535',
    success: '#3fb950',
  };
  const sc  = strokeOf[status] || strokeOf.idle;
  const dim = status === 'failed';

  const R  = 7;   // bow outer radius
  const HR = 3;   // bow hole radius
  const SX = R * 2 + 1;  // shaft start x
  const SL = 40;  // shaft length
  const SY = R;   // center Y

  return React.createElement('g', { opacity: dim ? 0.45 : 1 },
    // bow outer circle
    React.createElement('circle', { cx:R, cy:SY, r:R, fill:'none', stroke:sc, strokeWidth:1.5 }),
    // bow hole
    React.createElement('circle', { cx:R, cy:SY, r:HR, fill:'none', stroke:sc, strokeWidth:1 }),
    // shaft
    React.createElement('line', { x1:SX, y1:SY, x2:SX+SL, y2:SY, stroke:sc, strokeWidth:2 }),
    // teeth — 3 downward ticks near end of shaft
    React.createElement('line', { x1:SX+SL-14, y1:SY, x2:SX+SL-14, y2:SY+6, stroke:sc, strokeWidth:2 }),
    React.createElement('line', { x1:SX+SL-7,  y1:SY, x2:SX+SL-7,  y2:SY+4, stroke:sc, strokeWidth:2 }),
    React.createElement('line', { x1:SX+SL,    y1:SY, x2:SX+SL,    y2:SY+7, stroke:sc, strokeWidth:2 }),
    // — failed: diagonal cross ——
    status === 'failed' && React.createElement('line', {
      x1:0, y1:0, x2:SX+SL+2, y2:SY*2,
      stroke:'#f85149', strokeWidth:1, strokeDasharray:'3 2',
    }),
    status === 'failed' && React.createElement('line', {
      x1:SX+SL+2, y1:0, x2:0, y2:SY*2,
      stroke:'#f85149', strokeWidth:1, strokeDasharray:'3 2',
    }),
    // — success: checkmark ——
    status === 'success' && React.createElement('text', {
      x: SX+SL+5, y: SY+5,
      fill:'#3fb950', fontSize:12, fontWeight:'bold', fontFamily:'ui-monospace,monospace',
    }, '✓'),
    // — trying: pulsing dot ——
    status === 'trying' && React.createElement('circle', {
      cx: SX+SL+7, cy: SY, r:3.5,
      fill:'#f5e000', opacity:0.9,
    }),
  );
}

// ── Main component ────────────────────────────────────────────────────────────
function RpgTryKeysSection({ want, isChild, isControl, isFocused }) {
  const state   = want.state?.current || {};
  const scene   = state.scene  || {};
  const tried   = state.tried  || [];
  const ok      = state.ok;
  const target  = state.target || scene.target || '';
  const summary = state.summary || '';
  const mt = isChild || (isControl && !isFocused) ? 'mt-2' : 'mt-3';

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
    idle:    '#8b7020',
    trying:  '#f5e000',
    failed:  '#7a3535',
    success: '#3fb950',
  };

  const statusColor = ok ? '#3fb950' : state.error ? '#f85149' : '#8b949e';
  const statusText  = ok ? '✓ OPENED' : state.error ? '✗ FAILED' : '⟳ TRYING KEYS';

  return React.createElement('div', {
    className: `${mt} rounded-lg overflow-hidden`,
    style: { background:'#0d1117', border:'1px solid #30363d' },
  },

    // ── header bar ──────────────────────────────────────────────
    React.createElement('div', {
      className: 'flex items-center px-3 py-1.5',
      style: { background:'#161b22', borderBottom:'1px solid #30363d' },
    },
      React.createElement('span', {
        style: { fontFamily:'ui-monospace,monospace', fontSize:11, color:statusColor },
      }, statusText),
      target && React.createElement('span', {
        style: { marginLeft:'auto', fontFamily:'ui-monospace,monospace', fontSize:10, color:'#6e7681' },
      }, doorLabel),
    ),

    // ── scene ────────────────────────────────────────────────────
    React.createElement('div', { style: { padding:'10px 12px', overflowX:'auto' } },
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
            fontFamily:'ui-monospace,monospace',
            fill: doorOpen ? '#3fb950' : doorLocked ? '#f85149' : '#6e7681',
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
                fontFamily: 'ui-monospace,monospace',
              }, short),
            );
          }),
          allKeys.length === 0 && React.createElement('text', {
            x:0, y:18, fill:'#484f58', fontSize:11, fontFamily:'ui-monospace,monospace',
          }, '…'),
        ),
      ),
    ),
  );
}

window.__mywant.registerPlugin({
  types: ['rpg_try_keys'],
  ContentSection: RpgTryKeysSection,
  hideFinalResult: true,
});
