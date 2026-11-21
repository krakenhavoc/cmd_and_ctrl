package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iraxxa, Empress of Mars — Legendary Creature — Alien Warrior {2}{R}{R}, 5/4
// (Doctor Who):
//
//	"Trample
//	 Battle cry (Whenever this creature attacks, each other attacking creature
//	 gets +1/+0 until end of turn.)
//	 Paradox — Whenever you cast a spell from anywhere other than your hand, create
//	 a 2/2 red Alien Warrior creature token."
//
// Battle cry is the engine's trigger row (CR 702.91a). The Paradox condition reads
// the zone the spell left (a commander cast from the command zone, a flashback or
// an exile cast all count; a hand cast does not).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "52925539-423a-42ec-91b9-0d39e7df565c",
		Name:            "Iraxxa, Empress of Mars",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			BattleCry("Iraxxa, Empress of Mars"),
			On(game.EventCast, gapB87YouCastFromOutsideYourHand, "Paradox — create a 2/2 red Alien Warrior creature token",
				Do(CreateToken{Template: TokenCard("2/2 red Alien Warrior"), N: 1})),
		},
	})
}
