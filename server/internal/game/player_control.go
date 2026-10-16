package game

import (
	"errors"
	"slices"

	"github.com/google/uuid"
)

// player_control.go — one player controlling another (CR 723, ADR 0166,
// #2164).
//
// Word of Command uses it for CR 723.2's limited form: "You control
// that player until Word of Command finishes resolving. The player
// plays that card if able. … If the chosen card is cast as a spell, you
// control the player while that spell is resolving." Emrakul, the
// Promised End uses it for CR 723.1's whole turn (ADR 0167,
// turn_control.go): a record with Until "turn" waits for the player's
// next turn, applies for all of it, and ends at the rotation seam.
//
// # The record
//
// Game.PlayerControls lists the live player-controlling effects. A
// record names who controls whom, the card that did it, the card the
// controlled player must play, and the stage:
//
//   - ControlPlaying: Word of Command is still resolving. Its resolution
//     waits on a controlled_play prompt (mid-resolution, so it blocks
//     the table and holds the CR 704.3 boundary, resolution_pause.go).
//     The prompt's answer is the forced play. Until it is answered the
//     controlled player's mana comes from their lands alone, and the
//     chosen card ignores the timing its type would impose (CR 608.2g).
//   - ControlSpellResolving: the card was cast. The spell waits on the
//     stack like any other, and the controlled player answers for
//     themselves. Control applies again while that spell resolves, and
//     the record goes once it has.
//
// # Who decides
//
// ControllerOfPlayerLocked is the one question: who makes p's decisions
// right now (CR 723.5)? Prompts keep their real Chooser, because every
// resolver reads it as "whose hand, whose permanents". The actions
// layer re-addresses an answer from the controller to the controlled
// player, the view shows the prompt to the controller, and the
// enumerator offers it to the controller's seat.

// ControlStage is where a player-controlling effect is in its life.
type ControlStage string

const (
	// ControlPlaying is Word of Command still resolving: the controlled
	// player must play the chosen card.
	ControlPlaying ControlStage = "playing"
	// ControlSpellResolving is the forced spell cast and on the stack.
	// Control applies only while that spell resolves.
	ControlSpellResolving ControlStage = "spell_resolving"
	// ControlAwaitingTurn is a turn-long record (ADR 0167 §1) waiting
	// for the controlled player's next turn. It never applies.
	ControlAwaitingTurn ControlStage = "awaiting_turn"
	// ControlTurn is the controlled turn itself: live while both
	// players are in the game and that turn is the current one.
	ControlTurn ControlStage = "turn"
)

// ControlUntil is how long a record's control lasts (ADR 0167 §1).
type ControlUntil string

const (
	// ControlUntilResolution is Word of Command's: control over a
	// resolution (ADR 0166). A record that names no Until (one written
	// before ADR 0167) reads as this.
	ControlUntilResolution ControlUntil = "resolution"
	// ControlUntilTurn is Emrakul's: the whole of the controlled
	// player's next turn (CR 723.1).
	ControlUntilTurn ControlUntil = "turn"
)

// AnchoredExtraTurn is the extra turn a turn-long record owes after the
// controlled turn (ADR 0167 §3): "After that turn, that player takes an
// extra turn." Ref is minted from NextExtraRef when the record is
// created; Floor is NextExtraRef as the controlled turn begins, so the
// extra turns created during it (Ref > Floor) stay ahead of this one
// (CR 500.7).
type AnchoredExtraTurn struct {
	Ref   int `json:"ref"`
	Floor int `json:"floor"`
}

// PlayerControl is one live player-controlling effect (CR 723.2). Plain
// data, so a table under one is still a restore point.
type PlayerControl struct {
	// Controller makes Controlled's decisions (CR 723.5).
	Controller uuid.UUID `json:"controller"`
	Controlled uuid.UUID `json:"controlled"`
	// Source is the card whose effect this is, and SourceName its name
	// for the prompt.
	Source     uuid.UUID `json:"source"`
	SourceName string    `json:"sourceName,omitempty"`
	// Card is the card Controlled must play while ControlPlaying, and
	// the spell on the stack while ControlSpellResolving.
	Card  uuid.UUID    `json:"card"`
	Stage ControlStage `json:"stage"`
	// Prompt is the controlled_play prompt while ControlPlaying.
	// uuid.Nil while the play is being probed, before it is queued.
	Prompt uuid.UUID `json:"prompt,omitempty"`
	// Until is how long the control lasts (ADR 0167 §1). Empty reads
	// as ControlUntilResolution, so a record written before it existed
	// is a Word of Command record.
	Until ControlUntil `json:"until,omitempty"`
	// TurnSeq is the Turn.Seq of the controlled turn, set as it begins
	// (ControlTurn). Zero while the record waits.
	TurnSeq int `json:"turnSeq,omitempty"`
	// ExtraTurn is the extra turn owed after the controlled turn, nil
	// when the effect gives none.
	ExtraTurn *AnchoredExtraTurn `json:"extraTurn,omitempty"`
}

