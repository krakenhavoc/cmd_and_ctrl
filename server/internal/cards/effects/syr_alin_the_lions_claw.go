package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syr Alin, the Lion's Claw — Legendary Creature — Human Knight
// {3}{W}{W}, 4/4 (Foundations, #4465):
//
//	"First strike (This creature deals combat damage before creatures
//	 without first strike.)
//	 Whenever Syr Alin attacks, other creatures you control get +1/+1
//	 until end of turn."
//
// Restless Prairie's attack trigger: "other creatures you control" is the
// set as the trigger resolves (CR 611.2c), so Syr Alin is never in it and
// a creature that arrives afterwards gets nothing. First strike rides
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d9a4905-4a26-410e-936c-56de849e9ca3",
		Name:            "Syr Alin, the Lion's Claw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Syr Alin, the Lion's Claw — other creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl(), OtherThan(item.SourceCardID)), Power: 1, Toughness: 1,
						Label: "Syr Alin, the Lion's Claw — other creatures you control get +1/+1",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
