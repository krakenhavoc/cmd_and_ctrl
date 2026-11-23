package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blade of Shared Souls — Artifact — Equipment {2}{U} (Phyrexia: All Will
// Be One):
//
//	"For Mirrodin! (When this Equipment enters, create a 2/2 red Rebel
//	 creature token, then attach this to it.)
//	 Whenever this Equipment becomes attached to a creature, for as long as
//	 this Equipment remains attached to it, you may have that creature become
//	 a copy of another target creature you control.
//	 Equip {2}"
//
// For Mirrodin! is the shared keyword trigger (#5654). The attach trigger is
// Assimilation Aegis's copy shape (CopyWhileSourceAttached, CR 611.2b)
// with a target: "another" is any creature you control but the one this
// became attached to, read off the attach event, so the Rebel it arrives on
// can become a copy of another creature you control, and a lone creature has
// no other to copy (the trigger is removed, CR 603.3d). The "you may" is
// asked as it resolves (CR 603.5). The copy is a layer-1 effect (CR 707.2,
// CR 613.2a) of the target's copiable values, lasting only while this
// Equipment stays attached to that same creature: equipping another creature
// ends it on the first and the new attach triggers again.
//
// No simplification.
func init() {
	trigger := effects.On(game.EventAttach, effects.ThisBecameAttachedToACreature,
		"Blade of Shared Souls — you may have the equipped creature become a copy of another target creature you control",
		bladeOfSharedSoulsCopy)
	trigger.TargetsFrom = bladeOfSharedSoulsTargets
	effects.Register(effects.Spec{
		OracleID:     "bfecd486-8326-426b-ac2b-54946f6ae707",
		Name:         "Blade of Shared Souls",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ForMirrodin("Blade of Shared Souls"),
			effects.Optional(trigger, "Blade of Shared Souls — have the equipped creature become a copy of the target creature?"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// bladeOfSharedSoulsTargets is "another target creature you control": any
// but the creature this Equipment just became attached to. A pure read of
// the trigger context.
func bladeOfSharedSoulsTargets(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	return effects.TargetCreature("another target creature you control", effects.YouControl(), effects.OtherThan(tc.Event.Target))
}

// bladeOfSharedSoulsCopy is the resolution: the equipped creature becomes a
// copy of the chosen creature while this Equipment stays attached to it.
func bladeOfSharedSoulsCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if item.Trigger == nil {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.BecomeCopy{
			Targets:  []uuid.UUID{item.Trigger.Event.Target},
			Of:       t.ID,
			Duration: effects.CopyWhileSourceAttached,
			Label:    "Blade of Shared Souls — becomes a copy while attached",
		}.Apply(ctx)
	}
	return nil
}
