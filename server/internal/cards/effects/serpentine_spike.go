package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serpentine Spike — Sorcery {5}{R}{R}, devoid:
//
//	"Devoid (This card has no color.)
//	 Serpentine Spike deals 2 damage to target creature, 3 damage to
//	 another target creature, and 4 damage to a third target creature.
//	 If a creature dealt damage this way would die this turn, exile it
//	 instead."
//
// Devoid is declared in PrintedKeywords, and the engine reads it as CR
// 702.114a's colour-defining ability (#2152), so the card is colourless
// in every zone. Three target
// clauses, the second and third Distinct from those before them ("another",
// "a third"), each read back by its slot, so a target that left in
// response is skipped and the others are still dealt their damage
// (CR 608.2b). Each creature dealt damage is marked from the damage's
// continuation (ADR 0108 §1 decision 3).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "c983644d-6741-4aa1-aa68-a6e680c26bb6",
		Name:            "Serpentine Spike",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
			Distinct(TargetCreature("a third target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var steps []damageStep
			for slot, amount := range []int{2, 3, 4} {
				if t, ok := ctx.ClauseTarget(slot); ok {
					steps = append(steps, damageStep{Target: t.ID, Amount: amount})
				}
			}
			return dealDamageStepsThen(ctx, steps, ExileIfDealtDamageWouldDie(item, false))
		},
	})
}
