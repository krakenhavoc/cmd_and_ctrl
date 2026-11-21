package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gary Clone — Creature — Human Citizen {1}{W}, 1/3 (Fallout):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay {2}
//	 any number of times. When this creature enters, create that many tokens
//	 that are copies of it.)
//	 Whenever this creature attacks, each creature you control named Gary
//	 Clone gets +1/+0 until end of turn."
//
// Squad is Powder Ganger's (SquadMana, Squad). The attack trigger is Gruff
// Triplets' name read: every creature I control that is named Gary Clone as
// the trigger resolves (the token copies are), chosen once as it resolves
// (CR 611.2c), so a Clone that arrives later is not boosted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d9af0da0-5e10-40c5-afb2-3f44c016b927",
		Name:          "Gary Clone",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.Squad("Gary Clone"),
			effects.WheneverThisAttacks("Gary Clone — each creature you control named Gary Clone gets +1/+0 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match: func(_ *game.Game, controller uuid.UUID, c game.Card) bool {
						return c.Controller == controller && c.IsCreature() && game.HasName(c, "Gary Clone")
					},
					Power: 1,
					Label: "Gary Clone — +1/+0",
				})),
		},
	})
}
