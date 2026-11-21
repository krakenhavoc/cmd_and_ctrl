package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	glowcapLanternLookGrant    = "glowcap-lantern/look-at-top"
	glowcapLanternExploreGrant = "glowcap-lantern/attack-explores"
)

// Glowcap Lantern — Artifact — Equipment {G} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Equipped creature has 'You may look at the top card of your library
//	 any time' and 'Whenever this creature attacks, it explores.'
//	 Equip {2}"
//
// Two ADR 0093 grants. The look is the creature's controller's private look
// (AbilityGrant.LibraryTopVisible, the Courser's mechanism, game.LibraryTopOwner):
// it makes the top card of the library of whoever controls the equipped
// creature known to that player alone, and goes with the Equipment. The attack
// trigger is the creature's own: it explores under its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2034d41b-5e77-49a6-884e-d3a6017f84d0",
		Name:         "Glowcap Lantern",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key:               glowcapLanternLookGrant,
				LibraryTopVisible: game.LibraryTopOwner,
				Text:              "You may look at the top card of your library any time.",
			},
			{
				Key: glowcapLanternExploreGrant,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisAttacks("Glowcap Lantern — it explores", func(g *game.Game, item *game.StackItem) error {
						return effects.Explores{Explorer: item.SourceCardID}.Apply(effects.NewContext(g, item))
					}),
				},
				Text: "Whenever this creature attacks, it explores.",
			},
		},
		Static: []game.StaticAbility{
			effects.GrantAbilitiesToAttached(glowcapLanternLookGrant, glowcapLanternExploreGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
