package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// populate.go — populate (CR 701.36), "create a token that's a copy
// of a creature token you control".
//
// Its own file for the reason token_copy.go is: concurrent card
// batches collide on shared files, and populate is a keystone every
// Selesnya token card reaches for.
//
// Populate is CreateTokenCopy with the thing being copied chosen at
// resolution instead of fixed at announcement. It is deliberately NOT
// a targeting clause (CR 701.36a says "choose", not "target"), so
// hexproof, shroud and ward on the token are irrelevant and the
// engine's target machinery is not involved. The choice is the same
// choose_cards prompt every "choose a permanent you control" effect
// uses, over the controller's creature tokens only.
//
// What it does, in the order CR 701.36 says it:
//
//   - No creature token under your control (CR 701.36b): nothing
//     happens, no prompt is asked, and the rest of the card runs.
//   - Exactly one candidate, or several that are indistinguishable
//     copies of each other: no prompt either. A choice among copies
//     with the same copiable values (CR 707.2) changes nothing, and
//     asking it is a click the player had no say in.
//   - Otherwise one prompt, answered by the controller.
//
// The new token goes through CreateTokenCopy, which goes through the
// ordinary token-creation event, so Doubling Season and Parallel Lives
// double it and every "whenever a token enters" watcher sees it. A
// copy of a token copy copies the same copiable values, so it is
// unremarkable.

// Populate makes the controller choose a creature token they control
// and create a token that's a copy of it.
//
// Then is the rest of the sentence, and it is a continuation rather
// than the statement after Apply because the choice can be a prompt:
// anything written on the next line would run before the player has
// answered. It runs exactly once, after the copy exists (or straight
// away when there was nothing to copy), and is handed the IDs of the
// tokens the populate created — empty when it created none (CR
// 701.36b). Determined Iteration's "the token created this way gains
// haste" and "sacrifice it at the next end step" read that list; a
// card that says "populate" and nothing more leaves it nil.
//
// Except is CreateTokenCopy's "except it …" clause, applied to the
// copy and not to the original.
//
// Player defaults to the resolving item's controller.
type Populate struct {
	Player uuid.UUID
	Except func(t *game.Card)
	Then   func(ctx *Context, created []uuid.UUID) error
}

func (p Populate) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	candidates := permanentsControlledByMatching(ctx.Game, player,
		And(Creature(), IsTokenPredicate()))
	switch {
	case len(candidates) == 0:
		return p.finish(ctx, nil)
	case populateAllAlike(ctx.Game, candidates):
		return p.copyOne(ctx, player, candidates[0])
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract MillToZone follows: an undo restores the
	// game's fields in place, so a captured one would be the wrong one.
	item := ctx.Item
	source := ctx.Source()
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: "Populate — choose a creature token you control to copy",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			c := NewContext(g, item)
			if len(picked) == 0 {
				return p.finish(c, nil)
			}
			return p.copyOne(c, player, picked[0])
		},
	})
	return nil
}

// copyOne copies `id` and hands whatever token that made to Then. The
// event cursor is read before the copy so only tokens THIS populate
// made are reported: a replacement that doubles the token (Doubling
// Season) yields two, and a replacement that prevents it yields none.
func (p Populate) copyOne(ctx *Context, player, id uuid.UUID) error {
	cursor := b25LastEventSeq(ctx.Game)
	if err := (CreateTokenCopy{
		Controller: player,
		Copy:       id,
		N:          1,
		Except:     p.Except,
	}).Apply(ctx); err != nil {
		return err
	}
	return p.finish(ctx, b27TokensCreatedByAfter(ctx.Game, player, cursor))
}

func (p Populate) finish(ctx *Context, created []uuid.UUID) error {
	if p.Then == nil {
		return nil
	}
	return p.Then(ctx, created)
}

// populateAllAlike reports whether every candidate would produce the
// same copy: the same copiable values (CR 707.2), which is exactly
// what TokenCopyTemplate reads. Choosing between such tokens is not a
// choice.
func populateAllAlike(g *game.Game, ids []uuid.UUID) bool {
	first, ok := TokenCopyTemplate(g, ids[0])
	if !ok {
		return false
	}
	for _, id := range ids[1:] {
		next, ok := TokenCopyTemplate(g, id)
		if !ok || !sameCopiableValues(first, next) {
			return false
		}
	}
	return true
}

func sameCopiableValues(a, b game.Card) bool {
	if a.Name != b.Name || a.OracleID != b.OracleID || a.ScryfallID != b.ScryfallID ||
		a.TypeLine != b.TypeLine || a.Power != b.Power || a.Toughness != b.Toughness ||
		a.VariableToughness != b.VariableToughness || a.ManaCost != b.ManaCost {
		return false
	}
	return sameStrings(a.Colors, b.Colors) && sameStrings(a.Keywords, b.Keywords)
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
