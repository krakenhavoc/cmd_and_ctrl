package game

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// snapshot_drift_test.go is the reason this feature will still be
// correct in six months.
//
// clone.go and snapshot.go are both hand-written deep copies of the
// same domain types. Hand-written copies rot: someone adds a field to
// game.Game or game.Card, the compiler says nothing (a struct literal
// with named fields compiles fine when a field is missing), and the
// new field silently fails to survive an undo or a deploy. That class
// of bug is invisible until a player loses a counter.
//
// So every field on every type the snapshot touches must be
// CLASSIFIED here. Add a field to the domain and this test fails
// until you say what happens to it — carried, rebuilt, or dropped
// with a reason. It is a five-second edit when you know the answer
// and a useful interruption when you do not.
//
// It is deliberately a test and not a linter: it runs on every CI
// build, it names the exact field, and it tells you which file to
// edit.
//
// WHAT THIS FILE DOES NOT DO, and #1005 is the report that it was read
// as doing: it walks field NAMES and never reads a value. A row saying
// `carried` is a promise, and `dropped` is the only disposition this
// file holds to anything — TestDroppedFieldsAreAllCensused makes a
// dropped field name the census counter that accounts for it.
// snapshot_carried_test.go is the other half: it writes a value into
// every `carried` field, runs the real capture → restore, and reads it
// back off the restored game. Neither test is enough alone — one knows
// which fields exist, the other knows what happens to them.

// disposition records what the snapshot does with one field.
type disposition int

const (
	// carried: the snapshot serialises it and restore puts it back.
	carried disposition = iota
	// rebuilt: not serialised, because the restoring BINARY
	// reconstructs it (catalog lookups, process-lifetime singletons,
	// derived caches).
	rebuilt
	// dropped: not serialised and not reconstructable. Every entry
	// here must be counted by ContinuationCensus or be genuinely
	// irrelevant to game state; the reason string says which.
	dropped
)

type fieldPlan map[string]struct {
	how    disposition
	reason string
}

func plan(entries ...any) fieldPlan {
	out := fieldPlan{}
	for i := 0; i < len(entries); i += 3 {
		name := entries[i].(string)
		out[name] = struct {
			how    disposition
			reason string
		}{entries[i+1].(disposition), entries[i+2].(string)}
	}
	return out
}

