package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b695_helpers.go — the Mercadian Masques enchantments (#5619, S82).

// Mmq695SharesNameWithANontokenPermanent is the core of Cornered Market:
// the card (a spell being cast, or a land being played) has the same name
// as a nontoken permanent on the battlefield, any controller's. A name is
// matched on any face (CR 201.2b); a face-down permanent has no name
// (CR 708.2a) and matches nothing.
func Mmq695SharesNameWithANontokenPermanent(g *game.Game, card game.Card) bool {
	if g == nil {
		return false
	}
	for _, perm := range g.BattlefieldCardsForEffect() {
		if IsToken(perm) || perm.FaceDown {
			continue
		}
		if game.CardNameMatches(card, perm.Name) {
			return true
		}
	}
	return false
}

// Mmq695RedOrGreenCreatureSpell is High Seas' "red creature spells and green
// creature spells": a creature spell that is red, or green, or both. The
// colours are words of the source's text (CR 612.2).
func Mmq695RedOrGreenCreatureSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.IsCreature() && (q.Card.HasColor(q.Source.TextColor("R")) || q.Card.HasColor(q.Source.TextColor("G")))
	}
}

// Mmq695SorcerySpell is a sorcery spell (Squeeze).
func Mmq695SorcerySpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsSorcery() }
}

// TheCastersDiscardsACard is the body of "Whenever a player casts a spell,
// that player discards a card" (Oppression) and its colour-narrowed kin
// (Putrefaction): the caster, named by the cast event, chooses a card from
// their own hand to discard; with an empty hand they discard nothing.
func TheCastersDiscardsACard(g *game.Game, item *game.StackItem) error {
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: item.Trigger.Event.Actor,
		Source: item.SourceCardID,
		N:      1,
	})
	return nil
}
