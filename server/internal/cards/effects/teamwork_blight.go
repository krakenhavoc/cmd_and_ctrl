package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teamwork_blight.go — #1703: the card side of the two optional
// additional costs game/teamwork_blight_cost.go pays.

// Teamwork is CR 702.194a's "Teamwork N": "As an additional cost to
// cast this spell, you may tap any number of creatures you control
// with total power N or more." The caster names the creatures on
// cast_spell as teamwork_ids. The engine validates them the way it
// validates crew (effective power, untapped, no summoning-sickness
// check) and taps them with the spell on the stack.
//
// "Cast using teamwork" (CR 702.194b) is ctx.UsedTeamwork() at
// resolution.
//
//	OptionalCosts: []game.AdditionalCost{Teamwork(4)},   // HULK SMASH!
func Teamwork(n int) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.TeamworkKey,
		Teamwork: n,
		Label:    fmt.Sprintf("Teamwork %d", n),
	}
}

// OptionalBlight is "As an additional cost to cast this spell, you may
// blight N" (CR 701.68a): put N -1/-1 counters on a creature you
// control. The caster names the creature on cast_spell as blight_ids.
// The counters go through the CR 614 window as a COST, so Doubling
// Season does not double them and Vizier of Remedies does reduce them
// (see game.blightLocked).
//
// "If this spell's additional cost was paid" is ctx.BlightPaid() at
// resolution.
//
//	OptionalCosts: []game.AdditionalCost{OptionalBlight(2)},   // Pyrrhic Strike
func OptionalBlight(n int) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.BlightKey,
		Blight:   n,
		Label:    fmt.Sprintf("Blight %d", n),
	}
}

// UsedTeamwork is CR 702.194b's "if this spell was cast using
// teamwork": the caster announced the teamwork cost.
func (c *Context) UsedTeamwork() bool { return c.OptionalCostTimes(game.TeamworkKey) > 0 }

// BlightPaid is "if this spell's additional cost was paid" on a card
// whose additional cost is an optional blight.
func (c *Context) BlightPaid() bool { return c.OptionalCostTimes(game.BlightKey) > 0 }

// TeamworkUsed is the teamwork cards' mode count — "Choose one. If this
// spell was cast using teamwork, choose both instead" is
// ChooseOne(…).InsteadIf(2, TeamworkUsed). It reads the teamwork cost
// announced WITH the modes (CR 601.2b, #1655), never the board.
var TeamworkUsed = game.ModeConditionOnAnnouncement("used-teamwork", func(_ *game.Game, q game.ModeCountQuery) bool {
	return q.OptionalCostTimes(game.TeamworkKey) > 0
})

// BlightUsed is the same count for a card whose optional cost is a
// blight — "If this spell's additional cost was paid, choose both
// instead" (Pyrrhic Strike).
var BlightUsed = game.ModeConditionOnAnnouncement("blight-paid", func(_ *game.Game, q game.ModeCountQuery) bool {
	return q.OptionalCostTimes(game.BlightKey) > 0
})

// ModeClauseTarget is the still-legal target (CR 608.2b) announced for
// clause `slot` of mode occurrence `occ` — the read a bullet with TWO
// target clauses needs ("target creature you control deals damage …
// to target creature an opponent controls"), where ModeTarget would
// hand back whichever clause came first. (zero, false) when that
// clause's target has gone.
func ModeClauseTarget(ctx *Context, occ, slot int) (game.TargetRef, bool) {
	for _, t := range ctx.ModeTargets(occ) {
		if t.Slot == slot && ctx.IsTargetLegal(t) {
			return t, true
		}
	}
	return game.TargetRef{}, false
}

// checkTeamworkBlight is the Register guard for the two components.
// Each shape it refuses compiles and then behaves as something the
// card does not print:
//
//   - either component in the MANDATORY slot. Teamwork is optional by
//     definition (CR 702.194a); a mandatory blight exists in print only
//     as "blight N or pay {M}" and "blight X", neither of which this
//     component is.
//   - either one under the wrong key, or hand-rolled with a count below
//     one: "cast using teamwork" reads the key.
//   - either one repeating, or sharing a cost with another component:
//     the payment lists are one creature set per cost.
//   - two of the same on one card: the plan sums them, and no card
//     prints that.
func checkTeamworkBlight(spec Spec) {
	if ac := spec.AdditionalCost; ac != nil && (ac.Teamwork != 0 || ac.Blight != 0) {
		panic(fmt.Sprintf("effects.Register: %q puts teamwork or blight in the mandatory AdditionalCost slot — declare Teamwork(n) / OptionalBlight(n) in OptionalCosts (#1703)", spec.Name))
	}
	teamwork, blight := 0, 0
	for _, oc := range spec.OptionalCosts {
		if oc.Teamwork == 0 && oc.Blight == 0 {
			continue
		}
		if oc.Teamwork < 0 || oc.Blight < 0 || (oc.Teamwork != 0 && oc.Blight != 0) {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q is a malformed teamwork / blight cost — build it with Teamwork(n) or OptionalBlight(n)", spec.Name, oc.Key))
		}
		if oc.Teamwork > 0 {
			teamwork++
			if oc.Key != game.TeamworkKey {
				panic(fmt.Sprintf("effects.Register: %q declares a teamwork cost keyed %q — build it with Teamwork(n)", spec.Name, oc.Key))
			}
		}
		if oc.Blight > 0 {
			blight++
			if oc.Key != game.BlightKey {
				panic(fmt.Sprintf("effects.Register: %q declares a blight cost keyed %q — build it with OptionalBlight(n)", spec.Name, oc.Key))
			}
		}
		// A target-clause rewrite (Targets, set by WhenPaid) is NOT a
		// payment component and is allowed: "if this spell was cast
		// using teamwork, instead … target …" (Cruel Alliance, Too
		// Evil to Stay Dead) is what the field exists for (#1716).
		if oc.MaxPayments() > 1 || oc.ManaCost != "" || oc.DiscardCards != 0 || oc.Sacrifice != nil || oc.PayLifeX || oc.ChoosesOpponent {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q mixes teamwork / blight with another component — build it with Teamwork(n) or OptionalBlight(n)", spec.Name, oc.Key))
		}
	}
	if teamwork > 1 || blight > 1 {
		panic(fmt.Sprintf("effects.Register: %q declares two teamwork or two blight costs", spec.Name))
	}
	if teamwork > 0 && spec.TapCost != nil {
		panic(fmt.Sprintf("effects.Register: %q declares teamwork beside convoke / waterbend — no printed card does, and the two tap lists would share creatures", spec.Name))
	}
}
