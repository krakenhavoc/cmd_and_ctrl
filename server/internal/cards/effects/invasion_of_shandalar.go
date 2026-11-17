package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Shandalar — Battle — Siege, defense 4, for {3}{G}{G} (March
// of the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, return up to three target permanent cards
//	 from your graveyard to your hand."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the return and the defeated
// trigger that hands over Leyline Surge. Each target is checked again as
// the ability resolves (CR 608.2b), so one that has left the graveyard is
// skipped and the rest still return.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Invasion of Shandalar — return up to three target permanent cards from your graveyard to your hand",
		returnEachTargetCardToHand)
	enters.Targets = TargetCardInGraveyard("up to three target permanent cards from your graveyard", Permanent(), YouOwn()).WithCount(0, 3)
	Register(Spec{
		OracleID:     invasionOfShandalarOracleID,
		Name:         "Invasion of Shandalar",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			enters,
			DefeatedTrigger("Invasion of Shandalar — defeated: exile it, then cast Leyline Surge", SiegeDefeated()),
		},
	})
}

// invasionOfShandalarOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfShandalarOracleID = "b5fe62f0-d914-4789-afd3-3831dcdbed7b"
