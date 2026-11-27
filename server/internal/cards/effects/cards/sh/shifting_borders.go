package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shifting Borders — Instant — Arcane {3}{U} (Saviors of Kamigawa, #5601):
//
//	"Exchange control of two target lands.
//	 Splice onto Arcane {3}{U} (As you cast an Arcane spell, you may reveal
//	 this card from your hand and pay its splice cost. If you do, add this
//	 card's effects to that spell.)"
//
// Shifting Grift's exchange (CR 701.12) over two target lands. It is all or
// nothing: if either target is illegal as the spell resolves, neither land
// changes hands (CR 701.12a, 608.2b). Spliced onto an Arcane spell its
// text and its two targets ride the host spell (CR 702.47a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b377f717-d0c2-4db5-aba2-fa4bfa35de0f",
		Name:         "Shifting Borders",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
		Splice:       effects.SpliceOntoArcane("{3}{U}"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) != 2 {
				return nil
			}
			return effects.ExchangeControl{A: legal[0].ID, B: legal[1].ID, Label: "Shifting Borders — exchange control of two lands"}.Apply(ctx)
		},
	})
}
