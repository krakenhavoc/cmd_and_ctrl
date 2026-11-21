package testsru

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b93_test.go — set GAP, batch 93: Doctor Who enchantments (S82, #5948).

const (
	gapFourKnocks   = "64fe83b1-5aae-4433-82c3-a11c8c83137e"
	gapCrackInTime  = "a3099143-0146-435f-8d29-cd364aefbc2b"
	gapBanish       = "ae6ef097-bd52-4bf8-81a8-2e6c643fc805"
	gapRegenRestor  = "a1e691f3-a3e9-4cfa-a761-d7e020364dfa"
	gapPartingWays  = "58d5b285-eb0c-4478-b137-26a897075f70"
	gapSoundDrums   = "0afcc646-cb34-4c0e-b3f5-4115ad4faa8e"
	gapHeavenSent   = "1863ba60-f786-4e52-90fc-f3135ca0257f"
	gapUnearthly    = "9527d2dc-d2e2-4259-aefe-2342ea07bc63"
	gapBecomePilot  = "7e079dd3-3397-491f-a6c5-7be5ade687d3"
	gapEleventhHour = "6b3e1aaf-6909-400a-88f2-9e9fde9a0cf7"
	gapGallifrey    = "52680471-a407-418b-9e74-c32308322772"
)

func gapRegistered(t *testing.T, names map[string]string) {
	t.Helper()
	for oracle, name := range names {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// gapSettle passes priority until the stack is empty, keeping every card on
// top for any scry prompt and ordering simultaneous triggers as offered (or,
// with reverse, the other way round).
func gapSettle(t *testing.T, g *game.Game, reverse bool) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if order := effectstest.TriggerOrderPrompt(g); order != nil {
			ids := append([]uuid.UUID(nil), order.TriggerOrderIDs...)
			if reverse {
				slices.Reverse(ids)
			}
			if err := g.ResolveTriggerOrder(order.ID, order.Chooser, ids); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		answered := false
		for _, s := range g.Seats {
			if c := effectstest.ScryChoiceFor(g, s.ID); c != nil {
				if err := g.ResolveScry(c.ID, s.ID, nil, c.ScryCards); err != nil {
					t.Fatalf("ResolveScry: %v", err)
				}
				answered = true
			}
		}
		if answered {
			continue
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// gapToMyNextFirstMain walks to seat 0's next precombat main phase and
// settles what triggered.
func gapToMyNextFirstMain(t *testing.T, g *game.Game) { gapToMyNextFirstMainOrdered(t, g, false) }

func gapToMyNextFirstMainOrdered(t *testing.T, g *game.Game, reverse bool) {
	t.Helper()
	step := func() {
		gapSettle(t, g, reverse)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && g.Turn.ActiveSeat == 0; i++ {
		step()
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain); i++ {
		step()
	}
	gapSettle(t, g, reverse)
}

func gapTimeCounters(g *game.Game, id uuid.UUID) int {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterTime]
}

func TestS82GapB93EnchantmentsRegisteredFull(t *testing.T) {
	gapRegistered(t, map[string]string{
		gapFourKnocks: "Four Knocks", gapCrackInTime: "Crack in Time", gapBanish: "Banish to Another Universe",
		gapRegenRestor: "Regenerations Restored",
	})
}

func TestS82GapB93FourKnocksVanishesAndDrawsAtFirstMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	knocks := effectstest.CastCatalogSpell(t, g, "Four Knocks", "Enchantment", gapFourKnocks, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapTimeCounters(g, knocks); got != 4 {
		t.Fatalf("%d time counters on entry, want 4", got)
	}
	before := me.Hand.Size()
	gapToMyNextFirstMain(t, g)
	if got := gapTimeCounters(g, knocks); got != 3 {
		t.Errorf("after the upkeep: %d time counters, want 3", got)
	}
	if got := me.Hand.Size(); got != before+2 { // the draw step and the trigger
		t.Errorf("hand grew by %d, want 2 (draw step + Four Knocks)", got-before)
	}
	for want := 2; want >= 0; want-- {
		gapToMyNextFirstMain(t, g)
		if want > 0 && gapTimeCounters(g, knocks) != want {
			t.Fatalf("%d counters, want %d", gapTimeCounters(g, knocks), want)
		}
	}
	if _, here := effectstest.BattlefieldCard(g, knocks); here {
		t.Error("Four Knocks survived its last time counter")
	}
}

func TestS82GapB93CrackInTimeExilesOnEntryAndEachFirstMainThenReturnsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.B12Creature(g, opp.ID, "First Victim", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second Victim", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	crack := effectstest.CastCatalogSpell(t, g, "Crack in Time", "Enchantment", gapCrackInTime, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Fatal("the first creature was not exiled on entry")
	}
	_ = mine
	gapToMyNextFirstMain(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(second) {
		t.Fatal("the second creature was not exiled at the first main phase")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(crack) })
	g.RunStateChecksForTest()
	if effectstest.CountOnBattlefield(g, "First Victim", opp.ID) != 1 || effectstest.CountOnBattlefield(g, "Second Victim", opp.ID) != 1 {
		t.Error("both exiled creatures should return when Crack in Time leaves")
	}
}

func TestS82GapB93CrackInTimeCannotTargetYourOwnCreature(t *testing.T) {
	spec, _ := effects.Lookup(gapCrackInTime)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for _, tr := range spec.Triggered {
		if tr.Targets != nil {
			effectstest.RfPermAExpectLegal(t, g, me.ID, tr.Targets, map[uuid.UUID]bool{mine: false, theirs: true})
		}
	}
}

func TestS82GapB93BanishToAnotherUniverseCostsLessForHistoricAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := effectstest.SelfPricedMV(t, g, me, gapBanish, "Enchantment", "{4}{W}"); got != 5 {
		t.Errorf("no historic permanents: costs %d, want 5", got)
	}
	effectstest.BlbB1Push(g, me.ID, "Relic", "Artifact", "", 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Legend", "Legendary Creature — Human", "", 1, 1)
	effectstest.B12Permanent(g, me.ID, "A Saga", "Enchantment — Saga")
	effectstest.B12Permanent(g, me.ID, "Plain Enchantment", "Enchantment")
	if got := effectstest.SelfPricedMV(t, g, me, gapBanish, "Enchantment", "{4}{W}"); got != 2 {
		t.Errorf("an artifact, a legend and a Saga: costs %d, want 2", got)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	spec, _ := effects.Lookup(gapBanish)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{victim: true, land: false})
	banish := effectstest.CastCatalogSpell(t, g, "Banish to Another Universe", "Enchantment", gapBanish, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the permanent is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(banish) })
	g.RunStateChecksForTest()
	if effectstest.CountOnBattlefield(g, "Their Threat", opp.ID) != 1 {
		t.Error("the exiled permanent should return when the enchantment leaves")
	}
}

func TestS82GapB93RegenerationsRestoredGainsLifeAndTakesTheExtraTurnAtTheLastCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	regen := effectstest.CastCatalogSpell(t, g, "Regenerations Restored", "Enchantment", gapRegenRestor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapTimeCounters(g, regen); got != 12 {
		t.Fatalf("%d time counters on entry, want 12", got)
	}
	life := me.Life
	gapToMyNextFirstMain(t, g)
	if got := gapTimeCounters(g, regen); got != 11 {
		t.Fatalf("%d counters after one upkeep, want 11", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

// At the last counter two triggers happen together: Vanishing's sacrifice
// and Regenerations Restored's scry / life / exile. Its controller orders
// them, so the extra turn depends on the order: exiled first, there is one;
// sacrificed first, the enchantment is gone and cannot be exiled.
func TestS82GapB93RegenerationsRestoredExtraTurnDependsOnTheTriggerOrder(t *testing.T) {
	var extraTurns []int
	for _, reverse := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		regen := effectstest.CastCatalogSpell(t, g, "Regenerations Restored", "Enchantment", gapRegenRestor, nil)
		effectstest.PassPriorityAroundTable(t, g)
		g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, regen).Counters[game.CounterTime] = 1 })
		before := len(g.ExtraTurnsQueued())
		gapToMyNextFirstMainOrdered(t, g, reverse)
		if c, here := effectstest.BattlefieldCard(g, regen); here {
			t.Errorf("reverse=%v: the enchantment is still on the battlefield (%d counters)", reverse, c.Counters[game.CounterTime])
		}
		extraTurns = append(extraTurns, len(g.ExtraTurnsQueued())-before)
	}
	if extraTurns[0]+extraTurns[1] != 1 || (extraTurns[0] != 1 && extraTurns[1] != 1) {
		t.Errorf("extra turns by trigger order %v, want exactly one order to take the extra turn", extraTurns)
	}
}

func TestS82GapB93SagasAndAurasRegisteredFull(t *testing.T) {
	gapRegistered(t, map[string]string{
		gapPartingWays: "The Parting of the Ways", gapSoundDrums: "The Sound of Drums", gapHeavenSent: "Heaven Sent",
		gapUnearthly: "An Unearthly Child", gapBecomePilot: "Become the Pilot", gapEleventhHour: "The Eleventh Hour",
		gapGallifrey: "Gallifrey Stands",
	})
}

func TestS82GapB93SoundOfDrumsDoublesCombatDamageOfTheEnchantedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.AuraBear(g, opp.ID)
	other := effectstest.AuraBear(g, opp.ID)
	victim := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 20)
	effectstest.AuraCast(t, g, "The Sound of Drums", gapSoundDrums, host)
	if err := g.MarkCombatDamage(host, victim, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := bnsMarked(g, victim); got != 4 {
		t.Errorf("the enchanted creature dealt %d combat damage, want 4 (doubled)", got)
	}
	if err := g.MarkCombatDamage(other, victim, 3); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := bnsMarked(g, victim); got != 7 {
		t.Errorf("after an unenchanted creature's 3 the wall has %d damage, want 7", got)
	}
}

func TestS82GapB93SoundOfDrumsGoadsAndReturnsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	host := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "The Sound of Drums", gapSoundDrums, host)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(host, me.ID); err == nil {
		t.Error("the goaded creature attacked the goader while another player was open")
	}
	if err := g.DeclareAttacker(host, third.ID); err != nil {
		t.Errorf("attacking another player was refused: %v", err)
	}

	g2 := effectstest.NewCatalogGame(t)
	id, me2 := effectstest.ActivateFromGraveyard(t, g2, "The Sound of Drums", "Enchantment — Aura", gapSoundDrums, 0, 0, "{C}{C}{R}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	if !me2.Hand.Contains(id) || me2.Graveyard.Contains(id) {
		t.Error("the Aura did not return from the graveyard to its owner's hand")
	}
}

