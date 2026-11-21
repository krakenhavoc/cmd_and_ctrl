package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ginger, Queen of Sweets — Legendary Artifact Creature — Food Noble
// {6}, 6/4:
//
//	"When Ginger enters, you become the monarch.
//	 {2}, {T}, Sacrifice Ginger: You gain 6 life.
//	 At the beginning of each upkeep, if you're the monarch, create
//	 a Gingerbrute token."
//
// The monarch comes from WhenThisEntersYouBecomeTheMonarch. The upkeep
// trigger fires on EVERY player's upkeep (AtEachUpkeep) and carries an
// intervening "if" (CR 603.4): "you're the monarch" is read when the
// upkeep begins and again as the trigger resolves, so a crown stolen in
// response means no token. The token is the Gingerbrute from
// reality_fracture_fra_creature_b_helpers.go and always goes to
// Ginger's controller, whoever's upkeep it is. The sacrifice ability
// taps, so Ginger is summoning-sick for it like any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ea93956-085d-409b-bd95-dd669fa69eeb",
		Name:         "Ginger, Queen of Sweets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Ginger, Queen of Sweets"),
			effects.On(game.EventBeginUpkeep,
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.YoureTheMonarch(g, source.Controller)
				},
				"Ginger, Queen of Sweets — create a Gingerbrute token",
				effects.Do(effects.If{
					Cond: effects.CondMonarch{},
					Then: []effects.Applier{effects.CreateToken{Template: effects.GingerbruteToken(), N: 1}},
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice Ginger: You gain 6 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 6}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
