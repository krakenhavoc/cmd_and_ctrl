package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_land_drops_purpose_test.go — #2678: a card that lets its
// controller play additional lands declares how many, so the bot can
// price the drop, and the declaration agrees with the static the
// engine runs.

// Every catalog permanent with AdditionalLandPlays declares the same
// number as Purpose.ExtraLandDrops.
func TestEveryExtraLandDropIsDeclared(t *testing.T) {
	n := 0
	for _, spec := range All() {
		if spec.AdditionalLandPlays == 0 {
			continue
		}
		n++
		if spec.Purpose.ExtraLandDrops != spec.AdditionalLandPlays {
			t.Errorf("%s plays %d additional land(s) and declares ExtraLandDrops %d", spec.Name, spec.AdditionalLandPlays, spec.Purpose.ExtraLandDrops)
		}
	}
	if n == 0 {
		t.Fatal("no catalog card declares AdditionalLandPlays: the test is reading the wrong registry")
	}
}

func TestExtraLandDropsGuard(t *testing.T) {
	for _, c := range []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string
	}{
		{"on the card", purposeOnCard, game.Purpose{ExtraLandDrops: 1}, ""},
		{"on a mode", purposeOnMode, game.Purpose{ExtraLandDrops: 1}, "off the card"},
		{"on a triggered row", purposeOnTriggered, game.Purpose{ExtraLandDrops: 1}, "off the card"},
		{"negative", purposeOnCard, game.Purpose{ExtraLandDrops: -1}, "negative amount"},
	} {
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
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}
