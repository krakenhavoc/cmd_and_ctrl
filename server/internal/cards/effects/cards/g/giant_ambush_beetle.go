package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Ambush Beetle — Creature — Insect {3}{B/G}{R}, 4/3 (Alara Reborn, #5586):
//
//	"Haste
//	 When this creature enters, you may have target creature block it this turn
//	 if able."
//
// The target is chosen as the trigger goes on the stack and judged again as it
// resolves (CR 608.2b); the requirement is the Beetle itself, as the object it
// is then (CR 400.7), recorded until end of turn (Hunt Down's blocks-attacker
// record). It asks nothing unless the Beetle attacks and the creature can
// legally block it (CR 509.1c). The "may" is asked as it resolves.
//
// No simplification.
func init() {
	const label = "Giant Ambush Beetle — you may have target creature block it this turn if able"
	effects.Register(effects.Spec{
		OracleID:        "c2dbdf9d-3ac1-47d8-9ad8-57e7888e5cdd",
		Name:            "Giant Ambush Beetle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						ref, ok := g.PermanentRefForEffect(item.SourceCardID)
						if !ok {
							return nil
						}
						return effects.BlocksAttackerUntilEOT{Blocker: t.ID, Attacker: ref, Label: "Giant Ambush Beetle — blocks it if able"}.Apply(ctx)
					}),
					"Giant Ambush Beetle — have the target creature block it this turn if able?"),
				effects.TargetCreature("target creature")),
		},
	})
}
