package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worldsoul's Rage — Sorcery {X}{R}{G} (Murders at Karlov Manor, #4469):
//
//	"Worldsoul's Rage deals X damage to any target. Put up to X land cards
//	 from your hand and/or graveyard onto the battlefield tapped."
//
// The damage is the first instruction; the lands are a pick made after it,
// a choose_cards prompt over every land card in the caster's hand and
// graveyard with a ceiling of X in all, not X from each zone. The chosen
// lands enter together, tapped, each from the zone it is in when the pick is
// answered (a card that left in the meantime is skipped). The pick is
// keyed, so a table waiting on it is a restore point. With X = 0 nothing is
// asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b6ef144-bb98-4686-9718-204f1c3cf020",
		Name:         "Worldsoul's Rage",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if t, ok := ctx.ClauseTarget(0); ok && x > 0 {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}).Apply(ctx); err != nil {
					return err
				}
			}
			if x <= 0 {
				return nil
			}
			me := ctx.PlayerByID(ctx.Controller())
			if me == nil {
				return nil
			}
			var lands []uuid.UUID
			for _, z := range []*game.Zone{me.Hand, me.Graveyard} {
				if z == nil {
					continue
				}
				for _, c := range z.Cards {
					if c.IsLand() {
						lands = append(lands, c.InstanceID)
					}
				}
			}
			if len(lands) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Worldsoul's Rage — put up to X land cards from your hand and/or graveyard onto the battlefield tapped",
				Cards:    lands,
				Min:      0,
				Max:      x,
				ThenKey:  worldsoulsRageLands,
			})
			return nil
		},
	})
}

// worldsoulsRageLands is the pick answered: the lands enter together,
// tapped, under the caster's control, each from the zone it is in now.
var worldsoulsRageLands = game.RegisterChooseCardsThen("choose-cards/worldsouls-rage-lands", func(g *game.Game, r game.CardsChosen) error {
	var entries []game.BatchEntry
	for _, id := range r.Picked {
		z := g.FindCardZoneForEffect(id)
		if z == nil || (z.Kind != game.ZoneHand && z.Kind != game.ZoneGraveyard) {
			continue
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsLand() || c.Owner != r.Chooser {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: z.Kind})
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{Controller: r.Chooser, Tapped: true}, game.KeyedTail{})
})
