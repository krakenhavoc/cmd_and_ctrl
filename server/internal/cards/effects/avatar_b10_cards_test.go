package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b10_cards_test.go — Avatar: The Last Airbender batch 10 (#4459).

const (
	avB10RuinousOracle    = "7d3a0216-871b-4c1b-adb1-de99b832f577"
	avB10SoldOutOracle    = "6a69d4b1-72dd-4599-a16d-32299de5b0b9"
	avB10SwampsnareOracle = "fc532d1d-6cf5-4b59-afb3-325d0bc0db77"
	avB10TundraTankOracle = "2ded4b02-7aa4-46a1-b10f-8480eb942cd4"
	avB10WolfbatOracle    = "adc47a25-daab-49c7-93a5-5bbf2bc20cf9"
	avB10ConvictionOracle = "f034e8a3-82f7-4c70-93ea-0bf2688f66e1"
	avB10BoarQPineOracle  = "551e4356-5b1f-475c-a64b-9c50ea9b0368"
	avB10BumiBashOracle   = "629f247c-1e64-4c32-8892-0716ed469d1c"
	avB10CaveOracle       = "c2df54f1-a2db-4de3-bc96-5c07f4594752"
)

func TestAvatarB10CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		theRiseOfSozinOracleID:        "The Rise of Sozin",
		theRiseOfSozinOracleID + "#1": "Fire Lord Sozin",
		avB10RuinousOracle:            "Ruinous Waterbending",
		avB10SoldOutOracle:            "Sold Out",
		avB10SwampsnareOracle:         "Swampsnare Trap",
		avB10TundraTankOracle:         "Tundra Tank",
		avB10WolfbatOracle:            "Wolfbat",
		avB10ConvictionOracle:         "Zuko's Conviction",
		avB10BoarQPineOracle:          "Boar-q-pine",
		avB10BumiBashOracle:           "Bumi Bash",
		avB10CaveOracle:               "The Cave of Two Lovers",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// avB10Cast puts `c` into the active seat's hand, walks to a main phase
// and casts it with `params`.
func avB10Cast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

func avB10Refs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func avB10Graveyard(g *game.Game, p *game.Player, name, typeLine, manaCost, oracle string) uuid.UUID {
	return pushToGraveyard(g, p, game.Card{Name: name, TypeLine: typeLine, ManaCost: manaCost, OracleID: oracle, Power: 2, Toughness: 2})
}

// --- Boar-q-pine ----------------------------------------------------------------

func TestAvatarB10BoarQPineGrowsOnNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	boar := pushCatalogPermanent(g, me.ID, "Boar-q-pine", "Creature — Boar Porcupine", avB10BoarQPineOracle, false)

	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, boar, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after a creature spell, want 0", n)
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, boar, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after a noncreature spell, want 1", n)
	}
}

// --- Bumi Bash ------------------------------------------------------------------

