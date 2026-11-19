package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadhorde Arcanist — Creature — Zombie Wizard {1}{R}, 1/3:
//
//	"Trample
//	 Whenever this creature attacks, you may cast target instant or
//	 sorcery card with mana value less than or equal to this creature's
//	 power from your graveyard without paying its mana cost. If that
//	 spell would be put into your graveyard, exile it instead."
//
// Arcane Proxy's power-relative target and Daring Waverider's free cast
// from the graveyard with the exile rider.
func init() {
	Register(Spec{
		OracleID:        "4c05b382-58ab-4a2d-a81c-408ea273b6b6",
		Name:            "Dreadhorde Arcanist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Dreadhorde Arcanist — you may cast target instant or sorcery card with mana value less than or equal to this creature's power from your graveyard without paying its mana cost",
					freeCastTargetFromYourGraveyardExilingIt("Dreadhorde Arcanist")),
				RelativeToSource(
					TargetCardInGraveyard("target instant or sorcery card with mana value less than or equal to Dreadhorde Arcanist's power from your graveyard",
						YouOwn(), Or(Instant(), Sorcery())),
					ManaValueNoGreaterThanPower())),
		},
	})
}
