package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primaris Eliminator — Creature — Astartes Warrior {4}{B}, 3/2 (Warhammer
// 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"When this creature enters, choose one —
//	 • Executioner Round — Destroy target creature.
//	 • Hyperfrag Round — Creatures target player controls get -2/-2 until
//	   end of turn."
//
// A modal enters trigger: the mode, and so its target, is chosen as the
// trigger goes on the stack (CR 603.3c). The -2/-2 applies to the creatures
// the player controls as it resolves (CR 611.2c), so a creature that enters
// later is not shrunk.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Primaris Eliminator — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Executioner Round — Destroy target creature.",
			effects.TargetCreature("target creature"),
			effects.DestroyTheModesTarget),
		effects.ModeWithPurpose(effects.ModeDoing("Hyperfrag Round — Creatures target player controls get -2/-2 until end of turn.",
			effects.TargetPlayer("target player"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				owner := t.ID
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Controller == owner }),
					Power:     -2,
					Toughness: -2,
					Label:     "Primaris Eliminator — Hyperfrag Round",
				}.Apply(ctx)
			}),
			game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, OpponentsOnly: true}}),
	)
	effects.Register(effects.Spec{
		OracleID:     "7d679591-f8ea-4c4c-ab98-7b9e3438cf57",
		Name:         "Primaris Eliminator",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
