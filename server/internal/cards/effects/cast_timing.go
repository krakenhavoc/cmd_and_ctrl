package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_timing.go — S42, #1195: the card-side vocabulary for "when may
// this player begin to cast" (CR 307.1, CR 702.8).
//
// Two shapes, and they are ADR 0066's two homes said again:
//
//   - A statement a PERMANENT makes while it is on the battlefield.
//     Not code at all: `Spec.CastTimings`, derived from the
//     battlefield on every query, so two Vedalken Orreries compose and
//     one leaving cannot revoke the other's. Vedalken Orrery, Leyline
//     of Anticipation, Yeva, Teferi, Time Raveler's static.
//   - A statement an EFFECT makes, with a duration that outlives its
//     source. `GrantCastTiming`, stored on the player. Emergence
//     Zone sacrifices itself and the permission lasts the turn;
//     Teferi's +1 resolves and the planeswalker may die before your
//     next turn comes.
//
// The constructors below carry the two halves a card file gets wrong:
// the Affects clause ("you" versus "each opponent") and the printed
// label. A card that spells a clause by hand is a card that can spell
// it differently from the next one.

// CastAsThoughFlash is "You may cast spells as though they had flash"
// (CR 702.8) — Vedalken Orrery, Leyline of Anticipation, Alchemist's
// Refuge's granted half, Emergence Zone's.
//
// The zero filter is "spells". Narrow it with the two below.
func CastAsThoughFlash() game.CastTimingRule {
	return game.CastTimingRule{
		Timing: game.TimingFlash,
		Label:  "You may cast spells as though they had flash.",
	}
}

// CastKindAsThoughFlash is CastAsThoughFlash narrowed to a kind of
// spell: Yeva, Nature's Herald's creature spells, Shimmer Myr's
// artifact spells, Teferi, Time Raveler's sorcery spells.
//
// `label` is the printed clause, because the filter cannot be turned
// back into English and a player reading the log deserves the card's
// own words.
func CastKindAsThoughFlash(filter game.PermissionFilter, label string) game.CastTimingRule {
	return game.CastTimingRule{
		Timing: game.TimingFlash,
		Filter: filter,
		Label:  label,
	}
}

// OpponentsCastAtSorcerySpeed is "Each opponent can cast spells only
// any time they could cast a sorcery" — Teferi, Time Raveler and
// Teferi, Mage of Zhalfir.
//
// CR 101.2 is the engine's business, not the card's: the restriction
// beats every grant, including one an opponent's own Vedalken Orrery
// would make, because CastTimingOpenLocked reads the restrictions
// last.
func OpponentsCastAtSorcerySpeed() game.CastTimingRule {
	return game.CastTimingRule{
		Timing:  game.TimingSorcery,
		Affects: game.TimingAffectsEachOpponent,
		Label:   "Each opponent can cast spells only any time they could cast a sorcery.",
	}
}

// PlayersCastOnlyOnTheirOwnTurns is "Players can cast spells only
// during their own turns" — Dosan the Falling Leaf.
//
// NOT sorcery speed: it leaves every instant-speed window on a
// player's OWN turn open and shuts every window on anybody else's,
// which is why game.TimingYourTurnOnly is a value of its own.
func PlayersCastOnlyOnTheirOwnTurns() game.CastTimingRule {
	return game.CastTimingRule{
		Timing:  game.TimingYourTurnOnly,
		Affects: game.TimingAffectsEachPlayer,
		Label:   "Players can cast spells only during their own turns.",
	}
}

// GrantCastTiming is the EFFECT that stores a timing statement on a
// player — the half with a real duration.
//
// Two windows, and they are the two the cards print. The zero
// `UntilYourNextTurn` is "this turn" (Emergence Zone, Alchemist's
// Refuge, Winding Canyons); setting it is "until your next turn"
// (Teferi, Time Raveler's +1), stamped against the controller's own
// seat-turn count by ADR 0063's builder.
type GrantCastTiming struct {
	// Timing is the statement. Almost always game.TimingFlash: an
	// effect that RESTRICTS with a duration has no card yet.
	Timing game.GrantTiming

	// Filter narrows which spells. The zero filter is "spells".
	Filter game.PermissionFilter

	// UntilYourNextTurn picks the longer of the two windows.
	UntilYourNextTurn bool

	// Player is who the statement is about. uuid.Nil means the
	// controller, which is every card that prints "you may cast".
	// Winding Canyons' "target player may cast" is why the field
	// exists.
	Player uuid.UUID

	// Label is the clause as printed, for the log.
	Label string
}

func (e GrantCastTiming) Apply(ctx *Context) error {
	player := e.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	timing := e.Timing
	if timing == game.TimingNormal {
		timing = game.TimingFlash
	}
	var window game.Duration
	if e.UntilYourNextTurn {
		window = DurationUntilYourNextTurn(ctx, player)
	}
	// A zero Duration is stamped "until end of turn" by the one write
	// path (game.GrantCastTimingForEffect), which is what "this turn"
	// means and what a card file that forgets to say should get.
	//
	// The grant lands as a game.PlayerStatic (#1197): the player, the
	// source, the clause and the window are ITS fields, and the rule
	// below is only what the statement says.
	ctx.Game.GrantCastTimingForEffect(player, game.CastTimingRule{
		Timing: timing,
		Filter: e.Filter,
	}, e.Label, ctx.Source(), window)
	return nil
}
