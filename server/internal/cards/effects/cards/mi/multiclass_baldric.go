package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Multiclass Baldric — Artifact — Equipment {1} (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Equipped creature has lifelink if you control a Cleric, deathtouch if
//	 you control a Rogue, haste if you control a Warrior, and flying if you
//	 control a Wizard.
//	 As long as you have a full party, prevent all damage that would be dealt
//	 to equipped creature.
//	 Equip {2}"
//
// Four conditional keyword grants, each re-read every time the layers are
// computed (a Cleric that dies takes the lifelink with it), and a prevention
// that applies while you have a full party: four creatures in your party
// (CR 700.8c). "You" is the Equipment's controller. The creature that is
// equipped counts as one of your creatures when it is one.
//
// No simplification.
func init() {
	grant := func(subtype string, keyword string) game.StaticAbility {
		return effects.GrantToAttachedWhile(func(_ *game.Card, g *game.Game, source *game.Card) bool {
			return effects.ControlsPartyType(g, source.Controller, subtype)
		}, keyword)
	}
	effects.Register(effects.Spec{
		OracleID:     "5b173737-d33e-482f-a91e-f37e3e428e86",
		Name:         "Multiclass Baldric",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			grant("Cleric", "lifelink"),
			grant("Rogue", "deathtouch"),
			grant("Warrior", "haste"),
			grant("Wizard", "flying"),
		},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To: effects.ToEquippedCreature,
				While: func(g *game.Game, src *game.Card) bool {
					return effects.PartySizeInLayers(g, src.Controller) == 4
				},
				Label: "Multiclass Baldric: prevent all damage that would be dealt to equipped creature while you have a full party",
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
