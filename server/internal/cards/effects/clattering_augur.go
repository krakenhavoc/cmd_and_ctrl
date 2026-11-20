package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clattering Augur — Creature — Skeleton Shaman {1}{B}, 1/1 (Modern
// Horizons 2, #5501):
//
//	"This creature can't block.
//	 When this creature enters, you draw a card and you lose 1 life.
//	 {2}{B}{B}: Return this card from your graveyard to your hand."
//
// Craven Giant's restriction, Phyrexian Rager's enter trigger
// (drawACardAndLoseALife) and Abzan Devotee's graveyard ability, which
// works only while the card is in its owner's graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d88f249f-5bf1-448a-91d1-f81538dfd475",
		Name:         "Clattering Augur",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Clattering Augur — draw a card, lose 1 life", drawACardAndLoseALife),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{B}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
