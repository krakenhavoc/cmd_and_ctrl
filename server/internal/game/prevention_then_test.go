package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// prevention_then_test.go — ADR 0108 §8 (#1906): a prevention static's
// additional effect (ReplacementEffect.Then / ThenPer), the apply loop's
// measure of what it prevented, CR 615.12's zero-prevented run, and the
// same follow-up on the scoped preventDamage and preventCombatDamage
// shields (owner decision 2).

// thenCall is what a follow-up body was handed.
type thenCall struct {
	prevented  int // the params' Amount: "prevented this way"
	damage     int // the handed event's Amount: "that damage"
	source     uuid.UUID
	object     ObjectRef
	controller uuid.UUID
	damageFrom uuid.UUID
	target     uuid.UUID
	to         []TargetRef
}

func thenRecorder(calls *[]thenCall) BodyRef {
	key := fmt.Sprintf("%sthen-%d", testEffectKeyPrefix, testEffectKeySeq.Add(1))
	return DelayedBody(key, func(g *Game, item *StackItem, p EffectParams) error {
		c := thenCall{
			prevented: p.Amount, source: item.SourceCardID, object: item.SourceObject,
			controller: item.Controller, damageFrom: p.Object.ID, to: item.Targets,
		}
		if item.Trigger != nil {
			c.damage = item.Trigger.Event.Amount
			c.target = item.Trigger.Event.Target
		}
		*calls = append(*calls, c)
		return nil
	})
}

const (
	thenToThisOracle = "test-prevention-then-to-this"
	thenToYouOracle  = "test-prevention-then-to-you"
)

// preventionStatic is "If damage would be dealt to <who>, prevent that
// damage. <then>" — or, with ThenPerSource, "If a source would deal
// damage to <who>, …". `prevent` is how much of each event it prevents
// (0 is all of it).
func preventionStatic(toThis bool, per PreventionUnit, then BodyRef, prevent int) ReplacementEffect {
	return ReplacementEffect{
		Watches:    []EventKind{EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
			if ev.Kind != RepEventDamage || src == nil {
				return false
			}
			if toThis {
				return ev.DamageTarget == src.InstanceID
			}
			return ev.DamageTarget == src.Controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			if prevent > 0 && prevent < ev.DamageAmount {
				ev.DamageAmount -= prevent
				return nil
			}
			ev.Cancel()
			return nil
		},
		Then:    then,
		ThenPer: per,
		Label:   "test prevention static",
	}
}

// withPreventionStatics stubs the catalog: a permanent with oracle `key`
// has the one replacement.
func withPreventionStatics(t *testing.T, byKey map[string]ReplacementEffect) {
	t.Helper()
	m := map[string][]ReplacementEffect{}
	for k, r := range byKey {
		m[k] = []ReplacementEffect{r}
	}
	stubCatalogReplacements(t, m)
}

