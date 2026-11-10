package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pit of Offerings — Land — Cave (The Lost Caverns of Ixalan, #4470):
//
//	"This land enters tapped.
//	 When this land enters, exile up to three target cards from graveyards.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any of the exiled cards' colors."
//
// The trigger exiles the still-legal targets (CR 608.2b) in one move, and
// the second mana ability reads the cards THIS land exiled (Chrome Mox's
// record, keyed on the trigger's label and the land's own incarnation,
// CR 400.7), through their colours as they are now in exile: a colourless
// card or a land adds none, and with no coloured card exiled the ability
// produces nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "044d2788-6daa-4849-a813-1f577eef9295",
		Name:         "Pit of Offerings",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(pitOfOfferingsLabel, ExileTargetCards),
				TargetCardInGraveyard("up to three target cards from graveyards").WithCount(0, 3)),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "{T}: Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: pitOfOfferingsColors,
				Label:        "{T}: Add one mana of any of the exiled cards' colors",
			},
		},
	})
}

// pitOfOfferingsLabel is the trigger's stack label; the "exiled with this
// land" record keys on it.
const pitOfOfferingsLabel = "Pit of Offerings — exile up to three target cards from graveyards"

// pitOfOfferingsColors is every colour among the cards this land exiled.
func pitOfOfferingsColors(g *game.Game, _, source uuid.UUID) string {
	seen := map[string]bool{}
	for _, id := range b27ExiledWith(g, source, pitOfOfferingsLabel) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[strings.ToUpper(col)] = true
		}
	}
	return pipeString(colorsOnly(orderedManaSymbols(seen)))
}
