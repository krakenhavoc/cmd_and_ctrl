package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sharp-Eyed Rookie — Creature — Human Detective {1}{G}, 2/2 (Murders at
// Karlov Manor, #4469):
//
//	"Vigilance
//	 Whenever a creature you control enters, if its power is greater than
//	 this creature's power or its toughness is greater than this
//	 creature's toughness, put a +1/+1 counter on this creature and
//	 investigate."
//
// An intervening "if" (CR 603.4), so it is read twice. As the creature
// enters, its power and toughness (counters it enters with included) are
// read from the live card against the Rookie's. As the ability resolves,
// the entered creature is read as it is now, or as it last existed on the
// battlefield once it has left (Context.TriggeringPermanent), against the
// Rookie's present or last-known values (Context.SourcePermanent). The
// stat that is greater may have changed from power to toughness between
// the two reads and the ability still resolves. The Rookie's own entry
// never triggers it: nothing is greater than itself.
//
// No simplification.
func init() {
	const label = "Sharp-Eyed Rookie — +1/+1 counter and investigate"
	Register(Spec{
		OracleID:        "0642fb0d-73cb-4989-852c-3233fe36c3b4",
		Name:            "Sharp-Eyed Rookie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(condEnteredCreatureOutstatsSource{}, On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsCreature() &&
						(c.PowerForComparison() > source.PowerForComparison() || c.CurrentToughness() > source.CurrentToughness())
				}, label,
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, CreateToken{Template: ClueToken(), N: 1}))),
		},
	})
}

// condEnteredCreatureOutstatsSource is "its power is greater than this
// creature's power or its toughness is greater than this creature's
// toughness", for the creature a trigger was about against the trigger's
// own source, both last-known once they have left (CR 608.2h).
type condEnteredCreatureOutstatsSource struct{}

func (condEnteredCreatureOutstatsSource) Holds(ctx *Context) bool {
	entered, ok := ctx.TriggeringPermanent()
	if !ok {
		return false
	}
	src, ok := ctx.SourcePermanent()
	if !ok {
		return false
	}
	return entered.Power > src.Power || entered.Toughness > src.Toughness
}

func (condEnteredCreatureOutstatsSource) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
