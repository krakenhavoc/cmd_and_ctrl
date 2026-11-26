package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b381_cards_test.go — OLD batch 381 (Commander 2013, #5560): seven
// cards. Each test plays the real catalog entry through the engine.

const (
	oldB381Jar       = "3075dadd-240f-4455-9286-9f1d48f53a3f"
	oldB381Glory     = "5a8dd1b7-b63e-4997-9fe8-5e8816bc051b"
	oldB381Soulbeast = "5ea0c608-2c56-4889-a5d3-d435df515950"
	oldB381Hada      = "8db3152a-11b1-4c59-a745-1c40a9b235b7"
	oldB381Requiem   = "a72213e7-16f0-4fbe-91d3-750ca51a4336"
	oldB381Eye       = "e808a11e-29bd-4e99-a24e-67fa8f6fe502"
	oldB381Echo      = "fb98f6b7-5986-4c5d-98fc-e5c4106f48bf"
)

func TestOldB381EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB381Jar: "Jar of Eyeballs", oldB381Glory: "Tempt with Glory", oldB381Soulbeast: "Naya Soulbeast",
		oldB381Hada: "Hada Spy Patrol", oldB381Requiem: "Vile Requiem", oldB381Eye: "Eye of Doom", oldB381Echo: "Echo Mage",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB381Level(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterLevel, n); err != nil {
			t.Fatalf("level counters: %v", err)
		}
	})
}

func oldB381Size(t *testing.T, g *game.Game, id uuid.UUID, power, toughness int, what string) {
	t.Helper()
	if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != power || tg != toughness {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, tg, power, toughness)
	}
}

// --- Hada Spy Patrol --------------------------------------------------------

func TestOldB381HadaSpyPatrolLevels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Hada Spy Patrol", "Creature — Human Rogue", oldB381Hada, 1, 1)
	oldB381Size(t, g, id, 1, 1, "a level 0 Patrol")
	if effectstest.AvB13Restrictions(g, id)&game.CantBeBlocked != 0 {
		t.Error("a level 0 Patrol can't be blocked")
	}
	oldB381Level(t, g, id, 1)
	oldB381Size(t, g, id, 2, 2, "a level 1 Patrol")
	if effectstest.AvB13Restrictions(g, id)&game.CantBeBlocked == 0 {
		t.Error("a level 1 Patrol can be blocked")
	}
	if effectstest.P1g2HasKeyword(g, id, "shroud") {
		t.Error("a level 1 Patrol has shroud")
	}
	oldB381Level(t, g, id, 2)
	oldB381Size(t, g, id, 3, 3, "a level 3 Patrol")
	if !effectstest.P1g2HasKeyword(g, id, "shroud") || effectstest.AvB13Restrictions(g, id)&game.CantBeBlocked == 0 {
		t.Error("a level 3 Patrol lacks shroud or can be blocked")
	}
}

func TestOldB381HadaSpyPatrolLevelUpIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Hada Spy Patrol", "Creature — Human Rogue", oldB381Hada, 1, 1)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("level up: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, id, game.CounterLevel); n != 1 {
		t.Errorf("%d level counters, want 1", n)
	}
	oldB381Size(t, g, id, 2, 2, "the leveled Patrol")
}

// --- Echo Mage --------------------------------------------------------------

func oldB381EchoCopies(t *testing.T, level, wantLoss int) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Echo Mage", "Creature — Human Wizard", oldB381Echo, 2, 3)
	oldB381Level(t, g, mage, level)
	bolt := effectstest.CastCopyTarget(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	life := opp.Life
	effectstest.AddPool(t, g, me.ID, "{U}{U}")
	err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{
		Ref:    "grant:" + map[bool]string{false: "echo-mage/copy-once", true: "echo-mage/copy-twice"}[level >= 4] + ":0:0",
		Strict: true, Targets: effectstest.CardRefs(bolt)})
	if wantLoss == 0 {
		if err == nil {
			t.Fatal("a level 1 Echo Mage copied a spell")
		}
		return
	}
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if p := effectstest.CopyPromptFor(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != wantLoss {
		t.Errorf("the opponent lost %d, want %d", got, wantLoss)
	}
}

func TestOldB381EchoMageCopiesOnceAtLevelTwo(t *testing.T)   { oldB381EchoCopies(t, 2, 6) }
func TestOldB381EchoMageCopiesTwiceAtLevelFour(t *testing.T) { oldB381EchoCopies(t, 4, 9) }

// --- Tempt with Glory -------------------------------------------------------

func TestOldB381TemptWithGloryOpponentAccepts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 2, 2)
	_, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Tempt with Glory", TypeLine: "Sorcery", OracleID: oldB381Glory, ManaCost: "{5}{W}"}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if effectstest.LatestConfirmFor(g, p.ID) == nil {
			continue
		}
		effectstest.AnswerMayChoice(t, g, p.ID, p.ID == opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.B12Counter(t, g, mine, game.CounterPlusOne); n != 2 {
		t.Errorf("my Bear has %d counters, want 2 (one, plus one for the accepting opponent)", n)
	}
	if n := effectstest.B12Counter(t, g, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("their Elk has %d counters, want 1", n)
	}
}

