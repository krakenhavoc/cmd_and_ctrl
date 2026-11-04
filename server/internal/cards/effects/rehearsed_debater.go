package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rehearsed Debater — Creature — Djinn Bard {2}{W}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Vigilance
//	 Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, this creature gets +1/+1 until end of turn."
//
// Repartee is Informed Inkwright's trigger (informedInkwrightCast: the cast
// spell's own announced targets). The pump belongs to this permanent and is
// skipped if it has left or come back as a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e30ccf07-0969-4e81-ba43-c50a2ff9c040",
		Name:            "Rehearsed Debater",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast,
				"Rehearsed Debater — Repartee: +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Rehearsed Debater — pump"}.Apply(NewContext(g, item))
				}),
		},
	})
}
