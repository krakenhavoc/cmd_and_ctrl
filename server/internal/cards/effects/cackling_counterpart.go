package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cackling Counterpart — Instant for {1}{U}{U}:
//
//	"Create a token that's a copy of target creature you control.
//	 Flashback {5}{U}{U}"
//
// The first flashback card with a target clause, and the reason the
// cast path had to ride the Board's ordinary prompt chain rather
// than the bare-payload route the exile impulse button uses: a
// graveyard cast of this has to open the targeting UI exactly as a
// hand cast does.
//
// The token is a copy of the copiable values only (CR 707.2), which
// CreateTokenCopy already gets right — counters, damage and auras on
// the original are not copied, and the copy carries the original's
// oracle ID so every catalog hook keyed on it comes along.
//
// Sandbox simplification, inherited from CreateTokenCopy and stated
// here because it is invisible otherwise: the token's ETB *triggered*
// abilities fire, but a copied card whose ETB lives in Spec.AsEnters
// rather than Spec.Triggered does not get that clause.
func init() {
	Register(Spec{
		OracleID:         "9e2adca5-f39c-4a09-bcce-8238ebac2c4a",
		Name:             "Cackling Counterpart",
		Completeness:     CompletenessCaveats,
		Caveats:          []string{"The token copy skips the enters-the-battlefield effect of a card whose entry is an on-enter hook rather than a trigger."},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}{U}")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return CreateTokenCopy{
					Controller: item.Controller,
					Copy:       t.ID,
					N:          1,
				}.Apply(ctx)
			}
			// Every target left in response (CR 608.2c): the spell
			// does as much as it can, which is nothing.
			return nil
		},
	})
}
