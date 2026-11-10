package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Llanowar — Plane — Dominaria (#4870):
//
//	"All creatures have "{T}: Add {G}{G}."
//	 Whenever chaos ensues, untap all creatures you control."
//
// The static grants the mana ability to every creature on the battlefield,
// whoever controls it (a layer-6 grant, so it follows the creature's own
// summoning sickness: a creature tapped for mana needs to have been under its
// controller's control since their turn began, CR 302.6). The chaos ability
// untaps the planar controller's creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11b3fcb2-0b66-4d8b-8e66-a21a5da1c1f5",
		Name:         "Llanowar",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{TapForManaGrant(llanowarGrant, "{G}{G}", "Add {G}{G}", "{T}: Add {G}{G}.")},
		Static:       []game.StaticAbility{Labelled(GrantAbilities(allCreatures, llanowarGrant), "All creatures have \"{T}: Add {G}{G}.\"")},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Llanowar — untap all creatures you control", Do(UntapAllCreaturesYouControl{})),
		},
	})
}

// llanowarGrant is the granted bundle's key. An on-disk key.
const llanowarGrant = "llanowar/tap-for-gg"
