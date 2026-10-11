package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glistening Oil — Enchantment — Aura {B}{B}:
//
//		"Enchant creature
//		 Enchanted creature has infect.
//		 At the beginning of your upkeep, put a -1/-1 counter on enchanted
//		 creature.
//		 When this Aura is put into a graveyard from the battlefield,
//		 return it to its owner's hand."
//
//	  - "Enchant creature" is the ordinary Aura target clause, re-checked
//	    as a state-based action (CR 704.5m), so an Aura whose creature has
//	    gone falls into the graveyard.
//	  - Infect is a layer 6 grant on the host (GrantToAttached), so the
//	    engine's damage tail turns the creature's damage into poison and
//	    -1/-1 counters.
//	  - The upkeep trigger is YOUR upkeep, the Aura's controller's, not
//	    the creature's. It looks the host up as it resolves and does
//	    nothing if the Aura has fallen off by then.
//	  - The recursion fires on any trip from the battlefield to a
//	    graveyard (cardDied), the host dying included, and returns the
//	    card only while it is still the graveyard object that triggered
//	    it (returnSelfFromGraveyardToHand).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "645a4dee-a4e0-4850-a781-522aa718827f",
		Name:         "Glistening Oil",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			GrantToAttached(game.KeywordInfect),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Glistening Oil — put a -1/-1 counter on enchanted creature", glisteningOilCounter),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			}, "Glistening Oil — return it to its owner's hand", returnSelfFromGraveyardToHand),
		},
	})
}

func glisteningOilCounter(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return AddCounter{Target: host.InstanceID, Kind: game.CounterMinusOne, N: 1}.Apply(NewContext(g, item))
}
