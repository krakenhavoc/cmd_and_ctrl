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
//     KnownBy, the cached `effective` characteristic and
//     the battlefield position are per-instance state, not copiable
//     values (CR 707.2), and the cached characteristic in particular
//     is a pointer — a wholesale struct copy would have the token and
//     its original sharing one layer cache.
//   - It does not read the card's printed fields blindly: the copy
//     takes the COPIABLE values (CR 707.2) through
//     game.TokenCopySourceForEffect — the permanent's own, as other copy
//     effects changed them, while it is on the battlefield, and its
//     last-known ones when it has left and not moved since (CR 608.2h).
//     A card that has moved on (exiled from a graveyard by embalm or The
//     Scarab God, discarded) is copied as the card it now is.

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
	return createTokenCopies(ctx, c.Controller, tmpl, c.N)
}

// CreateTokenCopiesTogether is "for each <object>, create a token that's a
// copy of it" (#4594): N tokens copying each of Copies, made by ONE
// creation with a group per original, so every copy is created at the
// same time (CR 608.2f). The copies enter the battlefield in one event,
// each newcomer sees the others enter (CR 603.6a), and a "one or more"
// trigger or a "the first time you would create tokens" replacement sees
// one occurrence (CR 603.2c). Saheeli, the Gifted's −7 is the shape.
//
// Copies is the list the effect fixed as it began (a copy is never itself
// copied); an original that can no longer be found is skipped. Except is
// the "except …" clause, applied to each template (CR 707.9). Then is the
// rest of the effect, told the tokens that landed (TailOutcome.IDs), and
// runs whether or not a copy's entry stopped to ask a question; the zero
// tail is "nothing more".
type CreateTokenCopiesTogether struct {
	Controller uuid.UUID
	Copies     []uuid.UUID
	// N is how many tokens copy each original; zero means one.
	N      int
	Except func(t *game.Card)
	Then   game.KeyedTail
}

func (c CreateTokenCopiesTogether) Apply(ctx *Context) error {
	n := c.N
	if n <= 0 {
		n = 1
	}
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	groups := make([]game.TokenGroup, 0, len(c.Copies))
	for _, id := range c.Copies {
		tmpl, ok := TokenCopyTemplate(ctx.Game, id)
		if !ok {
			continue
		}
		if c.Except != nil {
			c.Except(&tmpl)
		}
		groups = append(groups, game.TokenGroup{Template: copiedToken(tmpl), Count: n})
	}
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: controller, Source: ctx.Source(), Groups: groups, Together: true,
	}, c.Then)
}

// TokenCopyGainsHaste is the "except it has haste" exception clause
// shared by every hasty-token-copy card — Reflection of Kiki-Jiki's
// activated ability and Electroduplicate's sorcery both print exactly
// this fragment on top of a scheduled sacrifice. Haste rides the
// token's printed Keywords, not a Layer 6 grant: a token has no
// catalog key of its own to hang one off.
func TokenCopyGainsHaste(t *game.Card) {
	for _, kw := range t.Keywords {
		if kw == "haste" {
			return
		}
	}
	t.Keywords = append(t.Keywords, "haste")
}

// SacrificeThisTokenAtEndStepGrant is the quoted ability "At the
// beginning of the end step, sacrifice this token." as a named bundle
// (ability_grant.go), for a copy effect that grants it as part of the
// copying process (Electroduplicate, Kindle the Inner Flame). `key` is the
// granting card's bundle key, an on-disk identity; `label` names the
// trigger on the stack.
//
// It is the token's own triggered ability, not a delayed trigger of the
// spell (CR 603.7): it triggers at the beginning of EVERY end step, whoever's
// turn it is, for as long as the token has it, and it sacrifices the
// object it is on (the resolving item's source).
func SacrificeThisTokenAtEndStepGrant(key, label string) AbilityGrant {
	return AbilityGrant{
		Key:  key,
		Text: "At the beginning of the end step, sacrifice this token.",
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, label, func(g *game.Game, item *game.StackItem) error {
				return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	}
}

// TokenCopyWithHasteAndGrant is the whole body of "Create a token that's a
// copy of target creature you control, except it has haste and '<ability>'"
// (Electroduplicate, Kindle the Inner Flame #4689). `grant` names the
// bundle the card declares in Spec.Grants (SacrificeThisTokenAtEndStepGrant).
//
// CR 707.9a: haste and the quoted ability are gained as part of the copying
// process, so both are copiable values of the token — haste on its printed
// Keywords (TokenCopyGainsHaste), the ability on its GrantedAbilities. A
// later copy of the token (a Clone, another token copy) has both too, and a
// token that loses all its abilities loses the ability and stays. The
// target is re-read out of ctx.LegalTargets(), so a creature that left in
// response is skipped (CR 608.2b).
func TokenCopyWithHasteAndGrant(grant string) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return CreateTokenCopy{
			Controller: item.Controller,
			Copy:       id,
			N:          1,
			Except: func(t *game.Card) {
				TokenCopyGainsHaste(t)
				TokenCopyGainsAbility(t, grant)
			},
		}.Apply(ctx)
	}
}

