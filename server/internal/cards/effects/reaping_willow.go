package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reaping Willow — Creature — Treefolk Cleric {1}{W/B}{W/B}{W/B}, 3/6
// (#3987):
//
//	"Lifelink
//	 This creature enters with two -1/-1 counters on it.
//	 {1}{W/B}, Remove two counters from this creature: Return target
//	 creature card with mana value 3 or less from your graveyard to the
//	 battlefield. Activate only as a sorcery."
//
// The two counters are of any kinds (RemoveCountersOfAnyKindsFromThis).
// The target is chosen as the ability is activated (CR 602.2b), from the
// activator's own graveyard; a card that has left it by resolution is
// an illegal target and nothing returns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce60402c-34cd-44fe-ae1d-4a5404298f5b",
		Name:            "Reaping Willow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Reaping Willow: enters with two -1/-1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{W/B}, Remove two counters from this creature: Return target creature card with mana value 3 or less from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{W/B}"), RemoveCountersOfAnyKindsFromThis(2)),
			Targets:      TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", YouOwn(), Creature(), ManaValueLE(3)),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 || ts[0].Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{Target: ts[0].ID, Dest: game.ZoneBattlefield}.Apply(ctx)
			},
		}},
	})
}
