// Package legal enumerates the moves a seat may make right now.
//
// It is the keystone of the S31 bot work (ADR 0024 §1): a policy —
// heuristic or model-backed — never invents an action, it picks an
// index into the closed list this package produces. The same list is
// what the client will consume so its timing predicates stop
// re-deriving the rules in TypeScript.
//
// Contract. Every Move returned by EnumerateFor would be ACCEPTED by
// actions.Dispatch if sent as-is by that seat right now — that is the
// soundness guarantee a bot depends on, and the tests hold it by
// dispatching every enumerated move against a cloned game. The
// converse is deliberately NOT promised: the engine is a sandbox and
// accepts plenty (a fifth land, a tapped attacker, a sacrifice for no
// reason) that is not a legal move in Magic. Where the engine is lax,
// this package is strict; where a verb is a sandbox affordance rather
// than a game action (move_card, change_life, advance_step, loyalty
// deltas), it is not enumerated at all.
//
// Locking. EnumerateFor takes the game's read lock through
// ReadSnapshot (which also refreshes the layer engine) and uses only
// the *ForEffect / *Locked read surfaces inside it, exactly as the
// protocol projection does. Callers must not hold g.mu.
package legal

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kind groups moves so a policy can reason about them coarsely
// ("is there anything but pass?") before looking at labels.
type Kind string

const (
	KindPass     Kind = "pass"
	KindLand     Kind = "land"
	KindCast     Kind = "cast"
	KindActivate Kind = "activate"
	KindMana     Kind = "mana"
	KindAttack   Kind = "attack"
	KindBlock    Kind = "block"
	KindChoice   Kind = "choice"
	KindMulligan Kind = "mulligan"
)

// Wire action types this package emits. Kept as strings rather than
// actions.Type so protocol may one day import legal without a cycle
// (actions imports protocol). legal_test asserts each equals the
// actions constant it mirrors.
const (
	TypePassPriority        = "pass_priority"
	TypeCastSpell           = "cast_spell"
	TypeActivateAbility     = "activate_ability"
	TypeActivateManaAbility = "activate_mana_ability"
	TypeDeclareAttacker     = "declare_attacker"
	TypeDeclareBlocker      = "declare_blocker"
	TypeResolveChoice       = "resolve_choice"
	TypeKeepHand            = "keep_hand"
	TypeMulligan            = "mulligan"
	TypeDiscardSelection    = "discard_selection"
)

// Move is one fully-specified thing a seat may do right now. Type
// and Params are exactly the wire ActionPayload fields that perform
// it; Player is the seat (the wire Player field for player-scoped
// actions). Label is human-readable and stable enough to show in a
// menu or feed to a model. Source is the instance ID of the card the
// move is about, when there is one.
type Move struct {
	Type   string          `json:"type"`
	Player uuid.UUID       `json:"player"`
	Params json.RawMessage `json:"params,omitempty"`
	Kind   Kind            `json:"kind"`
	Label  string          `json:"label"`
	Source uuid.UUID       `json:"source,omitempty"`

	// AlwaysLegal marks a move the engine cannot refuse whatever
	// else happens between enumeration and dispatch: passing
	// priority, and a choice kind's one unconditional answer (a
	// search's "fail to find", CR 701.19c).
	//
	// Every enumerated move is legal when it is enumerated — that is
	// this package's contract. AlwaysLegal is the stronger claim
	// that it is STILL legal after the board has moved or after some
	// other answer to the same prompt was rejected, and it exists so
	// an automated seat has somewhere to go when its preferred move
	// keeps bouncing. A seat owing a choice is offered nothing but
	// that choice's answers, so without it a bot whose every
	// considered answer is rejected has no pass to fall back on and
	// stops playing, holding the table (#544, #499).
	//
	// Only mark an answer whose acceptance depends on nothing. When
	// in doubt leave a kind unmarked: the cost is a bot that sleeps,
	// which is what it did before.
	AlwaysLegal bool `json:"always_legal,omitempty"`
	// Cost is what the move charges its own controller beyond the
	// mana, in the components Params cannot name. Nil — the
	// overwhelming majority — means "nothing but mana and the
	// choices Params already lists".
	Cost *MoveCost `json:"cost,omitempty"`
}

