package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Domri, Anarch of Bolas — Legendary Planeswalker — Domri {1}{R}{G},
// starting loyalty 3:
//
//	"Creatures you control get +1/+0.
//	 +1: Add {R} or {G}. Creature spells you cast this turn can't be
//	     countered.
//	 −2: Target creature you control fights target creature you don't
//	     control."
//
// The static is an ordinary layer 7c anthem over Domri's controller's
// creatures (b16Anthem), live while Domri is on the battlefield.
//
// THE +1 IS A LOYALTY ABILITY, NOT A MANA ABILITY (CR 605.1a), so it
// uses the stack and is activated only at sorcery speed (CR 606.3). As
// it resolves its controller picks {R} or {G}, and gets a "this turn"
// grant (ADR 0106 §4 decision 2, #1806): every creature spell they
// cast this turn can't be countered. "Cast" is the caster, so a
// creature spell of theirs that an opponent gained control of is still
// covered and a copy is not (CR 707.10). The grant is read when
// something tries to counter, so it covers creature spells cast after
// the ability resolved (CR 611.2c), and it ends at cleanup (CR 514.2).
//
// The −2 is two target clauses; a fighter that has become an illegal
// target or left means no damage at all (CR 701.14b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "afc2269c-d3b5-487d-9445-800c7a8e526b",
		Name:            "Domri, Anarch of Bolas",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Static:          []game.StaticAbility{b16Anthem(b41CreatureYouControlStatic, 1, 0)},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Add {R} or {G}. Creature spells you cast this turn can't be countered.",
				Cost:  LoyaltyCost(1),
				Effect: Do(
					AddMana{Produced: "{R|G}"},
					GrantCounterShield{From: "Domri, Anarch of Bolas", Grant: SpellsCantBeCounteredThisTurn(
						"Creature spells you cast this turn can't be countered.",
						game.CounterShieldYouCast, game.PermissionFilter{CreatureOnly: true})},
				),
			},
			{
				Label: "−2: Target creature you control fights target creature you don't control.",
				Cost:  LoyaltyCost(-2),
				Targets: Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature you don't control", OpponentControls()),
				),
				Effect: domriFight,
			},
		},
	})
}

// domriFight is Domri's −2: the slot-0 creature fights the slot-1
// creature, and only while both are still legal targets (CR 701.14b).
func domriFight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	mine, ok := ctx.ClauseTarget(0)
	if !ok || mine.Kind != game.TargetCard {
		return nil
	}
	theirs, ok := ctx.ClauseTarget(1)
	if !ok || theirs.Kind != game.TargetCard {
		return nil
	}
	return b10Fight(ctx, mine.ID, theirs.ID)
}
