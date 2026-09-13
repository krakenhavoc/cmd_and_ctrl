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
	"DiscardPending", carried, "",
	"Promises", carried, "",
	"Vote", carried, "",
	"PendingChoices", carried, "",
	"Events", carried, "",
	"eventSeq", carried, "",
	"lastKnownBattlefield", carried, "",
	"rngState", carried, "marshalled via rngSnapshot",
	"layerVersion", carried, "advanced by one on restore to force a recompute",
	"lastResolvedVersion", carried, "",

	"Listeners", rebuilt, "process-lifetime singletons installed by NewGame; a new binary's listener set wins",
	"BuiltinReplacements", rebuilt, "registered by NewGame, not per-game state",
	"rng", rebuilt, "rebuilt by wrapping the restored rngState",
	"mu", rebuilt, "a fresh receiver owns its own lock, exactly as Clone does",

	"TurnScopedStatics", dropped, "StaticAbility is two closures; counted in ContinuationCensus.TurnScopedStatics",
	"TurnScopedReplacements", dropped, "ReplacementEffect is three closures; counted in ContinuationCensus.TurnScopedReplacements",
	"testReplacements", dropped, "test-only injection slot; production has no path to it",
	"replacementsAppliedThisEvent", dropped, "per-pipeline-call scope, defer-cleared; always empty between Applies",
	"nextReplacementEventID", dropped, "mints keys for the map above, which restores empty",
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
	"BattleX", carried, "",
	"BattleY", carried, "",
	"Counters", carried, "",
	"IsCommander", carried, "",
	"AttackingTarget", carried, "",
	"BlockingTarget", carried, "",
	"GoadedBy", carried, "",
	"DamageMarked", carried, "",
	"FaceDown", carried, "",
	"KnownBy", carried, "",
	"EnteredBattlefieldAt", carried, "",
	"SummonedThisTurn", carried, "",
	"MarkedLethalByDeathtouch", carried, "",
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
	"CommanderDamage", carried, "",
	"LifeHistory", carried, "",
	"Eliminated", carried, "",
	"HandKept", carried, "",
	"MulligansTaken", carried, "",
	"DeckImported", carried, "",
	"UndosRemaining", carried, "",
	"DiscordID", carried, "",
	"DiscordAvatarHash", carried, "",
	"DisplayName", carried, "",
	"LosesAtNextSBA", carried, "",
	"CommanderCasts", carried, "",
	"Counters", carried, "",
	"MaxHandSize", carried, "",
	"ManaPool", carried, "",
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
	"Targets", carried, "",
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

	"targetSpec", rebuilt, "a spell's spec is re-derived from the catalog by oracle ID; an ability's is censused",
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
	"ColorOptions", carried, "",
	// Added by the mana pipeline (#352/#356). A restricted mana token
	// is game state that survives undo — clone.go deep-copies it at
	// clone.go:135 — so the snapshot must carry it too, or a restored
	// game would let the player spend restricted mana on anything.
	"ManaRestrictions", carried, "",
	"ReplacementEffectIDs", carried, "",
	"DamageAssignment", carried, "",
	"NoLegalTarget", carried, "",
	"PickTargetPlayers", carried, "",
	"PickTargetCards", carried, "",
	"PickTargetMin", carried, "",
	"PickTargetMax", carried, "",
	"SacrificeOptions", carried, "",
	"CopyOptions", carried, "",
	"ScryCards", carried, "",
	"TriggerOrderIDs", carried, "",
	"PayCost", carried, "",
	"SearchCards", carried, "",
	"SearchMax", carried, "",
	// S28 cascade: which card the "you may cast it without paying
	// its mana cost" prompt is offering. Carried for the same reason
	// SacrificeOptions is — the prompt is meaningless without it, and
	// a restored game that forgot it would render an offer about
	// nothing.
	"MayCastCard", carried, "",

	"replacementResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"pickTargetResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"copySpellResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"triggerResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"payUnlessResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"mayCastResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"searchResume", dropped, "continuation frame; counted in ContinuationCensus.ChoiceResumeFrames",
	"scryResume", dropped, "continuation closure; counted in ContinuationCensus.ChoiceResumeFrames",
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
		DamageAssignmentFrame{}, ManaPool{},
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
