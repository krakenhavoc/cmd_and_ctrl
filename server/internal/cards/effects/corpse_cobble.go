package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corpse Cobble — Instant {U}{B}:
//
//	"As an additional cost to cast this spell, sacrifice any number of
//	 creatures.
//	 Create an X/X blue and black Zombie creature token with menace,
//	 where X is the total power of the sacrificed creatures.
//	 Flashback {3}{U}{B}"
//
// The variable sacrifice is ADR 0100 §3's clause; X is ADR 0113 §1's
// record of WHICH creatures it took (owner decision 5 of ADR 0100): each
// one's power as it last existed on the battlefield (CR 608.2h), summed
// with negatives included and the total floored at zero (CR 107.1b).
//
// Zero creatures is a legal payment and makes a 0/0 Zombie, which the
// state-based actions then put into the graveyard — its "enters" and
// "dies" triggers still happen (the 2021-09-24 ruling). Cast with
// flashback, the additional cost is still paid (the same ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "5b0b6c8f-472a-4ee2-8368-3b17a2df498d",
		Name:             "Corpse Cobble",
		Completeness:     CompletenessFull,
		AdditionalCost:   SacrificeAnyNumberCost("any number of creatures", Creature()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: corpseCobbleZombie(ctx.SacrificedTotalPower()), N: 1}.Apply(ctx)
		},
	})
}

// corpseCobbleZombie is the X/X blue and black Zombie with menace. Built
// by hand because its size is decided at resolution, as Phyrexian
// Rebirth's Horror is. PrintedPTKnown, because 0/0 is a real size here:
// with nothing sacrificed the token must die to CR 704.5f.
func corpseCobbleZombie(x int) game.Card {
	return game.Card{
		Name:           "Zombie",
		TypeLine:       "Token Creature — Zombie",
		Power:          x,
		Toughness:      x,
		PrintedPTKnown: true,
		Colors:         []string{"U", "B"},
		Keywords:       []string{"menace"},
	}
}
