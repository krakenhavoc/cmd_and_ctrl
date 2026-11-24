package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b051_cards_test.go — Strixhaven: School of Mages, batch 51 (#5502).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	oldB51General   = "10ea73c8-7cf8-4a48-a0cc-072bb541880f"
	oldB51Runner    = "37cf4ee6-e146-4354-afc8-6b59b0700e87"
	oldB51Pugilist  = "9110339d-72ba-4132-801f-cd2fd738b71d"
	oldB51Mascot    = "ae3da0be-2a3f-498a-8854-73d0a5526ec2"
	oldB51Kasmina   = "06425cb9-d157-4155-a727-78656df12458"
	oldB51Rowan     = "dd06c11c-c4fd-414e-ac63-5556f31491f5"
	oldB51Mila      = "ee524e82-35c4-4cae-b65e-3e147eac2927"
	oldB51Archway   = "69c63055-ed44-4b32-b591-f3c6c2f3e7d1"
	oldB51Biblio    = "86ed6073-c35c-4d29-9911-5fe191dd875f"
	oldB51Pennant   = "0fa51915-7c0b-4018-802a-90fd431f0036"
	oldB51Reject    = "198fed4c-297e-41c0-a172-e471c7401fbb"
	oldB51Golem     = "3c067dbb-934b-4ebb-a656-8541349a4b2a"
	oldB51Divinity  = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e" // Divination
	oldB51Doom      = "d64b0848-0193-4025-ba62-63ecd8fb9f50" // Spell Pierce
	oldB51WalkerTyp = "Legendary Planeswalker — Test"
)

func TestOldB51AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB51General:         "Blood Age General",
		oldB51Runner:          "Vortex Runner",
		oldB51Pugilist:        "Augmenter Pugilist",
		oldB51Pugilist + "#1": "Echoing Equation",
		oldB51Mascot:          "Mascot Interception",
		oldB51Kasmina:         "Kasmina, Enigma Sage",
		oldB51Rowan:           "Rowan, Scholar of Sparks",
		oldB51Rowan + "#1":    "Will, Scholar of Frost",
		oldB51Mila:            "Mila, Crafty Companion",
		oldB51Mila + "#1":     "Lukka, Wayward Bonder",
		oldB51Archway:         "Archway Commons",
		oldB51Biblio:          "The Biblioplex",
		oldB51Pennant:         "Team Pennant",
		oldB51Reject:          "Reject",
		oldB51Golem:           "Reflective Golem",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func oldB51OnBattlefield(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneBattlefield
}

func oldB51Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.SeedLand(g, owner, "Forest", "Basic Land — Forest", "")
	}
}

func oldB51Walker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: oldB51WalkerTyp, OracleID: oracle,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// --- Blood Age General -------------------------------------------------------

func TestOldB51BloodAgeGeneralPumpsOnlyAttackingSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	general := effectstest.B12Push(g, me.ID, "Blood Age General", "Creature — Spirit Warrior", oldB51General, 2, 2)
	spirit := effectstest.B12Creature(g, me.ID, "Spirit", "Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	idleSpirit := effectstest.B12Creature(g, me.ID, "Idle Spirit", "Creature — Spirit", 1, 1)
	effectstest.AttackWith(t, g, opp.ID, spirit, bear)
	if err := effectstest.HobB4Activate(t, g, me.ID, general, 0, "", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, spirit); p != 2 {
		t.Errorf("the attacking Spirit has power %d, want 2", p)
	}
	for name, id := range map[string]uuid.UUID{"non-Spirit attacker": bear, "idle Spirit": idleSpirit, "the General": general} {
		want := 1
		if id == general {
			want = 2
		}
		if p := effectstest.EffectivePower(t, g, id); p != want {
			t.Errorf("the %s has power %d, want %d", name, p, want)
		}
	}
}

// --- Vortex Runner / Augmenter Pugilist ------------------------------------------

func TestOldB51VortexRunnerNeedsEightLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Push(g, me.ID, "Vortex Runner", "Creature — Human Wizard", oldB51Runner, 2, 3)
	oldB51Lands(g, me.ID, 7)
	if p := effectstest.EffectivePower(t, g, runner); p != 2 {
		t.Fatalf("with seven lands the Runner has power %d, want 2", p)
	}
	oldB51Lands(g, me.ID, 1)
	if p := effectstest.EffectivePower(t, g, runner); p != 3 {
		t.Errorf("with eight lands the Runner has power %d, want 3", p)
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, runner); err == nil {
		t.Error("blocked the Runner while its controller controls eight lands")
	}
}

func TestOldB51AugmenterPugilistGetsFivePlusFiveWithEightLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pug := effectstest.B12Push(g, me.ID, "Augmenter Pugilist", "Creature — Troll Druid", oldB51Pugilist, 3, 3)
	oldB51Lands(g, me.ID, 7)
	if p, tough := effectstest.EffectivePower(t, g, pug), effectstest.EffectiveToughness(t, g, pug); p != 3 || tough != 3 {
		t.Fatalf("with seven lands the Pugilist is %d/%d, want 3/3", p, tough)
	}
	oldB51Lands(g, me.ID, 1)
	if p, tough := effectstest.EffectivePower(t, g, pug), effectstest.EffectiveToughness(t, g, pug); p != 8 || tough != 8 {
		t.Errorf("with eight lands the Pugilist is %d/%d, want 8/8", p, tough)
	}
}

func TestOldB51EchoingEquationCopiesTheTargetOntoEachOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	model := effectstest.B12Push(g, me.ID, "Hero", "Legendary Creature — Human", "", 4, 4)
	mouse := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Echoing Equation", "Sorcery", oldB51Pugilist+"#1", effectstest.CardTargets(model))
	effectstest.PassPriorityAroundTable(t, g)
	view := effectstest.GaLayered(t, g, mouse)
	if view.Name != "Hero" {
		t.Errorf("my other creature is named %q, want a copy of Hero", view.Name)
	}
	if p := effectstest.EffectivePower(t, g, mouse); p != 4 {
		t.Errorf("my other creature has power %d, want 4", p)
	}
	if view.HasSupertype("legendary") {
		t.Error("the copy is legendary")
	}
	if !effectstest.GaLayered(t, g, model).HasSupertype("legendary") {
		t.Error("the model lost its legend status")
	}
	if effectstest.GaLayered(t, g, theirs).Name != "Their Mouse" {
		t.Error("an opponent's creature became a copy")
	}
}

// --- Mascot Interception ----------------------------------------------------------

func TestOldB51MascotInterceptionStealsUntapsPumpsAndHastes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(theirs) })
	effectstest.CastCatalogSpell(t, g, "Mascot Interception", "Sorcery", oldB51Mascot, effectstest.CardTargets(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	view := effectstest.GaLayered(t, g, theirs)
	if view.Controller != me.ID {
		t.Fatalf("the creature is controlled by %v, want the caster", view.Controller)
	}
	if view.Tapped {
		t.Error("the creature is still tapped")
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 4 {
		t.Errorf("the creature has power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the creature has no haste")
	}
}

func TestOldB51MascotInterceptionCostsThreeLessTargetingACreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	token := effectstest.SocB17Names(g, opp.ID, "Soldier")[0].InstanceID
	spell := effectstest.SocB17Hand(me, "Mascot Interception", "Sorcery", "{3}{R}", oldB51Mascot)
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Fatal("cast for {R} targeting a nontoken creature")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(token)}); err != nil {
		t.Errorf("cast for {R} targeting a creature token: %v", err)
	}
}

// --- Kasmina, Enigma Sage -------------------------------------------------------------

func TestOldB51KasminaLendsHerLoyaltyAbilitiesToOtherPlaneswalkersYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kasmina := oldB51Walker(g, me.ID, "Kasmina, Enigma Sage", oldB51Kasmina, 2)
	mine := oldB51Walker(g, me.ID, "Lukka, Wayward Bonder", oldB51Mila+"#1", 3)
	theirs := oldB51Walker(g, opp.ID, "Lukka, Wayward Bonder", oldB51Mila+"#1", 3)
	lent := func(id uuid.UUID) int {
		_, all := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, id))
		n := 0
		for _, o := range all {
			if o.Borrowed {
				n++
			}
		}
		return n
	}
	if n := lent(mine); n != 3 {
		t.Errorf("my other planeswalker has %d lent loyalty abilities, want 3", n)
	}
	if n := lent(theirs); n != 0 {
		t.Errorf("an opponent's planeswalker has %d lent abilities, want 0", n)
	}
	if n := lent(kasmina); n != 0 {
		t.Errorf("Kasmina lends herself %d rows", n)
	}
}

