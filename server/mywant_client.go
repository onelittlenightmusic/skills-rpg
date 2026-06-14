package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// ── mywant API client ────────────────────────────────────────────────────────

var mywantHTTP = &http.Client{Timeout: 10 * time.Second}

func mywantGet(base, path string) ([]byte, error) {
	resp, err := mywantHTTP.Get(base + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: %d %s", path, resp.StatusCode, body)
	}
	return body, nil
}

func mywantPost(base, path string, payload any) ([]byte, error) {
	b, _ := json.Marshal(payload)
	resp, err := mywantHTTP.Post(base+path, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("POST %s: %d %s", path, resp.StatusCode, body)
	}
	return body, nil
}

func mywantDelete(base, path string, payload any) error {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodDelete, base+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := mywantHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DELETE %s: %d %s", path, resp.StatusCode, body)
	}
	return nil
}

// ── Lifecycle webhook registration ───────────────────────────────────────────

// ensureLifecycleWebhook removes any duplicate registration for hookURL and
// registers a fresh one.  Returns the new webhook ID.
func ensureLifecycleWebhook(mywantURL, hookURL string) (string, error) {
	// List existing webhooks
	body, err := mywantGet(mywantURL, "/api/v1/lifecycle-webhooks")
	if err != nil {
		return "", fmt.Errorf("list lifecycle-webhooks: %w", err)
	}
	var resp struct {
		Webhooks []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"webhooks"`
	}
	_ = json.Unmarshal(body, &resp)

	// Remove any existing registration for this URL
	for _, wh := range resp.Webhooks {
		if wh.URL == hookURL {
			req, _ := http.NewRequest(http.MethodDelete,
				mywantURL+"/api/v1/lifecycle-webhooks/"+wh.ID, nil)
			r, _ := mywantHTTP.Do(req)
			if r != nil {
				r.Body.Close()
			}
		}
	}

	// Register fresh
	regBody, err := mywantPost(mywantURL, "/api/v1/lifecycle-webhooks",
		map[string]string{"url": hookURL})
	if err != nil {
		return "", fmt.Errorf("register lifecycle-webhook: %w", err)
	}
	var reg struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(regBody, &reg)
	return reg.ID, nil
}

// ── Want cleanup and creation ─────────────────────────────────────────────────

func cleanupMywantWants(mywantURL string, labels map[string]string) error {
	// Build label query
	query := "/api/v1/wants?"
	for k, v := range labels {
		query += "label=" + k + "=" + v + "&"
	}
	body, err := mywantGet(mywantURL, query)
	if err != nil {
		return fmt.Errorf("list wants: %w", err)
	}
	var resp struct {
		Wants []struct {
			Metadata struct {
				ID string `json:"id"`
			} `json:"metadata"`
		} `json:"wants"`
	}
	if err := json.Unmarshal(body, &resp); err != nil || len(resp.Wants) == 0 {
		return nil
	}
	ids := make([]string, 0, len(resp.Wants))
	for _, w := range resp.Wants {
		ids = append(ids, w.Metadata.ID)
	}
	log.Printf("[MYWANT-SETUP] cleaning up %d existing wants", len(ids))
	return mywantDelete(mywantURL, "/api/v1/wants", map[string]any{"ids": ids})
}

func createMywantWant(mywantURL string, w MywantSetupWant, nameToID map[string]string) (string, error) {
	labels := make(map[string]string)
	for k, v := range w.Labels {
		labels[k] = v
	}

	var ownerRefs []map[string]any
	if w.Owner != "" {
		ownerID := nameToID[w.Owner]
		ownerRefs = []map[string]any{{
			"apiVersion":         "mywant/v1",
			"kind":               "Want",
			"name":               w.Owner,
			"id":                 ownerID,
			"controller":         true,
			"blockOwnerDeletion": true,
		}}
	}

	payload := map[string]any{
		"metadata": map[string]any{
			"name":            w.Name,
			"type":            w.Type,
			"labels":          labels,
			"ownerReferences": ownerRefs,
		},
		"spec": map[string]any{
			"params": w.Params,
		},
	}

	body, err := mywantPost(mywantURL, "/api/v1/wants", payload)
	if err != nil {
		return "", fmt.Errorf("create want %q: %w", w.Name, err)
	}
	var resp struct {
		WantIDs []string `json:"want_ids"`
	}
	if err := json.Unmarshal(body, &resp); err != nil || len(resp.WantIDs) == 0 {
		return "", fmt.Errorf("create want %q: no id returned", w.Name)
	}
	return resp.WantIDs[0], nil
}

// ── initMywantStage ───────────────────────────────────────────────────────────

// initMywantStage is called asynchronously after a stage jump.
// It cleans up old wants, creates fresh ones, registers the lifecycle webhook,
// and runs any on_start_run commands.
func (s *Server) initMywantStage(stageID string, stage *Stage) {
	setup := stage.MywantSetup
	base := s.cfg.MywantURL
	log.Printf("[MYWANT-SETUP] initialising mywant for stage %s", stageID)

	// 1. Cleanup old wants + reset mywant's achievement watcher cache
	if len(setup.CleanupLabel) > 0 {
		if err := cleanupMywantWants(base, setup.CleanupLabel); err != nil {
			log.Printf("[MYWANT-SETUP] cleanup error: %v", err)
		}
	}
	// Clear the achievement watcher's status-transition cache so new want IDs
	// start fresh and the immediate "achieved" on creation is recorded as baseline.
	if err := mywantDelete(base, "/api/v1/lifecycle-webhooks/achievement-cache", nil); err != nil {
		log.Printf("[MYWANT-SETUP] achievement cache reset error: %v", err)
	}

	// small delay to let deletes propagate
	time.Sleep(500 * time.Millisecond)

	// 2. Create wants in declaration order; track name→ID for owner resolution.
	// A short sleep after each creation lets mywant's storage commit before the
	// next want tries to reference it as an owner.
	nameToID := make(map[string]string)
	for _, w := range setup.Wants {
		id, err := createMywantWant(base, w, nameToID)
		if err != nil {
			log.Printf("[MYWANT-SETUP] %v", err)
			continue
		}
		nameToID[w.Name] = id
		log.Printf("[MYWANT-SETUP] created %q (%s)", w.Name, id)
		time.Sleep(300 * time.Millisecond)
	}

	// 3. Register lifecycle webhook (RPG server's own hook endpoint)
	if setup.RegisterWebhook {
		rpgHookURL := fmt.Sprintf("http://localhost:%d/api/v1/hooks/mywant", s.cfg.Port)
		id, err := ensureLifecycleWebhook(base, rpgHookURL)
		if err != nil {
			log.Printf("[MYWANT-SETUP] webhook registration error: %v", err)
		} else {
			log.Printf("[MYWANT-SETUP] lifecycle webhook registered: %s", id)
		}
	}

	// 4. Run on_start commands sequentially
	for _, cmd := range setup.OnStartRun {
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			log.Printf("[MYWANT-SETUP] on_start_run %q: %v", cmd, err)
		}
	}

	log.Printf("[MYWANT-SETUP] stage %s initialised", stageID)
}
