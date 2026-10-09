package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hapatra, the Desert Fang — Legendary Creature — Human Cleric
// {2}{B}{B}{G}, 3/3:
//
//	"When Hapatra enters, for each opponent, put X -1/-1 counters on up
//	 to one target creature that player controls, where X is the greatest
//	 mana value among cards in your graveyard."
//
// Molten Primordial's clause-per-opponent targeting
// (moltenPrimordialClauses), so each pick is bound to its opponent and a
// creature that changed hands is an illegal target. X is read as the
// ability resolves, from the controller's graveyard.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Hapatra, the Desert Fang — X -1/-1 counters on up to one creature each opponent controls", rfCreatureCHapatraFangEffect)
	enters.TargetsFrom = moltenPrimordialClauses
	Register(Spec{
		OracleID:     "6894345f-52a6-46e4-b278-0783087daee1",
		Name:         "Hapatra, the Desert Fang",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
