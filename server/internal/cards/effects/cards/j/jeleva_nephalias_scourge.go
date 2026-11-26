package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeleva, Nephalia's Scourge — Legendary Creature — Vampire Wizard {1}{U}{B}{R},
// 1/3 (Commander 2013, S82 OLD b386):
//
//	"Flying
//	 When Jeleva enters, each player exiles the top X cards of their library,
//	 where X is the amount of mana spent to cast Jeleva.
//	 Whenever Jeleva attacks, you may cast an instant or sorcery spell from
//	 among cards exiled with Jeleva without paying its mana cost."
//
// The entry trigger reads the mana that paid for the permanent's own spell
// (CR 400.7d), then every player's top X cards leave their library in one
// exile linked to this Jeleva object (CR 607.2a). The attack trigger offers
// the controller a free cast of ONE instant or sorcery among the cards linked
// to the Jeleva that attacked (a Jeleva that left and came back has no claim on
// them), during the trigger's resolution (CR 608.2g); declining leaves the cards
// in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a014f283-c531-415c-ac00-e6773ea5d64d",
		Name:            "Jeleva, Nephalia's Scourge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ReadsManaSpent:  game.ReadsOwnPayment,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jeleva, Nephalia's Scourge — each player exiles the top X cards of their library, where X is the amount of mana spent to cast Jeleva",
				jelevaExile),
			effects.WheneverThisAttacks("Jeleva, Nephalia's Scourge — you may cast an instant or sorcery spell from among cards exiled with Jeleva without paying its mana cost",
				jelevaCast),
		},
	})
}

func jelevaExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := ctx.ManaSpentToCastThis().Total()
	ref, ok := ctx.SourceRef()
	if x <= 0 || !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Library == nil {
			continue
		}
		n := min(x, len(p.Library.Cards))
		for i := 0; i < n; i++ {
			ids = append(ids, p.Library.Cards[len(p.Library.Cards)-1-i].InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileLinkedKeyedForEffect(ref, ids, game.KeyedTail{})
}

func jelevaCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cards []game.Card
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsInstant() || c.IsSorcery() {
			cards = append(cards, *c)
			ids = append(ids, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	controller := ctx.Controller()
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  "Jeleva, Nephalia's Scourge",
		Label:       "Cast an instant or sorcery spell exiled with Jeleva without paying its mana cost",
	}, cards)
	effects.OfferResolutionCastOver(g, controller, ctx.Source(), "Jeleva, Nephalia's Scourge", game.MayCastKeywordFreeCast, effects.DeclineLeaveInExile, ids...)
	return nil
}
