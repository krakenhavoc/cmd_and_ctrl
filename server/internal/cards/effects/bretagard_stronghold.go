package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c792229b-4a0f-48d5-93e5-60bd4cae9c42",
		Name:         "Bretagard Stronghold",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{G}{W}{W}, {T}, Sacrifice this land: Put a +1/+1 counter on each of up to two target creatures you control. They gain vigilance and lifelink until end of turn. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{G}{W}{W}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "vigilance", "lifelink")),
			Targets:      TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := nothingIfGone(g.AddCounterByForEffect(item.Controller, t.ID, game.CounterPlusOne, 1)); err != nil {
						return err
					}
					if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance", "lifelink"},
						Label: "Bretagard Stronghold — vigilance and lifelink"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