func TestAvatarB10BumiBashDealsDamageEqualToYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		pushEarthbendLand(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	pushEarthbendLand(g, opp.ID, "Forest", "Basic Land — Forest")
	three := pushVanillaCreature(g, opp.ID, "Three", 1, 3)
	four := pushVanillaCreature(g, opp.ID, "Four", 1, 4)

	if _, err := avB10Cast(t, g, game.Card{Name: "Bumi Bash", TypeLine: "Sorcery", OracleID: avB10BumiBashOracle},
		game.CastSpellParams{Modes: []int{0}, Targets: avB10Refs(three)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, three) {
		t.Error("the 1/3 survived 3 damage (three lands)")
	}
	if _, err := avB10Cast(t, g, game.Card{Name: "Bumi Bash", TypeLine: "Sorcery", OracleID: avB10BumiBashOracle},
		game.CastSpellParams{Modes: []int{0}, Targets: avB10Refs(four)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, four) {
		t.Error("the 1/4 died to 3 damage; only the caster's lands count")
	}
}

func TestAvatarB10BumiBashDestroysANonbasicOrLandCreatureButNotABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	basic := pushEarthbendLand(g, opp.ID, "Forest", "Basic Land — Forest")
	nonbasic := pushEarthbendLand(g, opp.ID, "Strange Place", "Land")
	animated := pushPermanent(g, opp.ID, game.Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1})
	_ = me

	if _, err := avB10Cast(t, g, game.Card{Name: "Bumi Bash", TypeLine: "Sorcery", OracleID: avB10BumiBashOracle},
		game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: avB10Refs(basic)}); err == nil {
		t.Fatal("a basic land that is not a creature was a legal target")
	}
	for _, id := range []uuid.UUID{nonbasic, animated} {
		if _, err := avB10Cast(t, g, game.Card{Name: "Bumi Bash", TypeLine: "Sorcery", OracleID: avB10BumiBashOracle},
			game.CastSpellParams{Modes: []int{1}, Targets: avB10Refs(id)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if onBattlefield(g, id) {
			t.Errorf("%s survived the destroy mode", id)
		}
	}
}

// --- Sold Out -------------------------------------------------------------------

func TestAvatarB10SoldOutExilesAndCluesOnlyADamagedCreature(t *testing.T) {
	for _, damaged := range []bool{false, true} {
		name := "undamaged"
		if damaged {
			name = "damaged"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			victim := pushVanillaCreature(g, opp.ID, "Victim", 4, 4)
			if damaged {
				pinger := pushVanillaCreature(g, me.ID, "Pinger", 1, 1)
				g.WithWriteLock(func() {
					_ = g.DealDamageToCreatureKeyedForEffect(pinger, victim, 1, game.KeyedTail{})
				})
			}
			castCatalogSpell(t, g, "Sold Out", "Instant", avB10SoldOutOracle, avB10Refs(victim))
			passPriorityAroundTable(t, g)
			if !inExile(g, victim) {
				t.Fatal("the creature was not exiled")
			}
			want := 0
			if damaged {
				want = 1
			}
			if got := countOnBattlefield(g, "Clue", me.ID); got != want {
				t.Errorf("%d Clues, want %d", got, want)
			}
		})
	}
}

// --- Swampsnare Trap ------------------------------------------------------------

func TestAvatarB10SwampsnareTrapCostsLessAtAFlierAndShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	flier := pushPermanent(g, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 6, Toughness: 4, Keywords: []string{"flying"}})
	bear := pushVanillaCreature(g, opp.ID, "Bear", 6, 4)
	if got := selfPriced(t, g, me, avB10SwampsnareOracle, "Enchantment — Aura", "{2}{B}", game.ZoneHand, avB10Refs(flier)).ManaValue(); got != 2 {
		t.Errorf("at a flier: %d, want 2", got)
	}
	if got := selfPriced(t, g, me, avB10SwampsnareOracle, "Enchantment — Aura", "{2}{B}", game.ZoneHand, avB10Refs(bear)).ManaValue(); got != 3 {
		t.Errorf("at a ground creature: %d, want 3", got)
	}

	castCatalogSpell(t, g, "Swampsnare Trap", "Enchantment — Aura", avB10SwampsnareOracle, avB10Refs(bear))
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, bear); p != 1 || th != 1 {
		t.Errorf("enchanted 6/4 is %d/%d, want 1/1", p, th)
	}
}

// --- Tundra Tank ----------------------------------------------------------------

func TestAvatarB10TundraTankEntryGrantsIndestructibleToYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)

	castCatalogSpell(t, g, "Tundra Tank", "Artifact — Vehicle", avB10TundraTankOracle, nil)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Fatal("an opponent's creature was a legal target")
		}
	}
	pickTriggerTarget(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("the targeted creature has no indestructible")
	}
	if hasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("the opponent's creature gained indestructible")
	}
}

func TestAvatarB10TundraTankCrewedAttackFirebendsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tank := pushCatalogPermanent(g, me.ID, "Tundra Tank", "Artifact — Vehicle", avB10TundraTankOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tank {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
			}
		}
	})
	crew := pushVanillaCreature(g, me.ID, "Crewer", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tank, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("Crew 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(tank, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool %d after the crewed Tank attacked, want 1 ({R} from firebending 1)", got)
	}
}

// --- Wolfbat --------------------------------------------------------------------

