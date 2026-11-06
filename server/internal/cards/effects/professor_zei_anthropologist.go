package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "59a10335-72df-44a5-a514-5bac392cd282",
		Name:         "Professor Zei, Anthropologist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Discard a card: Draw a card.",
				Cost:    Plus(TapCost(), DiscardACard()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(DrawCards{N: 1}),
			},
			{
				Label:     "{1}, {T}, Sacrifice Professor Zei: Return target instant or sorcery card from your graveyard to your hand. Activate only during your turn.",
				Cost:      Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
				Condition: DuringYourTurn(),
				Targets:   TargetCardInGraveyard("target instant or sorcery card in your graveyard", Or(Instant(), Sorcery()), YouOwn()),
				Effect:    returnFirstLegalGraveyardTargetToHand,
			},
		},
	})
}