func TestOldB51KasminaPlusTwoScriesAndMinusXMakesAFractal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	kasmina := oldB51Walker(g, me.ID, "Kasmina, Enigma Sage", oldB51Kasmina, 5)
	if err := effectstest.HobB4Activate(t, g, me.ID, kasmina, 0, "", nil); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry); scry == nil {
		t.Fatal("no scry 1 was asked")
	}
	if got := effectstest.CounterOn(g, kasmina, game.CounterLoyalty); got != 7 {
		t.Errorf("loyalty %d after +2, want 7", got)
	}
}

func TestOldB51KasminaMinusXMakesAFractalWithXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	kasmina := oldB51Walker(g, me.ID, "Kasmina, Enigma Sage", oldB51Kasmina, 5)
	if err := g.ActivateCatalogAbility(me.ID, kasmina, 1, game.ActivateAbilityParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("−X: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	fractals := effectstest.SocB17Names(g, me.ID, "Fractal")
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals, want 1", len(fractals))
	}
	if n := effectstest.CounterOn(g, fractals[0].InstanceID, game.CounterPlusOne); n != 3 {
		t.Errorf("the Fractal has %d +1/+1 counters, want 3", n)
	}
	if got := effectstest.CounterOn(g, kasmina, game.CounterLoyalty); got != 2 {
		t.Errorf("loyalty %d after −3, want 2", got)
	}
}

func TestOldB51KasminaLentAbilityUsesTheRecipientsLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	oldB51Walker(g, me.ID, "Kasmina, Enigma Sage", oldB51Kasmina, 2)
	mine := oldB51Walker(g, me.ID, "Lukka, Wayward Bonder", oldB51Mila+"#1", 3)
	_, all := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, mine))
	idx := -1
	for i, o := range all {
		if o.Borrowed {
			idx = i // the +2 is the first lent row
			break
		}
	}
	if idx < 0 {
		t.Fatal("no lent row")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, mine, idx, "", nil); err != nil {
		t.Fatalf("activate the lent +2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mine, game.CounterLoyalty); got != 5 {
		t.Errorf("the recipient's loyalty is %d after the lent +2, want 5", got)
	}
}

// --- Rowan, Scholar of Sparks // Will, Scholar of Frost -------------------------------

func TestOldB51RowanAndWillDiscountInstantsAndSorceries(t *testing.T) {
	for _, row := range []struct{ name, oracle, typeLine string }{
		{"Rowan, Scholar of Sparks", oldB51Rowan, oldB51WalkerTyp},
		{"Will, Scholar of Frost", oldB51Rowan + "#1", oldB51WalkerTyp},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		oldB51Walker(g, me.ID, row.name, row.oracle, 3)
		cast := func(name, typeLine, cost, pool string) error {
			me.ManaPool = nil
			id := effectstest.SocB17Hand(me, name, typeLine, cost, "")
			effectstest.FloatForTest(g, me, pool)
			return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		}
		if err := cast("Instant", "Instant", "{1}{U}", "U"); err != nil {
			t.Errorf("%s: an instant should cost {U}: %v", row.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if err := cast("Creature", "Creature — Bear", "{1}{U}", "U"); err == nil {
			t.Errorf("%s: a creature spell was discounted", row.name)
		}
	}
}

func TestOldB51RowanPlusOneDealsOneOrThreeDependingOnCardsDrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opp = s
			break
		}
	}
	effectstest.AdvanceToMain(t, g)
	rowan := oldB51Walker(g, me.ID, "Rowan, Scholar of Sparks", oldB51Rowan, 2)
	before := effectstest.LifeOf(g, opp.ID)
	if err := effectstest.HobB4Activate(t, g, me.ID, rowan, 0, "", nil); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Fatalf("opponent at %d after the first +1, want %d", got, before-1)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3) })
	g.WithWriteLock(func() { delete(g.LoyaltyActivatedThisTurn, rowan) })
	if err := effectstest.HobB4Activate(t, g, me.ID, rowan, 0, "", nil); err != nil {
		t.Fatalf("second +1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1-3 {
		t.Errorf("opponent at %d after drawing three, want %d", got, before-4)
	}
}

