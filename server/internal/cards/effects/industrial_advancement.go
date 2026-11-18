package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Industrial Advancement — Enchantment {3}{R} (Streets of New Capenna
// Commander, S82):
//
//	"At the beginning of your end step, you may sacrifice a creature. If
//	 you do, look at the top X cards of your library, where X is that
//	 creature's mana value. You may put a creature card from among them
//	 onto the battlefield. Put the rest on the bottom of your library in a
//	 random order."
//
// Descendants' Fury's offer: a pick of which creature to sacrifice, or none
// (a "no" ends the ability). The mana value is read off the creature before
// it goes and rides the sacrifice's tail as N, so "that creature's mana
// value" is the value it had as it was sacrificed (CR 608.2h; a token has
// 0, and X is 0 off the stack). "If you do" is the tail's: it runs only when
// the creature really left. The looked-at cards then follow Kinnan, Bonder
// Prodigy's shape: an optional creature card onto the battlefield and the
// rest to the bottom in a random order, whether or not one was put.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a0aa9ea-610e-4910-b66d-fdac69e88adf",
		Name:         "Industrial Advancement",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Industrial Advancement — you may sacrifice a creature; if you do, look at the top X cards and put a creature card among them onto the battlefield",
				industrialAdvancementOffer),
		},
	})
}

// industrialAdvancementOffer asks which creature to sacrifice (nothing is a
// legal answer).
func industrialAdvancementOffer(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && !g.CantBeSacrificedForEffect(c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     ctx.Source(),
		Question:   "Industrial Advancement — you may sacrifice a creature",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    industrialAdvancementSacrificePicked,
	})
	return nil
}

// industrialAdvancementSacrificePicked sacrifices the pick, carrying its
// mana value to the look. An on-disk key: never renamed, never reused.
var industrialAdvancementSacrificePicked = game.RegisterChooseCardsThen("choose-cards/industrial-advancement-sacrifice",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		c, ok := g.LookupCardForEffect(r.Picked[0])
		if !ok {
			return nil
		}
		return g.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked, game.KeyedTail{
			Ref: industrialAdvancementSacrificed, Player: r.Chooser, Source: r.Source, N: c.ManaValue(),
		})
	})

// industrialAdvancementSacrificed is "if you do": the creature really went,
// so look at the top X cards. An on-disk key: never renamed, never reused.
var industrialAdvancementSacrificed = CardTail("then/industrial-advancement-sacrificed", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	cards := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), o.Tail.N)
	if len(cards) == 0 {
		return nil
	}
	return PutFromLibraryOntoBattlefield{
		Player:   ctx.Controller(),
		Cards:    cards,
		Match:    Creature(),
		Max:      1,
		Optional: true,
		Label:    "Industrial Advancement — you may put a creature card onto the battlefield",
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
})
