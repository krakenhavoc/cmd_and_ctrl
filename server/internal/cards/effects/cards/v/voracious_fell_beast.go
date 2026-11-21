package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voracious Fell Beast — Creature — Drake Beast {4}{B}{B}, 4/4 (The Lord of
// the Rings: Tales of Middle-earth):
//
//	"Flying
//	 When this creature enters, each opponent sacrifices a creature of
//	 their choice. Create a Food token for each creature sacrificed this
//	 way."
//
// Syphon Flesh's edict (EachPlayerSacrifices with ExceptController, one
// prompt for each opponent with a choice, in APNAP order) and its keyed
// "for each creature sacrificed this way" clause: the Food tokens wait until
// every opponent has chosen and the creatures have finished leaving, and
// there is one for each that really was sacrificed (an opponent with no
// creature sacrifices nothing and gives no Food), all created by the
// Beast's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e037c5e-3e07-4ffc-a388-7976459f6ff7",
		Name:            "Voracious Fell Beast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Voracious Fell Beast — each opponent sacrifices a creature; create a Food token for each creature sacrificed this way",
				func(g *game.Game, item *game.StackItem) error {
					return effects.EachPlayerSacrifices{
						ExceptController: true,
						Match:            effects.Creature(),
						Label:            "a creature",
						Keyed:            voraciousFellBeastEdicted,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var voraciousFellBeastEdicted = effects.RegisterEdictThen("prompt-run/voracious-fell-beast-food", func(ctx *effects.Context, sacrificed game.PromptedSacrifices) error {
	return effects.CreateToken{Template: effects.FoodToken(), N: sacrificed.Count()}.Apply(ctx)
})
