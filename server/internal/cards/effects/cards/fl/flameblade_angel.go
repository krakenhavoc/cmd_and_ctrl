package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flameblade Angel — Creature — Angel {4}{R}{R}, 4/4 (Shadows over
// Innistrad, S82 OLD):
//
//	"Flying
//	 Whenever a source an opponent controls deals damage to you or a
//	 permanent you control, you may have this creature deal 1 damage to that
//	 source's controller."
//
// Michiko Konda's trigger, widened to your permanents: one trigger per damage
// event, "that source's controller" read when it fires (Elesh Norn's
// B17DamageSourceController) and carried on the item, so a source that has
// left or changed hands still names the player who dealt it. The Angel deals
// the damage itself, as an object (Kederekt Parasite's "you may have this
// creature deal").
//
// No simplification.
func init() {
	const label = "Flameblade Angel — 1 damage to that source's controller"
	effects.Register(effects.Spec{
		OracleID:        "f912c739-27a2-4068-bf0c-0891309639df",
		Name:            "Flameblade Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || !effects.Cmdr2B08RecipientIsYouOrYourPermanent(ev.Target, g, source.Controller) {
					return false
				}
				dealer := effects.B17DamageSourceController(ev, g)
				return dealer != uuid.Nil && dealer != source.Controller
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Flameblade Angel — deal 1 damage to that source's controller?"},
			Key:            label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = effects.B17DamageSourceController(ev, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Params.Player
				if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
					return nil
				}
				ctx := effects.NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				return effects.DealDamage{SourceObject: &ref, Target: player, Amount: 1}.Apply(ctx)
			},
		}},
	})
}