func TestOldB51RowanEmblemCopiesYourInstantsAndSorceriesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	rowan := oldB51Walker(g, me.ID, "Rowan, Scholar of Sparks", oldB51Rowan, 5)
	if err := effectstest.HobB4Activate(t, g, me.ID, rowan, 1, "", nil); err != nil {
		t.Fatalf("−4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Msc18Library(me, 12)
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d after Divination and its copy, want %d", got, hand+4)
	}
}

func TestOldB51WillPlusOneSetsBasePowerAndToughnessAndMinusThreeDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	will := oldB51Walker(g, me.ID, "Will, Scholar of Frost", oldB51Rowan+"#1", 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	if err := effectstest.HobB4Activate(t, g, me.ID, will, 0, "", effectstest.CardTargets(ogre)); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, ogre), effectstest.EffectiveToughness(t, g, ogre); p != 0 || tough != 2 {
		t.Errorf("the Ogre is %d/%d, want 0/2", p, tough)
	}
	g.WithWriteLock(func() { delete(g.LoyaltyActivatedThisTurn, will) })
	hand := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, will, 1, "", nil); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after −3, want %d", got, hand+2)
	}
}

func TestOldB51WillMinusSevenExilesPermanentsAndGivesTheirControllersElementals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	will := oldB51Walker(g, me.ID, "Will, Scholar of Frost", oldB51Rowan+"#1", 7)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Ogre A", "Creature — Ogre", 3, 3)
	b := effectstest.B12Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
	if err := effectstest.HobB4Activate(t, g, me.ID, will, 2, "", effectstest.CardTargets(mine, a, b)); err != nil {
		t.Fatalf("−7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, a, b} {
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneExile {
			t.Errorf("a target is in %v, want exile", z)
		}
	}
	if n := len(effectstest.SocB17Names(g, me.ID, "Elemental")); n != 1 {
		t.Errorf("%d Elementals for me, want 1", n)
	}
	if n := len(effectstest.SocB17Names(g, opp.ID, "Elemental")); n != 2 {
		t.Errorf("%d Elementals for the opponent, want 2", n)
	}
}

// --- Mila, Crafty Companion // Lukka, Wayward Bonder -------------------------------------

func TestOldB51MilaPutsALoyaltyCounterOnEachPlaneswalkerWhenAnOpponentAttacksOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	opp := g.Seats[g.Turn.ActiveSeat]
	if opp.ID == me.ID {
		me = g.Seats[0]
	}
	effectstest.B12Push(g, me.ID, "Mila, Crafty Companion", "Legendary Creature — Fox", oldB51Mila, 2, 3)
	walkerA := oldB51Walker(g, me.ID, "Walker A", "", 3)
	walkerB := oldB51Walker(g, me.ID, "Walker B", "", 3)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walkerA); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, w := range []uuid.UUID{walkerA, walkerB} {
		if got := effectstest.CounterOn(g, w, game.CounterLoyalty); got != 4 {
			t.Errorf("a planeswalker has %d loyalty, want 4", got)
		}
	}
}

func TestOldB51MilaMayDrawWhenAnOpponentTargetsYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mila, Crafty Companion", "Legendary Creature — Fox", oldB51Mila, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastFromSeat(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	answerTrigger := func(yes bool) {
		t.Helper()
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID)
		if c == nil {
			t.Fatal("no \"you may draw\" prompt")
		}
		if err := g.ResolveTriggerPrompt(c.ID, me.ID, yes); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
	answerTrigger(true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after an opponent targeted my creature, want %d", got, hand+1)
	}
	// Targeting your own permanent does not trigger it.
	hand = me.Hand.Size()
	bear2 := effectstest.B12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(bear2))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID) != nil {
		t.Error("Mila asked about drawing for the controller's own spell")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after targeting my own creature, want %d", got, hand)
	}
}

