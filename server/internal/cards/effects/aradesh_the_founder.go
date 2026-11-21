package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aradesh, the Founder — Legendary Creature — Human Soldier {2}{W}, 1/4
// (Fallout):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)
//	 Whenever a creature you control attacks, if it enlisted a creature this
//	 combat, the creature that attacked gains double strike until end of
//	 turn. If that creature's power is 4 or greater, draw a card."
//
// A creature enlists only as it attacks (CR 702.154b), so "attacks and
// enlisted a creature this combat" is the attacker's EventEnlist
// (WheneverThisEnlists' event, here for any creature I control): it names
// the attacker in CardID and the creature tapped in Target, and goes on the
// stack with the attack triggers. Power is read as the ability resolves
// (CR 608.2h), so it sees the +X/+0 of the enlist trigger if that resolved
// first; an attacker that has left does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "32620b26-10b9-47a7-bfa7-ca590ab5fa50",
		Name:            "Aradesh, the Founder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
		Triggered: []game.TriggeredAbility{
			On(game.EventEnlist, gapYourCreatureEnlisted,
				"Aradesh, the Founder — the attacker gains double strike until end of turn; if its power is 4 or greater, draw a card",
				gapAradeshResolve),
		},
	})
}
