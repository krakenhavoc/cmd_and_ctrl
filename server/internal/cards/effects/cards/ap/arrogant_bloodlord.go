package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arrogant Bloodlord — Creature — Vampire Knight {1}{B}{B}, 4/4 (Rise of the
// Eldrazi, S82 OLD batch 484, #5581):
//
//	"Whenever this creature blocks or becomes blocked by a creature with
//	 power 1 or less, destroy this creature at end of combat."
//
// One block event names the blocker and the attacker (CR 509.3b, 509.3c): the
// Bloodlord's trigger fires when it is the blocker and the creature it blocks
// has power 1 or less, or when it is the attacker and a creature blocking it
// has power 1 or less, once for each such pair. The power is read as the
// block is declared. The destruction is Cinder Wall's delayed trigger at the
// beginning of end of combat.
//
// No simplification.
func init() {
	const label = "Arrogant Bloodlord — destroy this creature at end of combat"
	effects.Register(effects.Spec{
		OracleID:     "a35e1882-b878-49cf-a06c-be12a06d4977",
		Name:         "Arrogant Bloodlord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				other := uuid.Nil
				switch {
				case ev.CardID == source.InstanceID:
					other = ev.Target
				case ev.Target == source.InstanceID:
					other = ev.CardID
				}
				if other == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(other)
				return ok && c.IsCreature() && c.CurrentPower() <= 1
			}, label, effects.S82OldB364DestroyThisAtEndOfCombat(label)),
		},
	})
}
