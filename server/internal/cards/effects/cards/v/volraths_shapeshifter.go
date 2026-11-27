package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volrath's Shapeshifter — Creature — Phyrexian Shapeshifter {1}{U}{U}, 0/1
// (Stronghold, S82 OLD batch 746, #5627):
//
//	"As long as the top card of your graveyard is a creature card, this
//	 creature has the full text of that card and has the text "{2}: Discard a
//	 card." (This creature has that card's name, mana cost, color, types,
//	 abilities, power, and toughness.)
//	 {2}: Discard a card."
//
// The full-text seam (#6233, CR 612.6): the static is declared as data
// (FullTextOfTopGraveyardCreature) and the engine writes the top creature
// card's values onto the permanent before every layer pass, keeping the
// "{2}: Discard a card." grant named by Keep. The ability is the printed one
// too, so it works with no creature card on top. The discard is the
// activator's own choice (a real prompt over their hand), made as it resolves.
//
// No simplification.
func init() {
	const keep = "volraths-shapeshifter/discard"
	discard := effects.ActivatedAbility{
		Label:   "{2}: Discard a card.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    effects.ManaCost("{2}"),
		Effect: func(g *game.Game, item *game.StackItem) error {
			g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Controller, Source: item.SourceCardID, N: 1})
			return nil
		},
	}
	effects.Register(effects.Spec{
		OracleID:                       "16e9a4ee-0398-4e76-bc7e-dd2d52056aaf",
		Name:                           "Volrath's Shapeshifter",
		Completeness:                   effects.CompletenessFull,
		FullTextOfTopGraveyardCreature: &game.FullTextStatic{Keep: keep},
		Activated:                      []effects.ActivatedAbility{discard},
		Grants:                         []effects.AbilityGrant{{Key: keep, Text: "{2}: Discard a card.", Activated: []effects.ActivatedAbility{discard}}},
	})
}
