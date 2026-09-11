package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// surveil_lands.go — the Murders at Karlov Manor "surveil land"
// cycle:
//
//	"({T}: Add {X} or {Y}.)"
//	"This land enters tapped."
//	"When this land enters, surveil 1."
//
// Three of the ten, the three this deck plays. Like the shocklands
// these are nonbasic lands carrying two basic land types, so the
// mana ability is declared rather than left to the synthetic shape.
//
// S22: the surveil is real. These shipped in an earlier sprint with
// the enters-tapped half and the mana only, plus a long comment
// explaining that the surveil — the whole reason to play one of
// these over a Guildgate — was not implemented, because it needed a
// PendingChoice kind that did not exist. The Surveil primitive
// closes that: the ETB trigger queues the real look-at-one prompt,
// and a reanimator deck gets the graveyard card it was promised.
func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"ccfb8b4d-651c-418a-aa19-cb23105b3f2f", "Meticulous Archive", "W", "U"},
		{"216a2a92-9ca3-4ca3-8af7-686c13b04290", "Shadowy Backstreet", "W", "B"},
		{"08d80efc-9542-4ba2-824c-c8615d8d07f2", "Undercity Sewers", "U", "B"},
	} {
		// Bind per iteration: the closures below outlive the loop.
		name := t.name
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          name,
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, name+" — surveil 1",
						func(g *game.Game, item *game.StackItem) error {
							return Surveil{N: 1}.Apply(NewContext(g, item))
						})
				},
			}},
		})
	}
}
