package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Éowyn, Shieldmaiden — Legendary Creature — Human Knight {2}{U}{R}{W},
// 5/4 (EDHREC rank 4476):
//
//	"First strike
//	 At the beginning of combat on your turn, if another Human entered
//	 the battlefield under your control this turn, create two 2/2 red
//	 Human Knight creature tokens with trample and haste. Then if you
//	 control six or more Humans, draw a card."
//
// The Jeskai Human deck's payoff. Two hasty tramplers every combat is
// a clock on its own; the draw turns a wide Human board into a
// refill. The gate is the interesting part and it is the reason the
// card is built around cheap Humans rather than expensive ones: you
// have to have PLAYED one this turn.
//
// "If another Human entered … this turn" is an intervening-if clause
// (CR 603.4), so it is checked twice — once when the trigger would go
// on the stack, and again as it resolves. It reads the engine's
// per-turn subtype tally (EnteredWithSubtypeThisTurn), which counts
// permanents by the subtype they HAD as they entered, so a Human that
// entered and then died this turn still counts and a changeling
// counts. "ANOTHER" excludes Éowyn herself: when she is one of the
// Humans that entered this turn, the tally has to reach two.
//
// The "then if" is a second, separate check made as the trigger
// resolves, over the Humans on the battlefield at that moment (the
// Knight tokens are Humans, so they count towards the six they just
// helped make — which is printed, and is how the card actually
// draws).
//
// First strike rides PrintedKeywords; the tokens carry trample and
// haste, both canonical keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98a2a642-a30a-43e7-b47f-4776c22d1cb0",
		Name:            "Éowyn, Shieldmaiden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventStepBegan},
			AppliesTo: AllOf(
				StepBegan(game.StepBeginCombat, true),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b43AnotherHumanEnteredThisTurn(g, source)
				},
			),
			Key:    "Éowyn, Shieldmaiden — two 2/2 Human Knights, then draw on six Humans",
			Effect: eowynShieldmaidenKnights,
		}},
	})
}

// eowynShieldmaidenKnights is the combat trigger's resolution: the
// intervening if again, two Knights, then the draw on six Humans.
func eowynShieldmaidenKnights(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !b43AnotherHumanEnteredThisTurn(g, &src) {
		// CR 603.4: the intervening if is re-checked as the ability
		// resolves, and a false answer removes it from the stack doing
		// nothing.
		return nil
	}
	if err := (CreateToken{
		Controller: item.Controller,
		Template:   TokenCard("2/2 red Human Knight with trample and haste"),
		N:          2,
	}).Apply(ctx); err != nil {
		return err
	}
	if b43CreaturesOfSubtypeControlled(g, item.Controller, "Human") < 6 {
		return nil
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
