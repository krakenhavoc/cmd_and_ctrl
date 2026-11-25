package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mogg Assassin — Creature — Goblin Assassin {2}{R}, 2/1 (#6171):
//
//	"{T}: You choose target creature an opponent controls, and that
//	 opponent chooses target creature. Flip a coin. If you win the flip,
//	 destroy the creature you chose. If you lose the flip, destroy the
//	 creature your opponent chose."
//
// Both creatures are targets of the ability (CR 115.1, CR 601.2c via CR
// 602.2b). The first clause is the controller's own; the second is
// OfTheFirstTargetsController: an opponent's-choice clause whose chooser
// the engine names as the controller of the first target, so in a
// multiplayer game no other opponent chooses and nobody is asked "which
// opponent". The second clause is any creature, judged for the ability's
// controller, so the opponent may choose one of their own, the Assassin
// itself, or the creature already chosen.
//
// The flip is the keyed coin-flip continuation: won, the first clause's
// creature is destroyed; lost, the second's. A target that is no longer
// legal as the ability resolves is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36a6925a-621e-47d6-8cb3-1b2886e1986e",
		Name:         "Mogg Assassin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: You choose target creature an opponent controls, and that opponent chooses target creature. " +
				"Flip a coin. If you win the flip, destroy the creature you chose. If you lose the flip, destroy the creature your opponent chose.",
			Cost: effects.TapCost(),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				effects.OfTheFirstTargetsController(effects.TargetCreature("target creature chosen by that opponent")),
			),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Mogg Assassin — call the coin flip",
					ThenKey:  moggAssassinFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

var moggAssassinFlipped = effects.CoinFlipThen("coin-flip/mogg-assassin", func(ctx *effects.Context, r game.CoinFlipped) error {
	slot := 1
	if effects.WonFirstFlip(r) {
		slot = 0
	}
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
})
