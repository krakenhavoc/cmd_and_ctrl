package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrorwing Dragon — Creature — Dragon {3}{R}{R}, 4/5 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Flying
//	 Whenever a player casts an instant or sorcery spell that targets only
//	 this creature, that player copies that spell for each other creature
//	 they control that the spell could target. Each copy targets a
//	 different one of those creatures."
//
// "Targets only this creature" is every target of the spell, in each
// instance of the word "target", being the Dragon and no other object or
// player (the rulings). The copies are the CASTER's: one for each other
// creature that player controls, and none for anyone else's creatures,
// whoever controls the Dragon. A creature the spell couldn't target —
// hexproof from that player, protection, the clause's own restriction,
// or a target of a multi-target spell it fails — is ignored
// (game.SpellCouldTargetInsteadForEffect). The copies keep the
// original's mode, X and the costs paid for it (CR 707.10) and are not
// cast. The creatures are counted as the trigger resolves; when there
// are two or more, the caster puts the creatures in the order their
// copies go on the stack (the first goes on first, so it resolves last),
// and the original is beneath them all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ea14a03-c6a2-49ec-953d-4249a2c27e1d",
		Name:            "Mirrorwing Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, mirrorwingSpellTargetsOnlyThis,
				"Mirrorwing Dragon — that player copies the spell for each other creature they control it could target",
				mirrorwingCopies),
		},
	})
}

// mirrorwingSpellTargetsOnlyThis is "a player casts an instant or sorcery
// spell that targets only this creature".
func mirrorwingSpellTargetsOnlyThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	sole, ok := g.StackItemSoleTargetForEffect(ev.CardID)
	return ok && sole.Kind == game.TargetCard && sole.ID == source.InstanceID
}

// mirrorwingCopies finds the caster's other creatures the spell could
// target and copies the spell for each, asking the order when there is
// more than one.
func mirrorwingCopies(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell, caster, dragon := item.Trigger.Event.CardID, item.Trigger.Event.Actor, item.SourceCardID
	var creatures []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != caster || c.InstanceID == dragon || !c.IsCreature() {
			continue
		}
		if g.SpellCouldTargetInsteadForEffect(spell, caster, dragon, c.InstanceID) {
			creatures = append(creatures, c.InstanceID)
		}
	}
	switch len(creatures) {
	case 0:
		return nil
	case 1:
		g.CopySpellRetargetedForEffect(spell, caster, dragon, creatures[0])
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  caster,
		Source:   dragon,
		Question: "Mirrorwing Dragon — put your creatures in the order their copies go on the stack (the first goes on first and resolves last)",
		Cards:    creatures,
		Min:      len(creatures),
		Max:      len(creatures),
		ThenKey:  mirrorwingCopyInOrder,
		Carry:    []uuid.UUID{spell, dragon},
	})
	return nil
}

// mirrorwingCopyInOrder makes the copies in the order the caster put the
// creatures, keyed: the spell and the Dragon are the carry. A creature
// that is no longer a legal target by the answer gets no copy.
var mirrorwingCopyInOrder = game.RegisterChooseCardsThen("choose-cards/mirrorwing-dragon-copy-order", func(g *game.Game, r game.CardsChosen) error {
	// Unanswered, the caster has left the game, and so would the copies
	// they controlled (CR 800.4a).
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