// UntilKind is r.Until with the empty value read as
// ControlUntilResolution.
func (r PlayerControl) UntilKind() ControlUntil {
	if r.Until == "" {
		return ControlUntilResolution
	}
	return r.Until
}

// PendingChoiceControlledPlay is Word of Command's forced play (ADR
// 0164 §4). Chooser is Word's controller, FromPlayer the controlled
// player, DiscardOptions the one card. Its answers are a cast_spell
// that names the prompt (ControlledPlayCast) or an empty resolve_choice
// meaning "it can't be played", which the engine accepts only when that
// is true (the probe below).
const PendingChoiceControlledPlay PendingChoiceKind = "controlled_play"

var (
	// ErrPlayerControlled refuses an action by a player another player
	// controls (CR 723.5). Conceding is never refused (CR 723.6).
	ErrPlayerControlled = errors.New("game: another player is making this player's decisions")
	// ErrControlledPlayPossible refuses "it can't be played" while the
	// card can be played: "The player plays that card if able".
	ErrControlledPlayPossible = errors.New("game: the chosen card can be played, so it must be")
)

// controlledPlayProbe answers whether the controlled player can play the
// card under the record's restrictions. internal/legal registers it at
// init (SetControlledPlayProbe), so the engine accepts "it can't be
// played" exactly when the enumerator offers no play. Nil in a binary
// without internal/legal (this package's own tests): every card is then
// taken to be playable when the prompt is raised, and "it can't be
// played" is accepted.
var controlledPlayProbe func(g *Game, controlled, card uuid.UUID) bool

// SetControlledPlayProbe registers the probe. internal/legal calls it
// from init. The probe runs with g.mu held.
func SetControlledPlayProbe(f func(g *Game, controlled, card uuid.UUID) bool) {
	controlledPlayProbe = f
}

// controlLiveLocked reports whether a record applies right now.
// Caller must hold g.mu.
func (g *Game) controlLiveLocked(r PlayerControl) bool {
	if !g.playerInGameLocked(r.Controller) || !g.playerInGameLocked(r.Controlled) {
		return false
	}
	switch r.Stage {
	case ControlPlaying:
		return r.Prompt == uuid.Nil || g.choiceIndexLocked(r.Prompt) >= 0
	case ControlSpellResolving:
		return g.resolutionOpen && g.resolving != nil && g.resolving.hasCard &&
			g.resolving.card.InstanceID == r.Card
	case ControlTurn:
		return r.UntilKind() == ControlUntilTurn && g.Turn.Seq == r.TurnSeq
	case ControlSearching:
		return g.searchControlLiveLocked(r)
	}
	return false
}

// ControllerOfPlayerLocked is the player who makes p's decisions right
// now: p, unless a player-controlling effect applies (CR 723.5). The
// newest record wins (CR 723.1a). Caller must hold g.mu (read or write).
func (g *Game) ControllerOfPlayerLocked(p uuid.UUID) uuid.UUID {
	for i := len(g.PlayerControls) - 1; i >= 0; i-- {
		r := g.PlayerControls[i]
		if r.Controlled == p && g.controlLiveLocked(r) {
			return r.Controller
		}
	}
	return p
}

// ControllerOfPlayer is ControllerOfPlayerLocked under the read lock.
// Caller must NOT hold g.mu.
func (g *Game) ControllerOfPlayer(p uuid.UUID) uuid.UUID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.ControllerOfPlayerLocked(p)
}

// LandManaOnly is LandManaOnlyLocked under the read lock. Caller must
// NOT hold g.mu.
func (g *Game) LandManaOnly(p uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.LandManaOnlyLocked(p)
}