// MoveCost is the half of a move's price that Params does not carry.
//
// Params names every cost component the ACTOR chooses — which
// permanent is sacrificed, which cards are discarded — because the
// engine needs those choices to perform the move at all. It says
// nothing about the components the card simply CHARGES: Necropotence's
// 1 life, Griselbrand's 7, a planeswalker's −3. Those live on the
// ability shape in internal/game, which a policy may not import
// (ADR 0033 §3), so a policy holding only the wire payload prices
// "Pay 7 life: Draw seven cards" exactly like "{T}: Add {G}" — and a
// bot handed Griselbrand activates itself to 0 (#74).
//
// It rides beside Params rather than inside it on purpose. The Move
// doc promises Params is EXACTLY the ActionPayload that performs the
// move, so that {type, player, params} can be sent back unaltered;
// a field the dispatcher never reads has no business in there. This
// one is advice about the move, not part of it.
//
// A POINTER so that the frame budget is untouched by moves that cost
// nothing: no `cost` key is serialised for them at all.
type MoveCost struct {
	// Life is what the controller pays at announce (CR 118.4 /
	// 118.8). Always positive — the enumerator has already checked
	// the seat can pay it, and a seat that can pay exactly its whole
	// life total legally may, which is the trap.
	Life int `json:"life,omitempty"`

	// Loyalty is a loyalty ability's counter delta (CR 606.1),
	// SIGNED as printed: +1 adds a counter, −3 removes three. Zero
	// covers both "[0]" and "not a loyalty ability"; the two are the
	// same price even though they are not the same thing, and a
	// policy that needs to tell them apart has the label.
	Loyalty int `json:"loyalty,omitempty"`
}

// moveCost returns the MoveCost for a set of components, or nil when
// they are all free. Nil rather than a zero struct so the pointer's
// omitempty does the work and a cost-free move stays byte-identical
// on the wire to what it was before MoveCost existed.
func moveCost(life, loyalty int) *MoveCost {
	if life == 0 && loyalty == 0 {
		return nil
	}
	return &MoveCost{Life: life, Loyalty: loyalty}

}

// Options tunes the expansion caps. The zero value is usable.
type Options struct {
	// MaxExpansionPerSource caps how many concrete moves one source
	// card may expand into across its targets, modes and cost
	// choices. Combat and choices are capped separately by their own
	// structure. Default 12.
	MaxExpansionPerSource int

	// MaxX caps the X the enumerator will try for an {X} spell when
	// searching for the largest affordable value. Default 20.
	MaxX int
}

const (
	defaultMaxExpansionPerSource = 12
	defaultMaxX                  = 20
)

func (o Options) withDefaults() Options {
	if o.MaxExpansionPerSource <= 0 {
		o.MaxExpansionPerSource = defaultMaxExpansionPerSource
	}
	if o.MaxX <= 0 {
		o.MaxX = defaultMaxX
	}
	return o
}

// EnumerateFor returns every move seat may make in g right now, in a
// stable order: pass first, then lands, casts, activations, mana
// abilities, combat declarations, pending choices, mulligan window.
// Nil when the seat is unknown or the game is not active.
func EnumerateFor(g *game.Game, seat uuid.UUID) []Move {
	return EnumerateForWithOptions(g, seat, Options{})
}

// EnumerateForWithOptions is EnumerateFor with explicit caps.
func EnumerateForWithOptions(g *game.Game, seat uuid.UUID, opts Options) []Move {
	if g == nil || seat == uuid.Nil {
		return nil
	}
	var out []Move
	g.ReadSnapshot(func() {
		out = enumerateLocked(g, seat, opts.withDefaults())
	})
	return out
}

