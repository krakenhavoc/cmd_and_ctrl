package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Mar-Vell, Space-Born — Legendary Creature — Kree Soldier Hero
// {4}{W}, 4/4 (Marvel Super Heroes):
//
//	"Flying, vigilance
//	 Cosmic Awareness — As long as an opponent has cast a spell this turn,
//	 you may cast spells as though they had flash."
//
// Vedalken Orrery's sentence on a body, gated: one Spec.CastTimings entry
// (CastTimingRule.OpponentCastASpellThisTurn), derived from the battlefield
// on every query. The condition is read live from the other players' cast
// tallies, so a Mar-Vell that enters after the opponent's spell is covered at
// once, the window shuts when the turn ends and when Mar-Vell leaves or loses
// its abilities, and nothing sits on the stack first (a flash-granting trigger
// would do all three wrong). It is about casting, so it does not let its
// controller play a land at instant speed.
//
// No simplification.
func init() {
	const label = "As long as an opponent has cast a spell this turn, you may cast spells as though they had flash."
	Register(Spec{
		OracleID:        "b74c072d-77ac-460d-9d9c-75b25874f58b",
		Name:            "Captain Mar-Vell, Space-Born",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		CastTimings:     []game.CastTimingRule{CastAsThoughFlashWhileAnOpponentHasCast(label)},
	})
}
