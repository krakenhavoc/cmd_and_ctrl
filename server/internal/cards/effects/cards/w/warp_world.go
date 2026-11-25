package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warp World — Sorcery {5}{R}{R}{R} (Secret Lair Drop):
//
//	"Each player shuffles all permanents they own into their library, then
//	 reveals that many cards from the top of their library. Each player puts all
//	 artifact, creature, and land cards revealed this way onto the battlefield,
//	 then does the same for enchantment cards, then puts all cards revealed this
//	 way that weren't put onto the battlefield on the bottom of their library."
//
// Three stages joined by keyed tails, with what each reads riding the tail as data.
// First every permanent goes to its OWNER's library as one simultaneous move (a
// token goes there and ceases to exist, CR 111.7; a commander its owner sends to the
// command zone instead is not shuffled in). The count per player is the permanents
// of theirs that reached a library plus their tokens, and a player with a count
// shuffles their library. Then each player in APNAP order reveals that many cards from
// the top, and the artifact, creature and land cards among them enter together under
// their owners' control (CR 603.6a: each sees the others enter); then the revealed
// enchantment cards still in a library enter together (an Aura chooses what it
// enchants as it enters, CR 303.4f, and stays behind if it has nothing to); then the
// rest of the revealed cards still in a library go on the bottom of their owners'
// libraries in the owner's choice of order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50a228a2-c8b6-4416-b0ab-417926a9b9b6",
		Name:         "Warp World",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			var all, tokenOwners []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				p := g.PlayerByIDForEffect(c.Owner)
				if p == nil || p.Eliminated {
					continue
				}
				all = append(all, c.InstanceID)
				if c.IsToken() {
					tokenOwners = append(tokenOwners, c.Owner)
				}
			}
			return g.TuckCardsToLibraryKeyedForEffect(all, game.TuckOptions{},
				effects.WithCarry(ctx.ItemTail(warpWorldShuffled), tokenOwners...))
		},
	})
}

// warpWorldShuffled is the permanents shuffled in: count, shuffle, reveal and put
// the artifacts, creatures and lands. An on-disk key.
var warpWorldShuffled = effects.CardTail("tail/warp-world-shuffled", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	counts := map[uuid.UUID]int{}
	for _, owner := range o.Tail.Carry { // the tokens, one entry each
		counts[owner]++
	}
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok {
			counts[c.Owner]++
		}
	}
	var revealed []uuid.UUID
	for _, p := range effects.ApnapPlayers(g) {
		n := counts[p]
		if n == 0 {
			continue
		}
		if err := g.ShuffleLibraryForEffect(p); err != nil {
			return err
		}
		revealed = append(revealed, g.RevealTopOfLibraryForEffect(p, ctx.Source(), n, "Warp World — reveal that many cards")...)
	}
	return warpWorldPut(ctx, revealed, func(c game.Card) bool {
		return c.IsArtifact() || c.IsCreature() || c.IsLand()
	}, warpWorldEnchantments)
})

// warpWorldEnchantments is the second stage: the enchantment cards. An on-disk key.
var warpWorldEnchantments = effects.CardTail("tail/warp-world-enchantments", func(ctx *effects.Context, o game.TailOutcome) error {
	return warpWorldPut(ctx, o.Tail.Carry, func(c game.Card) bool { return c.IsEnchantment() }, warpWorldBottom)
})

// warpWorldBottom is the last stage: what is still in a library goes to the bottom,
// each owner ordering their own cards. An on-disk key.
var warpWorldBottom = effects.CardTail("tail/warp-world-bottom", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	byOwner := map[uuid.UUID][]uuid.UUID{}
	for _, id := range o.Tail.Carry {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok {
			byOwner[c.Owner] = append(byOwner[c.Owner], id)
		}
	}
	for _, p := range effects.ApnapPlayers(g) {
		ids := byOwner[p]
		if len(ids) == 0 {
			continue
		}
		if err := (effects.PutInLibraryInAnyOrder{
			Chooser:   p,
			Cards:     ids,
			From:      game.ZoneLibrary,
			Placement: game.LibraryPlaceBottom,
			Label:     "Warp World — put the rest on the bottom of your library in any order",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// warpWorldPut puts the cards among `revealed` that are still in a library and that
// `match` accepts onto the battlefield as one entry, then runs `next` with every
// revealed card riding its carry.
func warpWorldPut(ctx *effects.Context, revealed []uuid.UUID, match func(game.Card) bool, next game.TailRef) error {
	g := ctx.Game
	var entries []game.BatchEntry
	for _, id := range revealed {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && match(c) {
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneLibrary})
		}
	}
	tail := effects.WithCarry(ctx.Tail(next), revealed...)
	tail.Player, tail.Source = ctx.Controller(), ctx.Source()
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{}, tail)
}
