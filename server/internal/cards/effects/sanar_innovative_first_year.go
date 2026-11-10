package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanar, Innovative First-Year — Legendary Creature — Goblin Sorcerer
// {2}{U/R}{U/R} (Lorwyn Eclipsed, #4458):
//
//	"Vivid — At the beginning of your first main phase, reveal cards from
//	 the top of your library until you reveal X nonland cards, where X is
//	 the number of colors among permanents you control. For each of those
//	 colors, you may exile a card of that color from among the revealed
//	 cards. Then shuffle. You may cast the exiled cards this turn."
//
// X is counted once, as the ability resolves (the ruling), and "those
// colors" are the colors counted. The revealed cards stay on top of the
// library (revealing moves nothing); one choose_cards prompt per color, in
// WUBRG order, offers the revealed cards of that color still in the library
// (a card of two of the colors may be exiled for either, once), and the
// pile is shuffled before the cast permission is granted. The cast is an
// ordinary one: its costs and timing apply (the ruling). Lands have no color
// and are never offered.
func init() {
	Register(Spec{
		OracleID:     "2676ba6d-e7a7-48cf-9c86-6e2a11f03096",
		Name:         "Sanar, Innovative First-Year",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Sanar, Innovative First-Year — reveal until X nonland cards; for each color among your permanents, you may exile a card of that color; then shuffle; you may cast the exiled cards this turn (vivid)",
				sanarResolve),
		},
	})
}

func sanarResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	colors := colorsAmongPermanentsYouControl(g, item.Controller)
	x := len(colors)
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Library == nil {
		return nil
	}
	// Walk the library from the top (the last element) until X nonland cards.
	n, seen := 0, 0
	for i := len(p.Library.Cards) - 1; i >= 0 && seen < x; i-- {
		n++
		if !p.Library.Cards[i].IsLand() {
			seen++
		}
	}
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, n, "Sanar, Innovative First-Year — revealed cards")
	idx := make([]int, 0, len(colors))
	for _, col := range colors {
		for i, c := range game.AllColors {
			if c == col {
				idx = append(idx, i)
			}
		}
	}
	return sanarAsk(ctx, revealed, idx)
}

// sanarAsk asks about the next color still to go, or finishes when none is.
func sanarAsk(ctx *Context, revealed []uuid.UUID, idx []int) error {
	g := ctx.Game
	for len(idx) > 0 {
		col := game.AllColors[idx[0]]
		rest := append([]int(nil), idx[1:]...)
		var cands []uuid.UUID
		for _, id := range revealed {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
				continue
			}
			if c, ok := g.LookupCardForEffect(id); ok && c.HasColor(col) {
				cands = append(cands, id)
			}
		}
		if len(cands) == 0 {
			idx = rest
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: ctx.Controller(),
			Source:     ctx.Source(),
			Question:   "Sanar, Innovative First-Year — you may exile a " + game.ColorName(col) + " card from among the revealed cards",
			Cards:      cands,
			Min:        0,
			Max:        1,
			LiveZone:   game.ZoneLibrary,
			Carry:      revealed,
			CarryNums:  rest,
			ThenKey:    sanarPick,
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return sanarFinish(ctx, revealed)
}

// sanarPick is the keyed continuation of each color's prompt. It is set in
// init, not as a variable's initializer, because it asks about the next color
// through sanarAsk, which names it. An on-disk key: never renamed, never reused.
var sanarPick game.ChooseCardsThen

func init() {
	sanarPick = askedPick("choose-cards/sanar-innovative-first-year-exile", func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) > 0 {
			if err := (ExileTarget{Target: r.Picked[0]}).Apply(ctx); err != nil {
				return err
			}
		}
		return sanarAsk(ctx, r.Carry, r.Nums)
	})
}

// sanarFinish is "then shuffle. You may cast the exiled cards this turn".
func sanarFinish(ctx *Context, revealed []uuid.UUID) error {
	g := ctx.Game
	if err := (ShuffleLibrary{Player: ctx.Controller()}).Apply(ctx); err != nil {
		return err
	}
	var cards []game.Card
	for _, id := range revealed {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if ok && z != nil && z.Kind == game.ZoneExile {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   g.UntilEndOfTurnDuration(),
		CastOnly:   true,
		Source:     ctx.Source(),
		SourceName: "Sanar, Innovative First-Year",
		Label:      "Sanar, Innovative First-Year — cast the exiled cards this turn",
	}, cards)
	return nil
}
