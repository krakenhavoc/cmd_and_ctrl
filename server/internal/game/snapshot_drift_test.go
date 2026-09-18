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
	"Turn", carried, "",
	"MulligansOpen", carried, "",
	"Monarch", carried, "",
	"Initiative", carried, "",
	"UndoLimit", carried, "",
	"StartingSeat", carried, "",
	"StackMeta", carried, "",
	"PendingTriggers", carried, "",
	"DelayedTriggers", carried, "",
	"SplitSecondActive", carried, "",
	"LoyaltyActivatedThisTurn", carried, "",
	"SpellsCastThisTurn", carried, "",
	"LandsPlayedThisTurn", carried, "",
	"ExtraLandDropsThisTurn", carried, "",
	// Per-turn draw log (Sylvan Library's "cards in your hand drawn
	// this turn"). Carried for the same reason the other per-turn
	// tallies are: a restore mid-turn that forgot it would offer the
	// wrong candidate set, and the cards it names are still in hand.
	"DrawnThisTurn", carried, "",
	"TurnTally", carried, "",
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
	// #829 event batches. Carried for the same reason the per-turn
	// tallies are, and carried TOGETHER: the counter names the batch
	// the marks are recorded against, so a restore that kept one and
	// not the other would either double-fire a "whenever one or more"
	// trigger or swallow it.
	"eventBatch", carried, "",
	"oncePerBatchFired", carried, "",
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
	// #716 combat damage step participation. Carried for the reason
	// the three above are, and for one more: the window between the
	// two combat damage steps is a priority window, so an undo or a
	// deploy restore can land inside it. A restore that dropped the
	// record would let every first-striker deal its damage again in
	// the regular step.
	"firstStrikeStepParticipants", carried, "",
	"lastKnownBattlefield", carried, "",
	"lastKnownTriggerIdentity", carried, "",
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
	"testReplacements", dropped, "test-only injection slot; production has no path to it",
	"replacementsAppliedThisEvent", dropped, "non-empty between actions only for an event paused on a replacement prompt, and that prompt's resume frame is counted in ContinuationCensus.ChoiceResumeFrames; Clone deep-copies it for undo (#808)",
	"nextReplacementEventID", dropped, "mints keys for the map above, which restores empty",
	"enteringTokens", dropped, "non-empty between actions only for a created token whose battlefield entry is paused on a replacement prompt, and that prompt's resume frame is counted in ContinuationCensus.ChoiceResumeFrames; Clone copies it for undo (#762)",
	"recomputeCount", dropped, "test instrumentation for the layer fast-path, not game state",
	"simultaneousExit", dropped, "per-sweep scope, defer-cleared; a snapshot is never taken mid-wipe, so it is always empty between mutations",
)

var cardFields = plan(
	"InstanceID", carried, "",
	"Name", carried, "",
	"ScryfallID", carried, "",
	"OracleID", carried, "",
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
	"KnownBy", carried, "",
	"EnteredBattlefieldAt", carried, "",
	"SummonedThisTurn", carried, "",
	"MarkedLethalByDeathtouch", carried, "",
	// #683: carried — a restore that dropped it would let a 0/0 that
	// lost its last counter survive the next state-based check.
	"LostLastCounter", carried, "",
	// #683: printed data, carried with Toughness — a restore that
	// dropped it would let a `*` creature die to the toughness check
	// the first time it lost its last counter. Carried, and backfilled
	// from the printing when a file written before #683 has no key
	// (snapshot_backfill.go); every file this binary writes has it.
	"VariableToughness", carried, "",
	// Embedded BY VALUE in CardSnapshot rather than mirrored, so
	// every field it grows — S29's NotBeforeTurn, S32's Face — is
	// carried automatically and none of them appear in this plan.
	// That shortcut is only safe while the type stays pure data,
	// which TestEmbeddedDomainTypesStayPureData now proves.
	"ExilePlay", carried, "",
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
	// S26: the creature type named as the permanent entered. A
	// player's choice, so nothing can rebuild it.
	"NamedTribe", carried, "",
	// #742: the colour named as the permanent entered. A player's
	// choice, so nothing can rebuild it.
	"ChosenColor", carried, "",
	// S27 battles. Both are printed / chosen state with no other
	// source: a restore that lost StartingDefense would re-stamp
	// nothing (the stamp is idempotent and only fires on entry), and
	// one that lost ProtectorPlayerID would leave a battle nobody
	// defends and everybody may attack.
	"StartingDefense", carried, "",
	"ProtectorPlayerID", carried, "",

	"ManaAbilities", rebuilt, "closures; re-looked-up from the catalog by oracle ID, or censused when the card has none (a true token)",
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
	"Modes", carried, "",
	"XValue", carried, "",
	"Distribution", carried, "",
	"HoldPriority", carried, "",
	"CastFromZone", carried, "",
	"AltCost", carried, "",
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
	"Label", carried, "",
	"At", carried, "",
	"ControllerTurnOnly", carried, "",
	"CreatedTurn", carried, "",
	"Cards", carried, "",

	"Effect", dropped, "a closure; counted in ContinuationCensus.DelayedTriggerEffects",
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
	// Added by the mana pipeline (#352/#356). A restricted mana token
	// is game state that survives undo — clone.go deep-copies it at
	// clone.go:135 — so the snapshot must carry it too, or a restored
	// game would let the player spend restricted mana on anything.
	"ManaRestrictions", carried, "",
	// #742: how many tokens each colour of a one-pick-N-mana choice
	// mints (Gilded Lotus). Without it a restored pick adds one.
	"ManaAmounts", carried, "",
	"ReplacementEffectIDs", carried, "",
	"DamageAssignment", carried, "",
	"NoLegalTarget", carried, "",
	"PickTargetPlayers", carried, "",
	"PickTargetCards", carried, "",
	"PickTargetMin", carried, "",
	"PickTargetMax", carried, "",
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
	"TriggerOrderIDs", carried, "",
	"PayCost", carried, "",
	// #567: a prompt that stops the table though its kind does not
	// (cumulative upkeep's "sacrifice this unless you pay"). Carried
	// so a restored game gates exactly as the live one did; clone.go
	// gets it from the value copy every PendingChoice starts from.
	"ForceBlocks", carried, "",
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
	"modePickResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"pickTargetResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"copySpellResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"triggerResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"payUnlessResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"mayCastResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"optionPickResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"searchResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"scryResume", dropped, "continuation closure; counted in ContinuationCensus.ChoiceResumeFrames",
	"confirmResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"chooseColorResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"chooseCardsResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"coinFlipResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
)

// TestSnapshotCoversEveryDomainField is the drift guard.
func TestSnapshotCoversEveryDomainField(t *testing.T) {
	cases := []struct {
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

	for _, tc := range cases {
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
	all := map[string]fieldPlan{
		"Game":           gameFields,
		"Card":           cardFields,
		"Player":         playerFields,
		"Zone":           zoneFields,
		"StackItem":      stackItemFields,
		"DelayedTrigger": delayedTriggerFields,
		"PendingChoice":  pendingChoiceFields,
		"ScopedStatic":   scopedStaticFields,
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
		// Card.ExilePlay is one of these: CardSnapshot holds an
		// ExilePlayPermission by value, so the type's fields never
		// reach cardFields and a func or an unexported field added to
		// it would vanish across a restart with nothing complaining.
		// Listed from S32, when the permission started carrying a
		// face and stopped being a type nobody ever extends.
		ExilePlayPermission{},
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
