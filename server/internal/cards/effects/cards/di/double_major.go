package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Double Major — Instant {G}{U}:
//
//	"Copy target creature spell you control, except it isn't
//	 legendary if the spell is legendary. (A copy of a creature spell
//	 becomes a token.)"
//
// The first card in the catalog to copy a PERMANENT spell, and the
// reason #666 exists. The reminder text is the rule: CR 608.3f makes
// a resolving copy of a permanent spell a token rather than a
// permanent card, which is what keeps a bounce spell from turning
// the copy into a second physical card in your hand.
//
// Both halves of the printed sentence are engine primitives rather
// than anything this file does. The token is
// resolvePermanentSpellCopyLocked (game/spell_copy.go), through the
// one token-creation path, so Doubling Season doubles it and the
// creature's own enters-the-battlefield abilities fire. The except
// clause is CopySpell.Except, which edits the copiable values the
// copy is created with — and "if the spell is legendary" needs no
// condition, because RemoveSupertype does nothing when the supertype
// is not there. Dropping legendary is the whole point of the card in
// Commander: it is how you get a second copy of your commander.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ece44a82-dcf0-4439-bdd9-a09c99a6f159",
		Name:         "Double Major",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetSpell("target creature spell you control",
			effects.Creature(), effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CopySpell{
				StackID:    item.Targets[0].ID,
				Controller: ctx.Controller(),
				Except: func(v *game.PrintedValues) {
					v.RemoveSupertype("Legendary")
				},
			}.Apply(ctx)
		},
	})
}