// TokenCopyGainsAbility is "except it has '<ability>'" on a token copy's
// template (CR 707.9a): the named bundle joins the template's copiable
// GrantedAbilities, idempotently and in clause order.
func TokenCopyGainsAbility(t *game.Card, grant string) {
	v := game.PrintedValues{GrantedAbilities: t.GrantedAbilities}
	v.GrantAbility(grant)
	t.GrantedAbilities = v.GrantedAbilities
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
// copiable values (CR 707.2) of the card `cardID`: a permanent's
// copiable values, its last-known ones if it has just left the
// battlefield (CR 608.2h), and otherwise the card wherever it now sits
// (game.TokenCopySourceForEffect). Reports ok=false when no such card
// exists.
//
// The type line gains the "Token" supertype so IsToken and the
// client's token affordances recognise it. Everything else is the
// copied values: name, oracle ID, Scryfall ID (so the client renders
// the original's art), type line, P/T, mana cost, colours and the
// abilities earlier copy effects granted (CR 707.9a).
// VariableToughness travels with Toughness (#683): a token copy of a
// `*` creature copies the importer's 0 stand-in, so it copies the bit
// that says the 0 is a stand-in, and the toughness state check keeps
// skipping it after it loses its last counter, as it does the
// original.
func TokenCopyTemplate(g *game.Game, cardID uuid.UUID) (game.Card, bool) {
	src, ok := g.TokenCopySourceForEffect(cardID)
	if !ok {
		return game.Card{}, false
	}
	return tokenCopyTemplateFromCard(src), true
}

// tokenCopyTemplateFromCard is TokenCopyTemplate's body over a card already
// chosen as the copy source: the template carrying its copiable values.
func tokenCopyTemplateFromCard(src game.Card) game.Card {
	out := game.Card{
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
		// CR 707.9a: an ability an earlier copy effect granted is part of
		// the copiable values, so a token copy of that copy has it too.
		GrantedAbilities: append([]string(nil), src.GrantedAbilities...),
		// CR 707.9b: and so is a keyword an earlier copy exception took
		// away (#5675).
		LostKeywords: append([]string(nil), src.LostKeywords...),
	}
	if src.OracleID == "" && src.TokenKey != "" {
		// A copy of a TOKEN (CR 707.2) takes the token's catalog
		// identity: with no oracle ID the key is the only thing its
		// abilities are found by.
		out.TokenKey = src.TokenKey
		out.PrintedPTKnown = src.PrintedPTKnown
	}
	if src.Layout == game.LayoutTransform && len(src.Faces) == 2 {
		// CR 707.8a, ADR 0185: a token copy of a transforming permanent
		// or card, a token or a real card, is a transforming token with
		// both faces, entering with whichever face the original shows
		// and none of its counters (not copiable). Each face carries the
		// Token supertype, or a transform would turn the token into a
		// card.
		out.Layout = src.Layout
		out.Faces = game.CopiableValuesOf(src).Clone().Faces
		for i := range out.Faces {
			out.Faces[i].TypeLine = tokenTypeLine(out.Faces[i].TypeLine)
		}
		out.ActiveFace = src.ActiveFace
	} else if sv := game.SingleFacedCopiableValuesOf(src); sv.OracleID != src.OracleID || sv.TokenKey != src.TokenKey {
		// A token copy of the back face of a modal double-faced
		// permanent is single-faced (CR 707.8) but keeps that face's
		// catalog entry.
		if sv.OracleID != "" {
			out.OracleID = sv.OracleID
		} else {
			out.TokenKey = sv.TokenKey
		}
	}
	return out
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

// ExileTargetThenTokenCopy is "Exile target creature card from a
// graveyard. Create a token that's a copy of it, except …" (The Scarab
// God, Ardyn, the Usurper): the first still-legal card target is exiled,
// and only a card that really was exiled is copied, under the resolving
// item's controller, with the exception `then` was registered with
// (ExileThenTokenCopyTail) applied to the copy.
//
// #4253 slice 10a: the copy is the exile's keyed tail, so a table waiting
// on the exile's prompt is a restore point.
func ExileTargetThenTokenCopy(g *game.Game, item *game.StackItem, then game.TailRef) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(then)}.Apply(ctx)
	}
	return nil
}

// ExileThenTokenCopyTail registers ExileTargetThenTokenCopy's "create a
// token that's a copy of it, except …" under key (ledgered "tail <key>",
// an on-disk identity): exceptFor builds the exception against the asking
// item's context (its text, CR 612.2). Call it once, from a package-level
// var.
func ExileThenTokenCopyTail(key string, exceptFor func(ctx *Context) func(t *game.Card)) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		if !oneWentThisWay(o) {
			return nil
		}
		return CreateTokenCopy{
			Controller: ctx.Controller(),
			Copy:       o.IDs[0],
			N:          1,
			Except:     exceptFor(ctx),
		}.Apply(ctx)
	})
}