func TestAvatarB10WolfbatReturnsOnTheSecondDrawWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	// Not the active player: the draw step drew nothing for them.
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bat := avB10Graveyard(g, me, "Wolfbat", "Creature — Wolf Bat", "{2}{B}", avB10WolfbatOracle)

	hobA5Draw(t, g, me, nil)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("the first draw asked for {B}")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	fillPoolColored(me, "B", 1) // after the trigger resolved: passing moved the step
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if inGraveyardOf(g, me.ID, bat) {
		t.Fatal("Wolfbat is still in the graveyard after paying {B}")
	}
	back := findBattlefieldByName(g, "Wolfbat")
	if back == uuid.Nil {
		t.Fatal("Wolfbat did not return to the battlefield")
	}
	if n := countersOn(g, back, game.CounterFinality); n != 1 {
		t.Errorf("%d finality counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, back, "flying") {
		t.Error("Wolfbat has no flying")
	}

	// A third draw does not trigger again.
	hobA5Draw(t, g, me, nil)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the third draw asked again")
	}
}

func TestAvatarB10WolfbatDeclinedStaysInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bat := avB10Graveyard(g, me, "Wolfbat", "Creature — Wolf Bat", "{2}{B}", avB10WolfbatOracle)
	hobA5Draw(t, g, me, nil)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	fillPoolColored(me, "B", 1)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !inGraveyardOf(g, me.ID, bat) {
		t.Error("Wolfbat left the graveyard though {B} was not paid")
	}
}

// --- Zuko's Conviction ----------------------------------------------------------

func TestAvatarB10ZukosConvictionReturnsToHandOrBattlefieldTapped(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		name := "unkicked"
		if kicked {
			name = "kicked"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			dead := avB10Graveyard(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
			params := game.CastSpellParams{Strict: true, Targets: avB10Refs(dead)}
			if kicked {
				params.OptionalCosts = []int{0}
				fillPool(me, 4)
			}
			fillPoolColored(me, "B", 1)
			if _, err := avB10Cast(t, g, game.Card{Name: "Zuko's Conviction", TypeLine: "Instant", ManaCost: "{B}", OracleID: avB10ConvictionOracle}, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			if kicked {
				id := findBattlefieldByName(g, "Dead Bear")
				if id == uuid.Nil {
					t.Fatal("kicked: the card is not on the battlefield")
				}
				if !isTapped(g, id) {
					t.Error("kicked: it entered untapped")
				}
				if me.Hand.Contains(dead) {
					t.Error("kicked: it also went to the hand")
				}
				return
			}
			if !me.Hand.Contains(dead) {
				t.Error("unkicked: the card is not in hand")
			}
			if findBattlefieldByName(g, "Dead Bear") != uuid.Nil {
				t.Error("unkicked: it went to the battlefield")
			}
		})
	}
}

func TestAvatarB10ZukosConvictionCannotTargetAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := avB10Graveyard(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	fillPoolColored(me, "B", 1)
	if _, err := avB10Cast(t, g, game.Card{Name: "Zuko's Conviction", TypeLine: "Instant", ManaCost: "{B}", OracleID: avB10ConvictionOracle},
		game.CastSpellParams{Strict: true, Targets: avB10Refs(theirs)}); err == nil {
		t.Fatal("a creature card in an opponent's graveyard was a legal target")
	}
}

// --- Ruinous Waterbending -------------------------------------------------------

func TestAvatarB10RuinousWaterbendingShrinksAndGainsLifeOnlyWhenWaterbent(t *testing.T) {
	for _, paid := range []bool{false, true} {
		name := "unpaid"
		if paid {
			name = "waterbent"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			small := pushVanillaCreature(g, opp.ID, "Small", 2, 2)
			mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
			big := pushVanillaCreature(g, opp.ID, "Big", 4, 4)
			params := game.CastSpellParams{Strict: true}
			fillPoolColored(me, "B", 2)
			fillPool(me, 1)
			if paid {
				taps := []uuid.UUID{
					pushVanillaCreature(g, me.ID, "Tapper", 5, 5),
					pushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false),
				}
				params.OptionalCosts = []int{0}
				params.TapIDs = taps
				fillPool(me, 2)
			}
			life := me.Life
			if _, err := avB10Cast(t, g, game.Card{Name: "Ruinous Waterbending", TypeLine: "Sorcery — Lesson", ManaCost: "{1}{B}{B}", OracleID: avB10RuinousOracle}, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			passPriorityAroundTable(t, g)
			if onBattlefield(g, small) || onBattlefield(g, mine) {
				t.Fatal("a creature with toughness 2 or less survived -2/-2")
			}
			if p, th := effectivePTOf(t, g, big); p != 2 || th != 2 {
				t.Errorf("the 4/4 is %d/%d, want 2/2", p, th)
			}
			want := life
			if paid {
				want = life + 2
			}
			if me.Life != want {
				t.Errorf("life %d → %d, want %d", life, me.Life, want)
			}
			if !paid {
				return
			}
			// "This turn": a later death still gains 1.
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(big) })
			g.RunStateChecksForTest()
			passPriorityAroundTable(t, g)
			if me.Life != want+1 {
				t.Errorf("life %d after a later death, want %d", me.Life, want+1)
			}
		})
	}
}

