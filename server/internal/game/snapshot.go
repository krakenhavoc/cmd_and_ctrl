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
// Restore REFUSES anything it does not recognise rather than guessing.
// See ErrSchemaTooNew / ErrSchemaUnsupported and ADR 0041 for the
// version-skew policy this implements.
const SnapshotSchemaVersion = 5

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

	// StackMeta is a SLICE, not a map: map iteration order is
	// unspecified and the stack is ordered by Seq anyway. Sorted on
	// capture so two snapshots of the same state are byte-identical.
	StackMeta       []stackItemSnapshot      `json:"stackMeta,omitempty"`
	PendingTriggers []stackItemSnapshot      `json:"pendingTriggers,omitempty"`
	DelayedTriggers []delayedTriggerSnapshot `json:"delayedTriggers,omitempty"`

	LoyaltyActivatedThisTurn map[uuid.UUID]bool        `json:"loyaltyActivatedThisTurn,omitempty"`
	SpellsCastThisTurn       map[uuid.UUID]CastTally   `json:"spellsCastThisTurn,omitempty"`
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

	RNG               rngSnapshot          `json:"rng"`
	SourceOrdinals    map[uuid.UUID]uint64 `json:"sourceOrdinals,omitempty"`
	SourceOrdinalNext uint64               `json:"sourceOrdinalNext,omitempty"`

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
	Provenance CastProvenance `json:"provenance,omitzero"`
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
	ClassLevel        int       `json:"classLevel,omitempty"`
	Solved            bool      `json:"solved,omitempty"`
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
	ID            uuid.UUID         `json:"id"`
	Kind          StackItemKind     `json:"kind"`
	Controller    uuid.UUID         `json:"controller"`
	Owner         uuid.UUID         `json:"owner"`
	SourceCardID  uuid.UUID         `json:"sourceCardId"`
	SourceEpoch   int               `json:"sourceEpoch,omitempty"`
	Label         string            `json:"label,omitempty"`
	DoubledBy     uuid.UUID         `json:"doubledBy,omitempty"`
	DoubledByName string            `json:"doubledByName,omitempty"`
	Targets       []TargetRef       `json:"targets,omitempty"`
	Payload       []TargetRef       `json:"payload,omitempty"`
	Modes         []int             `json:"modes,omitempty"`
	XValue        int               `json:"xValue"`
	Distribution  map[uuid.UUID]int `json:"distribution,omitempty"`
	HoldPriority  bool              `json:"holdPriority"`
	CastFromZone  ZoneKind          `json:"castFromZone,omitempty"`
	AltCost       string            `json:"altCost,omitempty"`
	Foretold      bool              `json:"foretold,omitempty"`
	// FaceDown is CR 708.4 (#1194): the state the permanent this
	// spell becomes enters in. Carried, and it has to be — a restore
	// that lost it would resolve a morph on the stack into a face-UP
	// creature, revealing the card and giving it every ability
	// CR 708.2a says it does not have.
	FaceDown      FaceDownKind `json:"faceDown,omitempty"`
	AltCostExiles bool         `json:"altCostExiles,omitempty"`
	SplitSecond   bool         `json:"splitSecond"`
	IsCopy        bool         `json:"isCopy,omitempty"`
	Seq           uint64       `json:"seq"`
	Ordered       bool         `json:"ordered"`

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
	// #663: the event condition. On and ExpiresAfterTurn are data
	// and come back; the AppliesTo predicate and the Optional prompt
	// are closures and do not, exactly as Effect does not — and the
	// trigger is already counted once in
	// ContinuationCensus.DelayedTriggerEffects, so it is already
	// marked unrestorable and nothing here double-counts it.
	On       []EventKind `json:"on,omitempty"`
	Duration *Duration   `json:"duration,omitempty"`
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
	TriggerOrderIDs  []uuid.UUID    `json:"triggerOrderIds,omitempty"`
	PayCost          string         `json:"payCost,omitempty"`
	SearchCards      []uuid.UUID    `json:"searchCards,omitempty"`
	SearchMax        int            `json:"searchMax"`
	MayCastCard      uuid.UUID      `json:"mayCastCard,omitempty"`
	AcceptLabel      string         `json:"acceptLabel,omitempty"`
	LifeCost         int            `json:"lifeCost,omitempty"`
	DeclineLabel     string         `json:"declineLabel,omitempty"`
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

	// ScopedStatics is floating continuous effects with a duration
	// (Giant Growth's +3/+3, Act of Treason's theft). The wire key
	// stays `turnScopedStatics`, the name it had before the registry
	// grew the other CR 611.2 durations, so a census written by an
	// older binary still decodes.
	ScopedStatics int `json:"turnScopedStatics,omitempty"`

	// TurnScopedReplacements is floating until-end-of-turn
	// replacement effects (Fog).
	TurnScopedReplacements int `json:"turnScopedReplacements,omitempty"`

	// TurnScopedBlockRules is floating until-end-of-turn block rules
	// (Gingerbrute's "can't block this turn"). #750.
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
		Settings:          g.Settings,
		StartingSeat:      g.StartingSeat,
		SplitSecondActive: g.SplitSecondActive,
		EventSeq:          g.eventSeq,
		EventBatch:        g.eventBatch,
	}
	s.OncePerBatchFired = copyStringUint64Map(g.oncePerBatchFired)
	s.AnnouncedBlocks = copyUUIDPairMap(g.announcedBlocks)
	s.BlockedAttackers = copyBoolMap(g.blockedAttackers)
	s.AnnouncedAttacks = copyBoolMap(g.announcedAttacks)
	s.FirstStrikeStepParticipants = copyBoolMap(g.firstStrikeStepParticipants)
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

	// Turn-scoped registries: entirely closure-bearing, so only the
	// census and the labels survive. Dropping a Fog silently would be
	// worse than refusing the restore point, which is what this does.
	for _, st := range g.ScopedStatics {
		cen.ScopedStatics++
		cen.note("scoped static (%s): %s", st.Duration.Kind,
			labelOr(st.Label, st.Source.Name))
	}
	for _, re := range g.TurnScopedReplacements {
		cen.TurnScopedReplacements++
		cen.note("turn-scoped replacement: %s", labelOr(re.Label, "unnamed"))
	}
	for _, br := range g.TurnScopedBlockRules {
		cen.TurnScopedBlockRules++
		cen.note("turn-scoped block rule: %s", labelOr(br.Label, "unnamed"))
	}
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
		StartingDefense:          c.StartingDefense,
		ProtectorPlayerID:        c.ProtectorPlayerID,
		ManaAbilityCount:         len(c.ManaAbilities),
		ActivatedAbilityCount:    len(c.ActivatedAbilities),
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
			cloned := t
			cloned.Restrictions = copyStrings(t.Restrictions)
			out.ManaPool[i] = cloned
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
	out := stackItemSnapshot{
		ID:            s.ID,
		Kind:          s.Kind,
		Controller:    s.Controller,
		Owner:         s.Owner,
		SourceCardID:  s.SourceCardID,
		SourceEpoch:   s.SourceEpoch,
		Label:         s.Label,
		DoubledBy:     s.DoubledBy,
		DoubledByName: s.DoubledByName,
		Targets:       copyTargetRefs(s.Targets),
		Payload:       copyTargetRefs(s.Payload),
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
		Seq:           s.Seq,
		Ordered:       s.Ordered,
		Paid:          clonePaidCost(s.Paid),
		HasEffect:     s.Effect != nil,
		HasTargetSpec: s.targetSpec != nil,
		HasModeSpec:   s.modeSpec != nil,
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
	// #764: an ABILITY's ModeSpec came off the ability declaration,
	// reached by an index the item does not record, so it cannot be
	// recovered here either. A spell's is looked up by oracle ID.
	if s.modeSpec != nil && !spellSpecRederivable(out) {
		cen.StackTargetSpecs++
		cen.note("stack mode spec: %s", labelOr(s.Label, string(s.Kind)))
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
		CoinAllowStop:        c.CoinAllowStop,
		CoinCount:            c.CoinCount,
		CoinMaxUsefulWins:    c.CoinMaxUsefulWins,
		CoinWins:             c.CoinWins,
		ColorOptions:         copyStrings(c.ColorOptions),
		ColorPurpose:         c.ColorPurpose,
		ManaRestrictions:     copyStrings(c.ManaRestrictions),
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
	g.MulligansOpen = s.MulligansOpen
	g.Monarch = s.Monarch
	g.Initiative = s.Initiative
	g.Settings = s.Settings
	if s.Schema < settingsSchemaVersion {
		g.Settings = migrateLegacySettings(s.State, s.UndoLimit)
	}
	g.StartingSeat = s.StartingSeat
	g.SplitSecondActive = s.SplitSecondActive
	g.eventSeq = s.EventSeq
	g.eventBatch = s.EventBatch
	g.oncePerBatchFired = copyStringUint64Map(s.OncePerBatchFired)
	g.announcedBlocks = copyUUIDPairMap(s.AnnouncedBlocks)
	g.blockedAttackers = copyBoolMap(s.BlockedAttackers)
	g.announcedAttacks = copyBoolMap(s.AnnouncedAttacks)
	g.firstStrikeStepParticipants = copyBoolMap(s.FirstStrikeStepParticipants)

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
		StartingDefense:          c.StartingDefense,
		ProtectorPlayerID:        c.ProtectorPlayerID,
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
			cloned := t
			cloned.Restrictions = copyStrings(t.Restrictions)
			out.ManaPool[i] = cloned
		}
	}
	out.CastPermissions = cloneCastPermissions(p.CastPermissions)
	out.Statics = clonePlayerStatics(p.Statics)
	return out
}

