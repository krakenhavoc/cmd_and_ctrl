package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:          "ef47c447-bd7f-43c6-b4db-4a848c39820c",
		Name:              "Visions of Glory",
		Completeness:      effects.CompletenessFull,
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:  []game.AlternativeCost{effects.VisionsFlashbackAlternative("{8}{W}{W}")},
		SelfCostModifiers: []game.CostModifier{effects.VisionsFlashbackDiscount()},
		Purpose:           game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.B34CreaturesYouControlMatching(ctx.Game, ctx.Controller(), func(game.Card) bool { return true })
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: n}.Apply(ctx)
		},
	})
}
