package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Irradiate — Instant {3}{B} (Mirrodin, #5606):
//
//	"Target creature gets -1/-1 until end of turn for each artifact you
//	 control."
//
// Blight-Breath Catoblepas' -X/-X over a count taken as the spell resolves
// (CR 608.2h): the artifacts its controller controls then, so a response that
// sacrifices or adds artifacts changes the size.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84d45389-a085-44bc-a3fb-1a5f7cc6cbe0",
		Name:         "Irradiate",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := effects.B03ArtifactsControlled(ctx.Game, ctx.Controller())
			return effects.BoostUntilEOT{Target: id, Power: -n, Toughness: -n, Label: "Irradiate — -1/-1 for each artifact you control"}.Apply(ctx)
		},
	})
}
