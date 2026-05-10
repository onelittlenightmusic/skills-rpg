// RPG Switch card plugin — JIT loaded from ~/.mywant/custom-types/rpg-switch/view/plugin.jsx
const React = window.React;
const { useState, useEffect } = React;

function SwitchContentSection({ want, isChild, isControl, isFocused }) {
  const serverOn = want.state?.current?.on === true;
  const label = want.state?.current?.label || want.spec?.params?.label || 'Switch';
  const target = want.state?.current?.target || want.spec?.params?.target;

  const [localOn, setLocalOn] = useState(serverOn);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    setLocalOn(serverOn);
  }, [serverOn]);

  const handleToggle = async () => {
    if (pending) return;
    const next = !localOn;
    setLocalOn(next);
    setPending(true);

    const id = want.metadata?.id;
    if (!id) {
      setPending(false);
      return;
    }

    try {
      const resp = await fetch(`/api/v1/webhooks/${id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action: next ? 'on' : 'off' }),
      });
      if (!resp.ok) throw new Error('Failed to toggle');
    } catch (err) {
      console.error('[RpgSwitch] toggle failed:', err);
      setLocalOn(!next);
    } finally {
      setPending(false);
    }
  };

  const compact = isChild || (isControl && !isFocused);

  return React.createElement('div', {
    className: `${compact ? 'mt-2' : 'mt-4'} flex flex-col items-center gap-2`,
    onMouseDown: (e) => e.stopPropagation(),
    onTouchStart: (e) => e.stopPropagation(),
  },
    React.createElement('button', {
      onClick: (e) => { e.stopPropagation(); handleToggle(); },
      onMouseDown: (e) => e.stopPropagation(),
      disabled: pending,
      className: "relative focus:outline-none",
      style: { opacity: pending ? 0.7 : 1, background: 'none', border: 'none', padding: 0, cursor: 'pointer' }
    },
      React.createElement('div', {
        style: {
          width: compact ? 44 : 56,
          height: compact ? 24 : 30,
          borderRadius: 999,
          background: localOn
            ? 'linear-gradient(135deg, #22c55e, #16a34a)'
            : 'linear-gradient(135deg, #6b7280, #4b5563)',
          boxShadow: localOn
            ? '0 0 8px rgba(34,197,94,0.4), inset 0 1px 2px rgba(0,0,0,0.15)'
            : 'inset 0 1px 3px rgba(0,0,0,0.25)',
          position: 'relative',
          transition: 'background 0.2s ease',
        }
      },
        React.createElement('div', {
          style: {
            position: 'absolute',
            top: compact ? 3 : 4,
            left: localOn ? (compact ? 23 : 29) : (compact ? 3 : 4),
            width: compact ? 18 : 22,
            height: compact ? 18 : 22,
            borderRadius: '50%',
            background: '#ffffff',
            boxShadow: '0 1px 4px rgba(0,0,0,0.3)',
            transition: 'left 0.18s ease',
          }
        })
      )
    ),
    React.createElement('div', {
      className: "text-[10px] text-gray-400 dark:text-gray-500 font-medium tracking-wide uppercase"
    },
      localOn ? 'ON' : 'OFF'
    ),
    target && !compact && React.createElement('div', {
      className: "text-[9px] text-gray-500 font-mono"
    }, `target: ${target}`)
  );
}

window.__mywant.registerPlugin({
  types: ['rpg_switch'],
  ContentSection: SwitchContentSection,
});

// ── RPG Activate — Generator line-art ────────────────────────────────────────
function Generator({ status }) {
  // status: 'stopped' | 'starting' | 'running' | 'failed'
  const isRunning = status === 'running';
  const isStarting = status === 'starting';
  const isFailed = status === 'failed';

  const bodyStroke  = isRunning ? '#3fb950' : isFailed ? '#f85149' : isStarting ? '#f5e000' : '#6e7681';
  const bodyFill    = isRunning ? '#041a06' : isFailed ? '#1a0404' : '#0d1117';
  const lightColor  = isRunning ? '#3fb950' : isFailed ? '#f85149' : isStarting ? '#f5e000' : '#484f58';
  const lightGlow   = isRunning ? '0 0 6px #3fb950' : isFailed ? '0 0 6px #f85149' : isStarting ? '0 0 5px #f5e000' : 'none';

  // Layout: 110×70 viewBox
  const BX = 0, BY = 18, BW = 85, BH = 38;   // main body
  const EX = 4, EY = 22, EW = 36, EH = 30;   // engine block (left)
  const GX = 44, GY = 22, GW = 37, GH = 30;  // generator housing (right)
  const CX = 14, CY = 5, CW = 8, CH = 14;    // chimney
  const PX = BX + BW, PY = BY + 6;           // power output x, start y

  // Vent slits on engine block
  const vents = [26, 30, 34].map(y =>
    React.createElement('line', { key: y, x1: EX+4, y1: y, x2: EX+EW-6, y2: y,
      stroke: bodyStroke, strokeWidth: 0.6, opacity: 0.5 })
  );

  // Coil windings on generator housing (3 arcs approximated as rects)
  const coils = [48, 55, 62].map(x =>
    React.createElement('rect', { key: x, x, y: GY+4, width: 4, height: GH-8, rx: 1,
      fill: 'none', stroke: bodyStroke, strokeWidth: 0.8, opacity: 0.6 })
  );

  // Smoke puffs from chimney (running / starting)
  const smoke = (isRunning || isStarting) ? [
    { cx: CX+4, cy: 3, r: 2.5, op: 0.5 },
    { cx: CX+6, cy: -2, r: 2,  op: 0.3 },
  ].map((s, i) =>
    React.createElement('circle', { key: i, cx: s.cx, cy: s.cy, r: s.r,
      fill: isRunning ? '#238636' : '#8b7020', opacity: s.op })
  ) : [];

  // Power output lines (running: wavy; starting: partial)
  const powerLines = (isRunning || isStarting) ? [0, 10, 20].map((dy, i) => {
    const y = PY + dy;
    const len = isStarting ? 12 : 22;
    const waveY = isRunning
      ? `M ${PX} ${y} c 5 -3 10 3 ${len} 0`
      : `M ${PX} ${y} l ${len} 0`;
    return React.createElement('path', { key: i, d: waveY,
      fill: 'none', stroke: isRunning ? '#f5e000' : '#8b7020',
      strokeWidth: 1.5, opacity: isStarting ? 0.5 : 0.9 });
  }) : [];

  // Failed: X cross over body
  const failCross = isFailed ? [
    React.createElement('line', { key: 'x1', x1: BX+4, y1: BY+4, x2: BX+BW-4, y2: BY+BH-4,
      stroke: '#f85149', strokeWidth: 1.2, strokeDasharray: '4 2' }),
    React.createElement('line', { key: 'x2', x1: BX+BW-4, y1: BY+4, x2: BX+4, y2: BY+BH-4,
      stroke: '#f85149', strokeWidth: 1.2, strokeDasharray: '4 2' }),
  ] : [];

  return React.createElement('svg', {
    width: 132, height: 70,
    style: { display: 'block', overflow: 'visible' },
  },
    // smoke
    ...smoke,
    // chimney
    React.createElement('rect', { x: CX, y: CY, width: CW, height: CH,
      fill: bodyFill, stroke: bodyStroke, strokeWidth: 1.2 }),
    // main body
    React.createElement('rect', { x: BX, y: BY, width: BW, height: BH, rx: 2,
      fill: bodyFill, stroke: bodyStroke, strokeWidth: 1.8 }),
    // engine / generator divider
    React.createElement('line', { x1: GX, y1: BY, x2: GX, y2: BY+BH,
      stroke: bodyStroke, strokeWidth: 0.8, opacity: 0.4 }),
    // engine vents
    ...vents,
    // coil windings
    ...coils,
    // base feet
    React.createElement('rect', { x: 6, y: BY+BH, width: 14, height: 4, rx: 1,
      fill: bodyStroke, opacity: 0.5 }),
    React.createElement('rect', { x: BW-20, y: BY+BH, width: 14, height: 4, rx: 1,
      fill: bodyStroke, opacity: 0.5 }),
    // indicator light (panel)
    React.createElement('rect', { x: BX+BW-14, y: BY+3, width: 10, height: 8, rx: 1,
      fill: '#161b22', stroke: bodyStroke, strokeWidth: 0.7 }),
    React.createElement('circle', { cx: BX+BW-9, cy: BY+7, r: 2.5,
      fill: lightColor, style: { filter: lightGlow } }),
    // fail cross
    ...failCross,
    // power output terminals (3 small squares)
    ...[0, 10, 20].map((dy, i) =>
      React.createElement('rect', { key: i, x: PX, y: PY+dy-2, width: 4, height: 4, rx: 0.5,
        fill: 'none', stroke: bodyStroke, strokeWidth: 1 })
    ),
    // power lines
    ...powerLines,
  );
}

function RpgActivateSection({ want, isChild, isControl, isFocused }) {
  const state   = want.state?.current || {};
  const phase   = want.status?.phase || '';
  const target  = state.target || want.spec?.params?.target || '';
  const error   = state.error || '';
  const summary = state.summary || '';
  const pct     = state.achieving_percentage || 0;
  const mt      = isChild || (isControl && !isFocused) ? 'mt-2' : 'mt-3';

  const isRunning  = phase === 'achieved';
  const isFailed   = !!error || phase === 'failed';
  const isStarting = !isRunning && !isFailed;

  const genStatus = isRunning ? 'running' : isFailed ? 'failed' : 'starting';

  const statusColor = isRunning ? '#3fb950' : isFailed ? '#f85149' : '#f5e000';
  const statusText  = isRunning ? '⚡ RUNNING' : isFailed ? '✗ FAILED' : `⟳ STARTING${pct > 0 ? ` ${pct}%` : ''}`;

  const deviceLabel = target
    .replace(/_/g, ' ')
    .replace(/\b\w/g, c => c.toUpperCase())
    || 'Device';

  return React.createElement('div', {
    className: `${mt} rounded-lg overflow-hidden`,
    style: { background: '#0d1117', border: '1px solid #30363d' },
  },
    // header bar
    React.createElement('div', {
      className: 'flex items-center px-3 py-1.5',
      style: { background: '#161b22', borderBottom: '1px solid #30363d' },
    },
      React.createElement('span', {
        style: { fontFamily: 'ui-monospace,monospace', fontSize: 11, color: statusColor },
      }, statusText),
      target && React.createElement('span', {
        style: { marginLeft: 'auto', fontFamily: 'ui-monospace,monospace', fontSize: 10, color: '#6e7681' },
      }, deviceLabel),
    ),
    // generator scene
    React.createElement('div', { style: { padding: '10px 12px' } },
      React.createElement(Generator, { status: genStatus }),
    ),
  );
}

window.__mywant.registerPlugin({
  types: ['rpg_activate'],
  ContentSection: RpgActivateSection,
  hideFinalResult: true,
});
