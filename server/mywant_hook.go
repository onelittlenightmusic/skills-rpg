package server

import (
	"log"
	"os/exec"
)

// MywantHookEvent is the payload received from mywant's lifecycle webhook.
type MywantHookEvent struct {
	Event string            `json:"event"`
	Want  MywantWantInfo    `json:"want"`
	Rule  *MywantRuleRef    `json:"rule,omitempty"` // set by filtered rules (rpg_hook want type)
}

// MywantWantInfo holds the metadata of the want in the notification.
type MywantWantInfo struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Status          string                 `json:"status,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	OwnerReferences []MywantOwnerReference `json:"owner_references,omitempty"`
	FinalResult     any                    `json:"final_result,omitempty"`
}

// MywantOwnerReference is the owner reference info in the notification.
type MywantOwnerReference struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	Controller bool   `json:"controller"`
}

// HandleMywantHook processes a lifecycle event from mywant.
// It is called from handleMywantHook under s.mu to guarantee state consistency.
//
// Two processing paths:
//  1. Rule-based (dynamic): event.Rule is set by rpg_hook want type.  Action params
//     (activate, skip_if_achievement, require_achievements, run) come from Rule.Metadata.
//  2. Static (legacy): event.Rule is nil; hooks are matched from stage YAML mywant_hooks.
func (s *Server) HandleMywantHook(event MywantHookEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stage := s.state.Stages[s.state.CurrentStage]
	if stage == nil {
		return
	}
	locale := s.currentLocaleLocked()

	if event.Rule != nil {
		// ── Rule-based path ───────────────────────────────────────────────────
		meta := event.Rule.Metadata
		skipIf, _ := meta["skip_if_achievement"].(string)
		if skipIf != "" && has(s.state.Achievements, skipIf) {
			return
		}
		if reqs, ok := meta["require_achievements"].([]any); ok {
			for _, req := range reqs {
				if r, ok := req.(string); ok && !has(s.state.Achievements, r) {
					return
				}
			}
		}
		activate, _ := meta["activate"].(string)
		if activate != "" {
			s.applyActivateLocked(activate, event.Event, event.Want.Name, locale)
		}
		if cmds, ok := meta["run"].([]any); ok {
			for _, c := range cmds {
				if cmd, ok := c.(string); ok {
					go runShell(cmd)
				}
			}
		}
		return
	}

	// ── Static YAML path (backwards compat) ───────────────────────────────────
	for _, hook := range stage.MywantHooks {
		if hook.Event != "" && hook.Event != event.Event {
			continue
		}
		if !hookMatches(hook, event.Want, s.state.Achievements) {
			continue
		}
		if hook.Activate != "" {
			s.applyActivateLocked(hook.Activate, event.Event, event.Want.Name, locale)
		}
		for _, cmd := range hook.Run {
			go runShell(cmd)
		}
	}
}

// applyActivateLocked fires an activate control action and persists state.
// Must be called with s.mu held.
func (s *Server) applyActivateLocked(target, eventName, wantName string, locale *StageLocale) {
	in := ControlInput{Actor: ActorChap, Action: ActionActivate, Target: target}
	ev, res := applyControl(s.state, in, locale)
	ev.Narration = res.Narration
	s.state.EventHistory = append(s.state.EventHistory, ev)
	const maxHistory = 20
	if len(s.state.EventHistory) > maxHistory {
		s.state.EventHistory = s.state.EventHistory[len(s.state.EventHistory)-maxHistory:]
	}
	if err := s.persistLocked(); err != nil {
		log.Printf("[MYWANT-HOOK] persist error: %v", err)
	}
	log.Printf("[MYWANT-HOOK] activated %q via %s(%s): ok=%v", target, eventName, wantName, res.OK)
}

// hookMatches reports whether the hook rule matches the incoming want event.
func hookMatches(hook MywantHook, want MywantWantInfo, achievements []string) bool {
	if hook.SkipIfAchievement != "" && has(achievements, hook.SkipIfAchievement) {
		return false
	}
	for _, req := range hook.RequireAchievements {
		if !has(achievements, req) {
			return false
		}
	}
	if hook.MatchName != "" && want.Name != hook.MatchName {
		return false
	}
	if hook.MatchType != "" && want.Type != hook.MatchType {
		return false
	}
	for k, v := range hook.MatchLabel {
		if want.Labels[k] != v {
			return false
		}
	}
	if hook.MatchOwner != "" {
		found := false
		for _, ref := range want.OwnerReferences {
			if ref.Controller && ref.Name == hook.MatchOwner {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if hook.MatchChildRole != "" {
		if want.Labels["child-role"] != hook.MatchChildRole {
			return false
		}
	}
	return true
}

func runShell(cmd string) {
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		log.Printf("[MYWANT-HOOK] run %q: %v", cmd, err)
	}
}
