package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanctuary Wall — Artifact Creature — Wall {1}{W}, 0/4 (Murders at
// Karlov Manor, #4469):
//
//	"Defender
//	 {2}{W}, {T}: Tap target creature. You may put a stun counter on it.
//	 If you do, put a stun counter on this creature."
//
// The tap happens either way. The stun counter is a "you may" asked as
// the ability resolves (CR 608.2d); only a "yes" that actually put the
// counter on the target (it is still on the battlefield) puts one on the
// Wall (CR 608.2c, "if you do"), and the Wall's counter needs the Wall to
// still be there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83856c3a-652e-4afc-9dc0-0072324ac36f",
		Name:            "Sanctuary Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature. You may put a stun counter on it. If you do, put a stun counter on this creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  sanctuaryWallTap,
		}},
	})
}

func sanctuaryWallTap(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id := effects.FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return effects.MayChoice{
		Question: "Sanctuary Wall — put a stun counter on the creature (and on Sanctuary Wall)?",
		Then:     sanctuaryWallStun,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// sanctuaryWallStun is the "yes": stun the target, then, if that
// happened, the Wall. An on-disk key.
var sanctuaryWallStun = effects.MayYes("confirm/sanctuary-wall-stun", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	target := effects.CarryAt(r, 0)
	if !effects.OnBattlefield(ctx.Game, target) {
		return nil
	}
	if err := (effects.AddCounter{Target: target, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if wall := ctx.Item.SourceCardID; effects.OnBattlefield(ctx.Game, wall) {
		return effects.AddCounter{Target: wall, Kind: game.CounterStun, N: 1}.Apply(ctx)
	}
	return nil
})
