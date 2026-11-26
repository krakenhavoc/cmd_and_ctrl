package cardsco

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conundrum Sphinx — Creature — Sphinx {2}{U}{U}, 4/4:
//
//	"Flying
//	 Whenever this creature attacks, each player chooses a card name. Then
//	 each player reveals the top card of their library. If the card a player
//	 revealed has the name they chose, that player puts it into their hand.
//	 If it doesn't, that player puts it on the bottom of their library."
//
// Each player, the attacker first and then in turn order, is asked for a name
// by a keyed card-name prompt (Demonic Consultation's). The prompts are
// private to their chooser and the names ride the tail as data (KeyedTail
// Words), so nobody sees another player's name before every name is in and
// the reveals begin (CR 608.2c). A name is free text (CR 201.2); a split or
// double-faced card matches under any face (CardNameMatches). Each library
// is then revealed at the top and the card goes to its owner's hand when it
// matches, else to the bottom of its library; an empty library does nothing.
// A player who has left the game is not asked.
//
// No simplification.
func init() {
	conundrumNamed = game.RegisterTail("conundrum-sphinx/named", conundrumNamedBody)
	effects.Register(effects.Spec{
		OracleID:        "fb1bd39c-c870-4f63-9e4f-32c94ac46555",
		Name:            "Conundrum Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Conundrum Sphinx — each player chooses a card name, then reveals the top card of their library",
				func(g *game.Game, item *game.StackItem) error {
					players := conundrumPlayers(g, item.Controller)
					return conundrumAsk(g, item.SourceCardID, players, nil)
				}),
		},
	})
}

// conundrumPlayers is the players still in the game, starting with the
// controller and going around the table (CR 101.4 order).
func conundrumPlayers(g *game.Game, first uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			ids = append(ids, p.ID)
		}
	}
	i := slices.Index(ids, first)
	if i < 0 {
		return ids
	}
	return append(slices.Clone(ids[i:]), ids[:i]...)
}

// conundrumAsk asks the next player who has not named a card, or runs the
// reveals once every name is in.
func conundrumAsk(g *game.Game, source uuid.UUID, players []uuid.UUID, names []string) error {
	if len(players) == 0 {
		return nil
	}
	if len(names) >= len(players) {
		return conundrumReveal(g, source, players, names)
	}
	g.QueueCardNameChoiceKeyedForEffect(players[len(names)], source,
		"Conundrum Sphinx — choose a card name",
		game.KeyedTail{Ref: conundrumNamed, Player: players[0], Source: source, Carry: players, Words: names})
	return nil
}

// conundrumNamed is registered in init: it asks for the next name, which
// refers back to it.
var conundrumNamed game.TailRef

func conundrumNamedBody(g *game.Game, o game.TailOutcome) error {
	names := append(slices.Clone(o.Tail.Words), o.Tail.Word)
	return conundrumAsk(g, o.Tail.Source, o.Tail.Carry, names)
}

// conundrumReveal reveals each player's top card in turn and sends it to the
// hand when it has the name that player chose, else to the bottom.
func conundrumReveal(g *game.Game, source uuid.UUID, players []uuid.UUID, names []string) error {
	for i, player := range players {
		name := names[i]
		revealed := g.RevealTopOfLibraryForEffect(player, source, 1, "Conundrum Sphinx — the top card of the library")
		if len(revealed) == 0 {
			continue
		}
		if err := (effects.TakeFromLibraryToHand{
			Player: player,
			Cards:  revealed,
			Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.CardNameMatches(c, name) },
			All:    true,
			Label:  "Conundrum Sphinx — it has the chosen name, so it goes into hand",
			Key:    conundrumTake,
		}).Apply(effects.KeyedContext(g, player, source)); err != nil {
			return err
		}
	}
	return nil
}

var conundrumTake = effects.RegisterTake("choose-cards/conundrum-sphinx-take", effects.TakeDef{Then: effects.TakeRestOnBottomInRandomOrder})
