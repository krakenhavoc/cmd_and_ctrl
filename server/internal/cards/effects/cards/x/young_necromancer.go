package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Young Necromancer — Creature — Human Warlock {4}{B}, 2/3 (Modern Horizons
// 2, #5501):
//
//	"When this creature enters, you may exile two cards from your graveyard.
//	 When you do, return target creature card from your graveyard to the
//	 battlefield."
//
// The "may" is asked as the enters trigger resolves (CR 603.5); a graveyard of
// fewer than two cards cannot pay it. The two cards are chosen from the
// graveyard and exiled, and only then does the reflexive trigger go on the
// stack and choose its target (CR 603.12), so the cards just exiled are not
// candidates and a creature card still in the graveyard is. The return is
// under your control and does nothing if the card has left (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17455491-c295-4130-99b4-1b5c6ac1cd25",
		Name:         "Young Necromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Young Necromancer — you may exile two cards from your graveyard; when you do, return target creature card from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Graveyard == nil || p.Graveyard.Size() < 2 {
						return nil
					}
					ids := make([]uuid.UUID, 0, p.Graveyard.Size())
					for _, c := range p.Graveyard.Cards {
						ids = append(ids, c.InstanceID)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:    item.Controller,
						FromPlayer: item.Controller,
						Source:     item.SourceCardID,
						Question:   "Young Necromancer — exile two cards from your graveyard",
						Cards:      ids,
						Min:        2,
						Max:        2,
						LiveZone:   game.ZoneGraveyard,
						ThenKey:    youngNecromancerExiled,
						Asked:      game.NewAskedItem(item),
					})
					return nil
				}),
				"Young Necromancer — exile two cards from your graveyard?"),
		},
	})
}

// youngNecromancerExiled is the pick answered: exile the two cards, then the
// reflexive trigger. An on-disk key: never renamed, never reused.
var youngNecromancerExiled = effects.AskedPick("choose-cards/young-necromancer-exile-two", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 2 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(youngNecromancerReflexive))
})

// youngNecromancerReflexive is the exile's continuation: when both cards
// reached exile, the reflexive trigger. An on-disk key: never renamed, never
// reused.
var youngNecromancerReflexive = effects.CardTail("then/young-necromancer-when-you-do", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 2 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label: "Young Necromancer — return target creature card from your graveyard to the battlefield",
		Cards: o.IDs,
		Body:  youngNecromancerReturnBody,
	}.Apply(ctx)
})

var youngNecromancerReturnBody = game.ReflexiveBody("young-necromancer/return", effects.SimpleBody(
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
		}
		return nil
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())
	}))
