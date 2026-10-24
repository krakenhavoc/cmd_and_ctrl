package game

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// replacement_resume_keyed_test.go — #4253: a CR 616 prompt's resume
// frame is data. The catalog-card restores (dredge, Retriever Phoenix,
// riot, an ordering prompt) are in effects/replacement_resume_restore_test.go;
// these pin the codec and the one keyed shape the engine owns alone, a
// commander's CR 903.9b bounce.

// Every field of ReplacementEvent says what a keyed frame does with it.
func TestReplacementEventFieldsAreClassified(t *testing.T) {
	rt := reflect.TypeOf(ReplacementEvent{})
	live := map[string]bool{}
	for i := 0; i < rt.NumField(); i++ {
		name := rt.Field(i).Name
		live[name] = true
		if _, ok := replacementEventFields[name]; !ok {
			t.Errorf(`ReplacementEvent.%s is not classified in replacementEventFields
(replacement_resume_keyed.go). Say whether a keyed CR 616 frame carries it
(add it to replacementEventSnapshot, both directions), gates the frame
(it holds a closure or belongs to a kind that is not keyed), or is a cache.
Until then a frame whose event sets it is not keyed.`, name)
		}
	}
	for name := range replacementEventFields {
		if !live[name] {
			t.Errorf("replacementEventFields lists %q, which ReplacementEvent no longer has", name)
		}
	}
}

// keyedFrameSample is an event with every carried field set, the way no
// real event is, so the round trip below reads each one back.
func keyedFrameSample() *ReplacementEvent {
	id := func(s string) uuid.UUID { return uuid.NewSHA1(uuid.Nil, []byte(s)) }
	return &ReplacementEvent{
		ID: 41, Kind: RepEventMove, Canceled: true, Actor: id("actor"), Source: id("source"), mustSettleNow: true,
		DrawPlayer: id("drawer"), DrawCount: 2,
		drawPending:    []pendingDraw{{player: id("p"), count: 3, applied: []ReplacementEffectID{7, 9}}},
		drawInstead:    drawInsteadRun{ref: DrawInsteadRef{key: "di"}, source: id("di-src")},
		drawTail:       4,
		drawThen:       DrawThen{Ref: DrawThenRef{key: "dt"}, Player: id("dt-p"), Source: id("dt-s"), N: 5},
		keywordInstead: keywordInsteadRun{ref: KeywordInsteadRef{key: "ki"}, source: id("ki-src")},
		KeywordAction:  KeywordActionLearn, KeywordActionCount: 1,
		keywordAction: &keywordActionTail{
			cards: []uuid.UUID{id("kc")}, players: []uuid.UUID{id("kp")}, ask: true, choice: PendingChoiceScry, land: id("land"),
		},
		CounterTarget: id("ct"), CounterName: CounterPlusOne, CounterDelta: 3, CounterPlayer: id("cp"),
		StartingLoyalty: true, CounterPlacer: id("placer"), CounterFromCombatDamage: true, CounterFromCost: true,
		CounterFromTurnBasedAction: true,
		LifePlayer:                 id("lifer"), LifeDelta: -4, LifeFromRadiation: true,
		CardID: id("card"), OldZone: ZoneStack, NewZone: ZoneBattlefield, NewZoneOwner: id("owner"),
		ShuffleDestinationLibrary: true, ExiledWith: PermissionCardRef{ID: id("ew"), Epoch: 2},
		ExileCounter: "egg", Destruction: true, CantBeRegenerated: true, StateBasedDestruction: true,
		EntersTapped: true, EntersPrepared: true, EntersDevoured: 2, EntersWithHaste: true,
		entryRefused: true, EntersUnlocked: DoorMask(3), EntersAttacking: id("defender"),
		FaceDown: FaceDownKind("morph"), FaceDownListed: &FaceDownListing{Types: []string{"Creature"}, Subtypes: []string{"Elf"}, Power: 2, Toughness: 3},
		EntersAsCopyOf: &PrintedValues{Name: "Copied"}, copySourceID: id("copy"), copyUntilEndOfTurn: true,
		auraHost: TargetRef{Kind: TargetCard, ID: id("host")}, auraHostFixed: true, auraHostSettled: true,
		entersAsAura: true, exchangeTextWith: id("xt"),
		stackItem:          &StackItem{ID: id("item"), Controller: id("actor"), SourceCardID: id("card"), Label: "Spell"},
		putByEntering:      true,
		EntersWithCounters: map[string]int{CounterPlusOne: 2},
		entryResumable:     true,
		entryTail:          &entryTail{newObject: true, craftedWith: []ObjectRef{{ID: id("crafted"), Epoch: 1}}, batchIndex: 2},
		landPlay:           true, landPlayer: id("lander"),
		entryControllerSet: true, entryControllerPrior: id("prior"),
		zoneRoute: &zoneRoute{
			CardID: id("card"), Dst: ZoneHand, DstOwner: id("owner"), Actor: id("actor"), ToBottom: true, Depth: 2,
			FaceDown: FaceDownKind("morph"), HiddenBy: PermissionCardRef{ID: id("hb"), Epoch: 1},
			ExiledWith: PermissionCardRef{ID: id("rew"), Epoch: 3}, ExiledWithAbility: AbilityRef{Key: "k", Slot: "s", Ref: "own:0", Name: "n"},
			Mill: true, Discard: true, DiscardCause: DiscardCause("effect"), Source: id("rsrc"), MustSettleNow: true,
			Countered: true, Cause: MoveCause{Kind: MoveCauseKind("effect"), Controller: id("cc"), Item: id("ci")},
			Sacrifice: true, ViaBattlefieldLeave: true, Destruction: true, CantBeRegenerated: true, StateBased: true,
			AsCommander: true, commanderAnswer: commanderZoneAccept,
		},
		lifeAfterMove:       []lifeInstruction{{Source: id("ls"), Player: id("lp"), Delta: -2}},
		extraTurnsAfterMove: []extraTurnInstruction{{Source: id("xs"), Player: id("xp")}},
		asCommanderMove:     true,
		commanderAnswer:     commanderZoneDecline,
	}
}

