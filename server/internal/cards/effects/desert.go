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
// DECLARED SIMPLIFICATION (weaker than printed): the ping has no legal
// target today, so it cannot actually be used. The engine clears
// AttackingTarget when it ENTERS the end of combat step
// (Game.runStepEntryHooksLocked, deliberately, to keep the client's
// combat arrows drawn through the damage steps), where CR 511.3
// removes creatures from combat as that step ENDS. There is therefore
// no attacking creature during the one window this ability is allowed
// in, and the legal-target enumerator never offers it.
//
// The ability is declared anyway rather than dropped, on the posture
// Darksteel Citadel's indestructible took: the day the combat clear
// moves to the end of the step where the rules put it, this card
// starts working with no change to this file. Until then the Desert
// is a colourless land — and a DESERT, which is the cost Ifnir
// Deadlands eats.
func init() {
	Register(Spec{
		OracleID:     "195107ad-879d-4b02-a44a-a3ba70fedf88",
		Name:         "Desert",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The damage ability can't be used yet: attackers stop being attackers the moment the end of combat step begins, so there is never a legal target during the one window the card allows.",
		},
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
