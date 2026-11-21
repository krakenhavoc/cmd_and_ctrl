package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Irenicus's Vile Duplication — Sorcery {3}{U}:
//
//	"Create a token that's a copy of target creature you control,
//	 except the token has flying and it isn't legendary."
//
// Cackling Counterpart's target and copy, with Helm of the Host's
// "isn't legendary" template edit and a flying keyword added the way
// TokenCopyGainsHaste adds haste. Both exceptions are part of the
// copy's copiable values (CR 707.9b), so a copy of the token keeps
// them. The token carries the original's oracle ID, so its own
// triggers, statics and activated abilities come along.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3faf698-1dac-45ef-9c43-c8cead489ffc",
		Name:         "Irenicus's Vile Duplication",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.CreateTokenCopy{
					Controller: item.Controller,
					Copy:       t.ID,
					N:          1,
					Except: func(tok *game.Card) {
						tok.TypeLine = effects.RemoveLegendaryFromTypeLine(tok.TypeLine)
						if !effects.KeywordSliceContains(tok.Keywords, "flying") {
							tok.Keywords = append(tok.Keywords, "flying")
						}
					},
				}.Apply(ctx)
			}
			return nil
		},
	})
}
