package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tomb of Urami — Legendary Land (Saviors of Kamigawa, S82 OLD b593):
//
//	"{T}: Add {B}. Tomb of Urami deals 1 damage to you if you don't control an Ogre.
//	 {2}{B}{B}, {T}, Sacrifice all lands you control: Create Urami, a legendary 5/5
//	 black Demon Spirit creature token with flying."
//
// The mana ability's damage is a rider that reads the Ogre as the mana is made (CR
// 605.3b), through the layered subtypes, so an animated land that is an Ogre counts.
// The second ability sacrifices every land its controller has, the Tomb among them, as
// a cost (SacrificeAllYouControl, #6208); the token is made even with no land left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f002be6a-e459-49c4-b765-062e30107439",
		Name:         "Tomb of Urami",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "{T}: Add {B}. Tomb of Urami deals 1 damage to you if you don't control an Ogre.",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				ogre := "Ogre"
				if c, ok := g.LookupCardForEffect(source); ok {
					ogre = c.TextSubtype("Ogre")
				}
				if effects.ControlsA(ogre)(g, controller) {
					return nil
				}
				return g.DealDamageToPlayerForEffect(source, controller, 1)
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{B}, {T}, Sacrifice all lands you control: Create Urami, a legendary 5/5 black Demon Spirit creature token with flying.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}{B}"), effects.TapCost(), effects.SacrificeAllYouControl("lands you control", effects.Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("5/5 black Legendary Demon Spirit named Urami with flying"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
