// RPG art kit — the shared line-art vocabulary for every rpg_* want card.
//
// This module registers no plugin. It only exports drawing primitives, so that
// a door drawn on the try-keys card and a door drawn on the door card are the
// same door. Consumers import it by absolute URL:
//
//   import { Door, Generator, PALETTE } from '/api/v1/plugins/rpg-artkit.js';
//
// Absolute, not relative: the server compiles each plugin file on its own with
// esbuild's Transform (no bundling), so import specifiers survive into the
// browser and are resolved against the page, not against the plugin's folder.
const React = window.React;

// ── Palette ───────────────────────────────────────────────────────────────────
// GitHub-dark, the tone the first rpg card was drawn in.
export const PALETTE = {
  bg:        '#0d1117',
  bgHeader:  '#161b22',
  bgSunk:    '#030a04',
  border:    '#30363d',
  borderSoft:'#21262d',
  text:      '#cdd9e5',
  textDim:   '#8b949e',
  textFaint: '#6e7681',
  muted:     '#484f58',
  ok:        '#3fb950',
  okDeep:    '#238636',
  okBg:      '#041a06',
  bad:       '#f85149',
  badBg:     '#180505',
  warn:      '#f5e000',
  warnDim:   '#8b7020',
  gold:      '#c9971a',
  goldDim:   '#8b7020',
  info:      '#58a6ff',
  infoDeep:  '#388bfd',
  amber:     '#fbbf24',
  mono:      'ui-monospace,monospace',
};

const P = PALETTE;

// ── Door ──────────────────────────────────────────────────────────────────────
/**
 * The full-size door: a panel with a keyhole, or a panel swung open on a dark
 * doorway. Drawn into a `width` × `height` box with its own origin at 0,0, so
 * the caller places it with a transform.
 *
 * Three states, because a door has three: locked (red, padlocked), closed but
 * unlocked (grey — it will let you through, it just is not through yet), and
 * open (green, with the room beyond showing).
 */
export function Door({ open, locked, width = 50, height = 90 }) {
  const W = width, H = height;
  const s = W / 50;  // everything below was drawn at W=50

  if (open) {
    return React.createElement('g', null,
      React.createElement('rect', { x: 0, y: 0, width: W, height: H, rx: 3 * s,
        fill: 'none', stroke: P.border, strokeWidth: 1.5 * s }),
      // The leaf, swung towards the viewer — a trapezoid is the whole trick.
      React.createElement('path', {
        d: `M ${2*s} ${2*s} L ${18*s} ${9*s} L ${18*s} ${H-9*s} L ${2*s} ${H-2*s} Z`,
        fill: 'none', stroke: P.ok, strokeWidth: 2 * s,
      }),
      React.createElement('rect', { x: 18*s, y: 2*s, width: W-20*s, height: H-4*s, fill: P.bgSunk }),
      React.createElement('rect', { x: 18*s, y: 2*s, width: W-20*s, height: H-4*s,
        fill: 'none', stroke: P.okDeep, strokeWidth: 0.7 * s, strokeDasharray: '3 3', opacity: 0.5 }),
      React.createElement('circle', { cx: 34*s, cy: H/2, r: 3*s, fill: P.ok, opacity: 0.75 }),
    );
  }

  const sc = locked ? P.bad : P.textFaint;

  return React.createElement('g', null,
    React.createElement('rect', { x: 0, y: 0, width: W, height: H, rx: 3 * s,
      fill: locked ? P.badBg : P.bg, stroke: sc, strokeWidth: 2 * s }),
    React.createElement('rect', { x: 6*s, y: 6*s, width: W-12*s, height: H-12*s, rx: 2*s,
      fill: 'none', stroke: sc, strokeWidth: 0.6 * s, opacity: 0.3 }),
    // Keyhole: circle over a tapering slot.
    React.createElement('circle', { cx: W/2, cy: H*0.42, r: 5.5*s, fill: 'none', stroke: sc, strokeWidth: 1.5*s }),
    React.createElement('path', {
      d: `M ${W/2-3.5*s} ${H*0.42+4*s} L ${W/2} ${H*0.42+16*s} L ${W/2+3.5*s} ${H*0.42+4*s}`,
      fill: 'none', stroke: sc, strokeWidth: 1.5*s, strokeLinejoin: 'round',
    }),
    React.createElement('circle', { cx: W-11*s, cy: H/2, r: 4*s, fill: 'none', stroke: sc, strokeWidth: 1.5*s }),
    locked && React.createElement('g', null,
      React.createElement('rect', { x: W/2-7*s, y: H-22*s, width: 14*s, height: 10*s, rx: 2*s,
        fill: 'none', stroke: P.bad, strokeWidth: 1.2*s }),
      React.createElement('path', {
        d: `M ${W/2-4*s} ${H-22*s} L ${W/2-4*s} ${H-27*s} Q ${W/2} ${H-31*s} ${W/2+4*s} ${H-27*s} L ${W/2+4*s} ${H-22*s}`,
        fill: 'none', stroke: P.bad, strokeWidth: 1.2*s,
      }),
    ),
  );
}

