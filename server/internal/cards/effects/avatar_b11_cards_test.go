package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b11_cards_test.go — Avatar: The Last Airbender batch 11 (#4459).

const (
	avB11CombustionManOracle = "533b107c-fd66-45ef-b74e-e6bd9b75ea8f"
	avB11CrescentOracle      = "2886f95e-c4e8-4770-b7de-40985f336cfb"
	avB11CunningOracle       = "96335212-ec57-4beb-bab1-11c8eaa62119"
	avB11DisciplesOracle     = "73ca3132-d849-4cf9-beba-75ba09789332"
	avB11AttacksOracle       = "399fba23-fc13-440b-ab67-192eede6ecab"
	avB11CadetsOracle        = "5ebfa3f5-f80a-4845-bba7-75c4ddb88682"
	avB11RaiderOracle        = "b37e9bc5-6a2c-4f08-b856-3f8d0d16defb"
	avB11SagesOracle         = "08d1940a-ddc6-4b6d-b6c9-9d2aa470b906"
	avB11LessonOracle        = "a9282f91-638e-414b-8b3d-9a99e30aec96"
	avB11StudentOracle       = "1ad3be51-f68b-46d6-b58e-400b50e384f1"
)

func TestAvatarB11CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB11CombustionManOracle: "Combustion Man",
		avB11CrescentOracle:      "Crescent Island Temple",
		avB11CunningOracle:       "Cunning Maneuver",
		avB11DisciplesOracle:     "Deserter's Disciple",
		avB11AttacksOracle:       "Fire Nation Attacks",
		avB11CadetsOracle:        "Fire Nation Cadets",
		avB11RaiderOracle:        "Fire Nation Raider",
		avB11SagesOracle:         "Fire Sages",
		avB11LessonOracle:        "Firebending Lesson",
		avB11StudentOracle:       "Firebending Student",
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

func avB11Refs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

// avB11Creature seeds a catalog creature with its printed stats.
func avB11Creature(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human", OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// avB11Hand puts a card in a player's hand.
func avB11Hand(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// --- Combustion Man -------------------------------------------------------------

func avB11CombustionManAttacks(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	man := avB11Creature(g, me.ID, "Combustion Man", avB11CombustionManOracle, 4, 6)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, man)
	b04WaitForPick(t, g, me.ID)
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	return g, me, opp, man, victim
}

func TestAvatarB11CombustionManAsksTheControllerAndDestroysOnDecline(t *testing.T) {
	g, me, opp, _, victim := avB11CombustionManAttacks(t)
	if c := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID); c != nil {
		t.Fatal("the attacking player was asked; only the target's controller is")
	}
	c := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
	if c == nil {
		t.Fatal("the target's controller was not asked")
	}
	life := opp.Life
	if err := g.ResolvePayUnless(c.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if onBattlefield(g, victim) {
		t.Error("the target survived a declined offer")
	}
	if opp.Life != life {
		t.Errorf("life %d, want %d: no damage on a decline", opp.Life, life)
	}
}

func TestAvatarB11CombustionManDealsDamageEqualToHisPowerInstead(t *testing.T) {
	g, _, opp, man, victim := avB11CombustionManAttacks(t)
	c := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
	if c == nil {
		t.Fatal("the target's controller was not asked")
	}
	if a := c.PayAction(); a == nil || a.Kind != game.PayActionDamage || a.Count != 4 {
		t.Fatalf("payment %+v, want 4 damage (his power)", a)
	}
	life := opp.Life
	if err := g.ResolvePayUnless(c.ID, opp.ID, true); err != nil {
		t.Fatalf("take the damage: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, victim) {
		t.Error("the target was destroyed although its controller took the damage")
	}
	if opp.Life != life-4 {
		t.Errorf("life %d, want %d", opp.Life, life-4)
	}
	if !onBattlefield(g, man) {
		t.Error("Combustion Man left the battlefield")
	}
}

func TestAvatarB11CombustionManDamageTracksHisCurrentPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	man := avB11Creature(g, me.ID, "Combustion Man", avB11CombustionManOracle, 4, 6)
	setCounters(g, man, map[string]int{game.CounterPlusOne: 2})
	land := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, man)
	b04WaitForPick(t, g, me.ID)
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
	if c == nil {
		t.Fatal("the land's controller was not asked")
	}
	if a := c.PayAction(); a == nil || a.Count != 6 {
		t.Fatalf("payment %+v, want 6 damage", a)
	}
}

// --- Crescent Island Temple ------------------------------------------------------

func TestAvatarB11CrescentIslandTempleMakesAMonkPerShrine(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushPermanent(g, me.ID, game.Card{Name: "Other Shrine", TypeLine: "Legendary Enchantment — Shrine"})
	castCatalogSpell(t, g, "Crescent Island Temple", "Legendary Enchantment — Shrine", avB11CrescentOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 2 {
		t.Fatalf("%d Monks, want 2 (two Shrines)", got)
	}
	monk := findBattlefieldByName(g, "Monk")
	c, _ := battlefieldCard(g, monk)
	if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "R" || !hasEffectiveKeyword(t, g, monk, "prowess") {
		t.Errorf("Monk %+v is not a 1/1 red Monk with prowess", c)
	}

	castCatalogSpell(t, g, "Third Shrine", "Legendary Enchantment — Shrine", "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 3 {
		t.Errorf("%d Monks after another Shrine entered, want 3", got)
	}
	castCatalogSpell(t, g, "Plain Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 3 {
		t.Errorf("%d Monks after a non-Shrine entered, want 3", got)
	}
}

// --- Cunning Maneuver ------------------------------------------------------------

func TestAvatarB11CunningManeuverPumpsAndMakesAClue(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Cunning Maneuver", "Instant", avB11CunningOracle, avB11Refs(bear))
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
	if got := countOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Deserter's Disciple ---------------------------------------------------------

func TestAvatarB11DesertersDiscipleTargetsOnlyAnotherSmallCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	disciple := avB11Creature(g, me.ID, "Deserter's Disciple", avB11DisciplesOracle, 2, 2)
	small := pushVanillaCreature(g, me.ID, "Small", 2, 2)
	big := pushVanillaCreature(g, me.ID, "Big", 3, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	advanceToMain(t, g)
	for _, bad := range []uuid.UUID{disciple, big, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, disciple, 0, game.ActivateAbilityParams{Strict: true, Targets: avB11Refs(bad)}); err == nil {
			t.Fatalf("activated targeting %s", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, disciple, 0, game.ActivateAbilityParams{Strict: true, Targets: avB11Refs(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, small)&game.CantBeBlocked == 0 {
		t.Error("the target can be blocked")
	}
}

// --- Fire Nation Attacks ---------------------------------------------------------

func TestAvatarB11FireNationAttacksMakesFirebendingSoldiersAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Fire Nation Attacks", "Instant", avB11AttacksOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Soldier", me.ID); got != 2 {
		t.Fatalf("%d Soldiers, want 2", got)
	}
	soldier := findBattlefieldByName(g, "Soldier")
	c, _ := battlefieldCard(g, soldier)
	if c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("Soldier %+v is not a 2/2 red Soldier", c)
	}

	id := seedGraveyardCard(t, g, "Fire Nation Attacks", "Instant", avB11AttacksOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Soldier", me.ID); got != 4 {
		t.Errorf("%d Soldiers after flashback, want 4", got)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}

}

// The Soldier token carries its own firebending 1 (ADR 0083).
func TestAvatarB11FirebendingSoldierTokenFirebends(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soldier := pushToken(g, me.ID, avatarB11FirebendingSoldierToken())
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, soldier)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 1})
}

