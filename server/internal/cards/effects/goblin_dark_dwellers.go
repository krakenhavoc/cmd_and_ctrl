package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Dark-Dwellers — Creature — Goblin {3}{R}{R}, 4/4 (Foundations
// Jumpstart):
//
//	"Menace
//	 When this creature enters, you may cast target instant or sorcery
//	 card with mana value 3 or less from your graveyard without paying
//	 its mana cost. If that spell would be put into your graveyard,
//	 exile it instead."
//
// Daring Waverider's clause (freeCastTargetFromYourGraveyardExilingIt)
// at mana value 3. No simplification.
func init() {
	enters := WhenThisEnters("Goblin Dark-Dwellers — you may cast target instant or sorcery card with mana value 3 or less from your graveyard without paying its mana cost",
		freeCastTargetFromYourGraveyardExilingIt("Goblin Dark-Dwellers"))
	enters.Targets = TargetCardInGraveyard("target instant or sorcery card with mana value 3 or less from your graveyard",
		Or(Instant(), Sorcery()), ManaValueLE(3), YouOwn())
	Register(Spec{
		OracleID:        "bab0ab8d-74d2-49a3-b258-b93d19925d99",
		Name:            "Goblin Dark-Dwellers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
