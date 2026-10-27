package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// choose_source_keyed_test.go — #4253 slice 10c: a choose_source prompt's
// consequence is a registered key (RegisterChooseSourceThen) or the
// engine's own shield, so a table waiting on it is a restore point.

// chooseSourceProbeSeen is what the probe continuation was handed, one
// entry per run.
var chooseSourceProbeSeen []ChosenSource

var chooseSourceProbe = RegisterChooseSourceThen(testEffectKeyPrefix+"choose-source-probe", func(_ *Game, r ChosenSource) error {
	chooseSourceProbeSeen = append(chooseSourceProbeSeen, r)
	return nil
})

func openChooseSource(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceChooseSource {
			return c
		}
	}
	return nil
}

// The prompt carries its key and data and no frame, a strict restore keeps
// all of it, and the restored answer pins the source as the object it is
// now and hands the continuation the data.
func TestKeyedChooseSourceSurvivesARestore(t *testing.T) {
	chooseSourceProbeSeen = nil
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := pushColouredCreature(g, opp, "Red", []string{"R"})
	_ = pushColouredCreature(g, opp, "Green", []string{"G"})
	source := uuid.New()
	item := numberProbeItem(me.ID, source, uuid.New())
	g.WithWriteLock(func() {
		queued, err := g.ChooseDamageSourceThenForEffect(ChooseSourcePrompt{
			Chooser: me.ID, Source: source, Question: "choose a red source",
			Queries: []PermanentQuery{{Colors: []string{"R"}}},
			Then:    chooseSourceProbe,
			Carry:   []uuid.UUID{{1}, {2}},
			Word:    "probe",
			Nums:    []int{3, 4},
			Asked:   NewAskedItem(item),
		})
		if err != nil || !queued {
			t.Fatalf("queued %v, err %v", queued, err)
		}
	})
	c := openChooseSource(g)
	if c == nil || c.PickThen != chooseSourceProbe.Key() {
		t.Fatalf("prompt = %+v, want the probe's key", c)
	}
	if snap := g.CaptureSnapshot(); snap.Continuations.ChoiceResumeFrames != 0 {
		t.Fatalf("the prompt counts a closure: %+v", snap.Continuations)
	}

	g = restoreStrictForTest(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	c = openChooseSource(g)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != red {
		t.Fatalf("restored prompt = %+v, want the red creature", c)
	}
	if err := g.ResolveChooseSource(c.ID, opp.ID, []uuid.UUID{red}); err != ErrNotTheChooser {
		t.Errorf("another seat answered: err %v, want ErrNotTheChooser", err)
	}
	if err := g.ResolveChooseSource(c.ID, me.ID, []uuid.UUID{red}); err != nil {
		t.Fatal(err)
	}
	if len(chooseSourceProbeSeen) != 1 {
		t.Fatalf("the continuation ran %d times, want once", len(chooseSourceProbeSeen))
	}
	r := chooseSourceProbeSeen[0]
	if r.Ref.ID != red || r.Zone != ZoneBattlefield || r.Chooser != me.ID || r.Source != source {
		t.Errorf("handed %+v in %q for chooser %v source %v", r.Ref, r.Zone, r.Chooser, r.Source)
	}
	if !reflect.DeepEqual(r.Carry, []uuid.UUID{{1}, {2}}) || r.Word != "probe" || !reflect.DeepEqual(r.Nums, []int{3, 4}) {
		t.Errorf("handed data %v %q %v", r.Carry, r.Word, r.Nums)
	}
	if r.Item == nil || r.Item.ID != item.ID || r.Item.XValue != 2 || r.Item.Controller != me.ID {
		t.Errorf("handed item %+v, want the asking item rebuilt from data", r.Item)
	}
}

// A chooser who leaves while the prompt is open gives the continuation the
// zero ref (CR 800.4f), restored or not.
func TestKeyedChooseSourceDropGivesTheZeroRef(t *testing.T) {
	chooseSourceProbeSeen = nil
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[0]
	_ = pushColouredCreature(g, g.Seats[1], "Red", []string{"R"})
	// The card that asked is another seat's, so it outlives the chooser
	// (CR 800.4a) and the rest of the card still runs.
	source := departureTestSource(g, g.Seats[1].ID, "Probe")
	g.WithWriteLock(func() {
		if queued, err := g.ChooseDamageSourceThenForEffect(ChooseSourcePrompt{
			Chooser: me.ID, Source: source, Then: chooseSourceProbe,
		}); err != nil || !queued {
			t.Fatalf("queued %v, err %v", queued, err)
		}
	})
	g = restoreStrictForTest(t, g)
	if err := g.Concede(g.Seats[0].ID); err != nil {
		t.Fatal(err)
	}
	if openChooseSource(g) != nil {
		t.Error("a prompt outlived its chooser")
	}
	if len(chooseSourceProbeSeen) != 1 || chooseSourceProbeSeen[0].Ref.ID != uuid.Nil {
		t.Fatalf("the continuation saw %+v, want one run with the zero ref", chooseSourceProbeSeen)
	}
}

// A prompt that names neither a continuation nor a shield, or both, is a
// card bug and queues nothing.
func TestChooseSourceNeedsExactlyOneConsequence(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	_ = pushColouredCreature(g, me, "Red", []string{"R"})
	g.WithWriteLock(func() {
		if queued, err := g.ChooseDamageSourceThenForEffect(ChooseSourcePrompt{Chooser: me.ID}); err == nil || queued {
			t.Errorf("no consequence: queued %v err %v, want a refusal", queued, err)
		}
		if queued, err := g.ChooseDamageSourceThenForEffect(ChooseSourcePrompt{
			Chooser: me.ID, Then: chooseSourceProbe, Shield: &SourceShield{Next: &NextDamageShield{}},
		}); err == nil || queued {
			t.Errorf("both: queued %v err %v, want a refusal", queued, err)
		}
	})
	if openChooseSource(g) != nil {
		t.Error("a refused prompt was queued")
	}
}

// The engine's own consequence: a next-damage shield, a source shield and a
// redirection, each described by the card before the question, survive a
// restore with their data, and the answer registers them against the
// pinned source.
func TestChooseSourceShieldsSurviveARestore(t *testing.T) {
	red := []PermanentQuery{{Colors: []string{"R"}}}
	for name, build := range map[string]func(me, opp uuid.UUID) SourceShield{
		"next": func(me, _ uuid.UUID) SourceShield {
			return SourceShield{Next: &NextDamageShield{
				EffectSource: uuid.New(), Controller: me, ProtectPlayer: me,
				ProtectTypes: []string{"creature"}, Queries: red, Half: true, Label: "next",
			}}
		},
		"prevent": func(me, _ uuid.UUID) SourceShield {
			return SourceShield{Prevent: &DamageShield{
				EffectSource: uuid.New(), Controller: me, ProtectPlayer: me, Queries: red, Label: "prevent",
				Amount: 2, CombatOnly: true, Duration: Duration{Kind: UntilEndOfTurn},
			}}
		},
		"redirect": func(me, opp uuid.UUID) SourceShield {
			return SourceShield{Redirect: &DamageRedirection{
				EffectSource: uuid.New(), Controller: me, ProtectPlayer: me, To: opp, Queries: red, Label: "redirect",
			}}
		},
	} {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		source := pushColouredCreature(g, opp, "Red", []string{"R"})
		shield := build(me.ID, opp.ID)
		g.WithWriteLock(func() {
			if queued, err := g.ChooseDamageSourceThenForEffect(ChooseSourcePrompt{
				Chooser: me.ID, Source: uuid.New(), Queries: red, Shield: &shield,
			}); err != nil || !queued {
				t.Fatalf("%s: queued %v, err %v", name, queued, err)
			}
		})
		before := len(g.ScopedEffects)
		g = restoreStrictForTest(t, g)
		me = g.Seats[0]
		c := openChooseSource(g)
		if c == nil || c.PickThen != chooseSourceShieldThen.Key() || c.sourceShield == nil {
			t.Fatalf("%s: restored prompt = %+v, want the shield and its key", name, c)
		}
		if !reflect.DeepEqual(c.sourceShield, cloneSourceShield(&shield)) {
			t.Errorf("%s: restored as %+v, want %+v", name, c.sourceShield, shield)
		}
		if err := g.ResolveChooseSource(c.ID, me.ID, []uuid.UUID{source}); err != nil {
			t.Fatalf("%s: answer: %v", name, err)
		}
		if got := len(g.ScopedEffects); got != before+1 {
			t.Errorf("%s: %d scoped effects after the answer, want %d", name, got, before+1)
		}
	}
}

// Every field of the three shields survives the wire, the follow-up body's
// key included (BodyRef hides it from JSON, so each wire form carries it).
func TestSourceShieldWireRoundTrips(t *testing.T) {
	who, other := uuid.New(), uuid.New()
	red := []PermanentQuery{{Colors: []string{"R"}, Types: []string{"creature"}}}
	for name, shield := range map[string]SourceShield{
		"next": {Next: &NextDamageShield{
			EffectSource: uuid.New(), Controller: who, Source: ObjectRef{ID: other, Epoch: 3}, SourceZone: ZoneGraveyard,
			Queries: red, ProtectPlayer: who, ProtectTypes: []string{"creature"}, ProtectPermanent: other,
			Then: BodyRef{key: "probe/then"}, Half: true, CombatOnly: true, Label: "next",
		}},
		"prevent": {Prevent: &DamageShield{
			EffectSource: uuid.New(), Controller: who, Queries: red,
			Filter:            DamageSourceFilter{PowerBounded: true, PowerAtMost: 2, Colorless: true, NoCounters: []string{"+1/+1"}, Except: red, ExceptObjects: []ObjectRef{{ID: other, Epoch: 1}}},
			ProtectPermanents: []uuid.UUID{who, other}, Recipients: DamageRecipientFilter{Players: true, Match: red},
			AndDealtBy: false, CombatOnly: true, Amount: 3, Then: BodyRef{key: "probe/then"}, To: other,
			ThenPerSource: true, UntilYourNextTurn: true, Duration: Duration{Kind: UntilEndOfTurn, Player: who}, Label: "prevent",
		}},
		"redirect": {Redirect: &DamageRedirection{
			EffectSource: uuid.New(), Controller: who, Queries: red, SourceFilter: DamageSourceFilter{Except: red},
			ProtectPlayer: who, ProtectTypes: []string{"land"}, ProtectSet: DamageRecipientFilter{Permanents: true, Match: red},
			CombatOnly: true, Next: true, Amount: 2, To: other, ToSourceController: true,
			Then: BodyRef{key: "probe/then"}, Optional: true, UntilYourNextTurn: true, Label: "redirect",
		}},
	} {
		raw, err := json.Marshal(snapshotSourceShield(&shield))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var wire sourceShieldSnapshot
		if err := json.Unmarshal(raw, &wire); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := restoreSourceShield(&wire); !reflect.DeepEqual(got, &shield) {
			t.Errorf("%s: restored %+v, want %+v", name, got, shield)
		}
		if keys := unknownSourceShieldKeys(&wire); len(keys) != 1 {
			t.Errorf("%s: an unregistered follow-up body names %v, want it refused", name, keys)
		}
	}
}
