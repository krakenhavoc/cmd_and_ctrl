package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resounding Roar — Instant {1}{G} (Shards of Alara):
//
//	"Target creature gets +3/+3 until end of turn.
//	 Cycling {5}{R}{G}{W} ({5}{R}{G}{W}, Discard this card: Draw a card.)
//	 When you cycle this card, target creature gets +6/+6 until end of turn."
//
// The spell is a plain until-end-of-turn pump of a target creature. The cycle
// trigger is Choking Tethers' graveyard trigger (CR 702.29c), mandatory, with
// its own target, +6/+6; both targets are re-judged as they resolve
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40ad5a52-37ad-4524-be06-9e263a907d31",
		Name:         "Resounding Roar",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return oldB508PumpFirstLegalTarget(ctx, 3, "Resounding Roar — +3/+3")
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{5}{R}{G}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.OldB508CycleTrigger("Resounding Roar — cycled: target creature gets +6/+6 until end of turn",
				effects.TargetCreature("target creature"),
				func(g *game.Game, item *game.StackItem) error {
					return oldB508PumpFirstLegalTarget(effects.NewContext(g, item), 6, "Resounding Roar — +6/+6")
				}),
		},
	})
}

// oldB508PumpFirstLegalTarget gives the first legal target creature +n/+n
// until end of turn.
func oldB508PumpFirstLegalTarget(ctx *effects.Context, n int, label string) error {
	id := effects.FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: label}.Apply(ctx)
}
