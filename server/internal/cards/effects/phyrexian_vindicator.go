package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Vindicator — Creature — Phyrexian Horror {W}{W}{W}{W}, 5/5:
//
//	"Flying
//	 If damage would be dealt to this creature, prevent that damage. When damage is prevented this way, this creature deals that much damage to any other target."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect, one
// application per recipient in a damage instance, and the additional
// effect is a reflexive trigger (CR 603.12, New Way Forward's shape): it
// goes on the stack with the amount prevented, its controller chooses
// "any other target" then, and it deals that much damage from the
// Vindicator. Damage that can't be prevented is dealt, and nothing
// triggers, because nothing was prevented this way (CR 615.12).
//
// "Other" is enforced on the clause (Another) and again at resolution.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "9b5cfbb7-21ed-491d-b77d-547e4d30a7da",
		Name:            "Phyrexian Vindicator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				Then:  vindicatorPreventedBody,
				Label: "Phyrexian Vindicator — prevent damage to it",
			}),
		},
	})
}

var (
	// The additional effect: nothing prevented, nothing triggers.
	vindicatorPreventedBody = game.DelayedBody("phyrexian-vindicator/prevented", vindicatorPrevented)
	// The reflexive trigger itself, targeting any other target.
	vindicatorStrikeBody = game.ReflexiveBody("phyrexian-vindicator/strike", vindicatorStrike,
		constTargets(func() *game.TargetSpec { return Another(b35TargetAnyOther()) }))
)

func vindicatorPrevented(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	t := WhenYouDo("Phyrexian Vindicator — deal the prevented damage to any other target", vindicatorStrikeBody)
	t.Params = game.EffectParams{Amount: p.Amount}
	return t.Apply(NewContext(g, item))
}

func vindicatorStrike(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || len(item.Targets) == 0 || item.Targets[0].ID == item.SourceCardID {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: p.Amount}.Apply(NewContext(g, item))
}
