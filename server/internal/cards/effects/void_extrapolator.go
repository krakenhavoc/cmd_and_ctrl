package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Void Extrapolator // Omit Variables — Creature — Aetherborn Warlock
// {1}{B}, 2/2 // Sorcery {U/B} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 Threshold — This creature gets +1/+1 as long as there are seven or
//	 more cards in your graveyard."
//
//	Omit Variables — "Mill three cards."
//
// Threshold is read live on every layer recompute.
//
// No simplification.
func init() {
	const id = "517f84b5-cf99-410d-96ea-910740231dac"
	Register(Spec{
		OracleID:     id,
		Name:         "Void Extrapolator",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Static:       []game.StaticAbility{fraThresholdSelfPT(1, 1)},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Omit Variables",
		Completeness: CompletenessFull,
		OnResolve:    omitVariablesResolve,
	})
}
