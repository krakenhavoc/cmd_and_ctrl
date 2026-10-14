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
// Paradigm is the printed keyword, and the engine does the rest
// (game/paradigm.go, #1866): the spell exiles itself as it resolves,
// and the first time you resolve one, each of your precombat main
// phases for the rest of the game offers you a free cast of a copy of
// it. The copy picks its own targets, so each turn's copy can make a
// token for a different player.
func init() {
	Register(Spec{
		OracleID:        "d88c3554-18db-4e5e-a979-2402990a0311",
		Name:            "Echocasting Symposium",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordParadigm},
		// Slot 0 makes the token: a copy of the creature in slot 1.
		Purpose: ForTargets(game.TargetPurpose{Slot: 0, Tokens: 1}),
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
