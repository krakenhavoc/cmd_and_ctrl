package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orzhov Pontiff — Creature — Human Cleric {1}{W}{B}, 1/1 (Guildpact):
//
//	"Haunt (When this creature dies, exile it haunting target creature.)
//	 When this creature enters or the creature it haunts dies, choose one —
//	 • Creatures you control get +1/+1 until end of turn.
//	 • Creatures you don't control get -1/-1 until end of turn."
//
// Haunt is CR 702.55 (#6083). The modal ability is one ability with two
// conditions; from exile "you" is the card's owner (CR 108.4), who chooses
// the mode as it is put on the stack (CR 700.2a). Each mode's set is locked
// as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	pontiff := effects.WhenThisEntersOrTheCreatureItHauntsDies("Orzhov Pontiff — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	pontiff.Modes = effects.ChooseOne(
		effects.ModeDoing("Creatures you control get +1/+1 until end of turn.", nil, orzhovPontiffPump),
		effects.ModeDoing("Creatures you don't control get -1/-1 until end of turn.", nil, orzhovPontiffShrink),
	)
	effects.Register(effects.Spec{
		OracleID:     "3620ef25-d9e2-4460-b25f-592ff7df681b",
		Name:         "Orzhov Pontiff",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Haunt("Orzhov Pontiff"), pontiff},
	})
}

func orzhovPontiffPump(_ *game.StackItem, ctx *effects.Context, _ int) error {
	return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
		Label: "Orzhov Pontiff — +1/+1"}.Apply(ctx)
}

func orzhovPontiffShrink(_ *game.StackItem, ctx *effects.Context, _ int) error {
	return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.Not(effects.YouControl())), Power: -1, Toughness: -1,
		Label: "Orzhov Pontiff — -1/-1"}.Apply(ctx)
}
