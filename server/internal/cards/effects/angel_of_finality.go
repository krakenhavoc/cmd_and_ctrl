package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Finality — Creature — Angel {3}{W}, 3/4 (EDHREC rank
// 4299):
//
//	"Flying
//	 When this creature enters, exile target player's graveyard."
//
// Bojuka Bog on a body, and worth a card in a format where the
// reanimator deck's whole plan is one graveyard. The 3/4 flier is not
// incidental: white's graveyard hate is usually a land or a two-mana
// enchantment, and this one blocks.
//
// The trigger is a real "target player" rather than an entry hook, for
// the same reason Bojuka Bog's is: the target is chosen when the
// trigger goes on the stack, so the graveyard that gets emptied is the
// one that exists at RESOLUTION. A player who mills themselves in
// response loses those cards too, and a reanimation spell cast in
// response resolves first and saves its creature. The two cards share
// one body (exileTargetPlayersGraveyard) because they are one
// sentence.
//
// Mandatory and never a fizzle: every seated player is a legal target,
// the Angel's own controller included, so a table with nothing in any
// graveyard still puts the trigger on the stack and exiles an empty
// pile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48d14b5c-711a-4f8a-9de5-55415cb7a79a",
		Name:            "Angel of Finality",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Angel of Finality — exile target player's graveyard",
					exileTargetPlayersGraveyard),
				TargetPlayer("target player")),
		},
	})
}