// The mirror carries every carried field, both ways, through JSON: the
// sample sets each one (checked here, so a new carried field has to be
// added to the sample), and the restored event equals it.
func TestReplacementEventMirrorRoundTrips(t *testing.T) {
	ev := keyedFrameSample()
	v := reflect.ValueOf(ev).Elem()
	for i := 0; i < v.NumField(); i++ {
		name := v.Type().Field(i).Name
		if replacementEventFields[name] == repFieldCarried && v.Field(i).IsZero() {
			t.Errorf("keyedFrameSample leaves the carried field %s at its zero value", name)
		}
	}
	for _, sub := range []struct {
		val  reflect.Value
		skip map[string]bool
	}{
		{reflect.ValueOf(ev.zoneRoute).Elem(), map[string]bool{"then": true, "simultaneousExit": true, "simultaneousEntry": true}},
		{reflect.ValueOf(ev.entryTail).Elem(), map[string]bool{"then": true, "batch": true}},
		{reflect.ValueOf(ev.keywordAction).Elem(), map[string]bool{"then": true, "amassed": true, "armyToken": true, "armySubtype": true}},
	} {
		for i := 0; i < sub.val.NumField(); i++ {
			name := sub.val.Type().Field(i).Name
			if !sub.skip[name] && sub.val.Field(i).IsZero() {
				t.Errorf("keyedFrameSample leaves %s.%s at its zero value: carry it in the mirror, or gate it in replacementFrameKeyable", sub.val.Type().Name(), name)
			}
		}
	}

	g := NewGame()
	raw, err := json.Marshal(g.snapshotReplacementEvent(ev, &ContinuationCensus{}))
	if err != nil {
		t.Fatal(err)
	}
	var back replacementEventSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	got := restoreReplacementEvent(&back)
	if got.stackItem == nil || got.stackItem.ID != ev.stackItem.ID || got.stackItem.Label != ev.stackItem.Label {
		t.Errorf("stack item %+v, want %+v", got.stackItem, ev.stackItem)
	}
	got.stackItem, ev.stackItem = nil, nil
	if !reflect.DeepEqual(got, ev) {
		t.Errorf("the event did not round-trip:\n got %+v\nwant %+v\nroute got %+v\nroute want %+v", got, ev, got.zoneRoute, ev.zoneRoute)
	}
}

