package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB53YouSacrificedANontoken is "whenever you sacrifice a nontoken
// permanent": the source's controller sacrificed it, the source itself
// included.
func GapB53YouSacrificedANontoken(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || source == nil || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// gapB53PumpAndHaste is Conduit Goblin's paid half: "another target
// creature you control gets +1/+0 and gains haste until end of turn". The
// target is read as the trigger resolves; one that is gone gets nothing
// (CR 608.2b).
var GapB53PumpAndHaste = EnergyPaid("pay-unless/gap-b53-conduit-goblin-pump-haste",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: 1, Label: "Conduit Goblin — +1/+0 until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Conduit Goblin — haste until end of turn"}.Apply(ctx)
	})

// gapB53ThisWasKicked is the AppliesTo of "if this creature was kicked,
// it enters with … haste": the source itself, when it arrived by a
// kicked spell (CR 702.33d, game.CardKickedTimes).
func GapB53ThisWasKicked(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target != nil && source != nil && target.InstanceID == source.InstanceID && game.CardKickedTimes(*source) > 0
}

// gapB53SelfBasePT is "this creature has base power and toughness P/T" as a
// layer 7b set on the source itself, so +1/+1 counters and anthems still
// count (CR 613.4b).
func GapB53SelfBasePT(power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7B_Set,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target != nil && source != nil && target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power = power
			c.Toughness = toughness
		},
	}
}

// gapB53SelfKeyword is "this creature has <keyword>" as a layer 6 grant on
// the source itself.
func GapB53SelfKeyword(keyword string) game.StaticAbility {
	s := B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target != nil && source != nil && target.InstanceID == source.InstanceID
	}, keyword)
	s.Label = "This creature has " + keyword + "."
	return s
}

// gapB53ThrasherExileTwo is the paid half of Party Thrasher's "you may
// discard a card. If you do, exile the top two cards of your library, then
// choose one of them. You may play that card this turn": run once a card has
// really been discarded (the run's count, not the hand's). An on-disk key.
var GapB53ThrasherExileTwo = game.RegisterDiscardThen("choose-cards/gap-b53-party-thrasher-exile-two",
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		return g.ExileTopKeyedForEffect(r.Player, r.Player, 2, KeyedContext(g, r.Player, r.Source).Tail(gapB53ThrasherChoose))
	})

// gapB53ThrasherChoose is "choose one of them" over the cards that reached
// exile. An on-disk key.
var gapB53ThrasherChoose = CardTail("gap-b53-party-thrasher/choose-exiled", func(ctx *Context, o game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Party Thrasher — choose one of the exiled cards; this turn you may play it",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneExile,
		ThenKey:    gapB53ThrasherChosen,
	})
	return nil
})

// gapB53ThrasherChosen gives the chosen card its window: its controller may
// play it this turn. An on-disk key.
var gapB53ThrasherChosen = game.RegisterChooseCardsThen("choose-cards/gap-b53-party-thrasher-play",
	func(g *game.Game, r game.CardsChosen) error {
		for _, id := range r.Picked {
			g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
				Player: r.Chooser,
				Label:  "Party Thrasher — you may play that card this turn",
			})
		}
		return nil
	})

// gapB53NoncreatureFromExile is Party Thrasher's convoke filter:
// "noncreature spells you cast from exile".
func GapB53NoncreatureFromExile(_ *game.Game, _ uuid.UUID, spell game.Card, from game.ZoneKind) bool {
	return from == game.ZoneExile && !spell.IsCreature()
}
