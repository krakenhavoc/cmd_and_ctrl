package game

// snapshot.go gives Game a SERIALISABLE deep copy — the on-disk twin
// of clone.go's in-memory one — so that a game survives the process
// that was hosting it. Deploying new server code should cost the
// table a page refresh, not the game.
//
// Why not reuse the existing dump? Every Apply already writes
// <dumpDir>/games/<id>.json, but that file holds a protocol.GameView:
// a redacted, denormalised projection built for a browser. It cannot
// reconstruct a *Game and was never meant to. This is the domain
// type, written whole.
//
// Why not replay the action log? Replay re-derives state by RUNNING
// THE CODE. The entire point of persisting is to deploy changed code,
// and changed card logic makes a replay diverge from the game the
// players were actually in. Sound for crash recovery on an identical
// binary; useless across the version boundary that motivates this.
// So: state, not derivation.
//
// clone.go is the checklist. Every field cloneLocked copies is a
// field this file must carry, and snapshot_drift_test.go fails the
// build when the domain grows a field neither of them knows about.
//
// ---------------------------------------------------------------
// WHAT DOES NOT SURVIVE, AND WHY
//
// Parts of a live game are Go function pointers, not data:
//
//   - StackItem.Effect — what an ability does when it resolves
//   - StackItem.targetSpec — the clause its targets were legal under
//   - DelayedTrigger.Effect — "at the beginning of the next end step"
//   - PendingChoice's six resume frames + scryResume — a paused
//     game literally holds the rest of the effect as a continuation
//   - ScopedStatic.Ability's AppliesTo / Apply — Giant Growth's +3/+3
//   - TurnScopedReplacements' AppliesTo / Replace — Fog
//   - Card.ManaAbilities / ActivatedAbilities on a token, which has
//     no oracle ID for the catalog to key on
//
// Some of those are re-derivable, because the closure was looked up
// from the card catalog by oracle ID in the first place, and the
// catalog is code — it is rebuilt when the new binary starts. Restore
// re-runs those lookups (see rehydrate* below).
//
// The rest are genuinely lost. Making them data-driven is the next
// phase of this work, and the codebase is already drifting that way:
// the S19/S20/S22 resume frames are named structs carrying mostly
// data, with a single Build/Then closure left at the bottom.
//
// Until then Snapshot does not pretend. It takes a CENSUS of every
// continuation it could not represent, and a snapshot with a
// non-empty census is not a restore point. Writers keep the last
// snapshot whose census WAS empty, so a restart rewinds the table to
// its most recent continuation-free state rather than resurrecting it
// wrong. That is the "roll back to a clean boundary" simplification,
// stated in full and enforced by a predicate instead of a comment.

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"time"

	"github.com/google/uuid"
)

// SnapshotSchemaVersion is the on-disk shape of GameSnapshot. Bump it
// whenever a change to the snapshot types would make an older file
// decode into a game that is subtly wrong — a renamed field, a
// changed unit, a semantic shift in an existing field. Adding a field
// that zero-values correctly does NOT need a bump.
//
// Restore REFUSES anything it does not recognise rather than guessing.
// See ErrSchemaTooNew / ErrSchemaUnsupported and ADR 0041 for the
// version-skew policy this implements.
const SnapshotSchemaVersion = 1

// minRestorableSchema is the oldest schema Restore still understands.
// Raise it only when carrying a migration forward stops being worth
// the code; every game written at a lower version becomes
// unrestorable the moment you do, so raise it deliberately.
const minRestorableSchema = 1

var (
	// ErrSchemaTooNew means the file was written by a NEWER server
	// than the one reading it — a rollback, or a mixed-version
	// fleet. Never guess: the new binary may have changed what an
	// existing field means.
	ErrSchemaTooNew = errors.New("game: snapshot schema is newer than this server understands")

	// ErrSchemaUnsupported means the file predates minRestorableSchema
	// and no migration path remains.
	ErrSchemaUnsupported = errors.New("game: snapshot schema is too old to restore")

	// ErrSnapshotNotRestorable is returned by RestoreStrict for a
	// snapshot whose continuation census is non-empty — it holds live
	// Go closures that this build cannot rebuild. See ContinuationCensus.
	ErrSnapshotNotRestorable = errors.New("game: snapshot holds continuations that cannot be rebuilt")
)

// ---------------------------------------------------------------
// The snapshot types
//
// Types that are pure data (Event, Turn, Vote, TargetRef, ManaToken,
// LifeChange, Characteristic, CastTally, DamageAssignmentFrame) are
// embedded by value rather than mirrored. Mirroring them would double
// the drift surface for no benefit, and snapshot_drift_test.go proves
// they stay pure by marshalling them.
//
// Types carrying funcs or unexported fields get an explicit mirror,
// because encoding/json cannot see the one and refuses the other.
// ---------------------------------------------------------------

// GameSnapshot is a complete, serialisable *Game.
type GameSnapshot struct {
	Schema int `json:"schema"`

	// TakenAt is when the snapshot was captured, for operator
	// triage ("how stale is the restore point?"). Not game state.
	TakenAt time.Time `json:"takenAt"`

	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	State     State     `json:"state"`

	Seats []playerSnapshot `json:"seats"`

	Battlefield *zoneSnapshot `json:"battlefield"`
	Stack       *zoneSnapshot `json:"stack"`
	Exile       *zoneSnapshot `json:"exile"`

	Turn              Turn      `json:"turn"`
	MulligansOpen     bool      `json:"mulligansOpen"`
	Monarch           uuid.UUID `json:"monarch"`
	Initiative        uuid.UUID `json:"initiative"`
	UndoLimit         int       `json:"undoLimit"`
	StartingSeat      int       `json:"startingSeat"`
	SplitSecondActive bool      `json:"splitSecondActive"`

	// StackMeta is a SLICE, not a map: map iteration order is
	// unspecified and the stack is ordered by Seq anyway. Sorted on
	// capture so two snapshots of the same state are byte-identical.
	StackMeta       []stackItemSnapshot      `json:"stackMeta,omitempty"`
	PendingTriggers []stackItemSnapshot      `json:"pendingTriggers,omitempty"`
	DelayedTriggers []delayedTriggerSnapshot `json:"delayedTriggers,omitempty"`

	LoyaltyActivatedThisTurn map[uuid.UUID]bool        `json:"loyaltyActivatedThisTurn,omitempty"`
	SpellsCastThisTurn       map[uuid.UUID]CastTally   `json:"spellsCastThisTurn,omitempty"`
	LandsPlayedThisTurn      map[uuid.UUID]int         `json:"landsPlayedThisTurn,omitempty"`
	DrawnThisTurn            map[uuid.UUID][]uuid.UUID `json:"drawnThisTurn,omitempty"`
	TurnTally                TurnTally                 `json:"turnTally"`
	DiscardPending           map[uuid.UUID]int         `json:"discardPending,omitempty"`

	// Promises is a slice because its live form is keyed by a
	// STRUCT (PromiseKey), and a JSON object key must be a string.
	Promises []promiseSnapshot `json:"promises,omitempty"`

	Vote *Vote `json:"vote,omitempty"`

	PendingChoices []pendingChoiceSnapshot `json:"pendingChoices,omitempty"`

	Events   []Event `json:"events,omitempty"`
	EventSeq uint64  `json:"eventSeq"`

	// LastKnownBattlefield is CR 603.10 LKI. Empty in steady state —
	// entries live for the duration of one LTB-emitting mutation —
	// but carried so a round-trip is exact rather than nearly exact.
	LastKnownBattlefield map[uuid.UUID]Characteristic `json:"lastKnownBattlefield,omitempty"`

	RNG rngSnapshot `json:"rng"`

	LayerVersion        uint64 `json:"layerVersion"`
	LastResolvedVersion uint64 `json:"lastResolvedVersion"`

	// Continuations records what could not be represented. Empty
	// census == full-fidelity restore point.
	Continuations ContinuationCensus `json:"continuations"`
}

