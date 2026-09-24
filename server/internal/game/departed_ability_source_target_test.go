package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// departed_ability_source_target_test.go covers #1429 (ADR 0072
// amendment 2026-09-24 "#1429", ADR 0018 note 2026-09-24): when an
// ABILITY resolves after its source has left the battlefield, the
// CR 608.2b target re-check judges protection against the source as it
// last existed on the battlefield (CR 608.2h), not against the card in
// the zone it went to.
//
// Every scenario uses #1417's layer 5 painter and bleacher
// (departed_source_protection_test.go), which apply on the battlefield
// only, so the source's last-known colour and its graveyard colour
// DIFFER, and asserts that difference in setup. Without the fix the
// graveyard card's printed colour decides: a painted-red source's
// ability resolves through protection from red, and a bleached red
// source's ability fizzles against it.

// pushTargetedAbilityItem puts an ability item of `kind` from `src` on
// top of the stack, targeting `victim` under "target creature", and
// returns a pointer to a flag its Effect sets. The item is stamped the
// way the real paths stamp it, read NOW: StackItem.SourceObject is the
// object the card is (#1418, sourceObjectRefLocked), and an activated
// item also gets the post-cost SourceEpoch. A test that wants a
// different stamp — the pre-cost object of a source sacrificed to its
// own ability, or no stamp at all (a pre-#1418 snapshot) — overwrites
// it with restamp.
func pushTargetedAbilityItem(t *testing.T, g *Game, kind StackItemKind, src, controller, victim uuid.UUID) (*StackItem, *bool) {
	t.Helper()
	resolved := new(bool)
	var item *StackItem
	g.WithWriteLock(func() {
		item = &StackItem{
			ID:           uuid.New(),
			Kind:         kind,
			Controller:   controller,
			Owner:        controller,
			SourceCardID: src,
			Label:        "#1429 probe — 2 damage to target creature",
			Targets:      []TargetRef{{Kind: TargetCard, ID: victim}},
			targetSpec:   anyCreatureSpec(),
			Effect: func(*Game, *StackItem) error {
				*resolved = true
				return nil
			},
			Seq: g.nextStackSeqLocked(),
		}
		item.SourceObject = g.sourceObjectRefLocked(src)
		if kind == StackItemActivated {
			item.SourceEpoch = g.cardObjectEpochLocked(src)
		}
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		g.StackMeta[item.ID] = item
	})
	return item, resolved
}

// restamp overwrites the item's SourceObject. The zero ref is an
// unstamped item, restored from a snapshot written before #1418.
func restamp(g *Game, item *StackItem, ref ObjectRef) {
	g.WithWriteLock(func() { item.SourceObject = ref })
}

// recheck is the CR 608.2b per-slot answer for the item's only target.
func recheck(g *Game, item *StackItem) bool {
	var ok bool
	g.WithWriteLock(func() { ok = g.TargetStillLegalForEffect(item, item.Targets[0]) })
	return ok
}

