package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_copy.go — "create a token that's a copy of that card".
//
// Its own file rather than primitives.go, per the convention
// enters_tapped.go set: concurrent card batches collide on shared
// files, and this one is a keystone several unrelated batches will
// want.
//
// CreateToken builds a token from a hand-written template
// (tokens.go). That covers every token that is its own card — a
// Treasure, a 1/1 Soldier — and none of the ones that are a copy of
// something already in the game: Hashaton's Zombie, an eternalized
// Champion of Wits, a Clone. Those need the copiable values read off
// a real card at resolution time.
//
// What makes this cheap is that a token's identity in this engine is
// already just a game.Card, and the catalog keys every ability hook
// on OracleID. Carry the copied card's oracle ID onto the token and
// its triggered abilities, static abilities, printed keywords, mana
// abilities and activated abilities all come along for free, because
// every one of those hooks does a catalog lookup rather than reading
// a field. A copy really is a copy, not a vanilla body wearing the
// original's name.
//
// Two things it deliberately does NOT do:
//
//   - It does not copy the ORIGINAL Card struct wholesale. Counters,
//     KnownBy, ExilePlay, the cached `effective` characteristic and
//     the battlefield position are per-instance state, not copiable
//     values (CR 707.2), and the cached characteristic in particular
//     is a pointer — a wholesale struct copy would have the token and
//     its original sharing one layer cache.
//   - It does not read last-known information: the copy is taken from
//     wherever the named card sits right now.

// CreateTokenCopy creates N tokens that are copies of the card
// `Copy`, which may be in ANY zone — Hashaton copies a card in the
// graveyard, eternalize copies one in exile, Clone copies a
// permanent on the battlefield.
//
// `Except` is the card's "except it's ..." clause, applied to the
// template after the copiable values are read and before the token
// is created. Nil means a straight copy.
//
// A `Copy` that can no longer be found creates nothing and is not an
// error: the card it named may have been exiled in response, and a
// resolution that can't do its job still shouldn't wedge the stack.
//
// Since #762 the token runs the ordinary battlefield-entry pipeline,
// so BOTH halves of an "enters the battlefield" clause reach it: the
// copied card's ETB *triggered abilities* fire (EventETB is emitted
// and the harvester reads the log), and its CR 614.12 Spec.AsEnters
// clause fires too, because fireETBHookLocked is on that path. A token
// copy of Adaptive Automaton names a creature type, as it should. That
// was the declared gap here until token creation became a replaceable
// event with a real entry.
type CreateTokenCopy struct {
	Controller uuid.UUID
	Copy       uuid.UUID
	N          int
	Except     func(t *game.Card)
}

func (c CreateTokenCopy) Apply(ctx *Context) error {
	if c.N <= 0 {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(ctx.Game, c.Copy)
	if !ok {
		return nil
	}
	if c.Except != nil {
		c.Except(&tmpl)
	}
	return CreateToken{Controller: c.Controller, Template: tmpl, N: c.N}.Apply(ctx)
}

// TokenCopyOfSingleTarget is the whole body of the "create a token
// that's a copy of target <thing>" family — Cackling Counterpart's
// creature, Relm's Sketching's artifact, creature or land. The target
// clause is the only thing those cards do not share, and it lives on
// the Spec.
//
// It reads the first still-legal card target and copies it under the
// resolving item's controller. Every target having left in response
// is not an error: the spell does as much as it can, which is nothing
// (CR 608.2c).
func TokenCopyOfSingleTarget(item *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
		}.Apply(ctx)
	}
	return nil
}

// TokenCopyTemplate builds a CreateToken template carrying the
// copiable values (CR 707.2) of the card `cardID`, wherever it
// currently sits. Reports ok=false when no such card exists.
//
// The type line gains the "Token" supertype so IsToken and the
// client's token affordances recognise it. Everything else is the
// printed card: name, oracle ID, Scryfall ID (so the client renders
// the original's art), type line, P/T, mana cost and colours.
// VariableToughness travels with Toughness (#683): a token copy of a
// `*` creature copies the importer's 0 stand-in, so it copies the bit
// that says the 0 is a stand-in, and the toughness state check keeps
// skipping it after it loses its last counter, as it does the
// original.
func TokenCopyTemplate(g *game.Game, cardID uuid.UUID) (game.Card, bool) {
	src, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return game.Card{}, false
	}
	return game.Card{
		Name:               src.Name,
		ScryfallID:         src.ScryfallID,
		OracleID:           src.OracleID,
		TypeLine:           tokenTypeLine(src.TypeLine),
		Power:              src.Power,
		Toughness:          src.Toughness,
		VariableToughness:  src.VariableToughness,
		ManaCost:           src.ManaCost,
		Colors:             append([]string(nil), src.Colors...),
		ProducedMana:       append([]string(nil), src.ProducedMana...),
		Keywords:           append([]string(nil), src.Keywords...),
		ManaAbilities:      append([]game.ManaAbilityShape(nil), src.ManaAbilities...),
		ActivatedAbilities: append([]game.ActivatedAbilityShape(nil), src.ActivatedAbilities...),
	}, true
}

// tokenTypeLine prepends the "Token" supertype to a printed type
// line. Already-tokenised lines are returned unchanged so a copy of
// a token stays a single "Token".
func tokenTypeLine(printed string) string {
	if printed == "" {
		return "Token"
	}
	if containsFoldASCII(printed, "token") {
		return printed
	}
	return "Token " + printed
}

// retypedTypeLine rebuilds a type line with its subtypes REPLACED by
// `subtypes`, keeping supertypes and card types.
//
// That is what "except it's a 4/4 black Zombie" means (CR 707.9a):
// the exception sets the creature type rather than adding to it, so
// a copy of a Human Wizard is a Zombie and not a Zombie Human Wizard.
// Eternalize spells the retained types out when it keeps them
// ("a 4/4 black Zombie Snake Wizard"), which is the cleanest evidence
// that the bare form replaces.
//
// Supertypes survive: a copy of a legendary creature is legendary,
// and the legend rule applies to it.
func retypedTypeLine(printed string, subtypes ...string) string {
	super, types, _ := game.ParseTypeLine(printed)
	head := make([]string, 0, len(super)+len(types))
	head = append(head, super...)
	head = append(head, types...)
	line := strings.Join(head, " ")
	if len(subtypes) == 0 {
		return line
	}
	return line + " — " + strings.Join(subtypes, " ")
}
