package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Legolas Greenleaf — Legendary Creature — Elf Archer {2}{G}, 2/2:
//
//	"Reach
//	 Legolas can't be blocked by creatures with power 2 or less.
//	 Whenever another legendary creature you control enters, put a
//	 +1/+1 counter on Legolas.
//	 Whenever Legolas deals combat damage to a player, draw a card."
//
// The evasion clause is a pair rule on Legolas himself (#750, ADR 0045
// addendum Decision 11): CantBeBlockedBy(OnSelf(), PowerLE(2)). The
// blocker's power is read live when blockers are declared, after every
// layer and counter, so a 2/2 pumped to 3/3 during declare attackers
// can block him (CR 509.1b). Legolas losing all his abilities loses
// the rule with them.
//
// The other two clauses are Gimli of the Glittering Caves' shapes: the
// "another legendary creature you control enters" condition reads
// effective supertypes, and the counter lands only if Legolas is still
// on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "beacbb51-51fe-47f7-a612-02cd93fbffdd",
		Name:            "Legolas Greenleaf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B29AnotherLegendaryCreatureYouControlEntered(ev, source, g)
			}, "Legolas Greenleaf — put a +1/+1 counter on Legolas", effects.PutCounterOnSelf),
			effects.WheneverThisDealsCombatDamageToAPlayer("Legolas Greenleaf — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
