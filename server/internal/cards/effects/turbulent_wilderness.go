package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turbulent Wilderness — Land — Forest Island (EDHREC rank 4179):
//
//	"({T}: Add {G} or {U}.)
//	 This land enters tapped unless your opponents control eight or
//	 more lands."
//
// The Simic member of the "catch-up" typed duals. It is in the batch
// because the cycle's Golgari member — Turbulent Fen, batch 34 —
// already established the shape, and a cycle registered in halves is
// a mana base that behaves differently depending on which two colours
// a deck happens to play.
//
// Two basic land types on a NONBASIC land derive no mana ability in
// this engine (the synthetic one fires only for the basic supertype),
// so the reminder-text pipe is declared. The tapped entry is a CR 614
// self-replacement rather than a tap-on-entry, so the land is never
// briefly untapped on the battlefield for something to notice.
//
// The condition counts every land the controller's OPPONENTS control,
// between them — post-layer types, so an animated land counts. Your
// own lands never count towards it, which is the whole design: the
// player who is behind on mana gets an untapped dual and the player
// who is ahead does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "bd8adca6-4f16-45f8-994a-fe55bd573bd0",
		Name:          "Turbulent Wilderness",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(b40CatchUpDualCondition())},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "U")},
	})
}
