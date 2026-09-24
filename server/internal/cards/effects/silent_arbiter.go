package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silent Arbiter — Artifact Creature — Construct {4}, 1/5 (EDHREC rank
// 1161):
//
//	"No more than one creature can attack each combat.
//	 No more than one creature can block each combat."
//
// The combat-wide count limit, both halves, and the card #1507 was
// built for. Each line is one declaration and binds every player at
// the table, the Arbiter's own controller included:
//
//   - the attack line is an AttackLimit (CR 508.1c), judged on the
//     declaration by DeclareAttacker / DeclareAttackers and withheld by
//     the legal-move enumerator, so a second attacker is refused with
//     illegal_attack / attack_limit and never offered;
//   - the block line is a BlockRule.Limit (CR 509.1b), a third set
//     check in the block validator, refused as declaration_limit.
//     Every stored block in the combat counts, whichever defender made
//     it: this engine stages each defender's declaration separately,
//     so in a multiplayer combat the first defender to block uses the
//     one up (ADR 0045 Decision 43).
func init() {
	Register(Spec{
		OracleID:     "1cdf30de-d88c-421a-80df-4917bbd2f09e",
		Name:         "Silent Arbiter",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackEachCombat(1)},
		BlockRules:   []game.BlockRule{NoMoreThanNCanBlockEachCombat(1)},
	})
}
