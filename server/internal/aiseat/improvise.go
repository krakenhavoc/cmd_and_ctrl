package aiseat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// Improvisation is a bot playing a card by hand, the way a human at
// this table does with the right-click override menu, because the
// catalog cannot execute the effect its line needs. ADR 0033 §8.
//
// The catalog covers a few hundred cards out of ~35,000. A bot
// holding one the engine cannot run would otherwise be stuck with it
// or would quietly skip it — and quietly skipping it is the worse of
// the two, because the table cannot tell the difference between a bot
// choosing not to cast and a bot that could not. Improvisation is the
// third option, and it comes with two properties that are the feature
// rather than decoration:
//
//   - ANNOUNCED. Every improvisation posts a chat line naming the
//     card and the intended effect, and stamps a tagged line in the
//     replay log. An unannounced improvisation is a bot cheating, so
//     the announcement is built and validated BEFORE anything is
//     dispatched, and a bundle whose announcement cannot be built is
//     refused rather than applied silently.
//
//   - REVERSIBLE, AS ONE THING. The steps commit through
//     ws.Room.ApplyBundle: all of them or none, one undo entry,
//     stamped uuid.Nil so any seated human can undo it without the
//     admin token, and flagged FreeUndo so doing so costs them none
//     of their own per-turn undo budget.
//
// Improvisation is also the ONE path on which a bot acts with
// caller-nil (admin) authority — it has to, because "destroy target
// creature" reaches an opponent's board and the seated-caller gates
// would refuse it. Everything above is the counterweight to that
// grant: the table is told, the record is greppable, and any human
// can put it back.
type Improvisation struct {
	// Card is the card being played by hand. Required — an
	// improvisation that will not say what it is for is refused.
	Card string
	// Effect is the intended effect in plain words, as the table
	// will be told it ("destroy Avacyn, Angel of Hope"). Required.
	Effect string
	// Reason is the policy's rationale. Recorded in the replay
	// always; shown in chat only to viewers who turned on "show bot
	// reasoning".
	Reason string
	// Steps are the sandbox verbs that apply the effect, in order.
	// They commit together or not at all.
	Steps []ImprovStep
}

// ImprovStep is one sandbox verb in an improvisation bundle. Type and
// Params are exactly the wire ActionPayload fields; Player is the
// wire Player field, required by the player-scoped verbs.
//
// Type is a string rather than an actions.Type for the same reason
// legal.Move.Type is: policies should not need the engine's packages
// to describe what they want done. TestImprovVerbsMatchActions pins
// the constants to the actions ones.
type ImprovStep struct {
	Type   string
	Player uuid.UUID
	Params json.RawMessage
}

// The sandbox verbs an improvisation may use. ADR 0033 §8 names these
// four and no others, and Validate enforces the list.
//
// The restriction is not arbitrary. These four move a card between
// zones, change a life total, add or remove a counter, and mark
// damage — between them they express most of what a card does, and
// each one is individually reversible by restoring the pre-bundle
// clone. Verbs that drive the turn structure (advance_step,
// pass_priority), that touch another seat's hidden zones
// (discard_selection), or that end a game (concede) are not
// improvisation: they are the bot playing the game, and they go
// through the enumerator like every other move.
const (
	VerbMoveCard   = "move_card"
	VerbChangeLife = "change_life"
	VerbAddCounter = "add_counter"
	VerbMarkDamage = "mark_damage"
)

// MaxImprovSteps bounds one bundle. A single card's manual execution
// is a handful of verbs — move it to the graveyard, drop two life,
// put a counter on something. A policy asking for thirty has lost the
// plot, and a bundle that large is one nobody at the table can check
// against the card's oracle text, which defeats the announcement.
const MaxImprovSteps = 12

