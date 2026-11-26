package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Great Aurora — Sorcery {6}{G}{G}{G} (Magic Origins):
//
//	"Each player shuffles all cards from their hand and all permanents
//	 they own into their library, then draws that many cards. Each player
//	 may put any number of land cards from their hand onto the battlefield.
//	 Exile The Great Aurora."
//
// Three stages joined by keyed tails, as Warp World's first one is. The
// spell exiles itself first (as Ill-Gotten Gains does), so it is on no
// player's graveyard. Then every card in a hand and every permanent, by its
// OWNER, goes into that owner's library as one simultaneous move (a token
// goes there, is counted as a permanent shuffled in and then ceases to exist
// before any card is drawn, CR 111.7, CR 704.5d; a commander its owner sends to the command zone instead is
// not shuffled in and does not count). Each player with a count shuffles
// and draws that many. Last each player, asked in turn order, may pick any
// number of the land cards now in their hand; the picks enter together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9d7341e-10e7-43f9-bee2-e8053b287acd",
		Name:         "The Great Aurora",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			if err := (effects.ExileTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
				return err
			}
			var all []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				p := g.PlayerByIDForEffect(c.Owner)
				if p == nil || p.Eliminated {
					continue
				}
				all = append(all, c.InstanceID)
			}
			for _, id := range effects.TablePlayers(ctx) {
				all = append(all, effects.AllHandCardIDs(g, id)...)
			}
			return g.TuckCardsToLibraryKeyedForEffect(all, game.TuckOptions{},
				ctx.ItemTail(greatAuroraShuffled))
		},
	})
}

// greatAuroraShuffled is the shuffle done: count what each player put in,
// shuffle and draw that many, then ask for lands. An on-disk key.
var greatAuroraShuffled = effects.CardTail("tail/the-great-aurora-shuffled", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	// A token that reached a library ceases to exist before anything is
	// drawn (CR 111.7, CR 704.5d), but it is counted: it was a permanent
	// its owner shuffled in. The tuck's outcome lists it with the cards, so
	// each token counts once, by way of that list.
	counts := map[uuid.UUID]int{}
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok {
			counts[c.Owner]++
		}
	}
	g.TokensCeaseToExistForEffect()
	players := effects.ApnapPlayers(g)
	for _, p := range players {
		n := counts[p]
		if n == 0 {
			continue
		}
		if err := g.ShuffleLibraryForEffect(p); err != nil {
			return err
		}
		if err := g.DrawNForEffect(p, n); err != nil {
			return err
		}
	}
	var legs []game.ChooseCardsPrompt
	var asked []uuid.UUID
	for _, p := range players {
		var lands []uuid.UUID
		for _, id := range effects.AllHandCardIDs(g, p) {
			if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
				lands = append(lands, id)
			}
		}
		if len(lands) == 0 {
			continue
		}
		asked = append(asked, p)
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    p,
			FromPlayer: p,
			Source:     ctx.Source(),
			Question:   "The Great Aurora — put any number of land cards from your hand onto the battlefield",
			Cards:      lands,
			Min:        0,
			Max:        len(lands),
			Zone:       game.ZoneHand,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, greatAuroraLands, ctx.Source(), ctx.RunData(asked))
	return err
})

// greatAuroraLands is every player's picked lands entering together, each
// under its owner's control. Keyed.
var greatAuroraLands = effects.RunThen("prompt-run/the-great-aurora-lands", func(ctx *effects.Context, r game.RunDone) error {
	g := ctx.Game
	picks := r.Picks()
	var entries []game.BatchEntry
	for _, p := range r.Carry {
		for _, id := range picks.By(p) {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
				continue
			}
			if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() && c.Owner == p {
				entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneHand})
			}
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{}, game.KeyedTail{})
})
