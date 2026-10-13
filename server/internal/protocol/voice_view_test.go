package protocol

import "testing"

// Tests for voice_view.go (ADR 0159 §6): FilterViewFor reduces the
// table's voice room to one viewer's access, and the link reaches only
// a granted viewer of a ready room.

func fullVoice(status string) *VoiceView {
	return &VoiceView{
		Enabled: true,
		Status:  status,
		URL:     "https://discord.com/channels/1/2",
		Access: map[string]string{
			"alice": VoiceAccessGranted,
			"bob":   VoiceAccessNotInServer,
			"gus":   VoiceAccessNoDiscord,
		},
	}
}

func TestFilterVoiceForEachViewer(t *testing.T) {
	cases := []struct {
		viewer, status string
		wantYou        string
		wantURL        bool
	}{
		{"alice", VoiceStatusReady, VoiceAccessGranted, true},
		{"alice", VoiceStatusPending, VoiceAccessGranted, false},
		{"alice", VoiceStatusIdle, VoiceAccessGranted, false},
		{"bob", VoiceStatusReady, VoiceAccessNotInServer, false},
		{"gus", VoiceStatusReady, VoiceAccessNoDiscord, false},
		{"stranger", VoiceStatusReady, VoiceAccessSpectator, false},
		{SpectatorViewerID, VoiceStatusReady, VoiceAccessSpectator, false},
	}
	for _, tc := range cases {
		got := filterVoiceFor(fullVoice(tc.status), tc.viewer)
		if got.You != tc.wantYou || (got.URL != "") != tc.wantURL {
			t.Errorf("%s / %s: you %q url %q; want %q, url %v", tc.viewer, tc.status, got.You, got.URL, tc.wantYou, tc.wantURL)
		}
		if got.Access != nil {
			t.Errorf("%s: the access map reached a player: %v", tc.viewer, got.Access)
		}
		if !got.Enabled || got.Status != tc.status {
			t.Errorf("%s: enabled %v status %q", tc.viewer, got.Enabled, got.Status)
		}
	}
}

func TestFilterVoiceForTheAdminIsWholeAndACopy(t *testing.T) {
	full := fullVoice(VoiceStatusReady)
	got := filterVoiceFor(full, "")
	if got.URL == "" || len(got.Access) != 3 {
		t.Fatalf("admin view: %+v", got)
	}
	got.Access["alice"] = "changed"
	if full.Access["alice"] != VoiceAccessGranted {
		t.Error("the admin's copy aliases the full view's access map")
	}
}

func TestFilterViewForCarriesVoice(t *testing.T) {
	v := GameView{Voice: fullVoice(VoiceStatusReady)}
	if got := FilterViewFor(v, "alice").Voice; got == nil || got.URL == "" || got.You != VoiceAccessGranted {
		t.Errorf("alice: %+v", got)
	}
	if got := FilterViewFor(v, "bob").Voice; got == nil || got.URL != "" {
		t.Errorf("bob: %+v", got)
	}
	if got := FilterViewFor(GameView{}, "alice").Voice; got != nil {
		t.Errorf("no voice on the full view, but the filtered one has %+v", got)
	}
}
