package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Edgewall Inn — Land (Wilds of Eldraine, #4471):
//
//	"This land enters tapped.
//	 As this land enters, choose a color.
//	 {T}: Add one mana of the chosen color.
//	 {3}, {T}, Sacrifice this land: Return target card that has an
//	 Adventure from your graveyard to your hand."
//
// The first three lines are Uncharted Haven's tapped entry, colour prompt
// and chosen-colour mana. The last is a graveyard return over any card
// you own that has an Adventure (an adventurer card, whichever half it
// is in the graveyard), re-checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4dfd33e8-7e30-493b-8564-f7df5f0257aa",
		Name:         "Edgewall Inn",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Edgewall Inn"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: effects.ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this land: Return target card that has an Adventure from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target card that has an Adventure in your graveyard", effects.YouOwn(), effects.HasAnAdventure()),
			Effect:  effects.ReturnFirstLegalGraveyardTargetToHand,
		}},
	})
}
