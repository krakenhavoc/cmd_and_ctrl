package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Professor Zei, Anthropologist — Legendary Creature — Human Advisor Ally
// {U/R}{U/R}, 0/3 (Avatar: The Last Airbender, #4459):
//
//	"{T}, Discard a card: Draw a card.
//	 {1}, {T}, Sacrifice Professor Zei: Return target instant or sorcery
//	 card from your graveyard to your hand. Activate only during your
//	 turn."
//
// Two ordinary activated abilities. The rummage is Charging
// Strifeknight's; the second is Archaeomancer's return paid for with
// Zei himself, gated on DuringYourTurn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59a10335-72df-44a5-a514-5bac392cd282",
		Name:         "Professor Zei, Anthropologist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Discard a card: Draw a card.",
				Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:     "{1}, {T}, Sacrifice Professor Zei: Return target instant or sorcery card from your graveyard to your hand. Activate only during your turn.",
				Cost:      effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
				Condition: effects.DuringYourTurn(),
				Targets:   effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
				Effect:    effects.ReturnFirstLegalGraveyardTargetToHand,
			},
		},
	})
}
