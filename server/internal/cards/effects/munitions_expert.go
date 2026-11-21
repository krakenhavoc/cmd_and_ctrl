package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Munitions Expert — Creature — Goblin {B}{R}, 1/1 (Modern Horizons 3):
//
//	"Flash
//	 When this creature enters, you may have it deal damage to target creature
//	 or planeswalker equal to the number of Goblins you control."
//
// The Goblins are counted as the ability resolves, Munitions Expert included.
func init() {
	const label = "Munitions Expert — you may have it deal damage to target creature or planeswalker equal to the number of Goblins you control"
	enters := Optional(Targeting(WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		n := len(permanentsControlledByMatching(g, item.Controller, Subtype("Goblin")))
		if n <= 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
		}
		return nil
	}), TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))),
		"Munitions Expert — have it deal damage equal to the number of Goblins you control?")
	Register(Spec{
		OracleID:        "a0b16b11-69d9-4e94-aae3-298aa44ce3d4",
		Name:            "Munitions Expert",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