// ChoiceChooser is the chooser of the open prompt with that ID. Caller
// must NOT hold g.mu.
func (g *Game) ChoiceChooser(id uuid.UUID) (uuid.UUID, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if i := g.choiceIndexLocked(id); i >= 0 {
		return g.PendingChoices[i].Chooser, true
	}
	return uuid.Nil, false
}

// playingRecordLocked is the live ControlPlaying record over p, or nil.
// Caller must hold g.mu.
func (g *Game) playingRecordLocked(p uuid.UUID) *PlayerControl {
	for i := len(g.PlayerControls) - 1; i >= 0; i-- {
		r := &g.PlayerControls[i]
		if r.Controlled == p && r.Stage == ControlPlaying && g.controlLiveLocked(*r) {
			return r
		}
	}
	return nil
}

// LandManaOnlyLocked reports whether p may make mana only with their
// lands' mana abilities: Word of Command's "the player can activate
// mana abilities only if they're from lands that player controls",
// while the forced play is under way. The auto-tapper and
// ActivateManaAbility both ask it. Caller must hold g.mu.
func (g *Game) LandManaOnlyLocked(p uuid.UUID) bool {
	return g.playingRecordLocked(p) != nil
}

// ControlledPlayCardLocked reports whether card is the one p is being
// made to play right now. Such a card is played during a resolution, so
// the timing its type would impose does not apply (CR 608.2g), and a
// land needs only CR 305's turn and land play. Caller must hold g.mu.
func (g *Game) ControlledPlayCardLocked(p, card uuid.UUID) bool {
	r := g.playingRecordLocked(p)
	return r != nil && r.Card == card
}

// LandPlayWindowOpenLocked is CR 305's window for p to play card: their
// main phase with the stack empty, or (ADR 0166) the card is the one
// Word of Command is making them play and it is their turn (CR 305.3).
// The land play allowance is asked separately. Caller must hold g.mu.
func (g *Game) LandPlayWindowOpenLocked(p, card uuid.UUID) bool {
	if g.SorcerySpeedOpenLocked(p) {
		return true
	}
	return g.ControlledPlayCardLocked(p, card) && g.activeSeatIDLocked() == p
}

// choiceIndexLocked is the queue index of the prompt with that ID, or
// -1. Caller must hold g.mu.
func (g *Game) choiceIndexLocked(id uuid.UUID) int {
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == id {
			return i
		}
	}
	return -1
}

// controlRecordForPromptLocked is the index of the record whose prompt
// is id, or -1. Caller must hold g.mu.
func (g *Game) controlRecordForPromptLocked(id uuid.UUID) int {
	for i, r := range g.PlayerControls {
		if r.Stage == ControlPlaying && r.Prompt == id {
			return i
		}
	}
	return -1
}

// ControlledPlay is what an effect asks for.
type ControlledPlay struct {
	// Controller makes the decisions; Controlled plays the card.
	Controller, Controlled uuid.UUID
	// Source is the card whose effect this is.
	Source     uuid.UUID
	SourceName string
	// Card is the chosen card, in Controlled's hand.
	Card uuid.UUID
}

