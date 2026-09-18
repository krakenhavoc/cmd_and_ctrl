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
			Type:   TypePassPriority,
			Player: seat,
			Kind:   KindPass,
			Label:  "Pass priority",
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
