package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summon: Esper Valigarmanda — Enchantment Creature — Saga Drake
// {3}{R}, 3/3 (Final Fantasy; deck request #3655, seam #3773):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Exile an instant or sorcery card from each graveyard.
//	 II, III, IV — Add {R} for each lore counter on this Saga. You may
//	 cast an instant or sorcery card exiled with this Saga, and mana of
//	 any type can be spent to cast that spell.
//	 Flying, haste"
//
// The pieces were all in the engine; this card is their assembly.
//
//   - Chapter I is one choice per graveyard, made by the Saga's
//     controller (not a target, so hexproof and protection are not
//     involved), and the chosen cards leave together. A graveyard with
//     no instant or sorcery card is skipped, and one with exactly one is
//     not asked about. The exile is LINKED to the Saga as an object
//     (CR 607.2a, ExileLinkedKeyedForEffect): the cards carry
//     Card.ExiledWith, so the three later chapters read "exiled with
//     this Saga" off the card, and a Saga that left and came back is a
//     new object with no claim on them (CR 400.7).
//   - The mana is read when the chapter resolves, from the Saga's lore
//     counters as the permanent is (or, if it has gone, as it last
//     existed there, CR 608.2h): II adds two {R}, III three, IV four.
//   - The cast is a grant over the instant and sorcery cards still
//     exiled with the Saga, for ONE cast ("an instant or sorcery card"),
//     at flash timing because it happens during the resolution
//     (CR 608.2g), at the card's own mana cost, and with
//     CastPermission.AnyType so the {R} just added — or any other
//     mana — pays any part of it (CR 106.1b: colourless included).
//     Cards not cast stay exiled, linked, for the next chapter.
//
// #4250: the cast is made DURING the chapter's resolution (CR 608.2g),
// through the resolution_cast prompt over the cards exiled with the Saga,
// paid then with the mana the chapter added; one permission, one cast. A
// card not cast waits for the next chapter, as printed.
func init() {
	const name = "Summon: Esper Valigarmanda"
	chapter := func(n int) game.TriggeredAbility {
		return ChapterTrigger(n, SagaChapterLabel(name, n, "add {R} for each lore counter, then you may cast an exiled instant or sorcery card"),
			valigarmandaManaAndCast)
	}
	Register(Spec{
		OracleID:        "cecd2f49-8ccb-472d-9cf1-fd0f4d93546d",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "exile an instant or sorcery card from each graveyard"),
				valigarmandaExileFromEachGraveyard),
			chapter(2),
			chapter(3),
			chapter(4),
		},
	})
}

// isInstantOrSorcery is the card filter chapter I picks by and chapters
// II-IV re-check.
func isInstantOrSorcery(c game.Card) bool {
	return c.IsInstant() || c.IsSorcery()
}

// valigarmandaExileFromEachGraveyard is chapter I: the controller picks
// one instant or sorcery card from each graveyard in turn, then the
// picks are exiled together, linked to the Saga.
func valigarmandaExileFromEachGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	seats := append([]uuid.UUID{ctx.Controller()}, ctx.Opponents()...)
	return valigarmandaPickFrom(ctx, seats, nil)
}

// valigarmandaPickFrom asks about the first of `seats` and recurses on
// the rest from the answer; with nobody left it exiles `picked`.
func valigarmandaPickFrom(ctx *Context, seats, picked []uuid.UUID) error {
	if len(seats) == 0 {
		return valigarmandaExile(ctx, picked)
	}
	seat, rest := seats[0], seats[1:]
	p := ctx.Game.PlayerByIDForEffect(seat)
	var cands []uuid.UUID
	if p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if isInstantOrSorcery(c) {
				cands = append(cands, c.InstanceID)
			}
		}
	}
	switch len(cands) {
	case 0:
		return valigarmandaPickFrom(ctx, rest, picked)
	case 1:
		return valigarmandaPickFrom(ctx, rest, append(append([]uuid.UUID(nil), picked...), cands[0]))
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: seat,
		Source:     ctx.Source(),
		Question:   "Summon: Esper Valigarmanda — exile an instant or sorcery card from this graveyard",
		Cards:      cands,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    valigarmandaPick,
		Asked:      game.NewAskedItem(ctx.Item),
		// The seats still to ask, then the cards picked so far.
		Carry: packIDLists(rest, picked),
	})
	return nil
}

// valigarmandaPick is one graveyard's card chosen, keyed (#4253 sweep): the
// seats still to ask and the picks so far ride the prompt as data, and the
// chapter's item as the asking item. Registered in init, because the body
// asks the next graveyard with it.
var valigarmandaPick game.ChooseCardsThen

func init() {
	valigarmandaPick = askedPick("choose-cards/summon-esper-valigarmanda-exile-each", func(ctx *Context, r game.CardsChosen) error {
		rest, so := unpackIDPair(r.Carry)
		return valigarmandaPickFrom(ctx, rest, append(so, r.Picked...))
	})
}

// valigarmandaExile exiles the picks together, linked to the Saga.
func valigarmandaExile(ctx *Context, picked []uuid.UUID) error {
	if len(picked) == 0 {
		return nil
	}
	link, _ := ctx.SourceRef() // no link when the Saga has gone: nothing will read it
	return ctx.Game.ExileLinkedKeyedForEffect(link, picked, game.KeyedTail{})
}

// valigarmandaManaAndCast is chapters II-IV: the mana, then the offer.
func valigarmandaManaAndCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	if n := info.Counters[game.CounterLore]; n > 0 {
		if err := (AddMana{Amount: ManaN(n, "R")}).Apply(ctx); err != nil {
			return err
		}
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cards []game.Card
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if isInstantOrSorcery(*c) {
			cards = append(cards, *c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneExile,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		AnyType:     true,
		AnyColor:    true,
		Duration:    g.UntilEndOfTurnDuration(),
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  "Summon: Esper Valigarmanda",
		Label:       "Summon: Esper Valigarmanda — cast an exiled instant or sorcery card, spending mana of any type",
	}, cards)
	// #4250: cast during the chapter's resolution (CR 608.2g), paying for
	// it then with the mana just added; one permission, one cast.
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, c.InstanceID)
	}
	offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Summon: Esper Valigarmanda",
		game.MayCastKeywordPaidCast, declineLeaveInExile, ids...)
	return nil
}
