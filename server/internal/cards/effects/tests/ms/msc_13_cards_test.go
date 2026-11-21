package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_13_cards_test.go — Marvel Super Heroes Commander, thirteenth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc13HerculesOracle   = "ee3dfdb2-07ef-4ce2-9806-0a5b7f19aa67"
	msc13GoliathOracle    = "d06d7065-27e5-4119-a226-6f6d990d72c7"
	msc13JackOracle       = "25b30117-03ef-49a5-8743-2d1c167d4bc3"
	msc13IcemanOracle     = "f01c779a-97be-4727-ab62-33579f52447a"
	msc13HappyOracle      = "daaf33d8-2ee2-41f8-9872-66b6493c6e13"
	msc13HyperionOracle   = "4f4a5966-3713-45c7-acc2-9ddb4ff73eab"
	msc13ScepterOracle    = "5f8941d8-f8f7-4ec3-86c3-88af8829080b"
	msc13KillmongerOracle = "aff21bd1-87b0-409c-bfad-08f3690ef220"
	msc13LockjawOracle    = "5efce07e-24c4-4520-a46f-9954b67c4e6e"
	msc13InvisibleOracle  = "085016ee-4429-453f-b7f3-d61065f1214e"
	msc13JocastaOracle    = "787e6a7b-64f9-48ba-bccd-68dda3489b0f"
)

func TestMsc13EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc13HerculesOracle:   "Hercules, Olympian Hero",
		msc13GoliathOracle:    "Goliath, Mass Manipulator",
		msc13JackOracle:       "Jack of Hearts, Volatile Hero",
		msc13IcemanOracle:     "Iceman and Firestar",
		msc13HappyOracle:      "Happy Hogan, Bodyguard",
		msc13HyperionOracle:   "Hyperion, Supreme Hero",
		msc13ScepterOracle:    "Loki's Scepter",
		msc13KillmongerOracle: "Killmonger, Ruthless Usurper",
		msc13LockjawOracle:    "Lockjaw, Slobbering Teleporter",
		msc13InvisibleOracle:  "Invisible Woman",
		msc13JocastaOracle:    "Jocasta, Automaton Avenger",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// msc13Hurt deals noncombat damage to a creature and settles the table.
func msc13Hurt(t *testing.T, g *game.Game, source, target uuid.UUID, n int) {
	t.Helper()
	effectstest.B27Damage(g, source, target, n)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
}

// msc13Cast puts a spell of the given colour into `p`'s hand and casts it.
func msc13Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost string, colors ...string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors,
			Owner: p.ID, Controller: p.ID,
		})
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- Hercules, Olympian Hero ---------------------------------------------

const msc13HerculesTypes = "Legendary Creature — Demigod Warrior Hero"

func TestMsc13HerculesGrowsAndIsIndestructibleWhenHeAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herc := effectstest.B12Push(g, me.ID, "Hercules, Olympian Hero", msc13HerculesTypes, msc13HerculesOracle, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, herc)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, herc); got != 1 {
		t.Fatalf("+1/+1 counters after attacking: %d, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, herc, "indestructible") {
		t.Error("Hercules is not indestructible after attacking")
	}
}

func TestMsc13HerculesIsNotIndestructibleWhenHeDoesNotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	herc := effectstest.B12Push(g, g.Seats[0].ID, "Hercules, Olympian Hero", msc13HerculesTypes, msc13HerculesOracle, 3, 3)
	if effectstest.HasEffectiveKeyword(t, g, herc, "indestructible") {
		t.Error("Hercules is indestructible without attacking")
	}
	if got := effectstest.PlusOnes(g, herc); got != 0 {
		t.Errorf("+1/+1 counters without attacking: %d, want 0", got)
	}
}

func TestMsc13HerculesGetsCountersOnlyForTheFirstDamageEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herc := effectstest.B12Push(g, me.ID, "Hercules, Olympian Hero", msc13HerculesTypes, msc13HerculesOracle, 3, 3)
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)

	msc13Hurt(t, g, pinger, herc, 2)
	if got := effectstest.PlusOnes(g, herc); got != 2 {
		t.Fatalf("first damage of 2: %d counters, want 2", got)
	}
	// Hercules is now 5/5 with 2 damage: a second hit this turn does nothing.
	msc13Hurt(t, g, pinger, herc, 1)
	if got := effectstest.PlusOnes(g, herc); got != 2 {
		t.Errorf("second damage the same turn: %d counters, want 2 (first time each turn only)", got)
	}
}