// gameFields classifies every field on game.Game.
var gameFields = plan(
	"ID", carried, "",
	"CreatedAt", carried, "",
	"State", carried, "",
	"Seats", carried, "",
	"Battlefield", carried, "",
	"Stack", carried, "",
	"Exile", carried, "",
	"PhasedOut", carried, "",
	"Turn", carried, "",
	"MulligansOpen", carried, "",
	"Monarch", carried, "",
	"Initiative", carried, "",
	"Settings", carried, "",
	"StartingSeat", carried, "",
	"StackMeta", carried, "",
	"PendingTriggers", carried, "",
	"DelayedTriggers", carried, "",
	"SplitSecondActive", carried, "",
	"LoyaltyActivatedThisTurn", carried, "",
	"SpellsCastThisTurn", carried, "",
	"ForetoldThisTurn", carried, "",
	"LandsPlayedThisTurn", carried, "",
	"ExtraLandDropsThisTurn", carried, "",
	// Per-turn draw log (Sylvan Library's "cards in your hand drawn
	// this turn"). Carried for the same reason the other per-turn
	// tallies are: a restore mid-turn that forgot it would offer the
	// wrong candidate set, and the cards it names are still in hand.
	"DrawnThisTurn", carried, "",
	"TurnTally", carried, "",
	// #1181: what has been ACTIVATED, per (object, printed ability),
	// in both scopes. Carried for a stronger reason than the per-turn
	// tallies: the game-lifetime half never refreshes, so a restore
	// that dropped it would give every exhaust ability on the board a
	// second use.
	"Activations", carried, "",
	// #628 CR 726 loop breaker. Carried for the same reason the
	// per-turn tallies are: a restore mid-loop that forgot the notice
	// would come back with automatic passing live again, and the
	// threshold is configuration a restore must not silently
	// re-default.
	"LoopNotice", carried, "",
	"LoopThreshold", carried, "",
	"DiscardPending", carried, "",
	"Promises", carried, "",
	"Vote", carried, "",
	"PendingChoices", carried, "",
	// The persisted snapshot copies the log; the undo CLONE shares
	// its backing array and records the length, and RestoreFrom
	// truncates to it (#629). Both restore the same log, which is
	// what "carried" means here.
	"Events", carried, "shared with the live log by Clone, copied by the persisted snapshot",
	"eventSeq", carried, "",
	// #1401: the public log's projection cache and the generation it
	// validates against. A restored game is a NEW *Game — zero
	// generation, empty slot — and its first view refolds the carried
	// Events from scratch, so there is nothing to serialise.
	"eventLogGen", rebuilt, "names this *Game's log history; a restored game is a new receiver and starts a new one",
	"logProjection", rebuilt, "derived cache of the public log; the first view of a restored game refolds Events",
	// #829 event batches. Carried for the same reason the per-turn
	// tallies are, and carried TOGETHER: the counter names the batch
	// the marks are recorded against, so a restore that kept one and
	// not the other would either double-fire a "whenever one or more"
	// trigger or swallow it.
	"eventBatch", carried, "",
	"oncePerBatchFired", carried, "",
	// #1289: a resolution paused on one of its own prompts holds the
	// CR 704.3 boundary. Carried with each choice's midResolution.
	"resolutionOpen", carried, "",
	"resolutionDepth", dropped, "not game state: it counts resolution functions on the Go stack, so it is zero between actions (#1289)",
	// #830 block-declaration lock-in, and #715's blocked state.
	// Carried for the same reason and in the same pair-wise way: the
	// map of announced pairings names what the blocked marks were
	// recorded for, so a restore that kept one and not the other
	// would either re-announce an attacker that is already blocked or
	// swallow a real block. blockedAttackers is carried for one more
	// reason — a restore that dropped it would hand a blocked
	// attacker's combat damage to the defending player (CR 509.1h).
	"announcedBlocks", carried, "",
	"blockedAttackers", carried, "GameSnapshot.BlockedAttackers",
	// #859 attack-declaration lock-in. Carried for the reason the two
	// above are: a restore that dropped it would announce an attacker
	// that has already attacked, and one that invented it would
	// swallow a declaration the battlefield is still carrying.
	"announcedAttacks", carried, "",
	// #1364: the last-known defending player of each attack. Carried
	// with announcedAttacks: a restore that dropped it would leave an
	// attacker whose planeswalker has left unblockable (CR 506.4c).
	"attackDefenders", carried, "",
	// #1279: which defenders have completed their block declaration.
	// Carried with the block maps above: a restore that dropped it
	// would re-ask a defender who had already declared, and one that
	// invented it would read an attacker unblocked before the
	// defender chose.
	"blocksDeclared", carried, "",
	// #716 combat damage step participation. Carried for the reason
	// the three above are, and for one more: the window between the
	// two combat damage steps is a priority window, so an undo or a
	// deploy restore can land inside it. A restore that dropped the
	// record would let every first-striker deal its damage again in
	// the regular step.
	"firstStrikeStepParticipants", carried, "",
	"lastKnownBattlefield", carried, "",
	"lastKnownTriggerIdentity", carried, "",
	"lastKnownCounters", carried, "",
	// #1379: CR 608.2h LKI for permanents that left the battlefield
	// this turn. Carried, unlike lastKnownStack: a restore that lands
	// with an ability on the stack that names a departed permanent must
	// still be able to read how it last existed.
	"lastKnownPermanents", carried, "",
	// #1255: CR 608.2h LKI for spells that left the stack this turn.
	// Its only readers are copy effects that name a spell without
	// targeting it — a storm trigger, Thousand-Year Storm, Doublecast's
	// delayed trigger — and every one of those is a closure the
	// snapshot already refuses to carry. Clone copies it for undo.
	"lastKnownStack", dropped, "read only by stack items and delayed triggers whose behaviour is a closure, counted in ContinuationCensus.StackEffects and ContinuationCensus.DelayedTriggerEffects; with no such reader the record is dead and restores empty",
	// ADR 0054: the key and the per-turn stream counters ARE the
	// randomness. Clone copies them (undo rewinds) and rngSnapshot
	// carries them (a restore continues every stream).
	"rngKey", carried, "rngSnapshot.Key",
	"rngCounters", carried, "rngSnapshot.Counters",
	"rngTurn", carried, "rngSnapshot.Turn",
	"sourceOrdinals", carried, "GameSnapshot.SourceOrdinals",
	"sourceOrdinalNext", carried, "GameSnapshot.SourceOrdinalNext",
	"layerVersion", carried, "advanced by one on restore to force a recompute",
	"lastResolvedVersion", carried, "",

	"Listeners", rebuilt, "process-lifetime singletons installed by NewGame; a new binary's listener set wins",
	"BuiltinReplacements", rebuilt, "registered by NewGame, not per-game state",
	"mu", rebuilt, "a fresh receiver owns its own lock, exactly as Clone does",

	"ScopedStatics", dropped, "StaticAbility is two closures; counted in ContinuationCensus.ScopedStatics",
	"TurnScopedReplacements", dropped, "ReplacementEffect is three closures; counted in ContinuationCensus.TurnScopedReplacements",
	"TurnScopedBlockRules", dropped, "BlockRule is two closures; counted in ContinuationCensus.TurnScopedBlockRules",
	"testReplacements", dropped, "test-only injection slot; production has no path to it",
	"replacementsAppliedThisEvent", dropped, "non-empty between actions only for an event paused on a replacement prompt, and that prompt's resume frame is counted in ContinuationCensus.ChoiceResumeFrames; Clone deep-copies it for undo (#808)",
	"nextReplacementEventID", dropped, "mints keys for the map above, which restores empty",
	"promptRuns", dropped, "non-empty between actions only for a printed sacrifice, discard or resolution-time pick instruction paused on its prompts (#1214), and each of those prompts is counted in ContinuationCensus.ChoiceResumeFrames through PendingChoice.promptRun; the run holds a continuation closure the snapshot could not carry anyway; Clone deep-copies it for undo (#1019, #1027)",
	"enteringTokens", dropped, "non-empty between actions only for a created token whose battlefield entry is paused on a replacement prompt, and that prompt's resume frame is counted in ContinuationCensus.ChoiceResumeFrames; Clone copies it for undo (#762)",
	"resolving", dropped, "the CR 707.10 self-copy source (#920); set between actions only for a resolution paused on a prompt, and that prompt's resume frame is counted in ContinuationCensus.ChoiceResumeFrames; it holds a *StackItem, whose Effect is a closure the snapshot could not carry anyway; Clone shares it for undo",
	"recomputeCount", dropped, "test instrumentation for the layer fast-path, not game state",
	"simultaneousExit", dropped, "per-sweep scope, defer-cleared; a snapshot is never taken mid-wipe, so it is always empty between mutations",
)

