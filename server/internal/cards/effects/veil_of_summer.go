package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veil of Summer — Instant {G}:
//
//	"Draw a card if an opponent has cast a blue or black spell this
//	 turn. Spells you control can't be countered this turn. You and
//	 permanents you control gain hexproof from blue and from black until
//	 end of turn. (You and they can't be the targets of blue or black
//	 spells or abilities your opponents control.)"
//
// Three sentences, in printed order:
//
//   - THE DRAW reads the turn's casts (anOpponentCastASpellThisTurn):
//     any EventCast this turn by an opponent whose spell was blue or
//     black. The colour is read off the card where it is now, the
//     catalog's existing reading of a past cast (b25CastIsMulticolored);
//     a copy is not cast (CR 707.10) and never counts.
//   - "SPELLS YOU CONTROL CAN'T BE COUNTERED THIS TURN" is a "this turn"
//     grant (ADR 0106 §4 decision 2, #1806), read at the counter gate,
//     so it covers spells already on the stack and spells cast later
//     this turn (CR 611.2c), and ends at cleanup (CR 514.2). "Control"
//     is the spell's current controller.
//   - THE HEXPROOF is two grants of "hexproof from [colour]" (CR 702.11e,
//     #3172, ADR 0038's amendment of 2026-10-13), one for blue and one
//     for black, each given twice over: to every permanent you control
//     as the spell resolves (GrantKeywordUntilEOT, CR 611.2c — a
//     permanent that arrives later this turn is not covered) and to you
//     (GainPlayerKeyword, until end of turn). An opponent's blue or
//     black spell or ability can no longer target either; a red or
//     colourless one still can, and your own spells always can.
func init() {
	Register(Spec{
		OracleID:     "002965be-a36f-4a09-9ce0-c6535bca1703",
		Name:         "Veil of Summer",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			blue, black := ctx.TextColor("U"), ctx.TextColor("B")
			if anOpponentCastASpellThisTurn(ctx, func(c game.Card) bool { return c.HasColor(blue) || c.HasColor(black) }) {
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := (GrantCounterShield{From: "Veil of Summer", Grant: SpellsCantBeCounteredThisTurn(
				"Spells you control can't be countered this turn.",
				game.CounterShieldYouControl, game.PermissionFilter{})}).Apply(ctx); err != nil {
				return err
			}
			hexproofFrom := []string{game.HexproofFromColor(blue), game.HexproofFromColor(black)}
			if err := (GrantKeywordUntilEOT{
				Match:    YouControl(),
				Keywords: hexproofFrom,
				Label:    "Veil of Summer — hexproof from blue and from black",
			}).Apply(ctx); err != nil {
				return err
			}
			for _, kw := range hexproofFrom {
				if err := (GainPlayerKeyword{
					Player:  ctx.Controller(),
					Keyword: kw,
					Label:   "Veil of Summer — " + kw,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// anOpponentCastASpellThisTurn reports whether an opponent of the
// context's controller has cast a spell this turn that `match` admits —
// Veil of Summer's "if an opponent has cast a blue or black spell this
// turn". A filtered scan over g.EventsThisTurn(), the bounded slice
// that starts at the real turn boundary, with each spell's card looked
// up where it sits now (b25CastIsMulticolored's reading): a spell
// whose card has since left every tracked zone is not counted, which
// is the weaker answer. Copies emit no EventCast (CR 707.10).
func anOpponentCastASpellThisTurn(ctx *Context, match func(game.Card) bool) bool {
	opponents := map[uuid.UUID]bool{}
	for _, id := range ctx.Opponents() {
		opponents[id] = true
	}
	for _, ev := range ctx.Game.EventsThisTurn() {
		if ev.Kind != game.EventCast || !opponents[ev.Actor] {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(ev.CardID); ok && match(c) {
			return true
		}
	}
	return false
}