func TestS82GapB93BecomeThePilotStealsPumpsAndRestrictsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	host := effectstest.AuraBear(g, opp.ID)
	commander := effectstest.AuraBear(g, opp.ID)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, commander).IsCommander = true })
	spec, _ := effects.Lookup(gapBecomePilot)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Targets, map[uuid.UUID]bool{host: true, commander: false})
	effectstest.AuraCast(t, g, "Become the Pilot", gapBecomePilot, host)
	if got := effectstest.ControllerOf(t, g, host); got != me.ID {
		t.Fatalf("controller %s, want the Aura's controller", got)
	}
	if p, tt := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
	blocker := effectstest.B12Creature(g, third.ID, "Blocker", "Creature — Bear", 2, 2)
	card := func(id uuid.UUID) *game.Card {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				return &g.Battlefield.Cards[i]
			}
		}
		t.Fatalf("card %v is not on the battlefield", id)
		return nil
	}
	check := func(target uuid.UUID, wantLegal bool, why string) {
		t.Helper()
		g.WithWriteLock(func() { card(host).AttackingTarget = target })
		g.ReadSnapshot(func() {
			if r := g.BlockPairRefusalLocked(card(host), card(blocker)); r.Legal() != wantLegal {
				t.Errorf("%s: block legal=%v, want %v (%q)", why, r.Legal(), wantLegal, r.Reason)
			}
		})
	}
	check(third.ID, false, "attacking a player who is not its owner")
	check(opp.ID, true, "attacking its owner")
	pw := effectstest.B12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test")
	check(pw, true, "attacking a permanent its owner controls")
	pw2 := effectstest.B12Permanent(g, third.ID, "Other Walker", "Legendary Planeswalker — Test")
	check(pw2, false, "attacking another player's planeswalker")
}

