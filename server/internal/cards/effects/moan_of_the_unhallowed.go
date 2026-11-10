package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moan of the Unhallowed — Sorcery {2}{B}{B} (Foundations Commander):
//
//	"Create two 2/2 black Zombie creature tokens.
//	 Flashback {5}{B}{B}"
//
// Flashback is the shared alternative cost (CR 702.34a): castable from the
// graveyard at sorcery speed for {5}{B}{B}, exiled as it leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ce87e5ed-2562-4563-9a5b-bd53c04ec4a5",
		Name:             "Moan of the Unhallowed",
		Purpose:          game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{CostWithPurpose(Flashback("{5}{B}{B}"), game.Purpose{Tokens: 2, CreatureTokens: 2})},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 black Zombie"), N: 2}.Apply(ctx)
		},
	})
}
