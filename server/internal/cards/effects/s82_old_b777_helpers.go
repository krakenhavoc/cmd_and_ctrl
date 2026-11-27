package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b777_helpers.go — S82 OLD batch 777 (Weatherlight, #5629).

// S82OldB777SacrificeUnlessDiscardALand is "sacrifice it unless you discard a
// land card" (Fallow Wurm): a player with no land card in hand sacrifices the
// permanent at once; otherwise they pick a land card to discard, or none, which
// sacrifices it if it is still the same permanent (CR 400.7).
func S82OldB777SacrificeUnlessDiscardALand(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Eliminated {
			return nil
		}
		hasLand := false
		if p.Hand != nil {
			for _, c := range p.Hand.Cards {
				if c.IsLand() {
					hasLand = true
				}
			}
		}
		if !hasLand {
			return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
		}
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   item.Controller,
			Source:   item.SourceCardID,
			N:        1,
			UpTo:     true,
			Question: name + " — discard a land card, or sacrifice it?",
			Asked:    game.NewAskedItem(item),
		}, s82OldB777Discarded, nil)
	}
}

// s82OldB777Discarded is the discard answered: only a land card is a legal
// pick, and no pick at all sacrifices the permanent. An on-disk key: never
// renamed, never reused.
var s82OldB777Discarded = game.RegisterDiscardThenRule("choose-cards/s82-old-b777-sacrifice-unless-discard-land",
	func(picked []game.Card) bool { return len(picked) == 1 && picked[0].IsLand() },
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Item == nil || len(r.Discarded.Cards()) > 0 {
			return nil
		}
		if !SourceIsStillThisPermanent(g, r.Item) {
			return nil
		}
		return SacrificePermanent{Target: r.Item.SourceCardID}.Apply(NewContext(g, r.Item))
	})

// S82OldB777SacrificeAForestOrThisAnswer is the keyed answer to Rogue
// Elephant's "sacrifice it unless you sacrifice a Forest": index 0 asks for a
// Forest, any other index sacrifices the permanent. An on-disk key.
var S82OldB777SacrificeAForestOrThisAnswer = sacrificeOneOrThisAnswer(
	"option-pick/s82-old-b777-sacrifice-this-unless-you-sacrifice-a-forest", "a Forest", HasSubtype("Forest"))

// S82OldB777TopCreatureCardOfYourGraveyard is the topmost creature card of
// `player`'s graveyard, the top being the last card in the zone (CR 404.2),
// and false when there is none.
func S82OldB777TopCreatureCardOfYourGraveyard(g *game.Game, player uuid.UUID) (uuid.UUID, bool) {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return uuid.Nil, false
	}
	for i := len(p.Graveyard.Cards) - 1; i >= 0; i-- {
		if p.Graveyard.Cards[i].IsCreature() {
			return p.Graveyard.Cards[i].InstanceID, true
		}
	}
	return uuid.Nil, false
}

// S82OldB777EachPlayerDiscardsAnyNumberThenDraws is Flux's first sentence:
// each player, in turn order from the active player, discards any number of
// cards and then draws that many (the player's own leg), and `then` runs
// once every player has answered.
func S82OldB777EachPlayerDiscardsAnyNumberThenDraws(ctx *Context, question string, then game.PromptRunThen) error {
	return ctx.Game.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil, game.DiscardPrompt{
		Source:   ctx.Source(),
		N:        1 << 20,
		UpTo:     true,
		Question: question,
	}, then, []uuid.UUID{ctx.Controller()})
}

// S82OldB777DrawAsManyAsDiscarded is the per-seat tail of that sentence.
func S82OldB777DrawAsManyAsDiscarded(g *game.Game, leg game.RunLeg) error {
	if n := len(leg.Landed); n > 0 {
		return g.DrawNForEffect(leg.Seat, n)
	}
	return nil
}
