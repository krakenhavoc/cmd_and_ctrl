package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teysa, Orzhov Scion — Legendary Creature — Human Advisor {1}{W}{B},
// 2/3 (EDHREC rank 3474):
//
//	"Sacrifice three white creatures: Exile target creature.
//	 Whenever another black creature you control dies, create a 1/1
//	 white Spirit creature token with flying."
//
// The two halves feed each other: black creatures dying make white
// Spirits, and three white creatures exile anything.
//
//   - The removal is a sacrifice clause with a count of three (#747,
//     SacrificeN) over white creatures, with no tap and no mana, so it
//     can be activated at instant speed as often as the board allows.
//     The target is chosen before the cost is paid (CR 601.2c before
//     601.2h), so a white creature may be targeted and then be one of
//     the three sacrificed, as on paper; its target is then gone and
//     the ability does nothing on resolution (CR 608.2b). The engine
//     does not forbid that overlap. Teysa is white, so she may be one
//     of the three.
//   - The Spirit trigger is the shared dies watcher (diedCreature),
//     narrowed to another creature the controller controlled that is
//     black. The three Spirits a removal activation makes are white,
//     so they never retrigger it. "Black" is read as the permanent
//     last existed on the battlefield (leftAsColor, #1689, CR
//     603.10a), so a creature painted black only by an effect
//     (Darkest Hour, a black-making static, an Aura) counts, and one
//     an effect painted another colour does not — the card in the
//     graveyard is not consulted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8191342b-b25e-4c4d-8f69-aee662148ff4",
		Name:         "Teysa, Orzhov Scion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice three white creatures: Exile target creature.",
			Cost:    effects.SacrificeN(3, "three white creatures", effects.Creature(), effects.OfColor("W")),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsColor(ev, dead, source.TextColor("B"))
			}, "Teysa, Orzhov Scion — create a 1/1 white Spirit with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
	})
}
