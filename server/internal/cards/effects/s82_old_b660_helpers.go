package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b660_helpers.go — the Invasion domain cards that pump by the
// number of basic land types among the lands their controller controls
// (#5616). DomainCount is the count; these are the layer 7c statics.

// S82OldB660SelfPumpDomain is "Domain — This creature gets +p/+t for each
// basic land type among lands you control." (Kavu Scout, Wayfaring Giant): a
// layer 7c modification (CR 613.4c) of the creature itself, the count read on
// every pass so it follows the lands. Unlike DomainPT it is not a
// characteristic-defining ability: the printed power and toughness stay and
// the bonus is added to them.
func S82OldB660SelfPumpDomain(power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := DomainCount(g, source.Controller)
			c.Power += power * n
			c.Toughness += toughness * n
		},
	}
}

// S82OldB660DomainAura is a Domain Aura that pumps (sign 1) or shrinks (sign
// -1) its creature by 1/1 for each basic land type among the lands its
// controller controls (Strength of Unity, Exotic Curse).
func S82OldB660DomainAura(oracleID, name string, sign int, aura game.AuraPurpose) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: aura},
		Static: []game.StaticAbility{
			PumpAttachedPer(sign, sign, func(g *game.Game, source *game.Card) int {
				return DomainCount(g, source.Controller)
			}),
		},
	}
}
