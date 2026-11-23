package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Staff of Eden, Vault's Key — Legendary Artifact {6} (Assassin's Creed,
// #5948):
//
//	"When Staff of Eden enters, put target legendary permanent card not named
//	 Staff of Eden, Vault's Key from a graveyard onto the battlefield under
//	 your control.
//	 {T}: Draw a card for each permanent you control but don't own."
//
// The entry trigger targets a legendary permanent card (an artifact, creature,
// enchantment, land, planeswalker or battle card, CR 110.4) in any graveyard
// and puts it onto the battlefield under the Staff's controller, whoever owned
// it; a card that left the graveyard in response gets nothing (CR 608.2b). The
// {T} ability counts, as it resolves, the permanents its controller controls
// and does not own (a stolen creature, a reanimated card from another
// graveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4fe63e7-bb4c-4aff-b74f-d385d0128def",
		Name:         "Staff of Eden, Vault's Key",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Staff of Eden — put target legendary permanent card from a graveyard onto the battlefield under your control", effects.ReanimateChosenUnderYourControl("Staff of Eden")),
				effects.TargetCardInGraveyard("target legendary permanent card not named Staff of Eden, Vault's Key from a graveyard",
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsLegendary() && c.IsPermanent() && c.Name != "Staff of Eden, Vault's Key"
					}),
			),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card for each permanent you control but don't own.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  staffOfEdenDraw,
		}},
	})
}

func staffOfEdenDraw(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.Owner != item.Controller {
			n++
		}
	}
	return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
}
