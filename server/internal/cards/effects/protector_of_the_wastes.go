package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Protector of the Wastes — Creature — Dragon {4}{W}{W}, 5/5 (Tarkir:
// Dragonstorm Commander):
//
//	"Flying
//	 When this creature enters or becomes monstrous, exile up to two target
//	 artifacts and/or enchantments controlled by different players.
//	 {4}{W}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)"
//
// One printed ability with two conditions (OnAny over the entry and
// EventBecameMonstrous). The clause is "up to two" artifacts and/or
// enchantments with the set rule that no two share a controller
// (EachDifferentController, read as each target is judged again on resolution,
// CR 608.2b), so two permanents of one player are refused at announce and a
// permanent that changed hands to the other target's controller is dropped.
// Every target still legal is exiled. Monstrosity is the shared ability
// (Ember Swallower's), which does nothing once the creature is monstrous.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4534e260-06fd-4054-8077-505c2f6c5767",
		Name:            "Protector of the Wastes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{4}{W}"), 3)},
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventETB, game.EventBecameMonstrous}, Self,
				"Protector of the Wastes — exile up to two target artifacts and/or enchantments controlled by different players",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetPermanent("up to two target artifacts and/or enchantments controlled by different players",
					Or(Artifact(), Enchantment())).WithCount(0, 2).EachDifferent(EachDifferentController())),
		},
	})
}
