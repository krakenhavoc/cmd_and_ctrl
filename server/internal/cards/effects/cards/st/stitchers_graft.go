package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stitcher's Graft — Artifact — Equipment {1} (Innistrad Remastered, #5948):
//
//	"Equipped creature gets +3/+3.
//	 Whenever equipped creature attacks, it doesn't untap during its controller's
//	 next untap step.
//	 Whenever this Equipment becomes unattached from a permanent, sacrifice that
//	 permanent.
//	 Equip {2}"
//
// The attack trigger marks the attacker (the creature the event names) so its
// controller's next untap step skips it. The last sentence is Grafted
// Exoskeleton's: one trigger on the unattach event and on the Graft leaving the
// battlefield while attached (its last-known host), sacrificing the permanent it
// came off if that is still on the battlefield. An Equipment moved to another
// creature sacrifices the old host; one that is attached to nothing does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b73a5ec-c52e-4074-a7cb-78d6eb2adef0",
		Name:         "Stitcher's Graft",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Stitcher's Graft — equipped creature doesn't untap during its controller's next untap step",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
						return nil
					}
					return effects.DoesntUntapNextUntapStep{
						Targets: []uuid.UUID{item.Trigger.Event.CardID},
						Label:   "Stitcher's Graft — doesn't untap during its controller's next untap step",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.OnAny([]game.EventKind{game.EventUnattach, game.EventLTB}, effects.Self,
				"Stitcher's Graft — sacrifice the permanent it became unattached from", effects.GapB07SacrificeTheFormerHost),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
