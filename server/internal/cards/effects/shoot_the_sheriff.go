package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shoot the Sheriff — Instant {1}{B} (EDHREC rank 4177):
//
//	"Destroy target non-outlaw creature. (Assassins, Mercenaries,
//	 Pirates, Rogues, and Warlocks are outlaws. Everyone else is fair
//	 game.)"
//
// Doom Blade with the colour restriction swapped for a batch
// restriction. In Commander the restriction bites much less often
// than "nonblack" does — most decks run few or no creatures from
// those five types — which is why a two-mana unconditional-looking
// kill spell sits at this rank rather than much higher.
//
// "Non-outlaw" is the whole rules content and it is a BATCH, not a
// creature type: nothing has the subtype "Outlaw", so the predicate
// is the complement of membership in the five types the reminder
// text lists (b40Outlaw). It reads effective subtypes, so a
// changeling is an outlaw and cannot be shot, and a Bear an effect
// has turned into a Pirate cannot be shot either — both of which are
// the printed card, because the restriction is checked on the target
// as it is on the battlefield, at announce and again at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2038138e-2e52-41e4-90fc-443fa054c5c0",
		Name:         "Shoot the Sheriff",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target non-outlaw creature", Not(b40Outlaw())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
