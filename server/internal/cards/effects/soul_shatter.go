package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Shatter — Instant {2}{B} (EDHREC rank 1156):
//
//	"Each opponent sacrifices a creature or planeswalker with the
//	 greatest mana value among creatures and planeswalkers they
//	 control."
//
// The edict that takes the biggest thing. EachPlayerSacrifices with a
// per-player predicate: the fan-out hands each opponent's own ID to
// the clause, which admits a creature or planeswalker whose mana
// value no other creature or planeswalker THAT player controls
// exceeds — ties all qualify and the player picks among them (CR
// 700.3). Not targeted, so hexproof does not help, and an opponent
// with no creature or planeswalker sacrifices nothing. Mana value is
// read off the battlefield, where a token is zero and X is zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "615927c2-3fb0-4e64-a1a7-55fb56de1423",
		Name:         "Soul Shatter",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				ExceptController: true,
				Match:            b10GreatestManaValueCreatureOrPlaneswalkerYouControl(),
				Label:            "a creature or planeswalker with the greatest mana value",
			}.Apply(ctx)
		},
	})
}
