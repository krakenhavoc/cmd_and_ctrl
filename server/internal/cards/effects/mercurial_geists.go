package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mercurial Geists — Creature — Spirit {2}{U}{R}, 1/3 (Eldritch Moon,
// #5541):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, this creature gets
//	 +3/+0 until end of turn."
//
// Kiln Fiend's cast trigger on a flyer.
//
// No simplification.
func init() {
	const label = "Mercurial Geists — +3/+0 until end of turn"
	Register(Spec{
		OracleID:        "1f6bff6c-a2f0-4ed7-8169-acb4d31a7187",
		Name:            "Mercurial Geists",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), label, thisGetsUntilEndOfTurn(3, 0, label)),
		},
	})
}
