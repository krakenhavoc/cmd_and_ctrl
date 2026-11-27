package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aphotic Wisps — Instant {B} (Shadowmoor, #5591):
//
//	"Target creature becomes black and gains fear until end of turn.
//	 Draw a card."
//
// Crimson Wisps' shape in black with fear. A target that is gone as the
// spell resolves fizzles the whole spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db951ec2-d751-479d-af1f-885d3cf50027",
		Name:         "Aphotic Wisps",
		Purpose:      aphoticWispsPurpose(),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.OldB537WispsResolve("Aphotic Wisps", "B", "black",
			func(ctx *effects.Context, t game.TargetRef) error {
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"fear"}, Label: "Aphotic Wisps — fear"}.Apply(ctx)
			}),
	})
}

func aphoticWispsPurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "fear"))
	p.Draws = 1
	return p
}
