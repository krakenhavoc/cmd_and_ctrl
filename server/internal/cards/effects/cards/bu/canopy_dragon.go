package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Canopy Dragon — Creature — Dragon {4}{G}{G}, 4/4 (Mirage, S82 OLD b808):
//
//	"Trample
//	 {1}{G}: This creature gains flying and loses trample until end of turn."
//
// One effect, one timestamp (CR 613.7): a layer 6 grant of flying and removal
// of trample, locked to this object as the ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a9e59ee4-02a6-498c-82f0-f4f2ff334d65",
		Name:            "Canopy Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: This creature gains flying and loses trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
					"Canopy Dragon — gains flying and loses trample",
					game.AddKeywordsMod("flying"), game.RemoveKeywordsMod("trample"))
			},
		}},
	})
}
