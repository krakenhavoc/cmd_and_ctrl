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
// ONE SIMPLIFICATION, strictly weaker than printed:
//
//   - PROTECTION FROM BLACK AND FROM GREEN is not granted.
//     Protection is a parameterised keyword and CR 702.16b tests the
//     quality against the SOURCE of a spell or ability, which the
//     engine's targeting choke point never receives — see
//     ADR 0038 for why that is a structural gap and not a to-do.
//     Omitting it loses the Sword its evasion and its removal
//     protection against two colours, which makes the card worse,
//     not better.
//
// The discard used to be random with a caveat saying so, because the
// engine's only discard surface here was DiscardRandomForEffect.
// #651 made an effect's discard a real prompt addressed to the
// discarding player, so the damaged player now chooses their own card
// as printed (CR 701.8a) and that caveat is gone. The untap is not
// behind a "then" — it happens as the trigger resolves, while the
// discard prompt is still open — which is the printed card: the lands
// come back whether or not the opponent has decided yet.
//
// Deferred until protection lands (CR 702.16, #662).
func init() {
	Register(Spec{
		OracleID:     "d0901053-6de0-46d0-9ee3-8d40510236c1",
		Name:         "Sword of Feast and Famine",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Protection from black and from green isn't granted."},
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				damaged := ev.Target
				return game.NewTriggeredItem(source, "Sword of Feast and Famine — discard, untap your lands",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player: damaged,
							Source: item.SourceCardID,
							N:      1,
						})
						return untapAllLandsControlledBy(g, item.Controller, ctx)
					})
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
