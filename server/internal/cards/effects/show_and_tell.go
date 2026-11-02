package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Show and Tell — Sorcery {2}{U}:
//
//	"Each player may put an artifact, creature, enchantment, or land card
//	 from their hand onto the battlefield."
//
// One pick per player, in APNAP order (CR 101.4: the active player chooses,
// then each other player in turn order, and the actions then happen
// simultaneously). A player with no such card in hand is not asked; "may"
// is a floor of zero. The picks, and the players still to ask, ride the
// prompts as data, so a table waiting on any seat is a restore point. The
// picks enter together as one entry (CR 603.6a), each under its owner's
// control, put by that player (CR 110.2a), so each newcomer sees the
// others arrive. A planeswalker or battle card is not offered: the card
// types the clause names are the only ones.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b83a3ba0-249e-4c39-bbf0-cb005413f7d2",
		Name:         "Show and Tell",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return showAndTellAskNext(ctx, apnapPlayers(ctx.Game), nil)
		},
	})
}

// showAndTellCandidates is the cards a player may put: an artifact,
// creature, enchantment or land card in their hand.
func showAndTellCandidates(g *game.Game, player uuid.UUID) []uuid.UUID {
	return handCardsMatching(g, player, Or(Artifact(), Creature(), Enchantment(), Land()))
}

// showAndTellAskNext asks the next player, in APNAP order, whose hand holds
// an eligible card; once nobody is left it puts every pick onto the
// battlefield together.
func showAndTellAskNext(ctx *Context, order, picks []uuid.UUID) error {
	g, item := ctx.Game, ctx.Item
	for i, player := range order {
		candidates := showAndTellCandidates(g, player)
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  player,
			Source:   item.SourceCardID,
			Question: "Show and Tell — you may put an artifact, creature, enchantment, or land card from your hand onto the battlefield",
			Cards:    candidates,
			Min:      0,
			Max:      1,
			ThenKey:  showAndTellPick,
			LiveZone: game.ZoneHand,
			Carry:    packIDLists(order[i+1:], picks),
			Asked:    game.NewAskedItem(item),
		})
		return nil
	}
	entries := make([]game.BatchEntry, 0, len(picks))
	for _, id := range picks {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneHand})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{PutByEnteringPlayer: true}, game.KeyedTail{})
}

// showAndTellPick is one player's pick answered: the next player is asked,
// or the picks enter together. Registered in init because it asks the next
// player, which names the key again.
var showAndTellPick game.ChooseCardsThen

func init() {
	showAndTellPick = askedPick("choose-cards/show-and-tell-pick", func(ctx *Context, r game.CardsChosen) error {
		lists := unpackIDLists(r.Carry)
		for len(lists) < 2 {
			lists = append(lists, nil)
		}
		rest, picks := lists[0], lists[1]
		return showAndTellAskNext(ctx, rest, append(append([]uuid.UUID(nil), picks...), r.Picked...))
	})
}
