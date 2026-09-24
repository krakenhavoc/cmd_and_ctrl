package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// attack_limits.go — constructors for Spec.AttackLimits, the CR 508.1c
// count limits on an attack declaration (#1507, #1534). The engine
// half — the one check both declaration verbs and the legal-move
// enumerator run — is game/attack_limits.go; see ADR 0045 Decisions 44
// and 46.
//
//	AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackEachCombat(1)},    // Silent Arbiter
//	AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackYouEachCombat(2)}, // Crawlspace
//	AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackThisEachCombat(1)}, // The Eternal Wanderer
//	AttackLimits: []game.AttackLimit{AsLongAs(ThisIsTapped,                  // Mirri, Weatherlight Duelist
//	    NoMoreThanNCanAttackYouEachCombat(1))},
//
// One constructor per printed line, because the difference between
// them is the one a card file must not get wrong: "can attack each
// combat" binds every attack at the table, the card's own
// controller's included; "can attack YOU" counts only attacks on the
// card's controller — the player, not a planeswalker they control;
// "can attack <this>" counts only attacks on the permanent itself.

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

// NoMoreThanNCanAttackThisEachCombat — "No more than N creatures can
// attack <this permanent> each combat" (The Eternal Wanderer: one).
// Only creatures attacking the permanent itself count: its controller,
// and any other planeswalker they control, may be attacked by any
// number (#1534).
func NoMoreThanNCanAttackThisEachCombat(n int) game.AttackLimit {
	return game.AttackLimit{Scope: game.AttackLimitAttackingThis, Max: n}
}

// AsLongAs gates a limit on a condition of its source — "As long as
// Mirri is tapped, no more than one creature can attack you each
// combat" (#1534). The condition is read live at every check; one that
// turns true after attackers are declared unmakes none of them, the
// same rule as a limit that arrives late.
func AsLongAs(cond func(g *game.Game, source *game.Card) bool, l game.AttackLimit) game.AttackLimit {
	l.While = cond
	return l
}

// ThisIsTapped is the condition "as long as <this permanent> is
// tapped", for AsLongAs.
func ThisIsTapped(_ *game.Game, source *game.Card) bool {
	return source != nil && source.Tapped
}
