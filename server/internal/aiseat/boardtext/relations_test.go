package boardtext_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func attachmentBoard() *protocol.GameView {
	v := view()
	v.Battlefield.Cards = []protocol.CardView{
		{InstanceID: "h", Name: "Danitha Capashen, Paragon", Controller: "a", TypeLine: "Creature", Power: 3, Toughness: 3},
		{InstanceID: "e", Name: "Swiftfoot Boots", Controller: "a", TypeLine: "Artifact — Equipment", AttachedTo: &protocol.TargetRefView{Kind: "card", ID: "h"}},
		{InstanceID: "o", Name: "Oblivion Ring", Controller: "a", TypeLine: "Enchantment"},
		{InstanceID: "p", Name: "Curse", Controller: "a", TypeLine: "Enchantment — Aura", AttachedTo: &protocol.TargetRefView{Kind: "player", ID: "b"}},
	}
	v.Exile.Cards = []protocol.CardView{{InstanceID: "x", Name: "Grave Pact", TypeLine: "Enchantment"}}
	v.DelayedTriggers = []protocol.DelayedTriggerView{{ID: "d", Source: "o", Cards: []string{"x"}}}
	return v
}

func TestBattlefieldSaysWhatIsAttachedAndWhatIsHeld(t *testing.T) {
	got := boardtext.Render(attachmentBoard(), "a", boardtext.Options{})
	for _, want := range []string{
		"Swiftfoot Boots (attached to Danitha Capashen, Paragon)",
		"Oblivion Ring (exiling Grave Pact)",
		"Curse (attached to ",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestBattlefieldNeverNamesAHiddenExile(t *testing.T) {
	v := attachmentBoard()
	v.Exile.Cards[0] = protocol.CardView{InstanceID: "x", FaceDown: true}
	got := boardtext.Render(v, "a", boardtext.Options{})
	if strings.Contains(got, "exiling") || strings.Contains(got, "Grave Pact") {
		t.Errorf("hidden exile leaked:\n%s", got)
	}
	if !strings.Contains(got, "Oblivion Ring") {
		t.Errorf("O-Ring itself should still be listed:\n%s", got)
	}
}

func TestBattlefieldStepKeyedDelayedTriggerIsNotAHold(t *testing.T) {
	v := attachmentBoard()
	v.DelayedTriggers[0].At = "end"
	got := boardtext.Render(v, "a", boardtext.Options{})
	if strings.Contains(got, "exiling") {
		t.Errorf("an end-step return is not an until-hold:\n%s", got)
	}
}
