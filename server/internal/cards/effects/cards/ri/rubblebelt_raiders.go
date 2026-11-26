package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rubblebelt Raiders — Creature — Human Warrior {1}{R/G}{R/G}{R/G}, 3/3
// (Gatecrash, S82 OLD #5565):
//
//	"Whenever this creature attacks, put a +1/+1 counter on it for each
//	 attacking creature you control."
//
// Wingmate Roc's attack count, read as the trigger resolves: every creature
// its controller controls that is attacking at that moment, the Raiders
// included while it is still attacking (CR 508.1). The counters go on the
// Raiders only while it is still the same object on the battlefield, and not at
// all if none of the attackers is left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7904dd02-f231-4538-b1bb-81bce69309b1",
		Name:         "Rubblebelt Raiders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Rubblebelt Raiders — put a +1/+1 counter on it for each attacking creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
						return nil
					}
					n := countAttackingCreaturesOf(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
	})
}

// countAttackingCreaturesOf is "each attacking creature you control".
func countAttackingCreaturesOf(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == controller && c.AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n
}