func TestMsc13HerculesGetsCountersAgainOnTheNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herc := effectstest.B12Push(g, me.ID, "Hercules, Olympian Hero", msc13HerculesTypes, msc13HerculesOracle, 3, 3)
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	msc13Hurt(t, g, pinger, herc, 1)
	if got := effectstest.PlusOnes(g, herc); got != 1 {
		t.Fatalf("first turn: %d counters, want 1", got)
	}
	effectstest.PowerUpPassTo(t, g, 1)
	msc13Hurt(t, g, pinger, herc, 1)
	if got := effectstest.PlusOnes(g, herc); got != 2 {
		t.Errorf("next turn's first damage: %d counters, want 2", got)
	}
}

func TestMsc13HerculesGetsNoCountersFromDamageThatKillsHim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herc := effectstest.B12Push(g, me.ID, "Hercules, Olympian Hero", msc13HerculesTypes, msc13HerculesOracle, 3, 3)
	smasher := effectstest.B12Creature(g, opp.ID, "Smasher", "Creature — Test", 5, 5)
	msc13Hurt(t, g, smasher, herc, 5)
	if g.Battlefield.Contains(herc) {
		t.Fatal("Hercules survived lethal damage")
	}
	if got := effectstest.PlusOnes(g, herc); got != 0 {
		t.Errorf("a dead Hercules has %d counters, want 0", got)
	}
}

func TestMsc13HerculesCountsSimultaneousDamageAsOneTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herc := effectstest.B12Push(g, me.ID, "Hercules, Olympian Hero", msc13HerculesTypes, msc13HerculesOracle, 3, 3)
	a := effectstest.B12Creature(g, opp.ID, "Pinger A", "Creature — Test", 1, 9)
	b := effectstest.B12Creature(g, opp.ID, "Pinger B", "Creature — Test", 1, 9)
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(a, herc, 1)
			return g.DealDamageToCreatureForEffect(b, herc, 1)
		})
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, herc); got != 2 {
		t.Errorf("two simultaneous 1-damage hits: %d counters, want 2 (one trigger, that many)", got)
	}
}

// --- Goliath, Mass Manipulator -------------------------------------------

const msc13GoliathTypes = "Legendary Creature — Human Scientist Hero"

func TestMsc13GoliathHasReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Goliath, Mass Manipulator", msc13GoliathTypes, msc13GoliathOracle, 2, 2)
	effectstest.AssertKeywords(t, g, id, "reach")
}

func TestMsc13GoliathPowerUpPutsTwoCountersAndDrawsForEachBigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	gol := effectstest.PushPowerUpPermanent(g, me.ID, "Goliath, Mass Manipulator", msc13GoliathTypes, "{1}{G}", msc13GoliathOracle)
	effectstest.B12Creature(g, me.ID, "Big Friend", "Creature — Test", 5, 5)
	effectstest.B12Creature(g, me.ID, "Small Friend", "Creature — Test", 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Test", 7, 7)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	before := me.Hand.Size()

	if got := effectstest.ChargedPowerUp(t, g, me.ID, gol); got != "{4}{G}" {
		t.Fatalf("power-up price: %q, want {4}{G}", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, gol, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, gol); got != 2 {
		t.Errorf("+1/+1 counters: %d, want 2", got)
	}
	// Goliath is 4/4 after the counters, so he and the 5/5 count; the 3/3
	// and the opponent's 7/7 do not.
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("cards drawn: %d, want 2", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, gol, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second power-up activation was accepted")
	}
}

func TestMsc13GoliathPowerUpCostsHisManaCostLessTheTurnHeEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	gol := effectstest.CastPowerUpCreature(t, g, me, "Goliath, Mass Manipulator", msc13GoliathTypes, "{1}{G}", "{C}{G}", msc13GoliathOracle, 2)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, gol); got != "{3}" {
		t.Errorf("the turn he entered: %q, want {3}", got)
	}
}

// --- Jack of Hearts, Volatile Hero ---------------------------------------

const msc13JackTypes = "Legendary Creature — Alien Human Hero"

