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
//   - (RETIRED by ADR 0041 phase 3 tier 3b: Gingerbrute's and Fog's
//     block rules and replacements are both ScopedEffect records now)
//   - Card.ManaAbilities / ActivatedAbilities, when an ability
//     closure was stamped onto the INSTANCE and the catalog cannot
//     hand it back. #521 emptied the large and ordinary case of
//     this: a token template's abilities are registered under a
//     synthetic token key (game/token_key.go) and re-derived like a
//     printed card's, so a Treasure on the battlefield no longer
//     holds every restore point in the game open.
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
	"encoding/json"
	"errors"
	"fmt"
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
// Neither does adding a field whose zero value is wrong for older
// files, when restore can tell the field is ABSENT (not merely false)
// and recompute it from printed data the new binary has. That is a
// backfill, and it has to say where its answer comes from and what it
// cannot know. Card.VariableToughness is the one such field; see
// snapshot_backfill.go.
//
// The OTHER direction is also a reason to bump, and it is why v2
// exists (#623, ADR 0064 Decision 7): emblems are additive and
// restore correctly from a v1 file, but a v1 BINARY reading a v2 file
// would restore a game with the emblems silently missing. There is no
// per-field way to say "refuse this file if you don't know what an
// emblem is", so the version is it. checkSchema already refuses a
// file newer than the reader (ErrSchemaTooNew); the bump is what
// makes it fire. The cost is that a pre-emblem binary refuses EVERY
// post-emblem restore point, emblem or not.
//
// v3 is #945: a stored game.CastPermission carried its own window as
// `untilTurn` / `whileInZone`, and now carries ADR 0063's
// `duration` object instead. A v2 file decodes into the zero
// Duration, which reads as "until end of turn, unstamped" — so every
// granted permission in a restored pre-v3 game would lapse at the
// next cleanup step, including the airbend and warp grants that
// should have lasted for as long as the card stayed in exile. That is
// a semantic shift in an existing field, which is exactly what this
// constant is for.
//
// v4 is #1032 (ADR 0075): the table's configuration moved from the
// single `undoLimit` field into the `settings` object. A v3 file has
// no `settings` key, and its zero value is not a table anyone set up:
// StartingLife and CommanderDamage 0, and an UndoLimit of 0 that the
// v3 binary's Start would have rewritten to the default. So a pre-v4
// file is migrated ONCE, by version, in migrateLegacySettings — and a
// v4 file is restored as written, which is what lets a deliberate
// "no undos" survive a restore. The other direction needs the bump
// too: a v3 binary reading a v4 file would drop every setting.
//
// v5 is #521, and it is the EMBLEM shape of the argument rather than
// the Treasure one. The new `tokenKey` on a card is additive and
// zero-values correctly in the direction that usually matters: a v4
// file has no key, restore reads none, and the card comes back
// exactly as a v4 binary would have restored it. What forces the
// bump is the other direction. Before this change a board holding a
// live token was censused and never written as a restore point at
// all, so no file a v4 binary could be handed contained one; after
// it, restore points full of Treasures, Food and Clues are the
// normal output. A v4 binary reading one would drop the key it does
// not know about and restore those tokens as blank artifacts —
// silently, with a Treasure that no longer taps for mana and a Food
// nobody can eat. There is no per-field way to say "refuse this file
// if you do not know what a token key is", so the version is it, and
// the cost is the one v2 accepted: a pre-#521 binary refuses every
// post-#521 restore point, tokens or not, rather than restoring one
// wrong.
//
// v6 is #1199 (ADR 0084), and it is the emblem shape of the argument
// again rather than the Treasure one. The new `phasedOut` holding
// slice zero-values correctly in the direction that usually matters: a
// v5 file has no key, restore reads none, and a game in which nothing
// was phased out comes back exactly as a v5 binary would have restored
// it. What forces the bump is the other direction. A v5 binary handed
// a v6 file would drop the key it does not know about and restore the
// game with somebody's WHOLE BOARD silently gone — Teferi's Protection
// is four permanents on a light board and twenty on a heavy one, and
// none of them would be anywhere. There is no per-field way to say
// "refuse this file if you do not know what phasing is", so the
// version is it, and the cost is the one v2 and v5 accepted.
//
// v7 is #1497 (ADR 0041 phase 3), and it is the emblem shape of the
// argument once more. `scopedEffects` is new and zero-values correctly
// for every older file: a v6 restore point never held a live scoped
// continuous effect, because the census kept every one of them out.
// What forces the bump is the other direction. A v6 binary handed a v7
// file would drop the key it does not know and restore an earthbent
// land as a plain land, an Agent of Treachery's theft as nothing — and
// v7 files hold these routinely, because making them restore points is
// the whole change. v7 also READS the tier-2 keys (`body` and
// `condition` on a delayed trigger, `body` on a stack item) before any
// binary writes them, and refuses a file naming an effect key it
// cannot interpret (ErrUnknownEffectKey) — which is what lets the mod
// vocabulary and, later, the effect bodies grow within v7 without a
// bump each (ADR 0041 P4). The refusal covers every way a newer v7
// vocabulary can appear in a scoped effect, so each is additive within
// v7: a mod KIND, a duration KIND or CONDITION (bare ints on disk,
// checked against DurationKind.Known / DurationCondition.Known), and a
// JSON FIELD on the record, an affected member, a mod or its duration
// (unknownScopedEffectFields, which encoding/json would otherwise drop
// in silence). Anything else about a scoped effect that would change
// what an older v7 file means — a kind whose meaning changes, a field
// repurposed — is a new kind or a bump, never an edit.
//
// THE COMPATIBILITY RULE, and what enforces it (#522, ADR 0044
// decision 7). The paragraphs above are the judgement; these are the
// tests that make somebody exercise it.
//
//   - Within a version, changes are ADDITIVE ONLY: a new key whose zero
//     value is right for every older file AND whose absence a binary
//     from before it can live with (the other direction — v2, v5 and
//     v6 were all bumps for that half alone). Nothing else is.
//   - A key renamed, removed or retyped, a unit changed, or an existing
//     key given a new meaning FORCES A BUMP. So does an additive change
//     that fails either half of the line above.
//   - testdata/snapshot_shape/v<N>.txt records the on-disk shape of this
//     version, every JSON path and its type, derived by reflection.
//     TestSnapshotShapeIsRecorded fails on any difference. An additive
//     one is recorded in the SAME change by re-running with
//     -update-shape; a non-additive one cannot be recorded at all under
//     the old number — -update-shape refuses — so it has to come with a
//     bump, which writes v<N+1>.txt and freezes v<N>.txt.
//   - testdata/snapshots/v<N>/ is a frozen set of restore points written
//     by the binary that introduced version N, and testdata/snapshots/
//     real/ holds scrubbed ones copied off cmd-dev. FIXTURES ARE NEVER
//     EDITED OR REGENERATED; a bump adds a directory
//     (TestWriteSnapshotCorpus) and every older one stays. On every CI
//     run the corpus test in internal/cards/effects restores each
//     fixture strictly and fails if any key or value in it is missing
//     from a fresh capture of the restored game — which is exactly what
//     a rename looks like to yesterday's file. A bump that migrates an
//     old shape lists the migrated paths in that test's
//     corpusMigrations; nothing else excuses a difference.
//   - A card whose catalog entry lost abilities between the writing and
//     the reading binary is not a schema question at all: restore flags
//     it and reports it (abilityShortfallOf), and the corpus fails on it.
//
// AGENTS.md ("Snapshot compatibility") has the commands.
//
// Restore REFUSES anything it does not recognise rather than guessing.
// See ErrSchemaTooNew / ErrSchemaUnsupported and ADR 0041 for the
// version-skew policy this implements.
const SnapshotSchemaVersion = 7

// settingsSchemaVersion is the first schema that carries
// GameSnapshot.Settings. Older files are migrated from UndoLimit.
const settingsSchemaVersion = 4

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

	// ErrUnknownEffectKey means the file names an effect key — a
	// ScopedEffect mod kind, or an ADR 0041 P2 body or condition key —
	// that this binary cannot interpret. Every key a binary writes is
	// one it registered, so this is the rollback case, and the answer
	// is ErrSchemaTooNew's: refuse, keep the file (ADR 0041 P4).
	ErrUnknownEffectKey = errors.New("game: snapshot names an effect this server cannot interpret")
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

	// turnSeqPresent distinguishes a current pre-game snapshot (Seq is
	// present and zero) from a pre-ADR-0059 snapshot (Seq is absent).
	// Decode metadata only; not game state and never written.
	turnSeqPresent bool

	// unknownEffectFields lists JSON keys on a scopedEffects record, an
	// affected member, a mod or a duration — or on any delayed trigger's
	// or stack item's params, down through its filter and object — that
	// this binary's types do not have (#1497, #1568 reviews; ADR 0041 P4). encoding/json would drop
	// them silently; checkEffectKeys refuses them instead. Decode
	// metadata only; not game state and never written.
	unknownEffectFields []string

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

	// PhasedOut is the CR 702.26 holding slice (#1199, ADR 0084). Not
	// a zone in the CR 400 sense — see Game.PhasedOut — but a *Zone in
	// the code, so it snapshots and restores through the same two
	// helpers with no new mirror type and no new census counter:
	// there is nothing in it but plain Cards.
	PhasedOut *zoneSnapshot `json:"phasedOut,omitempty"`

	Turn          Turn      `json:"turn"`
	MulligansOpen bool      `json:"mulligansOpen"`
	Monarch       uuid.UUID `json:"monarch"`
	Initiative    uuid.UUID `json:"initiative"`
	// UndoLimit is the pre-v4 home of the undo budget. Read only when
	// migrating an older file (migrateLegacySettings); a v4 capture
	// leaves it zero and it is omitted.
	UndoLimit int `json:"undoLimit,omitempty"`
	// Settings is the table's configuration (ADR 0075). Carried from
	// v4; see settingsSchemaVersion.
	Settings          TableSettings `json:"settings"`
	StartingSeat      int           `json:"startingSeat"`
	SplitSecondActive bool          `json:"splitSecondActive"`

	// Outcome is the result of an ended game (ADR 0057 Decision 5):
	// nil while active and for a table ended by End(). omitempty, so
	// no schema bump. A restore point is removed once a game ends, so
	// this matters for fixtures and forensics rather than restarts.
	Outcome *GameOutcome `json:"outcome,omitempty"`
	// ActiveSeatLeftPending is ADR 0057 Decision 3's deferred
	// departure of an active player who lost by an effect
	// mid-resolution, consumed by the next SBA loss pass.
	ActiveSeatLeftPending bool `json:"activeSeatLeftPending,omitempty"`

	// StackMeta is a SLICE, not a map: map iteration order is
	// unspecified and the stack is ordered by Seq anyway. Sorted on
	// capture so two snapshots of the same state are byte-identical.
	StackMeta       []stackItemSnapshot      `json:"stackMeta,omitempty"`
	PendingTriggers []stackItemSnapshot      `json:"pendingTriggers,omitempty"`
	DelayedTriggers []delayedTriggerSnapshot `json:"delayedTriggers,omitempty"`

	// ScopedEffects is ADR 0041 phase 3's data-backed continuous
	// effects (scoped_effects.go, #1497): carried verbatim, because a
	// record holds nothing but data. v7.
	ScopedEffects []ScopedEffect `json:"scopedEffects,omitempty"`

	LoyaltyActivatedThisTurn map[uuid.UUID]bool        `json:"loyaltyActivatedThisTurn,omitempty"`
	SpellsCastThisTurn       map[uuid.UUID]CastTally   `json:"spellsCastThisTurn,omitempty"`
	ForetoldThisTurn         map[uuid.UUID]int         `json:"foretoldThisTurn,omitempty"`
	LandsPlayedThisTurn      map[uuid.UUID]int         `json:"landsPlayedThisTurn,omitempty"`
	ExtraLandDropsThisTurn   map[uuid.UUID]int         `json:"extraLandDropsThisTurn,omitempty"`
	DrawnThisTurn            map[uuid.UUID][]uuid.UUID `json:"drawnThisTurn,omitempty"`
	TurnTally                TurnTally                 `json:"turnTally"`

	// Activations is the per-(object, ability) activation record
	// (#1181). Carried because its game-lifetime half IS game state a
	// player can lose on: a restore that forgot it would hand every
	// exhaust ability on the board back.
	Activations ActivationTally `json:"activations,omitempty"`

	// LoopNotice / LoopThreshold are the CR 726 loop breaker (#628).
	// Both carried: a restore that dropped the notice would resume a
	// table into a live loop with automatic passing back on, and one
	// that dropped the threshold would silently re-default a game a
	// test had configured.
	LoopNotice    *LoopNotice `json:"loopNotice,omitempty"`
	LoopThreshold int         `json:"loopThreshold,omitempty"`

	DiscardPending map[uuid.UUID]int `json:"discardPending,omitempty"`

	// Promises is a slice because its live form is keyed by a
	// STRUCT (PromiseKey), and a JSON object key must be a string.
	Promises []promiseSnapshot `json:"promises,omitempty"`

	Vote *Vote `json:"vote,omitempty"`

	PendingChoices []pendingChoiceSnapshot `json:"pendingChoices,omitempty"`

	Events   []Event `json:"events,omitempty"`
	EventSeq uint64  `json:"eventSeq"`

	// EventBatch is the live Event.Batch counter and OncePerBatchFired
	// the per-ability record of the batch each "whenever one or more …"
	// ability last fired for (#829, event_batch.go). Both restore as
	// zero / empty from a file written before them, which reads as
	// "no batch has fired yet" — the safe direction: the first event
	// after the restore opens a fresh batch and every OncePerBatch
	// ability is free to fire for it. No schema bump.
	EventBatch        uint64            `json:"eventBatch,omitempty"`
	OncePerBatchFired map[string]uint64 `json:"oncePerBatchFired,omitempty"`

	// ResolutionOpen is Game.resolutionOpen (#1289,
	// resolution_pause.go): a resolution has begun and the CR 704.3
	// boundary after it has not run. Read with each choice's
	// MidResolution. A file written before it restores false, which is
	// the old behaviour (the boundary is not held). No schema bump.
	ResolutionOpen bool `json:"resolutionOpen,omitempty"`

	// AnnouncedBlocks / BlockedAttackers are what the block
	// declaration's lock-in produced (#830, #715, blockers.go): which
	// blocker has had its EventBlock announced against which
	// attacker, and which attackers are BLOCKED (CR 509.1h) — the
	// same set that has had its one EventBecomesBlocked (CR 506.4).
	// Both are empty outside a combat with blockers declared, and a
	// file written before them restores as empty — which reads as
	// "nothing announced and nothing blocked yet", so the next
	// lock-in announces the declaration the battlefield already
	// carries. No schema bump.
	//
	// BlockedAttackers keeps the wire key the field was born with
	// (#830 called it announcedBecameBlocked, before #715 gave the
	// same set its rules name): identical contents, identical
	// lifetime, so a file written before the rename restores a
	// mid-combat blocked state rather than losing it.
	AnnouncedBlocks  map[uuid.UUID]uuid.UUID `json:"announcedBlocks,omitempty"`
	BlockedAttackers map[uuid.UUID]bool      `json:"announcedBecameBlocked,omitempty"`

	// AnnouncedAttacks is the attack declaration's half of the same
	// bookkeeping (#859, attackers.go): the creatures that have had
	// their one EventAttack this combat, plus the permanents put onto
	// the battlefield already attacking, which never get one
	// (CR 506.3c). Empty outside combat, and a file written before it
	// restores as empty — which reads as "nothing announced yet", so
	// the next lock-in announces the declaration the battlefield
	// already carries. No schema bump.
	AnnouncedAttacks map[uuid.UUID]bool `json:"announcedAttacks,omitempty"`

	// BlocksDeclared is the set of defending players whose CR 509.1
	// block declaration is complete this combat (#1279,
	// Game.blocksDeclared). Empty outside the declare-blockers step's
	// combat, and a file written before it restores as empty — which
	// reads as "every defender still pending", the conservative
	// direction: the defender is asked again and ninjutsu waits. No
	// schema bump.
	BlocksDeclared map[uuid.UUID]bool `json:"blocksDeclared,omitempty"`

	// AttacksDeclared is whether this combat's attack declaration has
	// passed its CR 508.1d requirement checkpoint (#1571,
	// Game.attacksDeclared). False outside the declare-attackers step's
	// combat, and a file written before it restores as false — the
	// declaration is judged again at the active player's next pass,
	// which can only ask for an attack the requirements already asked
	// for. No schema bump.
	AttacksDeclared bool `json:"attacksDeclared,omitempty"`

	// AttackDefenders is each attacker's defending player as its
	// attack was last pointed (#1364, Game.attackDefenders) — what
	// lets a creature whose planeswalker or battle has left still be
	// blocked (CR 506.4c). Empty outside combat. A file written
	// before it restores as empty, which is the old behaviour: such
	// an attacker cannot be blocked until combat ends. No schema bump.
	AttackDefenders map[uuid.UUID]uuid.UUID `json:"attackDefenders,omitempty"`

	// FirstStrikeStepParticipants is the CR 510.4 / 702.7c
	// participation record for the combat damage steps (#716): the
	// combatants that had first strike or double strike as the first
	// one began. Empty outside a combat that had a first-strike step,
	// and a file written before it restores as empty — which reads as
	// "there was no first-strike step", so every combatant deals
	// damage in the regular step. That is the right answer for every
	// older file except one paused in the priority window between the
	// two steps, which could not exist before the steps did. No schema
	// bump.
	FirstStrikeStepParticipants map[uuid.UUID]bool `json:"firstStrikeStepParticipants,omitempty"`

	// LastKnownBattlefield is CR 603.10 LKI. Empty in steady state —
	// entries live for the duration of one LTB-emitting mutation —
	// but carried so a round-trip is exact rather than nearly exact.
	LastKnownBattlefield     map[uuid.UUID]Characteristic     `json:"lastKnownBattlefield,omitempty"`
	LastKnownTriggerIdentity map[uuid.UUID]triggerIdentityLKI `json:"lastKnownTriggerIdentity,omitempty"`
	// LastKnownCounters is lastKnownBattlefield's sibling for a card's
	// counters (#1218) — see the field doc on game.go.
	LastKnownCounters map[uuid.UUID]map[string]int `json:"lastKnownCounters,omitempty"`
	// LastKnownPermanents is CR 608.2h LKI for permanents that left the
	// battlefield this turn (#1379) — see permanent_lki.go.
	LastKnownPermanents map[uuid.UUID][]PermanentInfo `json:"lastKnownPermanents,omitempty"`

	// LastKnownStack is CR 608.2h LKI for spells that left the stack
	// this turn without resolving (#1255, stack_lki.go). Carried since
	// ADR 0041 phase 3 tier 4 (#1497, P9): its readers — storm's copies
	// of a countered spell, "copy that spell" — become data in that
	// tier, and a restore that dropped it would make a restored storm
	// miss them. Sorted by card ID. Additive within v7 under P10's
	// second rule: a binary before it drops the key from files it
	// writes itself, so carrying it makes nothing worse for it.
	LastKnownStack []lastKnownSpellSnapshot `json:"lastKnownStack,omitempty"`

	RNG               rngSnapshot          `json:"rng"`
	SourceOrdinals    map[uuid.UUID]uint64 `json:"sourceOrdinals,omitempty"`
	SourceOrdinalNext uint64               `json:"sourceOrdinalNext,omitempty"`

	LayerVersion        uint64 `json:"layerVersion"`
	LastResolvedVersion uint64 `json:"lastResolvedVersion"`

	// Continuations records what could not be represented. Empty
	// census == full-fidelity restore point.
	Continuations ContinuationCensus `json:"continuations"`
}

