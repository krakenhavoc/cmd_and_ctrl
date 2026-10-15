package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// landfall_doubler_test.go — #3146: a landfall doubler declares how many
// extra times each landfall trigger fires, and a fetch land declares the
// land its activation puts onto the battlefield.

func TestLandfallDoublerDeclarations(t *testing.T) {
	for _, c := range []struct{ oracle, name string }{
		{"3bcf090c-e890-4a9f-a8aa-6079e4ec9947", "Ancient Greenwarden"},
		{"5ade11c0-41dd-4b6a-9f5b-c5903a3a0d7f", "Elesh Norn, Mother of Machines"},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Purpose.LandfallExtraTriggers != 1 || len(spec.TriggerDoublers) == 0 {
			t.Errorf("%s declares %d extra landfall triggers with %d doublers, want 1 with a doubler",
				c.name, spec.Purpose.LandfallExtraTriggers, len(spec.TriggerDoublers))
		}
	}
}

func TestLandfallDoublerGuard(t *testing.T) {
	cases := []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string // "" = accepted
	}{
		{"on the card", purposeOnCard, game.Purpose{LandfallExtraTriggers: 1}, ""},
		{"on a triggered row", purposeOnTriggered, game.Purpose{LandfallExtraTriggers: 1}, "off the card"},
		{"on an activated row", purposeOnActivated, game.Purpose{LandfallExtraTriggers: 1}, "off the card"},
		{"negative", purposeOnCard, game.Purpose{LandfallExtraTriggers: -1}, "negative amount"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				checkPurpose("Test Card", "card", c.slot, c.p)
			}()
			switch {
			case c.want == "" && msg != "":
				t.Errorf("refused: %s", msg)
			case c.want != "" && !strings.Contains(msg, c.want):
				t.Errorf("got %q, want a refusal containing %q", msg, c.want)
			}
		})
	}
	// A card with no TriggerDoubler cannot declare one.
	var msg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				msg, _ = r.(string)
			}
		}()
		checkSpecPurposes(Spec{Name: "Test Card", Purpose: game.Purpose{LandfallExtraTriggers: 1}})
	}()
	if !strings.Contains(msg, "no TriggerDoubler") {
		t.Errorf("a card with no doubler: got %q, want a refusal", msg)
	}
}

func TestFetchLandDeclarations(t *testing.T) {
	for _, c := range []struct {
		oracle, name    string
		lands, untapped int
	}{
		{"09dd85aa-47bc-4713-a9b9-8b52ff2285ed", "Misty Rainforest", 1, 1},
		{"67d60b24-d429-4ded-90d9-06e49f28c396", "Verdant Catacombs", 1, 1},
		{"ef86989d-ce80-4e55-aece-7d11710eeffa", "Polluted Delta", 1, 1},
		{"032b8a0d-491a-4a12-ab9f-689010054d5b", "Prismatic Vista", 1, 1},
		{"a75445d3-1303-4bb5-89ad-26ea93fecd48", "Evolving Wilds", 1, 0},
		{"1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a", "Terramorphic Expanse", 1, 0},
		{"0c85b8f7-0bd0-4680-9ec5-d4b110460a54", "Fabled Passage", 1, 0},
		{"2549bc57-9ffb-4053-9f10-f2a5f792b845", "Myriad Landscape", 2, 0},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		found := false
		for _, a := range spec.Activated {
			if a.Purpose.Lands == 0 {
				continue
			}
			found = true
			if a.Purpose.Lands != c.lands || a.Purpose.LandsUntapped != c.untapped || a.Purpose.Answers != game.AnswerValue {
				t.Errorf("%s declares %+v, want %d lands, %d untapped, answers value", c.name, a.Purpose, c.lands, c.untapped)
			}
		}
		if !found {
			t.Errorf("%s declares no lands on any activated row", c.name)
		}
	}
}