func TestMsc13JackOfHeartsHasFlyingAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Jack of Hearts, Volatile Hero", msc13JackTypes, msc13JackOracle, 2, 2)
	effectstest.AssertKeywords(t, g, id, "flying", "haste")
}

func TestMsc13JackOfHeartsDiesForItsPowerToEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jack := effectstest.B12Push(g, me.ID, "Jack of Hearts, Volatile Hero", msc13JackTypes, msc13JackOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	imp := effectstest.B12Creature(g, opp.ID, "Their Imp", "Creature — Imp", 1, 1)
	effectstest.B27Kill(g, jack)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(imp) {
		t.Error("a creature with toughness 2 or less survived Jack's 2 damage")
	}
	if !g.Battlefield.Contains(ox) {
		t.Fatal("the 3/3 died to 2 damage")
	}
	if got := effectstest.DamageMarkedOn(g, ox); got != 2 {
		t.Errorf("the 3/3 has %d damage, want 2", got)
	}
}

func TestMsc13JackOfHeartsPowerUpDiscardsTheHandDrawsThreeAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	jack := effectstest.PushPowerUpPermanent(g, me.ID, "Jack of Hearts, Volatile Hero", msc13JackTypes, "{1}{R}{R}", msc13JackOracle)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	old := me.Hand.Size()
	if old == 0 {
		t.Fatal("setup: the hand is empty")
	}
	gy := me.Graveyard.Size()
	if got := effectstest.ChargedPowerUp(t, g, me.ID, jack); got != "{4}{R}{R}" {
		t.Fatalf("power-up price: %q, want {4}{R}{R}", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, jack, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand after discarding it and drawing three: %d, want 3", got)
	}
	if got := me.Graveyard.Size() - gy; got != old {
		t.Errorf("cards discarded: %d, want %d", got, old)
	}
	if got := effectstest.PlusOnes(g, jack); got != 2 {
		t.Errorf("+1/+1 counters: %d, want 2", got)
	}
}

func TestMsc13JackOfHeartsDeathDamageUsesHisLastKnownPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jack := effectstest.B12Push(g, me.ID, "Jack of Hearts, Volatile Hero", msc13JackTypes, msc13JackOracle, 2, 2)
	effectstest.AddCountersForTest(g, jack, game.CounterPlusOne, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	small := effectstest.B12Creature(g, opp.ID, "Their Wall Two", "Creature — Wall", 0, 5)
	effectstest.B27Kill(g, jack)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 0/4 survived Jack's four damage")
	}
	if got := effectstest.DamageMarkedOn(g, small); got != 4 {
		t.Errorf("the 0/5 has %d damage, want 4", got)
	}
}

// --- Iceman and Firestar -------------------------------------------------

const msc13IcemanTypes = "Legendary Creature — Mutant Hero"

func TestMsc13IcemanAndFirestarTapsATargetWhenYouCastABlueSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Iceman and Firestar", msc13IcemanTypes, msc13IcemanOracle, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	msc13Cast(t, g, me, "Blue Spell", "Sorcery", "{U}", "U")
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, victim); !c.Tapped {
		t.Error("the targeted creature was not tapped")
	}
}

func TestMsc13IcemanAndFirestarDoesNothingForANonBlueNonRedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Iceman and Firestar", msc13IcemanTypes, msc13IcemanOracle, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	msc13Cast(t, g, me, "Green Spell", "Sorcery", "{G}", "G")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a green spell asked for a target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("a green spell offered a discard")
	}
	if c, _ := effectstest.BattlefieldCard(g, victim); c.Tapped {
		t.Error("a green spell tapped a creature")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d for a green spell", hand, got)
	}
}

func TestMsc13IcemanAndFirestarLootsWhenYouCastARedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Iceman and Firestar", msc13IcemanTypes, msc13IcemanOracle, 3, 3)
	effectstest.SeedLibrary(me, "Drawn")
	discard := effectstest.PushHandCard(g, me)
	msc13Cast(t, g, me, "Red Spell", "Sorcery", "{R}", "R")
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, discard)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand after discard one, draw one: %d, want %d", got, hand)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestMsc13IcemanAndFirestarLootDrawsNothingWhenYouDeclineToDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Iceman and Firestar", msc13IcemanTypes, msc13IcemanOracle, 3, 3)
	effectstest.SeedLibrary(me, "Drawn")
	effectstest.PushHandCard(g, me)
	msc13Cast(t, g, me, "Red Spell", "Sorcery", "{R}", "R")
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("declining the discard changed the hand: %d -> %d", hand, got)
	}
}