// A printed-colourless source painted red puts its ability on the stack
// at a creature, the creature gains protection from red, and the source
// dies (colourless again in the graveyard). The source was red as it
// last existed, so the target is illegal and the ability fizzles —
// for a trigger (no epoch; the instance-ID rule), an activation that
// left after it was announced (the SourceEpoch record), and an
// activation whose cost sacrificed its own source (post-cost stamp).
func TestDepartedAbilitySourceFizzlesAgainstProtectionFromItsLastKnownColour(t *testing.T) {
	for _, tc := range []struct {
		name          string
		kind          StackItemKind
		sacrificeCost bool
		unstamped     bool
	}{
		{"triggered", StackItemTriggered, false, false},
		{"activated", StackItemActivated, false, false},
		{"activated, source sacrificed as its cost", StackItemActivated, true, false},
		{"triggered, unstamped (pre-#1418 snapshot)", StackItemTriggered, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			painted := map[uuid.UUID]bool{}
			withLKIColourPainters(t, painted, nil)

			victim := pushColouredCreature(g, opp, "Victim", []string{"W"})
			src := pushLKISource(g, me.ID, nil)
			painted[src] = true
			pushLKIPainter(g, me.ID, lkiPainterOracle)
			if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
				t.Fatalf("setup: the painted source is %v on the battlefield, want red", got)
			}
			preCost := refOf(t, g, src)
			if tc.sacrificeCost {
				destroy(t, g, src)
			}
			item, resolved := pushTargetedAbilityItem(t, g, tc.kind, src, me.ID, victim)
			if tc.sacrificeCost {
				// The activation stamps the object BEFORE its costs
				// (#1418), and SourceEpoch after them.
				restamp(g, item, preCost)
			}
			if tc.unstamped {
				restamp(g, item, ObjectRef{})
			}
			if !tc.sacrificeCost {
				destroy(t, g, src)
			}
			if got := liveColours(t, g, src); len(got) != 0 {
				t.Fatalf("setup: the card in the graveyard is %v; it must be colourless "+
					"there, or this test proves nothing", got)
			}
			if !recheck(g, item) {
				t.Fatal("setup: the target is illegal before it has protection")
			}
			g.WithWriteLock(func() {
				c := findBattlefieldCard(g, victim)
				c.Keywords = append(c.Keywords, "protection from red")
				g.recomputeLayersLocked()
			})

			if recheck(g, item) {
				t.Error("CR 608.2b / 608.2h: the source was RED as it last existed, so a " +
					"pro-red target is illegal — the re-check read the colourless graveyard card (#1429)")
			}
			resolveTop(t, g)
			if *resolved {
				t.Error("the ability resolved; with its only target illegal it must fizzle")
			}
		})
	}
}

// The reverse: a printed-red source bleached colourless dies and is red
// again in the graveyard. It last existed colourless, so a pro-red
// target stays legal and the ability resolves.
func TestDepartedAbilitySourceResolvesAgainstProtectionFromItsGraveyardColour(t *testing.T) {
	for _, tc := range []struct {
		name      string
		kind      StackItemKind
		unstamped bool
	}{
		{"triggered", StackItemTriggered, false},
		{"activated", StackItemActivated, false},
		{"triggered, unstamped (pre-#1418 snapshot)", StackItemTriggered, true},
	} {
		kind := tc.kind
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			bleached := map[uuid.UUID]bool{}
			withLKIColourPainters(t, nil, bleached)

			victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
			src := pushLKISource(g, me.ID, []string{"R"})
			bleached[src] = true
			pushLKIPainter(g, me.ID, lkiBleacherOracle)
			if got := liveColours(t, g, src); len(got) != 0 {
				t.Fatalf("setup: the bleached source is %v on the battlefield, want colourless", got)
			}
			item, resolved := pushTargetedAbilityItem(t, g, kind, src, me.ID, victim)
			if tc.unstamped {
				restamp(g, item, ObjectRef{})
			}
			if !recheck(g, item) {
				t.Fatal("setup: a colourless source may target a pro-red creature")
			}
			destroy(t, g, src)
			if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
				t.Fatalf("setup: the card in the graveyard is %v, want red", got)
			}

			if !recheck(g, item) {
				t.Error("CR 608.2h: the source was colourless as it last existed, so the " +
					"pro-red target is still legal — the re-check read the red graveyard card (#1429)")
			}
			resolveTop(t, g)
			if !*resolved {
				t.Error("the ability fizzled; its target was legal for the source as it last existed")
			}
		})
	}
}

// CR 400.7: an activation's source died and came back as a new object,
// painted red. The ability is judged by the OLD object — colourless as
// it last existed — not by the live red card with the same instance ID.
func TestActivatedAbilityOfAReturnedSourceIsJudgedByTheDepartedObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	painted := map[uuid.UUID]bool{}
	withLKIColourPainters(t, painted, nil)

	victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	src := pushLKISource(g, me.ID, nil)
	item, resolved := pushTargetedAbilityItem(t, g, StackItemActivated, src, me.ID, victim)
	destroy(t, g, src)
	returnUnderOwner(t, g, src, me.ID)
	painted[src] = true
	pushLKIPainter(g, me.ID, lkiPainterOracle)
	if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
		t.Fatalf("setup: the new object is %v, want red", got)
	}

	if !recheck(g, item) {
		t.Error("CR 400.7: the returned red card is a new object and not the ability's source; " +
			"the colourless object that activated it may target a pro-red creature")
	}
	resolveTop(t, g)
	if !*resolved {
		t.Error("the ability fizzled against the new object's colour")
	}
}

