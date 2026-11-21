package cardsp

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "044d2788-6daa-4849-a813-1f577eef9295",
		Name:         "Pit of Offerings",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(pitOfOfferingsLabel, effects.ExileTargetCards),
				effects.TargetCardInGraveyard("up to three target cards from graveyards").WithCount(0, 3)),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "{T}: Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
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
	for _, id := range effects.B27ExiledWith(g, source, pitOfOfferingsLabel) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[strings.ToUpper(col)] = true
		}
	}
	return effects.PipeString(effects.ColorsOnly(effects.OrderedManaSymbols(seen)))
}
