package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2034d41b-5e77-49a6-884e-d3a6017f84d0",
		Name:         "Glowcap Lantern",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:               glowcapLanternLookGrant,
				LibraryTopVisible: game.LibraryTopOwner,
				Text:              "You may look at the top card of your library any time.",
			},
			{
				Key: glowcapLanternExploreGrant,
				Triggered: []game.TriggeredAbility{
					WheneverThisAttacks("Glowcap Lantern — it explores", func(g *game.Game, item *game.StackItem) error {
						return Explores{Explorer: item.SourceCardID}.Apply(NewContext(g, item))
					}),
				},
				Text: "Whenever this creature attacks, it explores.",
			},
		},
		Static: []game.StaticAbility{
			GrantAbilitiesToAttached(glowcapLanternLookGrant, glowcapLanternExploreGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
