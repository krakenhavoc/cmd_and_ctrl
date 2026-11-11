package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giggling Skitterspike — Artifact Creature — Toy {4}, 1/1 (Duskmourn
// Commander, #5368):
//
//	"Indestructible
//	 Whenever this creature attacks, blocks, or becomes the target of a
//	 spell, it deals damage equal to its power to each opponent.
//	 {5}: Monstrosity 5."
//
// One printed trigger, one row, watching attack, block and becomes-target.
// "Of a spell" is read off the stack (SelfTargetedByASpell), so an ability
// that targets it does not fire. The damage is the power the creature has
// as the trigger resolves, its last-known power if it has left (CR 608.2h),
// dealt as one damage instance to each opponent. Monstrosity is the shared
// constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca02cb31-9850-48bf-bc8d-ed86a4d4d4f7",
		Name:            "Giggling Skitterspike",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{5}"), 5)},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock, game.EventBecomesTarget},
				func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source) || SelfTargetedByASpell(ev, source, c, g)
				},
				"Giggling Skitterspike — damage equal to its power to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					power := powerNowOrLastKnown(g, item.SourceCardID)
					if power <= 0 {
						return nil
					}
					return originOfThorDealToEachOpponent(NewContext(g, item), item.SourceCardID, power)
				}),
		},
	})
}
