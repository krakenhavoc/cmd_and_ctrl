package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archon of Coronation — Creature — Archon {4}{W}{W}, 5/5:
//
//	"Flying
//	 When this creature enters, you become the monarch.
//	 As long as you're the monarch, damage doesn't cause you to lose life.
//	 (When a creature deals combat damage to you, its controller still
//	 becomes the monarch.)"
//
// The static is a CR 614 replacement on the life the damage would cost
// you (ReplacementEvent.LifeFromDamage, the window CR 120.3a opens), cancelled
// while you are the monarch. The damage is still dealt: lifelink, "is dealt
// damage" triggers and the monarch steal on combat damage (CR 724.2) all see it,
// and a drain or a life payment is not damage, so it costs life as usual.
// It is read from the battlefield, so it ends when the Archon leaves or
// loses its abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61f55e39-4028-4c91-910e-16b88c73beb5",
		Name:            "Archon of Coronation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifeFromDamage && ev.LifeDelta < 0 &&
					src != nil && src.Controller != uuid.Nil && ev.LifePlayer == src.Controller &&
					g.Monarch == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Archon of Coronation — damage doesn't cause you to lose life while you're the monarch",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Archon of Coronation — you become the monarch", Do(BecomeTheMonarch{})),
		},
	})
}
