package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helmut Zemo, Mastermind — Legendary Creature — Human Noble Villain
// {2}{U}, 2/2 (Marvel Super Heroes Commander):
//
//	"Whenever Helmut Zemo attacks, you may cast target instant or sorcery
//	 card with mana value less than or equal to his power from your
//	 graveyard. If that spell would be put into your graveyard, exile it
//	 instead. If you cast a spell this way, put a +1/+1 counter on Helmut
//	 Zemo."
//
// Emet-Selch's offer (OfferGraveyardCast): the cast is made as the ability
// resolves, at the card's mana cost, and the spell is exiled if it would
// go to the graveyard. The power bound is read live when the legal set is
// computed and again on resolution (Carmen's technique), so a counter in
// response widens it. The +1/+1 counter is the offer's follow-up and runs
// only when the spell really is cast.
//
// No simplification.
const Msc6ZemoOracleID = "657c5473-f153-4dd2-94a0-d477cbc2451d"

var msc6ZemoCast = game.RegisterCastFollowUp("msc6/helmut-zemo-mastermind-counter",
	func(g *game.Game, f game.CastFollowUp) error {
		if !effects.OnBattlefield(g, f.Source) {
			return nil
		}
		return g.AddCounterForEffect(f.Source, game.CounterPlusOne, 1)
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:     Msc6ZemoOracleID,
		Name:         "Helmut Zemo, Mastermind",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.Msc6ThisAttacks,
					"Helmut Zemo — you may cast target instant or sorcery card from your graveyard",
					msc6ZemoOffer),
				effects.TargetCardInGraveyard("target instant or sorcery card with mana value less than or equal to Helmut Zemo's power in your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()),
					func(g *game.Game, caster uuid.UUID, c game.Card) bool {
						return c.ManaValue() <= effects.B26GreatestPowerOfOracleYouControl(g, caster, Msc6ZemoOracleID)
					})),
		},
	})
}

func msc6ZemoOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.OfferGraveyardCast{Target: t.ID, SourceName: "Helmut Zemo, Mastermind", FollowUp: msc6ZemoCast}.Apply(ctx)
	}
	return nil
}
