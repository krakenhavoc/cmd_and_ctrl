package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinel's Mark — Enchantment — Aura {1}{W} (Ravnica Allegiance):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+2 and has vigilance.
//	 Addendum — When this Aura enters, if you cast it during your main phase,
//	 enchanted creature gains lifelink until end of turn."
//
// The addendum is an intervening-if trigger (CR 603.4) over the Aura's cast
// provenance and the phase it entered in (castInYourMainPhase).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac34bc09-ba8a-4bfb-bc45-e9c3e4d7bdc0",
		Name:            "Sentinel's Mark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
			effects.GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, castInYourMainPhase),
				"Sentinel's Mark — enchanted creature gains lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					aura, ok := g.LookupCardForEffect(ctx.Source())
					if !ok || aura.AttachedTo.Kind != game.TargetCard {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: aura.AttachedTo.ID, Keywords: []string{"lifelink"},
						Label: "Sentinel's Mark — lifelink until end of turn"}.Apply(ctx)
				}),
		},
	})
}

// castInYourMainPhase is "if you cast it during your main phase" on a permanent:
// its caster is its controller and it is that player's main phase as it enters,
// the phase it was cast in, since a spell cannot change phase on the stack.
func castInYourMainPhase(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && source.Provenance.Caster == source.Controller && effects.B14InYourMainPhase(g, source.Controller)
}
