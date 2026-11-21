package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b25_test.go — S82 reprint pass, batch 25 (#5908). Each test plays
// the real catalog entry.

const (
	rprb25BonehoardOracle       = "83c69c05-173d-4c1b-9541-1dde474fef5f"
	rprb25CrushContrabandOracle = "ecbc234d-e1d3-4bcc-81d4-fffc2c7ea16a"
	rprb25MyrsmithOracle        = "7fc3798a-3907-4f50-98a1-25b8e1977197"
	rprb25ProtectorOracle       = "a9581345-87dc-4fbd-acc1-bb629f2cf40f"
	rprb25RaffOracle            = "13044ca1-76a3-43b2-a8e4-3bbc692c30c5"
	rprb25RazerunnersOracle     = "7b3e05e1-31b4-46af-b3d0-f2467f8551e9"
	rprb25PlowOracle            = "4cdf524a-1a99-4143-823b-d3c5c867c580"
	rprb25ColossusOracle        = "f89f7c7d-6eb3-4335-b9df-0ccaa9c50558"
	rprb25MasteryOracle         = "4234f07e-583f-40db-8b47-7e16c7316e19"
	rprb25DanceOracle           = "fdfac093-a5c0-47b4-9052-e0af517d83cf"
	rprb25JaceOracle            = "2274c7ae-5a40-4fd4-a4ac-6f56b23034e4"
)

func TestRprB25CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprb25BonehoardOracle:       "Bonehoard",
		rprb25CrushContrabandOracle: "Crush Contraband",
		rprb25MyrsmithOracle:        "Myrsmith",
		rprb25ProtectorOracle:       "Primeval Protector",
		rprb25RaffOracle:            "Raff Capashen, Ship's Mage",
		rprb25RazerunnersOracle:     "Goblin Razerunners",
		rprb25PlowOracle:            "Colossal Plow",
		rprb25ColossusOracle:        "Peacewalker Colossus",
		rprb25MasteryOracle:         "Elemental Mastery",
		rprb25DanceOracle:           "Dance of the Manse",
		rprb25JaceOracle:            "Jace, Architect of Thought",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, %d caveats", name, def.Name, def.Completeness, len(def.Caveats))
		}
	}
}

func TestRprB25BonehoardCountsCreatureCardsInEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[0], g.Seats[1]
	effectstest.PutCreatureCardsInGraveyard(me, 1)
	effectstest.PutCreatureCardsInGraveyard(other, 2)
	other.Graveyard.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
		Owner: other.ID, Controller: other.ID,
	})
	hoard, germ := effectstest.CastLivingWeapon(t, g, "Bonehoard", rprb25BonehoardOracle)
	if host := effectstest.AttachmentHostOf(t, g, hoard); host.ID != germ {
		t.Fatalf("Bonehoard is attached to %+v, want the Germ", host)
	}
	if p, tg := effectstest.EffectivePower(t, g, germ), effectstest.EffectiveToughness(t, g, germ); p != 3 || tg != 3 {
		t.Errorf("the Germ is %d/%d, want 3/3 (three creature cards, the land does not count)", p, tg)
	}
	victim := effectstest.B12Creature(g, other.ID, "Victim", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	if p := effectstest.EffectivePower(t, g, germ); p != 4 {
		t.Errorf("the Germ is %d/_ after another creature card died, want 4", p)
	}
	spec, _ := effects.Lookup(rprb25BonehoardOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Cost.Mana != "{2}" {
		t.Errorf("the equip row is %+v, want one Equip {2}", spec.Activated)
	}
}

// --- Crush Contraband --------------------------------------------------------

func rprb25Permanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
}

func TestRprB25CrushContrabandExilesBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := rprb25Permanent(g, opp.ID, "Trinket", "Artifact")
	ench := rprb25Permanent(g, opp.ID, "Aura Thing", "Enchantment")
	effectstest.D1658Cast(t, g, "Crush Contraband", "Instant", rprb25CrushContrabandOracle, game.CastSpellParams{
		Strict: true, Modes: []int{0, 1}, Targets: effectstest.CardRefs(art, ench),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, art) != game.ZoneExile || effectstest.B12ZoneOf(g, ench) != game.ZoneExile {
		t.Errorf("artifact in %v, enchantment in %v, want both exiled", effectstest.B12ZoneOf(g, art), effectstest.B12ZoneOf(g, ench))
	}
}

