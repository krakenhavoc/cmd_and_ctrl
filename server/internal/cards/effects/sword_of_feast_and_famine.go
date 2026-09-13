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
// TWO SIMPLIFICATIONS, both strictly weaker than printed:
//
//   - PROTECTION FROM BLACK AND FROM GREEN is not granted.
//     Protection is a parameterised keyword and CR 702.16e tests the
//     quality against the SOURCE of a spell or ability, which the
//     engine's targeting choke point never receives — see
//     ADR 0038 for why that is a structural gap and not a to-do.
//     Omitting it loses the Sword its evasion and its removal
//     protection against two colours, which makes the card worse,
//     not better.
//   - THE DISCARD IS RANDOM. Printed, the damaged player chooses.
//     The engine's discard surface is DiscardRandomForEffect and
//     there is no "opponent chooses" prompt yet; random is the
//     established stand-in across the catalog. Random is not
//     uniformly weaker in theory, but it is the house convention and
//     changing it is a discard-prompt problem, not a Sword problem.
//
// Deferred to whichever sprint lands protection (CR 702.16) and an
// opponent-chooses discard prompt.
func init() {
	Register(Spec{
		OracleID:     "d0901053-6de0-46d0-9ee3-8d40510236c1",
		Name:         "Sword of Feast and Famine",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Protection from black and from green isn't granted.", "The damaged player discards at random instead of choosing a card."},
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
						if err := (DiscardCards{Player: damaged, N: 1}.Apply(ctx)); err != nil {
							return err
						}
						return untapAllLandsControlledBy(g, item.Controller, ctx)
					})
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
