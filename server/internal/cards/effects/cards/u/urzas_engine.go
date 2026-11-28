package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Engine — Artifact Creature — Juggernaut {5}, 1/5 (Alliances, #5634):
//
//	"Trample
//	 {3}: This creature gains banding until end of turn.
//	 {3}: Attacking creatures banded with this creature gain trample until end
//	 of turn."
//
// Banding is the engine's own (CR 702.22, ADR 0224): a keyword granted until
// end of turn is read at the attack declaration like a printed one. The second
// ability gives trample to the creatures still in the Engine's band as it
// resolves (Game.BandMatesForEffect), the Engine itself not counted among those
// "banded with" it.
//
// No simplification.
func init() {
	const name = "Urza's Engine"
	effects.Register(effects.Spec{
		OracleID:        "bf6ea878-044b-40a8-a21b-0191a948a69a",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}: This creature gains banding until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{game.KeywordBanding}, Label: name + " — banding until end of turn"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{3}: Attacking creatures banded with this creature gain trample until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{3}"),
				Effect:  effects.OldB815GrantTrampleToTheBand,
			},
		},
	})
}
