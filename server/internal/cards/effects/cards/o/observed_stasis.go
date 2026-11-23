package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Observed Stasis — Enchantment — Aura {3}{U}:
//
//	"Flash
//	 Enchant creature an opponent controls
//	 When this Aura enters, remove enchanted creature from combat. Then
//	 draw a card for each tapped creature its controller controls.
//	 Enchanted creature loses all abilities and can't attack or block."
//
// The entry trigger reads the host through the Aura's attachment. The
// creature leaves combat first; the draw is then counted over every
// tapped creature its controller controls, the enchanted one included
// if it is tapped, and goes to the Aura's controller. The static half is
// Pacifism's restriction plus the layer-6 removal of all abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c729806-3a23-491e-847f-23a394d27291",
		Name:            "Observed Stasis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(effects.OpponentControls()),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.LoseAllAbilities(),
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Observed Stasis — remove enchanted creature from combat, then draw a card for each tapped creature its controller controls",
				observedStasisEntry),
		},
	})
}

func observedStasisEntry(g *game.Game, item *game.StackItem) error {
	host := effects.AttachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	hostID, hostController := host.InstanceID, host.Controller
	// Removing a creature that is not in combat does nothing.
	_ = g.RemoveFromCombatForEffect(hostID)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == hostController && c.IsCreature() && c.Tapped {
			n++
		}
	}
	return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
}
