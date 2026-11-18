package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visions of Glory — Sorcery {4}{W} (Midnight Hunt Commander):
//
//	"Create a 1/1 white Human creature token for each creature you control.
//	 Flashback {8}{W}{W}. This spell costs {X} less to cast this way, where X
//	 is the greatest mana value of a commander you own on the battlefield or
//	 in the command zone."
//
// The tokens are counted as the spell resolves (CR 608.2). The flashback
// discount is the Visions cycle's (visionsFlashbackDiscount).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "ef47c447-bd7f-43c6-b4db-4a848c39820c",
		Name:              "Visions of Glory",
		Completeness:      CompletenessFull,
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:  []game.AlternativeCost{visionsFlashbackAlternative("{8}{W}{W}")},
		SelfCostModifiers: []game.CostModifier{visionsFlashbackDiscount()},
		Purpose:           game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := b34CreaturesYouControlMatching(ctx.Game, ctx.Controller(), func(game.Card) bool { return true })
			return CreateToken{Template: TokenCard("1/1 white Human"), N: n}.Apply(ctx)
		},
	})
}