func restoreStackItem(s *stackItemSnapshot) *StackItem {
	out := &StackItem{
		ID:            s.ID,
		Kind:          s.Kind,
		Controller:    s.Controller,
		Owner:         s.Owner,
		SourceCardID:  s.SourceCardID,
		SourceEpoch:   s.SourceEpoch,
		Label:         s.Label,
		DoubledBy:     s.DoubledBy,
		DoubledByName: s.DoubledByName,
		Targets:       copyTargetRefs(s.Targets),
		Payload:       copyTargetRefs(s.Payload),
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
		Seq:           s.Seq,
		Ordered:       s.Ordered,
		Paid:          clonePaidCost(s.Paid),
		// Effect stays nil. A SPELL does not need one — resolution
		// dispatches through EffectResolver by oracle ID — but an
		// ability does, which is why a stack item with an Effect is
		// counted in the census and keeps the snapshot from being a
		// restore point.
	}
	if s.HasTargetSpec && spellSpecRederivable(*s) {
		out.targetSpec = CatalogTargetSpec(s.OracleID)
	}
	if s.HasModeSpec && s.Kind == StackItemSpell && s.OracleID != "" && CatalogModeSpec != nil {
		out.modeSpec = CatalogModeSpec(s.OracleID)
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
		// Effect, AppliesTo and Optional stay nil; see the census.
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
