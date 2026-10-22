package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dream Strix — Creature — Bird Illusion {2}{U}, 3/2 (EDHREC rank
// 25105):
//
//	"Flying
//	 When this creature becomes the target of a spell, sacrifice it.
//	 When this creature dies, learn. (You may reveal a Lesson card you
//	 own from outside the game and put it into your hand, or discard a
//	 card to draw a card.)"
//
// The sacrifice is Departed Deckhand's: a trigger on the Strix being
// targeted by a SPELL (an ability's target does nothing), whose
// resolution sacrifices it if it is still the same permanent. The
// sacrifice is a death, so the dies trigger learns (CR 701.48a,
// learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bfd6a73e-d126-4afe-99db-98cf716a1772",
		Name:            "Dream Strix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, SelfTargetedByASpell,
				"Dream Strix — sacrifice it",
				SacrificeThisIfStillOnBattlefield),
			WhenThisDies("Dream Strix — learn", Do(Learn{})),
		},
	})
}
