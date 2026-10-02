package game

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// prevent_next_from_source_test.go — ADR 0107 §6 (#1860): the shield
// against the next damage from a source (CR 615.8), its CR 615.9
// recheck, the source choice (CR 609.7a) and the CR 615.5 follow-up.

// testFollowUp registers a follow-up body that records what it was
// handed.
type followUpCall struct {
	amount     int
	controller uuid.UUID
	source     uuid.UUID
	damageFrom uuid.UUID
	colors     []string
}

func testFollowUp(calls *[]followUpCall) BodyRef {
	key := fmt.Sprintf("%sfollow-up-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	return DelayedBody(key, func(g *Game, item *StackItem, p EffectParams) error {
		c := followUpCall{amount: p.Amount, controller: item.Controller, source: item.SourceCardID, damageFrom: p.Object.ID}
		if item.Trigger != nil {
			c.colors = item.Trigger.Event.Colors
		}
		*calls = append(*calls, c)
		return nil
	})
}

// shieldAgainst registers a shield protecting `player` against the
// object `source` is right now.
func shieldAgainst(t *testing.T, g *Game, player, source uuid.UUID, then BodyRef, qs ...PermanentQuery) {
	t.Helper()
	g.WithWriteLock(func() {
		ref, zone, ok := g.DamageSourceRefLocked(source)
		if !ok {
			t.Fatal("source in no zone")
		}
		if !g.PreventNextDamageFromSourceForEffect(NextDamageShield{
			Controller: player, Source: ref, SourceZone: zone, Queries: qs,
			ProtectPlayer: player, Then: then, Label: "test shield",
		}) {
			t.Fatal("no shield registered")
		}
	})
}

func lifeOf(g *Game, id uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.playerByIDLocked(id).Life })
	return n
}

// CR 615.8: the next instance of damage from the chosen source is
// prevented whatever its size; any later instance is dealt normally, and
// another source's damage is never touched.
func TestNextDamageShieldPreventsTheNextInstanceFromItsSource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	other := pushColouredCreature(g, opp, "Other", []string{"R"})
	shieldAgainst(t, g, me.ID, dragon, BodyRef{})
	start := lifeOf(g, me.ID)

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(other, me.ID, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToPlayerForEffect(dragon, me.ID, 7); err != nil {
			t.Fatal(err)
		}
	})
	if got := lifeOf(g, me.ID); got != start-1 {
		t.Fatalf("life %d, want %d: the other source's 1 dealt, the dragon's 7 prevented", got, start-1)
	}
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		if len(g.ScopedEffects) != 0 {
			t.Fatalf("spent shield outlived its instance: %+v", g.ScopedEffects)
		}
		if err := g.DealDamageToPlayerForEffect(dragon, me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Errorf("life %d, want %d: the dragon's next instance is dealt normally", got, start-3)
	}
}

// One instance, several events: an attacker that deals combat damage to
// two things at once is one instance of damage (CR 615.8). A shield with
// nothing protected prevents all of it, and only it.
func TestNextDamageShieldCoversEveryEventOfOneInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trampler := pushColouredCreature(g, opp, "Trampler", []string{"G"})
	blocker := pushColouredCreature(g, me, "Blocker", []string{"W"})
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(trampler)
		g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, Source: ref, SourceZone: zone, Label: "Awe Strike"})
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		g.markCombatDamageOnCardLocked(blocker, 2, trampler, CombatStepRegular)
		g.markCombatDamageToPlayerLocked(me.ID, trampler, 4, CombatStepRegular)
	})
	if got := damageOn(g, blocker); got != 0 {
		t.Errorf("blocker has %d damage, want 0", got)
	}
	if got := lifeOf(g, me.ID); got != start {
		t.Errorf("life %d, want %d: the trampled-over damage is the same instance", got, start)
	}
}

// CR 615.9 / 609.7b: the property is rechecked when the source would deal
// damage. A source that no longer matches gets through and does not use
// the shield up; once it matches again, the shield works.
func TestNextDamageShieldRechecksTheSourcesProperty(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Shifter", []string{"R"})
	shieldAgainst(t, g, me.ID, src, BodyRef{}, PermanentQuery{Colors: []string{"R"}})
	start := lifeOf(g, me.ID)
	setColour := func(c string) {
		g.WithWriteLock(func() {
			findBattlefieldCard(g, src).Colors = []string{c}
			g.layerVersion.Add(1)
			g.RecomputeLayersIfStaleLocked()
		})
	}
	setColour("U")
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d: a blue source is not a red one", got, start-2)
	}
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Mods[0].SpentBatch != 0 {
		t.Fatalf("shield after a failed recheck = %+v, want it unspent", g.ScopedEffects)
	}
	setColour("R")
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 5)
	})
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Errorf("life %d, want %d: red again, so the 5 is prevented", got, start-2)
	}
}