// pushStaticCreature puts a 3/3 creature carrying `oracle`'s stubbed
// replacements onto the battlefield.
func pushStaticCreature(g *Game, owner *Player, oracle string) uuid.UUID {
	c := NewCard("Phantom", owner.ID)
	c.TypeLine = "Creature — Spirit"
	c.Power, c.Toughness = 3, 3
	c.OracleID = oracle
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// "If damage would be dealt to this creature, prevent that damage. Remove
// a +1/+1 counter": blocked by three creatures, all of it is prevented and
// the additional effect happens once (the Phantom Centaur ruling) — with
// the instance's total, from the static's object.
func TestPreventionStaticRunsOncePerRecipientPerInstance(t *testing.T) {
	var calls []thenCall
	withPreventionStatics(t, map[string]ReplacementEffect{
		thenToThisOracle: preventionStatic(true, ThenPerRecipient, thenRecorder(&calls), 0),
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phantom := pushStaticCreature(g, me, thenToThisOracle)
	var attackers []uuid.UUID
	for i := range 3 {
		attackers = append(attackers, pushColouredCreature(g, opp, fmt.Sprintf("Attacker %d", i), []string{"R"}))
	}
	g.WithWriteLock(func() {
		for i, a := range attackers {
			g.markCombatDamageOnCardLocked(phantom, i+1, a, CombatStepRegular)
		}
		g.runStateChecksLocked()
	})
	if got := damageOn(g, phantom); got != 0 {
		t.Errorf("phantom has %d damage, want 0: all of it is prevented", got)
	}
	if len(calls) != 1 {
		t.Fatalf("additional effect ran %d times, want once for one recipient in one instance: %+v", len(calls), calls)
	}
	c := calls[0]
	if c.prevented != 6 || c.damage != 6 || c.source != phantom || c.controller != me.ID || c.target != phantom {
		t.Errorf("follow-up got %+v, want 6 of 6 prevented, from the phantom, for its controller", c)
	}
	var epoch int
	g.WithWriteLock(func() { epoch = findBattlefieldCard(g, phantom).ObjectEpoch })
	if c.object != (ObjectRef{ID: phantom, Epoch: epoch}) {
		t.Errorf("follow-up's source object %+v, want the phantom as it is (epoch %d)", c.object, epoch)
	}
	// A second combat damage step is a second instance (CR 510.4).
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		g.markCombatDamageOnCardLocked(phantom, 2, attackers[0], CombatStepRegular)
		g.runStateChecksLocked()
	})
	if len(calls) != 2 || calls[1].prevented != 2 {
		t.Errorf("a later instance: calls %+v, want a second run with 2", calls)
	}
}

// "If a source would deal damage to you, prevent that damage and put an
// incarnation counter": two sources at once are two applications (the
// Nine Lives ruling), and one source dealing damage twice in one
// instruction is one.
func TestPreventionStaticRunsOncePerSourcePerInstance(t *testing.T) {
	var calls []thenCall
	withPreventionStatics(t, map[string]ReplacementEffect{
		thenToYouOracle: preventionStatic(false, ThenPerSource, thenRecorder(&calls), 0),
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushReplacementSource(g, thenToYouOracle, me.ID)
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	b := pushColouredCreature(g, opp, "B", []string{"G"})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		g.markCombatDamageToPlayerLocked(me.ID, a, 2, CombatStepRegular)
		g.markCombatDamageToPlayerLocked(me.ID, b, 3, CombatStepRegular)
		g.runStateChecksLocked()
	})
	if got := lifeOf(g, me.ID); got != start {
		t.Fatalf("life %d, want %d: all of it prevented", got, start)
	}
	if len(calls) != 2 {
		t.Fatalf("additional effect ran %d times, want once per source: %+v", len(calls), calls)
	}
	if calls[0].damageFrom != a || calls[0].prevented != 2 || calls[1].damageFrom != b || calls[1].prevented != 3 {
		t.Errorf("follow-ups %+v, want 2 from A then 3 from B", calls)
	}
	calls = nil
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		if err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToPlayerForEffect(a, me.ID, 1); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(a, me.ID, 2)
		}); err != nil {
			t.Fatal(err)
		}
	})
	if len(calls) != 1 || calls[0].prevented != 3 {
		t.Errorf("one source, one instance: calls %+v, want one run with 3", calls)
	}
}

// The apply loop owes what Replace PREVENTED, not the event's amount: "If
// a source would deal damage to you, prevent 1 of that damage" prevents 1
// of 4, and the follow-up sees 1 of 4.
func TestPreventionStaticOwesWhatItsReplacePrevented(t *testing.T) {
	var calls []thenCall
	withPreventionStatics(t, map[string]ReplacementEffect{
		thenToYouOracle: preventionStatic(false, ThenPerSource, thenRecorder(&calls), 1),
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushReplacementSource(g, thenToYouOracle, me.ID)
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 4) })
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Fatalf("life %d, want %d", got, start-3)
	}
	if len(calls) != 1 || calls[0].prevented != 1 || calls[0].damage != 4 {
		t.Errorf("follow-up %+v, want 1 prevented of 4", calls)
	}
}

