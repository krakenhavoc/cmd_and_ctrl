package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jeskai Devotee — Creature — Orc Monk {1}{R}, 2/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"Flurry — Whenever you cast your second spell each turn, this
//	 creature gets +1/+1 until end of turn.
//	 {1}: Add {U}, {R}, or {W}. Activate only once each turn."
//
// Flurry is an ability word (CR 207.2c): the trigger is Poised
// Practitioner's YouCastYourSecondSpellEachTurn, and the pump
// thisCreatureUntilEOT, which does nothing if the Devotee has left. The
// mana ability is Mardu Devotee's: a filter paid with {1} and no tap,
// limited by the declarative OncePerTurn bit, counted per object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6dcdd313-6df7-4d44-a4cb-0346a4fdb4e1",
		Name:         "Jeskai Devotee",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Jeskai Devotee — flurry: this creature gets +1/+1 until end of turn",
				thisCreatureUntilEOT("Jeskai Devotee — +1/+1", 1, 1)),
		},
		ManaAbilities: []ManaAbility{{
			Cost:        ManaAbilityCost{Mana: "{1}"},
			Produced:    "{U|R|W}",
			Label:       "{1}: Add {U}, {R}, or {W}",
			OncePerTurn: true,
		}},
	})
}
