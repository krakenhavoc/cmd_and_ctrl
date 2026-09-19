package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert — Land — Desert (EDHREC rank 4539):
//
//	"{T}: Add {C}.
//	 {T}: This land deals 1 damage to target attacking creature.
//	 Activate only during the end of combat step."
//
// The original Desert, from Arabian Nights, and still the reason
// "Desert" is a land subtype anyone cares about. It is a colourless
// land that kills an x/1 attacker for free once a turn cycle — and
// the timing restriction is what keeps it honest: the damage lands in
// the END OF COMBAT step, after combat damage has already been dealt,
// so it does not save you from the attack it punishes.
//
// Two abilities on one land, and only one can be used per untap
// because both tap it.
//
//   - {T}: Add {C} is a CR 605 mana ability — no stack, no response.
//   - The ping is an ordinary activated ability that uses the stack,
//     gated by an activation CONDITION (CR 602.1b, #743) rather than
//     by sorcery speed: the end of combat step is not a main phase,
//     so SorcerySpeed would have forbidden exactly the window the
//     card opens.
//
// "Target ATTACKING creature" is read at announce and re-checked at
// resolution (CR 608.2b), so a creature removed from combat in
// response fizzles the ability. The target is any attacking creature,
// including your own — the card says nothing about whose it is, and
// pinging your own attacker to trigger something is a real line.
//
// Cycling ({2}, discard) was never printed on this card; the Deserts
// that cycle are the Amonkhet ones.
//
// The ping had no legal target when this card was written: the engine
// cleared AttackingTarget on ENTRY to the end of combat step, so there
// was no attacking creature during the one window the ability is
// allowed in. It was declared anyway rather than dropped, on the
// posture Darksteel Citadel's indestructible took — and #785 moved the
// clear to where CR 511.3 puts it, as the step ENDS, so the card
// started working with no change to this file. The caveat came off
// with that fix.
func init() {
	Register(Spec{
		OracleID:     "195107ad-879d-4b02-a44a-a3ba70fedf88",
		Name:         "Desert",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{T}: This land deals 1 damage to target attacking creature. Activate only during the end of combat step.",
			Cost:      TapCost(),
			Condition: DuringStep(game.StepEndCombat),
			Targets:   TargetCreature("target attacking creature", AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
