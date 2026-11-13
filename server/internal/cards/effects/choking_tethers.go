package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a60348fa-b182-4667-9e38-15bdf3511f82",
		Name:         "Choking Tethers",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to four target creatures").WithCount(0, 4),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tapEachLegalTarget(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					InGraveyard(On(game.EventCycle, Self, "Choking Tethers — cycled: you may tap target creature",
						tapTheFirstLegalTarget)),
					"Choking Tethers — tap target creature?"),
				TargetCreature("target creature"),
			),
		},
	})
}
