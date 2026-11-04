package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9076a886-efe0-4c99-8661-f0fd7a7b3129",
		Name:         "Paramecia Coloniex",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Paramecia Coloniex — mill three cards", Do(MillCards{N: 3})),
			Optional(
				WhenThisDies("Paramecia Coloniex — exile it",
					exileItFromTheGraveyardThen("Paramecia Coloniex — put target creature card from your graveyard on top of your library", parameciaColoniexTuckBody)),
				"Paramecia Coloniex — exile it to put a creature card from your graveyard on top of your library?"),
		},
	})
}

// parameciaColoniexTuckBody is the reflexive half: "put target creature card
// from your graveyard on top of your library". An on-disk key.
var parameciaColoniexTuckBody = game.ReflexiveBody("paramecia-coloniex/tuck", simpleBody(PutChosenTargetOnTopOfLibrary),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())
	}))
