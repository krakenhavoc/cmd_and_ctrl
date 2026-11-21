package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nicol Bolas, God-Pharaoh — Legendary Planeswalker — Bolas {4}{U}{B}{R},
// loyalty 7:
//
//	"+2: Target opponent exiles cards from the top of their library
//	     until they exile a nonland card. Until end of turn, you may
//	     cast that card without paying its mana cost.
//	 +1: Each opponent exiles two cards from their hand.
//	 −4: Nicol Bolas deals 7 damage to target opponent, creature an
//	     opponent controls, or planeswalker an opponent controls.
//	 −12: Exile each nonland permanent your opponents control."
//
// Printed loyalty reaches the card through deck import (ADR 0032 §1), so
// Spec.StartingLoyalty is not set.
//
//   - +2 is Fevered Suspicion's exile-until-a-nonland-card
//     (MillToZone with an Until clause, so each exile is its own
//     CR 614 window), with the hit given a free-cast grant (free_cast_
//     grants.go) that lasts the whole turn: it is cast in an ordinary
//     main phase under the card's own timing, so a sorcery is still
//     cast at sorcery speed. Lands exiled on the way stay exiled, and a
//     library with no nonland card exiles itself out and grants nothing.
//   - +1 makes EACH opponent exile two cards of THEIR choice from hand,
//     as an exile (not a discard, so no discard trigger and nothing in a
//     graveyard for a recursion deck). An opponent with one card exiles
//     it and one with none exiles nothing. Each opponent gets their
//     own prompt, queued together.
//   - −4 is All Will Be One's target clause (an opponent, a creature an
//     opponent controls or a planeswalker an opponent controls) with a
//     fixed 7, dealt by Bolas.
//   - −12 exiles every nonland permanent the controller's opponents
//     control, simultaneously (ExileAllMatching): no dies triggers and
//     indestructible does not matter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb33e07b-9599-4979-991f-df0c43ddac31",
		Name:         "Nicol Bolas, God-Pharaoh",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Target opponent exiles cards from the top of their library until they exile a nonland card. Until end of turn, you may cast that card without paying its mana cost.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect:  nicolBolasPlusTwo,
			},
			{
				Label:  "+1: Each opponent exiles two cards from their hand.",
				Cost:   effects.LoyaltyCost(1),
				Effect: nicolBolasPlusOne,
			},
			{
				Label:   "−4: Nicol Bolas deals 7 damage to target opponent, creature an opponent controls, or planeswalker an opponent controls.",
				Cost:    effects.LoyaltyCost(-4),
				Targets: effects.B12TargetOpponentOrTheirCreatureOrPlaneswalker(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 7)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 7}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−12: Exile each nonland permanent your opponents control.",
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile, OpponentsOnly: true}},
				Cost:    effects.LoyaltyCost(-12),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileAllMatching{Match: effects.And(effects.Nonland(), effects.OpponentControls())}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// nicolBolasPlusTwo exiles from the target opponent's library until a
// nonland card, and lets the controller cast that card for free.
func nicolBolasPlusTwo(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return effects.MillToZone{
		Player:   target,
		To:       game.ZoneExile,
		Until:    effects.UntilNonland,
		ThenTail: ctx.ItemTail(nicolBolasGodPharaohMilled),
	}.Apply(ctx)
}

// nicolBolasPlusOne asks each opponent to exile two cards from their
// hand (all of them, with fewer than two).
func nicolBolasPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Eliminated || p.Hand == nil || p.Hand.Size() == 0 {
			continue
		}
		hand := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
		n := min(2, len(hand))
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  opp,
			Source:   item.SourceCardID,
			Question: "Nicol Bolas, God-Pharaoh — exile two cards from your hand",
			Cards:    hand,
			Min:      n,
			Max:      n,
			LiveZone: game.ZoneHand,
			ThenKey:  nicolBolasExileTwo,
		})
	}
	return nil
}

// nicolBolasGodPharaohMilled is the clause after the mill (#4253 slice 9).
var nicolBolasGodPharaohMilled = effects.CardTail("nicol-bolas-god-pharaoh/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	controller, source := ctx.Controller(), ctx.Source()
	exiled := o.IDs

	var hits []uuid.UUID
	for _, id := range exiled {
		if c, found := ctx.Game.LookupCardForEffect(id); found && !c.IsLand() {
			hits = append(hits, id)
		}
	}
	effects.GrantFreeCasts(ctx.Game, controller, source, "Nicol Bolas, God-Pharaoh", game.TimingNormal, "", hits)
	return nil
})

// nicolBolasExileTwo is the two cards exiled from an opponent's hand, keyed
// (#4253). It reads nothing but the pick.
var nicolBolasExileTwo = game.RegisterChooseCardsThen("choose-cards/nicol-bolas-god-pharaoh-exile-two", func(g *game.Game, r game.CardsChosen) error {
	g.ExileCardsForEffect(r.Picked)
	return nil
})