var cardFields = plan(
	"InstanceID", carried, "",
	"Name", carried, "",
	"ScryfallID", carried, "",
	"OracleID", carried, "",
	// #521: the synthetic catalog key a TOKEN carries instead of an
	// oracle ID. Carried for the same reason GrantedAbilities is — it
	// is a catalog KEY, not a closure, and it is the whole of what
	// makes a token's abilities findable on the other side of a
	// restore. Drop it and a Treasure comes back a blank artifact.
	"TokenKey", carried, "",
	"TypeLine", carried, "",
	"Power", carried, "",
	"Toughness", carried, "",
	"ManaCost", carried, "",
	"ProducedMana", carried, "",
	"Colors", carried, "",
	"ColorIdentity", carried, "",
	"StartingLoyalty", carried, "",
	"Keywords", carried, "",
	// Multi-face model (#357). Printed per-face data: which face is
	// up decides TypeLine, ManaCost and P/T, so a restore that lost
	// ActiveFace would resurrect an MDFC on the wrong side.
	"Layout", carried, "",
	"Faces", carried, "",
	"ActiveFace", carried, "",
	"NeedsEffect", carried, "",
	"Owner", carried, "",
	"Controller", carried, "",
	"Tapped", carried, "",
	"NextUntapSkips", carried, "",
	"BattleX", carried, "",
	"BattleY", carried, "",
	"Counters", carried, "",
	"IsCommander", carried, "",
	"AttackingTarget", carried, "",
	"BlockingTarget", carried, "",
	"GoadedBy", carried, "",
	"DamageMarked", carried, "",
	// #667 regeneration shields. Carried for the reason DamageMarked
	// is: it is per-turn state on one permanent that nothing can
	// re-derive, and a restore that dropped it would let a creature
	// the table has already paid to protect die to the next Doom
	// Blade. Not a counter, so it is its own field rather than a
	// Counters entry (see Card.RegenerationShields).
	"RegenerationShields", carried, "",
	"FaceDown", carried, "",
	// ADR 0069: carried — the kind is the rule. A restore that
	// dropped it would bring back a face-down object with no viewers
	// row, no CR 708.2 answer and no catalog suppression. restoreCard
	// reads an ABSENT kind on a face-down card as FaceDownExiled, the
	// only face-down object that could exist before the field did.
	"FaceDownKind", carried, "",
	// #1270 / CR 708.2: the body an effect LISTED for a face-down
	// object. Carried — nothing can re-derive that Yedora's face-down
	// card is a Forest land rather than the default 2/2.
	"FaceDownListed", carried, "",
	// #1271 / CR 613.7f: the face-change timestamp. Carried with
	// EnteredBattlefieldAt, which it competes with for the layer sort.
	"FaceTurnedAt", carried, "",
	"KnownBy", carried, "",
	"EnteredBattlefieldAt", carried, "",
	// #936 / CR 400.7: the object's serial number, and the epoch half
	// of every per-object tally key. Carried, not rebuilt — nothing
	// can re-derive how many times a card has changed zones, and a
	// restore that zeroed it would silently merge the returning
	// object's "only once each turn" counts with the ones the object
	// before it wrote. Absent in a file written before the field
	// existed, which decodes as zero: the same answer a card that has
	// never moved gives. It is also the CR 400.7 object identity a
	// granted cast permission names (ADR 0066), so a restore that reset
	// it would revive every permission ever granted against the card —
	// the one direction this field must not fail in. Listed once: the
	// row was written twice, and a map keeps the last one.
	"ObjectEpoch", carried, "",
	"SummonedThisTurn", carried, "",
	"MarkedLethalByDeathtouch", carried, "",
	// #683: carried — a restore that dropped it would let a 0/0 that
	// lost its last counter survive the next state-based check.
	"LostLastCounter", carried, "",
	// Printed data, carried with Toughness for the same reason
	// VariableToughness is, one sign the other way: a restore that
	// dropped it would leave a living weapon Germ on the battlefield
	// as a 0/0 nothing can kill.
	"PrintedPTKnown", carried, "",
	// #683: printed data, carried with Toughness — a restore that
	// dropped it would let a `*` creature die to the toughness check
	// the first time it lost its last counter. Carried, and backfilled
	// from the printing when a file written before #683 has no key
	// (snapshot_backfill.go); every file this binary writes has it.
	"VariableToughness", carried, "",
	// S24 attachments (ADR 0036). Carried, not rebuilt: which sword
	// is on which creature is not derivable from anything else, and
	// a restore that dropped it would silently un-equip the board.
	"AttachedTo", carried, "",
	"AttachedAt", carried, "",
	// The layer-2 control baseline. Carried rather than rebuilt: a
	// restore that dropped it would re-capture the CURRENT (stolen)
	// controller as the base, and the creature would never go home.
	"BaseController", carried, "",
	// S16.5 copy effects (#159 / #335). Which card a permanent is a
	// copy of is not derivable from anything else on the board, and
	// a restore that lost it would resurrect every clone as the 0/0
	// it is printed as. Pure data by construction — see copy.go on
	// why PrintedValues carries no closures.
	"PrintedSelf", carried, "",
	// #665 / CR 707.9a: the ability bundles a copy effect's "except"
	// clause granted. Carried, and carriable at all, because it holds
	// catalog KEYS rather than closures — the abilities themselves
	// are static catalog data the restoring binary already has. A
	// restore that dropped it would leave a Phantasmal Image copy
	// with no sacrifice trigger and no Illusion type.
	"GrantedAbilities", carried, "",
	// S26: the creature type named as the permanent entered. A
	// player's choice, so nothing can rebuild it.
	"NamedTribe", carried, "",
	// #742: the colour named as the permanent entered. A player's
	// choice, so nothing can rebuild it.
	"ChosenColor", carried, "",
	// #980, CR 614.12 / CR 702.16k: the player named as the permanent
	// entered. A player's choice, so nothing can rebuild it — and it
	// is the whole of what True-Name Nemesis's protection reads.
	"ChosenPlayer", carried, "",
	// #1210, CR 614.12: the card NAME named as the permanent entered.
	// A player's choice, so nothing can rebuild it — and unlike the
	// three above there is not even a vocabulary to rebuild it from,
	// because CR 201.2 lets a player name any card name at all.
	"ChosenName", carried, "",
	// #653 / #664, CR 400.7d: what the spell that became this
	// permanent was cast for — the alternative cost and the optional
	// additional costs, one record. Carried, and it is the field here
	// with the least room to be anything else: it was copied off a
	// StackItem that no longer exists, so a restore that dropped it
	// could not rebuild it from any other part of the game. A Phlage
	// that escaped would come back hard-cast and sacrifice itself; a
	// kicked Gatekeeper of Malakir would come back unkicked.
	"Provenance", carried, "",
	// ADR 0071 (#757): the CR 716.2 level and CR 719.3 solved
	// designations. Carried, and the reason is sharper than for the
	// two above — both zero values are LEGAL states ("level 1",
	// "unsolved"), so a restore that dropped them would come back
	// wrong and say nothing about it.
	"ClassLevel", carried, "",
	"Solved", carried, "",
	// ADR 0071 amendment (#1321): the CR 701.64 harnessed designation.
	// Carried for Solved's reason — the zero value ("not harnessed")
	// is a legal state, so a restore that dropped it would come back
	// wrong and say nothing.
	"Harnessed", carried, "",
	// ADR 0090 (#1328): the CR 722.3a prepared designation, the
	// CR 722.3c copy's not-a-card marker, and the permanent object the
	// copy is kept in exile by. Carried for Solved's reason — every
	// zero value is a legal state, so a restore that dropped them would
	// come back wrong and say nothing.
	"Prepared", carried, "",
	"PrepareCopy", carried, "",
	"PreparedBy", carried, "",
	// ADR 0091 (#1331): the hideaway link. Carried — the zero value is
	// a legal state ("not hidden by anything"), so a restore that
	// dropped it would say nothing and leave the card orphaned.
	"HiddenBy", carried, "",
	// #1199 / CR 702.26, ADR 0084. All four are the phased-out status
	// and all four are legal zero values, so a restore that dropped
	// them would bring a phased board back under the wrong player's
	// untap step, or an Aura back without its host.
	"PhasedOutBy", carried, "",
	"PhaseInLockedBy", carried, "",
	"PhasedOutIndirect", carried, "",
	"TapOnPhaseIn", carried, "",
	// S27 battles. Both are printed / chosen state with no other
	// source: a restore that lost StartingDefense would re-stamp
	// nothing (the stamp is idempotent and only fires on entry), and
	// one that lost ProtectorPlayerID would leave a battle nobody
	// defends and everybody may attack.
	"StartingDefense", carried, "",
	"ProtectorPlayerID", carried, "",
	// #522: a restore brought this card back with fewer catalog
	// abilities than were captured. Carried, and it has to be: the
	// owner's rule is that nothing but the card leaving the game
	// clears it, and a later restore point is not the card leaving.
	"AbilitiesLostOnRestore", carried, "",

	"ManaAbilities", rebuilt, "closures; re-looked-up from the catalog by oracle ID, or by TokenKey for a token (#521), and censused only when the catalog cannot return them",
	"ActivatedAbilities", rebuilt, "same as ManaAbilities",
	"effective", rebuilt, "layer-engine characteristic cache; restore forces a recompute",
)

