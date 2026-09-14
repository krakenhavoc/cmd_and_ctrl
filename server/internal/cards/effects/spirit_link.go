package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spirit Link — Enchantment — Aura for {W} (EDHREC rank 3094):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage, you gain that much
//	 life."
//
// Lifelink you can put on somebody else's creature, which is the
// whole trick: Spirit Link on the biggest attacker at the table means
// every swing it makes — at you or at anyone — pays you.
//
// THE WIDEST DAMAGE CONDITION IN THE ATTACHMENT CATALOG. The printed
// text says "deals damage", full stop: combat damage, an activated
// ping, damage to a player, damage to a creature, damage to YOU. So
// this trigger cannot reuse attachedCreatureDealtDamageToOpponent
// (which filters to opponents) or the Swords' combat condition — it
// is the unfiltered form, and the only guards are that the damage
// came from the enchanted creature and that it was more than zero.
//
// "That much" is read off the event's Amount, which is the damage
// actually dealt after any reduction, so a prevented or reduced hit
// gains the reduced number.
//
// The life goes to the AURA's controller (CR 109.5's "you"), not to
// the creature's — the difference the card is played for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c77ff526-c0a8-45c7-9730-2e306a0d01b8",
		Name:         "Spirit Link",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.IsAttachedTo(ev.Source)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				amount := ev.Amount
				return game.NewTriggeredItem(source, "Spirit Link — gain that much life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: amount}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
