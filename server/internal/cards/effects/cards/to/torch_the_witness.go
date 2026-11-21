package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torch the Witness — Sorcery {X}{R} (Murders at Karlov Manor, #4469):
//
//	"Torch the Witness deals twice X damage to target creature. If excess
//	 damage was dealt to that creature this way, investigate. (Create a
//	 Clue token. It's an artifact with '{2}, Sacrifice this token: Draw a
//	 card.')"
//
// Archaic's Agony's excess reading (socB20DealDamageWithExcess): the
// damage that actually landed beyond what was lethal, after the damage
// marked earlier this turn (CR 120.4a), so a prevented or reduced hit
// counts only what got through. Any excess at all makes one Clue
// (CR 701.16a); none, or a target that left in response, makes none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51473069-68dc-4e87-b345-d7cbcfdeea74",
		Name:         "Torch the Witness",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := effects.SocB20DealDamageWithExcess(ctx, t.ID, 2*ctx.X())
				if err != nil || excess <= 0 {
					return err
				}
				return effects.Investigate{N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
