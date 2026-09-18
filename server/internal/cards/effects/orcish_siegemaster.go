package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orcish Siegemaster — 0/5 Creature — Orc Soldier for {2}{R} (EDHREC
// rank 4377):
//
//	"Trample
//	 Other Orcs and Goblins you control have trample.
//	 Whenever this creature attacks, it gets +X/+0 until end of
//	 turn, where X is the greatest power among creatures you
//	 control."
//
// A three-mana lord that turns a goblin swarm's chump-blockable
// bodies into a real clock, and then attacks as whatever your biggest
// creature is. The 0/5 base is the joke: alone it hits for nothing,
// and next to one fatty it hits for the fatty's number twice.
// Roadmap batch 42 (#449) filed it under cost modification, which
// #93 shipped; what it actually needs is a tribal keyword grant and
// an until-end-of-turn pump, both of which the catalog already had.
//
// One TribeFilter covers "Orcs and Goblins" because the filter ORs
// its tribe list — a creature that is either matches — and Others
// plus YoursOnly are the printed "other … you control". A creature
// that is both an Orc and a Goblin gets trample once; the grant
// dedupes.
//
// The Siegemaster's own trample is PrintedKeywords rather than a
// self-match on the grant, because "other" excludes it and the
// printed keyword is where the engine looks for a card's own.
//
// X is read when the trigger RESOLVES, not when attackers are
// declared (CR 608.2f), so a creature that dies to a blocker's first
// strike is too late to shrink it but a removal spell in the declare-
// attackers step is not. The Siegemaster counts itself, since the
// clause says "creatures you control" and not "other": on an empty
// board it is 0/5 and gets +0/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d852099f-46c6-4676-860c-20a582f733d1",
		Name:            "Orcish Siegemaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{
				Tribes:    []string{"Orc", "Goblin"},
				Others:    true,
				YoursOnly: true,
			}, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Orcish Siegemaster — +X/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					x := b42GreatestPowerControlledBy(g, item.Controller)
					return BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  x,
						Label:  "Orcish Siegemaster — +X/+0 until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
