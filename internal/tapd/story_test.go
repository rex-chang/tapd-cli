package tapd_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rex-chang/tapd-cli/internal/config"
	"github.com/rex-chang/tapd-cli/internal/tapd"
)

func TestClient_GetStories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stories" {
			t.Errorf("expected path /stories, got %s", r.URL.Path)
		}

		wsID := r.URL.Query().Get("workspace_id")
		if wsID != "test_ws" {
			t.Errorf("expected workspace_id test_ws, got %s", wsID)
		}

		resp := map[string]interface{}{
			"status": 1,
			"data": []map[string]interface{}{
				{
					"Story": map[string]interface{}{
						"id":          "111",
						"name":        "test story 1",
						"status":      "status_1",
						"creator":     "user_1",
						"description": "desc 1",
					},
				},
				{
					"Story": map[string]interface{}{
						"id":          "222",
						"name":        "test story 2",
						"status":      "status_2",
						"creator":     "user_2",
						"description": "desc 2",
					},
				},
			},
			"info": "success",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := &config.Config{
		WorkspaceID: "test_ws",
		APIUser:     "user",
		APIToken:    "token",
	}

	client := tapd.NewClient(cfg, tapd.WithBaseURL(ts.URL))

	stories, err := client.GetStories()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stories) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(stories))
	}

	if stories[0].Story.ID != "111" || stories[1].Story.Name != "test story 2" {
		t.Errorf("parsed story list does not match expected")
	}
}

func TestClient_GetStories_Failed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"status": 0,
			"info":   "API rate limit exceeded",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := &config.Config{
		WorkspaceID: "test_ws",
		APIUser:     "user",
		APIToken:    "token",
	}

	client := tapd.NewClient(cfg, tapd.WithBaseURL(ts.URL))

	_, err := client.GetStories()
	if err == nil {
		t.Fatalf("expected error from API, got nil")
	}

	expectedErrSubstring := "TAPD 返回错误: API rate limit exceeded"
	if err.Error() != parseErrStr("解析需求列表失败", "TAPD 返回错误") && err.Error() != expectedErrSubstring { // simple check, standard json decoding shouldn't fail
		if err.Error() != expectedErrSubstring {
			t.Errorf("unexpected error string: %v", err)
		}
	}
}

func parseErrStr(errStr1, errStr2 string) string {
	return "test check"
}
