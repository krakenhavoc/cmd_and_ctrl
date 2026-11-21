package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk, Apex Predator — Legendary Planeswalker — Garruk {5}{B}{G}, loyalty
// 5 (Modern Horizons 3 Commander, #5948):
//
//	"+1: Destroy another target planeswalker.
//	 +1: Create a 3/3 black Beast creature token with deathtouch.
//	 −3: Destroy target creature. You gain life equal to its toughness.
//	 −8: Target opponent gets an emblem with "Whenever a creature attacks
//	 you, it gets +5/+5 and gains trample until end of turn.""
//
// The first +1 targets a planeswalker other than Garruk himself (CR
// 115.7's "another" is object identity). The −3 reads the creature's
// toughness as it was before it was destroyed (CR 608.2h) and gains the life
// whether or not the creature was destroyed (an indestructible one still
// gives it): the second sentence is not an "if you do". The −8 gives an
// opponent the emblem (Ob Nixilis Reignited's shape): it triggers for each
// creature that attacks that player (not one of their planeswalkers), and the
// attacker gets +5/+5 and trample until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ef9a8bdc-3361-440d-89f6-0bf5ff7a09e8",
		Name:            "Garruk, Apex Predator",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &EmblemSpec{
			Label: "Garruk, Apex Predator emblem",
			Text:  "Whenever a creature attacks you, it gets +5/+5 and gains trample until end of turn.",
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor != uuid.Nil && ev.Target == source.Controller
				},
				Key: "Garruk, Apex Predator emblem — the attacker gets +5/+5 and trample",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Garruk, Apex Predator emblem — the attacker gets +5/+5 and trample")
					item.Params.Object = game.ObjectRef{ID: ev.CardID}
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					attacker := item.Params.Object.ID
					if !onBattlefield(g, attacker) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Target: attacker, Power: 5, Toughness: 5, Label: "Garruk, Apex Predator emblem — +5/+5"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"trample"}, Label: "Garruk, Apex Predator emblem — trample"}.Apply(ctx)
				},
			}},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Destroy another target planeswalker.",
				Cost:    LoyaltyCost(1),
				Targets: Another(TargetPermanent("another target planeswalker", Planeswalker())),
				Effect:  destroyFirstLegalCardTarget,
			},
			{
				Label:   "+1: Create a 3/3 black Beast creature token with deathtouch.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("3/3 black Beast with deathtouch"), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−3: Destroy target creature. You gain life equal to its toughness.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCreature("target creature"),
				Effect:  garrukApexPredatorDestroy,
			},
			{
				Label:   "−8: Target opponent gets an emblem with \"Whenever a creature attacks you, it gets +5/+5 and gains trample until end of turn.\"",
				Cost:    LoyaltyCost(-8),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect:  targetOpponentGetsAnEmblem,
			},
		},
	})
}

// garrukApexPredatorDestroy is the −3: destroy the creature, then gain life
// equal to its toughness as it was before it was destroyed (CR 608.2h).
func garrukApexPredatorDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID}, withN(ctx.Tail(garrukApexPredatorGainLife), c.CurrentToughness()))
	}
	return nil
}
