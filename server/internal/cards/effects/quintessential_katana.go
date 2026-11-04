package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quintessential Katana — {W} Artifact — Equipment (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Equipped creature gets +1/+1 and has "Whenever this creature deals
//	 combat damage, untap it and you gain 2 life."
//	 Whenever a Ninja you control enters, you may attach this Equipment to
//	 it.
//	 Equip {2}"
//
// The granted trigger is Biorganic Carapace's bundle (ADR 0093): the
// creature's own ability, so "it" and "you" are the equipped creature and
// its controller. "Deals combat damage" is to anything (Lost Jitte's
// reading, not the Swords'), and it is one trigger per combat damage step
// however many objects the creature damaged (OncePerBatch): a blocked
// trampler that hits its blocker and the player untaps once and gains 2
// life once. A double striker triggers in each of the two damage steps.
//
// The attach is Doc Ock's Tentacles' optional enters trigger with a Ninja
// test (the subtype is a word of the text, CR 612.2); it does nothing if
// the Ninja or the Katana left before it resolves.
//
// No simplification.
func init() {
	const grant = "quintessential-katana/combat-damage-untap-gain"
	Register(Spec{
		OracleID:     "e9685c61-e33d-4ae0-9dd3-7fd0df5eb46f",
		Name:         "Quintessential Katana",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventDealDamage && ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID
				}, "Quintessential Katana — untap this creature and you gain 2 life", Do(
					Instruction(func(ctx *Context) error {
						return UntapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
					}),
					GainLife{Amount: 2},
				))),
			},
			Text: "Whenever this creature deals combat damage, untap it and you gain 2 life.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(grant),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype(source.TextSubtype("Ninja"))
			},
			Key: "Quintessential Katana — attach to the Ninja",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Quintessential Katana — attach to it?"},
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
