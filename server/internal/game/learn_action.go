package game

import (
	"sync"

	"github.com/google/uuid"
)

// learn_action.go — learn as an event a replacement effect can apply to
// (CR 701.48a, CR 614.1a, #3525), and the "if you would <keyword
// action>, <do this> instead" shape that makes it one.
//
// # The shape
//
// CR 701.48a: "'Learn' means 'You may discard a card. If you do, draw a
// card. If you didn't discard a card, you may reveal a Lesson card you
// own from outside the game and put it into your hand.'" Learn is one
// instruction, so it is one event: opened once at its one entry point
// (LearnForEffect) as a RepEventKeywordAction with KeywordActionLearn,
// the way a proliferate or a scry is, and performed only once the CR
// 614 window has settled on leaving it alone.
//
// What a printed card does to a learn is replace it. Retriever Phoenix:
// "As long as this card is in your graveyard, if you would learn, you
// may instead return this card to the battlefield." The replacement
// declares ReplacementEffect.KeywordInstead instead of a Replace, as a
// replaced draw declares DrawInstead: applying it cancels the learn
// (CR 614.6, the replaced event never happens, so the discard and the
// draw and the Lesson are not looked at) and records a registered body
// to run once the window has settled. The body may queue prompts.
//
// # Where it works from
//
// A card's own ability that works in the graveyard rides
// ReplacementEffect.FromGraveyard, which dredge introduced for a draw.
// The gather (gatherGraveyardReplacementsLocked) now looks in the
// LEARNING player's graveyard for a learn, as it looks in the drawing
// player's for a draw. Several of them in one graveyard are each a
// "may", asked in turn (allGraveyardOptions), so the player takes one or
// none; with a replacement that is not a graveyard "may" in the same
// window the ordinary CR 616 ordering prompt applies.
//
// # What is not a restore point
//
// A learn that has not been performed yet is not paused anywhere: the
// window either settles in one call or queues the replacement's yes/no.
// That prompt carries the generic CR 616 resume frame
// (ContinuationCensus.ChoiceResumeFrames), exactly as a dredge offer
// does, so a table with it open is not a restore point. The learn
// prompts that follow (the discard and the Lesson pick) are the
// closures learn.go already documents. This file adds no closure: the
// body is registered by key, the event carries data only, and nothing
// is added to the Game struct.

// LearnBody is what learn does once its window has left it alone: the
// "may discard, draw, or else take a Lesson from outside the game" of
// CR 701.48a. The catalog installs it at init (effects/learn.go), as it
// installs CatalogReplacements: the Lesson pick is a sideboard seam the
// game package does not own. Set once, before any game runs; a nil body
// makes learn do nothing.
var LearnBody func(g *Game, player, source uuid.UUID) error

// LearnForEffect is "learn" for `player`, on behalf of the card whose
// effect says so (`source`). It opens the CR 614 window on the action
// and, if nothing replaces it, performs it.
//
// It can pause: a replacement's yes/no, an ordering prompt, and the
// discard and Lesson questions of the learn itself are all queued, and
// a caller sequencing work behind it parks that work behind the prompt
// (the step walk does).
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) LearnForEffect(player, source uuid.UUID) error {
	_, err := g.runKeywordActionLocked(&ReplacementEvent{
		Kind:               RepEventKeywordAction,
		Actor:              player,
		Source:             source,
		KeywordAction:      KeywordActionLearn,
		KeywordActionCount: 1,
	})
	return err
}

// applyLearnLocked performs a learn the window left alone.
//
// Caller must hold g.mu.
func (g *Game) applyLearnLocked(ev *ReplacementEvent) error {
	if LearnBody == nil {
		return g.abandonKeywordActionLocked(ev)
	}
	return LearnBody(g, ev.Actor, ev.Source)
}

// KeywordInsteadFunc is the body of a keyword action replaced by an
// effect: "if you would learn, you may instead return this card to the
// battlefield". It runs with g.mu held once the replacement window has
// settled, may queue prompts, and must call done exactly once when it
// has finished: done runs the rest of the sentence the action was part
// of. `source` is the card the replacement belongs to.
type KeywordInsteadFunc func(g *Game, actor, source uuid.UUID, done func(g *Game) error) error

// KeywordInsteadRef names a registered KeywordInsteadFunc. The key is
// unexported, so the only way to hold one is RegisterKeywordInstead,
// which keeps a func out of ReplacementEffect and out of the event: the
// body lives in a registry and the effect holds its name.
type KeywordInsteadRef struct{ key string }

var (
	keywordInsteadMu     sync.RWMutex
	keywordInsteadBodies = map[string]KeywordInsteadFunc{}
)

// RegisterKeywordInstead registers a body under key and returns the ref
// a ReplacementEffect declares. Idempotent for one key. A body must
// capture scalars only, and the key must identify everything it
// captures. The key is an on-disk identity once a card ships with it:
// never renamed, never reused.
func RegisterKeywordInstead(key string, body KeywordInsteadFunc) KeywordInsteadRef {
	if key == "" || body == nil {
		panic("game.RegisterKeywordInstead: a key and a body are required")
	}
	keywordInsteadMu.Lock()
	keywordInsteadBodies[key] = body
	keywordInsteadMu.Unlock()
	return KeywordInsteadRef{key: key}
}

// keywordInsteadRun is a recorded, not yet run, substituted keyword
// action: which body, and which card it belongs to. Plain data, so a
// resume frame holding it is copied by value.
type keywordInsteadRun struct {
	ref    KeywordInsteadRef
	source uuid.UUID
}

// body looks the registered body up. Nil when nothing was recorded.
func (r keywordInsteadRun) body() KeywordInsteadFunc {
	if r.ref.key == "" {
		return nil
	}
	keywordInsteadMu.RLock()
	defer keywordInsteadMu.RUnlock()
	return keywordInsteadBodies[r.ref.key]
}

// recordKeywordInsteadLocked is what applying a KeywordInstead
// replacement does: cancel the action and remember which body to run.
func (g *Game) recordKeywordInsteadLocked(ev *ReplacementEvent, a activeReplacement) {
	ev.Canceled = true
	run := keywordInsteadRun{ref: a.effect.KeywordInstead}
	if a.source != nil {
		run.source = a.source.InstanceID
	}
	ev.keywordInstead = run
}
