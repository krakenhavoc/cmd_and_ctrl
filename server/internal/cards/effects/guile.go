package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guile — Creature — Elemental Incarnation {3}{U}{U}{U}, 6/6:
//
//	"This creature can't be blocked except by three or more creatures.
//	 If a spell or ability you control would counter a spell, instead
//	 exile that spell and you may play that card without paying its mana
//	 cost.
//	 When Guile is put into a graveyard from anywhere, shuffle it into
//	 its owner's library."
//
// Three lines, three seams:
//
//   - The block rule is Pathrazer of Ulamog's: MinBlockers(OnSelf(), 3),
//     so one or two blockers are refused at declaration.
//   - The middle line is a replacement of the act of countering
//     (#4056, RepEventCounterSpell, ADR 0113's 2026-10-19 amendment),
//     ExileSpellsYouWouldCounterInstead: the spell is exiled and NOT
//     countered, and the countering player may cast it for free. It
//     covers every counter the engine has (a hard counter, a declined
//     counter-unless-pays, a ward, Remand's and Hinder's destination
//     counters) because they share one window.
//   - The last line is the titans' graveyard trigger (#2038,
//     WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn).
//
// No simplification beyond the one every resolution-time free cast
// shares: the play is an ordinary cast taken after the counterspell
// finishes resolving, and closes on the player's next priority pass.
func init() {
	Register(Spec{
		OracleID:     "591bbf84-259d-4885-8b5f-b29c04222d41",
		Name:         "Guile",
		Completeness: CompletenessFull,
		BlockRules:   []game.BlockRule{MinBlockers(OnSelf(), 3)},
		Replacements: []game.ReplacementEffect{
			ExileSpellsYouWouldCounterInstead(),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn("Guile — shuffle it into its owner's library"),
		},
	})
}
