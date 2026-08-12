// RPG Switch card plugin — JIT loaded from ~/.mywant/custom-types/rpg-switch/view/plugin.jsx
const React = window.React;
const { useState, useEffect } = React;

// The generator line-art lives in the shared kit — rpg_generator draws the same
// machine from live device state, and the two must not drift apart.
import { Generator, PALETTE as P } from '/api/v1/plugins/rpg-artkit.js';

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

  return window.__mywant.createCardLayout({
    centerContent: true,
    content: React.createElement('div', {
      className: 'flex flex-col items-center gap-2',
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
    ),
  });
}

window.__mywant.registerPlugin({
  types: ['rpg_switch'],
  ContentSection: SwitchContentSection,
});

// ── RPG Activate ─────────────────────────────────────────────────────────────
function RpgActivateSection({ want, isChild, isControl, isFocused }) {
  const state   = want.state?.current || {};
  const phase   = want.status?.phase || '';
  const target  = state.target || want.spec?.params?.target || '';
  const error   = state.error || '';
  const summary = state.summary || '';
  const pct     = state.achieving_percentage || 0;

  const isRunning  = phase === 'achieved';
  const isFailed   = !!error || phase === 'failed';

  const genStatus = isRunning ? 'running' : isFailed ? 'failed' : 'starting';

  const statusColor = isRunning ? P.ok : isFailed ? P.bad : P.warn;
  const statusText  = isRunning ? '⚡ RUNNING' : isFailed ? '✗ FAILED' : `⟳ STARTING${pct > 0 ? ` ${pct}%` : ''}`;

  const deviceLabel = target
    .replace(/_/g, ' ')
    .replace(/\b\w/g, c => c.toUpperCase())
    || 'Device';

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
      target && React.createElement('span', {
        style: { marginLeft: 'auto', fontFamily: P.mono, fontSize: 10, color: P.textFaint },
      }, deviceLabel),
    ),
    content: React.createElement('div', { style: { height: '100%', overflow: 'hidden', display: 'flex', justifyContent: 'center', alignItems: 'center', padding: '10px 12px' } },
      React.createElement(Generator, { status: genStatus, label: target }),
    ),
  });
}

window.__mywant.registerPlugin({
  types: ['rpg_activate'],
  ContentSection: RpgActivateSection,
  hideFinalResult: true,
});
