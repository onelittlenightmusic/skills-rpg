// RPG Device card — a generator or an alarm, drawn from its live server state.
//
// The card is not just a readout of one boolean. What makes a device matter in
// this game is the door on the other side of it, so the door is on the card:
// power runs from the machine to the lock, and the lock answers.
const React = window.React;
const { useState } = React;

import { Generator, Alarm, Door, PALETTE as P } from '/api/v1/plugins/rpg-artkit.js';

const ART = { rpg_alarm: 'alarm', rpg_generator: 'generator' };

/** A device that is on does one thing; which thing depends on what it is. */
const VERB = {
  generator: { on: '⚡ RUNNING', off: '○ STOPPED', held: '⛔ BLOCKED' },
  alarm:     { on: '◉ RINGING', off: '○ SILENT',  held: '⛔ BLOCKED' },
};

function DeviceSection({ want, isFocused, isExpanded }) {
  const state = want.state?.current || {};
  const params = want.spec?.params || {};
  const type = want.metadata?.type || 'rpg_generator';
  const art = ART[type] || 'generator';

  const on          = state.on === true;
  const blocked     = state.blocked === true;
  const blockedBy   = state.blocked_by || '';
  const error       = state.error || '';
  const deviceId    = state.device_id || params.device_id || '';
  const label       = state.label || deviceId;
  const stageId     = state.stage_id || params.stage_id || '';
  const stageTitle  = state.stage_title || '';
  const gates       = Array.isArray(state.gates) ? state.gates : [];
  const controllable = state.controllable === true || params.controllable === true;

  // The want holds an intent; the server holds the fact. While they disagree the
  // card is mid-flight — either the poll has not run yet or the device is held.
  const desired = state.webhook_payload?.action || '';
  const inFlight = desired !== '' && desired !== (on ? 'on' : 'off') && !blocked && !error;

  // The click is optimistic only until the next poll answers, ~2s away.
  const [sent, setSent] = useState(null);
  const pending = inFlight || sent !== null;
  const shown = sent !== null ? sent : on;

  const request = async (next) => {
    const id = want.metadata?.id;
    if (!id || !controllable) return;
    setSent(next);
    try {
      const resp = await fetch(`/api/v1/webhooks/${id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action: next ? 'on' : 'off' }),
      });
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    } catch (err) {
      console.error('[RpgDevice] request failed:', err);
      setSent(null);
      return;
    }
    // Hand back to the server's answer once a poll has had time to land.
    setTimeout(() => setSent(null), 3000);
  };

  // Art status. "blocked" outranks off: a device that cannot be started is a
  // different problem from one nobody has started yet.
  const artStatus = error ? 'failed'
    : pending ? (art === 'alarm' ? (shown ? 'ringing' : 'stopping') : (shown ? 'running' : 'starting'))
    : shown ? (art === 'alarm' ? 'ringing' : 'running')
    : blocked ? 'blocked'
    : (art === 'alarm' ? 'silent' : 'stopped');

  const verb = VERB[art];
  const statusText = error ? '✗ ERROR'
    : pending ? '⟳ …'
    : shown ? verb.on
    : blocked ? verb.held
    : verb.off;
  const statusColor = error ? P.bad : shown ? (art === 'alarm' ? P.bad : P.ok) : blocked ? P.warnDim : P.textDim;

  const Art = art === 'alarm' ? Alarm : Generator;

  // The gated doors, small, to the right of the machine. A generator's doors
  // open when it runs; an alarm's doors open when it stops — so "what this
  // device is doing to that door right now" is the same question either way,
  // answered by the door's own locked/open state.
  const DOOR_W = 26, DOOR_H = 47, DOOR_GAP = 10;
  const doorStrip = gates.length > 0 && React.createElement('svg', {
    width: gates.length * (DOOR_W + DOOR_GAP), height: DOOR_H + 13,
    style: { display: 'block', overflow: 'visible', flexShrink: 0 },
  },
    ...gates.map((g, i) => React.createElement('g', {
      key: g.id, transform: `translate(${i * (DOOR_W + DOOR_GAP)}, 0)`,
    },
      React.createElement(Door, { open: g.open, locked: g.locked, width: DOOR_W, height: DOOR_H }),
      React.createElement('text', {
        x: DOOR_W / 2, y: DOOR_H + 10,
        textAnchor: 'middle', fontSize: 7, fontFamily: P.mono,
        fill: g.open ? P.ok : g.locked ? P.bad : P.textFaint,
      }, g.id.replace(/_/g, ' ')),
    )),
  );

  const held = blocked && blockedBy;

  // What a registry reader is waiting for, in the two states that are two
  // different jobs. Every other device is switched by chap and has nothing to
  // say here; this one cannot be switched at all, so if the card does not say
  // what the city is missing, nothing does.
  const readsKind   = state.reads_subtype || '';
  const readsValue  = state.reads_value || '';
  const needsPin    = state.reads_pinned === true;
  const isNamed     = state.reads_named === true;
  const isPinned    = state.reads_is_pinned === true;
  const needText = !readsKind ? ''
    : !isNamed  ? `add a ${readsKind} called ${readsValue} to the registry`
    : needsPin && !isPinned ? `${readsValue} is in the registry — pin it`
    : '';

  return window.__mywant.createCardLayout({
    className: 'rounded-lg overflow-hidden',
    style: { background: P.bg, border: `1px solid ${P.border}` },
    top: React.createElement('div', {
      className: 'flex items-center px-3 py-1.5',
      style: { background: P.bgHeader, borderBottom: `1px solid ${P.border}` },
    },
      React.createElement('span', {
        style: { fontFamily: P.mono, fontSize: 11, color: statusColor },
      }, statusText),
      React.createElement('span', {
        style: { marginLeft: 'auto', fontFamily: P.mono, fontSize: 10, color: P.textFaint },
      }, stageId + (stageTitle ? ` · ${stageTitle}` : '')),
    ),
    content: React.createElement('div', {
      style: {
        height: '100%', overflow: 'hidden', display: 'flex',
        alignItems: 'center', justifyContent: 'center', gap: 4, padding: '8px 10px',
      },
      onMouseDown: (e) => e.stopPropagation(),
      onTouchStart: (e) => e.stopPropagation(),
    },
      React.createElement(controllable ? 'button' : 'div', {
        onClick: controllable ? (e) => { e.stopPropagation(); request(!shown); } : undefined,
        onMouseDown: controllable ? (e) => e.stopPropagation() : undefined,
        disabled: controllable ? pending : undefined,
        title: controllable ? (shown ? `chap: deactivate ${deviceId}` : `chap: activate ${deviceId}`) : undefined,
        style: {
          background: 'none', border: 'none', padding: 0, flexShrink: 0,
          cursor: controllable ? 'pointer' : 'default',
          opacity: pending ? 0.7 : 1,
          outline: controllable && isFocused ? `1px dashed ${P.border}` : 'none',
        },
      },
        React.createElement(Art, { status: artStatus, label }),
      ),
      readsKind && React.createElement('div', {
        style: {
          display: 'flex', flexDirection: 'column', gap: 3, flexShrink: 0,
          fontFamily: P.mono, fontSize: 8, lineHeight: 1.3,
        },
      },
        React.createElement('div', { style: { color: isNamed ? P.ok : P.bad } },
          `${isNamed ? '✓' : '✗'} ${readsKind}: ${readsValue}`),
        needsPin && React.createElement('div', { style: { color: isPinned ? P.ok : P.bad } },
          `${isPinned ? '✓' : '✗'} pinned`),
      ),
      doorStrip,
    ),
    bottom: (held || error || needText) && React.createElement('div', {
      className: 'px-3 py-1 text-[10px] leading-snug',
      style: {
        fontFamily: P.mono,
        color: error ? P.bad : needText ? P.bad : P.warnDim,
        borderTop: `1px solid ${P.borderSoft}`,
      },
    }, error || needText || `held by ${blockedBy} — stop it first`),
  });
}

window.__mywant.registerPlugin({
  types: ['rpg_generator', 'rpg_alarm'],
  ContentSection: DeviceSection,
  hideFinalResult: true,
});