var playerFields = plan(
	"ID", carried, "",
	"Name", carried, "",
	"Seat", carried, "",
	"Life", carried, "",
	"Poison", carried, "",
	"Energy", carried, "",
	"Library", carried, "",
	"Hand", carried, "",
	"Graveyard", carried, "",
	"Command", carried, "",
	// #623 / CR 114: the other half of the command zone. Carried, and
	// it has to be — which emblems a player has is not derivable from
	// anything else on the board, and a restore that dropped them
	// would quietly un-ultimate a planeswalker.
	"Emblems", carried, "",
	"CommanderDamage", carried, "",
	"LifeHistory", carried, "",
	// The seat-turn counter "until your next turn" durations end on
	// (ADR 0063). Carried: a restore that dropped it would restart
	// every such effect's clock, and a departed seat's skipped turns
	// are not derivable from the board.
	"TurnsBegun", carried, "",
	"Eliminated", carried, "",
	"HandKept", carried, "",
	"MulligansTaken", carried, "",
	"DeckImported", carried, "",
	"UndosRemaining", carried, "",
	"DiscordID", carried, "",
	"DiscordAvatarHash", carried, "",
	"DisplayName", carried, "",
	// S31 bot seats. Carried, not rebuilt: which seat is a bot and
	// at what tier is not derivable from the board, and a restore
	// that dropped it would silently turn a bot into an empty chair
	// nobody is coming back to.
	"IsBot", carried, "",
	"BotTier", carried, "",
	"BotDeck", carried, "",
	"AttemptedEmptyDraw", carried, "",
	"CommanderCasts", carried, "",
	"Counters", carried, "",
	"MaxHandSize", carried, "",
	"LandDropsPerTurn", carried, "",
	"ManaPool", carried, "",
	// ADR 0066 granted cast and play permissions. Carried, not
	// rebuilt: who may cast what is not derivable from the board, and
	// a restore that dropped them would silently revoke a cascade hit
	// or a Snapcaster'd card nobody had cast yet. STANDING permissions
	// (Underworld Breach, Bolas's Citadel) are not in this slice at
	// all — they are re-derived from the battlefield on every query.
	"CastPermissions", carried, "",
	// #1197 granted player abilities ("you gain protection from
	// everything until your next turn"). Carried for the same reason
	// and by the same mechanism as the line above: plain data with no
	// closure, so PlayerSnapshot holds the engine type directly. The
	// DERIVED half — Leyline of Sanctity's "you have hexproof" — is
	// not in this slice at all and needs nothing, because it comes
	// back with the battlefield.
	"Statics", carried, "",
)