// playerSnapshot mirrors Player. Player is pure data today; it is
// mirrored anyway because its zones are, and because a mirror is
// where a future unexported field gets handled instead of silently
// dropped.
type playerSnapshot struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Seat      int           `json:"seat"`
	Life      int           `json:"life"`
	Poison    int           `json:"poison"`
	Energy    int           `json:"energy"`
	Library   *zoneSnapshot `json:"library"`
	Hand      *zoneSnapshot `json:"hand"`
	Graveyard *zoneSnapshot `json:"graveyard"`
	Command   *zoneSnapshot `json:"command"`
	// CommanderDamage is keyed by commander card instance ID since
	// S25 (#77). A snapshot written before that rekey restores with
	// player-ID keys, which read as damage from commanders that do
	// not exist: harmless (they render nowhere and can never reach
	// 21 again) but not migrated.
	CommanderDamage   map[uuid.UUID]int `json:"commanderDamage,omitempty"`
	LifeHistory       []LifeChange      `json:"lifeHistory,omitempty"`
	Eliminated        bool              `json:"eliminated"`
	HandKept          bool              `json:"handKept"`
	MulligansTaken    int               `json:"mulligansTaken"`
	DeckImported      bool              `json:"deckImported"`
	UndosRemaining    int               `json:"undosRemaining"`
	DiscordID         string            `json:"discordId,omitempty"`
	DiscordAvatarHash string            `json:"discordAvatarHash,omitempty"`
	DisplayName       string            `json:"displayName,omitempty"`
	IsBot             bool              `json:"isBot,omitempty"`
	BotTier           string            `json:"botTier,omitempty"`
	BotDeck           string            `json:"botDeck,omitempty"`
	LosesAtNextSBA    bool              `json:"losesAtNextSba"`
	CommanderCasts    map[uuid.UUID]int `json:"commanderCasts,omitempty"`
	Counters          map[string]int    `json:"counters,omitempty"`
	MaxHandSize       int               `json:"maxHandSize"`
	ManaPool          ManaPool          `json:"manaPool,omitempty"`
}

type zoneSnapshot struct {
	Kind  ZoneKind       `json:"kind"`
	Owner uuid.UUID      `json:"owner"`
	Cards []cardSnapshot `json:"cards,omitempty"`
}

// cardSnapshot mirrors Card minus two things:
//
//   - `effective`, the layer-engine characteristic cache. Derived
//     state; restore drops it and forces a recompute, exactly as
//     RestoreFrom does after an undo.
//   - the closure halves of ManaAbilities / ActivatedAbilities.
//     Those slices are non-empty only for tokens and token copies
//     (everything else looks its abilities up from the catalog by
//     oracle ID at use time). A token COPY carries the copied card's
//     oracle ID, so its abilities are re-derivable; a true token
//     (Treasure, Food, Clue, Blood) has no oracle ID and is counted
//     in the census instead.
type cardSnapshot struct {
	InstanceID               uuid.UUID           `json:"instanceId"`
	Name                     string              `json:"name"`
	ScryfallID               string              `json:"scryfallId,omitempty"`
	OracleID                 string              `json:"oracleId,omitempty"`
	TypeLine                 string              `json:"typeLine,omitempty"`
	Power                    int                 `json:"power"`
	Toughness                int                 `json:"toughness"`
	ManaCost                 string              `json:"manaCost,omitempty"`
	ProducedMana             []string            `json:"producedMana,omitempty"`
	Colors                   []string            `json:"colors,omitempty"`
	ColorIdentity            []string            `json:"colorIdentity,omitempty"`
	StartingLoyalty          int                 `json:"startingLoyalty"`
	Keywords                 []string            `json:"keywords,omitempty"`
	Layout                   string              `json:"layout,omitempty"`
	Faces                    []Face              `json:"faces,omitempty"`
	ActiveFace               int                 `json:"activeFace,omitempty"`
	PrintedSelf              *PrintedValues      `json:"printedSelf,omitempty"`
	NeedsEffect              bool                `json:"needsEffect"`
	Owner                    uuid.UUID           `json:"owner"`
	Controller               uuid.UUID           `json:"controller"`
	Tapped                   bool                `json:"tapped"`
	BattleX                  float64             `json:"battleX"`
	BattleY                  float64             `json:"battleY"`
	Counters                 map[string]int      `json:"counters,omitempty"`
	IsCommander              bool                `json:"isCommander"`
	AttackingTarget          uuid.UUID           `json:"attackingTarget"`
	BlockingTarget           uuid.UUID           `json:"blockingTarget"`
	GoadedBy                 uuid.UUID           `json:"goadedBy"`
	DamageMarked             int                 `json:"damageMarked"`
	FaceDown                 bool                `json:"faceDown"`
	KnownBy                  map[uuid.UUID]bool  `json:"knownBy,omitempty"`
	EnteredBattlefieldAt     int64               `json:"enteredBattlefieldAt"`
	SummonedThisTurn         bool                `json:"summonedThisTurn"`
	MarkedLethalByDeathtouch bool                `json:"markedLethalByDeathtouch"`
	ExilePlay                ExilePlayPermission `json:"exilePlay"`
	AttachedTo               TargetRef           `json:"attachedTo,omitempty"`
	AttachedAt               int64               `json:"attachedAt,omitempty"`
	BaseController           uuid.UUID           `json:"baseController,omitempty"`
	// NamedTribe is the CR 614.12 "as this enters, choose a creature
	// type" answer (S26). Carried rather than rebuilt: the choice was
	// made by a player and nothing in the catalog can re-derive it, so
	// a restore that lost it would leave a Cavern of Souls producing
	// mana for a tribe nobody named.
	NamedTribe        string    `json:"namedTribe,omitempty"`
	StartingDefense   int       `json:"startingDefense,omitempty"`
	ProtectorPlayerID uuid.UUID `json:"protectorPlayerId,omitempty"`

	// ManaAbilityCount / ActivatedAbilityCount record that the card
	// HAD intrinsic ability closures, so restore can tell the
	// difference between "none" and "some it must rebuild", and so
	// the census can name the card.
	ManaAbilityCount      int `json:"manaAbilityCount,omitempty"`
	ActivatedAbilityCount int `json:"activatedAbilityCount,omitempty"`
}

// stackItemSnapshot mirrors StackItem. Effect and targetSpec are both
// func-bearing; see rehydrateStackItem for which ones come back.
type stackItemSnapshot struct {
	ID           uuid.UUID         `json:"id"`
	Kind         StackItemKind     `json:"kind"`
	Controller   uuid.UUID         `json:"controller"`
	Owner        uuid.UUID         `json:"owner"`
	SourceCardID uuid.UUID         `json:"sourceCardId"`
	Label        string            `json:"label,omitempty"`
	Targets      []TargetRef       `json:"targets,omitempty"`
	Modes        []int             `json:"modes,omitempty"`
	XValue       int               `json:"xValue"`
	Distribution map[uuid.UUID]int `json:"distribution,omitempty"`
	HoldPriority bool              `json:"holdPriority"`
	CastFromZone ZoneKind          `json:"castFromZone,omitempty"`
	AltCost      string            `json:"altCost,omitempty"`
	SplitSecond  bool              `json:"splitSecond"`
	IsCopy       bool              `json:"isCopy,omitempty"`
	Seq          uint64            `json:"seq"`
	Ordered      bool              `json:"ordered"`

	// HasEffect / HasTargetSpec record the two closure slots so the
	// census can count what restore had to drop.
	HasEffect     bool `json:"hasEffect,omitempty"`
	HasTargetSpec bool `json:"hasTargetSpec,omitempty"`

	// OracleID is the source card's oracle ID, captured so restore
	// can re-derive a SPELL's target spec from the catalog without
	// having to find the card again (it may have moved zones).
	OracleID string `json:"oracleId,omitempty"`
}

