package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Odylic Wraith — Creature — Wraith {3}{B}, 2/2 (Weatherlight, #5629):
//
//	"Swampwalk (This creature can't be blocked as long as defending player
//	 controls a Swamp.)
//	 Whenever this creature deals damage to a player, that player discards a
//	 card."
//
// Swampwalk is landwalk, an engine keyword (CR 702.14c). The trigger is
// Abyssal Specter's: the damaged player chooses the card (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "613fa223-856e-49c4-8952-224895e00b5e",
		Name:            "Odylic Wraith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToAPlayer,
				"Odylic Wraith — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, effects.NewContext(g, item).Trigger().Event.Target)
				}),
		},
	})
}
