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
// Split second (CR 702.61) is declared below and enforced since #1519:
// nobody can cast a spell or activate a non-mana ability in response,
// which is what makes this a reliable answer to a lethal Exsanguinate.
// It shipped with #749 carrying a second caveat saying players could
// respond; #1519 cleared it.
//
// The life floor (#3172) is a third PlayerStatic payload
// (PlayerStatic.LifeFloorOne, game/life_floor.go) granted until end of
// turn with the two gates. It caps the life the damage COSTS, not the
// damage: the 2021-03-19 ruling keeps lifelink whole, and the part that
// cost no life rides the damage event as DamageNotLifeLoss.
func init() {
	Register(Spec{
		OracleID:        "66ca8a60-e028-4a5f-8177-860b888cb9d1",
		Name:            "Angel's Grace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CantLoseAndOpponentsCantWinThisTurn{Label: "Angel's Grace"}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.GrantLifeFloorForEffect(ctx.Controller(), "Angel's Grace: damage can't reduce your life total below 1", ctx.Source(), ctx.Game.UntilEndOfTurnDuration())
			return nil
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