// scopedStaticFields classifies game.ScopedStatic — the floating
// continuous-effect registry's entry type. It was not classified
// before S38, so a field added to it used to vanish across a restore
// with nothing complaining. The whole entry is dropped and censused;
// `Duration` is the half of it that is plain data and could be
// carried the day #515 makes the ability re-derivable, which is why
// it is classified `carried` rather than sharing the closure's fate.
var scopedStaticFields = plan(
	"Ability", dropped, "two closures; counted by ContinuationCensus.ScopedStatics",
	"Source", dropped, "rides with the ability; counted by ContinuationCensus.ScopedStatics",
	"Timestamp", dropped, "rides with the ability; counted by ContinuationCensus.ScopedStatics",
	"Duration", carried, "plain data (duration.go); carried by Clone and ready for #515",
	"Label", dropped, "reaches the operator through ContinuationCensus.Labels",
)

var zoneFields = plan(
	"Kind", carried, "",
	"Owner", carried, "",
	"Cards", carried, "",
)

var stackItemFields = plan(
	"ID", carried, "",
	"Kind", carried, "",
	"Controller", carried, "",
	"Owner", carried, "",
	"SourceCardID", carried, "",
	// CR 400.7 (#812): which OBJECT an ability's source was at
	// announce, so "attach this permanent" can refuse a source that
	// left — or that left and came back as a new object. Carried,
	// and it has to be: the epoch is a reading of a card that has
	// since moved, so nothing in the restored board could recompute
	// it, and a restore that lost it would let a bounced-and-replayed
	// Equipment be equipped by the old ability.
	"SourceEpoch", carried, "",
	// #1418: the same reading for EVERY ability item, triggers
	// included — the object "this" names. Carried for SourceEpoch's
	// reason: it is a reading of a card that has since moved.
	"SourceObject", carried, "",
	"Label", carried, "",
	"DoubledBy", carried, "",
	"DoubledByName", carried, "",
	"Targets", carried, "",
	// #636 reflexive triggers: a pending trigger's payload is what
	// the resolution that created it told it (the cards revealed,
	// the creature sacrificed). Carried, and it has to be — the
	// Effect reads its whole input from here, so a restore that lost
	// it would resolve the trigger against nothing.
	"Payload", carried, "",
	// #1223: the triggering event (the damage amount, the object
	// that left with its CR 603.10 characteristics). Carried, and it
	// has to be — a trigger paused on its CR 603.3d target prompt is
	// a restore point, and the event is unrecoverable from the
	// restored board: the object it describes has already moved, and
	// the last-known-information map it was read from is cleared as
	// the harvest ends.
	"Trigger", carried, "",
	"Modes", carried, "",
	"XValue", carried, "",
	"Distribution", carried, "",
	"HoldPriority", carried, "",
	"CastFromZone", carried, "",
	"AltCost", carried, "",
	// CR 702.143c (#658). Carried: a restore that lost it would make
	// "if this spell was foretold" false for a spell already on the
	// stack, and the fact cannot be recomputed — the card turned face
	// up as it was cast, so the object it was read from is gone.
	"Foretold", carried, "",
	// CR 708.4 (#1194, ADR 0082). Carried: a restore that lost it
	// would resolve a morph on the stack into a face-UP creature,
	// revealing the card to the table and handing it back every
	// ability CR 708.2a says it does not have.
	"FaceDown", carried, "",
	// CR 702.34a / CR 400.7g (ADR 0066). Carried for the reason
	// IsCopy is: a restore that lost it would route a flashed-back
	// spell to a graveyard instead of exile, and a card Snapcaster
	// gave flashback to could then be flashed back again forever.
	"AltCostExiles", carried, "",
	"SplitSecond", carried, "",
	// S30 spell copies (#95). Carried, and it has to be: a restore
	// that lost the flag would route a resolving copy to a graveyard
	// as though it were a card, putting a phantom Twincast in
	// somebody's yard where Tarmogoyf can count it.
	"IsCopy", carried, "",
	"Seq", carried, "",
	"Ordered", carried, "",
	// #789 / #761: what the announcement paid — the counters
	// removed, the life, and the mana tokens that left the pool.
	// Carried, and it has to be: the counters are off the board and
	// the Treasure that made the mana may be in a graveyard by the
	// time the item resolves, so a restore that lost the record
	// would resolve Painful Truths for zero cards and a "for each
	// counter removed this way" ability for nothing.
	"Paid", carried, "",

	"targetSpec", rebuilt, "a spell's spec is re-derived from the catalog by oracle ID; an ability's is censused",
	// #764: the ModeSpec an item was announced under, so the CR
	// 608.2b re-check can find the clause of the mode occurrence a
	// TargetRef names. Same disposition as targetSpec and for the
	// same reason: catalog data, keyed by oracle ID for a spell and
	// unreachable for an ability, which is why an ability carrying
	// one is counted in ContinuationCensus.StackTargetSpecs.
	"modeSpec", rebuilt, "a spell's mode spec is re-derived from the catalog by oracle ID; an ability's is censused",
	"Effect", dropped, "a closure; counted in ContinuationCensus.StackEffects (spells need none — they dispatch via EffectResolver)",
)

