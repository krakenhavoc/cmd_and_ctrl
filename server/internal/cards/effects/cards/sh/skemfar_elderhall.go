package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skemfar Elderhall — Land (Kaldheim, #5505):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {2}{B}{B}{G}, {T}, Sacrifice this land: Up to one target creature you
//	 don't control gets -2/-2 until end of turn. Create two 1/1 green Elf
//	 Warrior creature tokens. Activate only as a sorcery."
//
// "Up to one" means the ability may have no target and still makes its
// tokens; with a target that has become illegal the -2/-2 does nothing and
// the tokens are still made (CR 608.2b: it fizzles only if every target is
// illegal, and an ability with only an up-to-one target that was chosen and
// is now illegal does nothing at all).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70965b80-c8ad-4718-ae20-12a4d8228898",
		Name:         "Skemfar Elderhall",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}{B}{G}, {T}, Sacrifice this land: Up to one target creature you don't control gets -2/-2 until end of turn. Create two 1/1 green Elf Warrior creature tokens. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
			Cost:         effects.Plus(effects.ManaCost("{2}{B}{B}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets:      effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1),
			SorcerySpeed: true,
			Effect:       skemfarElderhallResolve,
		}},
	})
}

func skemfarElderhallResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Skemfar Elderhall — -2/-2 until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 2}.Apply(ctx)
}
