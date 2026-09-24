package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_spend_rider_test.go — the catalog half of #1547: mana that does
// something when it is spent. The engine contract (the filter
// vocabulary, one production is one "that mana") is pinned in
// game/mana_spend_rider_test.go; this file pins the printed cards end
// to end, through a real cast under strict mana.

const (
	pyromancersGogglesOracle = "f76bcbfe-483f-4e63-8425-76feca1abf3e"
	hallOfTheBanditLordOracl = "32fe7ac4-86f5-44af-9f73-ee8f6a9ce2ba"
	boseijuWhoSheltersOracle = "36937483-30cb-449a-8028-75017a124922"
	biophagusOracle          = "7c06a368-3306-48e0-825e-a8fdc81fa343"
	scaledNurturerOracle     = "13b96709-0e88-476b-9485-956e682bb818"
)

// strict is the payment every test here makes: the engine spends the
// pool, so the tokens — and their riders — are recorded.
var strict = game.CastSpellParams{Strict: true}

// castFromHandErr puts a card in `p`'s hand and casts it, returning the
// error rather than failing — the refusal tests need it.
func castFromHandErr(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID,
		})
	})
	return id, g.CastSpell(p.ID, id, params)
}

// tapCavernFor taps a Cavern of Souls for one coloured mana of `color`.
func tapCavernFor(t *testing.T, g *game.Game, owner, cavern uuid.UUID, color string) {
	t.Helper()
	activateManaFor(t, g, owner, cavern, 1, game.ManaAbilityParams{})
	pick := riderLatestManaPick(g, owner)
	if pick == nil {
		t.Fatal("Cavern queued no colour pick")
	}
	if err := g.ResolveManaChoice(pick.ID, owner, color); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
}

// counterByEffect runs the counter verb every catalog counterspell
// calls, and reports whether the spell is still on the stack after it.
func counterByEffect(t *testing.T, g *game.Game, spell uuid.UUID) (stillOnStack bool) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
		stillOnStack = g.Stack.Contains(spell)
	})
	return stillOnStack
}

// --- Cavern of Souls ------------------------------------------------

// The headline: an Elf cast with Cavern mana named Elf survives a real
// Counterspell, and resolves.
func TestCavernOfSoulsCreatureOfTheChosenTypeCantBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
	tapCavernFor(t, g, me.ID, cavern, "G")

	elf := castFromHandForTest(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", strict)
	if item := g.StackMeta[elf]; !item.SpellCantBeCounteredByMana() {
		t.Fatalf("the Elf's stack item does not record the Cavern rider: %+v", item.Paid.Mana)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	batch01OpponentCasts(t, g, opp, "Counterspell", b14CounterspellOracle, "{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: elf}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(elf) {
		t.Fatal("the Cavern-cast Elf was countered")
	}
}

// The other three ways to cast the same creature are all counterable:
// ordinary mana, Cavern's own COLOURLESS half, and a creature of
// another type (which Cavern's coloured mana cannot even pay for).
func TestCavernOfSoulsOtherwisePaidSpellsCanBeCountered(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		cost     string
		fund     func(t *testing.T, g *game.Game, me *game.Player, cavern uuid.UUID)
	}{
		{"an Elf paid with a Forest", "Creature — Elf Druid", "{G}", func(t *testing.T, g *game.Game, me *game.Player, _ uuid.UUID) {
			floatForTest(g, me, "G")
		}},
		{"an Elf whose generic half the colorless ability paid", "Creature — Elf Warrior", "{1}{G}", func(t *testing.T, g *game.Game, me *game.Player, cavern uuid.UUID) {
			activateManaFor(t, g, me.ID, cavern, 0, game.ManaAbilityParams{})
			floatForTest(g, me, "G")
		}},
		{"a Goblin, with Cavern mana named Elf floating beside the Mountain", "Creature — Goblin", "{R}", func(t *testing.T, g *game.Game, me *game.Player, cavern uuid.UUID) {
			tapCavernFor(t, g, me.ID, cavern, "R")
			floatForTest(g, me, "R")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
			tc.fund(t, g, me, cavern)

			id := castFromHandForTest(t, g, me, "Creature", tc.typeLine, tc.cost, "", strict)
			if counterByEffect(t, g, id) {
				t.Error("the spell was not countered — the rider fired on mana that did not carry it")
			}
		})
	}
}

