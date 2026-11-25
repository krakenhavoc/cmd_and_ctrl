package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grunn, the Lonely King — Legendary Creature — Ape Warrior {4}{G}{G}, 5/5
// (Dominaria):
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 If Grunn was kicked, it enters with five +1/+1 counters on it.
//	 Whenever Grunn attacks alone, double its power and toughness until end of
//	 turn."
//
// The counters are a CR 614.1c entry clause read off the cast (Vodalian
// Serpent's CountersPerKick, five for the one possible kick). The attack
// trigger fires for Grunn's own declaration when it is the only attacker
// (game.AttackedAlone, CR 506.5); doubling is CR 701.10: +X/+Y with X and Y its
// power and toughness as the trigger resolves, counters and pumps included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "e3d6d146-47f4-46f9-9a15-2ea649093668",
		Name:                       "Grunn, the Lonely King",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 5)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.AllOf(effects.ThisAttacked, effects.OldB208AttackedAlone),
				"Grunn, the Lonely King — double its power and toughness until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.DoublePowerAndToughnessOfEachUntilEOT(effects.NewContext(g, item),
						[]uuid.UUID{item.SourceCardID}, "Grunn, the Lonely King — double power and toughness")
				}),
		},
	})
}