// CR 609.7a: "If the player chooses a permanent spell, the effect will
// apply to any damage dealt by that spell and any damage dealt by the
// permanent that spell becomes." And CR 400.7: a permanent that left and
// came back is a new object the shield does not know.
func TestNextDamageShieldFollowsAPermanentSpellOntoTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := NewCard("Incoming", opp.ID)
	spell.TypeLine = "Creature — Elemental"
	spell.Power, spell.Toughness = 3, 3
	g.WithWriteLock(func() {
		g.Stack.PushTop(spell)
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*StackItem{}
		}
		g.StackMeta[spell.InstanceID] = &StackItem{ID: spell.InstanceID, Kind: StackItemSpell, Controller: opp.ID}
	})
	shieldAgainst(t, g, me.ID, spell.InstanceID, BodyRef{})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneStack}, ZoneRef{Kind: ZoneBattlefield}, spell.InstanceID); err != nil {
		t.Fatal(err)
	}
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(spell.InstanceID, me.ID, 3) })
	if got := lifeOf(g, me.ID); got != start {
		t.Fatalf("life %d, want %d: the permanent the spell became is the chosen source", got, start)
	}

	// A battlefield source that flickers is a new object.
	bear := pushColouredCreature(g, opp, "Bear", []string{"G"})
	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	shieldAgainst(t, g, me.ID, bear, BodyRef{})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneExile}, bear); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneExile}, ZoneRef{Kind: ZoneBattlefield}, bear); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, me.ID, 2) })
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Errorf("life %d, want %d: a flickered creature is a new object (CR 400.7)", got, start-2)
	}
}

// CR 615.5: the follow-up runs "immediately afterward" — as the instance
// of damage ends (ADR 0108 PR 0), not at the next priority boundary —
// with the amount prevented, the shield's controller and source, and the
// damage source's colours as it dealt the damage.
func TestNextDamageShieldRunsItsFollowUpWithTheAmountPrevented(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Red One", []string{"R"})
	var calls []followUpCall
	shieldAgainst(t, g, me.ID, src, testFollowUp(&calls))
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 4) })
	if len(calls) != 1 {
		t.Fatalf("follow-up ran %d times as the instance ended, want 1", len(calls))
	}
	c := calls[0]
	if c.amount != 4 || c.controller != me.ID || c.damageFrom != src || len(c.colors) != 1 || c.colors[0] != "R" {
		t.Errorf("follow-up got %+v, want 4 prevented from the red source, for %v", c, me.ID)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if len(calls) != 1 {
		t.Errorf("follow-up ran %d times, want exactly once", len(calls))
	}
}

// "The damage prevented this way" is the total for the INSTANCE: an
// attacker with trample blocked by two deals three events of combat
// damage at once, and a shield against it runs its follow-up once with
// 5, not three times (CR 615.5, 615.8).
func TestNextDamageShieldFollowUpRunsOncePerInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trampler := pushColouredCreature(g, opp, "Trampler", []string{"G"})
	b1 := pushColouredCreature(g, me, "Blocker One", []string{"W"})
	b2 := pushColouredCreature(g, me, "Blocker Two", []string{"W"})
	var calls []followUpCall
	then := testFollowUp(&calls)
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(trampler)
		g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, Source: ref, SourceZone: zone, Then: then})
		g.markCombatDamageOnCardLocked(b1, 2, trampler, CombatStepRegular)
		g.markCombatDamageOnCardLocked(b2, 2, trampler, CombatStepRegular)
		g.markCombatDamageToPlayerLocked(me.ID, trampler, 1, CombatStepRegular)
		g.runStateChecksLocked()
	})
	if len(calls) != 1 || calls[0].amount != 5 {
		t.Errorf("follow-up calls %+v, want one with all 5 prevented", calls)
	}
	// A new instance, after play moves on, is not the spent shield's.
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		g.markCombatDamageToPlayerLocked(me.ID, trampler, 3, CombatStepRegular)
		g.runStateChecksLocked()
	})
	if len(calls) != 1 {
		t.Errorf("follow-up ran again for a later instance: %+v", calls)
	}
}

