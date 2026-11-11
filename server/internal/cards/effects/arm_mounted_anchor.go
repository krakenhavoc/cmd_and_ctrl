package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arm-Mounted Anchor — Artifact — Equipment {3} (Lost Caverns of Ixalan
// Commander):
//
//	"Equipped creature gets +2/+2 and has menace.
//	 Whenever equipped creature deals combat damage to a player, draw two
//	 cards. Then discard two cards unless you discard a Pirate card.
//	 Equip {2}. This ability costs {2} less to activate if you have one or
//	 fewer cards in hand."
//
// The pump and menace are the shared attachment statics. The trigger is the
// Swords' condition (attachedCreatureDealtCombatDamageToPlayer) with Thirst
// for Knowledge's "draw, then discard two unless one qualifies" continuation
// (discardTwoUnlessOne): the drawn cards are among the discard choices.
//
// The discount is the equip ability's OWN cost clause, read when the cost is
// determined (CR 601.2f): the activating player's hand then.
//
// No simplification.
func init() {
	equip := EquipAbility("{2}")
	equip.CostModifiers = []game.CostModifier{
		CostsLessEach(cmdr2B21HandAtMostOneDiscount,
			"This ability costs {2} less to activate if you have one or fewer cards in hand."),
	}
	Register(Spec{
		OracleID:     "70f3bc78-201f-4c56-b87b-9e8f2db9f7f7",
		Name:         "Arm-Mounted Anchor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Arm-Mounted Anchor — draw two cards, then discard two cards unless you discard a Pirate card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
					Ref: cmdr2B21AnchorDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
				})
			},
		}},
		Activated: []ActivatedAbility{equip},
	})
}

// cmdr2B21AnchorDiscard is "then discard two cards unless you discard a
// Pirate card". An on-disk key: never renamed, never reused.
var cmdr2B21AnchorDiscard = discardTwoUnlessOne("arm-mounted-anchor-discard-two-unless-pirate",
	"Arm-Mounted Anchor — discard two cards, or one Pirate card",
	func(c game.Card) bool { return cmdr2B21HasSubtype(c, "Pirate") })