func TestRprB25CrushContrabandOneModeLeavesTheOtherAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := rprb25Permanent(g, opp.ID, "Trinket", "Artifact")
	ench := rprb25Permanent(g, opp.ID, "Aura Thing", "Enchantment")
	effectstest.D1658Cast(t, g, "Crush Contraband", "Instant", rprb25CrushContrabandOracle, game.CastSpellParams{
		Strict: true, Modes: []int{1}, Targets: effectstest.CardRefs(ench),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, ench) != game.ZoneExile {
		t.Error("the enchantment was not exiled")
	}
	if !g.Battlefield.Contains(art) {
		t.Error("the artifact was exiled by the enchantment mode")
	}
}

func TestRprB25CrushContrabandRefusesTheWrongTargetsAndNoModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ench := rprb25Permanent(g, opp.ID, "Aura Thing", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.D1658CastErr(t, g, "Crush Contraband", "Instant", rprb25CrushContrabandOracle, game.CastSpellParams{
		Strict: true, Modes: []int{0}, Targets: effectstest.CardRefs(ench),
	}); err == nil {
		t.Error("an enchantment was accepted as the artifact target")
	}
	if err := effectstest.D1658CastErr(t, g, "Crush Contraband", "Instant", rprb25CrushContrabandOracle, game.CastSpellParams{
		Strict: true, Modes: []int{1}, Targets: effectstest.CardRefs(bear),
	}); err == nil {
		t.Error("a creature was accepted as the enchantment target")
	}
}

// --- Myrsmith ----------------------------------------------------------------

func rprb25CastArtifactWithMyrsmith(t *testing.T, typeLine string) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Myrsmith", "Creature — Human Artificer", rprb25MyrsmithOracle, 2, 1)
	effectstest.CastCatalogSpell(t, g, "Spell", typeLine, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestRprB25MyrsmithPaysOneForAMyr(t *testing.T) {
	g, me := rprb25CastArtifactWithMyrsmith(t, "Artifact")
	effectstest.B06AddMana(me, "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("no payment prompt after casting an artifact spell")
	}
	if ask.PayCost != "{1}" {
		t.Errorf("optional cost %s", ask.PayCost)
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Myr")
	if len(ids) != 1 {
		t.Fatalf("%d Myr tokens, want 1", len(ids))
	}
	tok, _ := g.LookupCardForEffect(ids[0])
	if tok.Power != 1 || tok.Toughness != 1 || !tok.IsCreature() || !tok.IsArtifact() {
		t.Errorf("token %d/%d %q", tok.Power, tok.Toughness, tok.TypeLine)
	}
}

func TestRprB25MyrsmithDecliningMakesNothing(t *testing.T) {
	g, me := rprb25CastArtifactWithMyrsmith(t, "Artifact")
	effectstest.B06AddMana(me, "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if got := len(effectstest.BattlefieldIDsNamed(g, "Myr")); got != 0 {
		t.Errorf("%d Myr after declining", got)
	}
}

func TestRprB25MyrsmithIgnoresNonartifactSpells(t *testing.T) {
	g, _ := rprb25CastArtifactWithMyrsmith(t, "Sorcery")
	if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Error("a sorcery asked for {1}")
	}
}

// --- Primeval Protector ------------------------------------------------------

func TestRprB25PrimevalProtectorCostsLessPerOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, rprb25ProtectorOracle, "Creature — Avatar", "{10}{G}"); got != 9 {
		t.Errorf("Primeval Protector with two opposing creatures has mana value %d, want 9", got)
	}
}

func TestRprB25PrimevalProtectorPutsACounterOnEachOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	prot := effectstest.CastCatalogSpell(t, g, "Primeval Protector", "Creature — Avatar", rprb25ProtectorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, mine); got != 1 {
		t.Errorf("my other creature has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.PlusOneCounters(g, theirs); got != 0 {
		t.Errorf("an opponent's creature has %d +1/+1 counters, want 0", got)
	}
	if got := effectstest.PlusOneCounters(g, prot); got != 0 {
		t.Errorf("the Protector put %d counters on itself, want 0", got)
	}
}

// --- Raff Capashen, Ship's Mage ----------------------------------------------

func TestRprB25RaffGivesHistoricSpellsFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raff := effectstest.B12Push(g, me.ID, "Raff Capashen, Ship's Mage", "Legendary Creature — Human Wizard", rprb25RaffOracle, 3, 3)
	for _, kw := range []string{"flash", "flying"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, raff, kw) {
			t.Errorf("Raff has no %s", kw)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	cast := func(name, typeLine string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
			t.Fatal(err)
		}
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if err == nil {
			effectstest.PassPriorityAroundTable(t, g)
		}
		return err
	}
	for _, tc := range []struct{ Name, TypeLine string }{
		{"Rock", "Artifact"},
		{"Hero", "Legendary Creature — Human"},
		{"History", "Enchantment — Saga"},
	} {
		if err := cast(tc.Name, tc.TypeLine); err != nil {
			t.Errorf("historic %s in the upkeep: %v", tc.TypeLine, err)
		}
	}
	for _, tc := range []struct{ Name, TypeLine string }{
		{"Bear", "Creature — Bear"},
		{"Bolt", "Sorcery"},
		{"Ring", "Enchantment"},
	} {
		if err := cast(tc.Name, tc.TypeLine); err == nil {
			t.Errorf("non-historic %s was cast in the upkeep", tc.TypeLine)
		}
	}
}

// --- Goblin Razerunners ------------------------------------------------------

func TestRprB25RazerunnersSacrificesALandForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gob := effectstest.B12Push(g, me.ID, "Goblin Razerunners", "Creature — Goblin Warrior", rprb25RazerunnersOracle, 3, 4)
	land := effectstest.MrdPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("activated without paying {1}{R}")
	}
	effectstest.B06AddMana(me, "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without sacrificing a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if got := effectstest.CountersOn(g, gob, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func rprb25RazerunnersEndStep(t *testing.T, accept bool) (*game.Game, *game.Player, int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gob := effectstest.B12Push(g, me.ID, "Goblin Razerunners", "Creature — Goblin Warrior", rprb25RazerunnersOracle, 3, 4)
	effectstest.AddCounters(t, g, gob, game.CounterPlusOne, 2)
	before := opp.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestConfirmFor(g, me.ID)
	if ask == nil {
		t.Fatal("no \"you may\" prompt")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, accept); err != nil {
		t.Fatal(err)
	}
	return g, opp, before
}

func TestRprB25RazerunnersDealsItsCountersAtTheEndStep(t *testing.T) {
	_, opp, before := rprb25RazerunnersEndStep(t, true)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d life, want 2 (two +1/+1 counters)", got)
	}
}

func TestRprB25RazerunnersDecliningDealsNothing(t *testing.T) {
	_, opp, before := rprb25RazerunnersEndStep(t, false)
	if opp.Life != before {
		t.Errorf("the opponent lost %d life after declining", before-opp.Life)
	}
}

// --- Colossal Plow -----------------------------------------------------------

func TestRprB25ColossalPlowAddsKeptWhiteAndGainsThree(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	plow := s.PushVehicle("Colossal Plow", rprb25PlowOracle, 6, 3)
	crewer := s.PushCreature("Big Crewer", "Creature — Giant", 6, 6)
	s.CrewAndResolve(t, plow, 0, crewer)
	life := s.Me.Life
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, plow)
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := len(s.Me.ManaPool); got != 3 {
		t.Fatalf("mana pool %+v after the attack, want {W}{W}{W}", s.Me.ManaPool)
	}
	for _, m := range s.Me.ManaPool {
		if m.Color != "W" {
			t.Errorf("mana %+v, want white", m)
		}
	}
	if s.Me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, s.Me.Life)
	}
	effectstest.AdvanceTo(t, s.G, game.StepCombatDamage)
	if got := len(s.Me.ManaPool); got != 3 {
		t.Errorf("mana pool = %d after the combat steps, want the kept {W}{W}{W}", got)
	}
}

func TestRprB25ColossalPlowNeedsSixPowerToCrew(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	plow := s.PushVehicle("Colossal Plow", rprb25PlowOracle, 6, 3)
	small := s.PushCreature("Small Crewer", "Creature — Human", 5, 5)
	effectstest.ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, plow, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Error("crewed Colossal Plow with total power 5")
	}
}

// --- Peacewalker Colossus ----------------------------------------------------

func TestRprB25PeacewalkerColossusAnimatesAnotherVehicle(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	colossus := s.PushVehicle("Peacewalker Colossus", rprb25ColossusOracle, 6, 6)
	other := s.PushVehicle("Other Cart", "", 3, 3)
	effectstest.ToMain(t, s.G)
	effectstest.B06AddMana(s.Me, "W", "C")
	if err := s.G.ActivateCatalogAbility(s.Me.ID, colossus, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(colossus)}); err == nil {
		t.Fatal("the Colossus targeted itself")
	}
	if err := s.G.ActivateCatalogAbility(s.Me.ID, colossus, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(other)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, s.G)
	if c, ok := effectstest.BattlefieldCardByID(s.G, other); !ok || !c.IsCreature() {
		t.Error("the other Vehicle did not become a creature")
	}
	if c, ok := effectstest.BattlefieldCardByID(s.G, colossus); !ok || c.IsCreature() {
		t.Error("the Colossus became a creature itself")
	}
}

