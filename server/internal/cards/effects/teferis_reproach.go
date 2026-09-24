package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi's Reproach — Instant {2}{W}:
//
//	"Choose target opponent. Until that player's next turn, they gain
//	 protection from everything and their life total can't change. All
//	 nonland permanents they control phase out. Exile Teferi's
//	 Reproach."
//
// Teferi's Protection pointed the other way, and it is here because
// that is the sharpest proof #1200's shape is right: every one of the
// four clauses is the SAME machinery aimed at a seat that is not the
// caster, and not one line of engine code had to learn the
// difference.
//
//   - the life-total lock is a PlayerStatic on the target's seat, with
//     a CR 611.2 duration counted against THEIR seat-turn counter
//     (#1200, ADR 0085);
//   - the protection is the same grant on the same slice (#1197,
//     ADR 0072's amendment);
//   - the phase-out is CR 702.26 (#1199, ADR 0084), filtered to
//     nonland permanents and taken by CONTROLLER;
//   - the self-exile is #489's spellMovedItselfLocked, which keeps the
//     resolution frame from routing the card to a graveyard it moved
//     itself out of.
//
// WHY YOU CAST IT AT SOMEBODY ELSE. It is not a gift: the shield is a
// PRISON for a turn cycle. The target's whole nonland board is treated
// as though it does not exist (CR 702.26b) — no blockers, no mana from
// anything but lands, no activated abilities, nothing to sacrifice —
// and they cannot be targeted, cannot be damaged, and cannot pay life
// for anything while it holds. Aimed at the opponent whose turn is
// next, it takes that turn away; aimed at a player about to be killed
// by somebody else, it is a rescue you get to charge for.
//
// "UNTIL THAT PLAYER'S NEXT TURN", not yours. DurationUntilYourNextTurn
// takes the player it is about, so both grants are stamped against the
// TARGET's Player.TurnsBegun and end as the target's next turn begins
// (CR 611.2b, CR 500.1) — which is also the untap step the phased-out
// permanents come back in (CR 502.1), so the board and the shield
// arrive and leave together without either clause knowing about the
// other.
//
// NONLAND, so the target keeps their mana base: the card is a turn
// tax, not a Armageddon. Taken by CONTROLLER rather than owner, and
// snapshotted before anything moves, because the rule is simultaneous
// and a permanent attached to another permanent in the same list must
// not be dragged out twice (CR 702.26h).
//
// A target that became illegal in response (hexproof granted, the
// seat conceded) drops the whole spell: it is the only target, and
// CR 608.2b fizzles a spell whose targets are all illegal before
// OnResolve runs. The self-exile still happens, because it is part of
// the spell's own text — but a fizzled spell never reaches this
// function at all, and the CR 608.2n graveyard route takes it instead,
// which is what the rules say.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9381b4a5-a8e0-412d-bc7f-ae15afa0f135",
		Name:         "Teferi's Reproach",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			them, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			d := DurationUntilYourNextTurn(ctx, them)
			if err := (LockLifeTotal{
				Player:   them,
				Label:    "Teferi's Reproach — their life total can't change",
				Duration: d,
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (GainPlayerKeyword{
				Player:   them,
				Keyword:  ProtectionFromEverything,
				Label:    "Teferi's Reproach — protection from everything",
				Duration: d,
			}).Apply(ctx); err != nil {
				return err
			}
			// "All nonland permanents they control phase out"
			// (CR 702.26). Snapshotted before any of it moves, and
			// handed over in ONE call — see the doc comment.
			var theirs []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == them && !c.IsLand() {
					theirs = append(theirs, c.InstanceID)
				}
			}
			if err := (PhaseOut{Targets: theirs}).Apply(ctx); err != nil {
				return err
			}
			// "Exile Teferi's Reproach." The spell moves ITSELF, so
			// #489's spellMovedItselfLocked stops the resolution
			// frame putting it in the graveyard afterwards.
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