// The owed follow-up rides a restore point taken mid-instance — a combat
// damage step's instance is owed until a player would next receive
// priority — and the restored game runs it.
func TestOwedFollowUpSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	var calls []followUpCall
	shieldAgainst(t, g, me.ID, src, testFollowUp(&calls))
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(me.ID, src, 3, CombatStepRegular) })
	snap := g.CaptureSnapshot()
	if snap == nil || len(snap.PreventionFollowUps) != 1 || snap.PreventionFollowUps[0].Instance == 0 {
		t.Fatalf("snapshot owes %v, want the one follow-up, with its instance", snap)
	}
	r, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	r.WithWriteLock(func() { r.runStateChecksLocked() })
	if len(calls) != 1 || calls[0].amount != 3 {
		t.Errorf("restored game ran %+v, want one follow-up of 3", calls)
	}
}

// CR 615.12: applied to damage that can't be prevented, the shield
// prevents nothing, is not used up (CR 609.7b), and its additional effect
// still happens — with nothing prevented.
func TestNextDamageShieldUnderUnpreventableDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"B"})
	var calls []followUpCall
	shieldAgainst(t, g, me.ID, src, testFollowUp(&calls))
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		// Two unpreventable events of one instance: still one follow-up.
		if err := g.DamageInstanceForEffect(func() error {
			for _, n := range []int{2, 1} {
				if err := g.DealMarkedDamageForEffect(src, nil, me.ID, n, DamageMarks{CantBePrevented: true}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		g.runStateChecksLocked()
	})
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Fatalf("life %d, want %d: unpreventable damage is dealt", got, start-3)
	}
	if len(calls) != 1 || calls[0].amount != 0 {
		t.Fatalf("follow-up calls %+v, want one with 0 prevented", calls)
	}
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Mods[0].SpentBatch != 0 {
		t.Fatalf("shield = %+v, want it unspent", g.ScopedEffects)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Errorf("life %d, want %d: the unspent shield takes the next, preventable damage", got, start-3)
	}
}

// A shield with no chosen source and a property is "the next time a
// creature of the chosen type would deal damage to you" (Circle of
// Solace): the first matching source spends it.
func TestNextDamageShieldFromAnyMatchingSource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := NewCard("Elf", opp.ID)
	elf.TypeLine = "Creature — Elf Warrior"
	elf.Power, elf.Toughness = 1, 1
	g.Battlefield.PushTop(elf)
	goblin := NewCard("Goblin", opp.ID)
	goblin.TypeLine = "Creature — Goblin"
	goblin.Power, goblin.Toughness = 1, 1
	g.Battlefield.PushTop(goblin)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, ProtectPlayer: me.ID,
			Queries: []PermanentQuery{{Types: []string{"creature"}, Subtypes: []string{"Elf"}}}}) {
			t.Fatal("no shield")
		}
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(goblin.InstanceID, me.ID, 1)
		_ = g.DealDamageToPlayerForEffect(elf.InstanceID, me.ID, 2)
	})
	if got := lifeOf(g, me.ID); got != start-1 {
		t.Errorf("life %d, want %d: the Goblin's 1 dealt, the Elf's 2 prevented", got, start-1)
	}
}

// Shadowbane: "you and/or creatures you control". Damage to an opponent's
// creature is not protected.
func TestNextDamageShieldProtectsYouAndYourCreatures(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"B"})
	mine := pushColouredCreature(g, me, "Mine", []string{"W"})
	theirs := pushColouredCreature(g, opp, "Theirs", []string{"W"})
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, Source: ref, SourceZone: zone,
			ProtectPlayer: me.ID, ProtectTypes: []string{"creature"}})
		_ = g.DealDamageToCreatureForEffect(src, theirs, 1)
		_ = g.DealDamageToCreatureForEffect(src, mine, 1)
	})
	if damageOn(g, theirs) != 1 || damageOn(g, mine) != 0 {
		t.Errorf("theirs %d, mine %d: want 1 and 0", damageOn(g, theirs), damageOn(g, mine))
	}
}

// The shield is a ScopedEffect record of plain data: it survives a
// restore point, spent batch and all.
func TestNextDamageShieldSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	shieldAgainst(t, g, me.ID, src, BodyRef{}, PermanentQuery{Colors: []string{"R"}})
	snap := g.CaptureSnapshot()
	if snap == nil {
		t.Fatal("not a restore point")
	}
	r, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	start := lifeOf(r, me.ID)
	r.WithWriteLock(func() { _ = r.DealDamageToPlayerForEffect(src, me.ID, 3) })
	if got := lifeOf(r, me.ID); got != start {
		t.Errorf("life after restore %d, want %d", got, start)
	}
}

