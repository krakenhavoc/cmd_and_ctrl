package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Buried Ruin — Land (EDHREC rank 194):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Return target artifact card from
//	 your graveyard to your hand."
//
// A colourless land that is also a one-shot Regrowth for artifacts
// — played in every deck whose commander is an artifact or whose
// win condition is one.
//
// The recursion is an ordinary CR 602 activated ability with three
// cost components (Myriad Landscape's shape) and a graveyard target
// clause, the same TargetCardInGraveyard Eternal Witness and Sun
// Titan use: the picker opens the zone browser on your graveyard
// and offers only artifact cards. The cost is paid at announce, so
// the Ruin is gone before the ability resolves and a response that
// exiles the targeted card fizzles the ability with the land already
// spent — as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3644f316-f9a3-46c9-9b1e-747f86cf4ead",
		Name:         "Buried Ruin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this land: Return target artifact card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{
					Target: item.Targets[0].ID,
					Dest:   game.ZoneHand,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