// BeginControlledPlayForEffect starts Word of Command's control and its
// forced play (ADR 0166 §4). The record goes in first, because the
// probe asks the enumerator whether the card can be played under it
// (land mana only, the type's timing lifted). A card that can't be
// played ends the control at once: "The player plays that card if
// able". Otherwise the controlled_play prompt is queued for the
// controller; it is mid-resolution, so the resolution waits for it.
//
// Returns the prompt's ID, or uuid.Nil when nothing was queued.
//
// Caller must hold g.mu.
func (g *Game) BeginControlledPlayForEffect(cp ControlledPlay) uuid.UUID {
	if !g.playerInGameLocked(cp.Controller) || !g.playerInGameLocked(cp.Controlled) {
		return uuid.Nil
	}
	p := g.playerByIDLocked(cp.Controlled)
	if p == nil || p.Hand == nil || !p.Hand.Contains(cp.Card) {
		return uuid.Nil
	}
	// CR 723.1a: the newest effect is the one that works. Only one
	// card makes these, so an older Word record over the same player
	// can only be a stale one; it goes. A turn-long record (ADR 0167)
	// stays: Word of Command cast during an Emrakul turn controls the
	// player while it resolves, and the turn's record applies again
	// afterwards. A fresh slice, never a write into the old one, which
	// an undo entry may share (RestoreFrom).
	kept := make([]PlayerControl, 0, len(g.PlayerControls)+1)
	for _, r := range g.PlayerControls {
		if r.Controlled == cp.Controlled && r.UntilKind() == ControlUntilResolution {
			continue
		}
		kept = append(kept, r)
	}
	g.PlayerControls = append(kept, PlayerControl{
		Controller: cp.Controller,
		Controlled: cp.Controlled,
		Source:     cp.Source,
		SourceName: cp.SourceName,
		Card:       cp.Card,
		Stage:      ControlPlaying,
		Until:      ControlUntilResolution,
	})
	if controlledPlayProbe != nil && !controlledPlayProbe(g, cp.Controlled, cp.Card) {
		g.removeControlRecordLocked(len(g.PlayerControls) - 1)
		return uuid.Nil
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:           PendingChoiceControlledPlay,
		Chooser:        cp.Controller,
		FromPlayer:     cp.Controlled,
		Count:          1,
		Source:         cp.Source,
		Reason:         cp.SourceName,
		DiscardOptions: []uuid.UUID{cp.Card},
	})
	if id == uuid.Nil {
		g.removeControlRecordLocked(len(g.PlayerControls) - 1)
		return uuid.Nil
	}
	g.PlayerControls[len(g.PlayerControls)-1].Prompt = id
	return id
}

// removeControlRecordLocked drops the record at i. Caller must hold g.mu.
func (g *Game) removeControlRecordLocked(i int) {
	if i < 0 || i >= len(g.PlayerControls) {
		return
	}
	g.PlayerControls = slices.Delete(g.PlayerControls, i, i+1)
	if len(g.PlayerControls) == 0 {
		g.PlayerControls = nil
	}
}

// ControlledPlayCast is the forced play's "play it" answer: the
// controller casts the chosen card as the controlled player (CR 723.5,
// 608.2g), or plays it if it is a land. Every choice the cast asks for
// is in params, made by the controller. Refused, with the prompt left
// open:
//
//   - a prompt that is not this caller's controlled_play
//     (ErrPendingChoiceNotFound, ErrNotTheChooser), or another card
//     (ErrInvalidParam);
//   - a cast the engine refuses (whatever castSpellLocked returns: a
//     cost the lands can't pay, a target that is not legal). The
//     controller tries again or, if nothing works, says it can't be
//     played.
//
// "Cast anyway" is not offered: the card is played for what it costs.
// No player receives priority after it (CR 608.2g), so no action is
// noted. Then the prompt is answered, Word of Command's resolution can
// finish, and control moves to ControlSpellResolving for a spell or
// ends for a land.
//
// Caller must NOT hold g.mu.
func (g *Game) ControlledPlayCast(caller, choiceID, cardID uuid.UUID, params CastSpellParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := g.choiceIndexLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	c := g.PendingChoices[idx]
	if c.Kind != PendingChoiceControlledPlay {
		return ErrPendingChoiceNotFound
	}
	if c.Chooser != caller {
		return ErrNotTheChooser
	}
	if len(c.DiscardOptions) != 1 || c.DiscardOptions[0] != cardID {
		return ErrInvalidParam
	}
	ri := g.controlRecordForPromptLocked(choiceID)
	if ri < 0 {
		return ErrPendingChoiceNotFound
	}
	controlled := g.PlayerControls[ri].Controlled
	// The card is played for what it costs, from the controlled
	// player's lands (and pool), whatever the table's strict-mana
	// setting: paying is part of what Word of Command makes them do. The
	// controller chooses which lands by locking the rest
	// (LockedSources), and the auto-tapper reaches only lands
	// (LandManaOnlyLocked).
	params.ForceCast = false
	params.Strict = true
	params.AutoTap = true
	land := false
	if card, ok := g.LookupCardForEffect(cardID); ok {
		land = card.IsLand()
	}
	if err := g.castSpellLocked(controlled, cardID, params); err != nil {
		return err
	}
	// The cast may have queued or answered prompts, so both indexes are
	// looked up again.
	if i := g.choiceIndexLocked(choiceID); i >= 0 {
		g.dequeueChoiceLocked(i)
	}
	ri = g.controlRecordForPromptLocked(choiceID)
	if ri < 0 {
		return nil
	}
	if land || g.Stack == nil || !g.Stack.Contains(cardID) {
		g.removeControlRecordLocked(ri)
		return nil
	}
	g.PlayerControls[ri].Stage = ControlSpellResolving
	g.PlayerControls[ri].Prompt = uuid.Nil
	return nil
}

