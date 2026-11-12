package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Chorus — Enchantment {3}{W}{W} (Urza's Saga):
//
//	"Whenever a creature you control enters, you gain life equal to its
//	 toughness."
//
// Verdant Sun's Avatar's trigger on an enchantment: once per creature
// that enters under your control, a token included. The toughness is
// read as the trigger resolves (CR 608.2h), the creature as it is then,
// or as it last existed on the battlefield if it has left; a toughness of
// zero or less gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e66b12ff-45da-46dc-9148-bd8ea9b7fde0",
		Name:         "Angelic Chorus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Angelic Chorus — gain life equal to its toughness",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			},
			Effect: Do(GainLifeEqualToSubject{Stat: SubjectToughness}),
		}},
	})
}
