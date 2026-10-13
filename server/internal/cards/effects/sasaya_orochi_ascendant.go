package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sasaya, Orochi Ascendant // Sasaya's Essence — {1}{G}{G} Legendary
// Creature — Snake Monk 2/3 // Legendary Enchantment, a flip card
// (#3250, ADR 0151):
//
//	Top:     "Reveal your hand: If you have seven or more land cards in
//	          your hand, flip Sasaya."
//	Flipped: "Whenever a land you control is tapped for mana, add an
//	          additional one mana of any type that land produced for
//	          each other land you control with the same name as it."
//
// The cost is RevealYourHand (game.AbilityCost.RevealHand): the whole
// hand is shown to the table as the ability is activated, before
// anyone may respond, and an empty hand can still pay it. The land
// count is taken as the ability resolves (an "if" in the effect, not an
// intervening "if" on the cost), so cards cast or played in response
// count. A Sasaya that left and came back is a new object and flips
// nothing (#1432).
//
// The Essence is a triggered mana ability (CR 605.1b): one extra mana
// per OTHER land of the tapped land's name that its controller
// controls, each an independent pick among the types the tapped land
// produced. A land with no namesake adds nothing. It is green and
// {1}{G}{G} (CR 710.1c).
//
// No simplification.
const sasayaOracleID = "f07f3bf0-e1c8-4a24-87c7-1d6d0b5bd242"

func init() {
	Register(Spec{
		OracleID:     sasayaOracleID,
		Name:         "Sasaya, Orochi Ascendant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Reveal your hand: If you have seven or more land cards in your hand, flip Sasaya.",
			Cost:    RevealYourHand(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: flipThisIf(func(g *game.Game, controller uuid.UUID) bool {
				p := g.PlayerByIDForEffect(controller)
				if p == nil || p.Hand == nil {
					return false
				}
				lands := 0
				for i := range p.Hand.Cards {
					if p.Hand.Cards[i].IsLand() {
						lands++
					}
				}
				return lands >= 7
			}),
		}},
	})
	Register(Spec{
		OracleID:     sasayaOracleID + "#1",
		Name:         "Sasaya's Essence",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			WheneverYouTapALandForMana(
				"Sasaya's Essence — add one mana of any type that land produced for each other land with the same name",
				func(prod game.ManaProduced, _ *game.Card, g *game.Game) string {
					pipe := prod.ProducedColorPipe()
					if pipe == "" {
						return ""
					}
					others := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == prod.Controller && c.IsLand() &&
							c.InstanceID != prod.SourceID() && c.Name == prod.Source.Name {
							others++
						}
					}
					return strings.Repeat(pipe, others)
				}),
		},
	})
}
