package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Immerwolf — {1}{R}{G} Creature — Wolf 2/2 (#4071, ADR 0132 amendment
// 2026-10-16):
//
//	"Intimidate (This creature can't be blocked except by artifact
//	 creatures and/or creatures that share a color with it.)
//	 Each other creature you control that's a Wolf or a Werewolf gets
//	 +1/+1.
//	 Non-Human Werewolves you control can't transform."
//
// The lord is the shared tribal anthem over your OTHER Wolves and
// Werewolves. The third ability is the can't-transform seam
// (NonHumanWerewolvesYouControlCantTransform, game/cant_transform.go):
// "can't" beats every instruction to transform (CR 101.2), so a back
// face stays up whatever its own upkeep trigger, Moonmist or a
// nightbound werewolf's day turn-over says. A front face that is a Human
// Werewolf is not covered, so it still turns over, and once it has it is
// held there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c2445ee-84d5-47de-8b09-dbdc89ca2db7",
		Name:            "Immerwolf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, Others: true, YoursOnly: true}, 1, 1),
			NonHumanWerewolvesYouControlCantTransform(),
		},
	})
}
