package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Spock — Legendary Creature — Vulcan Human Scientist
// {2}{G}{W}{U}, 3/5 (Star Trek Commander, #4974):
//
//	"Strange new worlds — Whenever Commander Spock enters or attacks,
//	 reveal the top card of your library. You may cast that card without
//	 paying its mana cost if that spell's mana value is equal to or less
//	 than the number of differently named lands you control. If it's a
//	 land card, you may put it onto the battlefield tapped. If that card
//	 is still on top of your library, draw a card."
//
// "Strange new worlds" is an ability word; the trigger is one ability
// watching two events (Sun Titan).
//
// The reveal is public. The cast is made during the resolution
// (CR 608.2g) from the top of the library, through the resolution_cast
// prompt over a free, any-timing, cast-only permission capped at the
// number of differently named lands you control as the trigger resolves
// (Aetherworks Marvel's offer with Cosmic Cube's cap, judged on the face
// being cast). Declining moves nothing. A land card is then offered to be
// put onto the battlefield tapped (not played: no land drop is used,
// CR 305.4). Last, if the revealed card is still the top card of your
// library — not cast, not put — you draw a card, a real draw. An empty
// library reveals nothing and draws nothing.
func init() {
	Register(Spec{
		OracleID:     "60644afe-9645-45a1-9eac-70f327daa6a7",
		Name:         "Commander Spock",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks(commanderSpockLabel, commanderSpockStrangeNewWorlds),
		},
	})
}

const commanderSpockLabel = "Commander Spock — reveal the top card of your library; you may cast it free"

func commanderSpockStrangeNewWorlds(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you, source := item.Controller, ctx.Source()
	revealed := g.RevealTopOfLibraryForEffect(you, source, 1, commanderSpockLabel)
	if len(revealed) == 0 {
		return nil
	}
	perm := spellWithManaValueAtMost(trekB8DifferentlyNamedLands(g, you)).
		on(g.FreeCastFromLibraryPermission(you, source, "Commander Spock"))
	perm.LapseOnPass = game.LapseStaysInHand
	if len(freeCastCandidates(g, &perm, revealed)) > 0 && g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       source,
		Reason:       "Commander Spock",
		Cards:        revealed,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Don't cast it",
		Then:         commanderSpockAfterCast,
		Carry:        revealed,
	}) {
		return nil
	}
	return commanderSpockLandOrDraw(ctx, you, revealed[0])
}

// commanderSpockAfterCast is the rest of the resolution once the cast
// offer is answered. Carry is the revealed card. An on-disk key.
var commanderSpockAfterCast = game.RegisterChooseCardsThen("resolution-cast/commander-spock-after-cast",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Carry) == 0 {
			return nil
		}
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return commanderSpockLandOrDraw(ctx, r.Chooser, r.Carry[0])
	})

// commanderSpockLandOrDraw is the last two sentences: a land card still
// on top may be put onto the battlefield tapped; then, if the card is
// still on top, draw a card.
func commanderSpockLandOrDraw(ctx *Context, you, card uuid.UUID) error {
	if !commanderSpockOnTop(ctx.Game, you, card) {
		return nil
	}
	if c, ok := ctx.Game.LookupCardForEffect(card); ok && c.IsLand() {
		return MayChoice{
			Player:   you,
			Question: "Commander Spock — put " + c.Name + " onto the battlefield tapped?",
			YesLabel: "Put it onto the battlefield tapped",
			NoLabel:  "Leave it on top",
			Then:     commanderSpockLandAnswer,
			Carry:    []uuid.UUID{you, card},
		}.Apply(ctx)
	}
	return DrawCards{Player: you, N: 1}.Apply(ctx)
}

// commanderSpockLandAnswer puts the land onto the battlefield tapped on
// a yes, and draws once it has landed if the card is somehow still on
// top (a replacement kept it there); a no draws it. Carry is you, then
// the card. An on-disk key.
var commanderSpockLandAnswer = MayAnswer("confirm/commander-spock-land", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) < 2 {
		return nil
	}
	you, card := r.Carry[0], r.Carry[1]
	then := withCarry(ctx.ItemTail(commanderSpockDrawIfOnTop), you, card)
	if !r.Accepted || !commanderSpockOnTop(ctx.Game, you, card) {
		return then.Run(ctx.Game, nil)
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: card, From: game.ZoneLibrary}},
		game.ZoneEntryOptions{Controller: you, Tapped: true}, then)
})

// commanderSpockDrawIfOnTop is "if that card is still on top of your
// library, draw a card". Carry is you, then the card. An on-disk key.
var commanderSpockDrawIfOnTop = CardTail("trek-b31/commander-spock-draw-if-on-top", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) < 2 {
		return nil
	}
	you, card := o.Tail.Carry[0], o.Tail.Carry[1]
	if !commanderSpockOnTop(ctx.Game, you, card) {
		return nil
	}
	return DrawCards{Player: you, N: 1}.Apply(ctx)
})

// commanderSpockOnTop reports whether `card` is the top card of `you`'s
// library.
func commanderSpockOnTop(g *game.Game, you, card uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Library == nil || p.Library.Size() == 0 {
		return false
	}
	return p.Library.Cards[p.Library.Size()-1].InstanceID == card
}