type delayedTriggerSnapshot struct {
	ID                 uuid.UUID   `json:"id"`
	Controller         uuid.UUID   `json:"controller"`
	SourceCardID       uuid.UUID   `json:"sourceCardId"`
	Label              string      `json:"label,omitempty"`
	At                 Step        `json:"at"`
	ControllerTurnOnly bool        `json:"controllerTurnOnly"`
	CreatedTurn        int         `json:"createdTurn"`
	Cards              []uuid.UUID `json:"cards,omitempty"`
	HasEffect          bool        `json:"hasEffect,omitempty"`
}

// pendingChoiceSnapshot mirrors PendingChoice's DATA. Its seven
// continuation frames are the sharpest edge of this whole file: a
// game sitting on a prompt is a game whose next step is a Go closure.
// The data comes back; the continuation does not, which is precisely
// why a snapshot holding one is not a restore point.
type pendingChoiceSnapshot struct {
	ID                   uuid.UUID              `json:"id"`
	Kind                 PendingChoiceKind      `json:"kind"`
	Chooser              uuid.UUID              `json:"chooser"`
	FromPlayer           uuid.UUID              `json:"fromPlayer"`
	Count                int                    `json:"count"`
	Source               uuid.UUID              `json:"source"`
	Reason               string                 `json:"reason,omitempty"`
	ColorOptions         []string               `json:"colorOptions,omitempty"`
	ManaRestrictions     []string               `json:"manaRestrictions,omitempty"`
	ReplacementEffectIDs []ReplacementEffectID  `json:"replacementEffectIds,omitempty"`
	DamageAssignment     *DamageAssignmentFrame `json:"damageAssignment,omitempty"`
	NoLegalTarget        bool                   `json:"noLegalTarget"`
	PickTargetPlayers    []uuid.UUID            `json:"pickTargetPlayers,omitempty"`
	PickTargetCards      []uuid.UUID            `json:"pickTargetCards,omitempty"`
	PickTargetMin        int                    `json:"pickTargetMin"`
	PickTargetMax        int                    `json:"pickTargetMax"`
	SacrificeOptions     []uuid.UUID            `json:"sacrificeOptions,omitempty"`
	CopyOptions          []uuid.UUID            `json:"copyOptions,omitempty"`
	ScryCards            []uuid.UUID            `json:"scryCards,omitempty"`
	TriggerOrderIDs      []uuid.UUID            `json:"triggerOrderIds,omitempty"`
	PayCost              string                 `json:"payCost,omitempty"`
	SearchCards          []uuid.UUID            `json:"searchCards,omitempty"`
	SearchMax            int                    `json:"searchMax"`
	MayCastCard          uuid.UUID              `json:"mayCastCard,omitempty"`
	AcceptLabel          string                 `json:"acceptLabel,omitempty"`
	LifeCost             int                    `json:"lifeCost,omitempty"`
	DeclineLabel         string                 `json:"declineLabel,omitempty"`
	ChooseCards          []uuid.UUID            `json:"chooseCards,omitempty"`
	ChooseMin            int                    `json:"chooseMin,omitempty"`
	ChooseMax            int                    `json:"chooseMax,omitempty"`

	// ResumeFrames names the continuation slots that were populated.
	// Diagnostic only — nothing rebuilds them in this schema.
	ResumeFrames []string `json:"resumeFrames,omitempty"`
}

type promiseSnapshot struct {
	From  uuid.UUID `json:"from"`
	To    uuid.UUID `json:"to"`
	Count int       `json:"count"`
}

// rngSnapshot carries the game's position in its random stream.
//
// Why this is not just a seed: a game that has already shuffled N
// times is N draws into the stream. Restoring from the seed alone
// would re-deal the same numbers the game has already consumed. What
// has to persist is the source's CURRENT state, which is what
// *rand.PCG.MarshalBinary gives us.
//
// Kind discriminates three cases:
//
//	"pcg"      — engine-minted source; State holds its exact position
//	             and the restored game continues the same stream.
//	"external" — a caller supplied its own *rand.Rand (tests do).
//	             math/rand/v2 exposes no way to read a Rand's Source
//	             back out, so the position is unreachable. Recorded
//	             honestly and counted in the census rather than
//	             silently reseeded — a silent reseed is how you get a
//	             restore that quietly re-deals a library.
//	"none"     — no source yet (game not started).
type rngSnapshot struct {
	Kind  string `json:"kind"`
	State []byte `json:"state,omitempty"`
}

const (
	rngKindNone     = "none"
	rngKindPCG      = "pcg"
	rngKindExternal = "external"
)

// ContinuationCensus counts the live Go closures a snapshot could not
// represent. An empty census means the snapshot is a full-fidelity
// restore point; a non-empty one means restoring it would silently
// drop part of the game.
//
// It is a COUNT plus labels rather than a bool because phase 2's
// drain policy and an operator staring at a stuck deploy both want to
// know *what* is holding the game open, and because each counter
// drops to a permanent zero as the corresponding continuation becomes
// data-driven.
type ContinuationCensus struct {
	// StackEffects is stack items (on the stack or queued) whose
	// resolution behaviour is a closure.
	StackEffects int `json:"stackEffects,omitempty"`

	// StackTargetSpecs is items whose CR 608.2b re-check clause is a
	// closure and could not be re-derived from the catalog.
	StackTargetSpecs int `json:"stackTargetSpecs,omitempty"`

	// DelayedTriggerEffects is queued "at the beginning of the next
	// end step" instructions.
	DelayedTriggerEffects int `json:"delayedTriggerEffects,omitempty"`

	// ChoiceResumeFrames is paused prompts holding a continuation.
	ChoiceResumeFrames int `json:"choiceResumeFrames,omitempty"`

	// TurnScopedStatics is floating until-end-of-turn continuous
	// effects (Giant Growth's +3/+3).
	TurnScopedStatics int `json:"turnScopedStatics,omitempty"`

	// TurnScopedReplacements is floating until-end-of-turn
	// replacement effects (Fog).
	TurnScopedReplacements int `json:"turnScopedReplacements,omitempty"`

	// IntrinsicAbilityCards is cards whose ability closures live on
	// the instance rather than in the catalog and that carry no
	// oracle ID to re-derive from — true tokens.
	IntrinsicAbilityCards int `json:"intrinsicAbilityCards,omitempty"`

	// UnpersistableRNG marks a game whose random source belongs to
	// the caller and cannot be read back out.
	UnpersistableRNG bool `json:"unpersistableRng,omitempty"`

	// Labels names the blockers for an operator. Capped so a
	// pathological board cannot bloat the file.
	Labels []string `json:"labels,omitempty"`
}

const censusLabelCap = 24

// Empty reports whether the snapshot is a full-fidelity restore point.
func (c ContinuationCensus) Empty() bool {
	return c.StackEffects == 0 &&
		c.StackTargetSpecs == 0 &&
		c.DelayedTriggerEffects == 0 &&
		c.ChoiceResumeFrames == 0 &&
		c.TurnScopedStatics == 0 &&
		c.TurnScopedReplacements == 0 &&
		c.IntrinsicAbilityCards == 0 &&
		!c.UnpersistableRNG
}

// Total is the number of individual continuations counted.
func (c ContinuationCensus) Total() int {
	n := c.StackEffects + c.StackTargetSpecs + c.DelayedTriggerEffects +
		c.ChoiceResumeFrames + c.TurnScopedStatics +
		c.TurnScopedReplacements + c.IntrinsicAbilityCards
	if c.UnpersistableRNG {
		n++
	}
	return n
}

