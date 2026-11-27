package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidewater Minion — Creature — Elemental Minion {3}{U}{U}, 4/4 (Ravnica: City
// of Guilds, S82 OLD batch 574, #5600):
//
//	"Defender
//	 {4}: This creature loses defender until end of turn.
//	 {T}: Untap target permanent."
//
// Defender is the printed keyword; the first ability removes it until end of
// turn through the layered view. The second taps the Minion as a cost and
// untaps any permanent as it resolves (the Minion itself included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "80b4a9f7-008d-4979-a3df-7637fae68167",
		Name:            "Tidewater Minion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}: This creature loses defender until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{4}"),
				Effect:  effects.ThisLosesKeywordUntilEndOfTurn("defender", "Tidewater Minion — loses defender until end of turn"),
			},
			{
				Label:   "{T}: Untap target permanent.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target permanent"),
				Effect:  effects.B22UntapEachLegalTarget,
			},
		},
	})
}
