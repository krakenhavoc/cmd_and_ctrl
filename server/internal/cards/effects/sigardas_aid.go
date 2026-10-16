package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigarda's Aid — Enchantment {W} (the Squall deck request, #3918; also
// on the EDHREC top-1,000 list, #3654):
//
//	"You may cast Aura and Equipment spells as though they had flash.
//	 Whenever an Equipment you control enters, you may attach it to
//	 target creature you control."
//
// The first ability is a standing cast-timing statement (Gandalf's
// sorcery clause, Vedalken Orrery's), narrowed by
// PermissionFilter.AuraOrEquipmentOnly: an Aura or an Equipment spell
// opens at instant speed, nothing else does. It changes when you may
// cast them, not when you may activate equip (still sorcery speed).
//
// The second is Hammer of Nazahn's trigger without the Hammer's own
// entry: any Equipment entering under your control, the "you may" asked
// as it resolves, the target chosen as it goes on the stack. It attaches
// the ENTERING Equipment, read off the triggering event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f5f2792-0cc4-4caa-b798-2ee9597f2524",
		Name:         "Sigarda's Aid",
		Completeness: CompletenessFull,
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{AuraOrEquipmentOnly: true},
				"You may cast Aura and Equipment spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype("Equipment")
			},
			Key:     "Sigarda's Aid — attach that Equipment to target creature you control",
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return hammerOfNazahnAttachEnteringEquipment(g, item, item.Trigger.Event.CardID)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Sigarda's Aid — attach that Equipment to target creature you control?",
			},
		}},
	})
}
