package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraul Swarm — Creature — Insect Warrior {4}{B}, 4/1 (Jumpstart 2022,
// #5948):
//
//	"Flying
//	 {2}{B}, Discard a creature card: Return this card from your graveyard
//	 to your hand."
//
// Blessed Ghoul's graveyard row with Vampire Hounds' "Discard a creature
// card" cost: paid as the ability is activated (CR 602.2b), and only a
// creature card in hand can pay it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac97d8fc-15fa-4a60-96d3-4c09deafa83a",
		Name:            "Kraul Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Discard a creature card: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature)),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