// EnumerateLocked is EnumerateFor for a caller that ALREADY holds
// g's read lock — protocol.ViewOfGame builds its whole snapshot
// inside one ReadSnapshot, and stamping the move list from in there
// is what keeps the moves and the frame describing the same instant.
//
// Calling EnumerateFor from that position instead would take the
// read lock a second time, which sync.RWMutex only tolerates while
// no writer is queued: one Apply arriving mid-projection turns it
// into a deadlock. Callers that do NOT hold the lock must use
// EnumerateFor.
func EnumerateLocked(g *game.Game, seat uuid.UUID, opts Options) []Move {
	if g == nil || seat == uuid.Nil {
		return nil
	}
	return enumerateLocked(g, seat, opts.withDefaults())
}

// enumerateLocked is the lock-held core. Every helper it calls reads
// through the game's *ForEffect surfaces and never takes g.mu.
func enumerateLocked(g *game.Game, seat uuid.UUID, opts Options) []Move {
	if g.State != game.StateActive {
		return nil
	}
	p := playerByID(g, seat)
	if p == nil || p.Eliminated {
		return nil
	}
	e := &enumerator{g: g, p: p, seat: seat, opts: opts}

	// The mulligan window is its own world: the cursor is parked on
	// Untap, nobody holds priority, and the only verbs are keep and
	// mulligan.
	if g.MulligansOpen {
		e.mulliganMoves()
		return e.out
	}

	// Pending choices come first and, when one is owed by this seat,
	// they are the ONLY moves: the engine refuses pass_priority while
	// any choice is open (the client mirrors this in canPassPriority),
	// and a cast while a damage-assignment prompt is up is not a
	// decision anyone should be offered.
	if e.choiceMoves() {
		return e.out
	}
	if anyChoiceOpen(g) {
		return e.out
	}
	// Cleanup-step discard (CR 514.1): the cursor parks at Cleanup with
	// no priority holder until the active player discards down to
	// their maximum hand size. Nothing else is legal meanwhile.
	if e.cleanupDiscardMoves() {
		return e.out
	}

	holds := holdsPriority(g, seat)
	if holds {
		e.out = append(e.out, Move{
			Type:        TypePassPriority,
			Player:      seat,
			Kind:        KindPass,
			Label:       "Pass priority",
			AlwaysLegal: true,
		})
		e.castMoves()
		e.activatedMoves()
		e.manaMoves()
	}
	// Combat declarations are not priority-gated in the engine
	// (declare_attacker / declare_blocker only check the step and the
	// card's controller), and a defender must be able to block while
	// the active player still holds priority — see ADR 0024 §2.
	e.combatMoves()
	return e.out
}

type enumerator struct {
	g    *game.Game
	p    *game.Player
	seat uuid.UUID
	opts Options
	out  []Move
}

func (e *enumerator) add(m Move) { e.out = append(e.out, m) }

// mustJSON marshals a params struct; the structs are ours, so a
// failure is a programming error, not a runtime condition.
func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("legal: marshal params: %v", err))
	}
	return b
}

func playerByID(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

func holdsPriority(g *game.Game, seat uuid.UUID) bool {
	ph := g.Turn.PriorityHolder
	if ph < 0 || ph >= len(g.Seats) || g.Seats[ph] == nil {
		return false
	}
	return g.Seats[ph].ID == seat
}

func isActiveSeat(g *game.Game, seat uuid.UUID) bool {
	as := g.Turn.ActiveSeat
	if as < 0 || as >= len(g.Seats) || g.Seats[as] == nil {
		return false
	}
	return g.Seats[as].ID == seat
}

func stackEmpty(g *game.Game) bool {
	if g.Stack != nil && len(g.Stack.Cards) > 0 {
		return false
	}
	for _, item := range g.StackMeta {
		if item != nil {
			return false
		}
	}
	return true
}

func isMainPhase(g *game.Game) bool {
	return g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain
}

// sorcerySpeedOpen mirrors game.sorcerySpeedOpenLocked (CR 307.1):
// main phase, empty stack, and the seat is the active player.
func sorcerySpeedOpen(g *game.Game, seat uuid.UUID) bool {
	return isMainPhase(g) && stackEmpty(g) && isActiveSeat(g, seat)
}

func anyChoiceOpen(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil {
			return true
		}
	}
	return false
}

