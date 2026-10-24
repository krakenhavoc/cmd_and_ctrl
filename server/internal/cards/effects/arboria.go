package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arboria — World Enchantment {2}{G}{G}:
//
//	"Creatures can't attack a player unless that player cast a spell or
//	 put a nontoken permanent onto the battlefield during their last
//	 turn."
//
// #1997, ADR 0109 amendment 2026-10-23. A CR 508.1c restriction on every
// creature (CreaturesCantAttackPlayersWhoDidNothingLastTurn), read live
// at declaration off the defending player's record of their last turn
// (game/last_turn_actions.go): a spell cast, or a nontoken permanent
// they PUT onto the battlefield. The rulings shape both halves:
//
//   - "It's unusual for an ability to care who put a permanent onto the
//     battlefield, as opposed to under whose control it entered." The
//     engine records the putter: a land play's player, a permanent
//     spell's controller, the controller of the spell or ability whose
//     instruction put it (CR 608.2c), or the player an instruction names
//     ("each player puts", a search, a manifest; CR 110.2a).
//   - "Arboria's effect cares about the actions taken by players, not
//     their results." A countered spell, or a permanent that left before
//     the turn ended, still counts.
//   - "Arboria doesn't stop creatures from attacking planeswalkers": only
//     a player is refused.
//
// "Their last turn" is the most recent turn they took, so a player who
// has not taken a turn yet did neither and can't be attacked, and an
// extra turn is the turn that counts. A world permanent: the world rule
// (CR 704.5k) puts it into its owner's graveyard when a newer one
// enters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "acb3e93c-a1d3-458f-b8c3-c426cd359fa4",
		Name:         "Arboria",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CreaturesCantAttackPlayersWhoDidNothingLastTurn()},
	})
}
