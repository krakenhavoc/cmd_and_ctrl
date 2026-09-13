package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul of the Harvest — Creature — Elemental {4}{G}{G}, 6/6 (EDHREC
// rank 1636):
//
//	"Trample
//	 Whenever another nontoken creature you control enters, you may
//	 draw a card."
//
// Green's six-mana card-draw body: every real creature after it is
// a cantrip. Corpse Knight's condition
// (b13AnotherCreatureYouControlEntered) with the token exclusion —
// Avenger of Zendikar's Plants make nothing — and "you may" is the
// optional-trigger prompt, so the controller can decline a draw off
// an empty library. Post-layer type, so an animated land counts as
// printed; the Soul's own entry is "another" and does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6b9194dd-2296-4879-ac5e-f89b18431df6",
		Name:            "Soul of the Harvest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherNontokenCreatureYouControlEntered(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Soul of the Harvest — draw a card?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Soul of the Harvest — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
