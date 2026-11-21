package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paramecia Coloniex — {1}{B} Creature — Zombie Worm 2/2 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"When this creature enters, mill three cards.
//	 When this creature dies, you may exile it. When you do, put target
//	 creature card from your graveyard on top of your library."
//
// Undead Butler with Golgari Thug's destination. The "you may" is the
// dies trigger's own (Optional); the exile of the card out of the
// graveyard is checked, and only once it is really in exile is the
// CR 603.12 reflexive trigger created. Its target is chosen as it goes on
// the stack (CR 603.3d), after Paramecia has left the graveyard, so it
// can never put itself back. The put is Golgari Thug's tuck
// (PutChosenTargetOnTopOfLibrary).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9076a886-efe0-4c99-8661-f0fd7a7b3129",
		Name:         "Paramecia Coloniex",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Paramecia Coloniex — mill three cards", effects.Do(effects.MillCards{N: 3})),
			effects.Optional(
				effects.WhenThisDies("Paramecia Coloniex — exile it",
					effects.ExileItFromTheGraveyardThen("Paramecia Coloniex — put target creature card from your graveyard on top of your library", parameciaColoniexTuckBody)),
				"Paramecia Coloniex — exile it to put a creature card from your graveyard on top of your library?"),
		},
	})
}

// parameciaColoniexTuckBody is the reflexive half: "put target creature card
// from your graveyard on top of your library". An on-disk key.
var parameciaColoniexTuckBody = game.ReflexiveBody("paramecia-coloniex/tuck", effects.SimpleBody(effects.PutChosenTargetOnTopOfLibrary),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())
	}))
