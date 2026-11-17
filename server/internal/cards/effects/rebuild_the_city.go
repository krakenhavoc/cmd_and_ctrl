package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rebuild the City — Sorcery {3}{B}{R}{G} (March of the Machine: The
// Aftermath, #5476):
//
//	"Choose target land. Create three tokens that are copies of it, except
//	 they're 3/3 creatures in addition to their other types and they have
//	 vigilance and menace. (They're affected by summoning sickness.)"
//
// Applied Geometry's shape (CR 707.2, CR 707.9): the copiable values of the
// land as it is on resolution, with the exception applied to the template:
// base power and toughness 3/3 (a printed 3, not a `*` stand-in), the
// creature card type added to its own types, and vigilance and menace added
// to what it has. The three tokens are made by one creation. A land that
// has left the battlefield is an illegal target and the spell does nothing
// (CR 608.2b). The tokens are new objects, so they are summoning-sick like
// any creature that came under your control this turn (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2aac0ddc-1005-4049-a337-fe86e092b461",
		Name:         "Rebuild the City",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return CreateTokenCopy{
						Controller: ctx.Controller(),
						Copy:       t.ID,
						N:          3,
						Except: func(tok *game.Card) {
							tok.Power, tok.Toughness = 3, 3
							tok.VariableToughness = false
							tok.PrintedPTKnown = true
							tok.TypeLine = cmdrB23WithCardTypesInAddition(tok.TypeLine, "Creature")
							for _, kw := range []string{"vigilance", "menace"} {
								if !slices.Contains(tok.Keywords, kw) {
									tok.Keywords = append(tok.Keywords, kw)
								}
							}
						},
					}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
