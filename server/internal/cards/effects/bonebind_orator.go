package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonebind Orator — Creature — Squirrel Warlock Bard {1}{B}, 2/2
// (Bloomburrow, #4467):
//
//	"{3}{B}, Exile this card from your graveyard: Return another target
//	 creature card from your graveyard to your hand."
//
// Theoretical Necromancer's ability, word for word. It works only from
// the graveyard, at instant speed. The target is chosen before the cost
// is paid (CR 602.2b), so "another" keeps the Orator from naming itself
// even though the cost then exiles it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c79cf0e5-52dd-4af3-ab06-6cdaab526c23",
		Name:         "Bonebind Orator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}, Exile this card from your graveyard: Return another target creature card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{3}{B}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: Another(TargetCardInGraveyard("another target creature card from your graveyard", YouOwn(), Creature())),
			Effect:  returnTargetGraveyardCardToHand,
		}},
	})
}