/**
 * The door reduced to a mark on a map: a slab, or a slab swung aside. Used
 * where a door is one node in a diagram rather than the subject of the card.
 */
export function DoorGlyph({ cx, cy, open, locked, scale = 1 }) {
  const s = scale;
  if (open) {
    return React.createElement('rect', {
      x: cx - 3*s, y: cy - 7.5*s, width: 6*s, height: 13.5*s, rx: 1.5*s,
      fill: P.okDeep, transform: `rotate(-55,${cx},${cy})`,
    });
  }
  return React.createElement('rect', {
    x: cx - 3.75*s, y: cy - 9*s, width: 7.5*s, height: 18*s, rx: 1.5*s,
    fill: locked ? '#3d1f1f' : '#2d333b',
    stroke: locked ? P.bad : P.textFaint,
    strokeWidth: 1.2*s,
  });
}

// ── Key ───────────────────────────────────────────────────────────────────────
/**
 * The full-size key, drawn lying on its side in a 72×18 box.
 * status: 'idle' | 'trying' | 'failed' | 'success'
 */
export function Key({ status }) {
  const strokeOf = {
    idle:    P.gold,
    trying:  P.warn,
    failed:  '#7a3535',
    success: P.ok,
  };
  const sc  = strokeOf[status] || strokeOf.idle;
  const dim = status === 'failed';

  const R = 7, HR = 3, SX = R * 2 + 1, SL = 40, SY = R;

  return React.createElement('g', { opacity: dim ? 0.45 : 1 },
    React.createElement('circle', { cx: R, cy: SY, r: R,  fill: 'none', stroke: sc, strokeWidth: 1.5 }),
    React.createElement('circle', { cx: R, cy: SY, r: HR, fill: 'none', stroke: sc, strokeWidth: 1 }),
    React.createElement('line', { x1: SX, y1: SY, x2: SX+SL, y2: SY, stroke: sc, strokeWidth: 2 }),
    // Teeth: three ticks of unequal length, which is what stops it reading as a pin.
    React.createElement('line', { x1: SX+SL-14, y1: SY, x2: SX+SL-14, y2: SY+6, stroke: sc, strokeWidth: 2 }),
    React.createElement('line', { x1: SX+SL-7,  y1: SY, x2: SX+SL-7,  y2: SY+4, stroke: sc, strokeWidth: 2 }),
    React.createElement('line', { x1: SX+SL,    y1: SY, x2: SX+SL,    y2: SY+7, stroke: sc, strokeWidth: 2 }),
    status === 'failed' && React.createElement('line', {
      x1: 0, y1: 0, x2: SX+SL+2, y2: SY*2, stroke: P.bad, strokeWidth: 1, strokeDasharray: '3 2',
    }),
    status === 'failed' && React.createElement('line', {
      x1: SX+SL+2, y1: 0, x2: 0, y2: SY*2, stroke: P.bad, strokeWidth: 1, strokeDasharray: '3 2',
    }),
    status === 'success' && React.createElement('text', {
      x: SX+SL+5, y: SY+5, fill: P.ok, fontSize: 12, fontWeight: 'bold', fontFamily: P.mono,
    }, '✓'),
    status === 'trying' && React.createElement('circle', {
      cx: SX+SL+7, cy: SY, r: 3.5, fill: P.warn, opacity: 0.9,
    }),
  );
}