func (c *ContinuationCensus) note(format string, args ...any) {
	if len(c.Labels) >= censusLabelCap {
		return
	}
	c.Labels = append(c.Labels, fmt.Sprintf(format, args...))
}

// Restorable reports whether this snapshot can be restored at full
// fidelity — i.e. whether it is a usable restore point.
func (s *GameSnapshot) Restorable() bool { return s.Continuations.Empty() }

// ---------------------------------------------------------------
// Capture
// ---------------------------------------------------------------

// CaptureSnapshot captures a serialisable deep copy of the game.
//
// Named CaptureSnapshot rather than Snapshot because Game.Snapshot is
// long since taken by the shallow read-only copy the view layer uses.
// The two are unrelated: that one is a cheap struct copy for
// inspection, this one is the persistence format.
//
// It ALWAYS succeeds: every game has a snapshot, and the question of
// whether that snapshot is faithful is answered by
// GameSnapshot.Restorable rather than by an error. A caller writing
// restore points wants the predicate; a caller writing diagnostics
// wants the file either way.
//
// Concurrency: takes a read lock, like Clone.
func (g *Game) CaptureSnapshot() *GameSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.captureSnapshotLocked()
}

// captureSnapshotLocked is the unlocked variant. Caller must hold g.mu.
func (g *Game) captureSnapshotLocked() *GameSnapshot {
	s := &GameSnapshot{
		Schema:            SnapshotSchemaVersion,
		TakenAt:           time.Now().UTC(),
		ID:                g.ID,
		CreatedAt:         g.CreatedAt,
		State:             g.State,
		Turn:              g.Turn,
		MulligansOpen:     g.MulligansOpen,
		Monarch:           g.Monarch,
		Initiative:        g.Initiative,
		UndoLimit:         g.UndoLimit,
		StartingSeat:      g.StartingSeat,
		SplitSecondActive: g.SplitSecondActive,
		EventSeq:          g.eventSeq,
	}
	cen := &s.Continuations

	s.Battlefield = snapshotZone(g.Battlefield, cen)
	s.Stack = snapshotZone(g.Stack, cen)
	s.Exile = snapshotZone(g.Exile, cen)

	s.Seats = make([]playerSnapshot, len(g.Seats))
	for i, p := range g.Seats {
		s.Seats[i] = snapshotPlayer(p, cen)
	}

	// StackMeta: map → Seq-ordered slice. Sorting is what makes two
	// captures of one state byte-identical, which in turn is what
	// lets an operator diff restore points.
	if len(g.StackMeta) > 0 {
		s.StackMeta = make([]stackItemSnapshot, 0, len(g.StackMeta))
		for _, it := range g.StackMeta {
			s.StackMeta = append(s.StackMeta, snapshotStackItem(g, it, cen))
		}
		sort.Slice(s.StackMeta, func(i, j int) bool {
			if s.StackMeta[i].Seq != s.StackMeta[j].Seq {
				return s.StackMeta[i].Seq < s.StackMeta[j].Seq
			}
			// Seq is unique in practice; the ID tiebreak keeps the
			// order total anyway so the sort never depends on map order.
			return s.StackMeta[i].ID.String() < s.StackMeta[j].ID.String()
		})
	}
	if len(g.PendingTriggers) > 0 {
		s.PendingTriggers = make([]stackItemSnapshot, len(g.PendingTriggers))
		for i, it := range g.PendingTriggers {
			s.PendingTriggers[i] = snapshotStackItem(g, it, cen)
		}
	}
	if len(g.DelayedTriggers) > 0 {
		s.DelayedTriggers = make([]delayedTriggerSnapshot, len(g.DelayedTriggers))
		for i, d := range g.DelayedTriggers {
			s.DelayedTriggers[i] = snapshotDelayedTrigger(d, cen)
		}
	}

	s.LoyaltyActivatedThisTurn = copyBoolMap(g.LoyaltyActivatedThisTurn)
	s.SpellsCastThisTurn = copyTallyMap(g.SpellsCastThisTurn)
	s.LandsPlayedThisTurn = copyIntMap(g.LandsPlayedThisTurn)
	s.DrawnThisTurn = copyUUIDListMap(g.DrawnThisTurn)
	s.TurnTally = cloneTurnTally(g.TurnTally)
	s.DiscardPending = copyIntMap(g.DiscardPending)

	if len(g.Promises) > 0 {
		s.Promises = make([]promiseSnapshot, 0, len(g.Promises))
		for k, v := range g.Promises {
			s.Promises = append(s.Promises, promiseSnapshot{From: k.From, To: k.To, Count: v})
		}
		sort.Slice(s.Promises, func(i, j int) bool {
			if s.Promises[i].From != s.Promises[j].From {
				return s.Promises[i].From.String() < s.Promises[j].From.String()
			}
			return s.Promises[i].To.String() < s.Promises[j].To.String()
		})
	}

	if g.Vote != nil {
		s.Vote = cloneVote(g.Vote)
	}

	if len(g.PendingChoices) > 0 {
		s.PendingChoices = make([]pendingChoiceSnapshot, 0, len(g.PendingChoices))
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			s.PendingChoices = append(s.PendingChoices, snapshotPendingChoice(c, cen))
		}
	}

	if len(g.Events) > 0 {
		s.Events = make([]Event, len(g.Events))
		copy(s.Events, g.Events)
	}

	if len(g.lastKnownBattlefield) > 0 {
		s.LastKnownBattlefield = make(map[uuid.UUID]Characteristic, len(g.lastKnownBattlefield))
		for k, v := range g.lastKnownBattlefield {
			s.LastKnownBattlefield[k] = v
		}
	}

	// Turn-scoped registries: entirely closure-bearing, so only the
	// census and the labels survive. Dropping a Fog silently would be
	// worse than refusing the restore point, which is what this does.
	for _, st := range g.TurnScopedStatics {
		cen.TurnScopedStatics++
		cen.note("turn-scoped static: %s", labelOr(st.Label, st.Source.Name))
	}
	for _, re := range g.TurnScopedReplacements {
		cen.TurnScopedReplacements++
		cen.note("turn-scoped replacement: %s", labelOr(re.Label, "unnamed"))
	}
	// BuiltinReplacements and Listeners are deliberately NOT counted:
	// both are process-lifetime singletons installed by NewGame, so
	// the new binary rebuilds them itself. See restoreGame.

	s.RNG = snapshotRNG(g, cen)
	s.LayerVersion = g.layerVersion.Load()
	s.LastResolvedVersion = g.lastResolvedVersion.Load()
	return s
}

func snapshotRNG(g *Game, cen *ContinuationCensus) rngSnapshot {
	switch {
	case g.rngState != nil:
		b, err := g.rngState.MarshalBinary()
		if err != nil {
			// *rand.PCG.MarshalBinary does not fail in practice; if
			// it ever does, refuse to claim we captured the stream.
			cen.UnpersistableRNG = true
			cen.note("rng: PCG marshal failed: %v", err)
			return rngSnapshot{Kind: rngKindExternal}
		}
		return rngSnapshot{Kind: rngKindPCG, State: b}
	case g.rng != nil:
		// Caller-supplied source. Its position is unreachable.
		cen.UnpersistableRNG = true
		cen.note("rng: caller-supplied source, stream position unreadable")
		return rngSnapshot{Kind: rngKindExternal}
	default:
		return rngSnapshot{Kind: rngKindNone}
	}
}

func snapshotZone(z *Zone, cen *ContinuationCensus) *zoneSnapshot {
	if z == nil {
		return nil
	}
	out := &zoneSnapshot{Kind: z.Kind, Owner: z.Owner}
	if len(z.Cards) > 0 {
		out.Cards = make([]cardSnapshot, len(z.Cards))
		for i, c := range z.Cards {
			out.Cards[i] = snapshotCard(c, cen)
		}
	}
	return out
}

