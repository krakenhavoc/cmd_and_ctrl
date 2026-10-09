package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greenhouse Propagator — Creature — Cat Druid {2}{G}, 2/3:
//
//	"Whenever another creature you control enters, you gain 1 life.
//	 {T}: Add {G}."
//
// Essence Warden's trigger narrowed to the controller's own creatures
// ("another" excludes the Propagator's own entry), plus a plain
// tap-for-green mana ability, which is summoning-sick like any other
// creature's tap ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6b77cb9-49a5-4d1a-ae2d-02ae176dc1fa",
		Name:         "Greenhouse Propagator",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature()
			}, "Greenhouse Propagator — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