/** The key as an inventory item: small, labelled underneath, dimmed once spent. */
export function KeyGlyph({ x, y, label, used }) {
  const sc = used ? '#4a5568' : P.gold;
  const R = 3.5, HR = 1.4, SX = 8, SL = 18, SY = 3.5;
  const short = String(label).replace(/^key_/, '').slice(0, 7);
  return React.createElement('g', { transform: `translate(${x},${y})`, opacity: used ? 0.45 : 1 },
    React.createElement('circle', { cx: R, cy: SY, r: R,  fill: 'none', stroke: sc, strokeWidth: 1 }),
    React.createElement('circle', { cx: R, cy: SY, r: HR, fill: 'none', stroke: sc, strokeWidth: 0.7 }),
    React.createElement('line', { x1: SX, y1: SY, x2: SX+SL, y2: SY, stroke: sc, strokeWidth: 1.4 }),
    React.createElement('line', { x1: SX+SL-8, y1: SY, x2: SX+SL-8, y2: SY+4, stroke: sc, strokeWidth: 1.4 }),
    React.createElement('line', { x1: SX+SL-4, y1: SY, x2: SX+SL-4, y2: SY+3, stroke: sc, strokeWidth: 1.4 }),
    React.createElement('line', { x1: SX+SL,   y1: SY, x2: SX+SL,   y2: SY+5, stroke: sc, strokeWidth: 1.4 }),
    React.createElement('text', {
      x: (SX + SX+SL) / 2, y: SY*2 + 9,
      textAnchor: 'middle', fontSize: 6, fontFamily: P.mono, fill: sc,
    }, short),
  );
}

// ── The two who are in the world ──────────────────────────────────────────────
/** you — a stick figure, because you are the one without special powers. */
export function You({ cx, cy, scale = 1 }) {
  const s = scale;
  const p = { stroke: P.info, strokeWidth: 1.8 * s };
  return React.createElement('g', null,
    React.createElement('circle', { cx, cy, r: 5.25*s, fill: 'none', ...p }),
    React.createElement('line', { x1: cx,        y1: cy+5.25*s, x2: cx,        y2: cy+18*s,   ...p }),
    React.createElement('line', { x1: cx-7.5*s,  y1: cy+10.5*s, x2: cx+7.5*s,  y2: cy+10.5*s, ...p }),
    React.createElement('line', { x1: cx,        y1: cy+18*s,   x2: cx-6*s,    y2: cy+27*s,   ...p }),
    React.createElement('line', { x1: cx,        y1: cy+18*s,   x2: cx+6*s,    y2: cy+27*s,   ...p }),
  );
}

/** chap — a green head with two lit eyes. The agent, watching. */
export function ChapHead({ cx, cy, r = 5.25 }) {
  return React.createElement('g', null,
    React.createElement('circle', { cx, cy, r, fill: '#132d1a', stroke: P.ok, strokeWidth: 1.5 }),
    React.createElement('ellipse', { cx: cx - 2, cy: cy - 0.5, rx: 1.1, ry: 1.9, fill: P.ok }),
    React.createElement('ellipse', { cx: cx + 2, cy: cy - 0.5, rx: 1.1, ry: 1.9, fill: P.ok }),
  );
}

// ── Devices ───────────────────────────────────────────────────────────────────
const DEVICE_TONE = {
  running:  { stroke: P.ok,        fill: P.okBg,  lamp: P.ok,      glow: `0 0 6px ${P.ok}` },
  starting: { stroke: P.warn,      fill: P.bg,    lamp: P.warn,    glow: `0 0 5px ${P.warn}` },
  stopped:  { stroke: P.textFaint, fill: P.bg,    lamp: P.muted,   glow: 'none' },
  blocked:  { stroke: P.warnDim,   fill: P.bg,    lamp: P.warnDim, glow: 'none' },
  failed:   { stroke: P.bad,       fill: '#1a0404', lamp: P.bad,   glow: `0 0 6px ${P.bad}` },
};

/** A padlock badge, for a device something else is holding shut. */
function BlockedBadge({ x, y }) {
  return React.createElement('g', null,
    React.createElement('rect', { x, y: y + 5, width: 11, height: 8, rx: 1.5,
      fill: P.bg, stroke: P.bad, strokeWidth: 1.1 }),
    React.createElement('path', {
      d: `M ${x+3} ${y+5} L ${x+3} ${y+2} Q ${x+5.5} ${y-1.5} ${x+8} ${y+2} L ${x+8} ${y+5}`,
      fill: 'none', stroke: P.bad, strokeWidth: 1.1,
    }),
  );
}

/**
 * The generator: engine block, coil housing, chimney, and three terminals that
 * throw power to the right when it runs. The power lines are the point — they
 * are what connects a running machine to a door that opens because of it.
 *
 * status: 'running' | 'starting' | 'stopped' | 'blocked' | 'failed'
 */
