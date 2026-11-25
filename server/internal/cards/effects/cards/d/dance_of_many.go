package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dance of Many — Enchantment {U}{U} (Fallen Empires):
//
//	"When this enchantment enters, create a token that's a copy of target
//	 nontoken creature.
//	 When this enchantment leaves the battlefield, exile the token.
//	 When the token leaves the battlefield, sacrifice this enchantment.
//	 At the beginning of your upkeep, sacrifice this enchantment unless you
//	 pay {U}{U}."
//
// The three first abilities are linked (CR 607.2): the token the enters
// trigger made is stored on the enchantment as its linked permanent, an
// object with its CR 400.7 epoch (Card.LinkedPermanent, carried by the
// snapshot), as the creation's keyed tail settles. The enchantment's leave
// trigger reads the link from the enchantment as it last existed
// (CR 608.2h) and exiles that token if it is still on the battlefield. The
// third ability is the enchantment's own trigger on the token's leaving
// the battlefield, and sacrifices the enchantment only while it is still
// the object that made the token. If the enchantment has already left when
// its enters trigger resolves, the token is still made, linked to nothing,
// and stays. The upkeep payment is the shared pay-or-sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72e1a239-11a7-4dbe-b093-65d395e3ccc9",
		Name:         "Dance of Many",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Dance of Many — create a token that's a copy of target nontoken creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.CreateTokenCopiesTogether{
						Controller: ctx.Controller(),
						Copies:     []uuid.UUID{id},
						Then:       ctx.ItemTail(danceOfManyLinkToken),
					}.Apply(ctx)
				}),
				effects.TargetCreature("target nontoken creature", effects.Not(effects.IsTokenPredicate()))),
			effects.WhenThisLeaves("Dance of Many — exile the token", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				token := g.LinkedPermanentForEffect(ref)
				if token == nil {
					return nil
				}
				return effects.ExileTarget{Target: token.InstanceID}.Apply(ctx)
			}),
			effects.On(game.EventLTB, danceOfManyTokenLeft, "Dance of Many — the token left the battlefield; sacrifice this enchantment",
				effects.SacrificeThisIfStillOnBattlefield),
			effects.UpkeepSacrificeUnlessYouPay("Dance of Many", "{U}{U}"),
		},
	})
}

// danceOfManyLinkToken links the token the creation made to the
// enchantment, while the enchantment is still the object whose trigger
// made it (CR 400.7).
var danceOfManyLinkToken = effects.CardTail("dance-of-many/link-token", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	src, ok := ctx.SourcePermanent()
	if !ok || src.Left {
		return nil
	}
	ref, _ := ctx.SourceRef()
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID == o.IDs[0] {
			ctx.Game.SetLinkedPermanentForEffect(ref.ID, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
		}
	}
	return nil
})

// danceOfManyTokenLeft is "when the token leaves the battlefield": the
// permanent leaving is the one this enchantment is linked to.
func danceOfManyTokenLeft(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if source == nil || ev.CardID == uuid.Nil {
		return false
	}
	return source.LinkedPermanent.ID == ev.CardID
}
