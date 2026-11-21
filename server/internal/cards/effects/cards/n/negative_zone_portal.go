package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Negative Zone Portal — Artifact {2} (Marvel Super Heroes Commander,
// #4455):
//
//	"{2}, {T}: Exile target card from an opponent's graveyard. If it's a
//	 creature card, draw a card.
//	 At the beginning of your upkeep, if there are four or more creature
//	 cards exiled with this artifact, flip a coin. If you lose the flip,
//	 sacrifice this artifact and return a card exiled with it at random
//	 to its owner's hand."
//
// The exile is linked to the Portal (CR 607.2a: ExileTarget.Linked), so
// "exiled with this artifact" is the cards in exile that carry that
// link. Whether the card was a creature card is read before it leaves
// the graveyard, and the draw happens only if the card actually landed in
// exile (the exile's keyed tail). The upkeep trigger has an intervening
// "if" (CR 603.4): four or more CREATURE cards exiled with this Portal,
// checked as the upkeep begins and again as the trigger resolves.
//
// The coin is called like every flip in the engine; losing it sacrifices
// the Portal and returns one card at random from everything exiled with
// it (any card type, not only creatures), chosen from the set the Portal
// had when the flip was made. A Portal that left the battlefield or came
// back as a new object is not sacrificed again, but the card still
// returns.
//
// No simplification.
func init() {
	const upkeepLabel = "Negative Zone Portal — flip a coin"
	effects.Register(effects.Spec{
		OracleID:     "8430ae03-2d4c-4031-aa29-4d53815ded78",
		Name:         "Negative Zone Portal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Exile target card from an opponent's graveyard. If it's a creature card, draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn())),
			Effect:  negativeZonePortalExile,
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "four or more creature cards are exiled with this artifact",
				Test: func(ctx *effects.Context) bool { return negativeZonePortalCreatureCount(ctx) >= 4 },
			}, effects.AtYourUpkeep(upkeepLabel, negativeZonePortalFlip)),
		},
	})
}

// negativeZonePortalLinked lists the cards exiled with this Portal object.
func negativeZonePortalLinked(ctx *effects.Context) []*game.Card {
	c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return nil
	}
	return ctx.Game.ExiledWithForEffect(game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
}

func negativeZonePortalCreatureCount(ctx *effects.Context) int {
	n := 0
	for _, c := range negativeZonePortalLinked(ctx) {
		if c.IsCreature() {
			n++
		}
	}
	return n
}

func negativeZonePortalExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		return effects.ExileTarget{Target: t.ID, Linked: true,
			ThenTail: effects.WithN(ctx.ItemTail(negativeZonePortalDraw), effects.BoolToInt(c.IsCreature()))}.Apply(ctx)
	}
	return nil
}

// negativeZonePortalDraw is "if it's a creature card, draw a card": N
// says it was one, the outcome says it reached exile.
var negativeZonePortalDraw = effects.CardTail("negative-zone-portal/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N != 1 || len(o.IDs) != 1 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})

func negativeZonePortalFlip(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var carry []uuid.UUID
	for _, c := range negativeZonePortalLinked(ctx) {
		carry = append(carry, c.InstanceID)
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Negative Zone Portal — call the coin flip",
		ThenKey:  negativeZonePortalFlipped,
		Carry:    carry,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// negativeZonePortalFlipped is the settled call: a lost flip sacrifices
// the Portal and returns one of the cards exiled with it, at random, to
// its owner's hand. An on-disk key: never renamed, never reused.
var negativeZonePortalFlipped = effects.CoinFlipThen("coin-flip/negative-zone-portal", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return nil
	}
	if err := effects.SacrificeThisIfStillHere(ctx); err != nil {
		return err
	}
	var still []uuid.UUID
	for _, id := range r.Carry {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			still = append(still, id)
		}
	}
	for _, id := range ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), still, 1) {
		return effects.BounceToHand{Target: id}.Apply(ctx)
	}
	return nil
})