export function Generator({ status, label }) {
  const tone = DEVICE_TONE[status] || DEVICE_TONE.stopped;
  const isRunning  = status === 'running';
  const isStarting = status === 'starting';
  const isFailed   = status === 'failed';
  const isBlocked  = status === 'blocked';

  const BX = 0, BY = 18, BW = 85, BH = 38;
  const EX = 4, EW = 36;
  const GX = 44, GY = 22, GH = 30;
  const CX = 14, CY = 5, CW = 8, CH = 14;
  const PX = BX + BW, PY = BY + 6;

  const vents = [26, 30, 34].map(y =>
    React.createElement('line', { key: y, x1: EX+4, y1: y, x2: EX+EW-6, y2: y,
      stroke: tone.stroke, strokeWidth: 0.6, opacity: 0.5 })
  );

  const coils = [48, 55, 62].map(x =>
    React.createElement('rect', { key: x, x, y: GY+4, width: 4, height: GH-8, rx: 1,
      fill: 'none', stroke: tone.stroke, strokeWidth: 0.8, opacity: 0.6 })
  );

  const smoke = (isRunning || isStarting) ? [
    { cx: CX+4, cy: 3,  r: 2.5, op: 0.5 },
    { cx: CX+6, cy: -2, r: 2,   op: 0.3 },
  ].map((sm, i) =>
    React.createElement('circle', { key: i, cx: sm.cx, cy: sm.cy, r: sm.r,
      fill: isRunning ? P.okDeep : P.warnDim, opacity: sm.op })
  ) : [];

  // Running: the current leaves in waves. Starting: it only just reaches out.
  const powerLines = (isRunning || isStarting) ? [0, 10, 20].map((dy, i) => {
    const y = PY + dy;
    const len = isStarting ? 12 : 22;
    const d = isRunning ? `M ${PX} ${y} c 5 -3 10 3 ${len} 0` : `M ${PX} ${y} l ${len} 0`;
    return React.createElement('path', { key: i, d,
      fill: 'none', stroke: isRunning ? P.warn : P.warnDim,
      strokeWidth: 1.5, opacity: isStarting ? 0.5 : 0.9 });
  }) : [];

  const failCross = isFailed ? [
    React.createElement('line', { key: 'x1', x1: BX+4, y1: BY+4, x2: BX+BW-4, y2: BY+BH-4,
      stroke: P.bad, strokeWidth: 1.2, strokeDasharray: '4 2' }),
    React.createElement('line', { key: 'x2', x1: BX+BW-4, y1: BY+4, x2: BX+4, y2: BY+BH-4,
      stroke: P.bad, strokeWidth: 1.2, strokeDasharray: '4 2' }),
  ] : [];

  return React.createElement('svg', {
    width: 132, height: 78, style: { display: 'block', overflow: 'visible' },
  },
    ...smoke,
    React.createElement('rect', { x: CX, y: CY, width: CW, height: CH,
      fill: tone.fill, stroke: tone.stroke, strokeWidth: 1.2 }),
    React.createElement('rect', { x: BX, y: BY, width: BW, height: BH, rx: 2,
      fill: tone.fill, stroke: tone.stroke, strokeWidth: 1.8 }),
    React.createElement('line', { x1: GX, y1: BY, x2: GX, y2: BY+BH,
      stroke: tone.stroke, strokeWidth: 0.8, opacity: 0.4 }),
    ...vents,
    ...coils,
    React.createElement('rect', { x: 6, y: BY+BH, width: 14, height: 4, rx: 1, fill: tone.stroke, opacity: 0.5 }),
    React.createElement('rect', { x: BW-20, y: BY+BH, width: 14, height: 4, rx: 1, fill: tone.stroke, opacity: 0.5 }),
    React.createElement('rect', { x: BX+BW-14, y: BY+3, width: 10, height: 8, rx: 1,
      fill: P.bgHeader, stroke: tone.stroke, strokeWidth: 0.7 }),
    React.createElement('circle', { cx: BX+BW-9, cy: BY+7, r: 2.5,
      fill: tone.lamp, style: { filter: tone.glow } }),
    ...failCross,
    ...[0, 10, 20].map((dy, i) =>
      React.createElement('rect', { key: i, x: PX, y: PY+dy-2, width: 4, height: 4, rx: 0.5,
        fill: 'none', stroke: tone.stroke, strokeWidth: 1 })
    ),
    ...powerLines,
    isBlocked && React.createElement(BlockedBadge, { x: BX + 30, y: BY + 12 }),
    label && React.createElement('text', {
      x: BX + BW / 2, y: BY + BH + 18,
      textAnchor: 'middle', fontSize: 8, fontFamily: P.mono, fill: tone.stroke,
    }, label),
  );
}