// The declared caveat, pinned: with strict mana off the engine spends
// nothing (the payment is OnPaper), so no token rides onto the spell and
// the Elf can be countered — the weaker-than-printed direction.
func TestCavernOfSoulsWithStrictManaOffIsCounterable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
	tapCavernFor(t, g, me.ID, cavern, "G")

	elf := castFromHandForTest(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", game.CastSpellParams{})
	if !g.StackMeta[elf].Paid.OnPaper {
		t.Fatal("setup: a permissive cast recorded a payment")
	}
	if counterByEffect(t, g, elf) {
		t.Error("a permissive-mode Cavern cast was uncounterable, which the caveat says it is not")
	}
}

// The Cavern's restriction still holds under the rider: its coloured
// mana cannot pay for the Goblin at all, so the one token left in the
// pool after the Goblin cast above is the Cavern's.
func TestCavernOfSoulsManaDoesNotPayForTheWrongType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
	tapCavernFor(t, g, me.ID, cavern, "R")

	if _, err := castFromHandErr(g, me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "", strict); err == nil {
		t.Fatal("Cavern mana named Elf paid for a Goblin")
	}
}

// --- Delighted Halfling and Boseiju: the same rider, other filters --

func TestDelightedHalflingLegendarySpellCantBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	halfling := seedPermanentWithOracle(g, me.ID, "Delighted Halfling", "Creature — Halfling Citizen", delightedHalflingOracl)
	activateManaFor(t, g, me.ID, halfling, 1, game.ManaAbilityParams{Colors: []string{"G"}})

	id := castFromHandForTest(t, g, me, "Yeva, Nature's Herald", "Legendary Creature — Elf Shaman", "{G}", "", strict)
	if counterByEffect(t, g, id) != true {
		t.Error("a legendary spell cast with the Halfling's coloured mana was countered")
	}
}

// Boseiju's {C} pays for anything; only an instant or sorcery is
// protected by it.
func TestBoseijuProtectsOnlyInstantsAndSorceries(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		protected      bool
	}{
		{"an instant", "Instant", true},
		{"a sorcery", "Sorcery", true},
		{"a creature", "Creature — Bear", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			boseiju := seedPermanentWithOracle(g, me.ID, "Boseiju, Who Shelters All", "Legendary Land", boseijuWhoSheltersOracle)
			life := me.Life
			activateManaFor(t, g, me.ID, boseiju, 0, game.ManaAbilityParams{})
			if me.Life != life-2 {
				t.Fatalf("Boseiju cost %d life, want 2", life-me.Life)
			}
			id := castFromHandForTest(t, g, me, "Spell", tc.typeLine, "{1}", "", strict)
			if got := counterByEffect(t, g, id); got != tc.protected {
				t.Errorf("still on the stack after a counter = %v, want %v", got, tc.protected)
			}
		})
	}
}

// --- Pyromancer's Goggles -------------------------------------------

// A Lightning Bolt paid with the Goggles' {R} is copied: six damage.
func TestPyromancersGogglesCopiesARedInstantPaidWithItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	goggles := seedPermanentWithOracle(g, me.ID, "Pyromancer's Goggles", "Legendary Artifact", pyromancersGogglesOracle)
	activateManaFor(t, g, me.ID, goggles, 0, game.ManaAbilityParams{})
	before := lifeOf(g, victim.ID)

	params := strict
	params.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}
	bolt := castFromHandForTest(t, g, me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle, params)
	// Placed by the state check at the bottom of the cast, ABOVE the
	// Bolt, so the copy resolves first.
	trig := triggerOnStack(g, goggles)
	if trig == nil {
		t.Fatal("the Goggles trigger is not on the stack after the cast")
	}
	if trig.Seq < g.StackMeta[bolt].Seq {
		t.Error("the Goggles trigger is below the spell it copies")
	}
	settleKeepingCopyTargets(t, g, me.ID, victim.ID, 1)

	if got := lifeOf(g, victim.ID); got != before-6 {
		t.Errorf("victim life = %d, want %d (the Bolt and its copy)", got, before-6)
	}
}

