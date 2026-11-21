package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// As Foretold — Enchantment {2}{U} (Enchanting Tales):
//
//	"At the beginning of your upkeep, put a time counter on this enchantment.
//	 Once each turn, you may pay {0} rather than pay the mana cost for a spell
//	 you cast with mana value X or less, where X is the number of time counters
//	 on this enchantment."
//
// The upkeep trigger is Phyrexian Arena's shape and puts a time counter on the
// enchantment. The second sentence is a granted alternative cost (CR 118.9,
// ADR 0118) like Warped Space's, offered for every spell its controller casts
// from any zone where the printed mana cost could be paid. The cap is read off
// this enchantment as the offer is derived (MaxManaValueOfSource), an {X} in
// the spell's cost counting as 0 (CR 107.3b), and the once-each-turn limit is
// per As Foretold, spent as the spell becomes cast. Paying {0} is an
// alternative cost, so additional costs and cost increases still apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4b3e0af-5caa-47c1-b5be-c56a875cbb15",
		Name:         "As Foretold",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("As Foretold — put a time counter on this enchantment", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.PayZeroForASpellUpToTimeCountersOnceEachTurn()},
	})
}