func findBattlefield(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func cardName(g *game.Game, id uuid.UUID) string {
	if c := findBattlefield(g, id); c != nil {
		return c.Name
	}
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, z := range []*game.Zone{p.Hand, p.Graveyard, p.Command, p.Library} {
			if z == nil {
				continue
			}
			for _, c := range z.Cards {
				if c.InstanceID == id {
					return c.Name
				}
			}
		}
	}
	for _, z := range []*game.Zone{g.Stack, g.Exile} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if c.InstanceID == id {
				return c.Name
			}
		}
	}
	return id.String()[:8]
}

// unknownCardLabel is what a move label calls a card the seat is not
// entitled to recognise. Deliberately not the instance ID prefix
// cardName falls back to: a stable eight-hex handle for a card in
// somebody else's hand is exactly the correlation key ADR 0033 §3
// exists to withhold.
const unknownCardLabel = "a card"

// cardNameFor is cardName with the seat's own knowledge applied: a
// card the seat is not a knower of has no name this seat may be told.
//
// It exists for the branches that enumerate over a pool the seat does
// not own — the discard clause reads FromPlayer's hand, the search
// clause reads a library — where cardName would happily read the
// authoritative Card.Name and paste it into a Label that then travels
// to a Policy (ADR 0033 §3) and, since sub-PR 2, onto the wire as the
// viewer's own legal_moves.
//
// Today this changes nothing, and that is the point of having it
// rather than a comment saying "careful here". Both callers are safe
// by accident of their creators: QueueDiscardFromRevealedHand is the
// only path that makes a chooser != discarder choice and it reveals
// the hand to the chooser first, and queueSearchChoiceLocked marks
// the searcher a knower of every match. Neither guarantee is stated
// anywhere the author of the NEXT coercive-discard card would read.
// This makes the enumerator hold the line itself.
func cardNameFor(g *game.Game, id, seat uuid.UUID) string {
	if c := findCardAnywhere(g, id); c != nil && !c.IsKnownTo(seat) {
		return unknownCardLabel
	}
	return cardName(g, id)
}

// findCardAnywhere returns the card with this instance ID from any
// zone, or nil. Same search order as cardName.
func findCardAnywhere(g *game.Game, id uuid.UUID) *game.Card {
	if c := findBattlefield(g, id); c != nil {
		return c
	}
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, z := range []*game.Zone{p.Hand, p.Graveyard, p.Command, p.Library} {
			if z == nil {
				continue
			}
			for i := range z.Cards {
				if z.Cards[i].InstanceID == id {
					return &z.Cards[i]
				}
			}
		}
	}
	for _, z := range []*game.Zone{g.Stack, g.Exile} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == id {
				return &z.Cards[i]
			}
		}
	}
	return nil
}

func playerName(g *game.Game, id uuid.UUID) string {
	if p := playerByID(g, id); p != nil {
		return p.Name
	}
	return id.String()[:8]
}

// targetWire is the {kind, id} shape cast_spell / activate_ability /
// resolve_choice all use for a target reference.
type targetWire struct {
	Kind string `json:"kind"`
	ID   string `json:"id,omitempty"`
}

func wireTargets(refs []game.TargetRef) []targetWire {
	if len(refs) == 0 {
		return nil
	}
	out := make([]targetWire, 0, len(refs))
	for _, r := range refs {
		out = append(out, targetWire{Kind: string(r.Kind), ID: r.ID.String()})
	}
	return out
}

func targetLabel(g *game.Game, refs []game.TargetRef) string {
	if len(refs) == 0 {
		return ""
	}
	s := " targeting "
	for i, r := range refs {
		if i > 0 {
			s += " and "
		}
		switch r.Kind {
		case game.TargetPlayer:
			s += playerName(g, r.ID)
		default:
			s += cardName(g, r.ID)
		}
	}
	return s
}

func idStrings(ids []uuid.UUID) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
