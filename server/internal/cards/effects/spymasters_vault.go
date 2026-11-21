package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spymaster's Vault — Land (Modern Horizons 3, #5948):
//
//	"This land enters tapped unless you control a Swamp.
//	 {T}: Add {B}.
//	 {B}, {T}: Target creature you control connives X, where X is the
//	 number of creatures that died this turn."
//
// Castle Locthwain's entry and tap; X is read as the ability resolves
// (the turn's creature-death tally, tokens included), so a creature that
// dies in response raises it. X = 0 connives nothing (ConniveX).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69ddca4b-5cc0-45f3-b2e6-a047c8d601be",
		Name:         "Spymaster's Vault",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b06EntersTappedUnlessLandType("swamp")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}: Target creature you control connives X, where X is the number of creatures that died this turn.",
			Cost:    Plus(ManaCost("{B}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return ConniveX{Controller: item.Controller, Creature: t.ID, X: b11CreaturesDiedThisTurn(g)}.Apply(ctx)
			},
		}},
	})
}
