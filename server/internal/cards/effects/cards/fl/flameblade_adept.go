package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flameblade Adept — Creature — Jackal Warrior {R}, 1/2 (Jumpstart 2022,
// #5948):
//
//	"Menace
//	 Whenever you cycle or discard a card, this creature gets +1/+0 until end
//	 of turn."
//
// Menace is a printed keyword. The trigger watches the discard alone, which a
// cycle is (CR 702.29d; Ominous Sphinx's watch), so cycling a card is one
// trigger, not two. One trigger per card, so two cards discarded together are
// two boosts that stack. It acts only while the Adept is still the permanent
// that triggered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1c721e6-d0b1-4443-937d-89f7c3a459e6",
		Name:            "Flameblade Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Flameblade Adept — it gets +1/+0 until end of turn", func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Flameblade Adept — +1/+0"}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
