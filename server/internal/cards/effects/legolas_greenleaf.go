package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "beacbb51-51fe-47f7-a612-02cd93fbffdd",
		Name:            "Legolas Greenleaf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29AnotherLegendaryCreatureYouControlEntered(ev, source, g)
			}, "Legolas Greenleaf — put a +1/+1 counter on Legolas", putCounterOnSelf),
			WheneverThisDealsCombatDamageToAPlayer("Legolas Greenleaf — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
