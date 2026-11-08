package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Edgewall Inn — Land (Wilds of Eldraine, #4471):
//
//	"This land enters tapped.
//	 As this land enters, choose a color.
//	 {T}: Add one mana of the chosen color.
//	 {3}, {T}, Sacrifice this land: Return target card that has an
//	 Adventure from your graveyard to your hand."
//
// The first three lines are Uncharted Haven's (chosen_color_lands.go): a
// tapped entry (SelfEntersTapped), an as-enters colour prompt
// (ChooseColorAsEnters) and a mana ability read off the stored colour.
// The last is Bonebind Orator's graveyard return
// (returnTargetGraveyardCardToHand) narrowed to a card that has an
// Adventure (HasAnAdventure, CR 715.1), behind {3}, {T} and sacrificing
// the Inn; the target is chosen before the cost is paid (CR 602.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4dfd33e8-7e30-493b-8564-f7df5f0257aa",
		Name:         "Edgewall Inn",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Edgewall Inn"),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this land: Return target card that has an Adventure from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("target card that has an Adventure from your graveyard", YouOwn(), HasAnAdventure()),
			Effect:  returnTargetGraveyardCardToHand,
		}},
	})
}
