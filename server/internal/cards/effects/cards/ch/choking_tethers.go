package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Choking Tethers — Instant {3}{U} (Onslaught):
//
//	"Tap up to four target creatures.
//	 Cycling {1}{U}
//	 When you cycle this card, you may tap target creature."
//
// The spell taps each of its targets still legal as it resolves (CR
// 608.2b, tapEachLegalTarget); zero targets is a legal cast. Cycling is
// the keyword's own hand ability, and the cycle trigger fires from the
// graveyard, where the card already is (CR 702.29c), Howler's Heavy's
// InGraveyard(On(EventCycle, Self, …)). Its target is chosen as it goes on
// the stack, and the "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a60348fa-b182-4667-9e38-15bdf3511f82",
		Name:         "Choking Tethers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to four target creatures").WithCount(0, 4),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Choking Tethers — cycled: you may tap target creature",
						effects.TapTheFirstLegalTarget)),
					"Choking Tethers — tap target creature?"),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