// --- Fire Nation Cadets ----------------------------------------------------------

func TestAvatarB11FireNationCadetsFirebendOnlyWithALessonInYourGraveyard(t *testing.T) {
	for _, lesson := range []bool{false, true} {
		name := "no lesson"
		if lesson {
			name = "lesson"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			cadets := avB11Creature(g, me.ID, "Fire Nation Cadets", avB11CadetsOracle, 1, 2)
			pushToGraveyard(g, me, game.Card{Name: "Not a Lesson", TypeLine: "Instant"})
			if lesson {
				pushToGraveyard(g, me, game.Card{Name: "A Lesson", TypeLine: "Sorcery — Lesson"})
			}
			advanceToMain(t, g)
			declareAttack(t, g, opp.ID, cadets)
			passPriorityAroundTable(t, g)
			want := map[string]int{}
			if lesson {
				want["R"] = 2
			}
			mkWantPool(t, me, want)
		})
	}
}

func TestAvatarB11FireNationCadetsPumpForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cadets := avB11Creature(g, me.ID, "Fire Nation Cadets", avB11CadetsOracle, 1, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cadets, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	fillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, cadets, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, cadets); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
}

// --- Fire Nation Raider ----------------------------------------------------------

func TestAvatarB11FireNationRaiderMakesAClueOnlyAfterAnAttack(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		name := "no attack"
		if attacked {
			name = "attacked"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			advanceToMain(t, g)
			if attacked {
				declareAttack(t, g, opp.ID, bear)
				passPriorityAroundTable(t, g)
				mkAdvanceTo(t, g, game.StepPostcombatMain)
			}
			castCatalogSpell(t, g, "Fire Nation Raider", "Creature — Human Soldier", avB11RaiderOracle, nil)
			passPriorityAroundTable(t, g)
			passPriorityAroundTable(t, g)
			want := 0
			if attacked {
				want = 1
			}
			if got := countOnBattlefield(g, "Clue", me.ID); got != want {
				t.Errorf("%d Clues, want %d", got, want)
			}
		})
	}
}

