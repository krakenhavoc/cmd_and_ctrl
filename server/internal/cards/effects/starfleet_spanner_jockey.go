package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfleet Spanner Jockey — Creature — Human Artificer {1}{U}, 3/3
// (Star Trek):
//
//	Defender
//	As long as you've cast a noncreature spell this turn, this creature
//	can attack as though it didn't have defender.
//
// A self-only layer-6 static writing an unrestricted DefenderWaiver (ADR
// 0106's 2026-10-16 amendment, #1936) while its controller's cast tally for
// the turn holds a noncreature spell. The keyword stays (CR 702.3b is waived,
// not removed), so anything counting creatures with defender still counts
// it. "You" is the creature's controller, read live at declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d41ff489-7d74-4b06-b3a1-638f3e3303d1",
		Name:            "Starfleet Spanner Jockey",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			avB18CanAttackAsThoughNoDefenderWhile(
				"As long as you've cast a noncreature spell this turn, this creature can attack as though it didn't have defender.",
				func(g *game.Game, source *game.Card) bool {
					return mscM13CastNoncreatureSpellThisTurn(g, source.Controller)
				}),
		},
	})
}
