package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Linebreaker Baloth — Creature — Beast {3}{G}{G}, 4/5 (Dominaria United,
// #5487, seam #5658):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)
//	 This creature can't be blocked by creatures with power 2 or less."
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
// The block restriction is Cold-Blooded Crew's, judged as blockers are
// declared (CR 509.1b) on each would-be blocker's current power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c3b395b3-fea0-41d9-a3f2-6c90f89d8bdb",
		Name:            "Linebreaker Baloth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
