package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bellowing Fiend — Creature — Spirit {4}{B}, 3/3 (Tempest, S82 OLD
// batch 757, #5628):
//
//	"Flying
//	 Whenever this creature deals damage to a creature, this creature
//	 deals 3 damage to that creature's controller and 3 damage to you."
//
// Spiritmonger's "deals damage to a creature" trigger (combat or not).
// The two 3-damage hits are one object dealing them at once (one damage
// instance). "That creature's controller" is read as the trigger resolves,
// from the creature as it last existed if it has left (CR 608.2h), as
// Assembled Alphas does; "you" is the Fiend's controller, last known if the
// Fiend has gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37331d0e-edda-4791-8bcc-fb457f984877",
		Name:            "Bellowing Fiend",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToACreature,
				"Bellowing Fiend — 3 damage to that creature's controller and 3 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					victim := item.Trigger.Event.Target
					var theirs uuid.UUID
					if c, found := g.LookupCardForEffect(victim); found {
						if z := g.FindCardZoneForEffect(victim); z != nil && z.Kind == game.ZoneBattlefield {
							theirs = c.Controller
						}
					}
					if theirs == uuid.Nil {
						if info, found := g.LastKnownPermanentForEffect(victim); found {
							theirs = info.Controller
						}
					}
					return g.DamageInstanceForEffect(func() error {
						if theirs != uuid.Nil {
							if err := (effects.DealDamage{SourceObject: &ref, Target: theirs, Amount: 3}).Apply(ctx); err != nil {
								return err
							}
						}
						return effects.DealDamage{SourceObject: &ref, Target: item.Controller, Amount: 3}.Apply(ctx)
					})
				}),
		},
	})
}
