package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b35_helpers.go — shared helpers for Standard-set leftover batches 35 to
// 38 (#4452).

// b35GolemToken is the 3/3 colorless Golem artifact creature token the Big Score
// prints (Sandstorm Salvager, Legion Extruder). The tokens table's "3/3
// colorless Golem artifact" row is an enchantment Golem, so it is not reused.
func b35GolemToken() game.Card {
	return game.Card{Name: "Golem", TypeLine: "Token Artifact Creature — Golem", Power: 3, Toughness: 3}
}

// b35CreatureTokensYouControl lists the creature tokens `controller` controls.
func b35CreatureTokensYouControl(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.IsToken() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// b35IDSet is a CardPredicate over a fixed set of instance ids.
func b35IDSet(ids []uuid.UUID) CardPredicate {
	set := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return set[c.InstanceID] }
}

// b35YouDiscardLoseSacrifice is "you discard `discard` card(s), lose
// `life` life, and sacrifice `sac` creature(s)" (Greed's Gambit), in printed
// order: the discard prompt is queued first, the life is lost, then the
// sacrifice prompt is queued behind it.
func b35YouDiscardLoseSacrifice(discard, life, sac int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		p := item.Controller
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p, Source: item.SourceCardID, N: discard, Question: label + " — discard"})
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, p, -life); err != nil {
			return err
		}
		g.PlayerSacrificesNForEffect(item.SourceCardID, p, sacrificeSpec("a creature", Creature()), label+" — sacrifice", sac)
		return nil
	}
}

// b35BatToken is the 2/1 black Bat creature token with flying (Greed's Gambit).
func b35BatToken() game.Card {
	return game.Card{Name: "Bat", TypeLine: "Token Creature — Bat", Power: 2, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}}
}
