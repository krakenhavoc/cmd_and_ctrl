package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rose Noble — Legendary Creature — Human {3}{U}, 2/3:
//
//	"Ward {2}
//	 Whenever you cast a Doctor spell or creature spell with doctor's
//	 companion, draw a card.
//	 Doctor's companion (You can have two commanders if the other is the
//	 Doctor.)"
//
// Ward {2} is the usual triggered ability. The draw is one cast trigger over
// two kinds of spell: a Doctor (the subtype, read through the layers off the
// spell as cast) and a creature spell that prints doctor's companion. That
// keyword is a deck-construction rule with no rules text of its own, so it is
// the printed-list predicate GapHasDoctorsCompanion reads (the companions
// already on the catalog), widened here by Rose Noble herself: a second Rose
// Noble cast is a creature spell with doctor's companion. A Doctor that is also
// a creature spell with doctor's companion draws once, it is one spell. The
// trigger is yours only, and it triggers as the spell is cast, so it draws even
// if the spell is later countered. Doctor's companion changes only how a
// commander deck is built.
//
// No simplification.
func init() {
	const oracle = "cf66067b-8f5e-46f6-9d61-cffb8875d858"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Rose Noble",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Rose Noble — ward {2}"),
			effects.On(game.EventCast, effects.YouCast(func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return effects.GapIsDoctorCard(c) ||
					(c.IsCreature() && (effects.GapHasDoctorsCompanion(c) || c.OracleID == oracle))
			}), "Rose Noble — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