var delayedTriggerFields = plan(
	"ID", carried, "",
	"Controller", carried, "",
	"SourceCardID", carried, "",
	"SourceObject", carried, "", // #1418, CR 603.7d
	"Label", carried, "",
	"At", carried, "",
	"ControllerTurnOnly", carried, "",
	"CreatedSeq", carried, "",
	"Cards", carried, "",
	// #663's event condition. The data half comes back so a restored
	// game still knows WHAT was owed and until when.
	"On", carried, "",
	"Duration", carried, "",

	"Effect", dropped, "a closure; counted in ContinuationCensus.DelayedTriggerEffects",
	"AppliesTo", dropped, "a closure; its trigger is counted once in ContinuationCensus.DelayedTriggerEffects through Effect beside it",
	"Optional", dropped, "a prompt declaration holding a Chooser closure; its trigger is counted once in ContinuationCensus.DelayedTriggerEffects through Effect beside it",
)

var pendingChoiceFields = plan(
	"ID", carried, "",
	"Kind", carried, "",
	"Chooser", carried, "",
	"FromPlayer", carried, "",
	"Count", carried, "",
	"Source", carried, "",
	"Reason", carried, "",
	"CoinAllowStop", carried, "",
	"CoinCount", carried, "",
	"CoinMaxUsefulWins", carried, "",
	"CoinWins", carried, "",
	"ColorOptions", carried, "",
	"ColorPurpose", carried, "",
	// Added by the mana pipeline (#352/#356). A restricted mana token
	// is game state that survives undo — clone.go deep-copies it at
	// clone.go:135 — so the snapshot must carry it too, or a restored
	// game would let the player spend restricted mana on anything.
	"ManaRestrictions", carried, "",
	"ManaSourceKinds", carried, "",
	// #742: how many tokens each colour of a one-pick-N-mana choice
	// mints (Gilded Lotus). Without it a restored pick adds one.
	"ManaAmounts", carried, "",
	// #763: whether answering this pick is "a permanent was tapped for
	// mana" (CR 106.12a), which is what fires the triggered mana
	// abilities. A restored Birds pick that lost it would put the mana
	// in the pool and skip Wild Growth's {G}.
	"ManaTapped", carried, "",
	"ReplacementEffectIDs", carried, "",
	"DamageAssignment", carried, "",
	"NoLegalTarget", carried, "",
	"PickTargetPlayers", carried, "",
	"PickTargetCards", carried, "",
	"PickTargetMin", carried, "",
	"PickTargetMax", carried, "",
	// #1196's CR 115.7 retarget prompt. Carried, and that is the
	// point of building it out of data: the prompt is a question
	// about an object already on the stack, so unlike every other
	// prompt in this family it holds no continuation and a game
	// paused on one is a restorable snapshot.
	"RetargetItem", carried, "",
	"RetargetPolicy", carried, "",
	"RetargetOptional", carried, "",
	"RetargetSlot", carried, "",
	"RetargetReason", carried, "",
	// #764 mode_pick. Carried for the same reason ChooseCards is:
	// the offered options ARE the prompt, and a restored game that
	// forgot them would put a question with no answers in front of a
	// seat. The bounds travel with them because the answer is
	// validated against them.
	"ModeOptionIndex", carried, "",
	"ModeOptionLabel", carried, "",
	"ModeMin", carried, "",
	"ModeMax", carried, "",
	"ModeRepeatable", carried, "",
	"SacrificeOptions", carried, "",
	"CopyOptions", carried, "",
	"ScryCards", carried, "",
	"LibraryPlacement", carried, "",
	"LibraryTopCount", carried, "",
	"LibraryTopDepth", carried, "",
	"TriggerOrderIDs", carried, "",
	"PayCost", carried, "",
	// #997: the step a pay-or-else prompt has to be answered in
	// (Stasis, Pact of Negation, cumulative upkeep). Carried so a
	// restored game gates exactly as the live one did; clone.go gets
	// it from the value copy every PendingChoice starts from.
	"OwedInStep", carried, "",
	// #951: the object on the stack a counter-unless-pays prompt is
	// about. Carried for the same reason — the gate re-reads it, so a
	// restored game that forgot it would let the guarded spell resolve
	// for free, which is the bug the field exists to stop.
	"GuardsStackItem", carried, "",
	"SearchCards", carried, "",
	"SearchMax", carried, "",
	// S28 cascade: which card the "you may cast it without paying
	// its mana cost" prompt is offering. Carried for the same reason
	// SacrificeOptions is — the prompt is meaningless without it, and
	// a restored game that forgot it would render an offer about
	// nothing.
	"MayCastCard", carried, "",
	// The chained-choice prompts (chained_choice.go). The labels are
	// the card's own words for the two branches and the candidate set
	// / bounds are what the enumerator reads to offer legal answers —
	// all of it is the prompt, and a restored game that forgot any of
	// it would render a question nobody can answer.
	"AcceptLabel", carried, "",
	"LifeCost", carried, "",
	"DeclineLabel", carried, "",
	"ChooseCards", carried, "",
	"ChooseMin", carried, "",
	"ChooseMax", carried, "",
	// #568's option pick: the branches of "choose one of the
	// following", carried for the same reason ChooseCards is — the
	// options ARE the prompt, and a restored game that forgot them
	// would put a question with no answers in front of a seat.
	"PickOptions", carried, "",
	// #804's CR 726 shortcut prompt. Carried for the reason
	// LoopNotice is: the key is the only way back to the run the
	// answer is about, and a restored game that forgot it would put a
	// question about nothing in front of the loop's controller — or,
	// worse, take an answer and attach the allowance to no run at all,
	// which is a table that starts spinning again.
	"LoopShortcutKey", carried, "",
	"LoopShortcutCount", carried, "",
	"LoopShortcutRepeat", carried, "",

	"replacementResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"costCommanderResume", dropped, "a parked cost announcement (#1397); counted in ContinuationCensus.ChoiceResumeFrames",
	"modePickResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"pickTargetResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"copyResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"triggerResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"payUnlessResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"mayCastResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"optionPickResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"searchResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"scryResume", dropped, "continuation closure; counted in ContinuationCensus.ChoiceResumeFrames",
	"libraryOrderResume", dropped, "continuation closure; counted in ContinuationCensus.ChoiceResumeFrames",
	"midResolution", carried, "",
	"confirmResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"chooseColorResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"chooseCardsResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"coinFlipResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"promptRun", dropped, "the id of the prompted run this prompt is one leg of — a sacrifice (#1019), a discard (#1027) or one of the three resolution-time picks (#1214); the run's continuation lives on Game.promptRuns and is counted in ContinuationCensus.ChoiceResumeFrames through this field; Clone copies it with the rest of the choice",
)

// driftPlans is every domain type the snapshot touches, paired with the
// plan that classifies its fields.
//
// ONE list, because three tests walk it: the drift guard below, the
// dropped-field census check under it, and the `carried` enforcement in
// snapshot_carried_test.go. A domain type added to one list and not the
// others is exactly the drift this file exists to stop — and the third
// of those tests fails until a new type has a probe, so a plan cannot
// arrive with nothing checking its promises.
var driftPlans = []struct {
	sample any
	plan   fieldPlan
}{
	{Game{}, gameFields},
	{Card{}, cardFields},
	{Player{}, playerFields},
	{Zone{}, zoneFields},
	{StackItem{}, stackItemFields},
	{DelayedTrigger{}, delayedTriggerFields},
	{PendingChoice{}, pendingChoiceFields},
	{ScopedStatic{}, scopedStaticFields},
}

// driftPlanName is the type name a plan is keyed and reported under.
func driftPlanName(sample any) string { return reflect.TypeOf(sample).Name() }

// TestSnapshotCoversEveryDomainField is the drift guard.
func TestSnapshotCoversEveryDomainField(t *testing.T) {
	for _, tc := range driftPlans {
		rt := reflect.TypeOf(tc.sample)
		t.Run(rt.Name(), func(t *testing.T) {
			live := map[string]bool{}
			for i := 0; i < rt.NumField(); i++ {
				name := rt.Field(i).Name
				live[name] = true
				if _, ok := tc.plan[name]; !ok {
					t.Errorf(`%s.%s is not classified.

A field was added to the domain type and the snapshot does not know
about it, so it will NOT survive a deploy. Decide which it is and add
it to %sFields in snapshot_drift_test.go:

  carried  — add it to the mirror in snapshot.go (capture AND restore)
  rebuilt  — the restoring binary reconstructs it; say how
  dropped  — it cannot be carried; make sure ContinuationCensus
             counts it, or say why it is not game state

Consider whether clone.go needs it too — an undo has the same
problem.`, rt.Name(), name, strings.ToLower(rt.Name()[:1])+rt.Name()[1:])
				}
			}
			// A plan entry with no matching field is stale — the
			// field was renamed or removed.
			var stale []string
			for name := range tc.plan {
				if !live[name] {
					stale = append(stale, name)
				}
			}
			sort.Strings(stale)
			for _, name := range stale {
				t.Errorf("%sFields lists %q, which no longer exists on %s — remove it",
					strings.ToLower(rt.Name()[:1])+rt.Name()[1:], name, rt.Name())
			}
		})
	}
}

// TestDroppedFieldsAreAllCensused ties the `dropped` disposition to
// something enforceable: every dropped field's reason must name the
// census counter that accounts for it, or explicitly say it is not
// game state. Without this the `dropped` bucket would be a place to
// quietly lose things.
func TestDroppedFieldsAreAllCensused(t *testing.T) {
	all := map[string]fieldPlan{}
	for _, tc := range driftPlans {
		all[driftPlanName(tc.sample)] = tc.plan
	}
	censusFields := map[string]bool{}
	ct := reflect.TypeOf(ContinuationCensus{})
	for i := 0; i < ct.NumField(); i++ {
		censusFields[ct.Field(i).Name] = true
	}

	for typeName, p := range all {
		for field, d := range p {
			if d.how != dropped {
				continue
			}
			if d.reason == "" {
				t.Errorf("%s.%s is dropped with no reason", typeName, field)
				continue
			}
			// Either it names a census counter, or it declares
			// itself not-game-state.
			named := false
			for c := range censusFields {
				if strings.Contains(d.reason, "ContinuationCensus."+c) {
					named = true
					break
				}
			}
			notState := strings.Contains(d.reason, "not game state") ||
				strings.Contains(d.reason, "test-only") ||
				strings.Contains(d.reason, "always empty") ||
				strings.Contains(d.reason, "restores empty")
			if !named && !notState {
				t.Errorf(`%s.%s is dropped but its reason neither names a
ContinuationCensus counter nor declares the field to be outside game
state. reason = %q

A dropped field that nothing counts is state the server loses without
telling anyone.`, typeName, field, d.reason)
			}
		}
	}
}

// TestEmbeddedDomainTypesStayPureData guards the types snapshot.go
// embeds BY VALUE instead of mirroring (Event, Turn, Vote, ...). That
// shortcut is only safe while they hold no funcs and no unexported
// fields; encoding/json refuses the first and silently skips the
// second. Marshalling each one catches a func immediately, and the
// reflection walk catches an unexported field.
func TestEmbeddedDomainTypesStayPureData(t *testing.T) {
	samples := []any{
		Event{}, Turn{}, Vote{}, TargetRef{}, ManaToken{},
		LifeChange{}, Characteristic{}, CastTally{},
		// Event.CombatStep and DamageAssignmentFrame.CombatStep (#187)
		// are carried through these two with no plan entry; their zero
		// value "" is right for a file written before them, so no
		// schema bump (combat_step_snapshot_test.go).
		DamageAssignmentFrame{}, ManaPool{},
		// Player.CastPermissions is one of these: PlayerSnapshot holds
		// []CastPermission by value, so the type's fields never reach
		// playerFields and a func or an unexported field added to it
		// would vanish across a restart with nothing complaining.
		// Listed from S32, when the permission started carrying a face
		// and stopped being a type nobody ever extends; ADR 0066 made
		// the guard load-bearing, because the whole reason a granted
		// permission is a struct of flags rather than a predicate is
		// that it has to survive this check.
		CastPermission{},
	}
	for _, s := range samples {
		rt := reflect.TypeOf(s)
		t.Run(rt.Name(), func(t *testing.T) {
			if _, err := json.Marshal(s); err != nil {
				t.Fatalf(`%s is embedded by value in GameSnapshot but no
longer marshals: %v

Either remove the func field, or give the type an explicit mirror in
snapshot.go the way StackItem and PendingChoice have.`, rt.Name(), err)
			}
			if rt.Kind() != reflect.Struct {
				return
			}
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				if f.PkgPath != "" {
					t.Errorf(`%s.%s is unexported, but %s is embedded by value
in GameSnapshot — encoding/json will skip the field and it will vanish
across a restart. Give %s an explicit mirror in snapshot.go.`,
						rt.Name(), f.Name, rt.Name(), rt.Name())
				}
			}
		})
	}
}

