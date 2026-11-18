package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Precursor Golem — Artifact Creature — Golem {5}, 3/3:
//
//	"When this creature enters, create two 3/3 colorless Golem artifact
//	 creature tokens.
//	 Whenever a player casts an instant or sorcery spell that targets
//	 only a single Golem, that player copies that spell for each other
//	 Golem that spell could target. Each copy targets a different one of
//	 those Golems."
//
// Mirrorwing Dragon's copy-per-creature shape over Golems of any
// controller (CR 115.9c for "only a single"; the caster controls the
// copies, CR 707.10). The Golem it was cast on is the one each copy is
// moved off; "could target" is the spell's own targeting restrictions, so
// a hexproof Golem or one the spell can't legally name gets no copy. The
// caster orders the copies when more than one is made. "Golem" is a word
// of the source's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "668fc293-1b8f-4870-9784-4ac660925e30",
		Name:         "Precursor Golem",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Precursor Golem — create two 3/3 colorless Golem artifact creature tokens",
				Do(CreateToken{Template: TokenCard("3/3 colorless Golem artifact creature"), N: 2})),
			On(game.EventCast, precursorGolemSpellTargetsOnlyAGolem,
				"Precursor Golem — that player copies the spell for each other Golem it could target",
				precursorGolemCopies),
		},
	})
}

func precursorGolemSpellTargetsOnlyAGolem(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	sole, ok := g.StackItemSoleTargetForEffect(ev.CardID)
	if !ok || sole.Kind != game.TargetCard {
		return false
	}
	target, ok := g.LookupCardForEffect(sole.ID)
	return ok && onBattlefield(g, sole.ID) && target.HasSubtype(source.TextSubtype("Golem"))
}

func precursorGolemCopies(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell, caster := item.Trigger.Event.CardID, item.Trigger.Event.Actor
	sole, ok := g.StackItemSoleTargetForEffect(spell)
	if !ok || sole.Kind != game.TargetCard {
		return nil
	}
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		src = game.Card{}
	}
	var golems []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == sole.ID || !c.HasSubtype(src.TextSubtype("Golem")) {
			continue
		}
		if g.SpellCouldTargetInsteadForEffect(spell, caster, sole.ID, c.InstanceID) {
			golems = append(golems, c.InstanceID)
		}
	}
	switch len(golems) {
	case 0:
		return nil
	case 1:
		g.CopySpellRetargetedForEffect(spell, caster, sole.ID, golems[0])
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  caster,
		Source:   item.SourceCardID,
		Question: "Precursor Golem — put the Golems in the order their copies go on the stack (the first goes on first and resolves last)",
		Cards:    golems,
		Min:      len(golems),
		Max:      len(golems),
		ThenKey:  precursorGolemCopyInOrder,
		Carry:    []uuid.UUID{spell, sole.ID},
	})
	return nil
}

var precursorGolemCopyInOrder = game.RegisterChooseCardsThen("choose-cards/precursor-golem-copy-order", func(g *game.Game, r game.CardsChosen) error {
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
