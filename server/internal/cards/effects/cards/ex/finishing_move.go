package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Finishing Move — Sorcery {2}{G}:
//
//	"You get {TK}{TK}, then you may put a sticker on a nonland permanent
//	 you own. Target creature you control deals damage equal to its power
//	 to target creature you don't control."
//
// Three steps in printed order (CR 608.2c), written as ResolveSteps so the
// damage waits for the sticker questions (ADR 0208 §4): a power and
// toughness sticker put on the biter changes the damage it deals. The
// bite is Bite Down's (oneSidedBite), its power read as it deals.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d41dc3a5-58e4-444a-b388-5a32ccbed0eb",
		Name:         "Finishing Move",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.GetTickets{N: 2},
			effects.PutASticker{},
			effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(ctx.Item, ctx) }),
		),
	})
}
