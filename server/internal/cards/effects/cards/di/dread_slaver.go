package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dread Slaver — Creature — Zombie Horror {3}{B}{B}, 3/5 (Jumpstart 2022,
// #5948):
//
//	"Whenever a creature dealt damage by this creature this turn dies,
//	 return it to the battlefield under your control. That creature is a
//	 black Zombie in addition to its other colors and types."
//
// Zurgo Helmsmasher's trigger condition (the engine's record of which
// objects the source damaged this turn, CR 400.7) with Necroskitter's
// return of the dead card; the Zombie and black are added by Necromantic
// Selection's indefinite effect pinned to the returned object. A token that
// died is gone and returns nothing; a card that has left the graveyard
// since is not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8110f092-41b0-4e53-a7d7-f7e3cbc4a52a",
		Name:         "Dread Slaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Dread Slaver — return it to the battlefield under your control as a black Zombie in addition", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				dead := ctx.Trigger().Event.CardID
				if !effects.DiedCardStillInGraveyard(ctx, dead) {
					return nil
				}
				return effects.ReturnFromGraveyardTogether{
					Targets:    []uuid.UUID{dead},
					Controller: item.Controller,
					ThenTail:   dreadSlaverBlackZombie,
				}.Apply(ctx)
			}),
		},
	})
}

var dreadSlaverBlackZombie = effects.TogetherTail("dread-slaver/black-zombie",
	func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			if err := (effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Zombie")...), game.AddColorsMod(ctx.TextColors("B")...)},
				Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
				Label:    "Dread Slaver — it is a black Zombie in addition to its other colors and types",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
