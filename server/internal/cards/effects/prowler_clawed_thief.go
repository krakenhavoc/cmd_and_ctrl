package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prowler, Clawed Thief — Legendary Creature — Human Rogue Villain {1}{U}{B},
// 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Menace
//	 Whenever another Villain you control enters, Prowler connives."
//
// Baron Strucker's "another Villain you control enters" condition with the
// shared sourceConnives body (A.I.M. Scientists, Atlantean Skirmisher):
// draw, then discard, then a +1/+1 counter on Prowler for a nonland card.
// If Prowler has left by then the loot still happens and the counter is
// skipped (CR 701.50c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "76742bfc-ae99-4b06-b29d-e0e41d92ae14",
		Name:            "Prowler, Clawed Thief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return mshB2AnotherVillainEntered(ev, source, g)
			}, "Prowler — it connives", Do(Instruction(sourceConnives))),
		},
	})
}