func TestS82GapB93PartingOfTheWaysExilesFiveWithTimeCountersAndSuspend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Library, top first: a land, a {2} sorcery, a {4} creature, a card that
	// already has suspend is not available as a fixture, so a {1} instant, a land.
	land2 := effectstest.PushLibraryTyped(me, "Low Land", "Basic Land — Forest")
	inst := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: inst, Name: "Cheap Instant", TypeLine: "Instant", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	big := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: big, Name: "Big Creature", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Mid Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	land1 := effectstest.PushLibraryTyped(me, "Top Land", "Basic Land — Mountain")
	rest := effectstest.PushLibraryTyped(me, "Sixth Card", "Basic Land — Island")
	_ = rest
	// pushed last is on top, so reorder: the five exiled are land1, spell, big, inst, land2.
	effectstest.CastCatalogSpell(t, g, "The Parting of the Ways", "Enchantment — Saga", gapPartingWays, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, want := range map[string]int{"Mid Sorcery": 2, "Big Creature": 4, "Cheap Instant": 1} {
		found := false
		for _, c := range g.Exile.Cards {
			if c.Name == name {
				found = true
				if c.Counters[game.CounterTime] != want || !c.SuspendGranted {
					t.Errorf("%s: %d time counters (want %d), suspend granted=%v", name, c.Counters[game.CounterTime], want, c.SuspendGranted)
				}
			}
		}
		if !found {
			t.Errorf("%s was not exiled", name)
		}
	}
	for _, id := range []uuid.UUID{land1, land2} {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.Counters[game.CounterTime] != 0 || c.SuspendGranted {
			t.Errorf("a land was given time counters or suspend: %+v", c)
		}
	}
}

