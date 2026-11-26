package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emperor's Vanguard — Creature — Human Scout {3}{G}, 4/3 (Ixalan, #5532):
//
//	"Whenever this creature deals combat damage to a player, it explores."
//
// One trigger per combat damage event to a player (CR 510.2); the Vanguard
// is the explorer, read as the object that dealt the damage. A Vanguard
// that left before the trigger resolved still explores by its last-known
// information: the card is revealed and no counter is put on anything
// (CR 701.44c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5007425e-a626-46f7-ab7d-777af97ee724",
		Name:         "Emperor's Vanguard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Emperor's Vanguard — it explores",
				func(g *game.Game, item *game.StackItem) error {
					return effects.Explores{Explorer: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
