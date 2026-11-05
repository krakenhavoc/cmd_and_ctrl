package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b34_helpers.go — shared pieces for Star Trek Commander batch 34
// (#4974).

// trekB34ArtifactCreaturesYouControl is "artifact creatures you control",
// read on post-layer types: an artifact that is a creature only for now (a
// crewed Vehicle, a stationed Spacecraft) is one while it is, and the
// source counts itself when it is one.
func trekB34ArtifactCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
}

// trekB34GiftOfPerfection is Gift of Perfection's resolution:
//
//	"Each opponent reveals cards from the top of their library until they
//	 reveal a creature card. Assimilate one of the creature cards revealed
//	 this way. Then each opponent puts the other cards they revealed this
//	 way on the bottom of their library in a random order."
//
// Each opponent still in the game reveals in turn order from the caster
// (CR 101.4); nothing moves while they do. The caster then picks one of
// the creature cards (a reveal pick over every library at once, mandatory:
// "assimilate one" is not a may), which is assimilated straight out of its
// owner's library (trekB11Assimilate's entry, from the library). Every
// other revealed card, the creature cards not picked among them, goes to
// the bottom of its owner's library in a random order. With one creature
// card revealed there is no choice to make; with none, nothing is
// assimilated and every revealed card goes to the bottom.
func trekB34GiftOfPerfection(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var hits, rest []uuid.UUID
	for _, opp := range opponentsInTurnOrder(g, item.Controller) {
		run, hit := revealUntil(ctx, opp, func(c game.Card) bool { return c.IsCreature() },
			"Gift of Perfection — reveal until a creature card")
		for _, id := range run {
			if id == hit {
				hits = append(hits, id)
			} else {
				rest = append(rest, id)
			}
		}
	}
	if len(hits) <= 1 {
		return trekB34GiftOfPerfectionFinish(ctx, hits, rest)
	}
	queued, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
		Chooser:  item.Controller,
		Owner:    item.Controller,
		Source:   ctx.Source(),
		Question: "Gift of Perfection — assimilate one of the creature cards revealed this way",
		Cards:    hits,
		Min:      1,
		Max:      1,
		ThenKey:  trekB34GiftOfPerfectionPicked,
		Carry:    rest,
		Asked:    game.NewAskedItem(item),
	})
	if err != nil || queued > 0 {
		return err
	}
	// Nobody could be asked (the caster has left): nothing is
	// assimilated and every revealed card goes to the bottom.
	return trekB34GiftOfPerfectionFinish(ctx, nil, append(hits, rest...))
}

var trekB34GiftOfPerfectionPicked = game.RegisterChooseCardsThen("choose-cards/gift-of-perfection-assimilate",
	func(g *game.Game, r game.CardsChosen) error {
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		picked, left := revealPartition(r.Candidates, r.Picked)
		if len(picked) > 1 {
			picked, left = picked[:1], append(left, picked[1:]...)
		}
		return trekB34GiftOfPerfectionFinish(ctx, picked, append(left, r.Carry...))
	})

// trekB34GiftOfPerfectionFinish assimilates `picked` (at most one card) and
// puts `rest` on the bottom of their owners' libraries in a random order,
// one owner at a time.
func trekB34GiftOfPerfectionFinish(ctx *Context, picked, rest []uuid.UUID) error {
	for _, id := range picked {
		if err := trekB34AssimilateFromLibrary(ctx, id); err != nil {
			return err
		}
	}
	byOwner := map[uuid.UUID][]uuid.UUID{}
	var owners []uuid.UUID
	for _, id := range rest {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if _, seen := byOwner[c.Owner]; !seen {
			owners = append(owners, c.Owner)
		}
		byOwner[c.Owner] = append(byOwner[c.Owner], id)
	}
	for _, owner := range owners {
		if err := ctx.Game.PutOnBottomInRandomOrderForEffect(owner, game.ZoneLibrary, byOwner[owner]); err != nil {
			return err
		}
	}
	return nil
}

// trekB34AssimilateFromLibrary is the keyword action assimilate on a card
// in a library: trekB11Assimilate's entry (a +1/+1 counter, a Borg artifact
// creature that loses its other creature types), from the library.
func trekB34AssimilateFromLibrary(ctx *Context, card uuid.UUID) error {
	return putCardsFromZoneTogetherWith(ctx, game.ZoneLibrary, []uuid.UUID{card},
		game.ZoneEntryOptions{
			Controller:         ctx.Controller(),
			EntersWithCounters: map[string]int{game.CounterPlusOne: 1},
			EntersAs: &game.EntersAs{
				AddTypes:      []string{"Artifact"},
				CreatureTypes: ctx.TextSubtypes("Borg"),
				Label:         "Gift of Perfection — assimilated: a Borg artifact creature",
			},
		}, game.TailRef{}, 0)
}
