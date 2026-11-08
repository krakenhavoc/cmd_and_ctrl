package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Lord Zuko — Legendary Creature — Human Noble Ally {R}{W}{B}, 2/4:
//
//	"Firebending X, where X is Fire Lord Zuko's power. (Whenever this
//	 creature attacks, add X {R}. This mana lasts until end of combat.)
//	 Whenever you cast a spell from exile and whenever a permanent you
//	 control enters from exile, put a +1/+1 counter on each creature you
//	 control."
//
// Firebending is Avatar Roku's red mana kept until end of combat
// (KeepManaUntilEndOfCombat), with X read from Zuko's power as the
// trigger resolves (CR 608.2h): as he is now while he is still the
// object that attacked, as he last existed once he has left.
//
// The second ability is one trigger with two events. A cast from exile
// is Ghost-Spider's read (EventCast stamps the origin zone), narrowed to
// spells; an entry from exile is Extraordinary Journey's
// (Event.EnteredFrom), under your control, Zuko himself included. A
// permanent spell cast from exile enters from the stack, so it triggers
// the ability once, for the cast. The counters go on each creature you
// control as it resolves.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "a129423d-c886-4120-96f7-c2517ea7eb83",
		Name:         "Fire Lord Zuko",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fire Lord Zuko — firebending X, where X is Fire Lord Zuko's power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := 0
					if info, ok := ctx.SourcePermanent(); ok {
						n = info.Power
					}
					// X of zero adds nothing but still firebends (#5276).
					return firebending(n).Apply(ctx)
				}),
			OnAny([]game.EventKind{game.EventCast, game.EventETB}, fireLordZukoFromExile,
				"Fire Lord Zuko — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
	})
}

// fireLordZukoFromExile is "you cast a spell from exile" or "a permanent
// you control enters from exile".
func fireLordZukoFromExile(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventCast:
		return ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
	case game.EventETB:
		if ev.EnteredFrom != game.ZoneExile {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.Controller == source.Controller
	}
	return false
}
