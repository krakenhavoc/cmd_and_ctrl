package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b279_helpers.go — shared shapes of S82 OLD batch 279 (Eternal
// Masters, #5542): token rows and the "reveal a hand and discard every card
// with the named name" step.

// The token rows are added from this file's package-level initializer so the
// shared table's column alignment is not rewritten and the rows exist before
// any card file's init() asks for one. Each matches its printed token card.
var _ = registerOldB279Tokens()

func registerOldB279Tokens() bool {
	for key, card := range map[string]game.Card{
		// Tooth and Claw.
		"3/1 red Beast named Carnivore": {Name: "Carnivore", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 1, Colors: []string{"R"}},
		// Goblin Trenches.
		"1/1 red and white Goblin Soldier": {Name: "Goblin Soldier", TypeLine: "Token Creature — Goblin Soldier", Power: 1, Toughness: 1, Colors: []string{"R", "W"}},
		// Tidal Wave.
		"5/5 blue Wall with defender": {Name: "Wall", TypeLine: "Token Creature — Wall", Power: 5, Toughness: 5, Colors: []string{"U"}, Keywords: []string{"defender"}},
	} {
		if _, dup := TokenTable[key]; dup {
			panic("s82_old_b279_helpers: token key already in the table: " + key)
		}
		TokenTable[key] = card
	}
	return true
}

// RevealHandAndDiscardAllNonlandNamed is Cabal Therapy's second sentence once
// the name is chosen: `victim` reveals their hand and discards every nonland
// card with that name (CR 201.2: every face is checked), all at once and not
// by their choice. A land named does nothing, since "a nonland card name" can
// never be one.
func RevealHandAndDiscardAllNonlandNamed(ctx *Context, victim uuid.UUID, name string) error {
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	var named []uuid.UUID
	for _, c := range p.Hand.Cards {
		if !c.IsLand() && game.CardNameMatches(c, name) {
			named = append(named, c.InstanceID)
		}
	}
	if len(named) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(victim, named, ctx.Source(), game.KeyedTail{})
}
