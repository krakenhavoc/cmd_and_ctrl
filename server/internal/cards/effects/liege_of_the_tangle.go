package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liege of the Tangle — Creature — Elemental (8/8) for {6}{G}{G}:
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, you may choose any number of target lands you control and put an awakening counter on each of them. Each of those lands is an 8/8 green Elemental creature for as long as it has an awakening counter on it. They're still lands."
//
// ADR 0109 §2 (#1604): each chosen land gets an awakening counter and
// is an 8/8 green Elemental creature, still a land (layers 4, 5 and
// 7b), for as long as it has an awakening counter on it. One record
// per land, each timed by its own counter, so removing one land's
// counter frees only that land. "You may choose any number" is a
// target count from zero up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a85b845c-a196-43db-b740-6ce191345097",
		Name:            "Liege of the Tangle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: ThisDealtCombatDamageToAPlayer,
			Key:       "Liege of the Tangle — awaken any number of target lands you control",
			Targets:   TargetPermanent("any number of target lands you control", Land(), YouControl()).WithCount(0, 0),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := CounterThenWhileItHasIt(ctx, t.ID, "awakening",
						"Liege of the Tangle — an 8/8 green Elemental creature while it has an awakening counter",
						game.AddTypesMod("Creature"), game.AddSubtypesMod("Elemental"), game.SetColorsMod("G"),
						game.SetBasePowerMod(8), game.SetBaseToughnessMod(8)); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