// --- Happy Hogan, Bodyguard ----------------------------------------------

func TestMsc13HappyHoganHasVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Happy Hogan, Bodyguard", "Legendary Creature — Human Advisor", msc13HappyOracle, 3, 3)
	effectstest.AssertKeywords(t, g, id, "vigilance")
}

func TestMsc13HappyHoganLetsTheOwnerPutItSecondFromTheTopOrOnTheBottom(t *testing.T) {
	for _, bury := range []bool{false, true} {
		name := "second from the top"
		if bury {
			name = "on the bottom"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			lib := effectstest.SeedLibrary(opp, "Top", "Next")
			theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			effectstest.CastHoldCreature(t, g, "Happy Hogan, Bodyguard", "Legendary Creature — Human Advisor", msc13HappyOracle, 3, 3)
			if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
				t.Error("Happy Hogan may target his controller's own creature")
			}
			effectstest.PickCard(t, g, me.ID, theirs)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
				t.Fatal("the controller of Happy Hogan does not choose")
			}
			ask := effectstest.PutInLibraryChoiceFor(g, opp.ID)
			if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || ask.LibraryTopDepth != 2 {
				t.Fatalf("the OWNER is asked, with the top lane at depth 2: %+v", ask)
			}
			var top, bottom []uuid.UUID
			if bury {
				bottom = []uuid.UUID{theirs}
			} else {
				top = []uuid.UUID{theirs}
			}
			if err := g.ResolvePutInLibrary(ask.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if effects.OnBattlefield(g, theirs) {
				t.Fatal("the creature is still on the battlefield")
			}
			if bury {
				if opp.Library.Cards[0].InstanceID != theirs {
					t.Error("not on the bottom of its owner's library")
				}
			} else if got := effectstest.LibraryTopIDs(opp, 3); !effectstest.SameIDs(got, []uuid.UUID{lib[0], theirs, lib[1]}) {
				t.Errorf("top three %v, want [Top Their-Bear Next]", got)
			}
		})
	}
}

// --- Hyperion, Supreme Hero ----------------------------------------------

const msc13HyperionTypes = "Legendary Creature — Eternal Hero"

func TestMsc13HyperionHasFlashAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Hyperion, Supreme Hero", msc13HyperionTypes, msc13HyperionOracle, 7, 7)
	effectstest.AssertKeywords(t, g, id, "flash", "flying")
}

func TestMsc13HyperionPreventsAllButOneOfTheDamageToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hyperion, Supreme Hero", msc13HyperionTypes, msc13HyperionOracle, 7, 7)
	src := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Test", 5, 5)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 5) })
	if got := life - me.Life; got != 1 {
		t.Errorf("I took %d from a 5-damage source, want 1", got)
	}
	// One damage is already "all but 1 prevented": nothing to prevent.
	life = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 1) })
	if got := life - me.Life; got != 1 {
		t.Errorf("I took %d from a 1-damage source, want 1", got)
	}
	// An opponent is not protected.
	oppLife := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 5) })
	if got := oppLife - opp.Life; got != 5 {
		t.Errorf("the opponent took %d, want 5", got)
	}
}

func TestMsc13HyperionPreventsAllButOneOfTheDamageToYourHeroesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hyp := effectstest.B12Push(g, me.ID, "Hyperion, Supreme Hero", msc13HyperionTypes, msc13HyperionOracle, 7, 7)
	hero := effectstest.B12Creature(g, me.ID, "My Hero", "Creature — Human Hero", 4, 9)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 4, 9)
	theirHero := effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 4, 9)
	src := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Test", 5, 9)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"Hyperion himself", hyp, 1},
		{"a Hero I control", hero, 1},
		{"a non-Hero I control", bear, 5},
		{"an opponent's Hero", theirHero, 5},
	} {
		effectstest.B27Damage(g, src, c.id, 5)
		if got := effectstest.DamageMarkedOn(g, c.id); got != c.want {
			t.Errorf("%s has %d damage from a 5-damage source, want %d", c.name, got, c.want)
		}
	}
}