// The rider fires only on what its filter names. The Goggles' {R}
// paying for a red CREATURE, or for a colourless artifact, copies
// nothing and leaves the rider un-applied on the record.
func TestPyromancersGogglesRiderDoesNotFireOutsideItsFilter(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, cost string }{
		{"a red creature", "Creature — Goblin", "{R}"},
		{"a colourless artifact", "Artifact", "{1}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			goggles := seedPermanentWithOracle(g, me.ID, "Pyromancer's Goggles", "Legendary Artifact", pyromancersGogglesOracle)
			activateManaFor(t, g, me.ID, goggles, 0, game.ManaAbilityParams{})

			id := castFromHandForTest(t, g, me, "Spell", tc.typeLine, tc.cost, "", strict)
			if triggerOnStack(g, goggles) != nil || len(g.PendingTriggers) != 0 {
				t.Fatal("the Goggles triggered on a spell its filter does not name")
			}
			paid := g.StackMeta[id].Paid.Mana
			if len(paid) != 1 || len(paid[0].Riders) != 1 {
				t.Fatalf("the payment record = %+v, want the one Goggles token with its rider", paid)
			}
			if paid[0].Riders[0].Applied {
				t.Error("the rider is stamped Applied on a spell outside its filter")
			}
		})
	}
}

// The auto-tapper is a mana source planner, not a payment path of its
// own: a Bolt cast with auto-tap off an untapped Goggles taps the
// Goggles, spends its {R}, and is copied.
func TestPyromancersGogglesRiderFiresThroughTheAutoTapper(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	goggles := seedPermanentWithOracle(g, me.ID, "Pyromancer's Goggles", "Legendary Artifact", pyromancersGogglesOracle)
	before := lifeOf(g, victim.ID)

	params := game.CastSpellParams{
		Strict:  true,
		AutoTap: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}
	castFromHandForTest(t, g, me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle, params)
	if !battlefieldCardCopy(t, g, goggles).Tapped {
		t.Fatal("the auto-tapper did not tap the Goggles")
	}
	settleKeepingCopyTargets(t, g, me.ID, victim.ID, 1)

	if got := lifeOf(g, victim.ID); got != before-6 {
		t.Errorf("victim life = %d, want %d", got, before-6)
	}
}

// --- Hall of the Bandit Lord ----------------------------------------

// A creature cast with the Hall's {C} has haste on the battlefield; a
// creature cast without it does not.
func TestHallOfTheBanditLordGrantsHaste(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hall  bool
		haste bool
	}{
		{"paid with the Hall's mana", true, true},
		{"paid with a Wastes", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			hall := seedPermanentWithOracle(g, me.ID, "Hall of the Bandit Lord", "Legendary Land", hallOfTheBanditLordOracl)
			if tc.hall {
				life := me.Life
				activateManaFor(t, g, me.ID, hall, 0, game.ManaAbilityParams{})
				if me.Life != life-3 {
					t.Fatalf("the Hall cost %d life, want 3", life-me.Life)
				}
			} else {
				floatForTest(g, me, "C")
			}
			floatForTest(g, me, "G")
			bear := castFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", strict)
			passPriorityAroundTable(t, g)

			if got := effectiveAbilitiesContain(t, g, bear, "haste"); got != tc.haste {
				t.Errorf("haste = %v, want %v", got, tc.haste)
			}
		})
	}
}

