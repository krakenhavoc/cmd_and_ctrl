package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tuvok, Chief Tactical Officer — {2}{U} Legendary Creature — Vulcan
// Officer 2/3 (Star Trek Commander, #4974):
//
//	"Flash
//	 When Tuvok enters during the declare attackers step, choose up to
//	 two target attacking creatures. For each of them, you may reselect
//	 which player or permanent that creature is attacking. (A creature
//	 can't attack its controller.)"
//
// Portal Mage's trigger with "up to two" targets (CR 508.7): the trigger
// condition is read as Tuvok enters, so outside the declare attackers
// step there is no trigger at all. Each still-legal target gets its own
// "you may" question, one after the other; a target that left combat
// while an earlier question was open is skipped when its turn comes.
//
// No simplification.
func init() {
	t := On(game.EventETB, entersDuringDeclareAttackers,
		"Tuvok, Chief Tactical Officer — reselect what up to two target attacking creatures are attacking",
		tuvokReselectTargets)
	t.Targets = TargetCreature("up to two target attacking creatures", AttackingCreature()).WithCount(0, 2)
	Register(Spec{
		OracleID:        "05c5f17e-40f9-4259-9b03-03683e1c5ba6",
		Name:            "Tuvok, Chief Tactical Officer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

// tuvokReselectTargets asks one reselect question per still-legal
// target, in target order.
func tuvokReselectTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var attackers []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			attackers = append(attackers, t.ID)
		}
	}
	if len(attackers) == 0 {
		return nil
	}
	g.QueueReselectAttackForEffect(game.ReselectAttackPrompt{
		Chooser:  item.Controller,
		Attacker: attackers[0],
		Source:   item.SourceCardID,
		Question: g.ReselectAttackQuestion(attackers[0]),
		Next:     attackers[1:],
	})
	return nil
}
