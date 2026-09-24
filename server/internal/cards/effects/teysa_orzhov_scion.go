package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
//     so they never retrigger it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8191342b-b25e-4c4d-8f69-aee662148ff4",
		Name:         "Teysa, Orzhov Scion",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice three white creatures: Exile target creature.",
			Cost:    SacrificeN(3, "three white creatures", Creature(), OfColor("W")),
			Targets: TargetCreature("target creature"),
			Effect:  exileFirstLegalCardTarget,
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller == source.Controller && dead.HasColor("B")
			}, "Teysa, Orzhov Scion — create a 1/1 white Spirit with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
	})
}