// CR 400.7, the other half: a source SACRIFICED to its own ability
// (so SourceEpoch is its graveyard epoch) that has since come back as a
// new object. It was colourless as it last existed; the returned card is
// red. The ability is judged by the object it came from — the stamp
// names it before the cost was paid (#1418) — so a pro-red target stays
// legal. Without the stamp no record matches the post-cost epoch, the
// instance-ID rule does not answer for a card on the battlefield, and
// the re-check reads the live red card.
func TestSacrificedActivationSourceThatReturnedIsJudgedByTheSacrificedObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bleached := map[uuid.UUID]bool{}
	withLKIColourPainters(t, nil, bleached)

	victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	src := pushLKISource(g, me.ID, []string{"R"})
	bleached[src] = true
	pushLKIPainter(g, me.ID, lkiBleacherOracle)
	if got := liveColours(t, g, src); len(got) != 0 {
		t.Fatalf("setup: the bleached source is %v on the battlefield, want colourless", got)
	}
	preCost := refOf(t, g, src)
	destroy(t, g, src) // the sacrifice cost
	item, resolved := pushTargetedAbilityItem(t, g, StackItemActivated, src, me.ID, victim)
	restamp(g, item, preCost)
	delete(bleached, src)
	returnUnderOwner(t, g, src, me.ID)
	if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
		t.Fatalf("setup: the returned card is %v, want red", got)
	}

	if !recheck(g, item) {
		t.Error("CR 400.7: the ability came from the COLOURLESS object its cost sacrificed; " +
			"the red card on the battlefield is a new object and not its source")
	}
	resolveTop(t, g)
	if !*resolved {
		t.Error("the ability fizzled against the returned card's colour")
	}
}

// A trigger from a card IN A GRAVEYARD, whose card was a permanent
// earlier this turn. Its source is the graveyard card (red), not the
// permanent the card used to be (bleached colourless). The stamp names
// the graveyard object, which has no battlefield record, so the
// re-check reads the card where it is and a pro-red target is illegal.
// The instance-ID rule named the LAST permanent the card was, and let
// the target through.
func TestGraveyardTriggerFromACardThatDiedThisTurnIsJudgedAsTheGraveyardCard(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bleached := map[uuid.UUID]bool{}
	withLKIColourPainters(t, nil, bleached)

	victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	src := pushLKISource(g, me.ID, []string{"R"})
	bleached[src] = true
	pushLKIPainter(g, me.ID, lkiBleacherOracle)
	destroy(t, g, src)
	if got := liveColours(t, g, src); !slices.Equal(got, []string{"R"}) {
		t.Fatalf("setup: the card in the graveyard is %v, want red", got)
	}
	var matched bool
	g.WithWriteLock(func() {
		rec, ok := g.departedDamageSourceLocked(src, nil)
		matched = ok && len(rec.Characteristic.Colors) == 0
	})
	if !matched {
		t.Fatal("setup: the instance-ID rule must match the colourless record, or this test proves nothing")
	}

	// The trigger fires FROM the graveyard now (a Bloodghast-style
	// graveyard ability), so it is stamped with the graveyard object.
	item, resolved := pushTargetedAbilityItem(t, g, StackItemTriggered, src, me.ID, victim)
	var gy ObjectRef
	g.WithWriteLock(func() { gy = ObjectRef{ID: src, Epoch: g.cardObjectEpochLocked(src)} })
	if item.SourceObject != gy {
		t.Fatalf("setup: the trigger is stamped %+v, want the graveyard object %+v", item.SourceObject, gy)
	}

	if recheck(g, item) {
		t.Error("the trigger's source is the RED graveyard card, so a pro-red target is illegal; " +
			"it was judged as the colourless permanent the card used to be (#1418 / #1429)")
	}
	resolveTop(t, g)
	if *resolved {
		t.Error("the trigger resolved; with its only target illegal it must fizzle")
	}
}

