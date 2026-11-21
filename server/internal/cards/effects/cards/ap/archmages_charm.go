package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archmage's Charm — Instant {U}{U}{U}:
//
//	"Choose one —
//	 • Counter target spell.
//	 • Target player draws two cards.
//	 • Gain control of target nonland permanent with mana value 1 or less."
//
// The third mode is permanent (CR 611.2a: no duration is stated), so it
// is an indefinite control effect. The mana value is read through the
// layered view at announce and again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85cbb035-25df-43bb-ae06-ab098f448b02",
		Name:         "Archmage's Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Counter target spell.", effects.TargetSpell("target spell")),
				game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.Mode("Target player draws two cards.", effects.TargetPlayer("target player")),
				game.Purpose{Draws: 2}),
			effects.Mode("Gain control of target nonland permanent with mana value 1 or less.",
				effects.TargetPermanent("target nonland permanent with mana value 1 or less", effects.Nonland(), effects.ManaValueLE(1))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			t := item.Targets[0]
			switch {
			case ctx.HasMode(0):
				if t.Kind != game.TargetCard {
					return nil
				}
				return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
			case ctx.HasMode(1):
				if t.Kind != game.TargetPlayer {
					return nil
				}
				return effects.DrawCards{Player: t.ID, N: 2}.Apply(ctx)
			case ctx.HasMode(2):
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				return effects.GainControl{Target: t.ID, Duration: game.IndefiniteDuration(),
					Label: "Archmage's Charm — gain control"}.Apply(ctx)
			}
			return nil
		},
	})
}
