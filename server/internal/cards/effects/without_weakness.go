package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Without Weakness — Instant {1}{B} (Hour of Devastation, #5534):
//
//	"Target creature you control gains indestructible until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Divine Resilience's grant without its kicker: a target that left in
// response does nothing (CR 608.2b). The shared Cycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fffa9334-3576-4f70-9605-2ad062cdbc69",
		Name:         "Without Weakness",
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"indestructible"},
				Label:    "Without Weakness — indestructible until end of turn",
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
