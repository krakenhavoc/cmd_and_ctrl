package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruul War Plow — Artifact {4} (Guildpact):
//
//	"Creatures you control have trample.
//	 {1}{R}{G}: This artifact becomes a 4/4 Juggernaut artifact creature
//	 until end of turn."
//
// The trample is a layer 6 keyword grant to every creature the controller
// controls, itself included once animated. The animation is Mishra's
// Factory's: it becomes a creature in addition to being an artifact, a
// Juggernaut, with base power and toughness 4/4, until end of turn, and it
// follows only the permanent that activated it (CR 400.7). "Juggernaut" is a
// word of the source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dba119e-1b08-42e0-8d0a-b606bd90af02",
		Name:         "Gruul War Plow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && source != nil && target.Controller == source.Controller
		}, "trample")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{G}: This artifact becomes a 4/4 Juggernaut artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Artifact", "Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Juggernaut")...),
					}, game.SetBasePTMods(4, 4)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Gruul War Plow — a 4/4 Juggernaut artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
