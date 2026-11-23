package cardsth

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Grey Havens — Legendary Land (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"When The Grey Havens enters, scry 1.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color among legendary creature cards in your
//	 graveyard."
//
// The enters trigger is the ordinary scry (a fixed-amount rider, so the card
// declares it). The second mana ability is Mox Amber's derived mana turned to
// the graveyard: the colors of the legendary creature cards in your graveyard,
// read through the layers at the moment the ability is activated; with none, or
// only colorless ones, it adds no mana and still taps (CR 106.5, CR 105.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1a9695e-073b-4a65-b3ec-2cfddc23202a",
		Name:         "The Grey Havens",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("The Grey Havens — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				ProducedFunc: producedFromLegendaryCreatureCardsInYourGraveyard,
				Label:        "Add one mana of any color among legendary creature cards in your graveyard",
			},
		},
	})
}

// producedFromLegendaryCreatureCardsInYourGraveyard is the colors of the
// legendary creature cards in the controller's graveyard, as one pipe-joined
// choice.
func producedFromLegendaryCreatureCardsInYourGraveyard(g *game.Game, controller, _ uuid.UUID) string {
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Graveyard == nil {
		return ""
	}
	seen := map[string]bool{}
	for _, c := range p.Graveyard.Cards {
		ch := c.Effective()
		if !effects.HasFold(ch.Supertypes, "Legendary") || !effects.HasFold(ch.Types, "Creature") {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[strings.ToUpper(col)] = true
		}
	}
	return effects.PipeString(effects.ColorsOnly(effects.OrderedManaSymbols(seen)))
}
