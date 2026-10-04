package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Antler Skulkin — Artifact Creature — Scarecrow {5}, 3/3:
//
//	"{2}: Target white creature gains persist until end of turn."
//
// "White" is the creature's colour as the ability is put on the stack
// and again as it resolves (CR 608.2b). Persist is the engine's
// (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30572dd1-c982-4872-9f35-a91ed8b238f7",
		Name:         "Antler Skulkin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: Target white creature gains persist until end of turn.",
			Cost:    ManaCost("{2}"),
			Targets: TargetCreature("target white creature", OfColor("W")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return grantEachLegalTargetUntilEOT(NewContext(g, item), game.KeywordPersist,
					"Antler Skulkin — persist until end of turn")
			},
		}},
	})
}