func TestRprB25PeacewalkerColossusRefusesNonVehiclesAndOpposingVehicles(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	colossus := s.PushVehicle("Peacewalker Colossus", rprb25ColossusOracle, 6, 6)
	bear := s.PushCreature("Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(s.G, game.Card{
		InstanceID: uuid.New(), Name: "Their Cart", TypeLine: effectstest.CrewCardsVehicleTypeLine,
		Power: 3, Toughness: 3, Owner: s.Opp.ID, Controller: s.Opp.ID,
	})
	effectstest.ToMain(t, s.G)
	effectstest.B06AddMana(s.Me, "W", "C", "W", "C")
	for name, id := range map[string]uuid.UUID{"a Bear": bear, "an opposing Vehicle": theirs} {
		if err := s.G.ActivateCatalogAbility(s.Me.ID, colossus, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(id)}); err == nil {
			t.Errorf("%s was accepted as the target", name)
		}
	}
}

// --- Elemental Mastery -------------------------------------------------------

func TestRprB25ElementalMasteryMakesPowerManyHastyElementalsAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Elemental Mastery", "Enchantment — Aura", rprb25MasteryOracle, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
	effectstest.AdvanceToMain(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature lacks the granted ability")
	}
	effectstest.B16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, host) {
		t.Error("the enchanted creature did not tap")
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Elemental")
	if len(ids) != 3 {
		t.Fatalf("%d Elementals, want 3 (the host's power)", len(ids))
	}
	for _, id := range ids {
		c, _ := g.LookupCardForEffect(id)
		if c.Power != 1 || c.Toughness != 1 || c.Controller != me.ID || !effectstest.HasKeywordOf(g, id, "haste") {
			t.Errorf("token %+v, want a 1/1 red Elemental with haste under my control", c)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if g.Battlefield.Contains(id) {
			t.Error("an Elemental survived the end step")
		}
	}
}

// --- Dance of the Manse ------------------------------------------------------

func rprb25Grave(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func rprb25CastDance(t *testing.T, g *game.Game, x int, targets ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, g.Seats[g.Turn.ActiveSeat], strings.Repeat("C", x)+"WU")
	id, err := effectstest.CastWithTapParams(t, g, "Dance of the Manse", "Sorcery", "{X}{W}{U}", rprb25DanceOracle, game.CastSpellParams{
		Strict: true, XValue: x, Targets: effectstest.CardRefs(targets...),
	})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return id, err
}

func TestRprB25DanceOfTheManseReturnsUpToXCardsOfManaValueXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := rprb25Grave(me, "Rock", "Artifact", "{1}")
	ench := rprb25Grave(me, "Ward", "Enchantment", "{2}")
	if _, err := rprb25CastDance(t, g, 2, rock, ench); err != nil {
		t.Fatal(err)
	}
	for name, id := range map[string]uuid.UUID{"artifact": rock, "enchantment": ench} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the %s did not return", name)
		}
	}
	if c, _ := g.LookupCardForEffect(rock); c.IsCreature() {
		t.Error("an X of 2 made the artifact a creature")
	}
}

func TestRprB25DanceOfTheManseRefusesTooManyTooBigAndAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := rprb25Grave(me, "Rock A", "Artifact", "{1}")
	b := rprb25Grave(me, "Rock B", "Artifact", "{1}")
	big := rprb25Grave(me, "Big Rock", "Artifact", "{3}")
	aura := rprb25Grave(me, "Pacifism", "Enchantment — Aura", "{1}")
	creature := rprb25Grave(me, "Bear", "Creature — Bear", "{1}")
	for name, targets := range map[string][]uuid.UUID{
		"two targets with X = 1": {a, b},
		"a card of mana value 3": {big},
		"an Aura":                {aura},
		"a creature card":        {creature},
	} {
		x := 1
		if name == "a card of mana value 3" || name == "an Aura" || name == "a creature card" {
			x = 2
		}
		if _, err := rprb25CastDance(t, g, x, targets...); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

func TestRprB25DanceOfTheManseMakesFourFourCreaturesAtSixOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := rprb25Grave(me, "Rock", "Artifact", "{6}")
	ench := rprb25Grave(me, "Ward", "Enchantment", "{2}")
	if _, err := rprb25CastDance(t, g, 6, rock, ench); err != nil {
		t.Fatal(err)
	}
	for name, id := range map[string]uuid.UUID{"artifact": rock, "enchantment": ench} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("the %s did not return", name)
		}
		if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 4 || tg != 4 {
			t.Errorf("the %s is %d/%d, want 4/4", name, p, tg)
		}
		if c, _ := g.LookupCardForEffect(id); !c.IsCreature() {
			t.Errorf("the %s is not a creature", name)
		}
	}
	if c, _ := g.LookupCardForEffect(rock); !c.IsArtifact() {
		t.Error("the Rock lost its artifact type")
	}
}