func TestMsc13HyperionStopsPreventingOnceHeLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hyp := effectstest.B12Push(g, me.ID, "Hyperion, Supreme Hero", msc13HyperionTypes, msc13HyperionOracle, 7, 7)
	hero := effectstest.B12Creature(g, me.ID, "My Hero", "Creature — Human Hero", 4, 9)
	src := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Test", 5, 9)
	effectstest.B27Kill(g, hyp)
	g.RunStateChecksForTest()
	effectstest.B27Damage(g, src, hero, 5)
	if got := effectstest.DamageMarkedOn(g, hero); got != 5 {
		t.Errorf("my Hero took %d with Hyperion gone, want 5", got)
	}
}

// --- Loki's Scepter ------------------------------------------------------

func TestMsc13LokisScepterStealsUntapsAndMakesAVillainWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(victim) })
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	scepter := effectstest.CastHoldCreature(t, g, "Loki's Scepter", "Legendary Artifact", msc13ScepterOracle, 0, 0)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: scepter}); err == nil {
		t.Error("the Scepter may target a noncreature permanent")
	}
	_ = mine
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the creature is not under my control: %+v", c)
	}
	if c.Tapped {
		t.Error("the stolen creature was not untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("the stolen creature has no haste")
	}
	if !c.HasSubtype("Villain") || !c.HasSubtype("Bear") {
		t.Errorf("the stolen creature's subtypes: %q, want Bear and Villain", c.TypeLine)
	}
	effectstest.AdvanceToEndStep(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PowerUpPassTo(t, g, 1)
	back, _ := effectstest.BattlefieldCard(g, victim)
	if back.Controller != opp.ID {
		t.Error("the creature did not return at end of turn")
	}
}

func TestMsc13LokisScepterTapsForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scepter := effectstest.B12Push(g, me.ID, "Loki's Scepter", "Legendary Artifact", msc13ScepterOracle, 0, 0)
	spec, _ := effects.Lookup(msc13ScepterOracle)
	if len(spec.ManaAbilities) != 1 || !spec.ManaAbilities[0].Cost.Tap || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" {
		t.Fatalf("the Scepter's mana ability: %+v", spec.ManaAbilities)
	}
	_ = scepter
}

// --- Killmonger, Ruthless Usurper ----------------------------------------

const msc13KillmongerTypes = "Legendary Creature — Human Villain"

func TestMsc13KillmongerHasTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Killmonger, Ruthless Usurper", msc13KillmongerTypes, msc13KillmongerOracle, 3, 3)
	effectstest.AssertKeywords(t, g, id, "trample")
}

func TestMsc13KillmongerGetsPlusOnePowerForEachArtifactTheDefenderControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	km := effectstest.B12Push(g, me.ID, "Killmonger, Ruthless Usurper", msc13KillmongerTypes, msc13KillmongerOracle, 3, 3)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Other Rock", "Artifact Creature — Golem")
	effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	effectstest.B12Permanent(g, third.ID, "Third Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.DeclareAttack(t, g, opp.ID, km)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, km); got != 5 {
		t.Errorf("Killmonger's power attacking a player with two artifacts: %d, want 5", got)
	}
}

func TestMsc13KillmongerCombatDamageMakesThemSacrificeAnArtifactAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	km := effectstest.B12Push(g, me.ID, "Killmonger, Ruthless Usurper", msc13KillmongerTypes, msc13KillmongerOracle, 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	spare := effectstest.B12Permanent(g, opp.ID, "Their Spare", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.DeclareAttack(t, g, opp.ID, km)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the chosen artifact was not sacrificed")
	}
	if !g.Battlefield.Contains(spare) || !g.Battlefield.Contains(land) {
		t.Error("something other than the chosen artifact was sacrificed")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("Treasures I control: %d, want 1", got)
	}
}

func TestMsc13KillmongerStillMakesATreasureWhenTheyHaveNoArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	km := effectstest.B12Push(g, me.ID, "Killmonger, Ruthless Usurper", msc13KillmongerTypes, msc13KillmongerOracle, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, km)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("Treasures I control: %d, want 1", got)
	}
}

func TestMsc13KillmongerIgnoresNoncombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	km := effectstest.B12Push(g, me.ID, "Killmonger, Ruthless Usurper", msc13KillmongerTypes, msc13KillmongerOracle, 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(km, opp.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("Treasures after noncombat damage: %d, want 0", got)
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("an artifact was sacrificed for noncombat damage")
	}
}