// "If that mana is spent on a creature spell": the Hall's {C} paying
// for an artifact spell leaves the artifact without haste — the rider
// never applied.
func TestHallOfTheBanditLordDoesNotHasteANoncreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hall := seedPermanentWithOracle(g, me.ID, "Hall of the Bandit Lord", "Legendary Land", hallOfTheBanditLordOracl)
	activateManaFor(t, g, me.ID, hall, 0, game.ManaAbilityParams{})

	rock := castFromHandForTest(t, g, me, "Mind Stone", "Artifact", "{1}", "", strict)
	passPriorityAroundTable(t, g)
	c := battlefieldCardCopy(t, g, rock)
	for _, tok := range c.Provenance.Mana {
		for _, r := range tok.Riders {
			if r.Applied {
				t.Errorf("the haste rider applied to an artifact spell: %+v", r)
			}
		}
	}
	if effectiveAbilitiesContain(t, g, rock, "haste") {
		t.Error("an artifact paid for with the Hall's mana has haste")
	}
}

// --- Scaled Nurturer, Path of Ancestry, Biophagus -------------------

func TestScaledNurturerGainsLifeOnlyForADragon(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		gain           int
	}{
		{"a Dragon", "Creature — Dragon", 2},
		{"an Elf", "Creature — Elf", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			nurturer := seedPermanentWithOracle(g, me.ID, "Scaled Nurturer", "Creature — Dragon Druid", scaledNurturerOracle)
			activateManaFor(t, g, me.ID, nurturer, 0, game.ManaAbilityParams{})
			life := me.Life

			castFromHandForTest(t, g, me, "Creature", tc.typeLine, "{G}", "", strict)
			passPriorityAroundTable(t, g)
			if got := me.Life - life; got != tc.gain {
				t.Errorf("gained %d life, want %d", got, tc.gain)
			}
		})
	}
}

// Path of Ancestry scries when its mana casts a creature that shares a
// type with the commander, and not otherwise.
func TestPathOfAncestryScriesOnlyForTheCommandersTypes(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		scry           bool
	}{
		{"an Elf under an Elf commander", "Creature — Elf Warrior", true},
		{"a Goblin under an Elf commander", "Creature — Goblin", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			g.WithWriteLock(func() {
				me.Command.PushTop(game.Card{
					InstanceID: uuid.New(), Name: "Lathril, Blade of the Elves",
					TypeLine: "Legendary Creature — Elf Noble", ManaCost: "{2}{B}{G}",
					Owner: me.ID, Controller: me.ID, IsCommander: true,
				})
				for i := 0; i < 3; i++ {
					me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID})
				}
			})
			path := seedPermanentWithOracle(g, me.ID, "Path of Ancestry", "Land", pathOfAncestryOracle)
			activateManaFor(t, g, me.ID, path, 0, game.ManaAbilityParams{Colors: []string{"G"}})

			castFromHandForTest(t, g, me, "Creature", tc.typeLine, "{G}", "", strict)
			passPriorityAroundTable(t, g)
			scried := false
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceScry {
					scried = true
				}
			}
			if scried != tc.scry {
				t.Errorf("scry prompt queued = %v, want %v", scried, tc.scry)
			}
		})
	}
}

// Biophagus: a creature cast with its mana enters with one extra +1/+1
// counter; a sorcery cast with it has nothing to enter, and a creature
// cast without it gets none.
func TestBiophagusCreatureEntersWithACounter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		bio      bool
		counters int
	}{
		{"paid with Biophagus's mana", true, 1},
		{"paid with a Forest", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			bio := seedPermanentWithOracle(g, me.ID, "Biophagus", "Creature — Human Tyranid Wizard", biophagusOracle)
			if tc.bio {
				activateManaFor(t, g, me.ID, bio, 0, game.ManaAbilityParams{Colors: []string{"G"}})
			} else {
				floatForTest(g, me, "G")
			}
			bear := castFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{G}", "", strict)
			passPriorityAroundTable(t, g)
			if got := countersOn(g, bear, "+1/+1"); got != tc.counters {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.counters)
			}
		})
	}
}

// --- undo and the snapshot ------------------------------------------