// --- Jace, Architect of Thought ----------------------------------------------

const rprb25JaceType = "Legendary Planeswalker — Jace"

func TestRprB25JacePlusOneShrinksOpposingAttackersUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Jace, Architect of Thought", rprb25JaceType, rprb25JaceOracle, 4)
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	if got := effectstest.CountersOn(g, walker, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	attackerSeat := g.Seats[g.Turn.ActiveSeat]
	raider := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: attackerSeat.ID, Controller: attackerSeat.ID,
	})
	idle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Idle", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: attackerSeat.ID, Controller: attackerSeat.ID,
	})
	effectstest.DeclareAttack(t, g, me.ID, raider)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, raider); p != 2 {
		t.Errorf("the attacking creature has power %d, want 3-1", p)
	}
	if p := effectstest.EffectivePower(t, g, idle); p != 3 {
		t.Errorf("a creature that did not attack has power %d, want 3", p)
	}
}

func TestRprB25JacePlusOneIgnoresYourOwnAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := effectstest.RfpbWalker(g, me.ID, "Jace, Architect of Thought", rprb25JaceType, rprb25JaceOracle, 4)
	mine := effectstest.B16Creature(g, me.ID, "Mine", "Creature — Ogre", 3, 3, "G")
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("my own attacker has power %d, want 3", p)
	}
}

func TestRprB25JaceMinusTwoTakesOnePileAndBottomsTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := effectstest.SeedFactOrFictionLibrary(me, "Card One", "Card Two", "Card Three")
	walker := effectstest.RfpbWalker(g, me.ID, "Jace, Architect of Thought", rprb25JaceType, rprb25JaceOracle, 4)
	hand := me.Hand.Size()

	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil || len(split.ChooseCards) != 3 {
		t.Fatalf("the opponent should separate three cards: %+v", split)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, revealed[:1]); err != nil {
		t.Fatalf("separate: %v", err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !me.Hand.Contains(revealed[0]) || me.Hand.Size() != hand+1 {
		t.Errorf("the one-card pile did not reach my hand (hand %d -> %d)", hand, me.Hand.Size())
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 2 {
		t.Fatalf("the other pile is ordered onto the bottom by me: %+v", order)
	}
	want := []uuid.UUID{revealed[2], revealed[1]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, want, nil); err != nil {
		t.Fatalf("order the bottom: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, want) {
		t.Errorf("the bottom two are %v, want %v", got, want)
	}
	if got := effectstest.CountersOn(g, walker, game.CounterLoyalty); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestRprB25JaceMinusEightExilesANonlandFromEachLibraryAndOffersThemFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Jace, Architect of Thought", rprb25JaceType, rprb25JaceOracle, 8)
	spells := map[uuid.UUID]uuid.UUID{}
	lands := map[uuid.UUID]uuid.UUID{}
	var order []uuid.UUID
	for _, s := range g.Seats {
		ids := effectstest.SeedSearchLibrary(s,
			game.Card{Name: "Spell of " + s.Name, TypeLine: "Sorcery", ManaCost: "{2}{R}"},
			game.Card{Name: "Basic Forest", TypeLine: "Basic Land — Forest"},
		)
		spells[s.ID], lands[s.ID] = ids[0], ids[1]
		order = append(order, ids[0])
	}
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for range g.Seats {
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no search prompt for the next library")
		}
		var spell uuid.UUID
		for _, id := range c.SearchCards {
			for _, want := range spells {
				if id == want {
					spell = id
				}
			}
			for _, land := range lands {
				if id == land {
					t.Errorf("a land was offered for the nonland search")
				}
			}
		}
		if spell == uuid.Nil {
			t.Fatalf("the nonland card was not offered: %v", c.SearchCards)
		}
		effectstest.AnswerSearchByID(t, g, me.ID, spell)
	}
	for owner, id := range spells {
		if !g.Exile.Contains(id) {
			t.Errorf("the nonland card of %s was not exiled", owner)
		}
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, order...)
	if got := effectstest.CountersOn(g, walker, game.CounterLoyalty); got != 0 {
		t.Errorf("loyalty %d, want 0", got)
	}
}
