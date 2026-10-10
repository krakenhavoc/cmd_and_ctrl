package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5ea93956-085d-409b-bd95-dd669fa69eeb",
		Name:         "Ginger, Queen of Sweets",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Ginger, Queen of Sweets"),
			On(game.EventBeginUpkeep,
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return YoureTheMonarch(g, source.Controller)
				},
				"Ginger, Queen of Sweets — create a Gingerbrute token",
				func(g *game.Game, item *game.StackItem) error {
					if !YoureTheMonarch(g, item.Controller) {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: GingerbruteToken(), N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice Ginger: You gain 6 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 6}.Apply(NewContext(g, item))
			},
		}},
	})
}