func TestAvatarB10RuinousWaterbendingCannotBeWaterbentShort(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	tapper := pushVanillaCreature(g, me.ID, "Tapper", 1, 1)
	fillPoolColored(me, "B", 2)
	fillPool(me, 1)
	if _, err := avB10Cast(t, g, game.Card{Name: "Ruinous Waterbending", TypeLine: "Sorcery — Lesson", ManaCost: "{1}{B}{B}", OracleID: avB10RuinousOracle},
		game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TapIDs: []uuid.UUID{tapper}}); err == nil {
		t.Fatal("one tap and no spare mana paid waterbend {4}")
	}
}

// --- The Cave of Two Lovers -----------------------------------------------------

// avB10Fillers puts n plain cards on top of p's library, so the draw steps
// a test walks through draw them and not the cards it is about.
func avB10Fillers(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func TestAvatarB10CaveOfTwoLoversChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	mountain := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
		avB10Fillers(me, 3)
	})

	saga := castCatalogSpell(t, g, "The Cave of Two Lovers", "Enchantment — Saga", avB10CaveOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Ally", me.ID); got != 2 {
		t.Fatalf("%d Allies after chapter I, want 2", got)
	}
	ally := findBattlefieldByName(g, "Ally")
	if p, th := effectivePTOf(t, g, ally); p != 1 || th != 1 {
		t.Errorf("Ally is %d/%d, want 1/1", p, th)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Mountain")
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mountain) {
		t.Fatal("chapter II did not put the Mountain into the hand")
	}

	advanceToPrecombatMainOf(t, g, seat)
	pickCard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, land, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters on the earthbent land, want 3", got)
	}
	if _, _, creatureLand, hasty := earthbentBody(t, g, land); !creatureLand || !hasty {
		t.Error("the land is not a hasty land creature")
	}
	g.RunStateChecksForTest()
	if onBattlefield(g, saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestAvatarB10CaveOfTwoLoversFindsACaveButNotAForest(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	cave, forest := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: cave, Name: "Deep Cave", TypeLine: "Land — Cave", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		avB10Fillers(me, 3)
	})
	castCatalogSpell(t, g, "The Cave of Two Lovers", "Enchantment — Saga", avB10CaveOracle, nil)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest was offered")
	}
	answerSearchNamed(t, g, me.ID, "Deep Cave")
	if !me.Hand.Contains(cave) {
		t.Error("the Cave card is not in hand")
	}
}

// --- The Rise of Sozin // Fire Lord Sozin ---------------------------------------

func riseOfSozinRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "The Rise of Sozin // Fire Lord Sozin",
		Layout:   "transform",
		TypeLine: "Enchantment — Saga // Legendary Creature — Human Noble",
		OracleID: uuid.MustParse(theRiseOfSozinOracleID),
		CardFaces: []cards.CardFace{
			{Name: "The Rise of Sozin", TypeLine: "Enchantment — Saga", ManaCost: "{4}{B}{B}", Colors: []string{"B"}},
			{Name: "Fire Lord Sozin", TypeLine: "Legendary Creature — Human Noble", Power: "5", Toughness: "5", Colors: []string{"B", "R"}},
		},
	}
}

