package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Visions of Duplicity — Sorcery {2}{U} (Midnight Hunt Commander):
//
//	"Exchange control of two target creatures you don't control.
//	 Flashback {8}{U}{U}. This spell costs {X} less to cast this way, where X
//	 is the greatest mana value of a commander you own on the battlefield or
//	 in the command zone."
//
// Switcheroo's exchange (CR 701.12) with both targets restricted to
// creatures the caster doesn't control, so the two creatures swap
// controllers between two opponents (or do nothing if one controller owns
// both). All or nothing (CR 701.12a): if either target is illegal as the
// spell resolves, nothing is exchanged. No duration. The flashback discount
// is the Visions cycle's (visionsFlashbackDiscount).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "ba85f0b0-b08c-4e62-a56e-b3ab7d699175",
		Name:              "Visions of Duplicity",
		Completeness:      effects.CompletenessFull,
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:  []game.AlternativeCost{effects.VisionsFlashbackAlternative("{8}{U}{U}")},
		SelfCostModifiers: []game.CostModifier{effects.VisionsFlashbackDiscount()},
		Targets:           effects.TargetCreature("two target creatures you don't control", effects.OpponentControls()).WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) != 2 {
				return nil
			}
			return effects.ExchangeControl{
				A:     legal[0].ID,
				B:     legal[1].ID,
				Label: "Visions of Duplicity — exchange control",
			}.Apply(ctx)
		},
	})
}
