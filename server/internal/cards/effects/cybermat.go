package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cybermat — Artifact Creature — Robot {2}, 2/1 (Doctor Who, #5948):
//
//	"Skulk (This creature can't be blocked by creatures with greater power.)
//	 Whenever this creature attacks and isn't blocked, it gets +X/+0 until end
//	 of turn, where X is the number of attacking artifact creatures."
//
// Skulk is a printed keyword. The trigger is CR 509.3's completed block
// declaration (WhenAttacksAndIsNotBlockedEffect); X is counted as it
// resolves (CR 608.2h), including Cybermat itself while it is still attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca173221-1661-41f6-a353-29a32d6ccce0",
		Name:            "Cybermat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlockedEffect("Cybermat — it gets +X/+0 until end of turn, where X is the number of attacking artifact creatures",
				func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
					x := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.IsArtifact() && c.AttackingTarget != uuid.Nil {
							x++
						}
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Cybermat — +X/+0"}.Apply(NewContext(g, item))
				}),
		},
	})
}
