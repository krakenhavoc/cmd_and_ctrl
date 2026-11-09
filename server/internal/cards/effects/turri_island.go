package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turri Island — Plane — Ir (#4870):
//
//	"Creature spells cost {2} less to cast.
//	 Whenever chaos ensues, reveal the top three cards of your library. Put all
//	 creature cards revealed this way into your hand and the rest into your
//	 graveyard."
//
// The reduction applies to every player's creature spells and, as the ruling
// says, never takes the cost below {0} (CostsLess only removes generic mana).
// The chaos ability reveals, takes every creature card with no choice, and
// puts the rest into the graveyard (a put, not a mill: TakeRestIntoGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1300b5e7-d741-4723-bb15-dabd04cb0c11",
		Name:         "Turri Island",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Creature spells cost {2} less to cast.", CreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Turri Island — reveal the top three cards of your library; put all creature cards revealed this way into your hand and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TakeFromLibraryToHand{
						Player: item.Controller,
						Cards:  g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 3, "Turri Island — the top three cards of your library"),
						Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() },
						All:    true,
						Label:  "Turri Island — the top three cards of your library",
						Key:    turriIslandTake,
					}.Apply(ctx)
				}),
		},
	})
}

// turriIslandTake is the registered rest: everything not taken goes into the
// graveyard. An on-disk key: never renamed, never reused.
var turriIslandTake = RegisterTake("choose-cards/turri-island-take", TakeDef{Then: TakeRestIntoGraveyard})
