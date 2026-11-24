package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last Night Together — Sorcery {3}{R}{G} (Streets of New Capenna):
//
//	"Choose two target creatures. Untap them. Put two +1/+1 counters on
//	 each of them. They gain vigilance, indestructible, and haste until
//	 end of turn. After this main phase, there is an additional combat
//	 phase. Only the chosen creatures can attack during that combat
//	 phase."
//
// The untap, the counters and the keywords reach only the targets still
// legal at resolution (CR 608.2b). The additional combat comes directly
// after this main phase (CR 500.8), with no main phase after it.
//
// "Only the chosen creatures can attack during that combat phase" is a
// rule bound to that combat and to no other (#6087, ADR 0188 Decision
// 13): "can't attack" on every creature except the chosen ones, read
// live, so a creature that arrives later can't attack in it either, and
// neither can a chosen creature that has left and come back as a new
// object (CR 400.7). The chosen creatures are both targets as they stand
// on the battlefield at resolution: a target that has become illegal but
// is still there is not restricted, since a rules effect does not apply
// to an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Last Night Together"
	effects.Register(effects.Spec{
		OracleID:     "0ad48edf-49e2-4f3a-b50e-11231a616752",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance", "indestructible", "haste")),
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var chosen, legal []uuid.UUID
			for _, t := range ctx.Targets() {
				if t.Kind != game.TargetCard {
					continue
				}
				chosen = append(chosen, t.ID)
				if ctx.IsTargetLegal(t) {
					legal = append(legal, t.ID)
				}
			}
			if err := effects.UntapEach(ctx, legal); err != nil {
				return err
			}
			for _, id := range legal {
				if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, id := range legal {
				if err := (effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"vigilance", "indestructible", "haste"},
					Label:    name + " — vigilance, indestructible, and haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			ids := ctx.Game.AddPhasesForEffect(ctx.Source(), game.PhaseAnchor{Kind: game.AnchorThisMainPhase}, game.PhaseKindCombat)
			if len(ids) == 0 {
				return nil
			}
			ctx.Game.OnlyTheseCanAttackDuringCombatPhaseForEffect(ids[0], ctx.Source(), ctx.Controller(), chosen,
				name+" — only the chosen creatures can attack during this combat")
			return nil
		},
	})
}
