package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whiplash Wordsmith // Vicious Verse — Creature — Vampire Sorcerer
// {3}{B/R}, 3/3 // Sorcery {B/R} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 As long as an opponent was dealt noncombat damage this turn, this
//	 creature has flying and haste."
//
//	Vicious Verse — "Vicious Verse deals 1 damage to target opponent."
//
// Weaker than printed: the engine's turn tally records noncombat damage
// by the SOURCE's controller, not by who was dealt it, and a static
// ability only recomputes on a zone or counter change, so "an opponent
// was dealt noncombat damage this turn" has nothing to read and nothing
// to wake it. Rather than guess, the flying and haste line is left out;
// the card is the 3/3 body and Vicious Verse.
func init() {
	const id = "e1ffb884-a89e-4e2a-9beb-8bd5838b6a94"
	Register(Spec{
		OracleID:     id,
		Name:         "Whiplash Wordsmith",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The flying and haste it gets after an opponent is dealt noncombat damage isn't implemented — it is a 3/3 with no abilities of its own."},
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Vicious Verse",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    viciousVerseResolve,
	})
}