func TestS82GapB93PartingOfTheWaysTimeTravelsTwiceThenDestroysOneArtifactPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	clock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 1}})
	saga := effectstest.CastCatalogSpell(t, g, "The Parting of the Ways", "Enchantment — Saga", gapPartingWays, nil)
	effectstest.PassPriorityAroundTable(t, g)
	art1 := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	art2 := effectstest.B12Permanent(g, opp.ID, "Their Other Rock", "Artifact")
	art3 := effectstest.B12Permanent(g, third.ID, "Third Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TmtB1Card(t, g, clock).Counters[game.CounterTime]; n != 3 {
		t.Fatalf("the Clock has %d time counters after two time travels, want 3", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	// Chapter III: a target per opponent is chosen; pick one artifact each.
	effectstest.HobAnswerTargets(t, g, me.ID, art1, art3)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art1) || g.Battlefield.Contains(art3) {
		t.Error("a chosen artifact survived")
	}
	if !g.Battlefield.Contains(art2) || !g.Battlefield.Contains(mine) {
		t.Error("an artifact that was not chosen was destroyed (one per opponent, never your own)")
	}
	if _, here := effectstest.BattlefieldCard(g, saga); here {
		t.Error("the Saga stayed after its last chapter")
	}
}

func TestS82GapB93HeavenSentInvestigatesThenExilesAndMayBeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saga := effectstest.CastCatalogSpell(t, g, "Heaven Sent", "Enchantment — Saga", gapHeavenSent, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldByName(g, "Clue"); got != 1 {
		t.Fatalf("%d Clues after chapter I, want 1", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldByName(g, "Clue"); got != 2 {
		t.Fatalf("%d Clues after chapter II, want 2", got)
	}
	life := opp.Life
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	inExile := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == saga {
			inExile = true
		}
	}
	if !inExile {
		t.Fatal("the Saga was not exiled (no opponent was at 0 or less)")
	}
	granted := 0
	for _, perm := range me.CastPermissions {
		if perm.Zone == game.ZoneExile && perm.Label == "Heaven Sent — you may cast this Saga this turn" {
			granted++
		}
	}
	if granted != 1 {
		t.Errorf("%d cast permissions over the exiled Saga, want 1", granted)
	}
}

func TestS82GapB93HeavenSentDrawsSevenWhenAnOpponentIsAtZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 10; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Basic Land — Forest")
	}
	g.WithWriteLock(func() { opp.Life = 1 })
	effectstest.CastCatalogSpell(t, g, "Heaven Sent", "Enchantment — Saga", gapHeavenSent, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got < 7 {
		t.Errorf("drew %d cards, want at least 7 (plus the turn's draw)", got)
	}
}

