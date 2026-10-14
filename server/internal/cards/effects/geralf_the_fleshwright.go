package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geralf, the Fleshwright — Legendary Creature — Human Warlock {2}{U},
// 2/3:
//
//	"Whenever you cast a spell during your turn other than your first
//	 spell that turn, create a 2/2 blue and black Zombie Rogue creature
//	 token.
//	 Whenever a Zombie you control enters, put a +1/+1 counter on it for
//	 each other Zombie that entered the battlefield under your control
//	 this turn."
//
// "Other than your first spell" reads the turn's cast tally, which
// already counts the spell that triggered: two or more means this is
// not the first, whether or not Geralf was there for the first.
//
// The counters are counted as the trigger resolves (CR 608.2h) from the
// turn's tally of entries by subtype (EnteredWithSubtypeThisTurn, the
// Zombie as it entered), less the one entering. A Zombie that entered
// earlier and has since died still counts, as the rulings say.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a62e6cf4-1adb-4c28-b24b-91287bb23e96",
		Name:         "Geralf, the Fleshwright",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, youCastANonFirstSpellOnYourTurn,
				"Geralf, the Fleshwright — create a 2/2 Zombie Rogue",
				Do(CreateToken{Template: TokenCard("2/2 blue and black Zombie Rogue"), N: 1})),
			On(game.EventETB, aZombieEnteredUnderYourControl,
				"Geralf, the Fleshwright — put a +1/+1 counter on it for each other Zombie that entered this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					n := g.EnteredWithSubtypeThisTurn(item.Controller, "Zombie") - 1
					if n <= 0 {
						return nil
					}
					return AddCounter{Target: ctx.Trigger().Object.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
	})
}

// youCastANonFirstSpellOnYourTurn is "whenever you cast a spell during
// your turn other than your first spell that turn".
func youCastANonFirstSpellOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return false
	}
	active := g.Seats[g.Turn.ActiveSeat]
	return active != nil && active.ID == source.Controller && g.CastTallyFor(source.Controller).Total >= 2
}

// aZombieEnteredUnderYourControl is "whenever a Zombie you control
// enters", the source included when it is one.
func aZombieEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtype("Zombie")
}