// A frame whose event holds a closure stays a closure.
func TestAFrameHoldingAClosureIsNotKeyed(t *testing.T) {
	g := NewGame()
	base := func() *PendingChoice {
		return &PendingChoice{Kind: PendingChoiceOptionalReplacement, replacementResume: &replacementResumeFrame{
			ev:         &ReplacementEvent{Kind: RepEventMove, CardID: uuid.New(), OldZone: ZoneBattlefield, NewZone: ZoneHand},
			applicable: []activeReplacement{{id: builtinReplacementIDBase}},
		}}
	}
	if g.replacementFrameKeyable(base()) != ReplacementResumeMove {
		t.Fatal("a plain bounce's frame is not keyed")
	}
	// #4253 slice 2: the event kinds and prompt kinds keyed since.
	for name, mut := range map[string]func(c *PendingChoice){
		"a counter placement": func(c *PendingChoice) { c.replacementResume.ev.Kind = RepEventCounter },
		"a life change":       func(c *PendingChoice) { c.replacementResume.ev.Kind = RepEventLife },
		"a scry with no then": func(c *PendingChoice) {
			c.replacementResume.ev.Kind = RepEventKeywordAction
			c.replacementResume.ev.keywordAction = &keywordActionTail{choice: PendingChoiceScry}
		},
		"a copy prompt":       func(c *PendingChoice) { c.Kind = PendingChoiceCopyTarget },
		"an aura host prompt": func(c *PendingChoice) { c.Kind, c.replacementResume.applicable = PendingChoiceAuraHost, nil },
		"an entry sacrifice": func(c *PendingChoice) {
			c.Kind = PendingChoiceEntrySacrifice
			c.chooseCardsResume = &chooseCardsFrame{zone: ZoneBattlefield}
		},
	} {
		c := base()
		mut(c)
		if key := g.replacementFrameKeyable(c); key == "" {
			t.Errorf("%s: the frame is not keyed", name)
		}
	}
	for name, mut := range map[string]func(c *PendingChoice){
		"a route's then": func(c *PendingChoice) {
			c.replacementResume.ev.zoneRoute = &zoneRoute{then: func(*Game) error { return nil }}
		},
		"an entry tail's then": func(c *PendingChoice) {
			c.replacementResume.ev.entryTail = &entryTail{then: func(*Game, uuid.UUID) error { return nil }}
		},
		"an entry batch": func(c *PendingChoice) { c.replacementResume.ev.entryTail = &entryTail{batch: &entryBatch{}} },
		"a life tail":    func(c *PendingChoice) { c.replacementResume.ev.lifeTail = &lifeTail{} },
		"a counter tail": func(c *PendingChoice) {
			c.replacementResume.ev.Kind = RepEventCounter
			c.replacementResume.ev.counterTail = &counterTail{}
		},
		"a keyword action's then": func(c *PendingChoice) {
			c.replacementResume.ev.keywordAction = &keywordActionTail{then: func(*Game) error { return nil }}
		},
		"the Army you amassed": func(c *PendingChoice) {
			c.replacementResume.ev.keywordAction = &keywordActionTail{amassed: func(*Game, uuid.UUID) error { return nil }}
		},
		"an amass's Army": func(c *PendingChoice) {
			c.replacementResume.ev.keywordAction = &keywordActionTail{armyToken: Card{Name: "Army"}}
		},
		"a damage event": func(c *PendingChoice) { c.replacementResume.ev.Kind = RepEventDamage },
		"a test replacement": func(c *PendingChoice) {
			c.replacementResume.applicable = []activeReplacement{{id: testReplacementIDBase}}
		},
		"an unkeyed prompt kind": func(c *PendingChoice) { c.Kind = PendingChoiceSacrifice },
		"no gathered effect":     func(c *PendingChoice) { c.replacementResume.applicable = nil },
		"an entry card choice with a then": func(c *PendingChoice) {
			c.Kind = PendingChoiceEntryDiscardFromHand
			c.chooseCardsResume = &chooseCardsFrame{zone: ZoneHand, then: func(*Game, []uuid.UUID) error { return nil }}
		},
		"an entry card choice with no zone frame": func(c *PendingChoice) { c.Kind = PendingChoiceEntrySacrifice },
		"an entry card choice over the wrong zone": func(c *PendingChoice) {
			c.Kind = PendingChoiceEntrySacrifice
			c.chooseCardsResume = &chooseCardsFrame{zone: ZoneHand}
		},
		"a staged token's entry": func(c *PendingChoice) {
			c.replacementResume.ev.OldZone, c.replacementResume.ev.NewZone = "", ZoneBattlefield
			g.enteringTokens = []Card{{InstanceID: c.replacementResume.ev.CardID}}
		},
	} {
		c := base()
		mut(c)
		if key := g.replacementFrameKeyable(c); key != "" {
			t.Errorf("%s: the frame is keyed as %q", name, key)
		}
		g.enteringTokens = nil
	}
}

// bouncedCommanderTable bounces seat 0's commander and leaves the CR
// 903.9b prompt open.
func bouncedCommanderTable(t *testing.T) (*Game, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)
	g.mu.Lock()
	err := g.BounceToHandForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("BounceToHandForEffect: %v", err)
	}
	expectCommanderPrompt(t, g, owner)
	return g, cmdID
}