// UnmarshalJSON records whether the embedded Turn carried Seq without
// changing the public snapshot shape. The distinction matters only at
// zero: a current lobby snapshot legitimately has Seq 0, while an old
// active-game snapshot has no Seq key and needs the identity backfill.
func (s *GameSnapshot) UnmarshalJSON(data []byte) error {
	type snapshotAlias GameSnapshot
	if err := json.Unmarshal(data, (*snapshotAlias)(s)); err != nil {
		return err
	}
	var envelope struct {
		Turn map[string]json.RawMessage `json:"turn"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	_, s.turnSeqPresent = envelope.Turn["Seq"]
	if !s.turnSeqPresent {
		_, s.turnSeqPresent = envelope.Turn["seq"]
	}
	fields, err := unknownScopedEffectFields(data, s.Schema)
	if err != nil {
		return err
	}
	s.unknownEffectFields = fields
	return nil
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
	// Emblems is the other half of the command zone (CR 114, #623).
	// Absent from every pre-S40 file, which restores as an empty
	// zone — the right reading, because no game written before
	// emblems existed had one. See SnapshotSchemaVersion for why the
	// schema was bumped anyway.
	Emblems *zoneSnapshot `json:"emblems,omitempty"`
	// CommanderDamage is keyed by commander card instance ID since
	// S25 (#77). A snapshot written before that rekey restores with
	// player-ID keys, which read as damage from commanders that do
	// not exist: harmless (they render nowhere and can never reach
	// 21 again) but not migrated.
	CommanderDamage map[uuid.UUID]int `json:"commanderDamage,omitempty"`
	LifeHistory     []LifeChange      `json:"lifeHistory,omitempty"`
	// TurnsBegun is the seat-turn counter "until your next turn"
	// durations end on (ADR 0063). A file written before S38 restores
	// with 0 for every seat, which is a relative counter reading as
	// "nobody has had a turn yet" — an effect stamped after the
	// restore still ends on that player's next turn, because the
	// stamp is taken from the restored value.
	TurnsBegun         int               `json:"turnsBegun,omitempty"`
	Eliminated         bool              `json:"eliminated"`
	HandKept           bool              `json:"handKept"`
	MulligansTaken     int               `json:"mulligansTaken"`
	DeckImported       bool              `json:"deckImported"`
	UndosRemaining     int               `json:"undosRemaining"`
	DiscordID          string            `json:"discordId,omitempty"`
	DiscordAvatarHash  string            `json:"discordAvatarHash,omitempty"`
	DisplayName        string            `json:"displayName,omitempty"`
	IsBot              bool              `json:"isBot,omitempty"`
	BotTier            string            `json:"botTier,omitempty"`
	BotDeck            string            `json:"botDeck,omitempty"`
	AttemptedEmptyDraw bool              `json:"losesAtNextSba"`
	CommanderCasts     map[uuid.UUID]int `json:"commanderCasts,omitempty"`
	Counters           map[string]int    `json:"counters,omitempty"`
	MaxHandSize        int               `json:"maxHandSize"`
	// LandDropsPerTurn is the player's base land-play allowance
	// (#500). Absent from every pre-#500 snapshot, which would
	// restore as 0 — "may never play a land" — so restorePlayer maps
	// a non-positive value back to DefaultLandDropsPerTurn.
	LandDropsPerTurn int      `json:"landDropsPerTurn,omitempty"`
	ManaPool         ManaPool `json:"manaPool,omitempty"`
	// CastPermissions are the granted cast and play permissions this
	// player holds (ADR 0066). Carried: who may cast what is not
	// derivable from the board, and a restore that dropped them would
	// silently revoke a cascade hit or a Snapcaster'd card that had
	// not been cast yet. Pure data by construction — the type holds no
	// closures, which is what lets it be mirrored rather than
	// rebuilt.
	CastPermissions []CastPermission `json:"castPermissions,omitempty"`
	// Statics are the abilities this PLAYER was granted for a
	// duration — "you gain protection from everything until your next
	// turn" (#1197, CR 702.16i). Carried for the same reason
	// CastPermissions is: not derivable from the board, and pure data
	// by construction, so it is mirrored rather than rebuilt. A file
	// written before #1197 has none, which restores as a player with
	// no granted abilities — the right reading, because no game
	// written before the field existed had one. The DERIVED half
	// (Leyline of Sanctity's "you have hexproof") is not here and
	// needs nothing: it comes back with the battlefield.
	Statics []PlayerStatic `json:"statics,omitempty"`
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
//
// VariableToughness is a *bool so restore can tell a file written
// before #683, which has no such key, from one that says false.
// snapshotCard always sets it, so every file this binary writes
// carries the key; restore backfills it when the key is missing.
type cardSnapshot struct {
	InstanceID               uuid.UUID           `json:"instanceId"`
	Name                     string              `json:"name"`
	ScryfallID               string              `json:"scryfallId,omitempty"`
	OracleID                 string              `json:"oracleId,omitempty"`
	TokenKey                 string              `json:"tokenKey,omitempty"`
	TypeLine                 string              `json:"typeLine,omitempty"`
	Power                    int                 `json:"power"`
	Toughness                int                 `json:"toughness"`
	VariableToughness        *bool               `json:"variableToughness,omitempty"`
	ManaCost                 string              `json:"manaCost,omitempty"`
	ProducedMana             []string            `json:"producedMana,omitempty"`
	Colors                   []string            `json:"colors,omitempty"`
	ColorIdentity            []string            `json:"colorIdentity,omitempty"`
	StartingLoyalty          int                 `json:"startingLoyalty"`
	Keywords                 []string            `json:"keywords,omitempty"`
	GrantedAbilities         []string            `json:"grantedAbilities,omitempty"`
	Layout                   string              `json:"layout,omitempty"`
	Faces                    []Face              `json:"faces,omitempty"`
	ActiveFace               int                 `json:"activeFace,omitempty"`
	PrintedSelf              *PrintedValues      `json:"printedSelf,omitempty"`
	NeedsEffect              bool                `json:"needsEffect"`
	Owner                    uuid.UUID           `json:"owner"`
	Controller               uuid.UUID           `json:"controller"`
	Tapped                   bool                `json:"tapped"`
	NextUntapSkips           []untapSkipSnapshot `json:"nextUntapSkips,omitempty"`
	BattleX                  float64             `json:"battleX"`
	BattleY                  float64             `json:"battleY"`
	Counters                 map[string]int      `json:"counters,omitempty"`
	IsCommander              bool                `json:"isCommander"`
	AttackingTarget          uuid.UUID           `json:"attackingTarget"`
	BlockingTarget           uuid.UUID           `json:"blockingTarget"`
	GoadedBy                 uuid.UUID           `json:"goadedBy"`
	DamageMarked             int                 `json:"damageMarked"`
	RegenerationShields      int                 `json:"regenerationShields,omitempty"`
	FaceDown                 bool                `json:"faceDown"`
	FaceDownKind             FaceDownKind        `json:"faceDownKind,omitempty"`
	KnownBy                  map[uuid.UUID]bool  `json:"knownBy,omitempty"`
	EnteredBattlefieldAt     int64               `json:"enteredBattlefieldAt"`
	ObjectEpoch              int                 `json:"objectEpoch,omitempty"`
	SummonedThisTurn         bool                `json:"summonedThisTurn"`
	MarkedLethalByDeathtouch bool                `json:"markedLethalByDeathtouch"`
	LostLastCounter          bool                `json:"lostLastCounter,omitempty"`
	PrintedPTKnown           bool                `json:"printedPTKnown,omitempty"`
	AttachedTo               TargetRef           `json:"attachedTo,omitempty"`
	AttachedAt               int64               `json:"attachedAt,omitempty"`
	BaseController           uuid.UUID           `json:"baseController,omitempty"`
	FaceDownListed           *FaceDownListing    `json:"faceDownListed,omitempty"`
	FaceTurnedAt             int64               `json:"faceTurnedAt,omitempty"`
	// NamedTribe is the CR 614.12 "as this enters, choose a creature
	// type" answer (S26). Carried rather than rebuilt: the choice was
	// made by a player and nothing in the catalog can re-derive it, so
	// a restore that lost it would leave a Cavern of Souls producing
	// mana for a tribe nobody named.
	NamedTribe string `json:"namedTribe,omitempty"`
	// ChosenColor is the "as this enters, choose a color" answer
	// (#742). Carried for NamedTribe's reason: a player made it and
	// nothing can re-derive it.
	ChosenColor string `json:"chosenColor,omitempty"`
	// ChosenPlayer is the CR 614.12 "as this enters, choose a player"
	// answer (#980). Carried for NamedTribe's reason and one more: it
	// is the whole of what True-Name Nemesis's protection reads, so a
	// restore that lost it would bring the Nemesis back protected from
	// nobody — a rules change, silently, with nothing in the state to
	// say what went. Old snapshots have no key and their zero uuid
	// reads correctly as "nobody chosen".
	ChosenPlayer uuid.UUID `json:"chosenPlayer,omitempty"`
	// ChosenName is the CR 614.12 "as this enters, choose a card
	// name" answer (#1210). Carried for ChosenPlayer's reason and one
	// more: it is free text, so there is not even a vocabulary a
	// restore could have re-derived it from — a Pithing Needle that
	// came back with an empty name would silently stop restricting
	// the card it was played to stop.
	ChosenName string `json:"chosenName,omitempty"`
	// Provenance is CR 400.7d: what the spell that became this
	// permanent was cast for — the alternative cost (#653) and the
	// optional additional costs (#664, ADR 0073 §5), in one record.
	// Carried, and for a sharper reason than NamedTribe's: the stack
	// item it was copied from is gone by definition, so a restore that
	// dropped it could not rebuild it from anything. A Phlage that
	// escaped would come back having been hard-cast and sacrifice
	// itself on the next trigger, and a kicked Gatekeeper of Malakir
	// would come back unkicked, both silently.
	//
	// Absent in a file written before the field existed, which decodes
	// as the zero record: "not cast, or cast for its mana cost", the
	// answer every permanent gave before #653.
	//
	// No `omitzero`: `encoding/json` only honours that option from Go
	// 1.24 (#1492), and CI's pinned 1.22 toolchain — the one that
	// builds every fixture in testdata/snapshots and every deployed
	// binary — silently ignores it and always writes the field. A
	// contributor's newer local toolchain honouring the option is
	// what produced the divergence; always writing it, on every Go
	// version, is what removes it.
	Provenance CastProvenance `json:"provenance"`
	// ClassLevel is the CR 716.2 level designation and Solved the
	// CR 719.3 solved designation (ADR 0071 decision 6). Both carried,
	// for NamedTribe's reason and one more: they are legal zero
	// values, so a restore that dropped them would bring back a
	// level-3 Wizard Class as a level-1 one with two of its three
	// abilities gone, and a solved Case unsolved — silently, with
	// nothing in the state to say anything was lost. Old snapshots
	// have neither key, and their zero values read correctly as
	// "level 1, unsolved", which is why snapshot_backfill.go needs no
	// arm for them.
	ClassLevel int  `json:"classLevel,omitempty"`
	Solved     bool `json:"solved,omitempty"`
	// Harnessed is the CR 701.64 harnessed designation (ADR 0071
	// amendment, #1321), carried for the same reason as ClassLevel /
	// Solved: it is a legal zero value, and an old snapshot with no
	// key decodes as "not harnessed", which is what every game before
	// this amendment was.
	Harnessed bool `json:"harnessed,omitempty"`
	// Prepared, PrepareCopy and PreparedBy are ADR 0090's CR 722.3
	// state: the designation on the permanent, and the not-a-card
	// marker and permanent link on the copy it keeps in exile. Carried
	// for Solved's reason — all three zero values are legal states, so
	// a restore that dropped them would bring back an unprepared
	// permanent beside a copy nobody may cast, or a copy that is a
	// real card — and old snapshots decode as "not prepared, no copy",
	// which is what every game before ADR 0090 was.
	Prepared    bool `json:"prepared,omitempty"`
	PrepareCopy bool `json:"prepareCopy,omitempty"`
	// No `omitzero` on PreparedBy or HiddenBy below — see Provenance's
	// comment above (#1492).
	PreparedBy PermissionCardRef `json:"preparedBy"`
	// HiddenBy is ADR 0091's hideaway link: the permanent object that
	// exiled this card face down. Carried because nothing else records
	// it — a restore that dropped it would leave the card unplayable by
	// the land that hid it and unreadable by that land's controller.
	HiddenBy          PermissionCardRef `json:"hiddenBy"`
	StartingDefense   int               `json:"startingDefense,omitempty"`
	ProtectorPlayerID uuid.UUID         `json:"protectorPlayerId,omitempty"`

	// The CR 702.26 phased-out status (#1199, ADR 0084). Meaningful
	// only for a card in GameSnapshot.PhasedOut, and carried for
	// ClassLevel's reason: every one of them is a legal zero value, so
	// a restore that dropped them would bring back somebody's phased
	// board under the wrong player's untap step, or bring an Aura back
	// without its host.
	PhasedOutBy       uuid.UUID `json:"phasedOutBy,omitempty"`
	PhaseInLockedBy   uuid.UUID `json:"phaseInLockedBy,omitempty"`
	PhasedOutIndirect bool      `json:"phasedOutIndirect,omitempty"`
	TapOnPhaseIn      bool      `json:"tapOnPhaseIn,omitempty"`

	// ManaAbilityCount / ActivatedAbilityCount record that the card
	// HAD intrinsic ability closures, so restore can tell the
	// difference between "none" and "some it must rebuild", and so
	// the census can name the card. Since #522 restore also COMPARES
	// them against what the catalog handed back — see
	// abilityShortfallOf.
	ManaAbilityCount      int `json:"manaAbilityCount,omitempty"`
	ActivatedAbilityCount int `json:"activatedAbilityCount,omitempty"`

	// CatalogAbilities is the catalog entry this card resolved to at
	// capture, measured: how many abilities of each kind the capturing
	// binary's catalog held for it. Present only for a card that HAD an
	// entry (IsAutoCard), and nil for nearly everything in a library.
	// It is the other half of the #522 parity check: the instance
	// counts above exist only for tokens and token copies, because a
	// printed card reads its abilities from the catalog at use time, so
	// without this record a card whose entry vanished between two
	// binaries would come back silently inert.
	//
	// Additive, and no schema bump in either direction: a file written
	// before it has no record and restore checks only the instance
	// counts, which is what the binary before it did; a binary before
	// it reading a newer file drops the key and does the same.
	CatalogAbilities *AbilityCounts `json:"catalogAbilities,omitempty"`

	// AbilitiesLostOnRestore is Card.AbilitiesLostOnRestore. Absent
	// from every file written before #522, which reads as "not
	// flagged" — the answer every card gave before the flag existed.
	// No schema bump: a binary that predates it drops the key and
	// shows the card as automated again, which is cosmetic, not a
	// rules change.
	AbilitiesLostOnRestore bool `json:"abilitiesLostOnRestore,omitempty"`
}

// AbilityCounts is how many catalog abilities of each kind a card
// has, by the slot the engine reads them from (#522). It measures a
// catalog ENTRY, not the card: two binaries can disagree about it, and
// restore's parity check is where they are compared.
type AbilityCounts struct {
	Mana         int `json:"mana,omitempty"`
	Activated    int `json:"activated,omitempty"`
	Triggered    int `json:"triggered,omitempty"`
	Static       int `json:"static,omitempty"`
	Replacements int `json:"replacements,omitempty"`
}

// Total is every ability counted.
func (a AbilityCounts) Total() int {
	return a.Mana + a.Activated + a.Triggered + a.Static + a.Replacements
}

// fewerThan reports whether any slot of a is below the same slot of b.
// A slot ABOVE is not a shortfall: a new build adding abilities to a
// card is normal (owner decision on #515).
func (a AbilityCounts) fewerThan(b AbilityCounts) bool {
	return a.Mana < b.Mana || a.Activated < b.Activated ||
		a.Triggered < b.Triggered || a.Static < b.Static ||
		a.Replacements < b.Replacements
}

// catalogAbilityCounts measures the running binary's catalog entry
// for key. It reads through the per-slot hooks rather than the
// CardDef, so a test that stubs one slot is measured the way the
// engine reads it.
func catalogAbilityCounts(key string) AbilityCounts {
	var n AbilityCounts
	if key == "" {
		return n
	}
	if CatalogManaAbilities != nil {
		n.Mana = len(CatalogManaAbilities(key))
	}
	if CatalogActivatedAbilities != nil {
		n.Activated = len(CatalogActivatedAbilities(key))
	}
	if CatalogTriggers != nil {
		n.Triggered = len(CatalogTriggers(key))
	}
	if CatalogStaticAbilities != nil {
		n.Static = len(CatalogStaticAbilities(key))
	}
	if CatalogReplacements != nil {
		n.Replacements = len(CatalogReplacements(key))
	}
	return n
}

// AbilityShortfall is one card a restore brought back with fewer
// catalog abilities than the restore point recorded (#522). Reported
// by GameSnapshot.AbilityShortfalls and logged at ERROR by the boot
// restore path; the restored card is flagged AbilitiesLostOnRestore.
type AbilityShortfall struct {
	CardID   uuid.UUID
	Name     string
	OracleID string
	TokenKey string
	Zone     ZoneKind
	// Captured is what the file recorded: per slot, the larger of the
	// instance count and the catalog-entry count.
	Captured AbilityCounts
	// Restored is what this binary's catalog hands back for the same
	// key.
	Restored AbilityCounts
	// EntryMissing is the sharpest case: the capturing binary had a
	// catalog entry for the card and this one has none, so a spell
	// with no abilities at all — a Lightning Bolt — is caught too.
	EntryMissing bool
}

// abilityShortfallOf is the #522 parity check for one decoded card:
// did this binary's catalog hand back at least what the restore point
// recorded? It is ONE function because restoreCard (which flags the
// card) and AbilityShortfalls (which reports it) must never disagree
// about which cards were short.
//
// "Recorded" is two measurements. The instance counts
// (ManaAbilityCount / ActivatedAbilityCount) are the closures a token
// or token copy carried, which restore re-derives by key; the
// CatalogAbilities record is the entry a printed card reads at use
// time. Both are compared against the running catalog under
// abilityCatalogKey — the key capture measured under — so the two
// sides ask one question.
func abilityShortfallOf(c *cardSnapshot) (AbilityShortfall, bool) {
	captured := AbilityCounts{Mana: c.ManaAbilityCount, Activated: c.ActivatedAbilityCount}
	hadEntry := c.CatalogAbilities != nil
	if hadEntry {
		rec := *c.CatalogAbilities
		captured.Mana = max(captured.Mana, rec.Mana)
		captured.Activated = max(captured.Activated, rec.Activated)
		captured.Triggered = max(captured.Triggered, rec.Triggered)
		captured.Static = max(captured.Static, rec.Static)
		captured.Replacements = max(captured.Replacements, rec.Replacements)
	}
	if !hadEntry && captured.Total() == 0 {
		return AbilityShortfall{}, false
	}
	key := restoreAbilityKey(c)
	restored := catalogAbilityCounts(key)
	entryMissing := hadEntry && !IsAutoCard(key)
	if !entryMissing && !restored.fewerThan(captured) {
		return AbilityShortfall{}, false
	}
	return AbilityShortfall{
		CardID:       c.InstanceID,
		Name:         c.Name,
		OracleID:     c.OracleID,
		TokenKey:     c.TokenKey,
		Captured:     captured,
		Restored:     restored,
		EntryMissing: entryMissing,
	}, true
}

// AbilityShortfalls lists every card this binary restores with fewer
// catalog abilities than the snapshot recorded (#522), in zone order:
// battlefield, stack, exile, phased out, then each seat's library,
// hand, graveyard, command zone and emblems.
//
// Restore flags each of them AbilitiesLostOnRestore through the same
// check; this is the half the boot path logs. It reads the snapshot
// and the running catalog and nothing else, so it gives the same
// answer before or after Restore.
func (s *GameSnapshot) AbilityShortfalls() []AbilityShortfall {
	var out []AbilityShortfall
	walk := func(z *zoneSnapshot) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			if sf, ok := abilityShortfallOf(&z.Cards[i]); ok {
				sf.Zone = z.Kind
				out = append(out, sf)
			}
		}
	}
	walk(s.Battlefield)
	walk(s.Stack)
	walk(s.Exile)
	walk(s.PhasedOut)
	for i := range s.Seats {
		p := &s.Seats[i]
		walk(p.Library)
		walk(p.Hand)
		walk(p.Graveyard)
		walk(p.Command)
		walk(p.Emblems)
	}
	return out
}

// stackItemSnapshot mirrors StackItem. Effect and targetSpec are both
// func-bearing; see rehydrateStackItem for which ones come back.
type stackItemSnapshot struct {
	ID            uuid.UUID     `json:"id"`
	Kind          StackItemKind `json:"kind"`
	Controller    uuid.UUID     `json:"controller"`
	Owner         uuid.UUID     `json:"owner"`
	SourceCardID  uuid.UUID     `json:"sourceCardId"`
	SourceEpoch   int           `json:"sourceEpoch,omitempty"`
	SourceObject  *ObjectRef    `json:"sourceObject,omitempty"` // #1418; nil = unstamped
	Label         string        `json:"label,omitempty"`
	DoubledBy     uuid.UUID     `json:"doubledBy,omitempty"`
	DoubledByName string        `json:"doubledByName,omitempty"`
	Targets       []TargetRef   `json:"targets,omitempty"`
	Payload       []TargetRef   `json:"payload,omitempty"`
	// Trigger is the triggering event (#1223). Carried, and it has
	// to be: a targeted trigger waiting on its CR 603.3d prompt is a
	// restorable snapshot, and a restore that lost the event would
	// resolve Scrap Trawler against a mana value nothing on the
	// restored board remembers — the artifact it is about is in a
	// graveyard, where the characteristics it died with are gone.
	// Plain data, so unlike the Effect beside it there is nothing to
	// census.
	Trigger      *TriggerContext   `json:"trigger,omitempty"`
	Modes        []int             `json:"modes,omitempty"`
	XValue       int               `json:"xValue"`
	Distribution map[uuid.UUID]int `json:"distribution,omitempty"`
	HoldPriority bool              `json:"holdPriority"`
	CastFromZone ZoneKind          `json:"castFromZone,omitempty"`
	AltCost      string            `json:"altCost,omitempty"`
	Foretold     bool              `json:"foretold,omitempty"`
	// FaceDown is CR 708.4 (#1194): the state the permanent this
	// spell becomes enters in. Carried, and it has to be — a restore
	// that lost it would resolve a morph on the stack into a face-UP
	// creature, revealing the card and giving it every ability
	// CR 708.2a says it does not have.
	FaceDown      FaceDownKind `json:"faceDown,omitempty"`
	AltCostExiles bool         `json:"altCostExiles,omitempty"`
	SplitSecond   bool         `json:"splitSecond"`
	IsCopy        bool         `json:"isCopy,omitempty"`
	Uncopyable    bool         `json:"uncopyable,omitempty"` // #1574
	Seq           uint64       `json:"seq"`
	Ordered       bool         `json:"ordered"`
	Commutes      bool         `json:"commutes,omitempty"` // #1511

	// Paid is what the announcement cost (#789 / #761). Carried: a
	// restore that lost it would resolve a converge spell for zero
	// and a "for each counter removed this way" ability for nothing,
	// and neither number is recoverable from the restored board.
	Paid PaidCost `json:"paid,omitempty"`

	// HasEffect / HasTargetSpec / HasModeSpec record the catalog
	// slots so the census can count what restore had to drop.
	HasEffect     bool `json:"hasEffect,omitempty"`
	HasTargetSpec bool `json:"hasTargetSpec,omitempty"`
	HasModeSpec   bool `json:"hasModeSpec,omitempty"`

	// OracleID is the source card's oracle ID, captured so restore
	// can re-derive a SPELL's target spec from the catalog without
	// having to find the card again (it may have moved zones).
	OracleID string `json:"oracleId,omitempty"`

	// Body is ADR 0041 P2's effect-body key (tier 2, #1497): what a
	// fired delayed trigger resolves through. Restore re-derives the
	// item's Effect from it, and refuses a key this binary does not
	// have (ErrUnknownEffectKey). Params is the body's plain data.
	Body   string        `json:"body,omitempty"`
	Params *EffectParams `json:"params,omitempty"` // nil when zero: most items are spells
}

type delayedTriggerSnapshot struct {
	ID                 uuid.UUID   `json:"id"`
	Controller         uuid.UUID   `json:"controller"`
	SourceCardID       uuid.UUID   `json:"sourceCardId"`
	SourceObject       *ObjectRef  `json:"sourceObject,omitempty"` // #1418
	Label              string      `json:"label,omitempty"`
	At                 Step        `json:"at"`
	ControllerTurnOnly bool        `json:"controllerTurnOnly"`
	CreatedSeq         int         `json:"createdSeq,omitempty"`
	CreatedTurn        int         `json:"createdTurn,omitempty"`
	Cards              []uuid.UUID `json:"cards,omitempty"`
	HasEffect          bool        `json:"hasEffect,omitempty"`
	// #663: the event condition. On and ExpiresAfterTurn are data
	// and come back; the AppliesTo predicate and the Optional prompt
	// are closures and do not, exactly as Effect does not — and the
	// trigger is already counted once in
	// ContinuationCensus.DelayedTriggerEffects, so it is already
	// marked unrestorable and nothing here double-counts it.
	On       []EventKind `json:"on,omitempty"`
	Duration *Duration   `json:"duration,omitempty"`

	// Body, Condition and their params are ADR 0041 P2's data form
	// of what the trigger does and which event fires it (tier 2,
	// #1497). The v7 reader was in place before this writer (PR 1), so
	// no bump: a v7 binary without a key refuses the file
	// (ErrUnknownEffectKey) rather than misreading it.
	Body             string        `json:"body,omitempty"`
	Params           *EffectParams `json:"params,omitempty"` // nil when zero
	Condition        string        `json:"condition,omitempty"`
	CondParams       *EffectParams `json:"condParams,omitempty"` // nil when zero
	OptionalQuestion string        `json:"optionalQuestion,omitempty"`
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
	CoinAllowStop        bool                   `json:"coinAllowStop,omitempty"`
	CoinCount            int                    `json:"coinCount,omitempty"`
	CoinMaxUsefulWins    int                    `json:"coinMaxUsefulWins,omitempty"`
	CoinWins             int                    `json:"coinWins,omitempty"`
	ColorOptions         []string               `json:"colorOptions,omitempty"`
	ColorPurpose         ColorPurpose           `json:"colorPurpose,omitempty"`
	ManaRestrictions     []string               `json:"manaRestrictions,omitempty"`
	ManaRiders           []ManaSpendRider       `json:"manaRiders,omitempty"`
	ManaSourceKinds      ManaSourceKinds        `json:"manaSourceKinds,omitempty"`
	ManaAmounts          map[string]int         `json:"manaAmounts,omitempty"`
	ManaTapped           bool                   `json:"manaTapped,omitempty"`
	ReplacementEffectIDs []ReplacementEffectID  `json:"replacementEffectIds,omitempty"`
	DamageAssignment     *DamageAssignmentFrame `json:"damageAssignment,omitempty"`
	NoLegalTarget        bool                   `json:"noLegalTarget"`
	PickTargetPlayers    []uuid.UUID            `json:"pickTargetPlayers,omitempty"`
	PickTargetCards      []uuid.UUID            `json:"pickTargetCards,omitempty"`
	PickTargetMin        int                    `json:"pickTargetMin"`
	PickTargetMax        int                    `json:"pickTargetMax"`
	// #1196's CR 115.7 retarget prompt. Carried rather than counted
	// as a continuation because it IS the prompt: the item whose
	// targets are being changed, the printed sentence being applied,
	// whether declining is allowed, which slot is being asked, and
	// the card's own header for the rest of the walk. A restored
	// game with these missing would put a question about nothing in
	// front of a seat.
	RetargetItem     uuid.UUID      `json:"retargetItem,omitempty"`
	RetargetPolicy   RetargetPolicy `json:"retargetPolicy,omitempty"`
	RetargetOptional bool           `json:"retargetOptional,omitempty"`
	RetargetSlot     int            `json:"retargetSlot,omitempty"`
	RetargetReason   string         `json:"retargetReason,omitempty"`
	ModeOptionIndex  []int          `json:"modeOptionIndex,omitempty"`
	ModeOptionLabel  []string       `json:"modeOptionLabel,omitempty"`
	ModeMin          int            `json:"modeMin,omitempty"`
	ModeMax          int            `json:"modeMax,omitempty"`
	ModeRepeatable   bool           `json:"modeRepeatable,omitempty"`
	SacrificeOptions []uuid.UUID    `json:"sacrificeOptions,omitempty"`
	CopyOptions      []uuid.UUID    `json:"copyOptions,omitempty"`
	ScryCards        []uuid.UUID    `json:"scryCards,omitempty"`
	// ADR 0088: which lanes a put_in_library answer may use.
	LibraryPlacement LibraryPlacement `json:"libraryPlacement,omitempty"`
	// #1298: the put_in_library top lane's exact count and depth.
	LibraryTopCount int         `json:"libraryTopCount,omitempty"`
	LibraryTopDepth int         `json:"libraryTopDepth,omitempty"`
	TriggerOrderIDs []uuid.UUID `json:"triggerOrderIds,omitempty"`
	PayCost         string      `json:"payCost,omitempty"`
	SearchCards     []uuid.UUID `json:"searchCards,omitempty"`
	SearchMax       int         `json:"searchMax"`
	MayCastCard     uuid.UUID   `json:"mayCastCard,omitempty"`
	AcceptLabel     string      `json:"acceptLabel,omitempty"`
	LifeCost        int         `json:"lifeCost,omitempty"`
	DeclineLabel    string      `json:"declineLabel,omitempty"`
	// OwedInStep: the step a pay-or-else prompt has to be answered in
	// (#997). Carried so a restored game gates the same way, cheap
	// and honest even though every prompt that sets it today also
	// holds a continuation and so blocks the restore point.
	OwedInStep TurnStep `json:"owedInStep,omitempty"`
	// GuardsStackItem: the object on the stack whose fate a
	// counter-unless-pays prompt decides (#951). Carried for
	// OwedInStep's reason — a restored game gates the same way —
	// and because the ID is the only route back to what the
	// question was about.
	GuardsStackItem uuid.UUID   `json:"guardsStackItem,omitempty"`
	ChooseCards     []uuid.UUID `json:"chooseCards,omitempty"`
	ChooseMin       int         `json:"chooseMin,omitempty"`
	// #568: the branches of an option pick. Carried for the reason
	// ChooseCards is — the prompt is the options, and a restored game
	// that forgot them would render a question with no answers.
	PickOptions []ChoiceOption `json:"pickOptions,omitempty"`
	ChooseMax   int            `json:"chooseMax,omitempty"`
	// #804 CR 726 shortcut: which run the answer's allowance attaches
	// to, how many resolutions had happened when it was asked, and
	// whether this is the turn's second ask.
	LoopShortcutKey    string `json:"loopShortcutKey,omitempty"`
	LoopShortcutCount  int    `json:"loopShortcutCount,omitempty"`
	LoopShortcutRepeat bool   `json:"loopShortcutRepeat,omitempty"`
	// MidResolution is PendingChoice.midResolution (#1289).
	MidResolution bool `json:"midResolution,omitempty"`

	// ResumeFrames names the continuation slots that were populated.
	// Diagnostic only — nothing rebuilds them in this schema.
	ResumeFrames []string `json:"resumeFrames,omitempty"`
}

type promiseSnapshot struct {
	From  uuid.UUID `json:"from"`
	To    uuid.UUID `json:"to"`
	Count int       `json:"count"`
}

// rngSnapshot carries the game's randomness (ADR 0054 Decision 3).
//
// Why this is not just a seed: a game that has already shuffled is
// some draws into each stream. Restoring from the key alone would
// re-deal the draws the game has already consumed this turn. What has
// to persist is the key AND how far each stream has got, which is the
// counter map.
//
// Kind discriminates:
//
//	"keyed"    — the only kind this binary writes once a key exists.
//	             Key is the 32-byte secret key, Counters the draws each
//	             stream has taken on turn index Turn. The restored game
//	             continues every stream exactly.
//	"pcg"      — written by binaries before ADR 0054: State held a
//	             *rand.PCG position. Restores with a fresh key.
//	"external" — written before ADR 0054 for a caller-supplied
//	             *rand.Rand whose position was unreadable. Restores
//	             with a fresh key.
//	"none"     — no key yet (game not started, nothing drawn).
//
// A fresh key for the two old kinds loses nothing a player could
// notice: library orders live in the zones, so a new key changes only
// FUTURE draws. No SnapshotSchemaVersion bump: an older binary reading
// "keyed" hits restoreRNG's default branch and degrades to its global
// source until the next deploy, which beats abandoning every live game
// on a rollback (ADR 0041 Decision 5).
//
// The key never leaves the server: it is in this engine snapshot, on
// the server's disk, and never in a GameView, a replay line or a
// crash dump.
type rngSnapshot struct {
	Kind     string            `json:"kind"`
	State    []byte            `json:"state,omitempty"`
	Key      []byte            `json:"key,omitempty"`
	Counters map[string]uint64 `json:"counters,omitempty"`
	Turn     int               `json:"turn,omitempty"`
}

const (
	rngKindNone     = "none"
	rngKindKeyed    = "keyed"
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
	// StackEffects counted stack items (on the stack, queued, or
	// remembered in lastKnownStack) that restore could not rebuild: an
	// Effect that is a bare closure, or a target or mode clause with
	// neither an oracle ID nor a catalog ability ref behind it. RETIRED
	// by ADR 0041 phase 3 tier 4-final (#1497): every production path
	// names its item's body — a catalog row (catalog/activated,
	// catalog/triggered), a tier-2 body or a reflexive body — and the
	// one kind of item that still cannot (an ability carried on the
	// card instance, or a hand-built item) is folded into
	// IntrinsicAbilityCards. Nothing increments this. The field stays
	// so a census written by an older binary still decodes, and still
	// reads as not restorable.
	StackEffects int `json:"stackEffects,omitempty"`

	// StackTargetSpecs counted items whose CR 608.2b re-check clause is
	// a closure and could not be re-derived from the catalog. RETIRED by
	// ADR 0041 phase 3 tier 4 (#1497, P9): a stack item's clauses are
	// re-derived from its oracle ID (a spell), its catalog ability ref
	// (a stamped ability, tier 4-1) or its own registration (a CR 603.12
	// reflexive trigger's ReflexiveBody, tier 4-0), and an item that is
	// none of those is counted ONCE, in StackEffects. Nothing increments
	// this. The field stays so a census written by an older binary
	// still decodes, and still reads as not restorable.
	StackTargetSpecs int `json:"stackTargetSpecs,omitempty"`

	// DelayedTriggerEffects is queued "at the beginning of the next
	// end step" instructions.
	DelayedTriggerEffects int `json:"delayedTriggerEffects,omitempty"`

	// ChoiceResumeFrames is paused prompts holding a continuation.
	ChoiceResumeFrames int `json:"choiceResumeFrames,omitempty"`

	// ScopedStatics counted floating continuous effects held as
	// closures (Giant Growth's +3/+3, Act of Treason's theft). RETIRED
	// by ADR 0041 phase 3 tier 3a (#1497): every such effect is a
	// ScopedEffect record now, which the snapshot carries, and nothing
	// increments this. The field stays so a census written by an older
	// binary still decodes; the wire key stays `turnScopedStatics`, the
	// name it had before the registry grew the other CR 611.2
	// durations.
	ScopedStatics int `json:"turnScopedStatics,omitempty"`

	// TurnScopedReplacements counted floating until-end-of-turn
	// replacement effects held as closures (Fog, a prevention shield,
	// the Whip's redirect). RETIRED by ADR 0041 phase 3 tier 3b
	// (#1497): each is a ScopedEffect record now, which the snapshot
	// carries, and nothing increments this. The field stays so a
	// census written by an older binary still decodes.
	TurnScopedReplacements int `json:"turnScopedReplacements,omitempty"`

	// TurnScopedBlockRules counted floating until-end-of-turn block
	// rules held as closures (Gingerbrute's "can't block this turn",
	// #750). RETIRED by ADR 0041 phase 3 tier 3b (#1497): each is a
	// ScopedEffect record now, which the snapshot carries, and nothing
	// increments this. The field stays so a census written by an older
	// binary still decodes.
	TurnScopedBlockRules int `json:"turnScopedBlockRules,omitempty"`

	// IntrinsicAbilityCards is cards holding an ability closure on
	// the INSTANCE that the catalog cannot hand back — the ones a
	// restore would bring back with the ability missing.
	//
	// #521 narrowed what that means, and the counter is worth reading
	// twice because the old meaning was wider than it sounded. It
	// used to count every card carrying an instance ability and no
	// oracle ID, which tested whether there was an identity to look
	// up rather than whether anything unserialisable was actually
	// there — so a Treasure, whose ability is four fields of plain
	// data, censused itself and blocked every restore point in the
	// game for as long as it sat on the battlefield. A token template
	// now has a catalog key of its own, so the question asked here is
	// the honest one: CAN the registry return this ability? An
	// instance ability with no entry behind it — a closure stamped
	// onto one object at runtime — still counts, and must: the
	// counter is meant to become accurate, not unreachable.
	//
	// Since ADR 0041 tier 4-final (#1497) it also counts, once each, the
	// stack items (on the stack, queued, or in lastKnownStack) that
	// restore cannot rebuild: an Effect that is a closure with no Body,
	// or a target or mode clause nothing re-derives. In production the
	// only such item is an ability carried on a card instance, which is
	// this counter's own subject — and StackEffects, which used to count
	// them, is retired.
	IntrinsicAbilityCards int `json:"intrinsicAbilityCards,omitempty"`

	// UnpersistableRNG marked a game whose random source belonged to
	// the caller and could not be read back out. Since ADR 0054 every
	// game's randomness is a persistable key, so this binary never sets
	// it; the field stays so a census written by an older binary still
	// decodes (and still reads as not restorable).
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
		c.ScopedStatics == 0 &&
		c.TurnScopedReplacements == 0 &&
		c.TurnScopedBlockRules == 0 &&
		c.IntrinsicAbilityCards == 0 &&
		!c.UnpersistableRNG
}

// Total is the number of individual continuations counted.
func (c ContinuationCensus) Total() int {
	n := c.StackEffects + c.StackTargetSpecs + c.DelayedTriggerEffects +
		c.ChoiceResumeFrames + c.ScopedStatics +
		c.TurnScopedReplacements + c.TurnScopedBlockRules + c.IntrinsicAbilityCards
	if c.UnpersistableRNG {
		n++
	}
	return n
}

// Kinds returns the counters that are non-zero, keyed by each
// counter's JSON key — the name an operator already reads in a
// restore point's `continuations` (ADR 0041 P7, #1497): the per-kind
// skip tally the shutdown census logs is keyed by these. ScopedStatics
// is therefore "turnScopedStatics", its JSON key since before S38.
func (c ContinuationCensus) Kinds() map[string]int {
	out := map[string]int{}
	for name, n := range map[string]int{
		"stackEffects":           c.StackEffects,
		"stackTargetSpecs":       c.StackTargetSpecs,
		"delayedTriggerEffects":  c.DelayedTriggerEffects,
		"choiceResumeFrames":     c.ChoiceResumeFrames,
		"turnScopedStatics":      c.ScopedStatics,
		"turnScopedReplacements": c.TurnScopedReplacements,
		"turnScopedBlockRules":   c.TurnScopedBlockRules,
		"intrinsicAbilityCards":  c.IntrinsicAbilityCards,
	} {
		if n > 0 {
			out[name] = n
		}
	}
	if c.UnpersistableRNG {
		out["unpersistableRng"] = 1
	}
	return out
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
		Schema:                SnapshotSchemaVersion,
		TakenAt:               time.Now().UTC(),
		ID:                    g.ID,
		CreatedAt:             g.CreatedAt,
		State:                 g.State,
		Turn:                  g.Turn,
		MulligansOpen:         g.MulligansOpen,
		Monarch:               g.Monarch,
		Initiative:            g.Initiative,
		Settings:              g.Settings,
		StartingSeat:          g.StartingSeat,
		SplitSecondActive:     g.SplitSecondActive,
		Outcome:               cloneGameOutcome(g.Outcome),
		EventSeq:              g.eventSeq,
		EventBatch:            g.eventBatch,
		ResolutionOpen:        g.resolutionOpen,
		ActiveSeatLeftPending: g.ActiveSeatLeftPending,
		turnSeqPresent:        true,
	}
	s.OncePerBatchFired = copyStringUint64Map(g.oncePerBatchFired)
	s.AnnouncedBlocks = copyUUIDPairMap(g.announcedBlocks)
	s.BlockedAttackers = copyBoolMap(g.blockedAttackers)
	s.AnnouncedAttacks = copyBoolMap(g.announcedAttacks)
	s.AttackDefenders = copyUUIDPairMap(g.attackDefenders)
	s.BlocksDeclared = copyBoolMap(g.blocksDeclared)
	s.AttacksDeclared = g.attacksDeclared
	s.FirstStrikeStepParticipants = copyBoolMap(g.firstStrikeStepParticipants)
	cen := &s.Continuations

	s.Battlefield = snapshotZone(g.Battlefield, cen)
	s.Stack = snapshotZone(g.Stack, cen)
	s.Exile = snapshotZone(g.Exile, cen)
	s.PhasedOut = snapshotZone(g.PhasedOut, cen)

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
	// ADR 0041 phase 3 (#1497): data, so carried whole and never
	// counted by the census.
	s.ScopedEffects = deepCopyScopedEffects(g.ScopedEffects)

	s.LoyaltyActivatedThisTurn = copyBoolMap(g.LoyaltyActivatedThisTurn)
	s.SpellsCastThisTurn = copyTallyMap(g.SpellsCastThisTurn)
	s.ForetoldThisTurn = copyIntMap(g.ForetoldThisTurn)
	s.LandsPlayedThisTurn = copyIntMap(g.LandsPlayedThisTurn)
	s.ExtraLandDropsThisTurn = copyIntMap(g.ExtraLandDropsThisTurn)
	s.DrawnThisTurn = copyUUIDListMap(g.DrawnThisTurn)
	s.TurnTally = cloneTurnTally(g.TurnTally)
	s.Activations = cloneActivationTally(g.Activations)
	s.LoopNotice = cloneLoopNotice(g.LoopNotice)
	s.LoopThreshold = g.LoopThreshold
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
	if len(g.lastKnownTriggerIdentity) > 0 {
		s.LastKnownTriggerIdentity = make(map[uuid.UUID]triggerIdentityLKI, len(g.lastKnownTriggerIdentity))
		for k, v := range g.lastKnownTriggerIdentity {
			s.LastKnownTriggerIdentity[k] = v
		}
	}
	if len(g.lastKnownCounters) > 0 {
		s.LastKnownCounters = make(map[uuid.UUID]map[string]int, len(g.lastKnownCounters))
		for k, v := range g.lastKnownCounters {
			s.LastKnownCounters[k] = copyStringIntMap(v)
		}
	}
	s.LastKnownPermanents = cloneLastKnownPermanents(g.lastKnownPermanents)
	s.LastKnownStack = snapshotLastKnownStack(g.lastKnownStack, cen)

	// The turn-scoped block-rule registry retired with ADR 0041 tier
	// 3b, alongside its replacement twin: a Gingerbrute is a
	// ScopedEffect record now, carried above. Nothing increments
	// cen.TurnScopedBlockRules any more.
	// BuiltinReplacements and Listeners are deliberately NOT counted:
	// both are process-lifetime singletons installed by NewGame, so
	// the new binary rebuilds them itself. See restoreGame.

	s.RNG = snapshotRNG(g)
	s.SourceOrdinals = cloneSourceOrdinals(g.sourceOrdinals)
	s.SourceOrdinalNext = g.sourceOrdinalNext
	s.LayerVersion = g.layerVersion.Load()
	s.LastResolvedVersion = g.lastResolvedVersion.Load()
	return s
}

func snapshotRNG(g *Game) rngSnapshot {
	if g.rngKey == ([32]byte{}) {
		return rngSnapshot{Kind: rngKindNone}
	}
	return rngSnapshot{
		Kind:     rngKindKeyed,
		Key:      append([]byte(nil), g.rngKey[:]...),
		Counters: snapshotRNGCounters(g.rngCounters),
		Turn:     g.rngTurn,
	}
}

// snapshotRNGCounters copies the counters, writing an empty map as
// nil so a capture -> JSON -> restore -> capture round trip is exact
// (the JSON field is omitempty).
func snapshotRNGCounters(m map[string]uint64) map[string]uint64 {
	if len(m) == 0 {
		return nil
	}
	return cloneRNGCounters(m)
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

// abilityCatalogKey is the key a card's instance-carried ability
// closures are re-derived under, and it is deliberately ONE function
// so that the census (capture) and the rebuild (restore) can never
// disagree about what is recoverable. A counter that answered a
// different question from the code it guards would be worse than no
// counter at all.
//
// The oracle ID first, the token key second: CR 707.2 again — a token
// COPY carries the copied card's oracle ID and must resolve to that
// card, never to a token entry. The two are never both set on one
// object, so the order only ever settles the empty cases.
//
// Deliberately NOT CatalogKey: this is the key restore will have,
// off a decoded file, before the card is in a zone or a layer cache
// exists. CatalogKey's face suffix and CR 708.2a silence are
// questions about a LIVE object, and asking them here would make the
// census depend on state the restoring binary has not rebuilt yet.
func abilityCatalogKey(oracleID, tokenKey string) string {
	if oracleID != "" {
		return oracleID
	}
	return tokenKey
}

// restoreAbilityKey is abilityCatalogKey for a decoded snapshot card.
func restoreAbilityKey(c *cardSnapshot) string {
	return abilityCatalogKey(c.OracleID, c.TokenKey)
}

// intrinsicAbilitiesLost reports whether the ability closures carried
// on THIS object would be lost by a capture → restore round trip —
// the one question ContinuationCensus.IntrinsicAbilityCards is
// counting.
//
// It asks the registry rather than asking after an oracle ID (#521).
// The old predicate was `has instance abilities && no oracle ID`,
// which counted a Treasure token — four fields of plain data, every
// closure slot on its ManaAbilityShape nil — as an unrestorable
// continuation, and so refused to write a restore point for as long
// as one sat on the battlefield. A token template's abilities are now
// registered under its token key, so the honest question is whether
// the catalog can hand them back.
//
// The COUNT is what it compares, not mere presence, and that is the
// half that keeps the counter reachable: an ability stamped onto one
// instance at runtime, over and above whatever its entry declares,
// has no catalog entry of its own and is still lost. Restore rebuilds
// the catalog's list wholesale, so a card holding more abilities than
// its entry declares comes back with the extra one missing — which
// is exactly what this must keep counting.
func intrinsicAbilitiesLost(c Card) bool {
	mana, activated := len(c.ManaAbilities), len(c.ActivatedAbilities)
	if mana == 0 && activated == 0 {
		return false
	}
	key := abilityCatalogKey(c.OracleID, c.TokenKey)
	if key == "" {
		return true
	}
	if mana > 0 && (CatalogManaAbilities == nil || len(CatalogManaAbilities(key)) < mana) {
		return true
	}
	if activated > 0 && (CatalogActivatedAbilities == nil || len(CatalogActivatedAbilities(key)) < activated) {
		return true
	}
	return false
}

func snapshotCard(c Card, cen *ContinuationCensus) cardSnapshot {
	variableToughness := c.VariableToughness
	out := cardSnapshot{
		InstanceID:               c.InstanceID,
		Name:                     c.Name,
		ScryfallID:               c.ScryfallID,
		OracleID:                 c.OracleID,
		TokenKey:                 c.TokenKey,
		TypeLine:                 c.TypeLine,
		Power:                    c.Power,
		Toughness:                c.Toughness,
		VariableToughness:        &variableToughness,
		ManaCost:                 c.ManaCost,
		ProducedMana:             copyStrings(c.ProducedMana),
		Colors:                   copyStrings(c.Colors),
		ColorIdentity:            copyStrings(c.ColorIdentity),
		StartingLoyalty:          c.StartingLoyalty,
		Keywords:                 copyStrings(c.Keywords),
		GrantedAbilities:         copyStrings(c.GrantedAbilities),
		Layout:                   c.Layout,
		Faces:                    copyFaces(c.Faces),
		ActiveFace:               c.ActiveFace,
		PrintedSelf:              copyPrintedValues(c.PrintedSelf),
		NeedsEffect:              c.NeedsEffect,
		Owner:                    c.Owner,
		Controller:               c.Controller,
		Tapped:                   c.Tapped,
		NextUntapSkips:           snapshotUntapSkips(c.NextUntapSkips),
		BattleX:                  c.BattleX,
		BattleY:                  c.BattleY,
		Counters:                 copyStringIntMap(c.Counters),
		IsCommander:              c.IsCommander,
		AttackingTarget:          c.AttackingTarget,
		BlockingTarget:           c.BlockingTarget,
		GoadedBy:                 c.GoadedBy,
		DamageMarked:             c.DamageMarked,
		RegenerationShields:      c.RegenerationShields,
		FaceDown:                 c.FaceDown,
		FaceDownKind:             c.FaceDownKind,
		FaceDownListed:           c.FaceDownListed.clone(),
		FaceTurnedAt:             c.FaceTurnedAt,
		KnownBy:                  copyBoolMap(c.KnownBy),
		EnteredBattlefieldAt:     c.EnteredBattlefieldAt,
		ObjectEpoch:              c.ObjectEpoch,
		SummonedThisTurn:         c.SummonedThisTurn,
		MarkedLethalByDeathtouch: c.MarkedLethalByDeathtouch,
		LostLastCounter:          c.LostLastCounter,
		PrintedPTKnown:           c.PrintedPTKnown,
		AttachedTo:               c.AttachedTo,
		AttachedAt:               c.AttachedAt,
		BaseController:           c.BaseController,
		NamedTribe:               c.NamedTribe,
		Provenance:               c.Provenance.Clone(),
		ChosenColor:              c.ChosenColor,
		ChosenPlayer:             c.ChosenPlayer,
		ChosenName:               c.ChosenName,
		ClassLevel:               c.ClassLevel,
		Solved:                   c.Solved,
		Harnessed:                c.Harnessed,
		Prepared:                 c.Prepared,
		PrepareCopy:              c.PrepareCopy,
		PreparedBy:               c.PreparedBy,
		HiddenBy:                 c.HiddenBy,
		PhasedOutBy:              c.PhasedOutBy,
		PhaseInLockedBy:          c.PhaseInLockedBy,
		PhasedOutIndirect:        c.PhasedOutIndirect,
		TapOnPhaseIn:             c.TapOnPhaseIn,
		StartingDefense:          c.StartingDefense,
		ProtectorPlayerID:        c.ProtectorPlayerID,
		ManaAbilityCount:         len(c.ManaAbilities),
		ActivatedAbilityCount:    len(c.ActivatedAbilities),
		AbilitiesLostOnRestore:   c.AbilitiesLostOnRestore,
	}
	// #522: measure the catalog entry this card resolves to, so the
	// binary that restores the file can tell whether it still has it.
	if key := abilityCatalogKey(c.OracleID, c.TokenKey); IsAutoCard(key) {
		counts := catalogAbilityCounts(key)
		out.CatalogAbilities = &counts
	}
	// Intrinsic ability closures, censused only when the registry
	// cannot give them back (#521). See intrinsicAbilitiesLost.
	if intrinsicAbilitiesLost(c) {
		cen.IntrinsicAbilityCards++
		cen.note("intrinsic abilities the catalog cannot re-derive: %s", labelOr(c.Name, c.InstanceID.String()))
	}
	return out
}

func snapshotPlayer(p *Player, cen *ContinuationCensus) playerSnapshot {
	out := playerSnapshot{
		ID:                 p.ID,
		Name:               p.Name,
		Seat:               p.Seat,
		Life:               p.Life,
		Poison:             p.Poison,
		Energy:             p.Energy,
		Library:            snapshotZone(p.Library, cen),
		Hand:               snapshotZone(p.Hand, cen),
		Graveyard:          snapshotZone(p.Graveyard, cen),
		Command:            snapshotZone(p.Command, cen),
		Emblems:            snapshotZone(p.Emblems, cen),
		CommanderDamage:    copyIntMap(p.CommanderDamage),
		TurnsBegun:         p.TurnsBegun,
		Eliminated:         p.Eliminated,
		HandKept:           p.HandKept,
		MulligansTaken:     p.MulligansTaken,
		DeckImported:       p.DeckImported,
		UndosRemaining:     p.UndosRemaining,
		DiscordID:          p.DiscordID,
		DiscordAvatarHash:  p.DiscordAvatarHash,
		DisplayName:        p.DisplayName,
		IsBot:              p.IsBot,
		BotTier:            p.BotTier,
		BotDeck:            p.BotDeck,
		AttemptedEmptyDraw: p.AttemptedEmptyDraw,
		CommanderCasts:     copyIntMap(p.CommanderCasts),
		Counters:           copyStringIntMap(p.Counters),
		MaxHandSize:        p.MaxHandSize,
		LandDropsPerTurn:   p.LandDropsPerTurn,
	}
	if len(p.LifeHistory) > 0 {
		out.LifeHistory = make([]LifeChange, len(p.LifeHistory))
		copy(out.LifeHistory, p.LifeHistory)
	}
	if len(p.ManaPool) > 0 {
		out.ManaPool = make(ManaPool, len(p.ManaPool))
		for i, t := range p.ManaPool {
			out.ManaPool[i] = t.clone()
		}
	}
	out.CastPermissions = cloneCastPermissions(p.CastPermissions)
	out.Statics = clonePlayerStatics(p.Statics)
	return out
}

func snapshotStackItem(g *Game, s *StackItem, cen *ContinuationCensus) stackItemSnapshot {
	if s == nil {
		return stackItemSnapshot{}
	}
	return snapshotStackItemAs(s, oracleIDOfLocked(g, s.SourceCardID), cen)
}

// snapshotStackItemAs is snapshotStackItem with the source's oracle ID
// supplied: a spell in Game.lastKnownStack has left the stack, and the
// card it records is the one to read it from (a countered COPY is in
// no zone at all).
func snapshotStackItemAs(s *StackItem, oracleID string, cen *ContinuationCensus) stackItemSnapshot {
	out := stackItemSnapshot{
		ID:            s.ID,
		Kind:          s.Kind,
		Controller:    s.Controller,
		Owner:         s.Owner,
		SourceCardID:  s.SourceCardID,
		SourceEpoch:   s.SourceEpoch,
		SourceObject:  s.SourceObject.stamped(),
		Label:         s.Label,
		DoubledBy:     s.DoubledBy,
		DoubledByName: s.DoubledByName,
		Targets:       copyTargetRefs(s.Targets),
		Payload:       copyTargetRefs(s.Payload),
		Trigger:       cloneTriggerContext(s.Trigger),
		Modes:         copyInts(s.Modes),
		XValue:        s.XValue,
		Distribution:  copyIntMap(s.Distribution),
		HoldPriority:  s.HoldPriority,
		CastFromZone:  s.CastFromZone,
		AltCost:       s.AltCost,
		Foretold:      s.Foretold,
		FaceDown:      s.FaceDown,
		AltCostExiles: s.AltCostExiles,
		SplitSecond:   s.SplitSecond,
		IsCopy:        s.IsCopy,
		Uncopyable:    s.Uncopyable,
		Seq:           s.Seq,
		Ordered:       s.Ordered,
		Commutes:      s.Commutes,
		Paid:          clonePaidCost(s.Paid),
		HasEffect:     s.Effect != nil,
		HasTargetSpec: s.targetSpec != nil,
		HasModeSpec:   s.modeSpec != nil,
		OracleID:      oracleID,
		Body:          s.Body,
		Params:        effectParamsOrNil(s.Params),
	}
	// ADR 0041 P9 (#1497, tier 4): the census fold. An item is counted
	// ONCE, whatever it holds, because the question is one question —
	// can restore rebuild this item — and it has one of two answers:
	//
	//   - an Effect that is a bare closure cannot be rebuilt;
	//   - a target or mode clause can be rebuilt only when the item
	//     says where it came from: a SPELL by oracle ID, an ABILITY by
	//     its catalog row (Params.Ability, ability_ref.go, tier 4-1),
	//     or a CR 603.12 REFLEXIVE TRIGGER by its own Body registration
	//     (ReflexiveBody/reflexiveTargetSpecFor, tier 4-0).
	//
	// A KEYED item (a fired delayed trigger, ADR 0041 P2; a stamped
	// activated ability, or a reflexive trigger, P9) is data: restore
	// re-derives its Effect. StackTargetSpecs is retired: nothing
	// increments it.
	//
	// Tier 4-final (owner decision, 2026-09-25) retired StackEffects
	// too and folded what is left into IntrinsicAbilityCards: every
	// production path keys its item, and the one that cannot — an
	// ability carried on a card instance (Card.ActivatedAbilities),
	// which activatedAbilityRefFor never stamps — is that counter's own
	// subject. A hand-built item with a closure and no body (a test
	// fixture, a stubbed catalog row) lands here as well, so the census
	// still refuses to call a game holding one restorable.
	switch {
	case s.Effect != nil && s.Body == "":
		cen.IntrinsicAbilityCards++
		cen.note("stack effect: %s", labelOr(s.Label, string(s.Kind)))
	case (s.targetSpec != nil || s.modeSpec != nil) && !stackSpecsRederivable(out):
		cen.IntrinsicAbilityCards++
		cen.note("stack clause: %s", labelOr(s.Label, string(s.Kind)))
	}
	return out
}

// stackSpecsRederivable reports whether restore can rebuild this item's
// target and mode clauses: a spell from the catalog by oracle ID, a
// stamped ability from its catalog row, or a reflexive trigger from its
// own Body's registration (ADR 0041 P9, #1497, tier 4-0).
func stackSpecsRederivable(s stackItemSnapshot) bool {
	if spellSpecRederivable(s) {
		return true
	}
	if _, catalog := catalogBodySlot(s.Body); catalog && s.Params != nil && s.Params.Ability != nil {
		return true
	}
	if s.Body != "" && reflexiveTargetSpecFor(s.Body, s.SourceCardID, effectParamsValue(s.Params)) != nil {
		return true
	}
	return false
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
		SourceObject:       d.SourceObject.stamped(),
		Label:              d.Label,
		At:                 d.At,
		ControllerTurnOnly: d.ControllerTurnOnly,
		CreatedSeq:         d.CreatedSeq,
		HasEffect:          d.Body.key != "",
		Body:               d.Body.key,
		Params:             effectParamsOrNil(d.Params),
		Condition:          d.Condition.key,
		CondParams:         effectParamsOrNil(d.CondParams),
		OptionalQuestion:   d.OptionalQuestion,
	}
	if d.Duration != nil {
		dur := *d.Duration
		out.Duration = &dur
	}
	if len(d.On) > 0 {
		out.On = append([]EventKind(nil), d.On...)
	}
	if len(d.Cards) > 0 {
		out.Cards = append([]uuid.UUID(nil), d.Cards...)
	}
	// ADR 0041 P2 (#1497): a delayed trigger is data, so this counter
	// is RETIRED — ScheduleDelayedTriggerForEffect refuses a trigger
	// with no Body, and no other path builds one. The only way to meet
	// one here is a hand-built test fixture, and it is still counted,
	// because a trigger with nothing to do cannot be restored exactly.
	if d.Body.key == "" {
		cen.DelayedTriggerEffects++
		cen.note("delayed trigger without a body: %s", labelOr(d.Label, d.ID.String()))
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
		CoinAllowStop:        c.CoinAllowStop,
		CoinCount:            c.CoinCount,
		CoinMaxUsefulWins:    c.CoinMaxUsefulWins,
		CoinWins:             c.CoinWins,
		ColorOptions:         copyStrings(c.ColorOptions),
		ColorPurpose:         c.ColorPurpose,
		ManaRestrictions:     copyStrings(c.ManaRestrictions),
		ManaRiders:           copyManaRiders(c.ManaRiders),
		ManaSourceKinds:      c.ManaSourceKinds,
		ManaAmounts:          copyManaAmounts(c.ManaAmounts),
		ManaTapped:           c.ManaTapped,
		ReplacementEffectIDs: copyReplacementEffectIDs(c.ReplacementEffectIDs),
		NoLegalTarget:        c.NoLegalTarget,
		PickTargetPlayers:    copyUUIDs(c.PickTargetPlayers),
		PickTargetCards:      copyUUIDs(c.PickTargetCards),
		PickTargetMin:        c.PickTargetMin,
		PickTargetMax:        c.PickTargetMax,
		RetargetItem:         c.RetargetItem,
		RetargetPolicy:       c.RetargetPolicy,
		RetargetOptional:     c.RetargetOptional,
		RetargetSlot:         c.RetargetSlot,
		RetargetReason:       c.RetargetReason,
		ModeOptionIndex:      copyInts(c.ModeOptionIndex),
		ModeOptionLabel:      copyStrings(c.ModeOptionLabel),
		ModeMin:              c.ModeMin,
		ModeMax:              c.ModeMax,
		ModeRepeatable:       c.ModeRepeatable,
		SacrificeOptions:     copyUUIDs(c.SacrificeOptions),
		CopyOptions:          copyUUIDs(c.CopyOptions),
		ScryCards:            copyUUIDs(c.ScryCards),
		LibraryPlacement:     c.LibraryPlacement,
		LibraryTopCount:      c.LibraryTopCount,
		LibraryTopDepth:      c.LibraryTopDepth,
		TriggerOrderIDs:      copyUUIDs(c.TriggerOrderIDs),
		PayCost:              c.PayCost,
		SearchCards:          copyUUIDs(c.SearchCards),
		SearchMax:            c.SearchMax,
		MayCastCard:          c.MayCastCard,
		AcceptLabel:          c.AcceptLabel,
		DeclineLabel:         c.DeclineLabel,
		LifeCost:             c.LifeCost,
		OwedInStep:           c.OwedInStep,
		GuardsStackItem:      c.GuardsStackItem,
		ChooseCards:          copyUUIDs(c.ChooseCards),
		ChooseMin:            c.ChooseMin,
		ChooseMax:            c.ChooseMax,
		PickOptions:          cloneChoiceOptions(c.PickOptions),
		LoopShortcutKey:      c.LoopShortcutKey,
		LoopShortcutCount:    c.LoopShortcutCount,
		LoopShortcutRepeat:   c.LoopShortcutRepeat,
		MidResolution:        c.midResolution,
	}
	if c.DamageAssignment != nil {
		// Pure data (see the type), so a value copy with its own
		// slice is a complete copy.
		da := *c.DamageAssignment
		da.BlockerIDs = copyUUIDs(c.DamageAssignment.BlockerIDs)
		da.SourceLKI = copyCharacteristic(c.DamageAssignment.SourceLKI)
		out.DamageAssignment = &da
	}
	// The continuation slots, one by one. Each is a paused effect.
	for name, present := range map[string]bool{
		"replacementResume": c.replacementResume != nil,
		// #1397: a parked cost announcement waiting on a commander's
		// owner. Dropping it drops the announcement, not half of it —
		// nothing was paid — but the census still says so.
		"costCommanderResume": c.costCommanderResume != nil,
		"pickTargetResume":    c.pickTargetResume != nil,
		"copyResume":          c.copyResume != nil,
		"triggerResume":       c.triggerResume != nil,
		"payUnlessResume":     c.payUnlessResume != nil,
		"mayCastResume":       c.mayCastResume != nil,
		"searchResume":        c.searchResume != nil,
		"scryResume":          c.scryResume != nil,
		// ADR 0088's ordered placement: handed the answer, then
		// places the pile and runs the rest of the card.
		"libraryOrderResume": c.libraryOrderResume != nil,
		// The two chained-choice frames. A chain link is a
		// continuation like any other, and an entry missing here
		// would let the server write a restore point that silently
		// drops the rest of the card. See chained_choice.go.
		"confirmResume":     c.confirmResume != nil,
		"chooseCardsResume": c.chooseCardsResume != nil,
		// #742's resolution-time "choose a color".
		// #568's option pick, for the same reason.
		"optionPickResume":  c.optionPickResume != nil,
		"chooseColorResume": c.chooseColorResume != nil,
		"modePickResume":    c.modePickResume != nil,
		"coinFlipResume":    c.coinFlipResume != nil,
		// #1019's / #1027's prompted run — a sacrifice's or a
		// discard's. Not a frame ON the choice — the continuation
		// lives on the Game, keyed by this id, because the prompts of
		// one run share it — but it is a continuation the snapshot
		// cannot carry all the same, and a restore point taken
		// mid-fan-out would drop the rest of the card. Counted here so
		// the census says so.
		"promptRun": c.promptRun != uuid.Nil,
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
	if err := s.checkEffectKeys(); err != nil {
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
	if err := s.checkEffectKeys(); err != nil {
		return nil, err
	}
	if !s.Restorable() {
		return nil, fmt.Errorf("%w: %d continuation(s): %v",
			ErrSnapshotNotRestorable, s.Continuations.Total(), s.Continuations.Labels)
	}
	return s.restoreGame(), nil
}

// migrateLegacySettings builds the settings of a pre-v4 file, which
// knew only an undo limit (#1032). Every other setting takes its
// default, which is what those games were played under.
//
// The undo limit is kept when the v3 binary would have honoured it:
// once a game was active its Start had already rewritten any value
// <= 0 to the default, so an active or ended game's 0 was set on
// purpose with set_undo_limit and survives. A LOBBY game's 0 (or less)
// is the unset field Start was about to rewrite, and becomes the
// default here instead.
func migrateLegacySettings(state State, undoLimit int) TableSettings {
	out := DefaultTableSettings()
	switch {
	case undoLimit > 0:
		out.UndoLimit = undoLimit
	case state != StateLobby:
		out.UndoLimit = 0
	}
	return out
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
	legacyTurnIdentity := g.Turn.Seq == 0 && !s.turnSeqPresent
	if legacyTurnIdentity {
		// ADR 0059 deliberately keeps the legacy snapshot key "Number"
		// for Round. Old files have no Seq, so rebuild the monotone turn
		// identity from the RNG index they already used.
		g.Turn.Seq = g.Turn.Round*MaxPlayers + g.Turn.ActiveSeat
		g.Turn.OrderSeat = g.Turn.ActiveSeat
	}
	g.MulligansOpen = s.MulligansOpen
	g.Monarch = s.Monarch
	g.Initiative = s.Initiative
	g.Settings = s.Settings
	if s.Schema < settingsSchemaVersion {
		g.Settings = migrateLegacySettings(s.State, s.UndoLimit)
	}
	g.StartingSeat = s.StartingSeat
	g.SplitSecondActive = s.SplitSecondActive
	g.Outcome = cloneGameOutcome(s.Outcome)
	g.ActiveSeatLeftPending = s.ActiveSeatLeftPending
	g.eventSeq = s.EventSeq
	g.eventBatch = s.EventBatch
	g.resolutionOpen = s.ResolutionOpen
	g.oncePerBatchFired = copyStringUint64Map(s.OncePerBatchFired)
	g.announcedBlocks = copyUUIDPairMap(s.AnnouncedBlocks)
	g.blockedAttackers = copyBoolMap(s.BlockedAttackers)
	g.announcedAttacks = copyBoolMap(s.AnnouncedAttacks)
	g.attackDefenders = copyUUIDPairMap(s.AttackDefenders)
	g.blocksDeclared = copyBoolMap(s.BlocksDeclared)
	g.attacksDeclared = s.AttacksDeclared
	g.firstStrikeStepParticipants = copyBoolMap(s.FirstStrikeStepParticipants)

	g.Battlefield = restoreZone(s.Battlefield, ZoneBattlefield)
	g.Stack = restoreZone(s.Stack, ZoneStack)
	g.Exile = restoreZone(s.Exile, ZoneExile)
	g.PhasedOut = restoreZone(s.PhasedOut, ZonePhasedOut)

	g.Seats = make([]*Player, len(s.Seats))
	for i := range s.Seats {
		g.Seats[i] = restorePlayer(&s.Seats[i])
	}
	if legacyTurnIdentity {
		for seat, p := range g.Seats {
			if p == nil {
				continue
			}
			p.TurnsBegun = g.Turn.Round
			if seat > g.Turn.ActiveSeat {
				p.TurnsBegun--
			}
			if p.TurnsBegun < 0 {
				p.TurnsBegun = 0
			}
			for i := range p.CastPermissions {
				perm := &p.CastPermissions[i]
				if perm.NotBeforeSeq == 0 && perm.LegacyNotBeforeTurn > g.Turn.Round {
					perm.NotBeforeSeq = perm.LegacyNotBeforeTurn * MaxPlayers
				}
				perm.LegacyNotBeforeTurn = 0
			}
		}
	}

	// ADR 0041 P9 / Q3: the sources of stack abilities this binary's
	// catalog no longer has, flagged once every zone is restored.
	var lostSources []uuid.UUID
	if len(s.StackMeta) > 0 {
		g.StackMeta = make(map[uuid.UUID]*StackItem, len(s.StackMeta))
		for i := range s.StackMeta {
			it, ok := restoreStackItem(&s.StackMeta[i])
			if !ok {
				lostSources = append(lostSources, it.SourceCardID)
			}
			g.StackMeta[it.ID] = it
		}
	}
	if len(s.PendingTriggers) > 0 {
		g.PendingTriggers = make([]*StackItem, len(s.PendingTriggers))
		for i := range s.PendingTriggers {
			it, ok := restoreStackItem(&s.PendingTriggers[i])
			if !ok {
				lostSources = append(lostSources, it.SourceCardID)
			}
			g.PendingTriggers[i] = it
		}
	}
	for _, id := range lostSources {
		g.flagAbilitiesLostLocked(id)
	}
	// Tier 4-2: a stamped trigger whose row builds its clause from the
	// trigger context gets it back now that every zone is restored.
	for i := range s.StackMeta {
		g.rederiveTriggerClauseLocked(g.StackMeta[s.StackMeta[i].ID], &s.StackMeta[i])
	}
	for i := range s.PendingTriggers {
		g.rederiveTriggerClauseLocked(g.PendingTriggers[i], &s.PendingTriggers[i])
	}
	g.ScopedEffects = deepCopyScopedEffects(s.ScopedEffects)
	g.scopedEffectSeq = maxScopedEffectSeq(g.ScopedEffects)
	if len(s.DelayedTriggers) > 0 {
		g.DelayedTriggers = make([]*DelayedTrigger, len(s.DelayedTriggers))
		for i := range s.DelayedTriggers {
			g.DelayedTriggers[i] = restoreDelayedTrigger(&s.DelayedTriggers[i])
			if legacyTurnIdentity && g.DelayedTriggers[i].CreatedSeq == 0 {
				g.DelayedTriggers[i].CreatedSeq = s.DelayedTriggers[i].CreatedTurn * MaxPlayers
			}
		}
	}

	g.LoyaltyActivatedThisTurn = copyBoolMap(s.LoyaltyActivatedThisTurn)
	g.SpellsCastThisTurn = copyTallyMap(s.SpellsCastThisTurn)
	g.ForetoldThisTurn = copyIntMap(s.ForetoldThisTurn)
	g.TurnTally = cloneTurnTally(s.TurnTally)
	g.Activations = cloneActivationTally(s.Activations)
	g.LoopNotice = cloneLoopNotice(s.LoopNotice)
	g.LoopThreshold = s.LoopThreshold
	g.LandsPlayedThisTurn = copyIntMap(s.LandsPlayedThisTurn)
	g.ExtraLandDropsThisTurn = copyIntMap(s.ExtraLandDropsThisTurn)
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
	if len(s.LastKnownTriggerIdentity) > 0 {
		g.lastKnownTriggerIdentity = make(map[uuid.UUID]triggerIdentityLKI, len(s.LastKnownTriggerIdentity))
		for k, v := range s.LastKnownTriggerIdentity {
			g.lastKnownTriggerIdentity[k] = v
		}
	}
	if len(s.LastKnownCounters) > 0 {
		g.lastKnownCounters = make(map[uuid.UUID]map[string]int, len(s.LastKnownCounters))
		for k, v := range s.LastKnownCounters {
			g.lastKnownCounters[k] = copyStringIntMap(v)
		}
	}
	g.lastKnownPermanents = cloneLastKnownPermanents(s.LastKnownPermanents)
	g.lastKnownStack = restoreLastKnownStack(s.LastKnownStack)

	restoreRNG(g, s.RNG)
	g.sourceOrdinals = cloneSourceOrdinals(s.SourceOrdinals)
	g.sourceOrdinalNext = s.SourceOrdinalNext

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
	case rngKindKeyed:
		if len(s.Key) == 32 {
			var key [32]byte
			copy(key[:], s.Key)
			if key != ([32]byte{}) {
				g.rngKey = key
				g.rngCounters = cloneRNGCounters(s.Counters)
				g.rngTurn = s.Turn
				return
			}
		}
		// A damaged key is not a reason to hand the table a game
		// that cannot shuffle. Mint a fresh one: the restored game
		// is still fair, it just does not continue the exact
		// streams.
		g.rngKey = mintRNGKey()
		g.rngCounters = nil
	case rngKindPCG, rngKindExternal:
		// A file from before keyed streams. Its PCG position (or the
		// lack of one) cannot seed the keyed derivation, and nothing
		// already decided depends on it: library orders are stored
		// in the zones. A fresh key changes only future draws.
		g.rngKey = mintRNGKey()
		g.rngCounters = nil
	default:
		// Nothing drawn yet; Start (or the first draw) mints the key.
	}
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
		TokenKey:                 c.TokenKey,
		TypeLine:                 c.TypeLine,
		Power:                    c.Power,
		Toughness:                c.Toughness,
		ManaCost:                 c.ManaCost,
		ProducedMana:             copyStrings(c.ProducedMana),
		Colors:                   copyStrings(c.Colors),
		ColorIdentity:            copyStrings(c.ColorIdentity),
		StartingLoyalty:          c.StartingLoyalty,
		Keywords:                 copyStrings(c.Keywords),
		GrantedAbilities:         copyStrings(c.GrantedAbilities),
		Layout:                   c.Layout,
		Faces:                    copyFaces(c.Faces),
		ActiveFace:               c.ActiveFace,
		PrintedSelf:              copyPrintedValues(c.PrintedSelf),
		NeedsEffect:              c.NeedsEffect,
		Owner:                    c.Owner,
		Controller:               c.Controller,
		Tapped:                   c.Tapped,
		NextUntapSkips:           restoreUntapSkips(c.NextUntapSkips),
		BattleX:                  c.BattleX,
		BattleY:                  c.BattleY,
		Counters:                 copyStringIntMap(c.Counters),
		IsCommander:              c.IsCommander,
		AttackingTarget:          c.AttackingTarget,
		BlockingTarget:           c.BlockingTarget,
		GoadedBy:                 c.GoadedBy,
		DamageMarked:             c.DamageMarked,
		RegenerationShields:      c.RegenerationShields,
		FaceDown:                 c.FaceDown,
		FaceDownKind:             c.FaceDownKind,
		FaceDownListed:           c.FaceDownListed.clone(),
		FaceTurnedAt:             c.FaceTurnedAt,
		KnownBy:                  copyBoolMap(c.KnownBy),
		EnteredBattlefieldAt:     c.EnteredBattlefieldAt,
		ObjectEpoch:              c.ObjectEpoch,
		SummonedThisTurn:         c.SummonedThisTurn,
		MarkedLethalByDeathtouch: c.MarkedLethalByDeathtouch,
		LostLastCounter:          c.LostLastCounter,
		PrintedPTKnown:           c.PrintedPTKnown,
		AttachedTo:               c.AttachedTo,
		AttachedAt:               c.AttachedAt,
		BaseController:           c.BaseController,
		NamedTribe:               c.NamedTribe,
		Provenance:               c.Provenance.Clone(),
		ChosenColor:              c.ChosenColor,
		ChosenPlayer:             c.ChosenPlayer,
		ChosenName:               c.ChosenName,
		ClassLevel:               c.ClassLevel,
		Solved:                   c.Solved,
		Harnessed:                c.Harnessed,
		Prepared:                 c.Prepared,
		PrepareCopy:              c.PrepareCopy,
		PreparedBy:               c.PreparedBy,
		HiddenBy:                 c.HiddenBy,
		PhasedOutBy:              c.PhasedOutBy,
		PhaseInLockedBy:          c.PhaseInLockedBy,
		PhasedOutIndirect:        c.PhasedOutIndirect,
		TapOnPhaseIn:             c.TapOnPhaseIn,
		StartingDefense:          c.StartingDefense,
		ProtectorPlayerID:        c.ProtectorPlayerID,
		AbilitiesLostOnRestore:   c.AbilitiesLostOnRestore,
	}
	// #522: the parity check. A card this binary's catalog hands back
	// fewer abilities for than the file recorded is restored anyway
	// and flagged, for the rest of the game, as not automated. The
	// report half is GameSnapshot.AbilityShortfalls, through the same
	// function.
	if _, short := abilityShortfallOf(c); short {
		out.AbilitiesLostOnRestore = true
	}
	if c.VariableToughness != nil {
		out.VariableToughness = *c.VariableToughness
	} else {
		backfillVariableToughness(&out)
	}
	// Re-derive intrinsic abilities from the catalog. This is the
	// half of the closure problem that DOES have an answer: the
	// closures came out of the catalog by oracle ID, and the catalog
	// is code the new binary already has. A token copy carries the
	// copied card's oracle ID (CR 707.2), so it comes back whole.
	//
	// A TOKEN takes the same route since #521, under the synthetic
	// key its template registered (game/token_key.go) — which is the
	// whole of why a Treasure no longer blocks a restore point.
	//
	// An object with NEITHER identity cannot be looked up;
	// snapshotCard already counted it in the census, so a strict
	// restore never reaches this line with abilities to rebuild.
	if key := restoreAbilityKey(c); key != "" {
		if c.ManaAbilityCount > 0 && CatalogManaAbilities != nil {
			out.ManaAbilities = CatalogManaAbilities(key)
		}
		if c.ActivatedAbilityCount > 0 && CatalogActivatedAbilities != nil {
			out.ActivatedAbilities = CatalogActivatedAbilities(key)
		}
	}
	// ADR 0069: a snapshot written before FaceDownKind existed carries
	// no kind, and a face-down card with no kind has no rule attached
	// — no viewers row, no CR 708.2 answer. The only face-down object
	// that could exist in such a file is a Necropotence exile, so it
	// restores as one and comes back with the visibility it had
	// (nobody may look) rather than as an unclassified flag. No schema
	// bump: this is the field zero-valuing correctly, which
	// SnapshotSchemaVersion's policy says does not need one.
	if out.FaceDown && out.FaceDownKind == FaceDownNone {
		out.FaceDownKind = FaceDownExiled
	}
	// `effective` is intentionally left nil: it is a derived cache,
	// and restoreGame bumps layerVersion so the next read recomputes.
	return out
}

func restorePlayer(p *playerSnapshot) *Player {
	out := &Player{
		ID:                 p.ID,
		Name:               p.Name,
		Seat:               p.Seat,
		Life:               p.Life,
		Poison:             p.Poison,
		Energy:             p.Energy,
		Library:            restoreZone(p.Library, ZoneLibrary),
		Hand:               restoreZone(p.Hand, ZoneHand),
		Graveyard:          restoreZone(p.Graveyard, ZoneGraveyard),
		Command:            restoreZone(p.Command, ZoneCommand),
		Emblems:            restoreZone(p.Emblems, ZoneCommand),
		TurnsBegun:         p.TurnsBegun,
		Eliminated:         p.Eliminated,
		HandKept:           p.HandKept,
		MulligansTaken:     p.MulligansTaken,
		DeckImported:       p.DeckImported,
		UndosRemaining:     p.UndosRemaining,
		DiscordID:          p.DiscordID,
		DiscordAvatarHash:  p.DiscordAvatarHash,
		DisplayName:        p.DisplayName,
		IsBot:              p.IsBot,
		BotTier:            p.BotTier,
		BotDeck:            p.BotDeck,
		AttemptedEmptyDraw: p.AttemptedEmptyDraw,
		Counters:           copyStringIntMap(p.Counters),
		MaxHandSize:        p.MaxHandSize,
		LandDropsPerTurn:   p.LandDropsPerTurn,
	}
	// #500: a snapshot written before the field existed carries no
	// value for it, and restoring 0 would seat a player who may never
	// play a land again. Nothing in the catalog sets an allowance of
	// zero, so a non-positive restored value is an old snapshot and
	// means the default.
	if out.LandDropsPerTurn <= 0 {
		out.LandDropsPerTurn = DefaultLandDropsPerTurn
	}
	// #623: a pre-emblem file carries no emblem zone, so restoreZone
	// hands back one owned by nobody. newPlayer stamps the owner, so
	// stamp it here too and a restored seat is the shape a fresh one
	// is.
	if out.Emblems != nil {
		out.Emblems.Owner = out.ID
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
			out.ManaPool[i] = t.clone()
		}
	}
	out.CastPermissions = cloneCastPermissions(p.CastPermissions)
	out.Statics = clonePlayerStatics(p.Statics)
	return out
}

// restoreStackItem rebuilds one stack item. It reports false when the
// item was stamped with a catalog row this binary cannot find (ADR 0041
// P9, the owner's Q3): the item is restored as a manual one and the
// caller flags its source.
func restoreStackItem(s *stackItemSnapshot) (*StackItem, bool) {
	out := &StackItem{
		ID:            s.ID,
		Kind:          s.Kind,
		Controller:    s.Controller,
		Owner:         s.Owner,
		SourceCardID:  s.SourceCardID,
		SourceEpoch:   s.SourceEpoch,
		SourceObject:  s.SourceObject.value(),
		Label:         s.Label,
		DoubledBy:     s.DoubledBy,
		DoubledByName: s.DoubledByName,
		Targets:       copyTargetRefs(s.Targets),
		Payload:       copyTargetRefs(s.Payload),
		Trigger:       cloneTriggerContext(s.Trigger),
		Modes:         copyInts(s.Modes),
		XValue:        s.XValue,
		Distribution:  copyIntMap(s.Distribution),
		HoldPriority:  s.HoldPriority,
		CastFromZone:  s.CastFromZone,
		AltCost:       s.AltCost,
		Foretold:      s.Foretold,
		FaceDown:      s.FaceDown,
		AltCostExiles: s.AltCostExiles,
		SplitSecond:   s.SplitSecond,
		IsCopy:        s.IsCopy,
		Uncopyable:    s.Uncopyable,
		Seq:           s.Seq,
		Ordered:       s.Ordered,
		Commutes:      s.Commutes,
		Paid:          clonePaidCost(s.Paid),
		Body:          s.Body,
		Params:        effectParamsValue(s.Params),
		// Effect stays nil unless the item is keyed. A SPELL does not
		// need one — resolution dispatches through EffectResolver by
		// oracle ID — but an ability does, which is why a stack item
		// with an unkeyed Effect is counted in the census and keeps
		// the snapshot from being a restore point.
	}
	if _, ok := catalogBodySlot(s.Body); ok {
		// A stamped activated or triggered ability (P9): the row gives
		// back the Effect, the target clause and the mode clause
		// together, with the owner's Q2 name check. No row is Q3.
		return out, restoreCatalogAbility(out, s)
	}
	if s.Body != "" {
		out.Effect = bodyEffect(s.Body, out.Params)
	}
	if s.HasTargetSpec && spellSpecRederivable(*s) {
		// The clause the spell was ANNOUNCED under, not the printed
		// one: an alternative cost (cleave) or a paid optional cost
		// (a promised gift, ADR 0089 §3) may have rewritten it, and
		// both are recorded on the item. castTargetSpecForItem is the
		// one function that applies both rewrites, in the order the
		// announce path does.
		out.targetSpec = castTargetSpecForItem(s.OracleID, out)
	} else if s.HasTargetSpec && s.Body != "" {
		// A CR 603.12 reflexive trigger (ADR 0041 P9, #1497, tier 4):
		// the clause was declared alongside the body's registration
		// (ReflexiveBody), keyed the same way, so it is re-derived from
		// there rather than from a catalog row.
		out.targetSpec = reflexiveTargetSpecFor(s.Body, s.SourceCardID, out.Params)
	}
	if s.HasModeSpec && s.Kind == StackItemSpell && s.OracleID != "" && CatalogModeSpec != nil {
		out.modeSpec = CatalogModeSpec(s.OracleID)
	}
	return out, true
}

func restoreDelayedTrigger(d *delayedTriggerSnapshot) *DelayedTrigger {
	out := &DelayedTrigger{
		ID:                 d.ID,
		Controller:         d.Controller,
		SourceCardID:       d.SourceCardID,
		SourceObject:       d.SourceObject.value(),
		Label:              d.Label,
		At:                 d.At,
		ControllerTurnOnly: d.ControllerTurnOnly,
		CreatedSeq:         d.CreatedSeq,
		// checkEffectKeys has already refused a key this binary has
		// no body or condition for, so these refs are registered ones.
		Body:             BodyRef{key: d.Body},
		Params:           effectParamsValue(d.Params),
		Condition:        ConditionRef{key: d.Condition},
		CondParams:       effectParamsValue(d.CondParams),
		OptionalQuestion: d.OptionalQuestion,
	}
	if d.Duration != nil {
		dur := *d.Duration
		out.Duration = &dur
	}
	if len(d.On) > 0 {
		out.On = append([]EventKind(nil), d.On...)
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
		CoinAllowStop:        c.CoinAllowStop,
		CoinCount:            c.CoinCount,
		CoinMaxUsefulWins:    c.CoinMaxUsefulWins,
		CoinWins:             c.CoinWins,
		ColorOptions:         copyStrings(c.ColorOptions),
		ColorPurpose:         c.ColorPurpose,
		ManaRestrictions:     copyStrings(c.ManaRestrictions),
		ManaRiders:           copyManaRiders(c.ManaRiders),
		ManaSourceKinds:      c.ManaSourceKinds,
		ManaAmounts:          copyManaAmounts(c.ManaAmounts),
		ManaTapped:           c.ManaTapped,
		ReplacementEffectIDs: copyReplacementEffectIDs(c.ReplacementEffectIDs),
		NoLegalTarget:        c.NoLegalTarget,
		PickTargetPlayers:    copyUUIDs(c.PickTargetPlayers),
		PickTargetCards:      copyUUIDs(c.PickTargetCards),
		PickTargetMin:        c.PickTargetMin,
		PickTargetMax:        c.PickTargetMax,
		RetargetItem:         c.RetargetItem,
		RetargetPolicy:       c.RetargetPolicy,
		RetargetOptional:     c.RetargetOptional,
		RetargetSlot:         c.RetargetSlot,
		RetargetReason:       c.RetargetReason,
		ModeOptionIndex:      copyInts(c.ModeOptionIndex),
		ModeOptionLabel:      copyStrings(c.ModeOptionLabel),
		ModeMin:              c.ModeMin,
		ModeMax:              c.ModeMax,
		ModeRepeatable:       c.ModeRepeatable,
		SacrificeOptions:     copyUUIDs(c.SacrificeOptions),
		CopyOptions:          copyUUIDs(c.CopyOptions),
		ScryCards:            copyUUIDs(c.ScryCards),
		LibraryPlacement:     c.LibraryPlacement,
		LibraryTopCount:      c.LibraryTopCount,
		LibraryTopDepth:      c.LibraryTopDepth,
		TriggerOrderIDs:      copyUUIDs(c.TriggerOrderIDs),
		PayCost:              c.PayCost,
		SearchCards:          copyUUIDs(c.SearchCards),
		SearchMax:            c.SearchMax,
		MayCastCard:          c.MayCastCard,
		AcceptLabel:          c.AcceptLabel,
		DeclineLabel:         c.DeclineLabel,
		LifeCost:             c.LifeCost,
		OwedInStep:           c.OwedInStep,
		GuardsStackItem:      c.GuardsStackItem,
		ChooseCards:          copyUUIDs(c.ChooseCards),
		ChooseMin:            c.ChooseMin,
		ChooseMax:            c.ChooseMax,
		PickOptions:          cloneChoiceOptions(c.PickOptions),
		LoopShortcutKey:      c.LoopShortcutKey,
		LoopShortcutCount:    c.LoopShortcutCount,
		LoopShortcutRepeat:   c.LoopShortcutRepeat,
		midResolution:        c.MidResolution,
		// Every resume frame stays nil. This is the phase-1 line in
		// the sand, and the census is how it is enforced rather than
		// hoped for.
	}
	if c.DamageAssignment != nil {
		da := *c.DamageAssignment
		da.BlockerIDs = copyUUIDs(c.DamageAssignment.BlockerIDs)
		da.SourceLKI = copyCharacteristic(c.DamageAssignment.SourceLKI)
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

// copyManaAmounts deep-copies PendingChoice.ManaAmounts (#742).
func copyManaAmounts(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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

// copyCharacteristic deep-copies an optional characteristics snapshot
// — the CR 702.16e last-known information a damage-assignment frame
// carries (#662). Value copy plus its own slices, so a restored frame
// shares nothing with the live one.
func copyCharacteristic(in *Characteristic) *Characteristic {
	if in == nil {
		return nil
	}
	out := *in
	out.Types = copyStrings(in.Types)
	out.Subtypes = copyStrings(in.Subtypes)
	out.Supertypes = copyStrings(in.Supertypes)
	out.Colors = copyStrings(in.Colors)
	out.Abilities = copyStrings(in.Abilities)
	if len(in.GrantedAbilities) > 0 {
		out.GrantedAbilities = append([]GrantedAbility(nil), in.GrantedAbilities...)
	}
	return &out
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

// copyUUIDPairMap is copyBoolMap for a card-to-card map —
// Game.announcedBlocks, blocker to the attacker its EventBlock named
// (#830).
func copyUUIDPairMap(in map[uuid.UUID]uuid.UUID) map[uuid.UUID]uuid.UUID {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]uuid.UUID, len(in))
	for k, v := range in {
		out[k] = v
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

// copyStringUint64Map is copyStringIntMap for Game.oncePerBatchFired,
// whose values are batch ids (#829).
func copyStringUint64Map(in map[string]uint64) map[string]uint64 {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]uint64, len(in))
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
