package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bretagard Stronghold — Land:
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {G}{W}{W}, {T}, Sacrifice this land: Put a +1/+1 counter on each of up
//	 to two target creatures you control. They gain vigilance and lifelink
//	 until end of turn. Activate only as a sorcery."
//
// A tapped entry, a mana ability for {G}, and a sorcery-speed activated
// ability whose cost is mana, {T} and sacrificing the land. Targets are
// chosen as it is activated, and each is rechecked on resolution
// (CR 608.2b): one that has gone gets neither counter nor keywords, and the
// other still does. The land is sacrificed as a cost, so the ability
// resolves whether or not the land is still there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c792229b-4a0f-48d5-93e5-60bd4cae9c42",
		Name:         "Bretagard Stronghold",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{G}{W}{W}, {T}, Sacrifice this land: Put a +1/+1 counter on each of up to two target creatures you control. They gain vigilance and lifelink until end of turn. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{G}{W}{W}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance", "lifelink")),
			Targets:      effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := effects.NothingIfGone(g.AddCounterByForEffect(item.Controller, t.ID, game.CounterPlusOne, 1)); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance", "lifelink"},
						Label: "Bretagard Stronghold — vigilance and lifelink"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
