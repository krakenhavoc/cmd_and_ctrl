package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seafloor Stalker — Creature — Merfolk Rogue {2}{U}, 2/3 (Zendikar Rising,
// #5508):
//
//	"{4}{U}: This creature gets +1/+0 until end of turn and can't be blocked
//	 this turn. This ability costs {1} less to activate for each creature in
//	 your party. (Your party consists of up to one each of Cleric, Rogue,
//	 Warrior, and Wizard.)"
//
// The reduction is a cost modifier on the activation, counted as it is
// activated (CR 700.8, CR 601.2f via 602.2b): generic mana only, so with a
// full party {4}{U} costs {U}. The Stalker itself is the Rogue if nothing
// else is. The effect does nothing to a Stalker that has left (CR 400.7).
//
// No simplification.
func init() {
	const label = "Seafloor Stalker — it gets +1/+0 until end of turn and can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "d3ee7f5d-56b4-4d87-9642-dcd8449dfb45",
		Name:         "Seafloor Stalker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "{4}{U}: This creature gets +1/+0 until end of turn and can't be blocked this turn. This ability costs {1} less to activate for each creature in your party.",
			Cost:    effects.ManaCost("{4}{U}"),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.OldB094PartySizeForCost, "This ability costs {1} less to activate for each creature in your party."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: label}.Apply(ctx)
			},
		}},
	})
}