// --- Lockjaw, Slobbering Teleporter --------------------------------------

const msc13LockjawTypes = "Legendary Creature — Inhuman Dog Hero"

func TestMsc13LockjawGrowsAndMakesItselfAndAnotherCreatureUnblockableAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jaw := effectstest.B12Push(g, me.ID, "Lockjaw, Slobbering Teleporter", msc13LockjawTypes, msc13LockjawOracle, 1, 1)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Test", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Test", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Test", 2, 2)
	msc13Cast(t, g, me, "Some Sorcery", "Sorcery", "{U}", "U")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, jaw); got != 1 {
		t.Fatalf("+1/+1 counters at the beginning of combat: %d, want 1", got)
	}
	effectstest.PickCard(t, g, me.ID, friend)
	effectstest.PassPriorityAroundTable(t, g)
	refused := effectstest.Msc13BlockRefused(t, g, opp.ID, blocker, jaw, friend, plain)
	if !refused[jaw] || !refused[friend] {
		t.Errorf("Lockjaw or the chosen creature was blockable: %v", refused)
	}
	if refused[plain] {
		t.Error("an unchosen creature could not be blocked")
	}
}

func TestMsc13LockjawDoesNothingWithoutANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jaw := effectstest.B12Push(g, me.ID, "Lockjaw, Slobbering Teleporter", msc13LockjawTypes, msc13LockjawOracle, 1, 1)
	effectstest.B12Creature(g, me.ID, "Friend", "Creature — Test", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, jaw); got != 0 {
		t.Errorf("+1/+1 counters with no noncreature spell: %d, want 0", got)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for with no noncreature spell")
	}
}

func TestMsc13LockjawIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jaw := effectstest.B12Push(g, me.ID, "Lockjaw, Slobbering Teleporter", msc13LockjawTypes, msc13LockjawOracle, 1, 1)
	effectstest.CastHoldCreature(t, g, "Some Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, jaw); got != 0 {
		t.Errorf("+1/+1 counters after casting only a creature spell: %d, want 0", got)
	}
}

func TestMsc13LockjawNeedsNoSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jaw := effectstest.B12Push(g, me.ID, "Lockjaw, Slobbering Teleporter", msc13LockjawTypes, msc13LockjawOracle, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Test", 2, 2)
	msc13Cast(t, g, me, "Some Sorcery", "Sorcery", "{U}", "U")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, jaw); got != 1 {
		t.Fatalf("+1/+1 counters: %d, want 1", got)
	}
	refused := effectstest.Msc13BlockRefused(t, g, opp.ID, blocker, jaw)
	if !refused[jaw] {
		t.Error("Lockjaw alone was blockable")
	}
}

// --- Invisible Woman -----------------------------------------------------

const msc13InvisibleTypes = "Legendary Creature — Human Hero"

func TestMsc13InvisibleWomanMakesAWallAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Invisible Woman", msc13InvisibleTypes, msc13InvisibleOracle, 3, 3)
	msc13Cast(t, g, me, "Some Sorcery", "Sorcery", "{W}", "W")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Wall"); got != 1 {
		t.Fatalf("Walls I control: %d, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Wall" {
			continue
		}
		if !c.IsToken() {
			t.Error("the Wall is not a token")
		}
		if p, tg := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 0 || tg != 3 {
			t.Errorf("the Wall is %d/%d, want 0/3", p, tg)
		}
		effectstest.AssertKeywords(t, g, c.InstanceID, "defender", "reach")
	}
}

func TestMsc13InvisibleWomanMakesNoWallWithoutANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Invisible Woman", msc13InvisibleTypes, msc13InvisibleOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Wall"); got != 0 {
		t.Errorf("Walls with no noncreature spell: %d, want 0", got)
	}
}

func TestMsc13InvisibleWomanPaymentPumpsAndMakesATargetUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Invisible Woman", msc13InvisibleTypes, msc13InvisibleOracle, 3, 3)
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Test", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Test", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{R}{G}{W}{U}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PickCard(t, g, me.ID, runner)
	effectstest.PassPriorityAroundTable(t, g)
	// Two creatures I control: Invisible Woman and the runner.
	if got := effectstest.EffectivePower(t, g, runner); got != 4 {
		t.Errorf("the runner's power: %d, want 2 + 2", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, runner); err == nil {
		t.Error("the targeted creature was blocked")
	}
}

