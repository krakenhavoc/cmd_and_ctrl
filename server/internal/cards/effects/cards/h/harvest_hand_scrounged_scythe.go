package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvest Hand // Scrounged Scythe — Artifact Creature — Scarecrow {3}, 2/2 //
// Artifact — Equipment (Innistrad Remastered, #5948):
//
//	Harvest Hand: "When this creature dies, return it to the battlefield
//	 transformed under your control."
//	Scrounged Scythe: "Equipped creature gets +1/+1.
//	 As long as equipped creature is a Human, it has menace.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// A double-faced card, the back face its own catalog row (oracle ID plus "#1").
// The dies trigger returns the card from the graveyard with its back face up
// (CR 712.14a), under the trigger controller's control, as Optimus Prime, Hero's
// does. The Scythe's menace is a layer-6 grant gated on the host's types, as
// Butcher's Cleaver's lifelink.
//
// No simplification.
func init() {
	const oracle = "ffd3c928-af42-4d69-8d22-a4c72f93b1c7"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Harvest Hand",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Harvest Hand — return it to the battlefield transformed under your control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFromGraveyard{
						Target:      item.SourceCardID,
						Dest:        game.ZoneBattlefield,
						Controller:  item.Controller,
						Transformed: true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Scrounged Scythe",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.Labelled(effects.GrantToAttachedWhile(func(host *game.Card, _ *game.Game, _ *game.Card) bool {
				return host.HasSubtypeWord("Human")
			}, "menace"), "As long as equipped creature is a Human, it has menace."),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
