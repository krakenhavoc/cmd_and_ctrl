package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Army of the Damned — Sorcery {5}{B}{B}{B}:
//
//	"Create thirteen tapped 2/2 black Zombie creature tokens.
//	 Flashback {7}{B}{B}{B} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// Thirteen real tokens, tapped as printed — CreateTokenAdvanced's
// EntersTapped spec, the same shape Stern Lesson's Powerstone uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "75d667ec-86f4-4850-a3b6-e7a9fc7053b0",
		Name:             "Army of the Damned",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{7}{B}{B}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateTokenAdvanced{
				Controller: ctx.Controller(),
				Spec:       Token(BlackZombieToken()).EntersTapped(),
				N:          13,
			}.Apply(ctx)
		},
	})
}
