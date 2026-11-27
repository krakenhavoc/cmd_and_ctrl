package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ink-Treader Nephilim — Creature — Nephilim {R}{G}{W}{U}, 3/3 (Guildpact):
//
//	"Whenever a player casts an instant or sorcery spell, if that spell
//	 targets only this creature, copy the spell for each other creature
//	 that spell could target. Each copy targets a different one of those
//	 creatures."
//
// Precursor Golem's copy-per-creature shape, with three differences: the
// spell may be anyone's, it must target only the Nephilim (an intervening
// "if", CR 603.4, read as the spell is cast and again as the ability
// resolves), and the copies belong to the Nephilim's controller, who puts
// them on the stack (CR 707.10), not to the caster. "Could target" is the
// spell's own targeting restrictions (and the caster's: hexproof, protection),
// so a creature the spell can't legally name gets no copy. The controller
// orders the copies when more than one is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3846b499-59c8-4e77-b46a-1d697bcbded7",
		Name:         "Ink-Treader Nephilim",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, inkTreaderSpellTargetsOnlyThis,
				"Ink-Treader Nephilim — copy the spell for each other creature it could target",
				inkTreaderCopies),
		},
	})
}

// inkTreaderSpellTargetsOnlyThis is the trigger's condition and its
// intervening "if": an instant or sorcery spell whose only target is the
// Nephilim.
func inkTreaderSpellTargetsOnlyThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || source == nil {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	sole, ok := g.StackItemSoleTargetForEffect(ev.CardID)
	return ok && sole.Kind == game.TargetCard && sole.ID == source.InstanceID
}

func inkTreaderCopies(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell := item.Trigger.Event.CardID
	sole, ok := g.StackItemSoleTargetForEffect(spell)
	if !ok || sole.Kind != game.TargetCard || sole.ID != item.SourceCardID {
		return nil // the intervening "if" no longer holds (CR 603.4)
	}
	controller := item.Controller
	var others []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == sole.ID || !c.IsCreature() {
			continue
		}
		if g.SpellCouldTargetInsteadForEffect(spell, controller, sole.ID, c.InstanceID) {
			others = append(others, c.InstanceID)
		}
	}
	switch len(others) {
	case 0:
		return nil
	case 1:
		g.CopySpellRetargetedForEffect(spell, controller, sole.ID, others[0])
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "Ink-Treader Nephilim — put the creatures in the order their copies go on the stack (the first goes on first and resolves last)",
		Cards:    others,
		Min:      len(others),
		Max:      len(others),
		ThenKey:  inkTreaderCopyInOrder,
		Carry:    []uuid.UUID{spell, sole.ID},
	})
	return nil
}

var inkTreaderCopyInOrder = game.RegisterChooseCardsThen("choose-cards/ink-treader-nephilim-copy-order", effects.CopySpellForEachChosenInOrder)
