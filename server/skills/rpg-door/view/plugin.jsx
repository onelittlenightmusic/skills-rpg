// RPG Door card — one door, and everything standing between it and open.
//
// The board already draws this door as a tile you can or cannot walk through.
// What the tile cannot say is *why*, and why is the only actionable thing: a
// key chap may or may not be carrying, a generator that is not running, an
// alarm that has not been silenced. So the card draws the door with its
// conditions beside it, each as the thing it actually is.
//
// The whole scene is one viewBox'd SVG rather than laid-out HTML boxes: this
// card is shown at anything from a sidebar strip to a full-screen overlay, and
// a viewBox is the only way the drawing keeps its proportions across that range
// instead of being cropped at the small end.
const React = window.React;

import { Door, Key, Generator, Alarm, PALETTE as P } from '/api/v1/plugins/rpg-artkit.js';

// Scene coordinates. Everything below is in these units.
const VB_W = 216, VB_H = 106;
const DOOR_X = 10, DOOR_Y = 12, DOOR_W = 46, DOOR_H = 82;
const COL_X = 70;            // where the conditions column starts
const MARK_X = VB_W - 8;     // the ✓ / ✗ gutter

/** One condition: the thing itself, its name, and whether it is satisfied. */
function ConditionRow({ y, art, name, met, children }) {
  return React.createElement('g', { transform: `translate(0, ${y})` },
    children,
    React.createElement('text', {
      x: COL_X + 62, y: 12, fontSize: 9, fontFamily: P.mono,
      fill: met ? P.text : P.textDim,
    }, name),
    React.createElement('text', {
      x: MARK_X, y: 12, fontSize: 11, fontFamily: P.mono, textAnchor: 'end',
      fill: met ? P.ok : P.bad, fontWeight: 'bold',
    }, met ? '✓' : '✗'),
  );
}

function RpgDoorSection({ want }) {
  const state = want.state?.current || {};
  const params = want.spec?.params || {};

  const doorId = state.door_id || params.door_id || '';
  const stageId = state.stage_id || params.stage_id || '';
  const isOpen = state.open === true;
  const locked = state.locked !== false;
  const key = state.key || '';
  const keyHeld = state.key_held_by_chap === true;
  const requires = state.requires_device || '';
  const deviceOn = state.device_on === true;
  const blockedBy = state.blocked_by_device || '';
  const blockerOn = state.blocker_on === true;
  const from = state.between_from || '';
  const to = state.between_to || '';
  const summary = state.summary || '';
  // What kind of thing this door answers to, when it answers to one. The value
  // is deliberately not carried here — the door knows it, and the card is where
  // the player is meant to discover that it is MISSING, not to be handed it.
  const wantsThing = state.wants_thing_subtype || '';
  const wantsPinned = state.wants_thing_kind === 'pinned';

  // The device gate, drawn as the machine itself rather than as a word: the
  // same generator and alarm that have their own cards, small, standing here.
  const gate = requires
    ? { art: 'generator', id: requires, met: deviceOn }
    : blockedBy
      ? { art: 'alarm', id: blockedBy, met: !blockerOn }
      : null;

  const statusColor = isOpen ? P.ok : locked ? P.bad : P.textFaint;
  const statusText = isOpen ? '✓ OPEN' : locked ? '🔒 LOCKED' : '○ UNLOCKED';

  // A key chap is not carrying is not a key that is going to be tried.
  const keyStatus = isOpen ? 'success' : keyHeld ? 'idle' : 'failed';

  // Rows are stacked from the top of the column, so a door with only one
  // condition does not leave a hole where the other would have been.
  const rows = [];
  if (key) {
    rows.push({ met: keyHeld, name: key.replace(/^key_/, ''), node: (y) =>
      React.createElement('g', { key: 'key', transform: `translate(${COL_X}, ${y + 2}) scale(0.72)` },
        React.createElement(Key, { status: keyStatus }),
      ),
    });
  }
  // The empty nameplate. A gate scraped clean of its maker's name shows a plate
  // with nothing on it — which is the whole of what fortress1 asks the player to
  // notice, and it cannot be noticed on a card that only ever says "locked".
  if (wantsThing) {
    rows.push({ met: false, name: wantsThing, node: (y) =>
      React.createElement('g', { key: 'plate', transform: `translate(${COL_X - 2}, ${y - 2})` },
        React.createElement('rect', {
          x: 0, y: 0, width: 34, height: 18, rx: 2,
          fill: 'none', stroke: P.bad, strokeWidth: 1.2,
          strokeDasharray: wantsPinned ? '3 2' : undefined,
        }),
        // The scrape: a groove where the name was struck, polished flat.
        React.createElement('line', {
          x1: 5, y1: 9, x2: 29, y2: 9,
          stroke: P.bad, strokeWidth: 1, opacity: 0.45,
        }),
      ),
    });
  }
  if (gate) {
    rows.push({ met: gate.met, name: gate.id, node: (y) =>
      React.createElement('g', { key: 'gate', transform: `translate(${COL_X - 4}, ${y - 8}) scale(0.42)` },
        gate.art === 'alarm'
          ? React.createElement(Alarm, { status: gate.met ? 'silent' : 'ringing' })
          : React.createElement(Generator, { status: gate.met ? 'running' : 'stopped' }),
      ),
    });
  }

  const ROW_H = 34;
  const rowsTop = (VB_H - rows.length * ROW_H) / 2;

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
      }, [stageId, doorId].filter(Boolean).join(' · ')),
    ),
    content: React.createElement('div', {
      style: { height: '100%', width: '100%', overflow: 'hidden', padding: '6px 8px' },
    },
      React.createElement('svg', {
        viewBox: `0 0 ${VB_W} ${VB_H}`,
        width: '100%', height: '100%',
        preserveAspectRatio: 'xMidYMid meet',
        style: { display: 'block' },
      },
        // The door, with the rooms it joins named above and below it.
        React.createElement('text', {
          x: DOOR_X + DOOR_W / 2, y: 8,
          textAnchor: 'middle', fontSize: 7, fontFamily: P.mono, fill: P.textFaint,
        }, from),
        React.createElement('g', { transform: `translate(${DOOR_X}, ${DOOR_Y})` },
          React.createElement(Door, { open: isOpen, locked, width: DOOR_W, height: DOOR_H }),
        ),
        React.createElement('text', {
          x: DOOR_X + DOOR_W / 2, y: VB_H - 2,
          textAnchor: 'middle', fontSize: 7, fontFamily: P.mono, fill: P.textFaint,
        }, to),

        // What the lock is waiting on.
        ...rows.map((r, i) => {
          const y = rowsTop + i * ROW_H;
          return React.createElement(ConditionRow, {
            key: r.name, y, name: r.name, met: r.met,
          }, r.node(0));
        }),
        rows.length === 0 && React.createElement('text', {
          x: COL_X, y: VB_H / 2, fontSize: 9, fontFamily: P.mono, fill: P.muted,
        }, isOpen ? 'nothing in the way' : 'no key, no device'),
      ),
    ),
    bottom: summary && React.createElement('div', {
      className: 'px-3 py-1 text-[10px] leading-snug truncate',
      style: { fontFamily: P.mono, color: P.textDim, borderTop: `1px solid ${P.borderSoft}` },
    }, summary),
  });
}

window.__mywant.registerPlugin({
  types: ['rpg_door'],
  ContentSection: RpgDoorSection,
  hideFinalResult: true,
});