func TestOldB381TemptWithGloryAllDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	_, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Tempt with Glory", TypeLine: "Sorcery", OracleID: oldB381Glory, ManaCost: "{5}{W}"}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if effectstest.LatestConfirmFor(g, p.ID) == nil {
			continue
		}
		effectstest.AnswerMayChoice(t, g, p.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.B12Counter(t, g, mine, game.CounterPlusOne); n != 1 {
		t.Errorf("my Bear has %d counters, want 1", n)
	}
}

// --- Naya Soulbeast ---------------------------------------------------------

func TestOldB381NayaSoulbeastEntersWithTheRevealedManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for i, p := range g.Seats {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top Card", TypeLine: "Sorcery", ManaCost: "{" + string(rune('1'+i)) + "}", Owner: p.ID, Controller: p.ID})
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Naya Soulbeast", TypeLine: "Creature — Beast", OracleID: oldB381Soulbeast,
		ManaCost: "{6}{G}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, "{G}{G}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	want := 0
	for i := range g.Seats {
		want += i + 1
	}
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != want {
		t.Errorf("the Soulbeast has %d +1/+1 counters, want %d", n, want)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, id); p != want {
		t.Errorf("the Soulbeast has power %d, want %d", p, want)
	}
	if !effectstest.P1g2HasKeyword(g, id, "trample") {
		t.Error("the Soulbeast lacks trample")
	}
}

// --- Vile Requiem -----------------------------------------------------------

func TestOldB381VileRequiemUpkeepCounterIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Vile Requiem", "Enchantment", oldB381Requiem, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, id, "verse"); n != 1 {
		t.Errorf("%d verse counters, want 1", n)
	}
}

func TestOldB381VileRequiemDestroysUpToXNonblackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Vile Requiem", "Enchantment", oldB381Requiem, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, "verse", 2) })
	a := effectstest.B12Creature(g, opp.ID, "Ogre A", "Creature — Ogre", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Ogre B", "Creature — Ogre", 3, 3)
	c := effectstest.B12Creature(g, opp.ID, "Ogre C", "Creature — Ogre", 3, 3)
	effectstest.AddPool(t, g, me.ID, "{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a, b, c)}); err == nil {
		t.Fatal("three targets accepted with two verse counters")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a targeted creature survived")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("the untargeted creature died")
	}
	if g.Battlefield.Contains(id) {
		t.Error("the Requiem was not sacrificed")
	}
}

func TestOldB381VileRequiemCannotTargetABlackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Vile Requiem", "Enchantment", oldB381Requiem, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, "verse", 1) })
	black := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: black, Name: "Black Knight", TypeLine: "Creature — Knight", ManaCost: "{B}{B}", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AddPool(t, g, me.ID, "{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(black)}); err == nil {
		t.Error("a black creature was a legal target")
	}
}

// --- Eye of Doom ------------------------------------------------------------

func TestOldB381EyeOfDoomMarksOnePermanentPerPlayerThenDestroysThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	eye, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Eye of Doom", TypeLine: "Artifact", OracleID: oldB381Eye, ManaCost: "{4}"}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		pick := a
		if p.ID != me.ID {
			pick = b
		}
		effectstest.SlChoose(t, g, p.ID, pick)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.B12Counter(t, g, a, "doom"); n != 1 {
		t.Errorf("Bear A has %d doom counters, want 1", n)
	}
	if n := effectstest.B12Counter(t, g, b, "doom"); n != len(g.Seats)-1 {
		t.Errorf("Bear B has %d doom counters, want %d", n, len(g.Seats)-1)
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, eye, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a marked permanent survived")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an unmarked permanent was destroyed")
	}
}

// --- Jar of Eyeballs --------------------------------------------------------

func TestOldB381JarOfEyeballsCountsDeathsAndDigs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jar := effectstest.B12Push(g, me.ID, "Jar of Eyeballs", "Artifact", oldB381Jar, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, jar, "eyeball"); n != 2 {
		t.Fatalf("%d eyeball counters, want 2", n)
	}
	var top []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		top = append(top, id)
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	hand := effectstest.HandSize(me)
	if err := g.ActivateCatalogAbility(me.ID, jar, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, jar, "eyeball"); n != 0 {
		t.Errorf("%d eyeball counters left, want 0", n)
	}
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		effectstest.SlChoose(t, g, me.ID, top[2])
		effectstest.PassPriorityAroundTable(t, g)
	}
	if z := effectstest.B12ZoneOf(g, top[0]); z != game.ZoneLibrary {
		t.Errorf("the third card from the top is in %s: it was not looked at, so it stays in the library", z)
	}
	if z := effectstest.B12ZoneOf(g, top[2]); z != game.ZoneHand {
		t.Errorf("the chosen card is in %s, want the hand", z)
	}
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d → %d, want one card taken", hand, got)
	}
}
