package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summoning Materia — Artifact — Equipment {2}{G} (#3656):
//
//	"You may look at the top card of your library any time.
//	 As long as this Equipment is attached to a creature, you may cast
//	 creature spells from the top of your library.
//	 Equipped creature gets +2/+2 and has vigilance and "{T}: Add {G}."
//	 Equip {2}"
//
// Every line is an existing shape: the private look is
// LibraryTopVisible (Elven Chorus's), the cast permission is Elven
// Chorus's creature-only top-of-library permission gated by a
// Condition (#1314, Fortune Teller's Talent's slot) that reads the
// attachment live, and the rest is the ordinary Equipment kit.
//
// No simplification.
const summoningMateriaGrant = "summoning-materia/add-g"

func init() {
	effects.Register(effects.Spec{
		OracleID:          "2ca9e6e5-5eec-4593-8816-5c4317c79eb0",
		Name:              "Summoning Materia",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: effects.PlayFromTopOfYourLibrary(game.PermissionFilter{CreatureOnly: true},
				"Cast creature spells from the top of your library (Summoning Materia)"),
			Condition: summoningMateriaAttachedToACreature,
		}},
		Grants: []effects.AbilityGrant{effects.TapForManaGrant(summoningMateriaGrant, "{G}", "Add {G}", "{T}: Add {G}.")},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("vigilance"),
			effects.GrantAbilitiesToAttached(summoningMateriaGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// summoningMateriaAttachedToACreature is "as long as this Equipment is
// attached to a creature", read when the permission is asked for.
func summoningMateriaAttachedToACreature(g *game.Game, _, source uuid.UUID) bool {
	materia, ok := g.LookupCardForEffect(source)
	if !ok {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if materia.IsAttachedTo(c.InstanceID) && c.IsCreature() {
			return true
		}
	}
	return false
}
