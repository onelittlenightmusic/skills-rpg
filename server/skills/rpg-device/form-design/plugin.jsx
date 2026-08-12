// Lamp tile surface — the stack light drawn on an rpg_generator / rpg_alarm tile.
//
// It lives with the want type rather than in the GUI because it means nothing
// without one: `on` is a field only these two types have. A want type that
// carries its own card in view/ carries its own tile here, and installing the
// type installs the look.
//
// Loaded from ~/.mywant/custom-types/rpg-device/form-design/plugin.jsx.
const React = window.React;

/**
 * A stack light: one window, split green over red, exactly one half lit.
 *
 * The lit half is the whole message, so everything else is built to stay out of
 * its way. The two halves share a single rounded window — one clip path, filled
 * twice — rather than being two rounded rectangles stacked up. That distinction
 * is the difference between reading as one piece of hardware and reading as a
 * list of bars: separate rounded corners meeting in the middle put a light seam
 * across the body at every join, and the eye counts those seams as rows.
 *
 * For the same reason there is one divider and no lens ribs. A rib is a
 * horizontal line, and at 40px tall the eye cannot tell a rib from a join.
 *
 * Both halves are always drawn, and the unlit one stays a visible dark lens: a
 * lamp that vanishes when off reads as decoration that comes and goes, not as a
 * state.
 *
 * The window runs nearly the full width of the tile. It competes with the type
 * icon, which is drawn over the middle of every tile at roughly two thirds its
 * size — a narrow tower would sit entirely behind it, and a lamp you cannot see
 * is not reporting anything. Full width leaves lit colour showing on both sides
 * of the icon at every zoom.
 */
function Lamp({ want, width, height, color }) {
  // Read straight off the want: the tile reports what the device is, never what
  // a click meant it to be.
  const on = want.state?.current?.on === true;

  const pad = Math.max(1.5, Math.min(width, height) * 0.11);  // shell thickness
  const winX = pad, winY = pad;
  const winW = width - pad * 2, winH = height - pad * 2;
  const winR = Math.max(1, Math.min(width, height) * 0.1);
  const mid = winY + winH / 2;
  const divider = Math.max(1, height * 0.035);

  const LIT = { green: '#22c55e', red: '#ef4444' };
  const DARK = { green: '#0e2e1b', red: '#2f0e11' };

  // Stable per-size id: two tiles of the same size share one clip path, and a
  // random id would defeat that for no gain.
  const clipId = `lampwin-${Math.round(width)}-${Math.round(height)}`;

  const half = (y, h, hue, isLit) => (
    <g key={hue}>
      <rect x={winX} y={y} width={winW} height={h} fill={isLit ? LIT[hue] : DARK[hue]} />
      {/* The lit band: brightest just under the top of the half, which is where
          a lamp sitting behind a lens actually puts it. Drawn inside the clip,
          so it stops at the window edge like everything else. */}
      {isLit && (
        <rect x={winX} y={y + h * 0.12} width={winW} height={h * 0.16}
          fill="#fff" fillOpacity="0.28" />
      )}
      {/* Falloff towards the far end of the half, so the colour has a direction
          and the two halves cannot be mistaken for flat swatches. */}
      <rect x={winX} y={hue === 'green' ? y + h * 0.55 : y} width={winW} height={h * 0.45}
        fill="#000" fillOpacity={isLit ? 0.14 : 0.3} />
    </g>
  );

  return (
    <svg width={width} height={height}
      style={{ position: 'absolute', inset: 0, pointerEvents: 'none', borderRadius: 'inherit' }}
      aria-hidden="true">
      <defs>
        <clipPath id={clipId}>
          <rect x={winX} y={winY} width={winW} height={winH} rx={winR} />
        </clipPath>
      </defs>

      {/* The body: one shell, one colour, no internal edges. */}
      <rect width={width} height={height} rx="6" fill={color} />
      <rect width={width} height={height} rx="6" fill="#000" fillOpacity="0.62" />

      <g clipPath={`url(#${clipId})`}>
        {half(winY, winH / 2, 'green', on)}
        {half(mid, winH / 2, 'red', !on)}
        {/* One divider, on the centre line — the only horizontal in the piece. */}
        <rect x={winX} y={mid - divider / 2} width={winW} height={divider}
          fill="#000" fillOpacity="0.7" />
        {/* Glow, painted over its own half and clipped to the window: a soft
            overscan rather than a blur filter, which would be re-rasterised per
            tile per zoom and there are many tiles. */}
        {on
          ? <rect x={winX} y={winY} width={winW} height={winH / 2} fill={LIT.green} fillOpacity="0.2" />
          : <rect x={winX} y={mid} width={winW} height={winH / 2} fill={LIT.red} fillOpacity="0.2" />}
      </g>

      {/* The window's own rim, drawn last so the shell reads as sitting proud of
          the glass rather than behind it. */}
      <rect x={winX} y={winY} width={winW} height={winH} rx={winR}
        fill="none" stroke="#000" strokeOpacity="0.55"
        strokeWidth={Math.max(0.8, pad * 0.35)} />
    </svg>
  );
}

window.__mywant.registerTileSurface({
  id: 'lamp',
  render: (ctx) => React.createElement(Lamp, ctx),
});
