package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Reaver Cleaver — Legendary Artifact — Equipment {2}{R}:
//
//	"Equipped creature gets +1/+1 and has trample and 'Whenever this
//	 creature deals combat damage to a player or planeswalker, create
//	 that many Treasure tokens.'
//	 Equip {3}"
//
// The granted triggered ability is modelled the Sword-of-Fire-and-Ice
// way: the trigger lives on the EQUIPMENT, watching for the creature
// it is currently attached to dealing combat damage, rather than
// being written onto the creature's own ability list. The two are
// observably identical here — nothing cares whether the source of the
// treasure-creation is "the Equipment" or "the creature's granted
// ability" — and keeping it on the Equipment means the trigger goes
// away the instant the Cleaver is unattached or destroyed, with no
// separate bookkeeping.
//
// "A player or planeswalker" is wider than any Sword's "a player", so
// it gets its own condition (dealtCombatDamageToPlayerOrPlaneswalkerByAttached)
// rather than reusing attachedCreatureDealtCombatDamageToPlayer.
//
// The Treasure count reads ctx.Trigger().Event.Amount at RESOLUTION —
// the damage amount, as the trigger-context doc comment names it —
// rather than a value captured in Build, so a replacement that changed
// the damage in flight (Torbran, a prevention shield) is reflected in
// however many Treasures actually land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37a2a31d-51e6-4c07-b4c2-b206ad40eb42",
		Name:         "The Reaver Cleaver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.DealtCombatDamageToPlayerOrPlaneswalkerByAttached(ev, source, g)
			},
			Key:    "The Reaver Cleaver — create that many Treasure tokens",
			Effect: theReaverCleaverCreateTreasures,
		}},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}

func theReaverCleaverCreateTreasures(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	n := ctx.Trigger().Event.Amount
	if n <= 0 {
		return nil
	}
	return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: n}.Apply(ctx)
}