/**
 * The alarm: a beacon on a mount, ringing out in arcs. Silence is drawn as the
 * absence of the arcs, so an alarm that has been stopped reads at a glance as a
 * thing that is no longer doing anything.
 *
 * status: 'ringing' | 'silent' | 'stopping' | 'blocked' | 'failed'
 */
export function Alarm({ status, label }) {
  const isRinging  = status === 'ringing';
  const isStopping = status === 'stopping';
  const isFailed   = status === 'failed';
  const isBlocked  = status === 'blocked';

  const stroke = isRinging ? P.bad : isFailed ? P.bad : isStopping ? P.warn : P.textFaint;
  const fill   = isRinging ? P.badBg : isFailed ? '#1a0404' : P.bg;
  const lamp   = isRinging ? P.bad : isStopping ? P.warn : P.muted;
  const glow   = isRinging ? `0 0 8px ${P.bad}` : isStopping ? `0 0 5px ${P.warn}` : 'none';

  const CXc = 66, CYc = 52, DOME_R = 18;

  // Arcs sweeping over the beacon, 200°→340° through straight up.
  const wave = (r) => {
    const a1 = (200 * Math.PI) / 180, a2 = (340 * Math.PI) / 180;
    const x1 = CXc + r * Math.cos(a1), y1 = CYc + r * Math.sin(a1);
    const x2 = CXc + r * Math.cos(a2), y2 = CYc + r * Math.sin(a2);
    return `M ${x1.toFixed(1)} ${y1.toFixed(1)} A ${r} ${r} 0 0 1 ${x2.toFixed(1)} ${y2.toFixed(1)}`;
  };

  const waves = isRinging
    ? [26, 34, 42].map((r, i) =>
        React.createElement('path', { key: r, d: wave(r), fill: 'none',
          stroke: P.bad, strokeWidth: 1.6 - i * 0.35, opacity: 0.75 - i * 0.2 }))
    : isStopping
      ? [React.createElement('path', { key: 'w', d: wave(26), fill: 'none',
          stroke: P.warnDim, strokeWidth: 1.2, strokeDasharray: '4 3', opacity: 0.5 })]
      : [];

  return React.createElement('svg', {
    width: 132, height: 78, style: { display: 'block', overflow: 'visible' },
  },
    ...waves,
    // Beacon dome
    React.createElement('path', {
      d: `M ${CXc-DOME_R} ${CYc} A ${DOME_R} ${DOME_R} 0 0 1 ${CXc+DOME_R} ${CYc} Z`,
      fill, stroke, strokeWidth: 1.8,
    }),
    // Lamp inside the dome
    React.createElement('circle', { cx: CXc, cy: CYc - 6, r: 6, fill: lamp, opacity: isRinging ? 0.9 : 0.6,
      style: { filter: glow } }),
    // Mount
    React.createElement('rect', { x: CXc-24, y: CYc, width: 48, height: 9, rx: 1.5,
      fill, stroke, strokeWidth: 1.6 }),
    React.createElement('rect', { x: CXc-9, y: CYc+9, width: 18, height: 5, rx: 1,
      fill: stroke, opacity: 0.5 }),
    // Grille slits on the mount
    ...[-14, -5, 4, 13].map(dx =>
      React.createElement('line', { key: dx, x1: CXc+dx, y1: CYc+2.5, x2: CXc+dx, y2: CYc+6.5,
        stroke, strokeWidth: 0.7, opacity: 0.5 })
    ),
    isFailed && React.createElement('g', null,
      React.createElement('line', { x1: CXc-22, y1: CYc-20, x2: CXc+22, y2: CYc+6,
        stroke: P.bad, strokeWidth: 1.2, strokeDasharray: '4 2' }),
      React.createElement('line', { x1: CXc+22, y1: CYc-20, x2: CXc-22, y2: CYc+6,
        stroke: P.bad, strokeWidth: 1.2, strokeDasharray: '4 2' }),
    ),
    isBlocked && React.createElement(BlockedBadge, { x: CXc - 5.5, y: CYc - 14 }),
    label && React.createElement('text', {
      x: CXc, y: CYc + 28, textAnchor: 'middle', fontSize: 8, fontFamily: P.mono, fill: stroke,
    }, label),
  );
}
