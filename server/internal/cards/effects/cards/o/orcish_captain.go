package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orcish Captain — Creature — Orc Warrior {R}, 1/1 (Fallen Empires, S82 OLD
// b854, #5637):
//
//	"{1}: Flip a coin. If you win the flip, target Orc creature gets +2/+0 until
//	 end of turn. If you lose the flip, it gets -0/-2 until end of turn."
//
// The Orc is chosen as the ability goes on the stack (CR 602.2b), then the
// controller calls the flip as it resolves (CR 705.2, as Mogg Assassin's): won,
// +2/+0, lost, -0/-2, each only if the target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76c06de7-39ff-4b80-8e54-20910372cffd",
		Name:         "Orcish Captain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Flip a coin. If you win the flip, target Orc creature gets +2/+0 until end of turn. If you lose the flip, it gets -0/-2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target Orc creature", effects.HasSubtype("Orc")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Orcish Captain — call the coin flip",
					ThenKey:  orcishCaptainFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

var orcishCaptainFlipped = effects.CoinFlipThen("coin-flip/orcish-captain", func(ctx *effects.Context, r game.CoinFlipped) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	if effects.WonFirstFlip(r) {
		return effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Orcish Captain — +2/+0 until end of turn"}.Apply(ctx)
	}
	return effects.BoostUntilEOT{Target: t.ID, Toughness: -2, Label: "Orcish Captain — -0/-2 until end of turn"}.Apply(ctx)
})
