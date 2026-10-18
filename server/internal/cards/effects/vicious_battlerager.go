package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vicious Battlerager — Creature — Dwarf Barbarian {3}{B}, 1/5:
//
//	"When this creature enters, you take the initiative.
//	 Spiked Retribution — Whenever this creature becomes blocked by a
//	 creature, that creature's controller loses 5 life."
//
// The ETB is TakeTheInitiative (ADR 0203, #4096). Spiked Retribution
// triggers once per blocker (CR 509.3d): EventBlock is one event per
// blocker–attacker pair, Lim-Dûl's Cohort's shape, so two blockers are
// two triggers and ten life. "That creature's controller" is read as
// the trigger resolves, from the blocking object — the permanent as it
// is now, or as it last existed on the battlefield if it has left (CR
// 608.2h) — so a blocker that died to first strike still costs its
// controller the life. A creature that blocks and is then a new object
// is not the one the trigger names.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8942c72-16ce-4f2f-875f-97a56e85a03e",
		Name:         "Vicious Battlerager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Vicious Battlerager"),
			On(game.EventBlock, viciousBattleragerBlocked,
				"Vicious Battlerager — Spiked Retribution: the blocking creature's controller loses 5 life", viciousBattleragerRetribution),
		},
	})
}

// viciousBattleragerBlocked: this creature became blocked — EventBlock
// carries the blocker in CardID and the attacker in Target.
func viciousBattleragerBlocked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.Kind == game.EventBlock && ev.CardID != source.InstanceID && ev.Target == source.InstanceID
}

// viciousBattleragerRetribution costs the blocking creature's
// controller 5 life.
func viciousBattleragerRetribution(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	o := item.Trigger.Object
	info, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch})
	if !ok || info.Controller == uuid.Nil {
		return nil
	}
	return GainLife{Player: info.Controller, Amount: -5}.Apply(NewContext(g, item))
}