func snapshotCard(c Card, cen *ContinuationCensus) cardSnapshot {
	out := cardSnapshot{
		InstanceID:               c.InstanceID,
		Name:                     c.Name,
		ScryfallID:               c.ScryfallID,
		OracleID:                 c.OracleID,
		TypeLine:                 c.TypeLine,
		Power:                    c.Power,
		Toughness:                c.Toughness,
		ManaCost:                 c.ManaCost,
		ProducedMana:             copyStrings(c.ProducedMana),
		Colors:                   copyStrings(c.Colors),
		ColorIdentity:            copyStrings(c.ColorIdentity),
		StartingLoyalty:          c.StartingLoyalty,
		Keywords:                 copyStrings(c.Keywords),
		Layout:                   c.Layout,
		Faces:                    copyFaces(c.Faces),
		ActiveFace:               c.ActiveFace,
		PrintedSelf:              copyPrintedValues(c.PrintedSelf),
		NeedsEffect:              c.NeedsEffect,
		Owner:                    c.Owner,
		Controller:               c.Controller,
		Tapped:                   c.Tapped,
		BattleX:                  c.BattleX,
		BattleY:                  c.BattleY,
		Counters:                 copyStringIntMap(c.Counters),
		IsCommander:              c.IsCommander,
		AttackingTarget:          c.AttackingTarget,
		BlockingTarget:           c.BlockingTarget,
		GoadedBy:                 c.GoadedBy,
		DamageMarked:             c.DamageMarked,
		FaceDown:                 c.FaceDown,
		KnownBy:                  copyBoolMap(c.KnownBy),
		EnteredBattlefieldAt:     c.EnteredBattlefieldAt,
		SummonedThisTurn:         c.SummonedThisTurn,
		MarkedLethalByDeathtouch: c.MarkedLethalByDeathtouch,
		ExilePlay:                c.ExilePlay,
		AttachedTo:               c.AttachedTo,
		AttachedAt:               c.AttachedAt,
		BaseController:           c.BaseController,
		NamedTribe:               c.NamedTribe,
		StartingDefense:          c.StartingDefense,
		ProtectorPlayerID:        c.ProtectorPlayerID,
		ManaAbilityCount:         len(c.ManaAbilities),
		ActivatedAbilityCount:    len(c.ActivatedAbilities),
	}
	// Intrinsic ability closures. Re-derivable when the card carries
	// an oracle ID (a token COPY does — CR 707.2 copies the oracle
	// identity too), lost when it does not (Treasure, Food, Clue,
	// Blood are their ability and have no printing to look up).
	if (len(c.ManaAbilities) > 0 || len(c.ActivatedAbilities) > 0) && c.OracleID == "" {
		cen.IntrinsicAbilityCards++
		cen.note("intrinsic abilities with no oracle id: %s", labelOr(c.Name, c.InstanceID.String()))
	}
	return out
}

func snapshotPlayer(p *Player, cen *ContinuationCensus) playerSnapshot {
	out := playerSnapshot{
		ID:                p.ID,
		Name:              p.Name,
		Seat:              p.Seat,
		Life:              p.Life,
		Poison:            p.Poison,
		Energy:            p.Energy,
		Library:           snapshotZone(p.Library, cen),
		Hand:              snapshotZone(p.Hand, cen),
		Graveyard:         snapshotZone(p.Graveyard, cen),
		Command:           snapshotZone(p.Command, cen),
		CommanderDamage:   copyIntMap(p.CommanderDamage),
		Eliminated:        p.Eliminated,
		HandKept:          p.HandKept,
		MulligansTaken:    p.MulligansTaken,
		DeckImported:      p.DeckImported,
		UndosRemaining:    p.UndosRemaining,
		DiscordID:         p.DiscordID,
		DiscordAvatarHash: p.DiscordAvatarHash,
		DisplayName:       p.DisplayName,
		IsBot:             p.IsBot,
		BotTier:           p.BotTier,
		BotDeck:           p.BotDeck,
		LosesAtNextSBA:    p.LosesAtNextSBA,
		CommanderCasts:    copyIntMap(p.CommanderCasts),
		Counters:          copyStringIntMap(p.Counters),
		MaxHandSize:       p.MaxHandSize,
	}
	if len(p.LifeHistory) > 0 {
		out.LifeHistory = make([]LifeChange, len(p.LifeHistory))
		copy(out.LifeHistory, p.LifeHistory)
	}
	if len(p.ManaPool) > 0 {
		out.ManaPool = make(ManaPool, len(p.ManaPool))
		for i, t := range p.ManaPool {
			cloned := t
			cloned.Restrictions = copyStrings(t.Restrictions)
			out.ManaPool[i] = cloned
		}
	}
	return out
}

func snapshotStackItem(g *Game, s *StackItem, cen *ContinuationCensus) stackItemSnapshot {
	if s == nil {
		return stackItemSnapshot{}
	}
	out := stackItemSnapshot{
		ID:            s.ID,
		Kind:          s.Kind,
		Controller:    s.Controller,
		Owner:         s.Owner,
		SourceCardID:  s.SourceCardID,
		Label:         s.Label,
		Targets:       copyTargetRefs(s.Targets),
		Modes:         copyInts(s.Modes),
		XValue:        s.XValue,
		Distribution:  copyIntMap(s.Distribution),
		HoldPriority:  s.HoldPriority,
		CastFromZone:  s.CastFromZone,
		AltCost:       s.AltCost,
		SplitSecond:   s.SplitSecond,
		IsCopy:        s.IsCopy,
		Seq:           s.Seq,
		Ordered:       s.Ordered,
		HasEffect:     s.Effect != nil,
		HasTargetSpec: s.targetSpec != nil,
		OracleID:      oracleIDOfLocked(g, s.SourceCardID),
	}
	if s.Effect != nil {
		cen.StackEffects++
		cen.note("stack effect: %s", labelOr(s.Label, string(s.Kind)))
	}
	// A SPELL's target spec was looked up from the catalog by oracle
	// ID when it was cast, so the new binary can look it up again —
	// no census entry. An ABILITY's spec came off the ability
	// declaration, which is reached by an index the item does not
	// record, so it cannot be recovered here.
	if s.targetSpec != nil && !spellSpecRederivable(out) {
		cen.StackTargetSpecs++
		cen.note("stack target spec: %s", labelOr(s.Label, string(s.Kind)))
	}
	return out
}

// spellSpecRederivable reports whether restore can rebuild this
// item's targetSpec from the card catalog.
func spellSpecRederivable(s stackItemSnapshot) bool {
	return s.Kind == StackItemSpell && s.OracleID != "" && CatalogTargetSpec != nil
}

func oracleIDOfLocked(g *Game, cardID uuid.UUID) string {
	if cardID == uuid.Nil {
		return ""
	}
	if c := g.findCardByIDLocked(cardID); c != nil {
		return c.OracleID
	}
	return ""
}

func snapshotDelayedTrigger(d *DelayedTrigger, cen *ContinuationCensus) delayedTriggerSnapshot {
	if d == nil {
		return delayedTriggerSnapshot{}
	}
	out := delayedTriggerSnapshot{
		ID:                 d.ID,
		Controller:         d.Controller,
		SourceCardID:       d.SourceCardID,
		Label:              d.Label,
		At:                 d.At,
		ControllerTurnOnly: d.ControllerTurnOnly,
		CreatedTurn:        d.CreatedTurn,
		HasEffect:          d.Effect != nil,
	}
	if len(d.Cards) > 0 {
		out.Cards = append([]uuid.UUID(nil), d.Cards...)
	}
	if d.Effect != nil {
		cen.DelayedTriggerEffects++
		cen.note("delayed trigger: %s", labelOr(d.Label, d.ID.String()))
	}
	return out
}

