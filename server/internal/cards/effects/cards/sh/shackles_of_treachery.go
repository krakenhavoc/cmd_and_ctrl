package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shackles of Treachery — Sorcery {2}{R}:
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 Until end of turn, it gains haste and 'Whenever this creature deals damage,
//	 destroy target Equipment attached to it.'"
//
// Act of Treason's three clauses, with the haste and the granted trigger as one
// until-end-of-turn grant (Perigee Beckoner's GrantAbilitiesFor): one effect, one
// timestamp, pinned to the creature as it resolves (CR 611.2c). The granted
// trigger watches every damage the creature deals, combat or not, and its target
// is an Equipment attached to that creature, chosen as the trigger goes on the
// stack and judged against it again as it resolves (TargetSpec.AttachedToTheSource);
// with none attached nothing triggers a target and the trigger is removed
// (CR 603.3d). The Equipment is destroyed whoever controls it.
//
// No simplification.
func init() {
	const grant = "shackles-of-treachery/destroy-attached-equipment"
	effects.Register(effects.Spec{
		OracleID:     "3466e0be-3e59-4f5e-812e-3250c8bf19aa",
		Name:         "Shackles of Treachery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.On(game.EventDealDamage, effects.ThisDealtDamage,
						"Shackles of Treachery — destroy target Equipment attached to this creature", effects.DestroyTheFirstLegalTarget),
					effects.AttachedToTheSource(effects.TargetPermanent("target Equipment attached to it", effects.HasSubtype("Equipment")))),
			},
			Text: "Whenever this creature deals damage, destroy target Equipment attached to it.",
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GainControl{Target: t.ID, Duration: effects.DurationUntilEndOfTurn(ctx),
					Label: "Shackles of Treachery — gain control until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"},
					Label: "Shackles of Treachery — haste until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantAbilitiesFor{Target: t.ID, Keys: []string{grant},
					Label: "Shackles of Treachery — it destroys an Equipment attached to it when it deals damage"}.Apply(ctx)
			}
			return nil
		},
	})
}
