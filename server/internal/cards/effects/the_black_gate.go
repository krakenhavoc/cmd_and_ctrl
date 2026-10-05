package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Black Gate — Legendary Land — Gate:
//
//	"As The Black Gate enters, you may pay 3 life. If you don't, it
//	 enters tapped.
//	 {T}: Add {B}.
//	 {1}{B}, {T}: Choose a player with the most life or tied for most
//	 life. Target creature can't be blocked by creatures that player
//	 controls this turn."
//
// The entry is the shockland replacement with a three-life price. The
// ability targets a creature on announce and picks the player on
// resolution (the pick is a resolution-time choice, not a target). The
// pool is every seat tied for the highest life total among players still
// in the game, read when the ability resolves. The rule is a
// cantBeBlockedByPlayer ScopedEffect (#2172), pinned to the target object
// (CR 611.2c) and swept at end of turn; the blocker's controller is read
// live, so a creature that changes hands joins or leaves the barred set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40eb9904-dea3-47cf-963a-04821f98ba64",
		Name:         "The Black Gate",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnlessYouPayLife("The Black Gate", 3),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}: Choose a player with the most life or tied for most life. Target creature can't be blocked by creatures that player controls this turn.",
			Cost:    Plus(ManaCost("{1}{B}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ChoosePlayer{
					Among:    Players,
					Except:   notMostLife(g),
					Question: "The Black Gate — choose a player with the most life or tied for most life",
					Then:     blackGateBar,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// notMostLife is every seat that does not hold the highest life total
// among the players still in the game: what a "most life or tied for
// most life" pool leaves out.
func notMostLife(g *game.Game) []uuid.UUID {
	best, have := 0, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if !have || p.Life > best {
			best, have = p.Life, true
		}
	}
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		if p.Eliminated || p.Life < best {
			out = append(out, p.ID)
		}
	}
	return out
}

// blackGateBar is ChoosePlayer.Then: it reads only the Context it is
// handed, so an undo across the prompt resolves against the restored
// game. A target that left in response is skipped (CR 608.2b).
func blackGateBar(ctx *Context) error {
	p := ctx.ChosenPlayer()
	if p == uuid.Nil {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return CantBeBlockedThisTurnByPlayer{
			Target: t.ID,
			Player: p,
			Label:  "The Black Gate — can't be blocked by that player's creatures",
		}.Apply(ctx)
	}
	return nil
}
