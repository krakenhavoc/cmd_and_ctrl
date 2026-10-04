package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defiling Daemogoth — Creature — Demon {3}{B}{B}, 5/4:
//
//	"Menace
//	 Whenever a creature you control deals combat damage to a player,
//	 you gain 1 life.
//	 At the beginning of your end step, each opponent loses X life,
//	 where X is the amount of life you gained this turn."
//
// The combat trigger is Bident of Thassa's: one EventDealDamage per
// creature per player it hits, so three connecting attackers gain 3.
// X is read as the end-step trigger resolves, from the turn tally's
// LifeGained cell (b15LifeGainedThisTurn), so life gained in response
// still counts — the printed "the amount of life you gained this
// turn" has no snapshot. It is life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d3f7095a-b287-4858-9d08-242cf18e87cb",
		Name:            "Defiling Daemogoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Defiling Daemogoth — gain 1 life", Do(GainLife{Amount: 1})),
			AtYourEndStep("Defiling Daemogoth — each opponent loses life equal to the life you gained this turn", func(g *game.Game, item *game.StackItem) error {
				x := b15LifeGainedThisTurn(g, item.Controller)
				if x <= 0 {
					return nil
				}
				return eachOpponentLosesLife(g, item, x)
			}),
		},
	})
}
