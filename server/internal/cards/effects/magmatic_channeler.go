package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmatic Channeler — Creature — Human Wizard {1}{R}, 2/2 (Jumpstart 2022,
// #5948):
//
//	"As long as there are four or more instant and/or sorcery cards in your
//	 graveyard, this creature gets +3/+1.
//	 {T}, Discard a card: Exile the top two cards of your library, then
//	 choose one of them. You may play that card this turn."
//
// Ghitu Lavarunner's graveyard count (CR 604.2, layer 7c). The discard is a
// cost (CR 118.8); the exile is face up and only the chosen card may be
// played, this turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "382267da-c63c-432e-8a11-f803cf0e3a3a",
		Name:         "Magmatic Channeler",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "As long as there are four or more instant and/or sorcery cards in your graveyard, this creature gets +3/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: gapB157FourInstantOrSorceryCardsInYourGraveyard,
			Apply:     pumpBy(3, 1),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard a card: Exile the top two cards of your library, then choose one of them. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), DiscardACard()),
			Effect:  Do(Instruction(gapB157ExileTwoChooseOnePlay)),
		}},
	})
}

func gapB157FourInstantOrSorceryCardsInYourGraveyard(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID &&
		GraveyardAtLeast(4, MatchInstantOrSorcery)(g, source.Controller, source.InstanceID)
}