// resolveControlledPlayLocked answers a controlled_play prompt with
// resolve_choice. The only answer it takes is the empty one, "it can't
// be played", and only when the probe agrees (ErrControlledPlayPossible
// otherwise). The play itself is ControlledPlayCast. Caller must hold
// g.mu.
func (g *Game) resolveControlledPlayLocked(idx int, c *PendingChoice, picks []uuid.UUID) error {
	if len(picks) != 0 {
		return ErrInvalidParam
	}
	ri := g.controlRecordForPromptLocked(c.ID)
	if ri >= 0 && len(c.DiscardOptions) == 1 && controlledPlayProbe != nil &&
		controlledPlayProbe(g, g.PlayerControls[ri].Controlled, c.DiscardOptions[0]) {
		return ErrControlledPlayPossible
	}
	g.dequeueChoiceLocked(idx)
	if ri = g.controlRecordForPromptLocked(c.ID); ri >= 0 {
		g.removeControlRecordLocked(ri)
	}
	return nil
}

// prunePlayerControlsLocked drops the records that no longer apply and
// never will again: either player has left the game (CR 800.4a ends an
// effect that gives a leaving player control of a player), a forced
// play whose prompt is gone, or a forced spell that has left the stack
// and is not the resolution still open. A forced play whose controlled
// player left has its prompt withdrawn too. runStateChecksLocked calls
// it once a resolution has closed.
//
// A search-scoped record (#3998) goes once its search prompt is gone.
//
// A turn-long record (ADR 0167 §6) goes here only while it waits and
// its controlled player has left: that turn never begins, so there is
// neither control nor an extra turn. Every other turn-long record ends
// at the rotation seam (endTurnControlsLocked), which still owes the
// extra turn when only the controller has left. Caller must hold g.mu.
func (g *Game) prunePlayerControlsLocked() {
	if len(g.PlayerControls) == 0 {
		return
	}
	for i := len(g.PlayerControls) - 1; i >= 0; i-- {
		r := g.PlayerControls[i]
		gone := !g.playerInGameLocked(r.Controller) || !g.playerInGameLocked(r.Controlled)
		switch {
		case r.UntilKind() == ControlUntilTurn:
			if r.Stage != ControlAwaitingTurn || g.playerInGameLocked(r.Controlled) {
				continue
			}
		case r.Stage == ControlSearching:
			// #3998: the search is over once its prompt is answered or
			// withdrawn. A departed player ends the control, never the
			// search: the prompt stays the searcher's to answer.
			if !gone && r.Prompt != uuid.Nil && g.choiceIndexLocked(r.Prompt) >= 0 {
				continue
			}
		case gone:
			if r.Stage == ControlPlaying && r.Prompt != uuid.Nil {
				if idx := g.choiceIndexLocked(r.Prompt); idx >= 0 {
					g.dropChoiceLocked(idx)
				}
			}
		case r.Stage == ControlPlaying:
			if r.Prompt == uuid.Nil || g.choiceIndexLocked(r.Prompt) >= 0 {
				continue
			}
		case r.Stage == ControlSpellResolving:
			if g.Stack != nil && g.Stack.Contains(r.Card) {
				continue
			}
			if g.controlLiveLocked(r) {
				continue
			}
		}
		g.removeControlRecordLocked(i)
	}
}

// clonePlayerControls copies the record list, ExtraTurn included, so a
// clone never shares a pointer with the game it came from.
func clonePlayerControls(in []PlayerControl) []PlayerControl {
	if len(in) == 0 {
		return nil
	}
	out := slices.Clone(in)
	for i := range out {
		if out[i].ExtraTurn != nil {
			et := *out[i].ExtraTurn
			out[i].ExtraTurn = &et
		}
	}
	return out
}