func snapshotPendingChoice(c *PendingChoice, cen *ContinuationCensus) pendingChoiceSnapshot {
	out := pendingChoiceSnapshot{
		ID:                   c.ID,
		Kind:                 c.Kind,
		Chooser:              c.Chooser,
		FromPlayer:           c.FromPlayer,
		Count:                c.Count,
		Source:               c.Source,
		Reason:               c.Reason,
		ColorOptions:         copyStrings(c.ColorOptions),
		ManaRestrictions:     copyStrings(c.ManaRestrictions),
		ReplacementEffectIDs: copyReplacementEffectIDs(c.ReplacementEffectIDs),
		NoLegalTarget:        c.NoLegalTarget,
		PickTargetPlayers:    copyUUIDs(c.PickTargetPlayers),
		PickTargetCards:      copyUUIDs(c.PickTargetCards),
		PickTargetMin:        c.PickTargetMin,
		PickTargetMax:        c.PickTargetMax,
		SacrificeOptions:     copyUUIDs(c.SacrificeOptions),
		CopyOptions:          copyUUIDs(c.CopyOptions),
		ScryCards:            copyUUIDs(c.ScryCards),
		TriggerOrderIDs:      copyUUIDs(c.TriggerOrderIDs),
		PayCost:              c.PayCost,
		SearchCards:          copyUUIDs(c.SearchCards),
		SearchMax:            c.SearchMax,
		MayCastCard:          c.MayCastCard,
		AcceptLabel:          c.AcceptLabel,
		DeclineLabel:         c.DeclineLabel,
		LifeCost:             c.LifeCost,
		ChooseCards:          copyUUIDs(c.ChooseCards),
		ChooseMin:            c.ChooseMin,
		ChooseMax:            c.ChooseMax,
	}
	if c.DamageAssignment != nil {
		// Pure data (see the type), so a value copy with its own
		// slice is a complete copy.
		da := *c.DamageAssignment
		da.BlockerIDs = copyUUIDs(c.DamageAssignment.BlockerIDs)
		out.DamageAssignment = &da
	}
	// The continuation slots, one by one. Each is a paused effect.
	for name, present := range map[string]bool{
		"replacementResume": c.replacementResume != nil,
		"pickTargetResume":  c.pickTargetResume != nil,
		"copySpellResume":   c.copySpellResume != nil,
		"triggerResume":     c.triggerResume != nil,
		"payUnlessResume":   c.payUnlessResume != nil,
		"mayCastResume":     c.mayCastResume != nil,
		"searchResume":      c.searchResume != nil,
		"scryResume":        c.scryResume != nil,
		// The two chained-choice frames. A chain link is a
		// continuation like any other, and an entry missing here
		// would let the server write a restore point that silently
		// drops the rest of the card. See chained_choice.go.
		"confirmResume":     c.confirmResume != nil,
		"chooseCardsResume": c.chooseCardsResume != nil,
	} {
		if present {
			out.ResumeFrames = append(out.ResumeFrames, name)
		}
	}
	sort.Strings(out.ResumeFrames) // map iteration order is not stable
	for _, name := range out.ResumeFrames {
		cen.ChoiceResumeFrames++
		cen.note("pending choice %s holds %s", c.Kind, name)
	}
	return out
}

// ---------------------------------------------------------------
// Restore
// ---------------------------------------------------------------

// Restore rebuilds a *Game from a snapshot.
//
// It performs the schema check and rebuilds everything the snapshot
// carries plus everything the new binary can re-derive (listeners,
// built-in replacements, catalog-owned target specs). It does NOT
// refuse a snapshot whose census is non-empty — that judgement
// belongs to the caller, and a diagnostic tool wants the game even
// when a production restore would not. Use RestoreStrict for the
// restore-point path.
func (s *GameSnapshot) Restore() (*Game, error) {
	if err := s.checkSchema(); err != nil {
		return nil, err
	}
	return s.restoreGame(), nil
}

// RestoreStrict is Restore plus the full-fidelity requirement: a
// snapshot holding continuations this build cannot rebuild is
// rejected with ErrSnapshotNotRestorable rather than restored into a
// game that is quietly missing a Fog.
//
// This is the call the boot-time restore path makes.
func (s *GameSnapshot) RestoreStrict() (*Game, error) {
	if err := s.checkSchema(); err != nil {
		return nil, err
	}
	if !s.Restorable() {
		return nil, fmt.Errorf("%w: %d continuation(s): %v",
			ErrSnapshotNotRestorable, s.Continuations.Total(), s.Continuations.Labels)
	}
	return s.restoreGame(), nil
}

// checkSchema implements the version-skew policy: understand it, or
// refuse it. Never guess at a shape you do not recognise.
func (s *GameSnapshot) checkSchema() error {
	switch {
	case s.Schema > SnapshotSchemaVersion:
		return fmt.Errorf("%w: file is v%d, this server reads up to v%d",
			ErrSchemaTooNew, s.Schema, SnapshotSchemaVersion)
	case s.Schema < minRestorableSchema:
		return fmt.Errorf("%w: file is v%d, this server reads from v%d",
			ErrSchemaUnsupported, s.Schema, minRestorableSchema)
	}
	return nil
}

func (s *GameSnapshot) restoreGame() *Game {
	// NewGame, not a bare &Game{}: it installs the layer-version
	// listener, the trigger harvester, and the commander-zone
	// built-in replacement. Those are process-lifetime singletons
	// owned by the BINARY, not the game — which is exactly why they
	// need no serialisation and must not be restored from the file.
	// A new binary with a new listener gets it for free here.
	g := NewGame()

	g.ID = s.ID
	g.CreatedAt = s.CreatedAt
	g.State = s.State
	g.Turn = s.Turn
	g.MulligansOpen = s.MulligansOpen
	g.Monarch = s.Monarch
	g.Initiative = s.Initiative
	g.UndoLimit = s.UndoLimit
	g.StartingSeat = s.StartingSeat
	g.SplitSecondActive = s.SplitSecondActive
	g.eventSeq = s.EventSeq

	g.Battlefield = restoreZone(s.Battlefield, ZoneBattlefield)
	g.Stack = restoreZone(s.Stack, ZoneStack)
	g.Exile = restoreZone(s.Exile, ZoneExile)

	g.Seats = make([]*Player, len(s.Seats))
	for i := range s.Seats {
		g.Seats[i] = restorePlayer(&s.Seats[i])
	}

	if len(s.StackMeta) > 0 {
		g.StackMeta = make(map[uuid.UUID]*StackItem, len(s.StackMeta))
		for i := range s.StackMeta {
			it := restoreStackItem(&s.StackMeta[i])
			g.StackMeta[it.ID] = it
		}
	}
	if len(s.PendingTriggers) > 0 {
		g.PendingTriggers = make([]*StackItem, len(s.PendingTriggers))
		for i := range s.PendingTriggers {
			g.PendingTriggers[i] = restoreStackItem(&s.PendingTriggers[i])
		}
	}
	if len(s.DelayedTriggers) > 0 {
		g.DelayedTriggers = make([]*DelayedTrigger, len(s.DelayedTriggers))
		for i := range s.DelayedTriggers {
			g.DelayedTriggers[i] = restoreDelayedTrigger(&s.DelayedTriggers[i])
		}
	}

	g.LoyaltyActivatedThisTurn = copyBoolMap(s.LoyaltyActivatedThisTurn)
	g.SpellsCastThisTurn = copyTallyMap(s.SpellsCastThisTurn)
	g.TurnTally = cloneTurnTally(s.TurnTally)
	g.LandsPlayedThisTurn = copyIntMap(s.LandsPlayedThisTurn)
	g.DrawnThisTurn = copyUUIDListMap(s.DrawnThisTurn)
	g.DiscardPending = copyIntMap(s.DiscardPending)

	if len(s.Promises) > 0 {
		g.Promises = make(map[PromiseKey]int, len(s.Promises))
		for _, p := range s.Promises {
			g.Promises[PromiseKey{From: p.From, To: p.To}] = p.Count
		}
	}

	if s.Vote != nil {
		g.Vote = cloneVote(s.Vote)
	}

	if len(s.PendingChoices) > 0 {
		g.PendingChoices = make([]*PendingChoice, len(s.PendingChoices))
		for i := range s.PendingChoices {
			g.PendingChoices[i] = restorePendingChoice(&s.PendingChoices[i])
		}
	}

	if len(s.Events) > 0 {
		g.Events = make([]Event, len(s.Events))
		copy(g.Events, s.Events)
	}

	if len(s.LastKnownBattlefield) > 0 {
		g.lastKnownBattlefield = make(map[uuid.UUID]Characteristic, len(s.LastKnownBattlefield))
		for k, v := range s.LastKnownBattlefield {
			g.lastKnownBattlefield[k] = v
		}
	}

	restoreRNG(g, s.RNG)

	// Layer-engine counters, handled exactly as RestoreFrom does
	// after an undo and for the same reason: every restored Card
	// dropped its `effective` cache, so the next read must recompute
	// rather than serve a stale one. Bumping layerVersion past
	// lastResolvedVersion is what forces that.
	g.lastResolvedVersion.Store(s.LastResolvedVersion)
	g.layerVersion.Store(s.LayerVersion + 1)
	return g
}

