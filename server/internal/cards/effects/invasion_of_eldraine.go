package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of Eldraine // Prickle Faeries — March of the Machine (#5477).
//
// Front, Battle — Siege 4, {3}{B}: "When this Siege enters, target opponent
// discards two cards." Back, Prickle Faeries, Creature — Faerie 2/2: "Flying.
// At the beginning of each opponent's upkeep, if that player has two or fewer
// cards in hand, this creature deals 2 damage to them."
//
// The "if" is an intervening if (CR 603.4): asked as the upkeep begins and
// again as the ability resolves, of the upkeep's player.
//
// No simplification.
const invasionOfEldraineOracleID = "329fcb1e-7066-46c5-ba27-20ed2b18c1ba"

func init() {
	Register(Spec{
		OracleID:     invasionOfEldraineOracleID,
		Name:         "Invasion of Eldraine",
		Completeness: CompletenessFull,
		Battle:       &BattleSpec{Defense: 4, Subtype: BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Invasion of Eldraine — target opponent discards two cards", targetOpponentDiscards(2)),
				TargetPlayer("target opponent", Opponent())),
			DefeatedTrigger("Invasion of Eldraine — defeated: exile it, then cast Prickle Faeries", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:        invasionOfEldraineOracleID + "#1",
		Name:            "Prickle Faeries",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, AllOf(ByAnOpponent, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return prickleFaeriesHasTwoOrFewer(g, ev.Actor)
			}), "Prickle Faeries — 2 damage to that player (two or fewer cards in hand)",
				func(g *game.Game, item *game.StackItem) error {
					victim := item.Trigger.Event.Actor
					if !prickleFaeriesHasTwoOrFewer(g, victim) {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: victim, Amount: 2}.Apply(NewContext(g, item))
				}),
		},
	})
}

// prickleFaeriesHasTwoOrFewer is "that player has two or fewer cards in hand".
func prickleFaeriesHasTwoOrFewer(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && !p.Eliminated && p.Hand != nil && p.Hand.Size() <= 2
}
