package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oubliette — Enchantment {1}{B}{B}:
//
//	"When this enchantment enters, target creature phases out until
//	 this enchantment leaves the battlefield. Tap that creature as it
//	 phases in this way. (Auras and Equipment phase out with it. While
//	 permanents are phased out, they're treated as though they don't
//	 exist.)"
//
// The Arabian Nights removal spell, restored to phasing by the 2021
// errata, and the card that proves the OTHER half of CR 702.26: a
// phase-out with a duration on it rather than one on the untap step's
// clock.
//
// It is here because "phases out until ~ leaves the battlefield" is a
// shape of its own and would otherwise have been a seam left open.
// The engine's answer is Card.PhaseInLockedBy (ADR 0084 Decision 2):
// while Oubliette is on the battlefield the creature is skipped by
// CR 502.1's phase-in half, and the moment Oubliette leaves — Naturalised,
// destroyed, bounced, exiled — the creature phases back in, right
// then, in the state-based-action pass. It does NOT wait for an untap
// step, which is what "until" means on these cards and what makes
// this family (Out of Time is the other) work at all.
//
// "Tap that creature as it phases in this way" is the rider, and it
// is why Oubliette is removal rather than a delay: the creature comes
// back tapped, so it cannot block the turn it returns. Carried as
// Card.TapOnPhaseIn and consumed as it fires, because the phase-in it
// rides on is a duration ending and uses no stack (CR 702.26a) — there
// is nothing for a delayed trigger to sit on.
//
// WHY IT IS STRICTLY BETTER THAN AN EXILE HERE AND WORSE THERE, both
// as printed: the creature keeps its counters, its Auras and its
// Equipment (CR 702.26d, CR 702.26g), so a Walking Ballista comes back
// loaded — but it also never dies, so nothing sees it leave and no
// death trigger fires. Both are the engine's, not this file's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c753e9e3-9374-4e3c-8622-94576a8c1da3",
		Name:         "Oubliette",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetCreature("target creature"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				until := source.InstanceID
				return game.NewTriggeredItem(source,
					"Oubliette — target creature phases out until Oubliette leaves the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return PhaseOutUntilLeaves{
							Targets:      legalTargetCards(item, g),
							Until:        until,
							TapOnPhaseIn: true,
						}.Apply(ctx)
					})
			},
		}},
	})
}
