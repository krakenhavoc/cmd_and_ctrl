package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel's Grace — Instant {W}:
//
//	"Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 You can't lose the game this turn and your opponents can't win the
//	 game this turn. Until end of turn, damage that would reduce your
//	 life total to less than 1 reduces it to 1 instead."
//
// The GRANTED half of ADR 0057 Decision 4 (#749), and the one card in
// the first wave that uses it — the owner's answer to question 3 was
// to ship it declared incomplete rather than ship the registry with no
// real card behind it. The two gates are PlayerStatics on the caster
// with an until-end-of-turn Duration (the 2026-09-24 amendment's home
// for them), so they outlive the spell, which is in a graveyard a
// moment after it resolves, and end in the cleanup step (CR 514.2).
// "Your opponents" is relative to the caster and is stored once.
//
// TWO CLAUSES ARE MISSING, and both make the card WEAKER than printed:
//
//   - The life floor ("damage that would reduce your life total to
//     less than 1 reduces it to 1 instead") is a damage-RESULT
//     replacement that has to keep lifelink whole (the 2021-03-19
//     ruling). That is a separate seam — ADR 0057 Decision 8 scoped it
//     out. Without it the caster still can't LOSE this turn; they just
//     end the turn at 0 or less life and lose at the first check after
//     it unless they gain life first.
//   - Split second (CR 702.61) is not read from the card anywhere in
//     the engine yet — the stack's split-second gate is set only by a
//     sandbox cast flag — so players may respond to Angel's Grace.
//     #1519 is that seam; it clears this caveat.
func init() {
	Register(Spec{
		OracleID:     "66ca8a60-e028-4a5f-8177-860b888cb9d1",
		Name:         "Angel's Grace",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Damage can still reduce your life total below 1 this turn; you just can't lose the game until the turn ends.",
			"Split second isn't enforced — players can respond to it.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CantLoseAndOpponentsCantWinThisTurn{Label: "Angel's Grace"}.Apply(ctx)
		},
	})
}

// CantLoseAndOpponentsCantWinThisTurn is "You can't lose the game this
// turn and your opponents can't win the game this turn" — Angel's
// Grace's gates, granted to the resolving item's controller until end
// of turn (ADR 0057 Decision 4, the GRANTED source).
type CantLoseAndOpponentsCantWinThisTurn struct {
	// Label names the grant on the seat's badge tooltip and the model
	// prompt — the card's name.
	Label string
}

// Apply implements Primitive.
func (c CantLoseAndOpponentsCantWinThisTurn) Apply(ctx *Context) error {
	me := ctx.Controller()
	d := ctx.Game.UntilEndOfTurnDuration()
	for _, gate := range YouCantLoseOpponentsCantWin() {
		ctx.Game.GrantGameEndGateForEffect(me, gate, c.Label, ctx.Source(), d)
	}
	return nil
}
