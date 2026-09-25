package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hammer of Nazahn — Legendary Artifact — Equipment {4}:
//
//	"Whenever Hammer of Nazahn or another Equipment you control
//	 enters, you may attach that Equipment to target creature you
//	 control.
//	 Equipped creature gets +2/+0 and has indestructible.
//	 Equip {4}"
//
// The trigger condition is The Mighty Thor, Jane Foster's "an
// Equipment you control enters" (enteredUnderYourControl +
// HasSubtype("Equipment")), widened to include Hammer's OWN entry —
// "Hammer of Nazahn or another Equipment" is `another` false, exactly
// as Coercive Recruiter's own-or-another-Pirate condition is.
//
// The effect attaches the ENTERING Equipment, which is not
// necessarily this permanent (a second Equipment entering later
// attaches itself, not the Hammer) — the entering card's instance ID
// is read at resolution off the item's carried trigger context
// (item.Trigger.Event.CardID, #1223) rather than captured at trigger
// time. AttachForEffect is the "attach an Equipment that isn't the
// ability's own source" primitive TwoSlotAttachTargets already uses
// for Brass Squire and Magnetic Theft.
//
// "You may" + a target clause is Eternal Witness's shape: the
// controller answers yes/no, and on yes picks the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3955573-3db5-490e-903e-65e0172a9202",
		Name:         "Hammer of Nazahn",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("indestructible"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype("Equipment")
			},
			Key:     "Hammer of Nazahn — attach that Equipment to target creature you control",
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return hammerOfNazahnAttachEnteringEquipment(g, item, item.Trigger.Event.CardID)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Hammer of Nazahn — attach that Equipment to target creature you control?",
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}

func hammerOfNazahnAttachEnteringEquipment(g *game.Game, item *game.StackItem, entering uuid.UUID) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.AttachForEffect(entering, t)
	}
	return nil
}