// Improvisation errors. Every one of them means the bundle was
// REFUSED — nothing dispatched, nothing committed, nothing logged to
// the replay. A refused improvisation is not a failed one: the runner
// carries on and the policy takes an ordinary legal move instead.
//
// The table IS told, unless the refusal is ErrImprovNoCard: since
// #686 a refusal that can name its card posts a bot_improvisation line
// saying nothing happened. See RefusalAnnouncement.
var (
	ErrImprovNoCard   = errors.New("aiseat: improvisation must name the card")
	ErrImprovNoEffect = errors.New("aiseat: improvisation must state the intended effect")
	ErrImprovNoSteps  = errors.New("aiseat: improvisation has no steps")
	ErrImprovTooLong  = errors.New("aiseat: improvisation has too many steps")
	ErrImprovVerb     = errors.New("aiseat: improvisation may only use the sandbox verbs")
	ErrImprovParams   = errors.New("aiseat: improvisation step has unusable params")
	ErrImprovPlayer   = errors.New("aiseat: improvisation step needs a player")
)

// Improviser is an optional Policy extension: a policy that
// implements it is asked, before each decision, whether it wants to
// apply an effect by hand instead of taking one of the offered moves.
//
// It is out of band for the same reason Conceder is. A Decision names
// an INDEX into the closed move list, which is what makes a policy
// structurally unable to invent an action — and improvising is, by
// definition, doing something the move list does not contain. Rather
// than weaken Decision, improvisation gets its own narrow channel
// with its own validation and its own mandatory disclosure.
//
// Returning false is the overwhelmingly common answer.
//
// ctx carries the runner's hard think deadline, the same one Decide
// gets (ADR 0033 §10, and the 2026-09-19 amendment to §8). It is a
// parameter rather than an ambient assumption because the production
// implementation makes a network call: a hook with no deadline on the
// runner's own goroutine is a hook that can hold the table, and the
// table never waits on a bot. An implementation that overruns it
// simply does not improvise — there is nothing to fall back to and
// nothing that needs one, because not improvising is the normal
// answer.
type Improviser interface {
	Improvise(ctx context.Context, in Input) (Improvisation, bool)
}

// Announcer posts a server-originated chat line. Satisfied by
// *ws.Hub. The runner type-asserts its Broadcaster to this rather
// than widening Broadcaster, so the existing test fakes and the
// lobby wiring in #427 keep compiling unchanged.
type Announcer interface {
	BroadcastChat(gameID uuid.UUID, msg protocol.ChatPayload)
}

// improvVerbs is the allow-list Validate checks against.
var improvVerbs = map[string]bool{
	VerbMoveCard:   true,
	VerbChangeLife: true,
	VerbAddCounter: true,
	VerbMarkDamage: true,
}

// Validate reports whether the improvisation may be applied. It
// checks the announcement first and the mechanics second, which is
// the right order: a bundle that cannot be announced must not be
// applied, however well-formed its verbs are.
func (im Improvisation) Validate() error {
	if strings.TrimSpace(im.Card) == "" {
		return ErrImprovNoCard
	}
	if strings.TrimSpace(im.Effect) == "" {
		return ErrImprovNoEffect
	}
	if len(im.Steps) == 0 {
		return ErrImprovNoSteps
	}
	if len(im.Steps) > MaxImprovSteps {
		return fmt.Errorf("%w: %d > %d", ErrImprovTooLong, len(im.Steps), MaxImprovSteps)
	}
	for i, s := range im.Steps {
		if !improvVerbs[s.Type] {
			return fmt.Errorf("%w: step %d is %q", ErrImprovVerb, i, s.Type)
		}
		if len(s.Params) == 0 || !json.Valid(s.Params) {
			return fmt.Errorf("%w: step %d (%s)", ErrImprovParams, i, s.Type)
		}
		// change_life is player-scoped on the wire and the dispatcher
		// rejects it without a Player. Catching it here means the
		// bundle is refused whole rather than rolled back halfway
		// through, which keeps the failure off the replay.
		if s.Type == VerbChangeLife && s.Player == uuid.Nil {
			return fmt.Errorf("%w: step %d (%s)", ErrImprovPlayer, i, s.Type)
		}
	}
	return nil
}

// Verbs lists the step types in order, for the replay annotation.
func (im Improvisation) Verbs() []string {
	out := make([]string, 0, len(im.Steps))
	for _, s := range im.Steps {
		out = append(out, s.Type)
	}
	return out
}

