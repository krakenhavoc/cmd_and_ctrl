package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dread Wanderer — {B} Creature — Zombie Jackal 2/1 (Amonkhet, S82 old batch
// 248):
//
//	"This creature enters tapped.
//	 {2}{B}: Return this card from your graveyard to the battlefield. Activate
//	 only as a sorcery and only if you have one or fewer cards in hand."
//
// The graveyard ability is checked at activation (CR 602.5b) for timing and
// for the hand's size, and the card must still be in the graveyard as it
// resolves (CR 608.2b). It enters tapped again, from its own replacement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88154d64-221f-4742-bccb-b1b61a2e8bf0",
		Name:         "Dread Wanderer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}: Return this card from your graveyard to the battlefield. Activate only as a sorcery and only if you have one or fewer cards in hand.",
			Cost:         effects.ManaCost("{2}{B}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.YouHaveAtMostCardsInHand(1)(g, controller)
			},
			Effect: effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