// TestSnapshotMirrorsHaveNoFuncs proves the snapshot types themselves
// are serialisable all the way down — a func reaching a mirror struct
// would make json.Marshal fail at runtime, in production, on the
// write path, for one unlucky game.
func TestSnapshotMirrorsHaveNoFuncs(t *testing.T) {
	var walk func(rt reflect.Type, path string, seen map[reflect.Type]bool)
	walk = func(rt reflect.Type, path string, seen map[reflect.Type]bool) {
		if seen[rt] {
			return
		}
		seen[rt] = true
		switch rt.Kind() {
		case reflect.Func, reflect.Chan, reflect.UnsafePointer:
			t.Errorf("%s is a %s — GameSnapshot must be serialisable all the way down", path, rt.Kind())
		case reflect.Ptr, reflect.Slice, reflect.Array:
			walk(rt.Elem(), path+"[]", seen)
		case reflect.Map:
			walk(rt.Key(), path+"{key}", seen)
			walk(rt.Elem(), path+"{}", seen)
		case reflect.Struct:
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				if f.PkgPath != "" && rt != reflect.TypeOf(GameSnapshot{}) {
					// Unexported fields inside embedded stdlib types
					// (time.Time) are fine — they marshal via
					// MarshalJSON. Only flag our own.
					if !strings.HasPrefix(rt.PkgPath(), "github.com/krakenhavoc") {
						continue
					}
				}
				walk(f.Type, fmt.Sprintf("%s.%s", path, f.Name), seen)
			}
		}
	}
	// time.Time and uuid.UUID marshal via interfaces; skip their
	// internals.
	seen := map[reflect.Type]bool{}
	walk(reflect.TypeOf(GameSnapshot{}), "GameSnapshot", seen)

	// And prove it for real, not just structurally.
	if _, err := json.Marshal(GameSnapshot{}); err != nil {
		t.Errorf("zero GameSnapshot does not marshal: %v", err)
	}
}
