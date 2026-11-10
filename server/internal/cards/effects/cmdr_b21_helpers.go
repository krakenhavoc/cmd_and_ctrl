package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b21_helpers.go — helpers for the Commander deck cards, batch 21
// (#5368). Every name carries the cmdrB21 prefix.

// cmdrB21UpToTwoBasicsTapped is "Search your library for up to two basic
// land cards, put them onto the battlefield tapped, then shuffle."
func cmdrB21UpToTwoBasicsTapped(ctx *Context, name string) error {
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        name + " — up to two basic lands",
	}.Apply(ctx)
}

// cmdrB21OtherCreatureToughness is "the total toughness of other
// creatures you control", read as the entering card's controller sees the
// battlefield (effective toughness, the entering card excluded).
func cmdrB21OtherCreatureToughness(g *game.Game, src *game.Card) int {
	total := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.InstanceID != src.InstanceID && c.IsCreature() {
			total += c.CurrentToughness()
		}
	}
	return total
}

// cmdrB21CardTypes are the card types a "shares a card type" clause reads
// (CR 205.2a), lowercase.
var cmdrB21CardTypes = []string{"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"}

// cmdrB21BaralOffer is Baral and Kari Zev's resolution: offer the free cast,
// or make the Monkey Pirate at once when there is nothing to offer.
func cmdrB21BaralOffer(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	spellID := item.Trigger.Event.CardID
	spell, ok := g.LookupCardForEffect(spellID)
	if !ok {
		return cmdrB21RagavanToken(g, controller, source)
	}
	var shared []string
	for _, t := range cmdrB21CardTypes {
		if spell.HasCardType(t) {
			shared = append(shared, t)
		}
	}
	mv, _ := g.ManaValueForEffect(spell)
	if mv <= 0 || len(shared) == 0 {
		return cmdrB21RagavanToken(g, controller, source)
	}
	clause := freeCastSpell{filter: game.PermissionFilter{CardTypesAny: strings.Join(shared, " ")}}.withManaValueAtMost(mv - 1)
	perm := clause.on(g.FreeCastFromHandPermission(controller, source, "Baral and Kari Zev"))
	candidates := freeCastCandidates(g, &perm, allHandCardIDs(g, controller))
	if len(candidates) == 0 {
		return cmdrB21RagavanToken(g, controller, source)
	}
	if !g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Baral and Kari Zev",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Create First Mate Ragavan instead",
		Then:         cmdrB21BaralThen,
		Carry:        candidates,
	}) {
		return cmdrB21RagavanToken(g, controller, source)
	}
	return nil
}

// cmdrB21BaralThen is "If you don't": nothing was cast, so the Monkey
// Pirate is made. An on-disk key: never renamed, never reused.
var cmdrB21BaralThen = game.RegisterChooseCardsThen("resolution-cast/baral-and-kari-zev-ragavan", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		return nil
	}
	return cmdrB21RagavanToken(g, r.Chooser, r.Source)
})

// cmdrB21RagavanToken is First Mate Ragavan, which gains haste until end of
// turn.
func cmdrB21RagavanToken(g *game.Game, controller, source uuid.UUID) error {
	ctx := keyedContext(g, controller, source)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: controller,
		Source:     source,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("2/1 red Legendary Monkey Pirate named First Mate Ragavan")), Count: 1}},
	}, ctx.Tail(cmdrB21RagavanHaste))
}

// cmdrB21RagavanHaste is "it gains haste until end of turn" over the token
// that landed. An on-disk key.
var cmdrB21RagavanHaste = CardTail("baral-and-kari-zev/ragavan-haste", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Baral and Kari Zev — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
