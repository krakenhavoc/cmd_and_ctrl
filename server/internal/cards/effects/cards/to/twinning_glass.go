package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twinning Glass — Artifact {4} (Lorwyn, S82 OLD b547, #5593):
//
//	"{1}, {T}: You may cast a spell from your hand without paying its mana
//	 cost if it has the same name as a spell that was cast this turn."
//
// As the ability resolves the controller is offered every spell in their hand
// that shares a name with a spell cast this turn, by any player (the names the
// spells had on the stack, a split card's either name: Game.TurnTally.SpellNames),
// and may cast one without paying its mana cost, at the timing of the effect
// and not by the usual timing rules (CR 608.2g). A land card is not a spell. With no
// such card in the hand nothing is asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "907baed5-499b-4f20-bbb9-8f5bead37c30",
		Name:         "Twinning Glass",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: You may cast a spell from your hand without paying its mana cost if it has the same name as a spell that was cast this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				effects.OfferFreeHandCastOver(g, item.Controller, item.SourceCardID, twinningGlassCast,
					"Twinning Glass — cast a spell with the same name as a spell cast this turn from your hand without paying its mana cost?",
					effects.AnySpell, effects.S82OldB547HandCardsNamedLikeASpellCastThisTurn(g, item.Controller))
				return nil
			},
		}},
	})
}

var twinningGlassCast = effects.RegisterFreeHandCastOffer("choose-cards/twinning-glass-cast", "Twinning Glass")
