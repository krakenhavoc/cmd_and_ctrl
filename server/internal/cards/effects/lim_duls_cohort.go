package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lim-Dûl's Cohort — Creature — Zombie {1}{B}{B}, 2/3:
//
//	"Whenever this creature blocks or becomes blocked by a creature,
//	 that creature can't be regenerated this turn."
//
// The trigger is per creature (CR 509.3b, 509.3d): one block event per
// blocker–attacker pair, Giant Shark's shape, so the Cohort blocked by
// two creatures triggers twice, once for each. "That creature" is the
// other half of the pair the trigger saw, marked with ADR 0108 §2's
// turn-long CR 701.19c mark as the trigger resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1e3b97d2-8fda-4510-9697-f36ca9ca2ab0",
		Name:         "Lim-Dûl's Cohort",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.InstanceID == ev.CardID || source.InstanceID == ev.Target
			}, "Lim-Dûl's Cohort — that creature can't be regenerated this turn", limDulsCohortMark),
		},
	})
}

// limDulsCohortMark marks the creature on the other side of the block
// the trigger saw.
func limDulsCohortMark(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	other := ev.CardID
	if ev.CardID == item.SourceCardID {
		other = ev.Target
	}
	return CantBeRegeneratedThisTurn{Target: other}.Apply(NewContext(g, item))
}