// A Cavern token in the pool, an uncounterable spell on the stack and a
// hasted creature on the battlefield all survive a snapshot round-trip
// through JSON — and the snapshot is a restore point, because every
// rider is data. The same three survive an undo (clone + RestoreFrom).
func TestManaSpendRidersSurviveUndoAndTheSnapshot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)

	// A hasted Bear, via the Hall.
	hall := seedPermanentWithOracle(g, me.ID, "Hall of the Bandit Lord", "Legendary Land", hallOfTheBanditLordOracl)
	activateManaFor(t, g, me.ID, hall, 0, game.ManaAbilityParams{})
	floatForTest(g, me, "G")
	bear := castFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", strict)
	passPriorityAroundTable(t, g)

	// A Cavern pick, answered (a pool token) and an Elf cast with a
	// second one (a stack item).
	cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
	tapCavernFor(t, g, me.ID, cavern, "G")
	elf := castFromHandForTest(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", strict)
	g.WithWriteLock(func() {
		me.ManaPool.AddMana(game.ManaToken{
			Color: "G", Source: cavern,
			Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictType("Creature"), game.ManaRestrictSubtype("Elf")},
			Riders:       []game.ManaSpendRider{SpentSpellCantBeCountered(ManaRestrictCast)},
		})
	})

	check := func(t *testing.T, g *game.Game, label string) {
		t.Helper()
		if !effectiveAbilitiesContain(t, g, bear, "haste") {
			t.Errorf("%s: the Hall's Bear lost its haste", label)
		}
		if !g.StackMeta[elf].SpellCantBeCounteredByMana() {
			t.Errorf("%s: the Elf on the stack lost its rider", label)
		}
		pool := g.PlayerByID(me.ID).ManaPool
		if len(pool) != 1 || len(pool[0].Riders) != 1 || pool[0].Riders[0].Kind != game.ManaRiderCantBeCountered {
			t.Errorf("%s: the floating Cavern token = %+v, want its rider", label, pool)
		}
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("spend riders cost the table its restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	check(t, restored, "after the snapshot")
	if counterByEffect(t, restored, elf) != true {
		t.Error("after the snapshot: the Cavern-cast Elf was countered")
	}

	before := g.Clone()
	g.WithWriteLock(func() { g.PlayerByIDForEffect(me.ID).ManaPool = nil })
	g.RestoreFrom(before)
	check(t, g, "after the undo")
}

// A colour pick queued by Cavern's coloured ability carries the rider on
// the choice, and a snapshot taken while it is waiting restores a pick
// that still mints uncounterable mana.
func TestCavernPickCarriesItsRiderAcrossTheSnapshot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
	activateManaFor(t, g, me.ID, cavern, 1, game.ManaAbilityParams{})

	restored, err := g.CaptureSnapshot().RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	you := restored.PlayerByID(me.ID)
	pick := riderLatestManaPick(restored, me.ID)
	if pick == nil || len(pick.ManaRiders) != 1 {
		t.Fatalf("the restored pick = %+v, want one rider on it", pick)
	}
	if err := restored.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatal(err)
	}
	elf := castFromHandForTest(t, restored, you, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", strict)
	if counterByEffect(t, restored, elf) != true {
		t.Error("an Elf paid with a restored Cavern pick was countered")
	}
}

// settleKeepingCopyTargets passes priority until the stack is empty,
// answering every CR 707.10c "choose new targets" prompt by keeping the
// copy aimed at `victim` and every trigger-ordering prompt as offered.
func settleKeepingCopyTargets(t *testing.T, g *game.Game, caster, victim uuid.UUID, copies int) {
	t.Helper()
	answered := 0
	for i := 0; i < 48; i++ {
		if p := triggerOrderPromptFor(g, caster); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, caster, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := latestPickTarget(g, caster); p != nil {
			if err := g.ResolvePickTarget(p.ID, caster, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if answered != copies {
		t.Fatalf("answered %d copy-target prompts, want %d", answered, copies)
	}
}
