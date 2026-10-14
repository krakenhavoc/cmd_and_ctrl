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
type Incubate struct {
	Controller uuid.UUID
	N          int
}

func (i Incubate) Apply(ctx *Context) error {
	controller := i.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	opts := game.TokenEntryOptions{}
	if i.N > 0 {
		opts.Counters = map[string]int{"+1/+1": i.N}
	}
	_, err := ctx.Game.CreateTokensForEffect(controller, IncubatorToken(), 1, opts)
	return err
}