// Announcement is the chat line the table gets. It names the card,
// states the intended effect, says plainly that it was improvised
// rather than executed by the rules engine, and tells anyone reading
// that they can undo it — because the undo is the recourse, and a
// disclosure nobody knows how to act on is only half a disclosure.
//
// Capped at protocol.MaxChatTextLen. A long effect description is
// truncated rather than dropped: losing the tail of a sentence is
// recoverable from the replay, losing the whole announcement is not.
func (im Improvisation) Announcement() string {
	card := strings.TrimSpace(im.Card)
	effect := strings.TrimSpace(im.Effect)
	const suffix = " (improvised — the rules engine can't run this card, so I applied it by hand; any player can undo it)"
	const ellipsis = "…"
	head := card + ": " + effect
	if room := protocol.MaxChatTextLen - len(suffix); len(head) > room {
		cut := room - len(ellipsis)
		// Back up to a rune boundary so the truncation cannot emit
		// half a multi-byte character.
		for cut > 0 && !utf8.RuneStart(head[cut]) {
			cut--
		}
		if cut > 0 {
			head = head[:cut] + ellipsis
		} else {
			head = ""
		}
	}
	return head + suffix
}

// RefusalAnnouncement is the chat line for a bundle this runner would
// not apply. ADR 0033 §8, amended 2026-09-19 (#686).
//
// §8's refusal was silent, and silence is the failure the whole
// section exists to avoid. The opening paragraph of this file says it:
// a bot that quietly skips a card it cannot run is worse than one that
// improvises, because the table cannot tell the difference between a
// bot choosing not to cast and a bot that could not. A refusal is that
// same ambiguity one step further in — the bot DID try, and the table
// has even more reason to be told, because the card is now sitting in
// a graveyard having done nothing and a human can still apply it by
// hand.
//
// It deliberately does not promise what the seat will do next. The
// runner asks the policy for an ordinary move immediately afterwards
// and that move is frequently not a pass, so a line claiming one
// would be false about half the time.
//
// Empty when the improvisation cannot name its card: there is no
// truthful line to post about a bundle that will not say what it is.
func (im Improvisation) RefusalAnnouncement() string {
	card := strings.TrimSpace(im.Card)
	if card == "" {
		return ""
	}
	const suffix = ": the rules engine can't run this card and my attempt to apply it by hand was refused, so nothing was changed — the card's text did not happen."
	if room := protocol.MaxChatTextLen - len(suffix); len(card) > room {
		cut := room
		for cut > 0 && !utf8.RuneStart(card[cut]) {
			cut--
		}
		card = card[:cut]
	}
	return card + suffix
}

