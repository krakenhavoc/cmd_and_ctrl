package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sokka, Tenacious Tactician — Legendary Creature — Human Warrior Ally
// {1}{U}{R}{W}, 3/3:
//
//	"Menace, prowess (Whenever you cast a noncreature spell, this
//	 creature gets +1/+1 until end of turn.)
//	 Other Allies you control have menace and prowess.
//	 Whenever you cast a noncreature spell, create a 1/1 white Ally
//	 creature token."
//
// #706's GRANT proof card. Batch 49 (#456) filed "other Allies you
// control have menace and prowess" under #754, the seam for granting
// an ability to another permanent. For a keyword that seam is not
// needed: the grant is an ordinary layer-6 keyword grant, and prowess
// is a keyword the engine reads off the effective ability list — so a
// granted prowess triggers exactly like a printed one.
//
// Prowess is CUMULATIVE (CR 702.108b), and that is observable here: an
// Ally that prints prowess (Ty Lee, Chi Blocker) has two instances
// under Sokka and gets +2/+2 per noncreature spell, and two Sokkas'
// worth of grants would give it three. Menace is not, so it is
// granted once however many sources grant it.
//
// The Ally tokens this makes are themselves Allies, so each one has
// menace and prowess from the moment it enters — the grant is a
// static, re-read every layer pass (CR 611.3a), not a one-shot.
//
// No simplification.
func init() {
	allies := TribeFilter{Tribes: []string{"Ally"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "6b68acc2-b9d5-495b-8054-c04bae1349f1",
		Name:            "Sokka, Tenacious Tactician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", game.KeywordProwess},
		Static: []game.StaticAbility{
			TribalKeywordGrant(allies, "menace"),
			TribalKeywordGrant(allies, game.KeywordProwess),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Sokka, Tenacious Tactician — create a 1/1 Ally",
				Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}
