package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vein Ripper — Creature — Vampire Assassin {3}{B}{B}{B}, 6/5:
//
//	"Flying
//	 Ward—Sacrifice a creature.
//	 Whenever a creature dies, target opponent loses 2 life and you
//	 gain 2 life."
//
// The sacrifice-cost ward. When the ward trigger resolves, the spell's
// controller is asked whether to sacrifice a creature; yes chains a
// pick over the creatures THEY control (effects/ward.go), and no — or
// having no creature at all — counters the spell. Removal aimed at the
// Ripper therefore costs a card and a body, and the body it costs
// drains through the Ripper's own trigger before the removal resolves.
//
// The drain is Vengeful Bloodwitch's targeted trigger with a wider
// condition: ANY creature dying, anyone's, the Ripper included (a
// leaves-the-battlefield trigger looks back, CR 603.10a). The
// controller picks the opponent as the trigger goes on the stack, and
// the drain does nothing if that player has left by resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f83a768c-162d-46f0-8ebd-d9c6b2c69322",
		Name:            "Vein Ripper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardSacrifice("a creature", Creature()), "Vein Ripper — ward, sacrifice a creature"),
			{
				Watches: []game.EventKind{game.EventLTB},
				Key:     "Vein Ripper — target opponent loses 2 life and you gain 2 life",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b17SelfOrAnotherCreatureDied(ev, source, g)
				},
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return targetOpponentLosesAndYouGain(g, item, 2)
				},
			},
		},
	})
}
