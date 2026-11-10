package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// God-Pharaoh's Gift — Artifact {7}:
//
//	"At the beginning of combat on your turn, you may exile a creature card
//	 from your graveyard. If you do, create a token that's a copy of that
//	 card, except it's a 4/4 black Zombie. It gains haste until end of
//	 turn."
//
// Kishla Trawlers' choice (a pick of up to one creature card from your
// graveyard at resolution, not a target) into The Scarab God's copy: the
// token is copied from the card that really reached exile (a card that left
// the graveyard in response, or a replacement that sent it elsewhere, makes
// no token, CR 608.2b), and the exception REPLACES the copy's creature
// types with Zombie (CR 707.9a). The haste is an effect on the token, not
// part of the copy (CR 707.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f2d2fda-f098-4375-8fe9-63f929e3928c",
		Name:         "God-Pharaoh's Gift",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("God-Pharaoh's Gift — you may exile a creature card from your graveyard; create a 4/4 black Zombie token copy with haste",
				godPharaohsGiftOffer),
		},
	})
}

// godPharaohsGiftOffer asks for up to one creature card of the controller's
// graveyard. With none there is nothing to ask.
func godPharaohsGiftOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	me := ctx.Controller()
	var cards []uuid.UUID
	if p := g.PlayerByIDForEffect(me); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.HasCardType("creature") {
				cards = append(cards, c.InstanceID)
			}
		}
	}
	if len(cards) == 0 {
		return nil
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect([]game.ChooseCardsPrompt{{
		Chooser:    me,
		FromPlayer: me,
		Source:     ctx.Source(),
		Question:   "God-Pharaoh's Gift — you may exile a creature card from your graveyard to create a 4/4 black Zombie copy of it",
		Cards:      cards,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneGraveyard,
	}}, godPharaohsGiftExiled, ctx.Source(), ctx.RunData([]uuid.UUID{me}))
	return err
}

// godPharaohsGiftExiled exiles the card picked, if any. An on-disk key.
var godPharaohsGiftExiled = RunThen("prompt-run/god-pharaohs-gift-exile", func(ctx *Context, r game.RunDone) error {
	var chosen []uuid.UUID
	for _, player := range r.Carry {
		chosen = append(chosen, r.Picks().By(player)...)
	}
	if len(chosen) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(chosen, ctx.ItemTail(godPharaohsGiftCopy))
})

// godPharaohsGiftCopy is "if you do, create a token that's a copy of that
// card, except it's a 4/4 black Zombie", once the card has landed in exile.
// An on-disk key.
var godPharaohsGiftCopy = CardTail("then/god-pharaohs-gift-copy", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return CreateTokenCopiesTogether{
		Controller: ctx.Controller(),
		Copies:     o.IDs[:1],
		Except:     scarabGodZombieExceptionFor(ctx.Game.TextReaderForEffect(ctx.Item)),
		Then:       ctx.ItemTail(godPharaohsGiftHaste),
	}.Apply(ctx)
})

// godPharaohsGiftHaste is "it gains haste until end of turn", told the
// token by the copy that made it. An on-disk key.
var godPharaohsGiftHaste = CardTail("then/god-pharaohs-gift-haste", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
			Label: "God-Pharaoh's Gift — the token gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