func TestAvatarB10RiseOfSozinWrathsNamesAndExilesThenTransforms(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 5, 5)

	saga := importAndCast(t, g, riseOfSozinRow(), me)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	if onBattlefield(g, mine) || onBattlefield(g, theirs) {
		t.Fatal("chapter I left a creature alive")
	}

	// Chapter II: five Bolts across their zones; up to four are exiled.
	bolt := game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}"}
	var bolts []uuid.UUID
	bolts = append(bolts, pushToGraveyard(g, opp, bolt), pushToGraveyard(g, opp, bolt))
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			c := bolt
			c.InstanceID, c.Owner, c.Controller = uuid.New(), opp.ID, opp.ID
			opp.Library.PushTop(c)
			bolts = append(bolts, c.InstanceID)
		}
		c := bolt
		c.InstanceID, c.Owner, c.Controller = uuid.New(), opp.ID, opp.ID
		opp.Hand.PushTop(c)
		bolts = append(bolts, c.InstanceID)
	})
	other := pushToGraveyard(g, opp, game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"})

	advanceToPrecombatMainOf(t, g, seat)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "Lightning Bolt")
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after naming the card")
	}
	if searchOptionNamed(g, c, "Shock") != uuid.Nil {
		t.Error("a card with another name was offered")
	}
	beginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, bolts); err == nil {
		t.Error("all five copies were taken; the chapter says up to four")
	}
	c = searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the refused answer closed the search prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, bolts[:4]); err != nil {
		t.Fatalf("take four: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range bolts[:4] {
		if !inExile(g, id) {
			t.Errorf("Bolt %s was not exiled", id)
		}
	}
	if inExile(g, bolts[4]) || inExile(g, other) {
		t.Error("a card that was not taken was exiled")
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	assertReturnedAsAvatar(t, g, saga, me.ID, "Fire Lord Sozin", 5, 5)
	assertFaceInvariant(t, g)
}

func TestAvatarB10FireLordSozinFirebendsAndPaysXToReanimate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sozin := putBackFaceOntoBattlefield(t, g, riseOfSozinRow(), me)
	if !hasEffectiveKeyword(t, g, sozin, "menace") {
		t.Error("Fire Lord Sozin has no menace")
	}
	two := avB10Graveyard(g, opp, "Two", "Creature — Bear", "{1}{B}", "")
	one := avB10Graveyard(g, opp, "One", "Creature — Rat", "{B}", "")
	zero := avB10Graveyard(g, opp, "Zero", "Artifact Creature — Construct", "", "")
	three := avB10Graveyard(g, opp, "Three", "Creature — Ogre", "{2}{B}", "")
	mineDead := avB10Graveyard(g, me, "My Dead", "Creature — Bear", "{1}", "")

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sozin, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 3 {
		t.Fatalf("pool %d after attacking, want 3 (firebending 3)", got)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	c := payXPrompt(t, g, me.ID)
	if c.PayAmount.Max != 3 {
		t.Fatalf("X can go to %d, want 3 (the firebending mana lasts through combat)", c.PayAmount.Max)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err != nil {
		t.Fatalf("answer X: %v", err)
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no reflexive target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, avB10Refs(three, one)); err == nil {
		t.Fatal("targets with total mana value 4 were accepted for X = 3")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, avB10Refs(mineDead)); err == nil {
		t.Fatal("a card from another player's graveyard was accepted")
	}
	answerTriggerTargets(t, g, me.ID, two, one, zero)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"Two": two, "One": one, "Zero": zero} {
		if inGraveyardOf(g, opp.ID, id) {
			t.Errorf("%s is still in their graveyard", name)
		}
		back := findBattlefieldByName(g, name)
		if back == uuid.Nil {
			t.Errorf("%s is not on the battlefield", name)
			continue
		}
		if c := battlefieldCardFor(g, back); c.Controller != me.ID {
			t.Errorf("%s is controlled by %s, want the Sozin player", name, c.Controller)
		}
	}
	if !inGraveyardOf(g, opp.ID, three) {
		t.Error("the untargeted Three left the graveyard")
	}
}

func TestAvatarB10FireLordSozinDeclinedReanimatesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sozin := putBackFaceOntoBattlefield(t, g, riseOfSozinRow(), me)
	two := avB10Graveyard(g, opp, "Two", "Creature — Bear", "{1}{B}", "")

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sozin, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	c := payXPrompt(t, g, me.ID)
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer X: %v", err)
	}
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("declining the payment still made the reflexive trigger")
	}
	if !inGraveyardOf(g, opp.ID, two) {
		t.Error("a creature came back without the payment")
	}
}
