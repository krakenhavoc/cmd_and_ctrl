package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b324_helpers.go — helpers for Dragons of Tarkir, batch 324 (S82 older
// sets, #5551): the formidable creatures.

// OldB324Formidable is the "Activate only if creatures you control have total
// power 8 or greater" of a formidable ability (CR 602.5b): the total of the
// current power of every creature the activator controls, read as the ability is
// announced (CR 602.2) through the layers, so counters, anthems and pumps count
// (CR 208.3), the source included. A mana ability reads it the same way.
func OldB324Formidable(g *game.Game, controller, _ uuid.UUID) bool {
	return CreaturesYouControlTotalPowerAtLeast(8)(g, controller)
}

// OldB324FormidableLabel is the printed text of a formidable ability: the
// ability word, the ability and its activation condition.
func OldB324FormidableLabel(ability string) string {
	return "Formidable — " + ability + " Activate only if creatures you control have total power 8 or greater."
}

// OldB324FormidableRow turns an ability row into its formidable form: the
// printed label and the power condition (CR 602.5b).
func OldB324FormidableRow(row ActivatedAbility, ability string) ActivatedAbility {
	row.Label = OldB324FormidableLabel(ability)
	row.Condition = OldB324Formidable
	return row
}

// OldB324FormidableAttack is "Formidable — Whenever this creature attacks, if
// creatures you control have total power 8 or greater, <effect>": the attack
// trigger with an intervening "if" (CR 603.4), checked as the creature attacks and
// again as the ability resolves.
func OldB324FormidableAttack(label string, effect Effect) game.TriggeredAbility {
	return InterveningIfYou("creatures you control have total power 8 or greater",
		CreaturesYouControlTotalPowerAtLeast(8), WheneverThisAttacks(label, effect))
}

// OldB324EachPlayersLifeBecomesTheirCreatures is "Each player's life total
// becomes the number of creatures they control": every player still in the game
// is counted first, then each life total is set (CR 118.5), so one player's
// change can't move another's count. A player whose life total can't change is
// left as it is by the engine's own gate.
func OldB324EachPlayersLifeBecomesTheirCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	counts := map[uuid.UUID]int{}
	var order []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		order = append(order, p.ID)
		counts[p.ID] = CountControlled(g, p.ID, func(c game.Card) bool { return c.IsCreature() })
	}
	for _, id := range order {
		if err := B31LifeBecomes(ctx, id, counts[id]); err != nil {
			return err
		}
	}
	return nil
}
