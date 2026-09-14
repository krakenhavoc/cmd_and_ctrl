package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memorial to Folly — Land (EDHREC rank 2913):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {2}{B}, {T}, Sacrifice this land: Return target creature card
//	 from your graveyard to your hand."
//
// The black Memorial: a tapped Swamp that turns into a Raise Dead
// late. Enters tapped is the self-replacement; the mana is a plain
// {B}; the activation is mana, tap and sacrifice-this with a
// graveyard target clause (targetCreatureInYourGraveyard — the zone
// browser picker), returning the pick to hand if it is still there
// at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bc38f14-0314-4351-8138-e2b8bf041404",
		Name:         "Memorial to Folly",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, {T}, Sacrifice this land: Return target creature card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{2}{B}"), TapCost(), SacrificeThis()),
			Targets: targetCreatureInYourGraveyard(),
			Effect:  b27ReturnChosenGraveyardCardToHand,
		}},
	})
}
