package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adventurous Eater // Have a Bite — Creature — Human Warlock {2}{B}, 3/2,
// with the prepare spell Have a Bite, Sorcery {B} (Secrets of Strixhaven,
// #4456):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Have a Bite: "Put a +1/+1 counter on target creature. You gain 1
//	 life."
//
// A preparation card (CR 722): the permanent under the bare oracle ID and
// its prepare spell under "#1". The spell is Seed Suture's text word for
// word, so it shares that body (seedSutureResolve).
//
// No simplification.
func init() {
	const oracle = "02125db4-4507-467a-ac0c-406de2c7d533"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Adventurous Eater",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Have a Bite",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    seedSutureResolve,
	})
}
