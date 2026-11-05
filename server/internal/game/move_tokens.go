package game

import "github.com/google/uuid"

// move_tokens.go — a token creation after a replaced move (ADR 0178
// amendment of 2026-11-05, #4522).
//
// Kalitas, Traitor of Ghet: "If a nontoken creature an opponent controls
// would die, instead exile that card and create a 2/2 black Zombie
// creature token." The modified event is two instructions in printed
// order (CR 608.2c, 614.6): the exile, then the creation. Unlike Head of
// the Hunt's "when you do" (move_reflexive.go) the token is not a
// trigger: nothing goes on the stack and nobody can respond between the
// exile and the Zombie.
//
// The creation is a CR 701.7b event of its own, with its own CR 614
// window, so it can pause on a CR 616 ordering prompt (a doubler and a
// kind-swapper under the token's controller), and so can each token's
// entry. Like a life change (ADR 0178), the replacement only DECLARES the
// creation on the move event (CreateTokensAfterReplacedMoveForEffect),
// and the move's finisher (runMoveTailLocked) runs it once the card has
// landed. The move's own continuation (the next leg of a wipe, "then
// draw a card" behind a destroy) waits for it.
//
// A creation's continuation is a keyed tail all the way down (the
// creation's tokenTail, then the staged entries' tokenEntryRun), and a
// route cannot ride a keyed tail. So the rest of the move is PARKED on
// the game, the way an entry choice's frame is frozen behind its discard
// (Game.frozenEntries, entry_card_choice.go): a carrier event holding the
// route and the instructions not yet run, under an event ID the
// creation's tail names (moveTokensCreated). The tail takes the carrier
// back and goes on from a fresh copy, so an undo snapshot that shares the
// parked slice replays the same continuation instead of finding it
// consumed. The carrier is written to a restore point through the
// replacement-event mirror (GameSnapshot.MovesAwaitingTokens).
//
// The instruction survives a later replacement that redirects or cancels
// the move (CR 614.6: an impossible instruction is ignored, the others
// still happen), exactly as a declared life change does: a commander
// whose owner then moves it to the command zone still leaves a Zombie
// behind. A move abandoned before it happens discards it (CR 614.7).

// tokenMoveInstruction is one declared token creation: the source the
// creation names, the player who controls the tokens and what they are.
// All three are fixed as the replacement applies, so a source that
// leaves afterwards does not change who gets the tokens.
type tokenMoveInstruction struct {
	Source     uuid.UUID
	Controller uuid.UUID
	Groups     []TokenGroup
}

// cloneTokenMoveInstructions gives each instruction its own groups.
func cloneTokenMoveInstructions(in []tokenMoveInstruction) []tokenMoveInstruction {
	if in == nil {
		return nil
	}
	out := make([]tokenMoveInstruction, len(in))
	for i, t := range in {
		out[i] = t
		out[i].Groups = cloneTokenGroups(t.Groups)
	}
	return out
}

// CreateTokensAfterReplacedMoveForEffect declares, from a replacement's
// Replace callback, that the modified zone-change event also creates the
// tokens spec describes once the card has moved ("instead exile that
// card and create a 2/2 black Zombie creature token"). The creation runs
// through the ordinary token pipeline, so token and entry replacements
// apply to it and it may ask a CR 616 question; the move's continuation
// waits for it.
//
// spec names the Controller, the Source and the Groups. A creation
// attacking or blocking, or entering together, is refused: no card needs
// one. A creation of nothing, or for no controller, declares nothing.
//
// Returns ErrInvalidParam for an event that is not a zone change, or a
// refused spec.
//
// Caller must hold g.mu (Replace callbacks always do).
func (g *Game) CreateTokensAfterReplacedMoveForEffect(ev *ReplacementEvent, spec TokenCreation) error {
	if ev == nil || ev.Kind != RepEventMove ||
		spec.Attacking != uuid.Nil || spec.Together || spec.AttackChoicesLabel != "" {
		return ErrInvalidParam
	}
	groups := make([]TokenGroup, 0, len(spec.Groups))
	for _, grp := range spec.Groups {
		if grp.Attacking != uuid.Nil || grp.AttackOpponent != uuid.Nil || grp.Blocking != uuid.Nil {
			return ErrInvalidParam
		}
		if grp.Count > 0 {
			groups = append(groups, grp)
		}
	}
	if spec.Controller == uuid.Nil || len(groups) == 0 {
		return nil
	}
	// A fresh backing array: an undo snapshot taken while this event was
	// paused shares the old one (cloneReplacementResume copies too).
	n := len(ev.tokensAfterMove)
	ev.tokensAfterMove = append(ev.tokensAfterMove[:n:n], tokenMoveInstruction{
		Source:     spec.Source,
		Controller: spec.Controller,
		Groups:     cloneTokenGroups(groups),
	})
	return nil
}

