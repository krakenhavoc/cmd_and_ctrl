package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Bladecrafter — Creature — Faerie Rogue {2}{B}, 2/2 (Wilds of
// Eldraine Commander, #5368):
//
//	"Flying
//	 Whenever one or more Faeries you control deal combat damage to a
//	 player, put a +1/+1 counter on this creature.
//	 When this creature dies, each opponent loses X life and you gain X
//	 life, where X is its power."
//
// The counter trigger is "one or more": Alela's Faerie condition under
// OncePerBatch, so any number of Faeries connecting in one combat damage
// step is one trigger (CR 603.2c). The Bladecrafter is a Faerie and counts.
// The death trigger reads the power the creature last had on the
// battlefield (CR 603.10a, 608.2h) through the source's last-known
// information, counters and all; a power of zero or less drains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "164b064e-de50-468a-b874-b054897a5be9",
		Name:            "Faerie Bladecrafter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33FaerieYouControlDealtCombatDamageToPlayer(ev, source, g)
			}, "Faerie Bladecrafter — put a +1/+1 counter on this creature",
				Do(CounterOnThis{N: 1, Kind: game.CounterPlusOne}))),
			WhenThisDies("Faerie Bladecrafter — each opponent loses X life and you gain X life, where X is its power",
				func(g *game.Game, item *game.StackItem) error {
					info, ok := NewContext(g, item).SourcePermanent()
					if !ok || info.Power <= 0 {
						return nil
					}
					return b40DrainEachOpponent(info.Power)(g, item)
				}),
		},
	})
}