// A live source is read live, exactly as before: the painted-red
// permanent's ability may not keep a pro-red target, the bleached one's
// may, for both kinds.
func TestLiveAbilitySourceTargetRecheckStillReadsLive(t *testing.T) {
	for _, kind := range []StackItemKind{StackItemTriggered, StackItemActivated} {
		t.Run(string(kind), func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			painted, bleached := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
			withLKIColourPainters(t, painted, bleached)

			victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
			red := pushLKISource(g, me.ID, nil)
			painted[red] = true
			colourless := pushLKISource(g, me.ID, []string{"R"})
			bleached[colourless] = true
			pushLKIPainter(g, me.ID, lkiPainterOracle)
			pushLKIPainter(g, me.ID, lkiBleacherOracle)

			fromRed, _ := pushTargetedAbilityItem(t, g, kind, red, me.ID, victim)
			fromColourless, _ := pushTargetedAbilityItem(t, g, kind, colourless, me.ID, victim)
			if recheck(g, fromRed) {
				t.Error("a live painted-red source's ability may not keep a pro-red target")
			}
			if !recheck(g, fromColourless) {
				t.Error("a live bleached (colourless) source's ability may keep a pro-red target")
			}

			// A record CAN exist for a live epoch — an exit a replacement
			// stopped part-way leaves one (rememberDepartingPermanentLocked).
			// Plant a colourless one for the live red source: the live
			// object is still the answer.
			g.WithWriteLock(func() {
				c := findBattlefieldCard(g, red)
				stale := permanentInfoOf(c)
				stale.Left = true
				stale.Characteristic.Colors = nil
				g.lastKnownPermanents = map[uuid.UUID][]PermanentInfo{red: {stale}}
			})
			if recheck(g, fromRed) {
				t.Error("a stale record at the LIVE object's epoch must not replace the live read")
			}
		})
	}
}

// A spell is its own source and is read on the stack, as before — even
// when its card has a battlefield record this turn that the
// instance-ID rule would otherwise match. The state is built by hand
// (a card whose epoch is exactly one past its record, sitting on the
// stack) because no real path produces it: every route onto the stack
// bumps the epoch. The guard is what keeps a spell from ever being read
// as a permanent it used to be.
func TestSpellSourceTargetRecheckStillReadsTheSpell(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bleached := map[uuid.UUID]bool{}
	withLKIColourPainters(t, nil, bleached)

	victim := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")

	// An ordinary red spell: illegal against pro-red.
	redSpell := NewCard("Red Spell", me.ID)
	redSpell.TypeLine = "Instant"
	redSpell.Colors = []string{"R"}
	// A red card that was a bleached (colourless) permanent earlier this
	// turn, now on the stack with its epoch one past the record.
	lapsed := pushLKISource(g, me.ID, []string{"R"})
	bleached[lapsed] = true
	pushLKIPainter(g, me.ID, lkiBleacherOracle)
	destroy(t, g, lapsed)

	items := map[string]*StackItem{}
	g.WithWriteLock(func() {
		g.Stack.PushTop(redSpell)
		c, err := me.Graveyard.Remove(lapsed)
		if err != nil {
			t.Fatalf("setup: the lapsed card is not in the graveyard: %v", err)
		}
		g.Stack.PushTop(c)
		for name, id := range map[string]uuid.UUID{"red": redSpell.InstanceID, "lapsed": lapsed} {
			items[name] = &StackItem{
				ID: id, Kind: StackItemSpell, Controller: me.ID, Owner: me.ID, SourceCardID: id,
				Targets:    []TargetRef{{Kind: TargetCard, ID: victim}},
				targetSpec: anyCreatureSpec(),
			}
		}
		if rec, ok := g.departedDamageSourceLocked(lapsed, nil); !ok || len(rec.Characteristic.Colors) != 0 {
			t.Fatalf("setup: the instance-ID rule must match a colourless record for the "+
				"lapsed card, or this test proves nothing (ok=%v)", ok)
		}
	})

	if recheck(g, items["red"]) {
		t.Error("a red spell may not keep a pro-red target")
	}
	if recheck(g, items["lapsed"]) {
		t.Error("a spell is judged as the RED spell it is, never as the colourless permanent " +
			"its card was earlier in the turn")
	}
}
