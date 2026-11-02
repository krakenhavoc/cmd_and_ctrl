package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhino, Barreling Brute — Legendary Creature — Human Villain
// {3}{R}{R}{G}{G}, 6/7 (Marvel's Spider-Man, #4460):
//
//	"Vigilance, trample, haste
//	 Whenever Rhino attacks, if you've cast a spell with mana value 4 or
//	 greater this turn, draw a card."
//
// An intervening "if" (CR 603.4): the trigger-time check is the On
// predicate and the same condition is rechecked as the ability resolves.
// It reads the cast tally's highest mana value (CastTally.MaxManaValue),
// counted as each spell became cast with its X included (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9a6bf8a3-7640-4890-ac62-d5028f41978b",
		Name:            "Rhino, Barreling Brute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(condCastSpellWithManaValueAtLeast{N: 4},
				On(game.EventAttack, AllOf(ThisAttacked, castSpellWithManaValueAtLeast(4)),
					"Rhino — you've cast a spell with mana value 4 or greater this turn: draw a card",
					Do(DrawCards{N: 1}))),
		},
	})
}

// castSpellWithManaValueAtLeast is the trigger-time half of "if you've
// cast a spell with mana value N or greater this turn", read for the
// source's controller.
func castSpellWithManaValueAtLeast(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return g.CastTallyFor(source.Controller).MaxManaValue >= n
	}
}

// condCastSpellWithManaValueAtLeast is the resolution-time half.
type condCastSpellWithManaValueAtLeast struct{ N int }

func (c condCastSpellWithManaValueAtLeast) Holds(ctx *Context) bool {
	return ctx.Game.CastTallyFor(ctx.Controller()).MaxManaValue >= c.N
}

func (condCastSpellWithManaValueAtLeast) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
