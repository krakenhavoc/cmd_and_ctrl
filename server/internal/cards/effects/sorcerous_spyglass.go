package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorcerous Spyglass — Artifact {2} (Foundations):
//
//	"As this artifact enters, look at an opponent's hand, then choose any
//	 card name.
//	 Activated abilities of sources with the chosen name can't be activated
//	 unless they're mana abilities."
//
// Pithing Needle with a look first. The as-enters clause (CR 614.12) is two
// questions in a row, as Arachne, Psionic Weaver's: the opponent whose hand
// is looked at (only the Spyglass's controller learns the cards, Peek's
// "look at", not a reveal), then, from the continuation, the card name. With
// no opponent left to look at, the name is still chosen. The restriction is
// Pithing Needle's: it reads the name off this permanent each time an
// ability is announced, and an unanswered name stops nothing.
func init() {
	Register(Spec{
		OracleID:     "b2187f45-80ae-4ac4-9f83-5eb7a00978e2",
		Name:         "Sorcerous Spyglass",
		Completeness: CompletenessFull,
		AsEnters:     sorcerousSpyglassAsEnters,
		ActivationRestrictions: []game.ActivationRestriction{
			ChosenNameCantActivate(
				"Sorcerous Spyglass — activated abilities of sources with the chosen name can't be activated unless they're mana abilities.",
				true),
		},
	})
}

const sorcerousSpyglassName = "Sorcerous Spyglass"

// sorcerousSpyglassAsEnters asks for the opponent whose hand is looked at.
func sorcerousSpyglassAsEnters(card *game.Card, ctx *Context) error {
	if ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  card.Controller,
		Among:    Opponents.seats(ctx, card.Controller),
		Source:   card.InstanceID,
		Question: sorcerousSpyglassName + " — look at an opponent's hand",
		ThenKey:  sorcerousSpyglassLooked,
		Carry:    []uuid.UUID{card.InstanceID},
	}) != uuid.Nil {
		return nil
	}
	ctx.Game.QueueCardNameChoiceForEffect(card.Controller, card.InstanceID, sorcerousSpyglassName+" — choose a card name")
	return nil
}

// sorcerousSpyglassLooked is the opponent's answer: the controller looks at
// that hand, then names a card. An on-disk key: never renamed, never reused.
var sorcerousSpyglassLooked = game.RegisterOptionPickThen("option-pick/sorcerous-spyglass-look", func(g *game.Game, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if r.Option != nil {
		if p := g.PlayerByIDForEffect(r.Option.Player); p != nil && p.Hand != nil {
			for i := range p.Hand.Cards {
				p.Hand.Cards[i].AddKnower(r.Chooser)
			}
		}
	}
	g.QueueCardNameChoiceForEffect(r.Chooser, r.Carry[0], sorcerousSpyglassName+" — choose a card name")
	return nil
})
