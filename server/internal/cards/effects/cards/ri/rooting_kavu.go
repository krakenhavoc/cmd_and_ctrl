package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rooting Kavu — Creature — Kavu {2}{G}{G}, 4/3 (Invasion):
//
//	"When this creature dies, you may exile it. If you do, shuffle all
//	 creature cards from your graveyard into your library."
//
// Gamekeeper's "you may exile it" dies trigger: exiling is attempted only
// while the Kavu is still the card that died in the graveyard (CR 400.7), and
// "if you do" is whether it really reached exile. The shuffle takes the
// creature cards in the graveyard as that clause resolves, the Kavu itself
// already gone, and then shuffles the library (once, even if no card moved).
//
// No simplification.
func init() {
	const label = "Rooting Kavu — you may exile it; if you do, shuffle all creature cards from your graveyard into your library"
	effects.Register(effects.Spec{
		OracleID:     "e7438616-cc42-4a6d-a7e3-5856cb0782fa",
		Name:         "Rooting Kavu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisDies(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				dead := item.Trigger.Event.CardID
				if !effects.DiedCardStillInGraveyard(ctx, dead) {
					return nil
				}
				return effects.ExileTarget{Target: dead, ThenTail: ctx.ItemTail(rootingKavuExiled)}.Apply(ctx)
			}), "Rooting Kavu — exile it?"),
		},
	})
}

// rootingKavuExiled is "if you do": the creature cards in the graveyard are
// shuffled into the library once the Kavu really reached exile. An on-disk key.
var rootingKavuExiled = effects.CardTail("rooting-kavu/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	player := ctx.Controller()
	if p := ctx.Game.PlayerByIDForEffect(player); p != nil && p.Graveyard != nil {
		var ids []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				ids = append(ids, c.InstanceID)
			}
		}
		for _, id := range ids {
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return effects.ShuffleLibrary{Player: player}.Apply(ctx)
})
