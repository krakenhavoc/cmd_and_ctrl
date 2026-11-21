package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chronatog Totem — Artifact {3}:
//
//	"{T}: Add {U}.
//	 {1}{U}: This artifact becomes a 1/2 blue Atog artifact creature
//	 until end of turn.
//	 {0}: This creature gets +3/+3 until end of turn. You skip your next
//	 turn. Activate only once each turn and only if this permanent is a
//	 creature."
//
// The animation is one layered record at one timestamp (CR 613.7): the
// creature type (layer 4), the Atog subtype, blue (layer 5) and the
// base 1/2 (layer 7b), pinned to this object so a Totem flickered in
// response stays an artifact (CR 400.7). The third line is Chronatog's
// (chronatogPumpAndSkip), gated on the source being a creature at
// announce as well as on the once-per-turn tally (CR 602.1b). Both are
// checked when the ability is activated, not when it resolves, which is
// how the printed restriction reads.
//
// No simplification.
func init() {
	const pumpLabel = "{0}: This creature gets +3/+3 until end of turn. You skip your next turn. Activate only once each turn and only if this permanent is a creature."
	effects.Register(effects.Spec{
		OracleID:     "24542d66-e67a-4403-91a8-bb4231ea1f50",
		Name:         "Chronatog Totem",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}: This artifact becomes a 1/2 blue Atog artifact creature until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{1}{U}"),
				Effect:  chronatogTotemAnimate,
			},
			{
				Label:   pumpLabel,
				Purpose: game.Purpose{Answers: game.AnswerPump, SkipsTurns: 1},
				Cost:    effects.ManaCost("{0}"),
				Condition: effects.AllConditions(effects.OncePerTurnActivation(pumpLabel), func(g *game.Game, _, source uuid.UUID) bool {
					c, ok := g.LookupCardForEffect(source)
					return ok && c.IsCreature()
				}),
				Effect: effects.ChronatogPumpAndSkip("Chronatog Totem — +3/+3"),
			},
		},
	})
}

// chronatogTotemAnimate makes the Totem a 1/2 blue Atog artifact
// creature until end of turn, if it is still the permanent the ability
// came from.
func chronatogTotemAnimate(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	return effects.ScopedEffectFor{
		Target: item.SourceCardID,
		Mods: append([]game.Mod{
			game.AddTypesMod("Artifact", "Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Atog")...),
			game.SetColorsMod(ctx.TextColors("U")...),
		}, game.SetBasePTMods(1, 2)...),
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Chronatog Totem — becomes a 1/2 blue Atog artifact creature",
	}.Apply(ctx)
}
