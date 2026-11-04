package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steel-Plume Marshal — Creature — Bird Soldier {3}{W}{W}, 3/3
// (Jumpstart 2020):
//
//	"Flying
//	 Whenever this creature attacks, other attacking creatures you
//	 control with flying get +2/+2 until end of turn."
//
// Consul's Lieutenant's attack pump with a flying test beside it. The
// set is fixed as the trigger resolves (CR 611.2c): an attacking
// creature you control, other than the Marshal, that has flying then.
// A creature that gains flying later this turn does not get the bonus,
// and one that loses it keeps the bonus.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:        "e7ed9c56-c388-4188-98e2-307d35ca3088",
		Name:            "Steel-Plume Marshal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Steel-Plume Marshal — other attacking creatures you control with flying get +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match:     And(AttackingCreature(), YouControl(), OtherThan(item.SourceCardID), HasKeyword("flying")),
						Power:     2,
						Toughness: 2,
						Label:     "Steel-Plume Marshal — +2/+2",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
