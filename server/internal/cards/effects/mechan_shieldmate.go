package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mechan Shieldmate — Artifact Creature — Robot Soldier {1}{U}, 3/2 (Edge
// of Eternities, #4461):
//
//	"Defender
//	 As long as an artifact entered the battlefield under your control
//	 this turn, this creature can attack as though it didn't have
//	 defender."
//
// Hightide Hermit's waiver (the keyword stays, CR 609.4) as a self-only
// static switched on by the turn's entry tally for artifacts under its
// controller (EnteredWithCardTypeThisTurn): tokens count, and so does one
// that has since left. The Shieldmate's own entry is such an artifact, but
// it is summoning sick that turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d0025d29-d652-4af8-b367-13142d5532df",
		Name:            "Mechan Shieldmate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			avB18CanAttackAsThoughNoDefenderWhile(
				"As long as an artifact entered the battlefield under your control this turn, this creature can attack as though it didn't have defender.",
				func(g *game.Game, source *game.Card) bool {
					return g.EnteredWithCardTypeThisTurn(source.Controller, "artifact") > 0
				}),
		},
	})
}
