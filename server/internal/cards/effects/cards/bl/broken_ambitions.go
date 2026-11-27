package cardsbl

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Broken Ambitions — Instant {X}{U} (Lorwyn, #5593):
//
//	"Counter target spell unless its controller pays {X}. Clash with an
//	 opponent. If you win, that spell's controller mills four cards. (Each
//	 clashing player reveals the top card of their library, then puts that
//	 card on their choice of the top or bottom. A player wins if their card
//	 had a greater mana value.)"
//
// Condescend's pay-unless with X announced at the cast, then the clash. The
// spell's controller is read before the counter-unless is raised (CR
// 608.2h), so a spell that has left the stack still has its last controller
// mill on a win; they are the player whose spell it was, not the caster.
// With the spell's one target illegal as this resolves the whole spell does
// nothing (CR 608.2b), clash included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a3fba0c-50b1-4139-834e-c8fc8e1315bb",
		Name:         "Broken Ambitions",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			var controller uuid.UUID
			if spell := ctx.Game.StackItemForEffect(targets[0].ID); spell != nil {
				controller = spell.Controller
			}
			x := ctx.X()
			if err := (effects.CounterUnlessPaid{
				StackID:  targets[0].ID,
				Cost:     "{" + strconv.Itoa(x) + "}",
				Question: "Broken Ambitions — pay {" + strconv.Itoa(x) + "} or your spell is countered",
			}).Apply(ctx); err != nil {
				return err
			}
			then := ctx.ItemTail(brokenAmbitionsClashed)
			then.Carry = []uuid.UUID{controller}
			return effects.ClashWithAnOpponent{Then: then}.Apply(ctx)
		},
	})
}

// brokenAmbitionsClashed is "If you win, that spell's controller mills four
// cards". An on-disk key.
var brokenAmbitionsClashed = effects.CardTail("then/broken-ambitions-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) || len(o.Tail.Carry) == 0 || o.Tail.Carry[0] == uuid.Nil {
		return nil
	}
	return effects.MillCards{Player: o.Tail.Carry[0], N: 4}.Apply(ctx)
})