func TestAvatarB11FireNationRaiderTriggerDeclaresARaidFootprint(t *testing.T) {
	rows := game.CatalogTriggers(avB11RaiderOracle)
	if len(rows) != 1 {
		t.Fatalf("%d rows, want 1", len(rows))
	}
	fp := rows[0].Footprint
	if len(fp) != 2 || fp[0].Kind != game.FootprintCondition || fp[0].Cond != game.FootprintCondAttacked || fp[1].Kind != game.FootprintCreateToken {
		t.Errorf("footprint %+v, want the raid condition then a token", fp)
	}
}

// --- Fire Sages ------------------------------------------------------------------

func TestAvatarB11FireSagesFirebendsAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sages := avB11Creature(g, me.ID, "Fire Sages", avB11SagesOracle, 2, 2)
	advanceToMain(t, g)
	fillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, sages, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{1}{R}{R} was paid with green mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "R", 3)
	if err := g.ActivateCatalogAbility(me.ID, sages, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, sages, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	declareAttack(t, g, opp.ID, sages)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 1})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

// --- Firebending Lesson ----------------------------------------------------------

func TestAvatarB11FirebendingLessonDealsTwoOrFiveKicked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	three := pushVanillaCreature(g, opp.ID, "Three", 1, 3)
	four := pushVanillaCreature(g, opp.ID, "Four", 1, 4)
	two := pushVanillaCreature(g, opp.ID, "Two", 1, 2)
	advanceToMain(t, g)

	lesson := game.Card{Name: "Firebending Lesson", TypeLine: "Instant — Lesson", ManaCost: "{R}", OracleID: avB11LessonOracle}
	id := avB11Hand(g, me, lesson)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: avB11Refs(me.ID)}); err == nil {
		t.Fatal("a player was a legal target")
	}
	fillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: avB11Refs(three)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, three) {
		t.Error("unkicked, the 1/3 died to what should be 2 damage")
	}
	id = avB11Hand(g, me, lesson)
	fillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: avB11Refs(two)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, two) {
		t.Error("the 1/2 survived 2 damage")
	}

	id = avB11Hand(g, me, lesson)
	fillPoolColored(me, "R", 4)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, Targets: avB11Refs(four)}); err == nil {
		t.Fatal("a kicked cast was paid with four mana")
	}
	fillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, Targets: avB11Refs(four)}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, four) {
		t.Error("kicked, the 1/4 survived 5 damage")
	}
}

// --- Firebending Student ---------------------------------------------------------

func TestAvatarB11FirebendingStudentAddsRedEqualToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	student := avB11Creature(g, me.ID, "Firebending Student", avB11StudentOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, student, "prowess") {
		t.Error("Firebending Student has no prowess")
	}
	setCounters(g, student, map[string]int{game.CounterPlusOne: 2})
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, student)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepEndCombat)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}
