package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akoum Battlesinger — Creature — Human Berserker Ally {1}{R}, 1/1 (Worldwake, #5582):
//
//	"Haste
//	 Whenever this creature or another Ally you control enters, you may
//	 have Ally creatures you control get +1/+0 until end of turn."
//
// The Ally trigger is South Pole Voyager's (wwkThisOrAnotherAllyEntered),
// and the "you may" is asked as it resolves (CR 603.5). The Allies are
// the ones its controller controls as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb448589-28e1-49d4-ba3e-7ddb4341fdc5",
		Name:            "Akoum Battlesinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, wwkThisOrAnotherAllyEntered,
				"Akoum Battlesinger — Ally creatures you control get +1/+0 until end of turn", akoumBattlesingerPump),
				"Akoum Battlesinger — have Ally creatures you control get +1/+0 until end of turn?"),
		},
	})
}

// akoumBattlesingerPump gives each Ally creature its controller controls
// +1/+0 until end of turn.
func akoumBattlesingerPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return BoostUntilEOT{
		Match: And(Creature(), YouControl(), Subtype(ctx.TextSubtype("Ally"))),
		Power: 1,
		Label: "Akoum Battlesinger — +1/+0",
	}.Apply(ctx)
}
