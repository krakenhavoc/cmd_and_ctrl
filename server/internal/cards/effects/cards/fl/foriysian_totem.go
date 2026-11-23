package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foriysian Totem — Artifact {3}:
//
//	"{T}: Add {R}.
//	 {4}{R}: This artifact becomes a 4/4 red Giant artifact creature
//	 with trample until end of turn.
//	 As long as this artifact is a creature, it can block an
//	 additional creature each combat."
//
// The mana ability is an ordinary single-color tap ability
// (ManaAbilities). The animation is one ScopedEffectFor registration
// combining every clause of the middle line at one timestamp — the
// layer-4 type add, the layer-7b base P/T set, the layer-5 colour set
// and the layer-6 keyword grant — rather than four separate
// primitives, because the printed sentence is one continuous effect
// (CR 611.2) and the type change is what the third line's "as long as
// this artifact is a creature" reads.
//
// That third line is CanBlockAdditional (#1706) gated on `IsCreature`
// rather than on a fixed condition: it is true only while the middle
// ability's effect is live, and false again the instant cleanup or an
// answered removal spell (Vandalblast) takes the creature type away —
// which is exactly "as long as", read live on every recompute the way
// every other `target.IsCreature()` gate in the catalog is (see
// life_gated_statics.go, attack_requirements.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d053ea00-e727-4030-8c64-0d32ae57f169",
		Name:         "Foriysian Totem",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}: This artifact becomes a 4/4 red Giant artifact creature with trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{4}{R}"),
			Effect:  foriysianTotemAnimate,
		}},
		Static: []game.StaticAbility{
			effects.CanBlockAdditional(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature()
			}, 1),
		},
	})
}

func foriysianTotemAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	mods := append([]game.Mod{game.AddTypesMod("Artifact", "Creature"), game.AddSubtypesMod(ctx.TextSubtypes("Giant")...)},
		game.SetBasePTMods(4, 4)...)
	mods = append(mods, game.SetColorsMod(ctx.TextColors("R")...), game.AddKeywordsMod("trample"))
	return effects.ScopedEffectFor{
		Target:   ctx.Source(),
		Mods:     mods,
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Foriysian Totem — becomes a 4/4 red Giant artifact creature with trample",
	}.Apply(ctx)
}