func restoreRNG(g *Game, s rngSnapshot) {
	switch s.Kind {
	case rngKindPCG:
		pcg := &rand.PCG{}
		if err := pcg.UnmarshalBinary(s.State); err != nil {
			// A corrupt stream is not a reason to hand the table a
			// game that cannot shuffle. Mint a fresh source: the
			// restored game is still fair, it just does not continue
			// the exact stream. The census already told the operator
			// this file was a restore point, so this path means the
			// bytes were damaged, not that the design was wrong.
			pcg = newCryptoSeededPCG()
		}
		g.rngState = pcg
		g.rng = rand.New(pcg)
	case rngKindExternal:
		// The original source is unreachable. Give the restored game
		// a real, unpredictable source rather than leaving it nil —
		// a nil rng silently falls back to the process-global source,
		// which works but is the thing we moved away from.
		g.rngState = newCryptoSeededPCG()
		g.rng = rand.New(g.rngState)
	default:
		// Game never started; Start will mint one.
	}
}

// newCryptoSeededPCG mints a *rand.PCG seeded from crypto/rand.
//
// PCG rather than the default global source because PCG implements
// encoding.BinaryMarshaler — it can be written to a snapshot and
// resumed. Seeded from crypto/rand so the stream is unguessable to a
// player who knows the game ID or the wall clock; the seed never
// leaves the server.
func newCryptoSeededPCG() *rand.PCG {
	var b [16]byte
	if _, err := crand.Read(b[:]); err != nil {
		// crypto/rand.Read does not fail on any platform we run on
		// (and panics internally on most). Fall back to a clock seed
		// rather than returning a zero-seeded — i.e. identical for
		// every game — source.
		now := uint64(time.Now().UnixNano())
		return rand.NewPCG(now, now^0x9e3779b97f4a7c15)
	}
	return rand.NewPCG(
		binary.LittleEndian.Uint64(b[0:8]),
		binary.LittleEndian.Uint64(b[8:16]),
	)
}

func restoreZone(z *zoneSnapshot, fallback ZoneKind) *Zone {
	if z == nil {
		return newZone(fallback, uuid.Nil)
	}
	out := &Zone{Kind: z.Kind, Owner: z.Owner}
	if len(z.Cards) > 0 {
		out.Cards = make([]Card, len(z.Cards))
		for i := range z.Cards {
			out.Cards[i] = restoreCard(&z.Cards[i])
		}
	}
	return out
}

func restoreCard(c *cardSnapshot) Card {
	out := Card{
		InstanceID:               c.InstanceID,
		Name:                     c.Name,
		ScryfallID:               c.ScryfallID,
		OracleID:                 c.OracleID,
		TypeLine:                 c.TypeLine,
		Power:                    c.Power,
		Toughness:                c.Toughness,
		ManaCost:                 c.ManaCost,
		ProducedMana:             copyStrings(c.ProducedMana),
		Colors:                   copyStrings(c.Colors),
		ColorIdentity:            copyStrings(c.ColorIdentity),
		StartingLoyalty:          c.StartingLoyalty,
		Keywords:                 copyStrings(c.Keywords),
		Layout:                   c.Layout,
		Faces:                    copyFaces(c.Faces),
		ActiveFace:               c.ActiveFace,
		PrintedSelf:              copyPrintedValues(c.PrintedSelf),
		NeedsEffect:              c.NeedsEffect,
		Owner:                    c.Owner,
		Controller:               c.Controller,
		Tapped:                   c.Tapped,
		BattleX:                  c.BattleX,
		BattleY:                  c.BattleY,
		Counters:                 copyStringIntMap(c.Counters),
		IsCommander:              c.IsCommander,
		AttackingTarget:          c.AttackingTarget,
		BlockingTarget:           c.BlockingTarget,
		GoadedBy:                 c.GoadedBy,
		DamageMarked:             c.DamageMarked,
		FaceDown:                 c.FaceDown,
		KnownBy:                  copyBoolMap(c.KnownBy),
		EnteredBattlefieldAt:     c.EnteredBattlefieldAt,
		SummonedThisTurn:         c.SummonedThisTurn,
		MarkedLethalByDeathtouch: c.MarkedLethalByDeathtouch,
		ExilePlay:                c.ExilePlay,
		AttachedTo:               c.AttachedTo,
		AttachedAt:               c.AttachedAt,
		BaseController:           c.BaseController,
		NamedTribe:               c.NamedTribe,
		StartingDefense:          c.StartingDefense,
		ProtectorPlayerID:        c.ProtectorPlayerID,
	}
	// Re-derive intrinsic abilities from the catalog. This is the
	// half of the closure problem that DOES have an answer: the
	// closures came out of the catalog by oracle ID, and the catalog
	// is code the new binary already has. A token copy carries the
	// copied card's oracle ID (CR 707.2), so it comes back whole.
	//
	// A card with no oracle ID cannot be looked up; snapshotCard
	// already counted it in the census, so a strict restore never
	// reaches this line with abilities to rebuild.
	if c.OracleID != "" {
		if c.ManaAbilityCount > 0 && CatalogManaAbilities != nil {
			out.ManaAbilities = CatalogManaAbilities(c.OracleID)
		}
		if c.ActivatedAbilityCount > 0 && CatalogActivatedAbilities != nil {
			out.ActivatedAbilities = CatalogActivatedAbilities(c.OracleID)
		}
	}
	// `effective` is intentionally left nil: it is a derived cache,
	// and restoreGame bumps layerVersion so the next read recomputes.
	return out
}