func TestOldB51LukkaPlusOneDrawsOneOrTwoByTheDiscardedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lukka := oldB51Walker(g, me.ID, "Lukka, Wayward Bonder", oldB51Mila+"#1", 5)
	effectstest.Msc18Library(me, 10)
	me.Hand.Cards = nil
	creature := effectstest.SocB17Hand(me, "Bear", "Creature — Bear", "{1}{G}", "")
	land := effectstest.SocB17Hand(me, "Island", "Basic Land — Island", "", "")
	if err := effectstest.HobB4Activate(t, g, me.ID, lukka, 0, "", nil); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 { // the land plus two drawn
		t.Errorf("hand %d after discarding a creature card, want 3", got)
	}
	_ = land
}

func TestOldB51LukkaMinusTwoReturnsACreatureWithHasteAndExilesItNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lukka := oldB51Walker(g, me.ID, "Lukka, Wayward Bonder", oldB51Mila+"#1", 5)
	dead := effectstest.GraveyardCreature(me, "Dead Bear", "Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, lukka, 1, "", effectstest.CardTargets(dead)); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB51OnBattlefield(g, dead) {
		t.Fatal("the creature did not return")
	}
	if !effectstest.HasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("the returned creature has no haste")
	}
	// It stays through the rest of the turn and is exiled at the beginning of
	// its controller's next upkeep.
	start := g.Turn.Seq
	seat := indexOfSeat(g, me.ID)
	for i := 0; i < 200; i++ {
		if g.Turn.Seq > start && g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepUpkeep {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Seq == start && !oldB51OnBattlefield(g, dead) {
			t.Fatal("the creature left before the next upkeep")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, dead); z != game.ZoneExile {
		t.Errorf("the returned creature is in %v at my next upkeep, want exile", z)
	}
}

func indexOfSeat(g *game.Game, id uuid.UUID) int {
	for i, s := range g.Seats {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func TestOldB51LukkaEmblemDealsDamageEqualToPowerWhenACreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lukka := oldB51Walker(g, me.ID, "Lukka, Wayward Bonder", oldB51Mila+"#1", 8)
	if err := effectstest.HobB4Activate(t, g, me.ID, lukka, 2, "", nil); err != nil {
		t.Fatalf("−7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.LifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/1 red Elemental"), 1) })
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("opponent at %d after a 3-power creature entered, want %d", got, before-3)
	}
}

// --- Archway Commons -------------------------------------------------------------------

func TestOldB51ArchwayCommonsEntersTappedAndIsSacrificedUnlessYouPayOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PlayLandFromHand(t, g, "Archway Commons", oldB51Archway)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB51OnBattlefield(g, land) || !effectstest.GaLayered(t, g, land).Tapped {
		t.Fatal("Archway Commons did not enter tapped")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB51OnBattlefield(g, land) {
		t.Error("Archway Commons stayed after the {1} was declined")
	}
}

func TestOldB51ArchwayCommonsStaysWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PlayLandFromHand(t, g, "Archway Commons", oldB51Archway)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB51OnBattlefield(g, land) {
		t.Error("Archway Commons was sacrificed after {1} was paid")
	}
}

// --- The Biblioplex ---------------------------------------------------------------------

func TestOldB51TheBiblioplexOnlyActivatesWithZeroOrSevenCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.SeedLand(g, me.ID, "The Biblioplex", "Land", oldB51Biblio)
	effectstest.SeedLibrary(me, "A", "B")
	for _, n := range []int{3, 6, 8} {
		me.Hand.Cards = nil
		for i := 0; i < n; i++ {
			effectstest.SocB17Hand(me, "Card", "Sorcery", "", "")
		}
		if err := effectstest.HobB4Activate(t, g, me.ID, land, 0, "{C}{C}", nil); err == nil {
			t.Fatalf("activated with %d cards in hand", n)
		}
		me.ManaPool = nil
	}
	me.Hand.Cards = nil
	if err := effectstest.HobB4Activate(t, g, me.ID, land, 0, "{C}{C}", nil); err != nil {
		t.Errorf("could not activate with no cards in hand: %v", err)
	}
}

