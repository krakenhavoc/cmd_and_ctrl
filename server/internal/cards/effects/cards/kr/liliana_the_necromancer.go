package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, the Necromancer — Legendary Planeswalker — Liliana {3}{B}{B},
// loyalty 4 (Core Set 2019):
//
//	"+1: Target player loses 2 life.
//	 −1: Return target creature card from your graveyard to your hand.
//	 −7: Destroy up to two target creatures. Put up to two creature cards from
//	 graveyards onto the battlefield under your control."
//
// The +1 is Bloodhunter Bat's life loss. The −1 targets a creature card in the
// controller's graveyard and returns it to hand if still there (CR 608.2b).
// The −7 destroys the still-legal targets as one event, then asks for up to two
// creature cards from every graveyard as it resolves (CR 608.2c, Extract from
// Darkness's pool: a creature just destroyed is a legal pick, nothing is
// targeted); the picks enter as one simultaneous entry under the controller's
// control, whoever owned them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f3ad60a6-05d0-48a1-a804-7900d306d613",
		Name:            "Liliana, the Necromancer",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Target player loses 2 life.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 2}),
			},
			{
				Label:   "−1: Return target creature card from your graveyard to your hand.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				Effect:  effects.ReturnTargetCardToHand,
			},
			{
				Label:   "−7: Destroy up to two target creatures. Put up to two creature cards from graveyards onto the battlefield under your control.",
				Cost:    effects.LoyaltyCost(-7),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := effects.DestroyEachLegalTarget(item, ctx); err != nil {
						return err
					}
					candidates := effects.AllGraveyardsCardIDs(ctx, game.Card.IsCreature)
					if len(candidates) == 0 {
						return nil
					}
					max := 2
					if len(candidates) < max {
						max = len(candidates)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:     ctx.Controller(),
						Source:      item.SourceCardID,
						Question:    "Liliana, the Necromancer — put up to two creature cards from graveyards onto the battlefield under your control",
						Cards:       candidates,
						Min:         0,
						Max:         max,
						Destination: game.ChooseOntoBattlefield,
						ThenKey:     lilianaNecromancerPicked,
						Asked:       game.NewAskedItem(item),
					})
					return nil
				},
			},
		},
	})
}

// lilianaNecromancerPicked puts the picked creature cards onto the battlefield
// under the controller's control as one entry. An on-disk key: never renamed,
// never reused.
var lilianaNecromancerPicked = effects.ItemChoice("choose-cards/liliana-the-necromancer-reanimate", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	entries := make([]game.BatchEntry, 0, len(r.Picked))
	for _, id := range r.Picked {
		if id == uuid.Nil {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneGraveyard})
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{Controller: ctx.Controller()}, game.KeyedTail{})
})
