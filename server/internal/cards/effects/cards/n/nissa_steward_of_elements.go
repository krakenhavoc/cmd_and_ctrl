package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa, Steward of Elements — Legendary Planeswalker — Nissa {X}{G}{U},
// loyalty X (Modern Horizons 3 Commander, #5948):
//
//	"+2: Scry 2.
//	 0: Look at the top card of your library. If it's a land card or a
//	 creature card with mana value less than or equal to the number of
//	 loyalty counters on Nissa, you may put that card onto the battlefield.
//	 −6: Untap up to two target lands you control. They become 5/5
//	 Elemental creatures with flying and haste until end of turn. They're
//	 still lands."
//
// The printed loyalty is X: the announced X of the cast becomes that many
// loyalty counters as she enters (CR 306.5b, XCounters(loyalty)), and with
// X = 0 she enters with none and is put into the graveyard by the next
// state-based action (CR 704.5i). The 0 reads Nissa's loyalty counters as it
// resolves (her last-known count if she has left), and offers the top card
// only if it is a land or a creature whose mana value fits: otherwise the
// card stays on top. The −6 untaps the lands and then makes each a 5/5
// Elemental creature with flying and haste in addition to being a land,
// until end of turn (CR 611.2c: the affected set is fixed as it resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "fdb96b5c-ecac-4db3-8b9a-24f824149f91",
		Name:                       "Nissa, Steward of Elements",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterLoyalty)},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Scry 2.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Scry: 2},
				Effect:  effects.Do(effects.Scry{N: 2}),
			},
			{
				Label:  "0: Look at the top card of your library. If it's a land card or a creature card with mana value less than or equal to the number of loyalty counters on Nissa, you may put that card onto the battlefield.",
				Cost:   effects.LoyaltyCost(0),
				Effect: nissaStewardOfElementsLook,
			},
			{
				Label:   "−6: Untap up to two target lands you control. They become 5/5 Elemental creatures with flying and haste until end of turn. They're still lands.",
				Cost:    effects.LoyaltyCost(-6),
				Targets: effects.TargetPermanent("up to two target lands you control", effects.Land(), effects.YouControl()).WithCount(0, 2),
				Effect:  nissaStewardOfElementsAnimate,
			},
		},
	})
}

func nissaStewardOfElementsLook(g *game.Game, item *game.StackItem) error {
	loyalty := 0
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		loyalty = c.Counters[game.CounterLoyalty]
	}
	looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(looked) == 0 {
		return nil
	}
	return effects.PutFromLibraryOntoBattlefield{
		Player:   item.Controller,
		Cards:    looked,
		Match:    effects.Or(effects.Land(), effects.And(effects.Creature(), effects.ManaValueLE(loyalty))),
		Max:      1,
		Optional: true,
		Label:    "Nissa, Steward of Elements — you may put the card onto the battlefield",
	}.Apply(effects.NewContext(g, item))
}

func nissaStewardOfElementsAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var lands []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			lands = append(lands, t.ID)
		}
	}
	for _, id := range lands {
		if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range lands {
		if err := (effects.BecomeCreatureUntilEOT{
			Target:       id,
			Types:        []string{"Creature"},
			Subtypes:     []string{"Elemental"},
			SetPower:     5,
			SetToughness: 5,
			Label:        "Nissa, Steward of Elements — becomes a 5/5 Elemental creature",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "haste"}, Label: "Nissa, Steward of Elements — flying and haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
