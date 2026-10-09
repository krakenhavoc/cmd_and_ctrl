package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "9a104f73-597b-49e3-8088-13db31f61900",
		Name:            "Frostbite Pyromental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Frostbite Pyromental — draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
				}),
			AtEachStep(game.StepEnd, "Frostbite Pyromental — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}
