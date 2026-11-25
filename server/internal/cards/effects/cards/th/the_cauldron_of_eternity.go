package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Cauldron of Eternity — Legendary Artifact {10}{B}{B} (Throne of
// Eldraine, S82 old sets):
//
//	"This spell costs {2} less to cast for each creature card in your
//	 graveyard.
//	 Whenever a creature you control dies, put it on the bottom of its owner's
//	 library.
//	 {2}{B}, {T}, Pay 2 life: Return target creature card from your graveyard
//	 to the battlefield. Activate only as a sorcery."
//
// The reduction is Writhing Necromass's self cost modifier doubled: two less
// of the generic part for each creature card in the caster's graveyard. The
// trigger reads the creature as it last existed (CR 603.10a: it was a
// creature you controlled) and puts the card that went to the graveyard on the
// bottom of its owner's library through the shared route (a token has ceased
// to exist, a commander's owner is offered the command zone, CR 903.9); a card
// that has already left the graveyard is left alone. The ability is Rise
// Again's reanimation with a tap, mana and life cost at sorcery speed.
//
// No simplification.
func init() {
	creatures := effects.CountCardsInYourGraveyard(effects.Creature())
	effects.Register(effects.Spec{
		OracleID:     "a7308a1f-ec46-478e-86ea-33b2ebbd995f",
		Name:         "The Cauldron of Eternity",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return 2 * creatures(q) },
				"This spell costs {2} less to cast for each creature card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller
			}, "The Cauldron of Eternity — put the creature on the bottom of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.CardID
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.PutIntoLibrary{Card: id, ToBottom: true}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}, {T}, Pay 2 life: Return target creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.Reanimates(0)),
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(), effects.PayLife(2)),
			SorcerySpeed: true,
			Targets:      effects.TargetCreatureInYourGraveyard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, _ = effects.ReanimateSingleTarget(effects.NewContext(g, item), item.Controller)
				return nil
			},
		}},
	})
}
