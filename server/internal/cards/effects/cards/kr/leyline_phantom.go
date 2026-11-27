package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leyline Phantom — Creature — Illusion {4}{U}, 5/5 (Gatecrash, S82 OLD
// #5565):
//
//	"When this creature deals combat damage, return it to its owner's hand.
//	 (Return it only if it survived combat.)"
//
// One trigger for combat damage to anything (CR 510.2): it fires once per
// damage event, so damage to several blockers returns it once and the others
// find it gone. It resolves after the combat damage step's state-based
// actions, so a Phantom that was destroyed is in the graveyard and is not
// returned; it is bounced only while it is still the same object on the
// battlefield (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af642b8d-a085-4a07-8a1d-ca385d4700f7",
		Name:         "Leyline Phantom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, leylinePhantomDealtCombatDamage,
				"Leyline Phantom — return it to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if _, ok := ctx.SourceRef(); !ok || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.BounceToHand{Target: item.SourceCardID}.Apply(ctx)
				}),
		},
	})
}

// leylinePhantomDealtCombatDamage is "this creature deals combat damage": to a
// player, a creature, a planeswalker or a battle.
func leylinePhantomDealtCombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventDealDamage && ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID
}
