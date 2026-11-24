package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lavinia, Azorius Renegade — Legendary Creature — Human Soldier {W}{U},
// 2/2 (Ravnica Remastered, #5948):
//
//	"Each opponent can't cast noncreature spells with mana value greater
//	 than the number of lands that player controls.
//	 Whenever an opponent casts a spell, if no mana was spent to cast it,
//	 counter that spell."
//
// The first ability is a cast restriction on each opponent, measured
// against the lands the CASTER controls, declared as a mana-value
// ceiling (game.CastRestriction.ManaValueCeiling, #6085): the gate
// compares the spell's mana value with its announced X (CR 202.3e), so a
// Fireball's X is held to what the caster's lands allow, and the bot and
// the client's X picker stop at that X. A creature spell, and any spell
// Lavinia's controller casts, is never refused. Abilities are not spells.
//
// The second ability is Boromir, Warden of the Tower's: an intervening if
// (CR 603.4) read from what was actually paid, checked as the spell is
// cast and again as the ability resolves. It declares ReadsManaSpent, so
// a waived payment asks the caster what they paid outside the pool.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "81ba3a33-0e04-4f68-9a42-1dd717733b57",
		Name:           "Lavinia, Azorius Renegade",
		ReadsManaSpent: game.ReadsOpponentsCasts,
		Completeness:   effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Each opponent can't cast noncreature spells with mana value greater than the number of lands that player controls.",
			ManaValueCeiling: func(q game.CastQuery) (int, bool) {
				if q.Source.Controller == q.Controller || effects.MatchCastCard(q.Game, effects.Creature(), q.Controller, q.Card) {
					return 0, false
				}
				return effects.B02CountLandsControlledBy(q.Game, q.Controller), true
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AllOf(effects.AnOpponentCast(nil), func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.NoManaWasSpentToCast(g, ev.CardID)
			}), "Lavinia, Azorius Renegade — counter that spell", effects.CounterTheSpellIfNoManaWasSpent),
		},
	})
}
