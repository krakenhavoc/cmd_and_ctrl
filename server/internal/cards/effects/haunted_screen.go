package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunted Screen — Artifact {3} (Duskmourn: House of Horror, #4466):
//
//	"{T}: Add {W} or {B}.
//	 {T}, Pay 1 life: Add {G}, {U}, or {R}.
//	 {7}: Put seven +1/+1 counters on this artifact. It becomes a 0/0
//	 Spirit creature in addition to its other types. Activate only once."
//
// The life is a cost of the second mana ability, paid before it taps
// (Horizon Canopy's shape).
//
// "Activate only once" is the OnlyOnce bit (#5221, CR 602.5b): once per
// object, so a Screen that leaves and comes back may activate it again
// (CR 400.7), and it is not an exhaust ability.
//
// The effect runs in printed order: the counters go on while the Screen
// is still not a creature, then the type change, which prints no
// duration and so lasts until the end of the game (CR 611.2a), pinned to
// this object (CR 400.7). A 0/0 with seven counters is a 7/7. A Screen
// that left in response gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49f92c0b-16ac-4cf0-8ec3-e3638990ee75",
		Name:         "Haunted Screen",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|B}",
				Label:    "{T}: Add {W} or {B}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Life: 1},
				Produced: "{G|U|R}",
				Label:    "{T}, Pay 1 life: Add {G}, {U}, or {R}",
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{7}: Put seven +1/+1 counters on this artifact. It becomes a 0/0 Spirit creature " +
				"in addition to its other types. Activate only once.",
			Purpose:  game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
			Cost:     ManaCost("{7}"),
			OnlyOnce: true,
			Effect:   hauntedScreenBecomesASpirit,
		}},
	})
}

// hauntedScreenBecomesASpirit is the {7} ability's effect.
func hauntedScreenBecomesASpirit(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if err := (AddCounter{Target: self, Kind: game.CounterPlusOne, N: 7}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target: self,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Spirit")...),
			game.SetBasePowerMod(0),
			game.SetBaseToughnessMod(0),
		},
		Duration: g.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Haunted Screen — a 0/0 Spirit creature in addition to its other types",
	}.Apply(ctx)
}
