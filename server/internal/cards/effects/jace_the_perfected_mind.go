package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace, the Perfected Mind — Legendary Planeswalker — Jace {2}{U}{U/P},
// starting loyalty 5 (Phyrexia: All Will Be One, #5479):
//
//	"Compleated ({U/P} can be paid with {U} or 2 life. If life was paid,
//	 this planeswalker enters with two fewer loyalty counters.)
//	 +1: Until your next turn, up to one target creature gets −3/−0.
//	 −2: Target player mills three cards. Then if a graveyard has twenty or
//	     more cards in it, you draw three cards. Otherwise, you draw a card.
//	 −X: Target player mills three times X cards."
//
// Compleated is ADR 0183's replacement (effects.Compleated). The +1 is a
// modification pinned to the target
// until the activator's next turn (DurationUntilYourNextTurn). The −2's
// "Then" waits for the mill to finish (CR 608.2c; the mill is the keyed
// one), and "a graveyard" is any one player's graveyard, counted after the
// mill, so a graveyard that was already nearly full counts. The −X mills
// three times the announced X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "03dafab4-841f-41cb-8f2e-8188f5177837",
		Name:            "Jace, the Perfected Mind",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{Compleated()},
		// X = 0 mills nothing.
		XMatters: true,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Until your next turn, up to one target creature gets -3/-0.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  jacePerfectedMindPlusOne,
			},
			{
				Label:   "−2: Target player mills three cards. Then if a graveyard has twenty or more cards in it, you draw three cards. Otherwise, you draw a card.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetPlayer("target player"),
				Effect: Do(
					Instruction(func(ctx *Context) error { return JacePerfectedMindMill(ctx, 3) }),
					Instruction(jacePerfectedMindDraw),
				),
			},
			{
				Label:   "−X: Target player mills three times X cards.",
				Cost:    LoyaltyMinusX(),
				Targets: TargetPlayer("target player"),
				Effect: Do(Instruction(func(ctx *Context) error {
					return JacePerfectedMindMill(ctx, 3*ctx.X())
				})),
			},
		},
	})
}

func jacePerfectedMindPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return ScopedEffectFor{
		Target:   targets[0].ID,
		Mods:     []game.Mod{game.ModifyPTMod(-3, 0)},
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Jace, the Perfected Mind — -3/-0 until your next turn",
	}.Apply(ctx)
}

// jacePerfectedMindMill mills n cards from the still-legal target player.
func JacePerfectedMindMill(ctx *Context, n int) error {
	player, ok := FirstLegalPlayerTarget(ctx)
	if !ok || n <= 0 {
		return nil
	}
	return MillEachPlayerInTurn(ctx, []uuid.UUID{player}, ConstantMill(n), game.KeyedTail{})
}

// jacePerfectedMindDraw is the −2's second sentence: three cards when any
// player's graveyard holds twenty or more, otherwise one.
func jacePerfectedMindDraw(ctx *Context) error {
	n := 1
	for _, p := range ctx.Game.Seats {
		if p != nil && p.Graveyard != nil && len(p.Graveyard.Cards) >= 20 {
			n = 3
			break
		}
	}
	return DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
}
