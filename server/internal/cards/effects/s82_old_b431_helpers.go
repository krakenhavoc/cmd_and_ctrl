package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB431NoggleRansackerBody is Noggle Ransacker's trigger: each player
// draws two cards, then discards a card at random, one player at a time in
// turn order from the controller (CR 101.4). The discard waits for that
// player's draws through the draw continuation, so a replaced draw is still
// followed by the discard.
func S82OldB431NoggleRansackerBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range TablePlayers(ctx) {
		if err := g.DrawNThenForEffect(id, 2, game.DrawThen{Ref: s82OldB431RandomDiscard, Player: id, Source: item.SourceCardID}); err != nil {
			return err
		}
	}
	return nil
}

// s82OldB431RandomDiscard is "then discards a card at random". An on-disk key:
// never renamed, never reused.
var s82OldB431RandomDiscard = game.RegisterDrawThen("s82-old-b431/discard-a-card-at-random", func(g *game.Game, d game.DrawThen) error {
	return g.DiscardRandomForEffect(d.Player, 1)
})

// S82OldB431SeeBeyond is See Beyond: draw two cards, then shuffle a card from
// your hand into your library. The pick waits for the draws (CR 608.2c).
func S82OldB431SeeBeyond(item *game.StackItem, ctx *Context) error {
	return ctx.Game.DrawNThenForEffect(item.Controller, 2, game.DrawThen{Ref: s82OldB431SeeBeyondPick, Player: item.Controller, Source: item.SourceCardID})
}

// s82OldB431SeeBeyondPick asks for the card to shuffle away; an empty hand asks
// nothing. An on-disk key: never renamed, never reused.
var s82OldB431SeeBeyondPick = game.RegisterDrawThen("s82-old-b431/see-beyond-pick", func(g *game.Game, d game.DrawThen) error {
	hand := AllHandCardIDs(g, d.Player)
	if len(hand) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    d.Player,
		FromPlayer: d.Player,
		Source:     d.Source,
		Question:   "See Beyond — shuffle a card from your hand into your library",
		Cards:      hand,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneHand,
		ThenKey:    s82OldB431SeeBeyondShuffle,
	})
	return nil
})

// s82OldB431SeeBeyondShuffle puts the picked card into its owner's library and
// shuffles that library. An on-disk key: never renamed, never reused.
var s82OldB431SeeBeyondShuffle = game.RegisterChooseCardsThen("choose-cards/s82-old-b431-see-beyond-shuffle", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect(r.Picked[:1], game.TuckOptions{}, game.KeyedTail{Ref: ThenShuffle, Player: r.Chooser, Source: r.Source})
})

// S82OldB431CountLands counts the lands the controller controls with the basic
// land type word (read through the source's text, CR 612.2).
func S82OldB431CountLands(ctx *Context, subtype string) int {
	word := ctx.TextSubtype(subtype)
	return CountYouControl(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsLand() && c.HasSubtype(word)
	})
}

// S82OldB431AuraThatCouldEnchant is the search clause of Auratouched Mage: an
// Aura card that could enchant the creature `host`. While it is on the
// battlefield that is the Aura's enchant clause tested against it; once it has
// left, the clause is tested against the card as it last was (its printed
// characteristics), CR 608.2h.
func S82OldB431AuraThatCouldEnchant(g *game.Game, controller, host uuid.UUID) func(c game.Card) bool {
	return func(c game.Card) bool {
		if !c.HasSubtype("Aura") {
			return false
		}
		if OnBattlefield(g, host) {
			return CanBeAttachedTo(g, controller, c, host)
		}
		hostCard, ok := g.LookupCardForEffect(host)
		if !ok {
			return false
		}
		spec := game.TargetSpecFor(game.CatalogKey(c))
		if spec == nil {
			return true
		}
		if spec.CardOK == nil {
			return hostCard.IsCreature()
		}
		return spec.CardOK(g, controller, hostCard, game.ZoneBattlefield)
	}
}

// S82OldB431AuratouchedMageSearch is Auratouched Mage's trigger: search for an
// Aura card that could enchant it; if the Mage is still on the battlefield the
// Aura enters attached to it, otherwise the Aura card is revealed and put into
// your hand; then shuffle (both branches).
func S82OldB431AuratouchedMageSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	mage := item.SourceCardID
	pred := S82OldB431AuraThatCouldEnchant(g, item.Controller, mage)
	if OnBattlefield(g, mage) && !SourceIsNewObject(g, item) {
		return SearchLibrary{
			Player:           item.Controller,
			Predicate:        pred,
			Dest:             game.ZoneBattlefield,
			Limit:            1,
			Shuffle:          true,
			Optional:         true,
			Reason:           "Auratouched Mage — an Aura card that could enchant it",
			EntersAttachedTo: game.TargetRef{Kind: game.TargetCard, ID: mage},
		}.Apply(ctx)
	}
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: pred,
		Dest:      game.ZoneHand,
		Limit:     1,
		Shuffle:   true,
		Optional:  true,
		Reveal:    true,
		Reason:    "Auratouched Mage — an Aura card that could enchant it",
	}.Apply(ctx)
}
