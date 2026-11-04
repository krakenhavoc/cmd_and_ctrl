package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Sable, Mercenary Leader — Legendary Creature — Human Mercenary Hero
// {2}{W}, 2/3 (Marvel's Spider-Man):
//
//	"When Silver Sable enters, put a +1/+1 counter on another target
//	 creature.
//	 Whenever Silver Sable attacks, target modified creature you control
//	 gains lifelink until end of turn. (Equipment, Auras you control, and
//	 counters are modifications.)"
//
// "Another" is any creature but Silver Sable, yours or not (Captain America,
// Steve Rogers' Another wrapper); with no other creature there is no trigger.
// "Modified" is CR 700.9's word, read by the catalog's one predicate for it
// (b07IsModified), both when the target is chosen and when it is re-checked on
// resolution (CR 608.2b): a creature that has lost its last modification by
// then is no longer a legal target. Silver Sable herself, with a counter on
// her, is a legal pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40b82dee-98c7-4b86-88b8-28388cd0331f",
		Name:         "Silver Sable, Mercenary Leader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Silver Sable, Mercenary Leader — put a +1/+1 counter on another target creature",
					b36CounterOnChosenAnimal),
				Another(TargetCreature("another target creature"))),
			Targeting(
				WheneverThisAttacks("Silver Sable, Mercenary Leader — target modified creature you control gains lifelink until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Silver Sable, Mercenary Leader — lifelink"}.Apply(ctx)
					}),
				TargetCreature("target modified creature you control", YouControl(), modifiedCreature)),
		},
	})
}