// A commander's CR 903.9b bounce prompt is a restore point, and the
// restored prompt is answered both ways as the live one is.
func TestABouncedCommandersPromptIsARestorePoint(t *testing.T) {
	for _, apply := range []bool{true, false} {
		g, cmdID := bouncedCommanderTable(t)
		snap := throughJSON(t, g.CaptureSnapshot())
		if !snap.Restorable() {
			t.Fatalf("the open bounce prompt is not a restore point: %+v", snap.Continuations)
		}
		if snap.PendingChoices[0].PickThen != ReplacementResumeMove || snap.PendingChoices[0].ReplacementResume == nil {
			t.Fatalf("the prompt was written without its frame: %+v", snap.PendingChoices[0])
		}
		restored, err := snap.RestoreStrict()
		if err != nil {
			t.Fatalf("RestoreStrict: %v", err)
		}
		owner := restored.Seats[0]
		prompt := expectCommanderPrompt(t, restored, owner)
		if prompt.PickThen != "" {
			t.Errorf("the live prompt carries the tail key %q", prompt.PickThen)
		}
		if got := prompt.CommanderHeadedFor(); got != ZoneHand {
			t.Errorf("the restored prompt says the commander is headed for %q, want the hand", got)
		}
		if err := restored.ResolveOptionalReplacement(prompt.ID, owner.ID, apply); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
		if apply {
			assertOnlyIn(t, cmdID, owner.Command, owner.Hand, restored.Battlefield)
		} else {
			assertOnlyIn(t, cmdID, owner.Hand, owner.Command, restored.Battlefield)
		}
	}
}

// The rollback case: a binary that does not know the tail refuses the
// file rather than restoring a prompt with nothing behind it.
func TestAKeyedFrameWithAnUnknownTailIsRefused(t *testing.T) {
	g, _ := bouncedCommanderTable(t)
	for name, mut := range map[string]func(c *pendingChoiceSnapshot){
		"unknown tail":       func(c *pendingChoiceSnapshot) { c.PickThen = "replacement-resume/never-registered" },
		"tail of other kind": func(c *pendingChoiceSnapshot) { c.PickThen = ReplacementResumeDraw },
		"no tail":            func(c *pendingChoiceSnapshot) { c.PickThen = "" },
		"unknown draw body": func(c *pendingChoiceSnapshot) {
			c.ReplacementResume.Event.DrawInstead = &insteadRunSnapshot{Key: "never-registered"}
		},
		"unknown keyword-action prompt kind": func(c *pendingChoiceSnapshot) {
			c.ReplacementResume.Event.KeywordActionTail = &keywordActionTailSnapshot{Choice: "never-registered"}
		},
		"a counter tail over a move": func(c *pendingChoiceSnapshot) { c.PickThen = ReplacementResumeCounter },
		"a frame on an unkeyed prompt kind": func(c *pendingChoiceSnapshot) {
			c.Kind = PendingChoiceSacrifice
		},
	} {
		snap := throughJSON(t, g.CaptureSnapshot())
		mut(&snap.PendingChoices[0])
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: restore err = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}

// An effect the running binary no longer gathers is restored as one
// that no longer applies (CR 616.1f): a "yes" fires nothing, and the
// window goes on with the effects the board still has. Here the file
// names an effect that does not exist, so the commander's own CR 903.9b
// replacement is still unasked and is asked next.
func TestALostEffectInARestoredFrameNoLongerApplies(t *testing.T) {
	g, cmdID := bouncedCommanderTable(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	snap.PendingChoices[0].ReplacementResume.Applicable = []ReplacementEffectID{builtinReplacementIDBase + 999}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	owner := restored.Seats[0]
	prompt := expectCommanderPrompt(t, restored, owner)
	if f := prompt.replacementResume; f == nil || len(f.applicable) != 1 || !f.applicable[0].lost {
		t.Fatalf("the unknown effect was not restored as a lost one: %+v", f)
	}
	if err := restored.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if owner.Command.Contains(cmdID) {
		t.Fatal("a lost effect's yes moved the commander to the command zone")
	}
	again := expectCommanderPrompt(t, restored, owner)
	if again.ID == prompt.ID || again.ReplacementEffectIDs[0] != builtinReplacementIDBase {
		t.Fatalf("the window did not go on to the commander's own replacement: %+v", again)
	}
	if err := restored.ResolveOptionalReplacement(again.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Hand, owner.Command, restored.Battlefield)
}
