package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chishiro, the Shattered Blade — Legendary Creature — Snake Samurai
// {2}{R}{G}, 4/4 (Kamigawa: Neon Dynasty):
//
//	"Whenever an Aura or Equipment you control enters, create a 2/2 red
//	 Spirit creature token with menace.
//	 At the beginning of your end step, put a +1/+1 counter on each modified
//	 creature you control. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// The first ability watches every entry under your control, Chishiro's own
// included, and reads the entering card's effective subtypes. The second is
// the modified predicate of CR 700.9 (b07IsModified) over the creatures you
// control as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c98a4af3-3e58-4927-9a60-d2664cf0b139",
		Name:         "Chishiro, the Shattered Blade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && (c.IsAura() || c.HasSubtype("Equipment"))
			}, "Chishiro, the Shattered Blade — create a 2/2 red Spirit creature token with menace",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Spirit with menace"), N: 1})),
			effects.AtYourEndStep("Chishiro, the Shattered Blade — put a +1/+1 counter on each modified creature you control",
				effects.S82NeoB11CounterOnEachModified),
		},
	})
}
