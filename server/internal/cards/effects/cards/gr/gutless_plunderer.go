package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gutless Plunderer — Creature — Skeleton Pirate {2}{B}, 2/2 (Foundations,
// #4465):
//
//	"Deathtouch
//	 Raid — When this creature enters, if you attacked this turn, look at
//	 the top three cards of your library. You may put one of those cards
//	 back on top of your library. Put the rest into your graveyard."
//
// Gorehorn Raider's raid enters trigger (an intervening "if", CR 603.4).
// The look is private; the optional pick keeps one card on top and the rest
// are put, not milled, into the graveyard (stdB15LookKeepOneOnTopRestToGraveyard).
//
// No simplification.
func init() {
	const label = "Gutless Plunderer — Raid: look at the top three cards of your library, you may put one back on top and the rest into your graveyard"
	effects.Register(effects.Spec{
		OracleID:        "31265357-5ecb-4d83-80ff-1795b1dfdecb",
		Name:            "Gutless Plunderer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn), label,
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.StdB15LookKeepOneOnTopRestToGraveyard(ctx, 3,
							"Gutless Plunderer — you may put one of those cards back on top of your library; the rest go into your graveyard")
					}),
				}})),
		},
	})
}
