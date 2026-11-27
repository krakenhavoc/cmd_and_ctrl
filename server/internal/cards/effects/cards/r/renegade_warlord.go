package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renegade Warlord — Creature — Human Soldier {4}{R}, 3/3 (Tempest, S82 OLD
// b756, #5628):
//
//	"First strike
//	 Whenever this creature attacks, each other attacking creature gets +1/+0
//	 until end of turn."
//
// Hero of Bladehold's battle cry, written out as the trigger it is: every
// attacking creature but the Warlord, as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae63849d-7352-4a2b-9d64-f3005146bd22",
		Name:            "Renegade Warlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Renegade Warlord — each other attacking creature gets +1/+0",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.AttackingCreature(), effects.OtherThan(item.SourceCardID)),
						Power: 1,
						Label: "Renegade Warlord — +1/+0",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
