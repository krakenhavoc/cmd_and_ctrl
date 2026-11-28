package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB783RevealHandDrawPerMatching is "Target opponent reveals their hand.
// You draw a card for each <match> in it" (Baleful Stare, Withering Gaze): the
// hand is revealed to the table, the matching cards are counted once, and the
// controller draws that many. A target gone by resolution does nothing (CR
// 608.2b).
func S82OldB783RevealHandDrawPerMatching(ctx *Context, match func(game.Card) bool) error {
	victim := TargetedPlayer(ctx)
	if victim == uuid.Nil {
		return nil
	}
	p := ctx.PlayerByID(victim)
	if p == nil || p.Hand == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	n := 0
	for _, c := range p.Hand.Cards {
		if match(c) {
			n++
		}
	}
	return DrawCards{N: n}.Apply(ctx)
}

// S82OldB783LandOrColor reports a card that is a land of the given subtype
// ("Mountain", "Forest") or of the given color ("R", "G"): one card counts once.
func S82OldB783LandOrColor(subtype, color string) func(game.Card) bool {
	return func(c game.Card) bool { return c.HasSubtype(subtype) || c.HasColor(color) }
}

// S82OldB783OmenLook is Omen: "Look at the top three cards of your library,
// then put them back in any order. You may shuffle. Draw a card." The order is
// asked first, the shuffle second, and the draw follows on both answers (CR
// 608.2c).
func S82OldB783OmenLook(ctx *Context) error {
	return LookAtTop{
		Player:   ctx.Controller(),
		N:        3,
		ThenTail: ctx.ItemTail(s82OldB783AskShuffle),
	}.Apply(ctx)
}

var s82OldB783AskShuffle = CardTail("tail/s82-old-b783-ask-shuffle", func(ctx *Context, _ game.TailOutcome) error {
	return MayChoice{
		Question: "Omen — shuffle your library?",
		YesLabel: "Shuffle",
		NoLabel:  "Keep that order",
		Then:     s82OldB783ShuffleThenDraw,
	}.Apply(ctx)
})

var s82OldB783ShuffleThenDraw = MayAnswer("confirm/s82-old-b783-shuffle-then-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := ctx.Game.ShuffleLibraryForEffect(r.Chooser); err != nil {
			return err
		}
	}
	return ctx.Game.DrawNForEffect(r.Chooser, 1)
})

// S82OldB783SacrificeUnlessDiscardALand is "sacrifice this creature unless you
// discard a land card" (Thundering Wurm): no land card in hand sacrifices it at
// once; otherwise the controller picks a land card to discard, or none, which
// sacrifices it if it is still the same permanent (CR 400.7).
func S82OldB783SacrificeUnlessDiscardALand(name string) Effect {
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
		}, s82OldB783LandDiscarded, nil)
	}
}

// s82OldB783LandDiscarded is the discard answered: only a land card is a legal
// pick, and no pick at all sacrifices the permanent. An on-disk key: never
// renamed, never reused.
var s82OldB783LandDiscarded = game.RegisterDiscardThenRule("choose-cards/s82-old-b783-sacrifice-unless-discard-land",
	func(picked []game.Card) bool { return len(picked) == 1 && picked[0].IsLand() },
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Item == nil || len(r.Discarded.Cards()) > 0 || !SourceIsStillThisPermanent(g, r.Item) {
			return nil
		}
		return SacrificePermanent{Target: r.Item.SourceCardID}.Apply(NewContext(g, r.Item))
	})

// S82OldB783SacrificeUnlessDiscardAtRandom is "sacrifice this creature unless
// you discard a card at random" (Pillaging Horde): with an empty hand it is
// sacrificed at once; otherwise the controller chooses between discarding a
// card at random and sacrificing it. The permanent is judged as the object that
// entered (CR 400.7).
func S82OldB783SacrificeUnlessDiscardAtRandom(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Eliminated {
			return nil
		}
		ctx := NewContext(g, item)
		if p.Hand == nil || p.Hand.Size() == 0 {
			return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
		}
		return PickOption{
			Question: name + " — discard a card at random, or sacrifice " + name,
			Options: []game.ChoiceOption{
				{Label: "Discard a card at random"},
				{Label: "Sacrifice " + name},
			},
			ThenKey: s82OldB783DiscardAtRandomOrThis,
		}.Apply(ctx)
	}
}

var s82OldB783DiscardAtRandomOrThis = OptionPickThen("option-pick/s82-old-b783-discard-at-random-or-sacrifice", func(ctx *Context, r game.OptionPicked) error {
	if ctx.Item == nil || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	if r.Index == 0 {
		return ctx.Game.DiscardRandomForEffect(ctx.Controller(), 1)
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})

// S82OldB783AnswerSacrificeAForestOrThis is Plant Elemental's tax answered: the
// controller sacrifices a Forest of their choice, or the Elemental.
var S82OldB783AnswerSacrificeAForestOrThis = sacrificeOneOrThisAnswer("option-pick/s82-old-b783-sacrifice-this-unless-you-sacrifice-a-forest", "a Forest", Subtype("Forest"))
