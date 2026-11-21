package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiendlash — Artifact — Equipment {1}{R} (Adventures in the Forgotten Realms
// Commander):
//
//	"Equipped creature gets +2/+0 and has reach.
//	 Whenever equipped creature is dealt damage, it deals damage equal to its
//	 power to target player or planeswalker.
//	 Equip {2}{R}"
//
// Blazing Sunsteel's watch on the attached creature with a target picked as
// the ability goes on the stack. The engine opens one damage event per source,
// so the trigger waits for the damage to finish (AtBatchEnd, ThatMuchDamage):
// a creature dealt damage by two blockers at once triggers once (CR 120.3,
// CR 510.2, CR 603.2c). The damage is the creature's power as the ability
// resolves, from its last-known information if it died to the damage
// (CR 608.2h), and the creature is the source of it.
//
// No simplification.
func init() {
	const label = "Fiendlash — equipped creature deals damage equal to its power to target player or planeswalker"
	effects.Register(effects.Spec{
		OracleID:     "2d21dc80-2900-4dc3-9657-f17b50716481",
		Name:         "Fiendlash",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("reach"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
			},
			Targets:    effects.TargetPlayerOrPlaneswalker(),
			Key:        label,
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     fiendlashStrike,
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{R}")},
	})
}

// fiendlashStrike is the trigger's body: the damaged creature deals damage
// equal to its power to the target.
func fiendlashStrike(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	host := item.Trigger.Event.Target
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{Source: host, Target: t.ID, Amount: effects.B43PowerNowOrLastKnown(g, host)}.Apply(ctx)
	}
	return nil
}
