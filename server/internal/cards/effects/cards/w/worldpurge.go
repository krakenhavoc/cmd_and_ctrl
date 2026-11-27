package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worldpurge — Sorcery {4}{W/U}{W/U}{W/U}{W/U} (Shadowmoor, S82 older sets):
//
//	"Return all permanents to their owners' hands. Each player chooses up to
//	 seven cards in their hand, then shuffles the rest into their library. Each
//	 player loses all unspent mana."
//
// Every permanent is returned in one simultaneous event (BounceAllMatching, a
// token ceasing to exist as it leaves), and the rest of the spell runs from that
// event's keyed tail so a commander's replacement prompt settles first. Then
// each player, in turn order from the active player, with more than seven cards
// in hand is asked to choose seven of them to keep (a hand of seven or fewer
// keeps everything: "up to seven" and the rest is nothing); the cards not kept
// are tucked into the owner's library and the library is shuffled. Finally each
// player loses all unspent mana (CR 106.4), the caster's leftover included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "285ecbd5-ff7c-4919-ab0b-7fc05e5a0a0a",
		Name:         "Worldpurge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepBounce}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{
				Match:    func(*game.Game, uuid.UUID, game.Card) bool { return true },
				ThenTail: ctx.ItemTail(worldpurgeAfterTheBounce),
			}.Apply(ctx)
		},
	})
}

// worldpurgeKeep is how many cards of a hand a player keeps.
const worldpurgeKeep = 7

var worldpurgeAfterTheBounce = effects.MassTail("then/worldpurge-after-the-bounce", func(ctx *effects.Context, _ []game.Card, _ int) error {
	g := ctx.Game
	// A token that went to a hand is not a card and is gone as soon as the rules
	// look (CR 111.7): it is neither chosen nor shuffled in (The Great Aurora).
	g.TokensCeaseToExistForEffect()
	for _, id := range effects.SeatsFromActive(g) {
		hand := effects.AllHandCardIDs(g, id)
		if len(hand) > worldpurgeKeep {
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    id,
				FromPlayer: id,
				Source:     ctx.Source(),
				Question:   "Worldpurge — choose up to seven cards in your hand; shuffle the rest into your library",
				Cards:      hand,
				Min:        worldpurgeKeep,
				Max:        worldpurgeKeep,
				ThenKey:    worldpurgeKept,
				LiveZone:   game.ZoneHand,
			})
		}
	}
	for _, id := range effects.SeatsFromActive(g) {
		g.LoseAllUnspentManaForEffect(id)
	}
	return nil
})

var worldpurgeKept = game.RegisterChooseCardsThen("choose-cards/worldpurge-kept", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered {
		return nil
	}
	kept := make(map[uuid.UUID]bool, len(r.Picked))
	for _, id := range r.Picked {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, id := range r.Candidates {
		if !kept[id] {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect(rest, game.TuckOptions{}, game.KeyedTail{Ref: worldpurgeShuffle, Player: r.Chooser})
})

var worldpurgeShuffle = game.RegisterTail("then/worldpurge-shuffle", func(g *game.Game, o game.TailOutcome) error {
	return g.ShuffleLibraryForEffect(o.Tail.Player)
})
