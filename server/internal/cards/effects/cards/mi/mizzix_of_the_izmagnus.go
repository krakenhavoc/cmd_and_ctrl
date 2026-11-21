package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mizzix of the Izmagnus — Legendary Creature — Goblin Wizard {2}{U}{R},
// 2/2 (Commander 2015, #5546):
//
//	"Whenever you cast an instant or sorcery spell with mana value greater
//	 than the number of experience counters you have, you get an experience
//	 counter.
//	 Instant and sorcery spells you cast cost {1} less to cast for each
//	 experience counter you have."
//
// The trigger compares the spell's mana value (an announced X counts, CR
// 202.3e) with the caster's experience counters as the spell becomes cast,
// before the new counter: a 2-mana spell with one experience counter gives
// a second. The discount is Mindsplice Apparatus' CostsLessEach over instant
// and sorcery spells you cast, counting your experience counters (a player
// counter, CR 122.1) instead of counters on the source. It reduces generic
// mana only, and never changes the spell's mana value (CR 601.2f), so a
// discounted spell is still measured at its printed value.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f787c6cf-a4ff-487b-a7b0-0f516e68510a",
		Name:         "Mizzix of the Izmagnus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Or(effects.Instant(), effects.Sorcery()), c15ManaValueAboveYourExperience),
				"Mizzix of the Izmagnus — you get an experience counter", effects.YouGetAnExperienceCounter),
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(c15ExperienceOfTheModifiersController,
				"Instant and sorcery spells you cast cost {1} less to cast for each experience counter you have.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
	})
}

// c15ManaValueAboveYourExperience passes a spell whose mana value is greater
// than the number of experience counters the caster has.
func c15ManaValueAboveYourExperience(g *game.Game, caster uuid.UUID, c game.Card) bool {
	mv, ok := g.ManaValueForEffect(c)
	return ok && mv > effects.ExperienceCounters(g, caster)
}

// c15ExperienceOfTheModifiersController is the cost modifier's controller's
// experience counters: Mizzix's "for each experience counter you have".
func c15ExperienceOfTheModifiersController(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return effects.ExperienceCounters(q.Game, q.Source.Controller)
}