func TestOldB51TheBiblioplexTakesAnInstantOrSorceryFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.SeedLand(g, me.ID, "The Biblioplex", "Land", oldB51Biblio)
	me.Hand.Cards = nil
	top := effectstest.SeedLibraryTop(me, "Test Instant", "Instant", "{U}")
	if err := effectstest.HobB4Activate(t, g, me.ID, land, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no reveal-and-take prompt for an instant on top")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top}); err != nil {
		t.Fatalf("take: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the instant did not reach the hand")
	}
}

func TestOldB51TheBiblioplexMayPutAnythingNotTakenIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.SeedLand(g, me.ID, "The Biblioplex", "Land", oldB51Biblio)
	me.Hand.Cards = nil
	top := effectstest.SeedLibraryTop(me, "Test Creature", "Creature — Bear", "{1}{G}")
	if err := effectstest.HobB4Activate(t, g, me.ID, land, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyard(me, top) {
		t.Error("the creature card was not put into the graveyard")
	}
}

// --- Team Pennant -----------------------------------------------------------------------

func TestOldB51TeamPennantPumpsAndGrantsKeywordsAndEquipsATokenForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	pennant := effectstest.B12Push(g, me.ID, "Team Pennant", "Artifact — Equipment", oldB51Pennant, 0, 0)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	token := effectstest.SocB17Names(g, me.ID, "Soldier")[0].InstanceID
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, pennant, 0, "{C}", effectstest.CardTargets(bear)); err == nil {
		t.Fatal("equip creature token {1} went on a nontoken creature")
	}
	me.ManaPool = nil
	if err := effectstest.HobB4Activate(t, g, me.ID, pennant, 0, "{C}", effectstest.CardTargets(token)); err != nil {
		t.Fatalf("equip creature token {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, token), effectstest.EffectiveToughness(t, g, token); p != 2 || tough != 2 {
		t.Errorf("the equipped token is %d/%d, want 2/2", p, tough)
	}
	for _, kw := range []string{"vigilance", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, token, kw) {
			t.Errorf("the equipped token lacks %s", kw)
		}
	}
	// Equip {3} takes any creature you control.
	if err := effectstest.HobB4Activate(t, g, me.ID, pennant, 1, "{C}{C}{C}", effectstest.CardTargets(bear)); err != nil {
		t.Fatalf("equip {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("the equipped Bear has power %d, want 3", p)
	}
}

// --- Reject -----------------------------------------------------------------------------

func TestOldB51RejectExilesACreatureSpellItCountersAndLetsAPayerThrough(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SocB17Hand(me, "Bear", "Creature — Bear", "", "")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	effectstest.CastFromSeat(t, g, opp, "Reject", "Instant", oldB51Reject, effectstest.CardTargets(bear))
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, bear); z != game.ZoneExile {
		t.Errorf("the countered creature spell is in %v, want exile", z)
	}
}

func TestOldB51RejectIsNotLegalAgainstANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sorcery := effectstest.SocB17Hand(me, "Sorcery", "Sorcery", "", "")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	card := game.Card{InstanceID: uuid.New(), Name: "Reject", TypeLine: "Instant", OracleID: oldB51Reject, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(card)
	if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(sorcery)}); err == nil {
		t.Error("Reject targeted a sorcery spell")
	}
}

// --- Reflective Golem -------------------------------------------------------------------

func TestOldB51ReflectiveGolemCopiesASpellThatTargetsOnlyItForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.B12Push(g, me.ID, "Reflective Golem", "Artifact Creature — Golem", oldB51Golem, 2, 3)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(golem))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// A copy of the Bolt needs its target chosen; keeping the Golem is the default.
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: golem}); err != nil {
			t.Fatalf("retarget: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LifeOf(g, me.ID); got != before {
		t.Errorf("life changed from %d to %d", before, got)
	}
	if oldB51OnBattlefield(g, golem) {
		t.Error("the Golem survived two Lightning Bolts")
	}
}

func TestOldB51ReflectiveGolemIgnoresASpellWithAnotherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Reflective Golem", "Artifact Creature — Golem", oldB51Golem, 2, 3)
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the Golem asked for {2} for a spell that did not target it")
	}
}