// CR 615.12: under damage that can't be prevented the static prevents
// nothing, and its additional effect still happens once per unit — with
// nothing prevented, and the whole event as "that damage" (the Polukranos
// and Phyrexian Hydra rulings read one or the other).
func TestPreventionStaticUnderUnpreventableDamage(t *testing.T) {
	var calls []thenCall
	withPreventionStatics(t, map[string]ReplacementEffect{
		thenToThisOracle: preventionStatic(true, ThenPerRecipient, thenRecorder(&calls), 0),
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phantom := pushStaticCreature(g, me, thenToThisOracle)
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			for _, n := range []int{1, 1} {
				if err := g.DealMarkedDamageForEffect(src, nil, phantom, n, DamageMarks{CantBePrevented: true}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	if got := damageOn(g, phantom); got != 2 {
		t.Errorf("phantom has %d damage, want 2: unpreventable damage is dealt", got)
	}
	if len(calls) != 1 || calls[0].prevented != 0 || calls[0].damage != 2 {
		t.Errorf("follow-up %+v, want one run: 0 prevented of 2", calls)
	}
}

// A static whose object has left and come back is a new object (CR
// 400.7): an owed follow-up names the object it was, so a new object's
// application is a separate entry.
func TestPreventionStaticFollowUpIsKeyedOnItsObject(t *testing.T) {
	f := PreventionFollowUp{Static: true, Source: uuid.New(), Epoch: 1, Slot: 0, Unit: uuid.New(), Instance: 7}
	same := f
	same.Prevented = 3
	if !f.owes(same) {
		t.Error("same object, slot, unit and instance should be one owed follow-up")
	}
	for name, o := range map[string]PreventionFollowUp{
		"a new object": func() PreventionFollowUp { o := f; o.Epoch = 2; return o }(),
		"another slot": func() PreventionFollowUp { o := f; o.Slot = 1; return o }(),
		"another unit": func() PreventionFollowUp { o := f; o.Unit = uuid.New(); return o }(),
		"an instance":  func() PreventionFollowUp { o := f; o.Instance = 8; return o }(),
		"a shield":     {Seq: 4, Instance: 7},
	} {
		if f.owes(o) {
			t.Errorf("%s should be a separate follow-up", name)
		}
	}
}

// An owed static follow-up rides a restore point (a combat damage step's
// instance is owed until a player would next receive priority) and runs
// once in the restored game, its key intact.
func TestStaticFollowUpSurvivesARestorePoint(t *testing.T) {
	var calls []thenCall
	withPreventionStatics(t, map[string]ReplacementEffect{
		thenToThisOracle: preventionStatic(true, ThenPerRecipient, thenRecorder(&calls), 0),
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phantom := pushStaticCreature(g, me, thenToThisOracle)
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	b := pushColouredCreature(g, opp, "B", []string{"R"})
	g.WithWriteLock(func() { g.markCombatDamageOnCardLocked(phantom, 2, a, CombatStepRegular) })
	snap := g.CaptureSnapshot()
	if snap == nil || len(snap.PreventionFollowUps) != 1 || !snap.PreventionFollowUps[0].Static {
		t.Fatalf("snapshot owes %+v, want the static's follow-up", snap)
	}
	r, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	r.WithWriteLock(func() { r.runStateChecksLocked() })
	if len(calls) != 1 || calls[0].prevented != 2 || calls[0].source != phantom {
		t.Fatalf("restored game ran %+v, want one follow-up of 2 from the phantom", calls)
	}
	// The restored entry kept its key: more damage to the same object in
	// the same restored instance would have joined it. (A restored
	// combat damage step takes a new instance, damage_instance.go, so the
	// check is on the key itself.)
	f := snap.PreventionFollowUps[0]
	if f.Source != phantom || f.Unit != phantom || f.Slot != 0 {
		t.Errorf("captured entry %+v, want keyed on the phantom, slot 0, unit the phantom", f)
	}
	_ = b
}

// Owner decision 2: a charged shield (Test of Faith) owes its follow-up
// with what the charge prevented, out of the damage the event carried, and
// under CR 615.12 runs it with nothing prevented and keeps its charge.
func TestPreventDamageShieldRunsItsFollowUp(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushColouredCreature(g, me, "Bear", []string{"G"})
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	var calls []thenCall
	then := thenRecorder(&calls)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThenThisTurnForEffect(uuid.Nil, bear, 3, false,
			ShieldFollowUp{Controller: me.ID, Body: then}, "Test of Faith") {
			t.Fatal("no shield")
		}
		if err := g.DealMarkedDamageForEffect(src, nil, bear, 2, DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatal(err)
		}
	})
	if len(calls) != 1 || calls[0].prevented != 0 || calls[0].damage != 2 || calls[0].controller != me.ID {
		t.Fatalf("unpreventable: follow-up %+v, want 0 prevented of 2, for me", calls)
	}
	if g.ScopedEffects[0].Mods[0].Amount != 3 {
		t.Fatalf("charge %d, want 3: unpreventable damage doesn't reduce a shield", g.ScopedEffects[0].Mods[0].Amount)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 5) })
	if len(calls) != 2 || calls[1].prevented != 3 || calls[1].damage != 5 {
		t.Errorf("follow-up %+v, want 3 prevented of 5", calls)
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("spent shield kept: %+v", g.ScopedEffects)
	}
}

// Owner decision 2: Inkshield's "prevent all combat damage that would be
// dealt to you this turn. For each 1 damage prevented this way …" owes
// once per instance with the instance's total.
func TestPreventCombatDamageShieldRunsItsFollowUp(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	b := pushColouredCreature(g, opp, "B", []string{"R"})
	var calls []thenCall
	then := thenRecorder(&calls)
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		g.PreventCombatDamageThenThisTurnForEffect(uuid.Nil, me.ID, ShieldFollowUp{Controller: me.ID, Body: then}, "Inkshield")
		g.markCombatDamageToPlayerLocked(me.ID, a, 2, CombatStepRegular)
		g.markCombatDamageToPlayerLocked(me.ID, b, 4, CombatStepRegular)
		g.runStateChecksLocked()
		_ = g.DealDamageToPlayerForEffect(a, me.ID, 1)
	})
	if got := lifeOf(g, me.ID); got != start-1 {
		t.Errorf("life %d, want %d: combat damage prevented, the noncombat 1 dealt", got, start-1)
	}
	if len(calls) != 1 || calls[0].prevented != 6 {
		t.Errorf("follow-up %+v, want one run with 6", calls)
	}
}

