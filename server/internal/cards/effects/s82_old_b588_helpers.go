package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b588_helpers.go — Saviors of Kamigawa, batch 588 (#5601): the
// creatures whose bonuses depend on the size of their controller's hand.

// OldB588YouHaveAtLeastCardsInHand is "you have N or more cards in hand".
func OldB588YouHaveAtLeastCardsInHand(n int) func(g *game.Game, you uuid.UUID) bool {
	return YouHaveCardsInHandAtLeast(n)
}

// OldB588YouHaveMoreCardsInHandThanEachOpponent is "you have more cards in hand
// than each opponent": strictly more than the largest hand among the opponents
// still in the game (so with no opponent left it holds).
func OldB588YouHaveMoreCardsInHandThanEachOpponent(g *game.Game, you uuid.UUID) bool {
	me := g.PlayerByIDForEffect(you)
	if me == nil || me.Hand == nil {
		return false
	}
	largest := -1
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you || p.Hand == nil {
			continue
		}
		if n := len(p.Hand.Cards); n > largest {
			largest = n
		}
	}
	return len(me.Hand.Cards) > largest
}

// OldB588WhileYouHavePump is "As long as <condition about your hand>, this
// creature gets +p/+t": a layer 7c modify on this creature alone, which
// declares DependsOnHandSize so the layers are rebuilt as any hand changes.
func OldB588WhileYouHavePump(cond func(g *game.Game, you uuid.UUID) bool, power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:             label,
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7C_Modify,
		DependsOnHandSize: true,
		AppliesTo:         WhileYouHave(cond, SelfCreature),
		Apply:             PumpBy(power, toughness),
	}
}

// OldB588WhileYouHaveKeyword is "As long as <condition about your hand>, this
// creature has <keyword>": a layer 6 grant on this creature alone, with the same
// hand dependency.
func OldB588WhileYouHaveKeyword(cond func(g *game.Game, you uuid.UUID) bool, keyword, label string) game.StaticAbility {
	ab := KeywordGrant(WhileYouHave(cond, SelfCreature), keyword)
	ab.Label = label
	ab.DependsOnHandSize = true
	return ab
}

// OldB588ThisGetsPlusZeroPlusHandSize is Kitsune Loreweaver's effect: "This
// creature gets +0/+X until end of turn, where X is the number of cards in your
// hand", X counted once, as the ability resolves (CR 608.2h).
func OldB588ThisGetsPlusZeroPlusHandSize(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		n := 0
		if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
			n = len(p.Hand.Cards)
		}
		return BoostUntilEOT{Target: item.SourceCardID, Toughness: n, Label: label}.Apply(NewContext(g, item))
	}
}
