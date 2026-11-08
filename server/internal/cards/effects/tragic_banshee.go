package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tragic Banshee — Creature — Spirit {4}{B}, 5/3 (Foundations, #4465):
//
//	"Morbid — When this creature enters, target creature an opponent
//	 controls gets -1/-1 until end of turn. If a creature died this turn,
//	 that creature gets -13/-13 until end of turn instead."
//
// Ambush Gigapede's enters trigger with Tragic Slip's morbid "instead":
// the turn's tally of creatures that died is read once as the trigger
// resolves, and the target gets one shrink, -1/-1 or -13/-13, never both.
// Morbid is not an intervening "if" here, so the trigger always goes on
// the stack. A target that has left or changed control by resolution
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "888f58c0-1860-4df4-ad75-53e634150680",
		Name:         "Tragic Banshee",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Tragic Banshee — target creature an opponent controls gets -1/-1 (morbid: -13/-13) until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					n, label := 1, "Tragic Banshee — -1/-1 until end of turn"
					if b11CreaturesDiedThisTurn(g) > 0 {
						n, label = 13, "Tragic Banshee — morbid, -13/-13 until end of turn"
					}
					return BoostUntilEOT{Target: id, Power: -n, Toughness: -n, Label: label}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, 0, 0))),
		},
	})
}
