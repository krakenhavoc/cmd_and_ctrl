package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodwater Entity — Creature — Elemental {1}{U}{R}, 2/2 (Guilds of
// Ravnica):
//
//	"Flying
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature enters, you may put target instant or sorcery card
//	 from your graveyard on top of your library."
//
// Flying and prowess are printed keywords. The trigger is Mystic
// Sanctuary's: optional and targeted, so an empty graveyard means no
// trigger at all (CR 603.3d) and a card that leaves the graveyard in
// response makes it fizzle (CR 608.2b). The put goes through the shared exit
// primitive, so a commander card gets the CR 903.9 offer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94ca5db6-86e1-42b4-b776-977efa2e12b6",
		Name:            "Bloodwater Entity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Bloodwater Entity — put an instant or sorcery card from your graveyard on top of your library",
						effects.PutChosenTargetOnTopOfLibrary),
					"Bloodwater Entity — put target instant or sorcery card from your graveyard on top of your library?"),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn())),
		},
	})
}
