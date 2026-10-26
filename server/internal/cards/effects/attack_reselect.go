package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_reselect.go — the card side of CR 508.7, "reselect which
// player or permanent that creature is attacking" (#1329, ADR 0045
// amendment 2026-09-23). The engine owns the rule and the prompt
// (game/attack_reselect.go); this file owns the two printed trigger
// shapes that reach it.
//
// Every printed card of the family has the same trigger condition —
// "when this enters DURING THE DECLARE ATTACKERS STEP" — and a flash
// permanent carrying it is the whole trick: cast it after attackers
// are declared and before blockers, and an attack aimed at you lands
// on somebody else. Cast at any other time the permanent enters with
// no trigger at all (the condition is a trigger condition, read as
// the event happens, not an intervening "if").
//
// Whose attacker it is does not matter — a defending player flashing
// in Misleading Signpost redirects the ACTIVE player's creature, and
// CR 508.7c checks the new target against that creature's
// controller, not against the Signpost's. The prompt goes to the
// trigger's controller either way.

// entersDuringDeclareAttackers is the shared trigger condition: this
// permanent entered, and the turn is in the declare attackers step.
func entersDuringDeclareAttackers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID && g.Turn.Step == game.StepDeclareAttackers
}

// reselectTargetAttackerTrigger is "When this enters during the
// declare attackers step, you may reselect which player or permanent
// target attacking creature is attacking" — Misleading Signpost and
// Portal Mage, word for word.
//
// Targeted, so a table with no attacking creature puts no trigger on
// the stack at all (CR 603.3d), and one whose target has left combat
// by resolution fizzles (CR 608.2b). The "you may" is the prompt's
// first option, "Keep attacking …", asked at resolution where the
// card prints it rather than as a yes/no when the trigger fires.
func reselectTargetAttackerTrigger(label string) game.TriggeredAbility {
	t := On(game.EventETB, entersDuringDeclareAttackers, label, reselectTargetedAttacker)
	t.Targets = TargetCreature("target attacking creature", AttackingCreature())
	return t
}

// reselectTargetedAttacker is reselectTargetAttackerTrigger's effect.
// A package-level func so the item captures nothing (docs/adding-cards.md).
func reselectTargetedAttacker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		g.QueueReselectAttackForEffect(game.ReselectAttackPrompt{
			Chooser:  item.Controller,
			Attacker: t.ID,
			Source:   item.SourceCardID,
			Question: g.ReselectAttackQuestion(t.ID),
		})
		return nil
	}
	return nil
}

// reselectEachAttackerTrigger is Windshaper Planetar's untargeted
// sibling: "for each attacking creature, you may reselect which player
// or permanent that creature is attacking". One question per attacking
// creature, asked one after another in battlefield order — each is its
// own "you may", so declining one does not decline the rest.
//
// The set is the attacking creatures AS IT RESOLVES; a creature that
// leaves combat while an earlier question is open is skipped when its
// turn comes, because the engine's prompt builder finds it no longer
// attacking and asks nothing.
func reselectEachAttackerTrigger(label string) game.TriggeredAbility {
	return On(game.EventETB, entersDuringDeclareAttackers, label, reselectEachAttacker)
}

func reselectEachAttacker(g *game.Game, item *game.StackItem) error {
	var attackers []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil {
			attackers = append(attackers, c.InstanceID)
		}
	}
	if len(attackers) == 0 {
		return nil
	}
	// Keyed (#4253 slice 10b): the attackers still to ask about ride the
	// prompt as data, so each question is a restore point.
	g.QueueReselectAttackForEffect(game.ReselectAttackPrompt{
		Chooser:  item.Controller,
		Attacker: attackers[0],
		Source:   item.SourceCardID,
		Question: g.ReselectAttackQuestion(attackers[0]),
		Next:     attackers[1:],
	})
	return nil
}
