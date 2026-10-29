package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dancing from Dark to Dawn — Enchantment {3}{G}{G} (The Hobbit):
//
//	"Whenever you cast a creature spell, put X +1/+1 counters on target
//	 creature you control, where X is that spell's mana value.
//	 Landfall — Whenever a land you control enters, create a 2/2 green Bear
//	 creature token."
//
// The cast trigger is Necklace of Girion's targeted shape with Endrek Sahr's
// value: the target is a creature you control, chosen as the trigger goes on
// the stack and re-checked as it resolves (CR 608.2b), and X is read as it
// resolves (triggeringSpellManaValue, CR 608.2h), so an X the spell was cast
// for counts. With no creature to target the trigger is not put on the stack
// (CR 603.3d). The landfall trigger is the shared Landfall with a Bear token.
//
// No simplification.
func init() {
	cast := WheneverYouCast(Creature(), "Dancing from Dark to Dawn — put X +1/+1 counters on target creature you control, where X is that spell's mana value",
		func(g *game.Game, item *game.StackItem) error {
			x := triggeringSpellManaValue(g, item)
			if x <= 0 {
				return nil
			}
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
				}
			}
			return nil
		})
	cast.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "5ba482e9-fbb0-4d9f-a3a9-414892bcdfed",
		Name:         "Dancing from Dark to Dawn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cast,
			Landfall("Dancing from Dark to Dawn — create a 2/2 green Bear creature token",
				Do(CreateToken{Template: TokenCard("2/2 green Bear"), N: 1})),
		},
	})
}
