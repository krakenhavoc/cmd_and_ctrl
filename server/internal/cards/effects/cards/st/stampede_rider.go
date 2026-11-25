package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stampede Rider — Creature — Satyr {2}{R}, 2/3 (Theros Beyond Death, S82 OLD,
// #5514):
//
//	"Trample
//	 At the beginning of each combat, if you control a creature with power 4
//	 or greater, this creature gets +1/+1 until end of turn."
//
// The trigger fires at the beginning of EVERY player's combat and carries the
// intervening if (CR 603.4): it is checked as the step begins, and again as
// the trigger would resolve, against the creatures its controller controls
// (the Rider itself counts once it has power 4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "580112bd-6804-4981-97ad-0f5e43627b22",
		Name:            "Stampede Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a creature with power 4 or greater",
				func(g *game.Game, you uuid.UUID) bool {
					return len(effects.PermanentsControlledByMatching(g, you, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return effects.MatchCreatureWithPowerAtLeast(4)(c)
					})) > 0
				},
				effects.AtBeginningOfEachCombat("Stampede Rider — this creature gets +1/+1 until end of turn",
					effects.ThisGetsUntilEndOfTurn(1, 1, "Stampede Rider — +1/+1 until end of turn"))),
		},
	})
}
