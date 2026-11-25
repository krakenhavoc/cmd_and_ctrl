package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of Liliana — Legendary Enchantment {2}{B} (Eldritch Moon, S82 OLD,
// #5541):
//
//	"When Oath of Liliana enters, each opponent sacrifices a creature of
//	 their choice.
//	 At the beginning of each end step, if a planeswalker entered the
//	 battlefield under your control this turn, create a 2/2 black Zombie
//	 creature token."
//
// Rot-Tide Gargantua's edict. The second ability is an intervening "if"
// (CR 603.4) read from the per-turn entered-card-type tally, so a
// planeswalker that entered before the Oath did, or has since left, still
// counts; it is checked as each end step begins and again as the trigger
// resolves, and every end step is a chance (not only yours).
//
// No simplification.
func init() {
	const name = "Oath of Liliana"
	effects.Register(effects.Spec{
		OracleID:     "0bd7fd12-a0b7-4b4e-8a33-cace6e6bd6d1",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — each opponent sacrifices a creature of their choice",
				effects.Do(effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"})),
			effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEnd, false),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.S82OldB277PlaneswalkerEnteredThisTurn(g, source.Controller)
				}),
				name+" — create a 2/2 black Zombie creature token",
				effects.Do(effects.If{
					Cond: effects.S82OldB277CondPlaneswalkerEntered{}, // CR 603.4: checked again as it resolves
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}},
				})),
		},
	})
}
