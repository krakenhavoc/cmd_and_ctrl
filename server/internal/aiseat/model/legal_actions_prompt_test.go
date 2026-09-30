package model

import (
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0105 §8: the legal_actions digest reaches a bot's Input.View,
// because that is the same filtered frame a human gets. The prompt
// renderer picks fields rather than marshalling the view, and it does
// not pick this one: a model is shown the move list it chooses from,
// not a second summary of it. So the prompt is byte-identical with and
// without the digest.
func TestPromptIgnoresTheLegalActionsDigest(t *testing.T) {
	prompt := func(digest *protocol.LegalActionsView) (string, string) {
		t.Helper()
		fake := AlwaysIndex(0)
		p := testPolicy(t, fake, &stubB{index: 1, reason: "b"}, func(c *Config) { c.Deck = testDeck() })
		in := castWindow()
		in.View.LegalActions = digest
		decide(t, p, in, 2*time.Second)
		reqs := fake.Requests()
		if len(reqs) != 1 {
			t.Fatalf("calls = %d, want 1", len(reqs))
		}
		var system string
		for _, b := range reqs[0].System {
			system += b.Text
		}
		return system, reqs[0].User
	}
	sysWithout, userWithout := prompt(nil)
	sysWith, userWith := prompt(&protocol.LegalActionsView{
		Pass: true,
		Sources: map[string]*protocol.LegalSourceView{
			"bolt": {Kinds: []legal.Kind{legal.KindCast}, Moves: 1, Zones: []string{"hand"}, Faces: []int{0}},
		},
	})
	if sysWith != sysWithout {
		t.Error("the digest changed the static prompt block")
	}
	if userWith != userWithout {
		t.Errorf("the digest changed the per-decision prompt:\nwithout:\n%s\nwith:\n%s", userWithout, userWith)
	}
}
