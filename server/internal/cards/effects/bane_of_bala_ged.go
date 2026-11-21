package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bane of Bala Ged — Creature — Eldrazi {7}, 7/5 (Commander Masters):
//
//	"Whenever this creature attacks, defending player exiles two permanents
//	 they control."
//
// Nothing is targeted: as the trigger resolves, the defending player (read
// off what the creature attacked, CR 508.5) chooses two permanents they
// control and exiles them. With only one permanent they exile that one;
// with none, nothing happens. The choice is theirs, not the attacker's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38839894-1706-4e11-8310-5ea8dd8866d9",
		Name:         "Bane of Bala Ged",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked, "Bane of Bala Ged — defending player exiles two permanents they control", baneOfBalaGedExile),
		},
	})
}

func baneOfBalaGedExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
	if defender == uuid.Nil {
		return nil
	}
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == defender {
			mine = append(mine, c.InstanceID)
		}
	}
	n := min(2, len(mine))
	if n == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    defender,
		FromPlayer: defender,
		Source:     item.SourceCardID,
		Question:   "Bane of Bala Ged — exile two permanents you control",
		Cards:      mine,
		Min:        n,
		Max:        n,
		ThenKey:    gapB98ExileChosen,
		LiveZone:   game.ZoneBattlefield,
	})
	return nil
}
