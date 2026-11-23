package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banishment — Enchantment {3}{W} (Commander Legends: Battle for Baldur's
// Gate, #5948):
//
//	"Flash
//	 When this enchantment enters, exile target nonland permanent an opponent
//	 controls and all other nonland permanents your opponents control with the
//	 same name as that permanent until this enchantment leaves the
//	 battlefield."
//
// Banishing Light's exile (CR 610.3) over a group: the chosen permanent and
// every other nonland permanent an opponent controls with its name, as the
// trigger resolves, each exiled with a record of its own return, so all of
// them come back under their owners' control the moment this leaves. If this
// has left before the trigger resolves nothing is exiled (CR 610.3b); the name
// is the chosen permanent's as it last was, and a target gone by resolution
// makes the trigger do nothing (CR 608.2b). A token that is exiled ceases to
// exist and does not return (CR 111.8).
//
// No simplification.
func init() {
	const label = "Banishment — exile target nonland permanent an opponent controls and all other nonland permanents your opponents control with the same name"
	effects.Register(effects.Spec{
		OracleID:        "9c9b3bf7-afb7-4e9b-bdbc-e2b4d5ba477f",
		Name:            "Banishment",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim := effects.FirstLegalBattlefieldTarget(ctx)
				if victim == uuid.Nil {
					return nil
				}
				c, ok := g.LookupCardForEffect(victim)
				if !ok {
					return nil
				}
				ids := []uuid.UUID{victim}
				for _, o := range g.BattlefieldCardsForEffect() {
					if o.InstanceID != victim && o.Controller != item.Controller && !o.IsLand() && o.Name == c.Name {
						ids = append(ids, o.InstanceID)
					}
				}
				for _, id := range ids {
					if err := (effects.ExileUntil{Target: id, ThisLeaves: true,
						Label: "Banishment — the exiled card returns when Banishment leaves the battlefield"}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		},
	})
}