// improvise consults an Improviser policy and, if it wants one,
// validates, applies and announces the bundle. Returns true when an
// improvisation was applied, which tells the runner's act-loop to
// re-enumerate before deciding — the board just moved.
//
// Mirrors concede's shape deliberately: both are things a policy
// expresses outside Decision, both go through the room, both are
// logged.
func (r *Runner) improvise(ctx context.Context, in Input, started time.Time) bool {
	p, ok := r.policy.(Improviser)
	if !ok {
		return false
	}
	// The policy gets the runner's own hard deadline. Improvisation
	// is a decision the table is waiting on exactly like any other,
	// and the production improviser dials a model inside this call.
	ictx, cancel := context.WithTimeout(ctx, r.cfg.MaxThink)
	im, want := p.Improvise(ictx, in)
	cancel()
	if !want {
		return false
	}
	if err := im.Validate(); err != nil {
		// Refused, not applied. The policy gets no board change; it
		// will be asked for an ordinary move next. Logged at Warn
		// because a policy trying to improvise something it may not
		// is worth seeing.
		//
		// The table is told, when there is something truthful to tell
		// it — see RefusalAnnouncement. A bundle that could not name
		// its card is the one refusal that stays silent.
		r.improvRefused.Add(1)
		r.log.Warn("bot improvisation refused", "err", err, "card", im.Card)
		if line := im.RefusalAnnouncement(); line != "" {
			r.announce(protocol.ChatKindBotImprovisation, line, im.Reason)
		}
		return false
	}
	if ctx.Err() != nil {
		return false
	}

	text := im.Announcement()
	annotation := &protocol.ReplayAnnotation{
		Tag:      protocol.ReplayTagBotImprovisation,
		Seat:     r.seat.String(),
		SeatName: r.seatName(),
		Card:     strings.TrimSpace(im.Card),
		Effect:   strings.TrimSpace(im.Effect),
		Text:     text,
		Reason:   im.Reason,
		Steps:    im.Verbs(),
	}

	steps := make([]func() error, 0, len(im.Steps))
	for _, s := range im.Steps {
		steps = append(steps, func() error {
			return actions.Dispatch(r.room.Game, actions.Action{
				Type:   actions.Type(s.Type),
				Player: s.Player,
				// Caller uuid.Nil, per ADR 0033 §8. Two things follow
				// and both are wanted: the seated-caller gates are
				// bypassed (an improvised removal spell has to reach
				// an opponent's board), and the resulting undo entry
				// is nil-stamped, so Room.Undo's caller gate lets any
				// seated human pop it without the admin token.
				Caller: uuid.Nil,
				Params: s.Params,
			})
		})
	}

	r.pace(ctx, started)
	if ctx.Err() != nil {
		return false
	}

	view, seq, err := r.room.ApplyBundle(ws.Bundle{
		Caller: uuid.Nil,
		Steps:  steps,
		// Undoing a bot's improvisation is maintenance on a catalog
		// gap, not a take-back of your own play — see
		// ws.undoEntry.freeUndo.
		FreeUndo:   true,
		Annotation: annotation,
	})
	if err != nil {
		// ApplyBundle rolled the whole thing back, so the board is
		// exactly as the policy found it and nothing was announced.
		r.improvRefused.Add(1)
		r.log.Warn("bot improvisation rejected", "card", im.Card, "err", err)
		return false
	}

	r.applied.Add(1)
	r.improvisations.Add(1)
	r.log.Info("bot improvised",
		"card", im.Card, "effect", im.Effect, "reason", im.Reason,
		"steps", strings.Join(im.Verbs(), ","), "seq", seq)
	// Announce AFTER the commit: a line about a bundle that turned
	// out to be unapplicable would be a worse lie than silence. The
	// replay annotation landed inside the same commit, so the record
	// is atomic with the change even if the chat frame is dropped.
	r.announce(protocol.ChatKindBotImprovisation, text, im.Reason)
	if r.bc != nil {
		r.bc.BroadcastState(r.room.Game.ID, seq, view)
	}
	return true
}

// narrate posts the policy's reason for an ordinary move. Every
// client receives it; only clients with "show bot reasoning" on
// render it (the server has no per-player setting store — S11.5
// settings live in the browser — so the filter is necessarily on the
// client side). Off by default at the server too, via Config.Narrate,
// so a table that does not want the traffic does not carry it.
//
// Passes are skipped. They are most of what a bot does in a
// four-player game and "pass" narrates itself.
func (r *Runner) narrate(label, reason string) {
	if !r.cfg.Narrate || strings.TrimSpace(reason) == "" {
		return
	}
	r.announce(protocol.ChatKindBotReasoning, label, reason)
}

// announce posts one chat line as this seat, if the broadcaster can
// carry chat. A nil or chat-less broadcaster (every test, and the
// bot-vs-bot soak harness) loses the line; the replay annotation and
// the server log are what make the record durable.
func (r *Runner) announce(kind, text, reason string) {
	if r.bc == nil {
		return
	}
	a, ok := r.bc.(Announcer)
	if !ok {
		return
	}
	a.BroadcastChat(r.room.Game.ID, protocol.ChatPayload{
		AuthorID: r.seat.String(),
		// The seat's own name, not a "[BOT]" prefix: `kind` already
		// marks the line as a bot's, it cannot be spoofed by a player
		// who names themselves "[BOT] Kess", and the is_bot flag on
		// PlayerView (#427) is where the seat treatment belongs.
		AuthorName: r.seatName(),
		Text:       text,
		Kind:       kind,
		Reason:     reason,
	})
}

// seatName resolves this seat's display name, or the short seat ID if
// the player has gone (conceded and been reaped, say).
func (r *Runner) seatName() string {
	if p := r.room.Game.PlayerByID(r.seat); p != nil && p.Name != "" {
		return p.Name
	}
	return r.seat.String()[:8]
}
