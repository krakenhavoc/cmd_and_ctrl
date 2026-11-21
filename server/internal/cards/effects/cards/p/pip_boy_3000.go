package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pip-Boy 3000 — Artifact — Equipment {1} (Fallout, #5948):
//
//	"Whenever equipped creature attacks, choose one —
//	 • Sort Inventory — Draw a card, then discard a card.
//	 • Pick a Perk — Put a +1/+1 counter on that creature.
//	 • Check Map — Untap up to two target lands.
//	 Equip {2}"
//
// Argentum Armor's attack trigger (attachedCreatureAttacked, one trigger per
// declaration) with three modes. The loot is the shared body, the counter
// goes on the creature that attacked (still on the battlefield as that
// object, CR 608.2h) and the untap is an "up to two" target clause on its
// own mode, so the lands are only targeted when that mode is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e2ee630-9d37-4396-ae18-697c1a3a4e62",
		Name:         "Pip-Boy 3000",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			},
			Key:    "Pip-Boy 3000 — choose one",
			Effect: func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseOne(
				effects.ModeDoing("Sort Inventory — Draw a card, then discard a card.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.B16DrawThenDiscard(ctx.Game, item, 1)
					}),
				effects.ModeDoing("Pick a Perk — Put a +1/+1 counter on that creature.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						id := item.Trigger.Event.CardID
						if !effects.OnBattlefield(ctx.Game, id) {
							return nil
						}
						return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}),
				effects.ModeDoing("Check Map — Untap up to two target lands.",
					effects.TargetPermanent("up to two target lands", effects.Land()).WithCount(0, 2),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						for _, t := range ctx.ModeTargets(occ) {
							if !ctx.IsTargetLegal(t) || t.Kind != game.TargetCard {
								continue
							}
							if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
			),
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
