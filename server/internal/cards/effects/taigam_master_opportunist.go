package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taigam, Master Opportunist — Legendary Creature — Human Monk {1}{U},
// 2/2 (Tarkir: Dragonstorm, #4463):
//
//	"Flurry — Whenever you cast your second spell each turn, copy it, then
//	 exile the spell you cast with four time counters on it. If it doesn't
//	 have suspend, it gains suspend."
//
// The trigger is Cori Mountain Stalwart's flurry (the cast path bumps the
// per-turn tally before it emits EventCast, so a total of two is the
// second spell; spells cast before Taigam count). "Copy it" names the
// spell without targeting it, so a spell countered in response is still
// copied from last-known information (CR 608.2h) and the copy, made on
// the stack above the original, can't be given new targets or modes, and
// is not cast. Then the original is exiled from the stack (ExileTargetSpell,
// not a counter, CR 701.6 never applies) with Kang Prime's tail
// (game.GainSuspendExileTail): four time counters, and suspend if the card
// has none, so it comes back at the beginning of its owner's upkeeps and is
// cast free when the last counter is removed (CR 702.62a). A spell that
// already left the stack is not exiled and nothing more happens, and a
// commander the owner sent to the command zone instead was never exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49f2cbc3-bd03-48e6-b1c4-d7f315b8579d",
		Name:         "Taigam, Master Opportunist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Taigam, Master Opportunist — flurry: copy it, then exile the spell with four time counters; it gains suspend",
				taigamFlurry),
		},
	})
}

func taigamFlurry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell := item.Trigger.Event.CardID
	if err := (CopySpell{StackID: spell, Controller: item.Controller, FromLastKnown: true}).Apply(ctx); err != nil {
		return err
	}
	tail := ctx.Tail(game.GainSuspendExileTail)
	tail.N = 4
	return ExileTargetSpell{StackID: spell, ThenTail: tail}.Apply(ctx)
}
