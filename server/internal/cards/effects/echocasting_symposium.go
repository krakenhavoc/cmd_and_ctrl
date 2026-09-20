package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Echocasting Symposium — Sorcery — Lesson {4}{U}{U}:
//
//	"Target player creates a token that's a copy of target creature
//	 you control.
//	 Paradigm (Then exile this spell. After you first resolve a spell
//	 with this name, you may cast a copy of it from exile without
//	 paying its mana cost at the beginning of each of your first main
//	 phases.)"
//
// The body ships whole. Two target clauses in printed order — a
// player, then a creature you control — and the token is created
// under the TARGETED player's control, not the caster's, which is the
// political half of the card and the thing easiest to get wrong. The
// copy is CreateTokenCopy, so the token carries the copied creature's
// oracle ID and with it every catalogued trigger, static, mana and
// activated ability (CR 707.2).
//
// Either target having left in response is handled per slot: no
// player to make it, or no creature to copy, and nothing is created.
// A resolution with BOTH gone never runs at all (CR 608.2b).
//
// WHAT IS MISSING. Paradigm. It is a keyword with no shape anywhere
// in the engine: not an alternative cost, not a cast permission, not
// a delayed trigger. It is three clauses at once — the spell exiles
// itself instead of going to the graveyard, a per-name "have you
// resolved one of these yet" record is kept, and a recurring
// first-main-phase offer to cast a COPY of the card out of exile is
// granted for the rest of the game. The copy-from-exile half has no
// primitive (every free-cast permission in the engine casts the CARD,
// not a copy of it), and the "first main phases" window is a standing
// per-turn offer rather than a duration the engine can stamp. It is a
// new seam and this batch opened its row in docs/engine-seams.md.
//
// Shipping without it is strictly weaker: the spell goes to the
// graveyard and is never seen again, where the printed card would
// keep offering copies every turn.
func init() {
	Register(Spec{
		OracleID:     "d88c3554-18db-4e5e-a979-2402990a0311",
		Name:         "Echocasting Symposium",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Paradigm isn't implemented — the spell goes to your graveyard and never offers you the repeating copies it promises.",
		},
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetCreature("target creature you control", YouControl()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			maker, ok := ctx.ClauseTarget(0)
			if !ok || maker.Kind != game.TargetPlayer {
				return nil
			}
			copied, ok := ctx.ClauseTarget(1)
			if !ok || copied.Kind != game.TargetCard {
				return nil
			}
			return CreateTokenCopy{Controller: maker.ID, Copy: copied.ID, N: 1}.Apply(ctx)
		},
	})
}