func TestMsc13InvisibleWomanDecliningThePaymentDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Invisible Woman", msc13InvisibleTypes, msc13InvisibleOracle, 3, 3)
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Test", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, runner)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("declining the payment still asked for a target")
	}
	if got := effectstest.EffectivePower(t, g, runner); got != 2 {
		t.Errorf("the runner's power: %d, want 2", got)
	}
}

func TestMsc13InvisibleWomanCannotPayWithoutTheFourColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Invisible Woman", msc13InvisibleTypes, msc13InvisibleOracle, 3, 3)
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Test", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, runner)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{R}{G}{W}")
	// A "yes" the pool cannot fund degrades to a decline: nothing is spent.
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an unfunded payment still raised the reflexive trigger")
	}
	if got := effectstest.EffectivePower(t, g, runner); got != 2 {
		t.Errorf("the runner's power after an unfunded payment: %d, want 2", got)
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("mana left after an unfunded payment: %d, want all 3", got)
	}
}

// --- Jocasta, Automaton Avenger ------------------------------------------

const msc13JocastaTypes = "Legendary Artifact Creature — Robot Hero"

func msc13Commander(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
	})
}

func msc13JocastaInGraveyard(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Jocasta, Automaton Avenger", TypeLine: msc13JocastaTypes,
		OracleID: msc13JocastaOracle, ManaCost: "{3}", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func TestMsc13JocastaFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Jocasta, Automaton Avenger", msc13JocastaTypes, msc13JocastaOracle, 2, 2)
	effectstest.AssertKeywords(t, g, id, "flying")
}

func TestMsc13JocastaGrowsWhenYourCommanderDealsCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	joc := effectstest.B12Push(g, me.ID, "Jocasta, Automaton Avenger", msc13JocastaTypes, msc13JocastaOracle, 2, 2)
	cmdr := msc13Commander(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, cmdr)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, joc); got != 1 {
		t.Errorf("+1/+1 counters after my commander connected: %d, want 1", got)
	}
}

func TestMsc13JocastaIgnoresANonCommanderCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	joc := effectstest.B12Push(g, me.ID, "Jocasta, Automaton Avenger", msc13JocastaTypes, msc13JocastaOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, joc); got != 0 {
		t.Errorf("+1/+1 counters after a non-commander connected: %d, want 0", got)
	}
}

func TestMsc13JocastaReturnsTappedAndAttackingWhenYouAttackWithYourCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr := msc13Commander(g, me.ID)
	joc := msc13JocastaInGraveyard(me)
	effectstest.DeclareAttack(t, g, opp.ID, cmdr)
	if n := effectstest.TriggersOnStackFrom(g, joc); n != 1 {
		t.Fatalf("Jocasta triggers on the stack = %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if pick := effectstest.LatestOptionPickFor(g, me.ID); pick != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, joc)
	if !ok {
		t.Fatal("Jocasta did not return to the battlefield")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil || c.Controller != me.ID {
		t.Errorf("tapped=%v attacking=%v controller=%v, want tapped, attacking, mine", c.Tapped, c.AttackingTarget, c.Controller)
	}
}

func TestMsc13JocastaStaysInTheGraveyardWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr := msc13Commander(g, me.ID)
	joc := msc13JocastaInGraveyard(me)
	effectstest.DeclareAttack(t, g, opp.ID, cmdr)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(joc) {
		t.Error("Jocasta left the graveyard after the player declined")
	}
}

func TestMsc13JocastaDoesNotTriggerForANonCommanderAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	joc := msc13JocastaInGraveyard(me)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if n := effectstest.TriggersOnStackFrom(g, joc); n != 0 {
		t.Errorf("a non-commander attack triggered Jocasta %d times", n)
	}
}

func TestMsc13JocastaOnTheBattlefieldDoesNotUseTheGraveyardClause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr := msc13Commander(g, me.ID)
	joc := effectstest.B12Push(g, me.ID, "Jocasta, Automaton Avenger", msc13JocastaTypes, msc13JocastaOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cmdr)
	if n := effectstest.TriggersOnStackFrom(g, joc); n != 0 {
		t.Errorf("Jocasta on the battlefield triggered %d times for the graveyard clause", n)
	}
}