// Mod.To: the follow-up is handed the object it deals damage to, while it
// is still that object (CR 400.7); once it has left, nothing (the
// Acolyte's Reward and Vengeful Archon rulings: the damage is still
// prevented, and the follow-up deals none).
func TestShieldFollowUpCarriesItsRecipient(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushColouredCreature(g, me, "Bear", []string{"G"})
	victim := pushColouredCreature(g, opp, "Victim", []string{"B"})
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	var calls []thenCall
	then := thenRecorder(&calls)
	g.WithWriteLock(func() {
		g.PreventNextDamageThenThisTurnForEffect(uuid.Nil, bear, 10, false,
			ShieldFollowUp{Controller: me.ID, Body: then, To: victim}, "Acolyte's Reward")
		g.PreventNextDamageThenThisTurnForEffect(uuid.Nil, me.ID, 10, false,
			ShieldFollowUp{Controller: me.ID, Body: then, To: opp.ID}, "Vengeful Archon")
		_ = g.DealDamageToCreatureForEffect(src, bear, 2)
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	if len(calls) != 2 {
		t.Fatalf("calls %+v, want two", calls)
	}
	if len(calls[0].to) != 1 || calls[0].to[0] != (TargetRef{Kind: TargetCard, ID: victim}) {
		t.Errorf("first follow-up's target %+v, want the victim", calls[0].to)
	}
	if len(calls[1].to) != 1 || calls[1].to[0] != (TargetRef{Kind: TargetPlayer, ID: opp.ID}) {
		t.Errorf("second follow-up's target %+v, want the opponent", calls[1].to)
	}
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(victim)
		_ = g.DealDamageToCreatureForEffect(src, bear, 1)
	})
	if len(calls) != 3 || len(calls[2].to) != 0 {
		t.Errorf("victim gone: follow-up %+v, want no target", calls[len(calls)-1])
	}
}

// A follow-up's recipient on a mod with no prevention follow-up is a newer
// binary's shape: registration panics, and restore refuses the file.
func TestFollowUpRecipientIsRefusedWithoutAFollowUp(t *testing.T) {
	if p := followUpModProblem(Mod{Kind: ModPreventDamage, Amount: 1, To: []ObjectRef{{ID: uuid.New()}}}); p == "" {
		t.Error("a recipient with no Then should be a problem")
	}
	if p := followUpModProblem(Mod{Kind: ModAddTypes, To: []ObjectRef{{ID: uuid.New()}}, Then: "x"}); p == "" {
		t.Error("a recipient on a non-prevention kind should be a problem")
	}
	if p := followUpModProblem(Mod{Kind: ModPreventDamage, Then: "x", To: []ObjectRef{{ID: uuid.New()}, {ID: uuid.New()}}}); p == "" {
		t.Error("two recipients should be a problem")
	}

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushColouredCreature(g, me, "Bear", []string{"G"})
	var calls []thenCall
	then := thenRecorder(&calls)
	g.WithWriteLock(func() {
		g.PreventNextDamageThenThisTurnForEffect(uuid.Nil, bear, 3, false,
			ShieldFollowUp{Controller: me.ID, Body: then, To: opp.ID}, "Acolyte's Reward")
	})
	data, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var plain GameSnapshot
	if err := json.Unmarshal(data, &plain); err != nil {
		t.Fatal(err)
	}
	if _, err := plain.RestoreStrict(); err != nil {
		t.Fatalf("the file as written does not restore: %v", err)
	}
	stripped := bytes.Replace(data, []byte(`"then":"`+then.Key()+`",`), nil, 1)
	if bytes.Equal(stripped, data) {
		t.Fatal("no then in the file to strip")
	}
	var s GameSnapshot
	if err := json.Unmarshal(stripped, &s); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore = %v, want ErrUnknownEffectKey", err)
	}
}
