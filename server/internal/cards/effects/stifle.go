package effects

// Stifle — Instant {U} (EDHREC rank 4571):
//
//	"Counter target activated or triggered ability.
//	 (Mana abilities can't be targeted.)"
//
// The card the whole "counter an ability" family is named after, and
// the smallest possible proof of #1211: one target clause, one
// primitive, nothing else. Everything that makes it hard lives in the
// clause — an ability on the stack is a `game.StackMeta` entry with a
// synthetic id and no card in any zone, so until `TargetSpec.Abilities`
// there was nothing a card file could write here.
//
// The PARENTHETICAL needs no code. CR 605.3b keeps a mana ability off
// the stack entirely, so there is no item for the clause to enumerate
// and "mana abilities can't be targeted" is a description of the
// object model rather than a restriction on top of it.
//
// CR 701.5c is the other half: an ability that is countered does not
// go anywhere. It has no card to route, its source permanent stays on
// the battlefield untouched, and `counterAbilityLocked` deletes the
// item — which is why a Stifled fetchland is still in play and still
// tapped, and why nothing about the ability's cost is refunded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3b00911-ece7-4484-bc36-f211ce72b6cc",
		Name:         "Stifle",
		Completeness: CompletenessFull,
		Targets:      AbilityOnStack("target activated or triggered ability"),
		OnResolve:    counterTheTargetSpell,
	})
}
