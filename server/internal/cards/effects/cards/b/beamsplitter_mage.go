package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beamsplitter Mage — Creature — Vedalken Wizard {U}{R}, 2/2 (Guilds of
// Ravnica, #5525):
//
//	"Whenever you cast an instant or sorcery spell that targets only this
//	 creature, if you control one or more other creatures that spell could
//	 target, choose one of those creatures. Copy that spell. The copy targets
//	 the chosen creature."
//
// Mirrorwing Dragon's reading of "targets only this creature" and of "could
// target" (game.SpellCouldTargetInsteadForEffect), for the Mage's own
// controller's spells. The "if" is an intervening one (CR 603.4): it is read as
// the spell is cast and again as the trigger resolves. One creature is chosen
// when there are two or more; the copy is of the spell as last known if it has
// already resolved (CR 707.10) and targets the chosen creature, with no new
// targets to choose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa09e18c-e7cf-4f08-9cc4-324e36594063",
		Name:         "Beamsplitter Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.MirrorwingSpellTargetsOnlyThis(ev, source, ch, g) &&
					len(beamsplitterCreatures(g, ev.CardID, ev.Actor, source.InstanceID)) > 0
			}, "Beamsplitter Mage — if you control other creatures that spell could target, choose one of them and copy that spell for it",
				beamsplitterCopy),
		},
	})
}

// beamsplitterCreatures are the other creatures `caster` controls that a copy
// of the spell could target in place of `mage`.
func beamsplitterCreatures(g *game.Game, spell, caster, mage uuid.UUID) []uuid.UUID {
	var creatures []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != caster || c.InstanceID == mage || !c.IsCreature() {
			continue
		}
		if g.SpellCouldTargetInsteadForEffect(spell, caster, mage, c.InstanceID) {
			creatures = append(creatures, c.InstanceID)
		}
	}
	return creatures
}

// beamsplitterCopy rechecks the "if" and makes the copy, asking which creature
// when there are two or more.
func beamsplitterCopy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell, caster, mage := item.Trigger.Event.CardID, item.Trigger.Event.Actor, item.SourceCardID
	creatures := beamsplitterCreatures(g, spell, caster, mage)
	switch len(creatures) {
	case 0:
		return nil
	case 1:
		g.CopySpellRetargetedForEffect(spell, caster, mage, creatures[0])
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  caster,
		Source:   mage,
		Question: "Beamsplitter Mage — choose a creature for the copy to target",
		Cards:    creatures,
		Min:      1,
		Max:      1,
		ThenKey:  beamsplitterCopyFor,
		Carry:    []uuid.UUID{spell, mage},
	})
	return nil
}

// beamsplitterCopyFor copies the spell targeting the chosen creature, if it is
// still a creature the spell could target. An on-disk key: never renamed,
// never reused.
var beamsplitterCopyFor = game.RegisterChooseCardsThen("choose-cards/beamsplitter-mage-copy",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Carry) < 2 || !r.Answered || len(r.Picked) != 1 {
			return nil
		}
		g.CopySpellRetargetedForEffect(r.Carry[0], r.Chooser, r.Carry[1], r.Picked[0])
		return nil
	})
