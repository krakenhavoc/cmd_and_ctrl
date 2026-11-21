package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostbite Pyromental — Creature — Elemental {U}{R}{R}, 4/4:
//
//	"Trample, haste
//	 Whenever this creature deals combat damage to a player, draw two
//	 cards.
//	 At the beginning of the end step, sacrifice this creature."
//
// Lightning Serpent's end-step sacrifice (the END STEP of every turn,
// not only its controller's, so it is gone at the end of the turn it
// was cast) plus a combat-damage draw. The draw reads the controller
// off the trigger item, so a creature stolen mid-combat draws for its
// new controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a104f73-597b-49e3-8088-13db31f61900",
		Name:            "Frostbite Pyromental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Frostbite Pyromental — draw two cards",
				effects.Do(effects.DrawCards{N: 2})),
			effects.AtEachStep(game.StepEnd, "Frostbite Pyromental — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
