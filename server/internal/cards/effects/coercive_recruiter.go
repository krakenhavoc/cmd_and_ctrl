package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coercive Recruiter — Creature — Orc Pirate {4}{R}, 4/3 (issue
// #1112):
//
//	"Whenever this creature or another Pirate you control enters,
//	 gain control of target creature until end of turn. Untap that
//	 creature. Until end of turn, it gains haste and becomes a Pirate
//	 in addition to its other types."
//
// A repeatable Act of Treason that a Pirate deck sets off with its
// own creature spells, plus a coat of paint the theft leaves behind
// so the stolen creature can crew, fits an "attacks with a Pirate"
// payoff, or count for a Pirate tribal effect for the rest of the
// turn.
//
// The trigger condition is "this creature or another Pirate you
// control enters" — and the Recruiter is itself a Pirate, so that is
// exactly enteredUnderYourControl (Corsair Captain's helper, #236)
// with `another` false, filtered by isPirate: a card entering under
// your control that is a Pirate, the Recruiter's own entry included.
//
// The four printed sentences are Act of Treason's three primitives
// (#756) plus one more:
//
//   - GainControl, until end of turn (CR 613.1b) — reverts itself at
//     cleanup, no bookkeeping needed.
//   - UntapTarget — the creature you take is usually tapped.
//   - GrantKeywordUntilEOT{"haste"} — CR 302.6 would otherwise make
//     the theft useless for an immediate attack.
//   - "becomes a Pirate in addition to its other types" is a layer 4
//     type add, snapshotted to the stolen creature's own instance and
//     entry stamp exactly as BecomeCreatureUntilEOT pins a Vehicle
//     (CR 400.7, CR 611.2c) — StaticUntilEOT is the right tool rather
//     than BecomeCreatureUntilEOT itself, because that constructor's
//     empty-Types default adds Artifact and Creature, which this
//     card does not print.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad33530c-a8df-4c1c-a863-501e583290b6",
		Name:         "Coercive Recruiter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && isPirate(c)
			}, "Coercive Recruiter — gain control of target creature until end of turn",
				coerciveRecruiterEffect), TargetCreature("target creature")),
		},
	})
}

// coerciveRecruiterEffect is the resolution body: steal, untap, grant
// haste, grant the Pirate type — in printed order.
//
// Caller holds g.mu.
func coerciveRecruiterEffect(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	target := item.Targets[0].ID
	if err := (GainControl{
		Target:   target,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Coercive Recruiter — gain control until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"haste"},
		Label:    "Coercive Recruiter — haste until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	set := eotSnapshot(ctx, target, nil)
	if set == nil {
		return nil
	}
	return StaticUntilEOT{
		Label: "Coercive Recruiter — becomes a Pirate in addition to its other types",
		Ability: game.StaticAbility{
			Layer:     game.Layer4Type,
			AppliesTo: set.appliesTo(),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !eotHasType(c.Subtypes, "Pirate") {
					c.Subtypes = append(c.Subtypes, "Pirate")
				}
			},
		},
	}.Apply(ctx)
}
