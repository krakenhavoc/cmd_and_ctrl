package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avenger of the Fallen — Creature — Human Warrior {2}{B}, 2/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Deathtouch
//	 Mobilize X, where X is the number of creature cards in your
//	 graveyard. (Whenever this creature attacks, create X tapped and
//	 attacking 1/1 red Warrior creature tokens. Sacrifice them at the
//	 beginning of the next end step.)"
//
// Mobilize (CR 702.181) with a counted X (MobilizeX, mobilize.go): the
// creature cards in your graveyard are counted as the attack trigger
// resolves (CR 608.2h), so a creature that dies in response counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "158d0272-a850-4399-8afa-d0caa143c3cb",
		Name:            "Avenger of the Fallen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered:       []game.TriggeredAbility{effects.MobilizeX("Avenger of the Fallen", avengerOfTheFallenX)},
	})
}

// avengerOfTheFallenX is the number of creature cards in your graveyard.
func avengerOfTheFallenX(g *game.Game, item *game.StackItem) int {
	return effects.B11CreatureCardsInGraveyard(g, item.Controller)
}
