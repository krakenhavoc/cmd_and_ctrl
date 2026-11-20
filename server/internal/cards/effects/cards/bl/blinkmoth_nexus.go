package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blinkmoth Nexus — Land:
//
//	"{T}: Add {C}.
//	 {1}: This land becomes a 1/1 Blinkmoth artifact creature with flying
//	 until end of turn. It's still a land.
//	 {1}, {T}: Target Blinkmoth creature gets +1/+1 until end of turn."
//
// A manland built like Mutavault: the animation adds the Artifact and
// Creature types and the Blinkmoth subtype (layer 4), flying (layer 6)
// and a 1/1 base (layer 7b). It is colourless and keeps its mana
// ability. The pump targets any Blinkmoth creature, itself once it is
// animated, and is a {T} ability, so a land that came under its
// controller's control this turn cannot use it while it is a creature
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "40d45c02-6416-4e19-8fe3-0ddadf5ba627",
		Name:          "Blinkmoth Nexus",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}: This land becomes a 1/1 Blinkmoth artifact creature with flying until end of turn. It's still a land.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: append([]game.Mod{
							game.AddTypesMod("Artifact", "Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Blinkmoth")...),
							game.AddKeywordsMod("flying"),
						}, game.SetBasePTMods(1, 1)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Blinkmoth Nexus — a 1/1 Blinkmoth artifact creature with flying until end of turn",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{1}, {T}: Target Blinkmoth creature gets +1/+1 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetCreature("target Blinkmoth creature", effects.Subtype("Blinkmoth")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Blinkmoth Nexus — +1/+1"}.Apply(ctx)
				},
			},
		},
	})
}
