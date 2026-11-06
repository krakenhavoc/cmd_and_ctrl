package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prompto Argentum — Legendary Creature — Human Scout {1}{R}, 2/2 (Final
// Fantasy, #4462):
//
//	"Haste
//	 Selfie Shot — Whenever you cast a noncreature spell, if at least four
//	 mana was spent to cast it, create a Treasure token."
//
// Blazing Bomb's cast trigger: the amount spent on the spell is read off
// its payment as it is cast (mana of every kind counts). The intervening
// if (CR 603.4) is asked then; what was spent to cast a spell never
// changes, so asking again as the trigger resolves gives the same
// answer. The Treasure is made whether or not Prompto is still on the
// battlefield. "Selfie Shot" is a flavor word with no rules meaning.
//
// No simplification.
func init() {
	selfieShot := On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return YouCast(Noncreature())(ev, source, lki, g) && g.StackItemPaidForEffect(ev.CardID).Spent().Total() >= 4
	}, "Prompto Argentum — Selfie Shot: create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1}))
	Register(Spec{
		OracleID:        "e13d7600-0924-42eb-9d65-aa18548624a7",
		Name:            "Prompto Argentum",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{selfieShot},
	})
}
