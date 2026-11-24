package cardsmi

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Muse Vortex — Sorcery {X}{U}{U} (Commander 2021, S82 OLD b058, #5504):
//
//	"Exile the top X cards of your library. You may cast an instant or
//	 sorcery spell with mana value X or less from among them without paying
//	 its mana cost. Then put the exiled instant and sorcery cards that
//	 weren't cast this way into your hand and the rest on the bottom of your
//	 library in a random order."
//
// The exile is a mill to exile, so a card a replacement sent elsewhere is
// not "among them". The tail then offers ONE free cast of an instant or
// sorcery spell with mana value X or less as part of the resolution (CR
// 608.2g; Portent of Calamity's OfferResolutionCastForEffect flow: timing
// restrictions ignored, an {X} in the cast spell's cost is 0, additional
// costs still owed). Once the offer is answered, whether or not a spell was
// cast, the instant and sorcery cards still in exile go to the hand and
// every other exiled card goes to the bottom of the library in a random
// order (PutOnBottomInRandomOrderForEffect). Every prompt is keyed plain
// data, so a table waiting on one is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b65dc8d-a8bf-4fa1-a44e-6a3782d18c0f",
		Name:         "Muse Vortex",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return effects.MillToZone{
				Player:   item.Controller,
				To:       game.ZoneExile,
				N:        x,
				ThenTail: effects.WithN(ctx.ItemTail(museVortexExiled), x),
			}.Apply(ctx)
		},
	})
}

// museVortexExiled is the continuation of the exile: it offers the free cast
// over the cards that reached exile, then sorts what is left. N is X. An
// on-disk key.
var museVortexExiled = effects.CardTail("muse-vortex/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, controller, source := ctx.Game, ctx.Controller(), ctx.Source()
	exiled := effects.AbstractPerformanceStillExiled(g, o.IDs)
	if len(exiled) == 0 {
		return nil
	}
	x := o.Tail.N
	perm := game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Filter:      game.PermissionFilter{InstantOrSorceryOnly: true},
		Source:      source,
		SourceName:  "Muse Vortex",
		Label:       "Muse Vortex — cast an instant or sorcery spell with mana value " + strconv.Itoa(x) + " or less without paying its mana cost",
	}
	perm.MaxSpellManaValue = &x
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Muse Vortex",
		Cards:        exiled,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put the instants and sorceries into your hand",
		Then:         museVortexRest,
		Carry:        exiled,
	}) {
		return nil
	}
	return museVortexSort(g, controller, source, exiled)
})

// museVortexRest is "then put the exiled instant and sorcery cards that
// weren't cast into your hand and the rest on the bottom", once the cast is
// answered: Carry is every exiled card. An on-disk key.
var museVortexRest = game.RegisterChooseCardsThen("resolution-cast/muse-vortex-rest",
	func(g *game.Game, r game.CardsChosen) error {
		return museVortexSort(g, r.Chooser, r.Source, r.Carry)
	})

// museVortexSort puts the instant and sorcery cards still in exile into the
// hand and every other card still in exile on the bottom of the library in a
// random order.
func museVortexSort(g *game.Game, controller, source uuid.UUID, ids []uuid.UUID) error {
	var toHand, toBottom []uuid.UUID
	for _, id := range effects.AbstractPerformanceStillExiled(g, ids) {
		if c, ok := g.LookupCardForEffect(id); ok && (c.IsInstant() || c.IsSorcery()) {
			toHand = append(toHand, id)
		} else {
			toBottom = append(toBottom, id)
		}
	}
	if err := effects.GloriousPurposePutRest(g, controller, source, toHand); err != nil {
		return err
	}
	return g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneExile, toBottom)
}
