package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bronzebeak Moa — Creature — Bird {2}{G}{W}, 2/2:
//
//	"Whenever another creature you control enters, this creature gets
//	 +3/+3 until end of turn."
//
// Beast-Kin Ranger's trigger at +3/+3: each other creature entering under
// its controller's control pumps it until cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16f03175-9c57-4058-9c9d-9c4936e4c324",
		Name:         "Bronzebeak Moa",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Bronzebeak Moa — +3/+3 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: ctx.Source(), Power: 3, Toughness: 3, Label: "Bronzebeak Moa — +3/+3"}.Apply(ctx)
				}),
		},
	})
}
