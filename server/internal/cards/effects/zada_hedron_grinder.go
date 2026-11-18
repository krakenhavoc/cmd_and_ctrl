package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zada, Hedron Grinder — {3}{R} Legendary Creature — Goblin Ally, 3/3:
//
//	"Whenever you cast an instant or sorcery spell that targets only Zada,
//	 copy that spell for each other creature you control that the spell
//	 could target. Each copy targets a different one of those creatures."
//
// Mirrorwing Dragon's trigger for Zada's controller's own casts alone:
// "targets only Zada" is every target of the spell being Zada and nothing
// else (the rulings); the copies are the controller's, one for each other
// creature they control that the spell could target (a hexproof or
// protected creature, or one the spell's own restriction refuses, is
// ignored, game.SpellCouldTargetInsteadForEffect). The copies keep the
// original's mode, X and costs paid (CR 707.10) and are not cast. The
// creatures are counted as the trigger resolves; with two or more the
// controller orders their copies, as the Dragon's controller does, since
// the copies are put onto the stack in an order of their choosing (CR
// 603.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f7d5e1c-bf5b-4ab9-bd6e-2627b4d22b36",
		Name:         "Zada, Hedron Grinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && mirrorwingSpellTargetsOnlyThis(ev, source, ch, g)
			}, "Zada, Hedron Grinder — copy that spell for each other creature you control it could target",
				zadaCopies),
		},
	})
}

// zadaCopies finds the controller's other creatures the spell could target
// and copies the spell for each, asking the order when there is more than one.
func zadaCopies(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell, caster, zada := item.Trigger.Event.CardID, item.Controller, item.SourceCardID
	var creatures []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != caster || c.InstanceID == zada || !c.IsCreature() {
			continue
		}
		if g.SpellCouldTargetInsteadForEffect(spell, caster, zada, c.InstanceID) {
			creatures = append(creatures, c.InstanceID)
		}
	}
	switch len(creatures) {
	case 0:
		return nil
	case 1:
		g.CopySpellRetargetedForEffect(spell, caster, zada, creatures[0])
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  caster,
		Source:   zada,
		Question: "Zada, Hedron Grinder — put your creatures in the order their copies go on the stack (the first goes on first and resolves last)",
		Cards:    creatures,
		Min:      len(creatures),
		Max:      len(creatures),
		ThenKey:  zadaCopyInOrder,
		Carry:    []uuid.UUID{spell, zada},
	})
	return nil
}

// zadaCopyInOrder makes the copies in the order the controller put the
// creatures, keyed: the spell and Zada are the carry. A creature that is
// no longer a legal target by the answer gets no copy. An on-disk key.
var zadaCopyInOrder = game.RegisterChooseCardsThen("choose-cards/zada-hedron-grinder-copy-order", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) < 2 || !r.Answered {
		return nil
	}
	order := r.Picked
	if len(order) != len(r.Candidates) {
		order = r.Candidates
	}
	for _, id := range order {
		g.CopySpellRetargetedForEffect(r.Carry[0], r.Chooser, r.Carry[1], id)
	}
	return nil
})
