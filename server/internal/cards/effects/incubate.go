package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incubate.go — the keyword action "incubate N" (CR 701.53) and the
// Incubator token it makes (CR 111.10i), the first double-faced token
// (ADR 0185).
//
// An Incubator is a colourless Incubator artifact with "{2}: Transform
// this token." whose back face is a 0/0 colourless Phyrexian artifact
// creature named "Phyrexian Token". "Incubate N" creates one that
// enters with N +1/+1 counters on it (CR 701.53a).
//
// Three things make it small, and one thing makes it right:
//
//   - The token template carries a Back face (token_catalog.go), so the
//     token IS a double-faced permanent in the engine's own terms: the
//     same Layout / Faces / ActiveFace a double-faced card has, which is
//     what lets effects.Transform act on it with no token-specific code
//     (CanTransform's layout allowlist already admits it).
//   - The counters are the creation's own entry clause
//     (game.TokenEntryOptions.Counters), so they ride the CR 614 entry
//     pipeline: Hardened Scales and Doubling Season see an Incubator
//     enter exactly as they see any other permanent enter with counters.
//   - Counters are not a characteristic, so a transform keeps them
//     (CR 712.18). The back face is a 0/0, which is why the counters are
//     the whole of its size, and why PrintedPTKnown is set on the
//     template: an Incubator that transforms with none is a real 0/0 and
//     CR 704.5f puts it into the graveyard.
//   - The CREATION is the incubate; the transform is a later, separate
//     act. Nothing here transforms anything.

// IncubatorToken is the Incubator as PRINTED: the template an incubate
// creates a token from, front face up.
func IncubatorToken() game.Card { return tokenFromCatalog(printedIncubatorToken) }

// printedIncubatorToken is the Incubator as PRINTED, abilities
// included. It is the catalog's entry for this token (token_catalog.go).
func printedIncubatorToken() tokenTemplate {
	return tokenTemplate{
		Slug: "incubator",
		Card: game.Card{
			Name:     "Incubator",
			TypeLine: "Token Artifact — Incubator",
			// CR 704.5f: the Phyrexian Token face is a real 0/0, so an
			// Incubator turned over with no counters on it dies. The
			// flag describes the printed size, not the face that is up.
			PrintedPTKnown: true,
		},
		Text: "{2}: Transform this token.",
		Activated: []game.ActivatedAbilityShape{{
			Label: "{2}: Transform this token",
			Cost:  game.AbilityCost{Mana: "{2}"},
			// Turning over makes it a creature, which is a blocker or an
			// attacker the opponent did not have a moment ago.
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  Do(TransformThis{}),
		}},
		Back: &tokenTemplate{
			Card: game.Card{
				Name:     "Phyrexian Token",
				TypeLine: "Token Artifact Creature — Phyrexian",
				Power:    0,
				// Toughness 0: the counters are its whole size.
				Toughness: 0,
			},
		},
	}
}

// Incubate is "incubate N" (CR 701.53a): create an Incubator token that
// enters with N +1/+1 counters on it.
//
// N is the counter count and may be a computed X ("incubate X, where X
// is the number of creatures exiled this way", Sunfall). N <= 0 still
// creates the token: "incubate 0" is a legal instruction and makes an
// Incubator with no counters, which is a real, if sad, permanent.
//
// Controller defaults to the resolving ability's controller. Set it for
// "its controller incubates X" (Excise the Imperfect), where the
// instruction is carried out by someone other than "you".
//
// Times is "incubate 2 twice" / "incubate 3 X times" (ADR 0185,
// amendment 2026-10-14). Zero means once; a card whose count is a
// computed X that can be zero ("incubate 3 X times", Progenitor Exarch)
// skips the effect itself rather than passing 0. Each repetition is its
// OWN incubate: its own CR 701.7b creation window, its own token with
// its own N counters, and its own token-created event. It is never one
// token with the summed counters, which would be a different permanent
// (one body, one Hardened Scales bonus instead of two) and would not
// trigger a "whenever you create a token" ability twice.
//
// Then is the rest of the sentence ("then transform all Incubator
// tokens you control", Elesh Norn's chapter I). It runs once, after the
// LAST Incubator has entered, from the creation's continuation, so a
// CR 616 ordering prompt on any repetition holds it back rather than
// letting it read a board that is still short a token.
type Incubate struct {
	Controller uuid.UUID
	N          int
	Times      int
	Then       func(ctx *Context) error
}

func (i Incubate) Apply(ctx *Context) error {
	controller := i.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	times := i.Times
	if times < 1 {
		times = 1
	}
	return incubateRounds(ctx.Game, ctx.Item, controller, i.N, times, i.Then)
}

// incubateRounds creates one Incubator and, from the creation's
// continuation, the `left - 1` that follow it, then runs `then`. The
// stack item rides the closure (and not a *Context or a *Card) for the
// reason living weapon's does: the continuation resolves against
// whichever game the creation finished in.
func incubateRounds(g *game.Game, item *game.StackItem, controller uuid.UUID, n, left int, then func(ctx *Context) error) error {
	opts := game.TokenEntryOptions{}
	if n > 0 {
		opts.Counters = map[string]int{"+1/+1": n}
	}
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: controller,
		Groups:     []game.TokenGroup{{Template: ruleDefinedToken(IncubatorToken()), Count: 1, Entry: opts}},
	}, func(g *game.Game, _ []uuid.UUID) error {
		if left > 1 {
			return incubateRounds(g, item, controller, n, left-1, then)
		}
		if then == nil {
			return nil
		}
		return then(NewContext(g, item))
	})
}

// IsIncubator matches an Incubator token: a token whose face-up side is
// the Incubator, "Token Artifact — Incubator". A turned-over one is a
// Phyrexian Token and no longer an Incubator token, so "transform
// target Incubator token" (CR 701.27) cannot pick it and turn it back;
// the printed cards name the front face, and the subtype is what says
// so. Read through HasSubtype, so a token copy of an Incubator (which
// copies the faces) matches too.
func IsIncubator() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return IsToken(c) && c.HasSubtype("Incubator")
	}
}

// TargetIncubatorYouControl is the clause "target Incubator token you
// control" (Elvish Vatkeeper, Progenitor Exarch).
func TargetIncubatorYouControl() *game.TargetSpec {
	return TargetPermanent("target Incubator token you control", IsIncubator(), YouControl())
}

// transformTargetIncubator is the body shared by every "transform
// target Incubator token you control": it turns over each legal target
// (there is one) and does nothing for a target that is gone or is no
// longer an Incubator (CR 608.2b: the predicate is re-run at
// resolution, so LegalTargets is already empty then).
func transformTargetIncubator(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (Transform{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// TransformAllIncubators is "transform all Incubator tokens you
// control" (Glissa, the Argent Etchings): the set is read once, now,
// and each member is turned over (CR 701.27a). Members of the set, so
// the #1432 new-object question is not asked of them.
func TransformAllIncubators(ctx *Context) error {
	controller := ctx.Controller()
	var ids []uuid.UUID
	for _, c := range ctx.Game.Battlefield.Cards {
		if c.Controller == controller && IsIncubator()(ctx.Game, controller, c) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (Transform{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