// CR 609.7a's candidate list: permanents, spells, face-up command-zone
// cards and objects something refers to — but never a card in a hand.
func TestDamageSourceCandidates(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := pushColouredCreature(g, opp, "Red", []string{"R"})
	blue := pushColouredCreature(g, opp, "Blue", []string{"U"})
	bolt := pushRedSpell(g, opp)
	cmdr := NewCard("Commander", opp.ID)
	cmdr.TypeLine = "Legendary Creature — Human"
	cmdr.Colors = []string{"R"}
	opp.Command.PushTop(cmdr)
	inHand := NewCard("Hidden", opp.ID)
	inHand.Colors = []string{"R"}
	opp.Hand.PushTop(inHand)
	dead := NewCard("Dead", opp.ID)
	dead.TypeLine = "Creature — Zombie"
	dead.Colors = []string{"R"}
	opp.Graveyard.PushTop(dead)
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*StackItem{}
		}
		g.StackMeta[bolt] = &StackItem{ID: bolt, Kind: StackItemSpell, Controller: opp.ID}
		trig := uuid.New()
		g.StackMeta[trig] = &StackItem{ID: trig, Kind: StackItemTriggered, Controller: opp.ID, SourceCardID: dead.InstanceID}
		g.DelayedTriggers = append(g.DelayedTriggers, &DelayedTrigger{ID: uuid.New(), Controller: opp.ID, SourceCardID: inHand.InstanceID})
	})
	var all, reds []uuid.UUID
	g.WithWriteLock(func() {
		all = g.DamageSourceCandidatesLocked(nil)
		reds = g.DamageSourceCandidatesLocked([]PermanentQuery{{Colors: []string{"R"}}})
	})
	has := func(list []uuid.UUID, id uuid.UUID) bool {
		for _, x := range list {
			if x == id {
				return true
			}
		}
		return false
	}
	for _, id := range []uuid.UUID{red, blue, bolt, cmdr.InstanceID} {
		if !has(all, id) {
			t.Errorf("candidates %v miss %v", all, id)
		}
	}
	if has(all, inHand.InstanceID) {
		t.Error("a card in a hand is never a candidate, even one a delayed trigger names")
	}
	if !has(all, dead.InstanceID) {
		t.Error("the source of a trigger on the stack is a candidate wherever it is (CR 609.7a)")
	}
	if has(reds, blue) || !has(reds, red) || !has(reds, bolt) {
		t.Errorf("red candidates %v: want the red creature and the bolt, not the blue creature", reds)
	}
	_ = me
}

// The choice end to end: the prompt offers the candidates, the answer is
// pinned as the object it is now, and Then gets it.
func TestChooseDamageSourcePinsTheAnswer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := pushColouredCreature(g, opp, "Red", []string{"R"})
	_ = pushColouredCreature(g, opp, "Green", []string{"G"})
	var got ObjectRef
	var gotZone ZoneKind
	g.WithWriteLock(func() {
		queued, err := g.ChooseDamageSourceThenForEffect(ChooseSourcePrompt{
			Chooser: me.ID, Question: "choose a red source",
			Queries: []PermanentQuery{{Colors: []string{"R"}}},
			Then: func(_ *Game, ref ObjectRef, zone ZoneKind) error {
				got, gotZone = ref, zone
				return nil
			},
		})
		if err != nil || !queued {
			t.Fatalf("queued %v, err %v", queued, err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d choices open, want 1", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceChooseSource || len(c.ChooseCards) != 1 || c.ChooseCards[0] != red {
		t.Fatalf("prompt = %s over %v, want choose_source over the red creature", c.Kind, c.ChooseCards)
	}
	if !ChoiceBlocksTable(c.Kind) {
		t.Error("choose_source must block the table: the shield is the rest of the card")
	}
	if err := g.ResolveChooseSource(c.ID, opp.ID, []uuid.UUID{red}); err != ErrNotTheChooser {
		t.Errorf("another seat answered: err %v, want ErrNotTheChooser", err)
	}
	if err := g.ResolveChooseSource(c.ID, me.ID, []uuid.UUID{red}); err != nil {
		t.Fatal(err)
	}
	if got.ID != red || gotZone != ZoneBattlefield {
		t.Errorf("Then got %+v in %q, want the red creature on the battlefield", got, gotZone)
	}
}

// No legal source: nothing is asked, and Then still runs, with nothing.
func TestChooseDamageSourceWithNoCandidate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ran := false
	g.WithWriteLock(func() {
		queued, err := g.ChooseDamageSourceThenForEffect(ChooseSourcePrompt{
			Chooser: me.ID, Queries: []PermanentQuery{{Colors: []string{"R"}}},
			Then: func(_ *Game, ref ObjectRef, _ ZoneKind) error {
				ran = ref.ID == uuid.Nil
				return nil
			},
		})
		if err != nil || queued {
			t.Fatalf("queued %v, err %v; want nothing asked", queued, err)
		}
	})
	if !ran {
		t.Error("Then did not run with the zero ref")
	}
}