func TestS82GapB93UnearthlyChildFindsADoctorACompanionOrAVehicleThenBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Top first: Filler, Filler, Vehicle, Companion-less, Doctor, ...
	doctor := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: doctor, Name: "The Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryTyped(me, "Bottom Filler", "Basic Land — Forest")
	companion := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: companion, Name: "Rose Tyler", OracleID: "2f56de72-80d7-48c6-9b57-84fbd252c699", TypeLine: "Legendary Creature — Human", Owner: me.ID, Controller: me.ID})
	vehicle := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: vehicle, Name: "Ship", TypeLine: "Artifact — Vehicle", Owner: me.ID, Controller: me.ID})
	skip := effectstest.PushLibraryTyped(me, "Skipped", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "An Unearthly Child", "Enchantment — Saga", gapUnearthly, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(vehicle) || me.Hand.Contains(skip) {
		t.Fatalf("chapter I did not take the Vehicle and only it (hand has vehicle=%v)", me.Hand.Contains(vehicle))
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(companion) {
		t.Error("chapter II did not take the card with doctor's companion")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(doctor) {
		t.Error("chapter III did not take the Doctor card")
	}
}

func TestS82GapB93EleventhHourSearchesCreatesFoodAndHumanThenCopiesAsPrisonerZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doctor := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: doctor, Name: "A Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", ManaCost: "{2}{U}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryTyped(me, "Not A Doctor", "Basic Land — Island")
	victim := effectstest.B12Creature(g, g.Seats[1].ID, "Their Elf", "Creature — Elf Warrior", 2, 3)
	effectstest.CastCatalogSpell(t, g, "The Eleventh Hour", "Enchantment — Saga", gapEleventhHour, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(doctor) {
		t.Fatal("chapter I did not fetch the Doctor card")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "Food") != 1 || effectstest.CountOnBattlefield(g, "Human", me.ID) != 1 {
		t.Fatalf("chapter II: Food=%d Human=%d", effectstest.CountBattlefieldByName(g, "Food"), effectstest.CountOnBattlefield(g, "Human", me.ID))
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	var pz *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Prisoner Zero" {
			pz = &g.Battlefield.Cards[i]
		}
	}
	if pz == nil {
		t.Fatal("no Prisoner Zero token")
	}
	if !pz.IsToken() || pz.Controller != me.ID || pz.Power != 2 || pz.Toughness != 3 {
		t.Errorf("Prisoner Zero: token=%v controller ok=%v %d/%d, want a 2/3 token under my control", pz.IsToken(), pz.Controller == me.ID, pz.Power, pz.Toughness)
	}
	if !pz.IsLegendary() || !pz.HasSubtype("Alien") || pz.HasSubtype("Elf") || pz.HasSubtype("Warrior") {
		t.Errorf("Prisoner Zero type line %q, want a legendary Alien with no other creature types", pz.TypeLine)
	}
}

func TestS82GapB93GallifreyStandsReturnsDoctorsAndWinsAtThirteen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	d1, d2 := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: d1, Name: "Doc One", TypeLine: "Legendary Creature — Time Lord Doctor", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: d2, Name: "Not A Doc", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Gallifrey Stands", "Legendary Enchantment", gapGallifrey, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(d1) || me.Hand.Contains(d2) {
		t.Errorf("the Doctor returned=%v, the Bear returned=%v; want true and false", me.Hand.Contains(d1), me.Hand.Contains(d2))
	}
	// Twelve Doctors in play and one in hand: the upkeep put makes thirteen.
	for i := 0; i < 12; i++ {
		effectstest.B12Creature(g, me.ID, fmt.Sprintf("Doctor Body %d", i), "Legendary Creature — Time Lord Doctor", 1, 1)
	}
	if g.State != game.StateActive {
		t.Fatalf("the game ended early: %v", g.State)
	}
	for i := 0; i < 400 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, d1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.State == game.StateActive {
		t.Error("controlling thirteen Doctors at the upkeep did not win the game")
	}
}
