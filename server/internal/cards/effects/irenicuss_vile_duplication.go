package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a3faf698-1dac-45ef-9c43-c8cead489ffc",
		Name:         "Irenicus's Vile Duplication",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return CreateTokenCopy{
					Controller: item.Controller,
					Copy:       t.ID,
					N:          1,
					Except: func(tok *game.Card) {
						tok.TypeLine = removeLegendaryFromTypeLine(tok.TypeLine)
						if !keywordSliceContains(tok.Keywords, "flying") {
							tok.Keywords = append(tok.Keywords, "flying")
						}
					},
				}.Apply(ctx)
			}
			return nil
		},
	})
}
