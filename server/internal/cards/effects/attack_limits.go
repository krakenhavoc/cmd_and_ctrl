package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// attack_limits.go — constructors for Spec.AttackLimits, the CR 508.1c
// count limits on an attack declaration (#1507). The engine half —
// the one check both declaration verbs and the legal-move enumerator
// run — is game/attack_limits.go; see ADR 0045 Decision 44.
//
//	AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackEachCombat(1)},    // Silent Arbiter
//	AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackYouEachCombat(2)}, // Crawlspace
//
// Two constructors because the family prints exactly two lines, and
// the difference between them is the one a card file must not get
// wrong: "can attack each combat" binds every attack at the table,
// the card's own controller's included; "can attack YOU" counts only
// attacks on the card's controller — the player, not a planeswalker
// they control.

// NoMoreThanNCanAttackEachCombat — "No more than N creatures can
// attack each combat" (Silent Arbiter and Dueling Grounds: one;
// Caverns of Despair: two). Every attacking creature counts.
func NoMoreThanNCanAttackEachCombat(n int) game.AttackLimit {
	return game.AttackLimit{Scope: game.AttackLimitEachCombat, Max: n}
}

// NoMoreThanNCanAttackYouEachCombat — "No more than N creatures can
// attack you each combat" (Crawlspace: two; Judoon Enforcers: one).
// Only creatures attacking this permanent's controller count, so in a
// four-player game the other opponents may still be attacked by any
// number.
func NoMoreThanNCanAttackYouEachCombat(n int) game.AttackLimit {
	return game.AttackLimit{Scope: game.AttackLimitAttackingYou, Max: n}
}