func restorePlayer(p *playerSnapshot) *Player {
	out := &Player{
		ID:                p.ID,
		Name:              p.Name,
		Seat:              p.Seat,
		Life:              p.Life,
		Poison:            p.Poison,
		Energy:            p.Energy,
		Library:           restoreZone(p.Library, ZoneLibrary),
		Hand:              restoreZone(p.Hand, ZoneHand),
		Graveyard:         restoreZone(p.Graveyard, ZoneGraveyard),
		Command:           restoreZone(p.Command, ZoneCommand),
		Eliminated:        p.Eliminated,
		HandKept:          p.HandKept,
		MulligansTaken:    p.MulligansTaken,
		DeckImported:      p.DeckImported,
		UndosRemaining:    p.UndosRemaining,
		DiscordID:         p.DiscordID,
		DiscordAvatarHash: p.DiscordAvatarHash,
		DisplayName:       p.DisplayName,
		IsBot:             p.IsBot,
		BotTier:           p.BotTier,
		BotDeck:           p.BotDeck,
		LosesAtNextSBA:    p.LosesAtNextSBA,
		Counters:          copyStringIntMap(p.Counters),
		MaxHandSize:       p.MaxHandSize,
	}
	// clonePlayer guarantees these two are non-nil even when empty;
	// match it so a restored game and a cloned one are the same shape.
	out.CommanderDamage = copyIntMap(p.CommanderDamage)
	if out.CommanderDamage == nil {
		out.CommanderDamage = make(map[uuid.UUID]int)
	}
	out.CommanderCasts = copyIntMap(p.CommanderCasts)
	if out.CommanderCasts == nil {
		out.CommanderCasts = make(map[uuid.UUID]int)
	}
	if len(p.LifeHistory) > 0 {
		out.LifeHistory = make([]LifeChange, len(p.LifeHistory))
		copy(out.LifeHistory, p.LifeHistory)
	}
	if len(p.ManaPool) > 0 {
		out.ManaPool = make(ManaPool, len(p.ManaPool))
		for i, t := range p.ManaPool {
			cloned := t
			cloned.Restrictions = copyStrings(t.Restrictions)
			out.ManaPool[i] = cloned
		}
	}
	return out
}

func restoreStackItem(s *stackItemSnapshot) *StackItem {
	out := &StackItem{
		ID:           s.ID,
		Kind:         s.Kind,
		Controller:   s.Controller,
		Owner:        s.Owner,
		SourceCardID: s.SourceCardID,
		Label:        s.Label,
		Targets:      copyTargetRefs(s.Targets),
		Modes:        copyInts(s.Modes),
		XValue:       s.XValue,
		Distribution: copyIntMap(s.Distribution),
		HoldPriority: s.HoldPriority,
		CastFromZone: s.CastFromZone,
		AltCost:      s.AltCost,
		SplitSecond:  s.SplitSecond,
		IsCopy:       s.IsCopy,
		Seq:          s.Seq,
		Ordered:      s.Ordered,
		// Effect stays nil. A SPELL does not need one — resolution
		// dispatches through EffectResolver by oracle ID — but an
		// ability does, which is why a stack item with an Effect is
		// counted in the census and keeps the snapshot from being a
		// restore point.
	}
	if s.HasTargetSpec && spellSpecRederivable(*s) {
		out.targetSpec = CatalogTargetSpec(s.OracleID)
	}
	return out
}

func restoreDelayedTrigger(d *delayedTriggerSnapshot) *DelayedTrigger {
	out := &DelayedTrigger{
		ID:                 d.ID,
		Controller:         d.Controller,
		SourceCardID:       d.SourceCardID,
		Label:              d.Label,
		At:                 d.At,
		ControllerTurnOnly: d.ControllerTurnOnly,
		CreatedTurn:        d.CreatedTurn,
		// Effect stays nil; see the census.
	}
	if len(d.Cards) > 0 {
		out.Cards = append([]uuid.UUID(nil), d.Cards...)
	}
	return out
}

func restorePendingChoice(c *pendingChoiceSnapshot) *PendingChoice {
	out := &PendingChoice{
		ID:                   c.ID,
		Kind:                 c.Kind,
		Chooser:              c.Chooser,
		FromPlayer:           c.FromPlayer,
		Count:                c.Count,
		Source:               c.Source,
		Reason:               c.Reason,
		ColorOptions:         copyStrings(c.ColorOptions),
		ManaRestrictions:     copyStrings(c.ManaRestrictions),
		ReplacementEffectIDs: copyReplacementEffectIDs(c.ReplacementEffectIDs),
		NoLegalTarget:        c.NoLegalTarget,
		PickTargetPlayers:    copyUUIDs(c.PickTargetPlayers),
		PickTargetCards:      copyUUIDs(c.PickTargetCards),
		PickTargetMin:        c.PickTargetMin,
		PickTargetMax:        c.PickTargetMax,
		SacrificeOptions:     copyUUIDs(c.SacrificeOptions),
		CopyOptions:          copyUUIDs(c.CopyOptions),
		ScryCards:            copyUUIDs(c.ScryCards),
		TriggerOrderIDs:      copyUUIDs(c.TriggerOrderIDs),
		PayCost:              c.PayCost,
		SearchCards:          copyUUIDs(c.SearchCards),
		SearchMax:            c.SearchMax,
		MayCastCard:          c.MayCastCard,
		AcceptLabel:          c.AcceptLabel,
		DeclineLabel:         c.DeclineLabel,
		LifeCost:             c.LifeCost,
		ChooseCards:          copyUUIDs(c.ChooseCards),
		ChooseMin:            c.ChooseMin,
		ChooseMax:            c.ChooseMax,
		// Every resume frame stays nil. This is the phase-1 line in
		// the sand, and the census is how it is enforced rather than
		// hoped for.
	}
	if c.DamageAssignment != nil {
		da := *c.DamageAssignment
		da.BlockerIDs = copyUUIDs(c.DamageAssignment.BlockerIDs)
		out.DamageAssignment = &da
	}
	return out
}

// ---------------------------------------------------------------
// small helpers
// ---------------------------------------------------------------

func labelOr(s, fallback string) string {
	if s != "" {
		return s
	}
	return fallback
}

func copyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	return append([]string(nil), in...)
}

func copyInts(in []int) []int {
	if len(in) == 0 {
		return nil
	}
	return append([]int(nil), in...)
}

func copyUUIDs(in []uuid.UUID) []uuid.UUID {
	if len(in) == 0 {
		return nil
	}
	return append([]uuid.UUID(nil), in...)
}

func copyTargetRefs(in []TargetRef) []TargetRef {
	if len(in) == 0 {
		return nil
	}
	return append([]TargetRef(nil), in...)
}

func copyReplacementEffectIDs(in []ReplacementEffectID) []ReplacementEffectID {
	if len(in) == 0 {
		return nil
	}
	return append([]ReplacementEffectID(nil), in...)
}

func copyIntMap(in map[uuid.UUID]int) map[uuid.UUID]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// copyUUIDListMap deep-copies a per-player list map (DrawnThisTurn),
// giving each player's slice its own backing array so a capture cannot
// be grown by the game it was taken from.
func copyUUIDListMap(in map[uuid.UUID][]uuid.UUID) map[uuid.UUID][]uuid.UUID {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID][]uuid.UUID, len(in))
	for k, v := range in {
		out[k] = append([]uuid.UUID(nil), v...)
	}
	return out
}

func copyBoolMap(in map[uuid.UUID]bool) map[uuid.UUID]bool {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyStringIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyTallyMap(in map[uuid.UUID]CastTally) map[uuid.UUID]CastTally {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]CastTally, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// copyFaces deep-copies a face list. Face is pure printed data
// (strings and ints), so a value copy of each element is enough.
func copyFaces(in []Face) []Face {
	if len(in) == 0 {
		return nil
	}
	return append([]Face(nil), in...)
}

// copyPrintedValues deep-copies a card's stashed pre-copy printed
// values (CR 707 — see copy.go). nil in, nil out: the overwhelming
// majority of cards are not copies of anything.
//
// The value itself is pure data, which is why PrintedValues
// deliberately excludes the closure-bearing ability slices: a mirror
// that could not marshal would fail on the write path, in
// production, for one unlucky game.
func copyPrintedValues(in *PrintedValues) *PrintedValues {
	if in == nil {
		return nil
	}
	out := in.Clone()
	return &out
}
