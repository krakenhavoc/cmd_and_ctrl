package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crime // Punishment — split card (Breaking News, #5929), oracle
// c3df3574-c4da-415f-a56b-807ed8e41d73:
//
//	Crime {3}{W}{B} Sorcery: "Put target creature or enchantment card from
//	 an opponent's graveyard onto the battlefield under your control."
//	Punishment {X}{B}{G} Sorcery: "Destroy each artifact, creature, and
//	 enchantment with mana value X."
//
// Two specs, Punishment under "<oracle_id>#1" as Dusk // Dawn is. Either
// half is cast from the hand (CR 709.3); it has no fuse.
//
// Crime is Puppeteer Clique's reanimation: one target, a creature or
// enchantment card in a graveyard whose owner is an opponent, legal again
// as it resolves (CR 608.2b), entering under your control (an Aura chooses
// what it enchants as it enters, CR 303.4f).
//
// Punishment destroys every artifact, creature and enchantment with mana
// value exactly X as one simultaneous event, yours as well as theirs. A
// token has mana value 0 unless it is a copy of a card (CR 202.3), and a
// permanent's X is 0 on the battlefield (CR 107.3m), so X = 0 also destroys
// an X-cost permanent. An indestructible permanent survives.
//
// No simplification.
func init() {
	const oracle = "c3df3574-c4da-415f-a56b-807ed8e41d73"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Crime",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature or enchantment card from an opponent's graveyard", Or(Creature(), Enchantment()), Not(YouOwn())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			return ReturnFromGraveyard{Target: legal[0].ID, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Punishment",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return DestroyAllMatching{
				Match: And(Or(Artifact(), Creature(), Enchantment()), punishmentManaValueIs(x)),
			}.Apply(ctx)
		},
	})
}

// punishmentManaValueIs is "with mana value X" for the resolving spell's X.
func punishmentManaValueIs(x int) CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv == x
	}
}
