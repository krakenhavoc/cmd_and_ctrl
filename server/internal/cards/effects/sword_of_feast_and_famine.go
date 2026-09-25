package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Feast and Famine — Artifact — Equipment for {3}:
//
//	"Equipped creature gets +2/+2 and has protection from black and
//	 from green.
//	 Whenever equipped creature deals combat damage to a player, that
//	 player discards a card and you untap all lands you control.
//	 Equip {2}"
//
// The untap clause is the reason this Sword is the one people build
// around: connecting refunds your whole mana base, so the attack and
// the second main phase are both fully funded. That half is exact
// here.
//
// PROTECTION FROM BLACK AND FROM GREEN is a layer-6 grant to the
// equipped creature (GrantToAttached), enforced since #662 /
// ADR 0072: a black or green spell cannot target the wearer
// (CR 702.16b), a black or green Aura falls off it and a black or
// green Equipment unattaches (CR 702.16c-d), black or green damage to
// it is prevented (CR 702.16e), and a black or green creature cannot
// block it (CR 702.16f). Against the two colours that print the most
// removal, that is most of why the card is played.
//
// The discard used to be random with a caveat saying so, because the
// engine's only discard surface here was DiscardRandomForEffect.
// #651 made an effect's discard a real prompt addressed to the
// discarding player, so the damaged player now chooses their own card
// as printed (CR 701.8a) and that caveat is gone. The untap is not
// behind a "then" — it happens as the trigger resolves, while the
// discard prompt is still open — which is the printed card: the lands
// come back whether or not the opponent has decided yet.
func init() {
	Register(Spec{
		OracleID:     "d0901053-6de0-46d0-9ee3-8d40510236c1",
		Name:         "Sword of Feast and Famine",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from black", "protection from green"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Sword of Feast and Famine — discard, untap your lands",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player: item.Trigger.Event.Target,
					Source: item.SourceCardID,
					N:      1,
				})
				return untapAllLandsControlledBy(g, item.Controller, ctx)
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
