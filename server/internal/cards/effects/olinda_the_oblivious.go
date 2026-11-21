package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Olinda the Oblivious — Legendary Creature — Skunk Rogue {3}{B}, 3/3
// (Mystery Booster Commander, #5948):
//
//	"Whenever Olinda enters or attacks, for each opponent, put an odor counter
//	 on up to one target creature that player controls without an odor counter
//	 on it. Each of those creatures has "At the beginning of your upkeep, you
//	 lose 2 life" for as long as it has an odor counter on it. (They continue
//	 to smell after Olinda has left the battlefield.)"
//
// One ability with two conditions (WhenThisEntersOrAttacks), a clause per
// opponent as Molten Primordial has, each "up to one target creature that
// player controls" narrowed to creatures with no odor counter. Each creature
// still legal at resolution gets the counter and then the granted upkeep
// trigger for as long as it has the counter (GrantWhileItHasCounter, Makeshift
// Mannequin's shape): the grant belongs to the creature, so it outlasts
// Olinda, and "your" is whoever controls the creature at the upkeep.
//
// No simplification.
func init() {
	const grant = "olinda-the-oblivious/odor-upkeep"
	const label = "Olinda the Oblivious — for each opponent, put an odor counter on up to one target creature that player controls without one; each has an upkeep life loss while it has the counter"
	ability := WhenThisEntersOrAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
				continue
			}
			if err := GrantWhileItHasCounter(ctx, t.ID, "odor",
				"Olinda the Oblivious — loses 2 life at its controller's upkeep while it has an odor counter", grant); err != nil {
				return err
			}
		}
		return nil
	})
	ability.TargetsFrom = olindaClauses
	Register(Spec{
		OracleID:     "6ec8a9b8-233f-4de5-8f4c-3d677fbf2f30",
		Name:         "Olinda the Oblivious",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "At the beginning of your upkeep, you lose 2 life.",
			Triggered: []game.TriggeredAbility{
				AtYourUpkeep("Odor — you lose 2 life", Do(youLoseLife(2))),
			},
		}},
		Triggered: []game.TriggeredAbility{ability},
	})
}

// olindaClauses is one "up to one target creature that player controls
// without an odor counter on it" clause per opponent seat, in seat order
// (Molten Primordial's list, with the counter narrowing). The list reads the
// controller as the ability triggered and the seat list, neither of which the
// board changes.
func olindaClauses(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	controller := source.Controller
	if obj := tc.Object; obj != nil && obj.ID == source.InstanceID && obj.Controller != uuid.Nil {
		controller = obj.Controller
	}
	noOdor := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters["odor"] <= 0 }
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		clauses = append(clauses, TargetCreature("up to one target creature "+p.Name+" controls without an odor counter on it",
			ControlledBy(p.ID), noOdor).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}
