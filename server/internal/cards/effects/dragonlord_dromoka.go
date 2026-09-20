package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonlord Dromoka — Legendary Creature — Elder Dragon {4}{G}{W},
// 5/7:
//
//	"This spell can't be countered.
//	 Flying, lifelink
//	 Your opponents can't cast spells during your turn."
//
// The Elder Dragon that plays around the whole blue half of a
// Commander table: it resolves through a counterspell, and once it is
// down nobody may answer anything on your turn — no removal on your
// attackers, no flash blocker, no instant-speed anything until you
// pass the turn.
//
// All three clauses are engine vocabulary already:
//
//   - CantBeCountered is the S23 rider, honoured at the counter choke
//     point, so a Counterspell aimed at Dromoka resolves and does
//     nothing (CR 701.6a) rather than fizzling.
//   - The keywords ride PrintedKeywords like any other printed pair.
//   - The lockout is a CR 101.2 cast restriction (#760, ADR 0073),
//     which means the refusal happens at ANNOUNCE with nothing paid,
//     and the same predicate the engine refuses on is the one the
//     move enumerator and the client's cast gate read — so a bot is
//     never offered the cast and the client greys the card rather
//     than firing an action the server will reject.
//
// The restriction is DERIVED from Dromoka being on the battlefield
// on every query, so it lifts the instant she leaves and a second
// copy changes nothing. A nil `match` is "spells", which is what she
// prints — no narrowing to a type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "82cdd612-3322-4372-810f-1ff106ea8e6a",
		Name:            "Dragonlord Dromoka",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flying", "lifelink"},
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCastDuringYourTurn(
				"Dragonlord Dromoka — your opponents can't cast spells during your turn.", nil),
		},
	})
}