// continueMoveTokensLocked runs the next declared token creation, with
// the rest of the move parked behind it, or, when none is left, the
// move's life instructions and then its route continuation. A controller
// who has left the game gets none (CR 800.4a).
//
// Caller must hold g.mu.
func (g *Game) continueMoveTokensLocked(tokens []tokenMoveInstruction, seq moveLifeSequence) error {
	for len(tokens) > 0 {
		in := tokens[0]
		tokens = tokens[1:]
		if p := g.playerByIDLocked(in.Controller); p == nil || p.Eliminated {
			continue
		}
		id := g.parkMoveForTokensLocked(tokens, seq)
		ev := &ReplacementEvent{
			Kind:            RepEventCreateTokens,
			Actor:           in.Controller,
			Source:          in.Source,
			TokenController: in.Controller,
			TokenGroups:     cloneTokenGroups(in.Groups),
			// ADR 0178 §5: a move paid as a cost cannot pause, and
			// neither can the creation that is part of it.
			mustSettleNow: seq.mustSettleNow,
			tokenTail:     &tokenTail{keyed: KeyedTail{Ref: moveTokensCreated, N: int(id)}},
		}
		err := g.createTokensLocked(ev)
		if err != nil {
			// The creation ended on an error without reaching its tail.
			// The rest of the move is still owed.
			if p := g.takeMoveAwaitingTokensLocked(id); p != nil {
				if tailErr := g.resumeMoveAfterTokensLocked(p); tailErr != nil {
					g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: tailErr.Error()})
				}
			}
		}
		return err
	}
	return g.continueMoveLifeLocked(seq)
}

// moveTokensCreated is a declared creation's continuation: it takes the
// move parked under the event ID N back and goes on with it. It runs on
// every terminal outcome of the creation, tokens made or not. Registered
// in an init, because its body starts the next creation, which names it.
var moveTokensCreated TailRef

func init() {
	moveTokensCreated = RegisterTail("engine/move-tokens-created", func(g *Game, o TailOutcome) error {
		p := g.takeMoveAwaitingTokensLocked(ReplacementEventID(o.Tail.N))
		if p == nil {
			return nil
		}
		return g.resumeMoveAfterTokensLocked(p)
	})
}

// parkMoveForTokensLocked parks what is left of a move while one of its
// token creations runs: the creations not yet run, the life instructions
// and a copy of the route, on a carrier event under a fresh event ID,
// which it returns.
//
// Caller must hold g.mu.
func (g *Game) parkMoveForTokensLocked(tokens []tokenMoveInstruction, seq moveLifeSequence) ReplacementEventID {
	id := ReplacementEventID(g.nextReplacementEventID.Add(1))
	carrier := &ReplacementEvent{
		ID:              id,
		Kind:            RepEventMove,
		mustSettleNow:   seq.mustSettleNow,
		lifeAfterMove:   append([]lifeInstruction(nil), seq.remaining...),
		tokensAfterMove: cloneTokenMoveInstructions(tokens),
	}
	if seq.hasRoute {
		route := seq.route
		carrier.zoneRoute = cloneZoneRoute(&route)
	}
	next := make([]*ReplacementEvent, 0, len(g.movesAwaitingTokens)+1)
	next = append(next, g.movesAwaitingTokens...)
	g.movesAwaitingTokens = append(next, carrier)
	return id
}

// takeMoveAwaitingTokensLocked removes and returns the move parked under
// id, or nil when there is none. Copy on write: an undo snapshot sharing
// the slice keeps its own.
//
// Caller must hold g.mu.
func (g *Game) takeMoveAwaitingTokensLocked(id ReplacementEventID) *ReplacementEvent {
	var found *ReplacementEvent
	next := make([]*ReplacementEvent, 0, len(g.movesAwaitingTokens))
	for _, p := range g.movesAwaitingTokens {
		if found == nil && p.ID == id {
			found = p
			continue
		}
		next = append(next, p)
	}
	if found == nil {
		return nil
	}
	if len(next) == 0 {
		next = nil
	}
	g.movesAwaitingTokens = next
	return found
}

// resumeMoveAfterTokensLocked goes on with a parked move from a fresh
// copy of its carrier, which is never written: an undo snapshot may
// still hold it.
//
// Caller must hold g.mu.
func (g *Game) resumeMoveAfterTokensLocked(p *ReplacementEvent) error {
	seq := moveLifeSequence{
		remaining:     append([]lifeInstruction(nil), p.lifeAfterMove...),
		mustSettleNow: p.mustSettleNow,
	}
	if p.zoneRoute != nil {
		seq.route = *cloneZoneRoute(p.zoneRoute)
		seq.hasRoute = true
	}
	return g.continueMoveTokensLocked(cloneTokenMoveInstructions(p.tokensAfterMove), seq)
}

// snapshotMovesAwaitingTokensLocked writes Game.movesAwaitingTokens. One
// that is not data is counted instead.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) snapshotMovesAwaitingTokensLocked(cen *ContinuationCensus) []replacementEventSnapshot {
	var out []replacementEventSnapshot
	for _, p := range g.movesAwaitingTokens {
		if !g.replacementEventKeyable(p) {
			cen.ChoiceResumeFrames++
			cen.note("a move parked behind its token creation is not data")
			continue
		}
		out = append(out, g.snapshotReplacementEvent(p, cen))
	}
	return out
}

// restoreMovesAwaitingTokensLocked is snapshotMovesAwaitingTokensLocked's
// inverse. A later event must not reuse a parked ID.
//
// Caller must hold g.mu (write).
func (g *Game) restoreMovesAwaitingTokensLocked(snaps []replacementEventSnapshot) {
	for i := range snaps {
		p := restoreReplacementEvent(&snaps[i])
		if uint64(p.ID) > g.nextReplacementEventID.Load() {
			g.nextReplacementEventID.Store(uint64(p.ID))
		}
		g.movesAwaitingTokens = append(g.movesAwaitingTokens, p)
	}
}
