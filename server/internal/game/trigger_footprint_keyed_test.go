package game

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/uuid"
)

// trigger_footprint_keyed_test.go — ADR 0018's #3013 amendment, sub-PR
// 3. An engine keyword trigger's footprint is declared in
// keyedBodyFootprints, by hand, next to the body it describes. This is
// that table's conformance test, the keyed-body half of
// cards/effects/footprint_conformance_test.go: it resolves each body's
// real item on a board and fails on anything the body changes or emits
// outside what the check reads off its declaration.

// keyedFootprintCases builds, for each body in keyedBodyFootprints, the
// item the engine makes for it from a source on the battlefield.
var keyedFootprintCases = map[string]func(g *Game, src *Card) *StackItem{
	prowessPumpBody.key: func(g *Game, src *Card) *StackItem {
		return prowessTrigger.Build(Event{}, src, Characteristic{}, g)
	},
}

func TestEveryKeyedFootprintBodyHasAConformanceCase(t *testing.T) {
	for _, body := range KeyedFootprintBodies() {
		if _, ok := keyedFootprintCases[body]; !ok {
			t.Errorf("keyedBodyFootprints declares %q and keyedFootprintCases has no case for it", body)
		}
	}
	for body := range keyedFootprintCases {
		if _, ok := keyedBodyFootprints[body]; !ok {
			t.Errorf("keyedFootprintCases has %q, which keyedBodyFootprints does not declare", body)
		}
	}
}

// keyedConfState is what a keyed body could change, continuing effects
// aside (checked on their own): every player's life, counters and card
// zones, and every object's zone and state.
type keyedConfState struct {
	players map[uuid.UUID]string
	objects map[uuid.UUID]string
}

func captureKeyedConformance(g *Game) keyedConfState {
	s := keyedConfState{players: map[uuid.UUID]string{}, objects: map[uuid.UUID]string{}}
	add := func(z *Zone, zone string) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			s.objects[c.InstanceID] = fmt.Sprintf("%s|%s|%v|%v|%d|%v", zone, c.Controller, c.Counters, c.Tapped, c.DamageMarked, c.AttachedTo)
		}
	}
	add(g.Battlefield, "battlefield")
	add(g.Exile, "exile")
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		s.players[p.ID] = fmt.Sprintf("%d|%d|%v|%d|%d|%d", p.Life, p.Energy, p.Counters,
			zoneLen(p.Library), zoneLen(p.Hand), zoneLen(p.Graveyard))
		add(p.Library, "library")
		add(p.Hand, "hand")
		add(p.Graveyard, "graveyard")
	}
	return s
}

func zoneLen(z *Zone) int {
	if z == nil {
		return 0
	}
	return len(z.Cards)
}

func TestKeyedBodyFootprintsDoOnlyWhatTheyDeclare(t *testing.T) {
	bodies := KeyedFootprintBodies()
	sort.Strings(bodies)
	for _, body := range bodies {
		build := keyedFootprintCases[body]
		if build == nil {
			continue // reported by TestEveryKeyedFootprintBodyHasAConformanceCase
		}
		t.Run(body, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			id := pushBattlefieldForTest(g, me.ID, "Keyed Probe", "Creature — Monk", "")
			pushBattlefieldForTest(g, g.Seats[1].ID, "Bystander", "Creature — Bear", "")
			var item *StackItem
			g.WithWriteLock(func() {
				src := findBattlefieldCard(g, id)
				item = build(g, src)
			})
			item.ID = uuid.New()
			if item.Body != body {
				t.Fatalf("the case builds a %q item, want %q", item.Body, body)
			}

			var acc triggerAccess
			var ok bool
			var before keyedConfState
			var events, scoped int
			var err error
			g.WithWriteLock(func() {
				acc, ok = g.triggerAccessLocked(item)
				before = captureKeyedConformance(g)
				events, scoped = len(g.Events), len(g.ScopedEffects)
				err = item.Effect(g, item)
			})
			if !ok {
				t.Fatal("the check cannot describe the body's own item")
			}
			if err != nil {
				t.Fatalf("Effect: %v", err)
			}
			after := captureKeyedConformance(g)

			for _, ev := range g.Events[events:] {
				if !acc.emits[ev.Kind] {
					t.Errorf("emitted %q, which the footprint does not declare (declared %v)", ev.Kind, acc.emits)
				}
			}
			for p, b := range before.players {
				if after.players[p] != b {
					t.Errorf("player %s changed undeclared: %q → %q", p, b, after.players[p])
				}
			}
			for o, b := range before.objects {
				if a := after.objects[o]; a != b && !acc.writes[o] {
					t.Errorf("object %s changed undeclared: %q → %q", o, b, a)
				}
			}
			for o := range after.objects {
				if _, was := before.objects[o]; !was {
					t.Errorf("new object %s undeclared", o)
				}
			}
			// A new continuing effect may name only written objects.
			if len(g.ScopedEffects) == scoped {
				t.Error("the body registered nothing: the probe did nothing, so it proves nothing")
			}
			for _, se := range g.ScopedEffects[scoped:] {
				if se.Scope != ScopeNone {
					t.Errorf("registered a scoped rule %v, which the footprint does not declare", se.Scope)
				}
				for _, a := range se.Affected {
					if !acc.writes[a.ID] {
						t.Errorf("registered an effect on %s, which the footprint does not write", a.ID)
					}
				}
				for _, m := range se.Mods {
					if !characteristicModKinds[m.Kind] {
						t.Errorf("registered a %q mod, which is more than a characteristic change", m.Kind)
					}
				}
			}
			if !acc.reads[id] {
				t.Error("the footprint does not read its source")
			}
		})
	}
}

// A token template's keyword triggers are read off its keywords, the
// way the harvester reads them once the token is on the battlefield.
func TestTokenKeywordWatches(t *testing.T) {
	for _, c := range []struct {
		keywords []string
		want     []EventKind
	}{
		{nil, nil},
		{[]string{"flying"}, nil},
		{[]string{KeywordProwess}, []EventKind{EventCast}},
		{[]string{KeywordEvolve}, []EventKind{EventETB}},
	} {
		got := tokenKeywordWatches("", c.keywords)
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%v: watches %v, want %v", c.keywords, got, c.want)
		}
	}
}

// A card's printed keywords apply to the object itself and are no static
// (#4236), so a definition that has only them bears no rules; any static
// beside them still does.
func TestPrintedKeywordsBearNoRules(t *testing.T) {
	if global, own := defBearsRules(&CardDef{PrintedKeywords: []string{"flying"}}); global || own {
		t.Errorf("printed keywords alone: global %v own %v, want neither", global, own)
	}
	other := StaticAbility{Layer: Layer6Ability}
	if _, own := defBearsRules(&CardDef{PrintedKeywords: []string{"flying"}, Static: []StaticAbility{other}}); !own {
		t.Error("a static beside the printed keywords must still bear the object's own rules")
	}
}
