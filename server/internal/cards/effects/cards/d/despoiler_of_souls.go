package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Despoiler of Souls — Creature — Horror {B}{B}, 3/1 (Magic Origins, S82 older
// sets):
//
//	"This creature can't block.
//	 {B}{B}, Exile two other creature cards from your graveyard: Return this
//	 card from your graveyard to the battlefield."
//
// Bloodsoaked Champion's static and graveyard return, with an exile cost for
// the {B}{B}: the activator names two creature cards in their own graveyard,
// neither of them the Despoiler (a card can't pay for itself), exiled as the
// ability is activated (CR 602.2b). It can't be activated with fewer than two
// other creature cards there. The return is checked again as the ability
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cabb0cf3-124a-4f5f-9ca7-2d45a72159ca",
		Name:         "Despoiler of Souls",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}, Exile two other creature cards from your graveyard: Return this card from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{B}{B}"),
				game.AbilityCost{ExileCards: effects.ExileCardsFromGraveyard(2, "two other creature cards", func(c game.Card) bool { return c.IsCreature() })}),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
