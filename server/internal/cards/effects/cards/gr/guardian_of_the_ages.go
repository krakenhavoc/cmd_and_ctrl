package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guardian of the Ages — Artifact Creature — Golem {7}, 7/7 (Magic 2014, S82 OLD
// batch 401, #5563):
//
//	"Defender
//	 When a creature attacks you or a planeswalker you control, if this creature
//	 has defender, it loses defender and gains trample."
//
// Blood Reckoning's attack trigger (one per creature attacking you or a
// planeswalker you control, CR 508.3a) with an intervening "if" (CR 603.4): it is
// checked as the trigger fires and again as it resolves, reading defender through
// the layers. The change has no stated duration, so it lasts as long as the
// Guardian stays on the battlefield (CR 611.2a): one layer-6 effect removing
// defender and granting trample, pinned to the object (CR 400.7).
//
// No simplification.
func init() {
	hasDefender := func(g *game.Game, id uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(id)
		return ok && game.HasKeyword(&c, "defender")
	}
	effects.Register(effects.Spec{
		OracleID:        "1bf88f66-b673-4652-981b-3ea5d28f5c29",
		Name:            "Guardian of the Ages",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil &&
					effects.AttacksYouOrYourPlaneswalker(g, ev.Target, source.Controller) &&
					hasDefender(g, source.InstanceID)
			},
			Key: "Guardian of the Ages — it loses defender and gains trample",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Guardian of the Ages — it loses defender and gains trample")
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !hasDefender(g, item.SourceCardID) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				d, ok := effects.DurationWhileSourceRemains(ctx, item.SourceCardID)
				if !ok {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     []game.Mod{game.RemoveKeywordsMod("defender"), game.AddKeywordsMod("trample")},
					Duration: d,
					Label:    "Guardian of the Ages — loses defender and gains trample",
				}.Apply(ctx)
			},
		}},
	})
}
