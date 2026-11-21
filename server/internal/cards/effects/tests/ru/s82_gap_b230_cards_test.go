package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Legends: Battle for Baldur's Gate, GAP batch 230 (S82, #5948).

const (
	gapAstralConfrontationOracle = "80788ac7-19d6-42d2-9f6e-676d7b5c912e"
	gapContrabandLivestockOracle = "fbafe993-8118-40cc-9cb7-ee3f62a2c888"
	gapGuidingBoltOracle         = "b433cf1f-a7f3-4b68-8bfb-e4196f6e3a12"
	gapLaezelsAcrobaticsOracle   = "3efe6ac8-f92e-4108-913b-077e41c0a841"
	gapRobbersOracle             = "7d10ad59-e90d-4726-bd55-5680d7c21638"
	gapBaldursGateOracle         = "da307ea2-4df7-4d6b-be0f-9dc6ac93db61"
	gapGondGateOracle            = "4306938b-c0db-4e63-a4fb-61628e5ff41f"
	gapHeapGateOracle            = "35922a30-6b84-44dd-a2f0-306554a1ae90"
	gapStarlitSanctumOracle      = "d16298ac-67bd-4f9d-9979-23c1b7e4b359"
	gapXenagosOracle             = "942a318e-90eb-4f10-bb30-2e10f5cc922a"
)

func TestS82GapB230CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapAstralConfrontationOracle: "Astral Confrontation",
		gapContrabandLivestockOracle: "Contraband Livestock",
		gapGuidingBoltOracle:         "Guiding Bolt",
		gapLaezelsAcrobaticsOracle:   "Lae'zel's Acrobatics",
		gapRobbersOracle:             "You're Confronted by Robbers",
		gapBaldursGateOracle:         "Baldur's Gate",
		gapGondGateOracle:            "Gond Gate",
		gapHeapGateOracle:            "Heap Gate",
		gapStarlitSanctumOracle:      "Starlit Sanctum",
		gapXenagosOracle:             "Xenagos, the Reveler",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// gapB230Cast puts the card in the active seat's hand under id and casts it,
// strictly and for free.
func gapB230Cast(t *testing.T, g *game.Game, id uuid.UUID, name, typeLine, oracle string, params game.CastSpellParams) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.D5681Spell(g, me, id, name, typeLine, "", oracle)
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestGapGuidingBoltDestroysABigCreatureAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 9)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	if err := gapB230Cast(t, g, uuid.New(), "Guiding Bolt", "Instant", gapGuidingBoltOracle, game.CastSpellParams{Targets: effectstest.CardTarget(small)}); err == nil {
		t.Fatal("Guiding Bolt targeted a creature with power 3")
	}
	if err := gapB230Cast(t, g, uuid.New(), "Guiding Bolt", "Instant", gapGuidingBoltOracle, game.CastSpellParams{Targets: effectstest.CardTarget(big)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(big) {
		t.Error("the target was not destroyed")
	}
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil || len(scry.ScryCards) != 2 {
		t.Fatalf("scry prompt %+v, want 2 cards", scry)
	}
}

func TestGapContrabandLivestockGivesTheControllerATokenByTheRoll(t *testing.T) {
	for _, tc := range []struct {
		natural    int
		name       string
		power, tgh int
	}{{1, "Ox", 4, 4}, {9, "Ox", 4, 4}, {10, "Boar", 2, 2}, {19, "Boar", 2, 2}, {20, "Goat", 0, 1}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		me, opp, _ := effectstest.BlbB1Seats(g)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		effectstest.D5681Spell(g, me, id, "Contraband Livestock", "Instant", "", gapContrabandLivestockOracle)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if z := g.FindCardZoneForEffect(victim); z == nil || z.Kind != game.ZoneExile {
			t.Fatalf("natural %d: the creature is in %v, want exile", tc.natural, z)
		}
		found := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsToken() && c.Name == tc.name {
				found++
				if c.Controller != opp.ID {
					t.Errorf("natural %d: the %s is controlled by %v, want the exiled creature's controller", tc.natural, tc.name, c.Controller)
				}
				if c.Power != tc.power || c.Toughness != tc.tgh {
					t.Errorf("natural %d: %s is %d/%d, want %d/%d", tc.natural, tc.name, c.Power, c.Toughness, tc.power, tc.tgh)
				}
			}
		}
		if found != 1 {
			t.Errorf("natural %d: %d %s tokens, want 1", tc.natural, found, tc.name)
		}
	}
}

func TestGapRobbersTapsUpToThreeOrCreatesThreeSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	d := effectstest.B12Creature(g, opp.ID, "D", "Creature — Bear", 2, 2)
	if err := gapB230Cast(t, g, uuid.New(), "You're Confronted by Robbers", "Instant", gapRobbersOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(a, b, c, d)}); err == nil {
		t.Fatal("tapped four creatures")
	}
	if err := gapB230Cast(t, g, uuid.New(), "You're Confronted by Robbers", "Instant", gapRobbersOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(a, b, c)}); err != nil {
		t.Fatalf("cast stall: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("a target is untapped")
		}
	}
	if effectstest.B16Tapped(t, g, d) {
		t.Error("an untargeted creature is tapped")
	}
	if err := gapB230Cast(t, g, uuid.New(), "You're Confronted by Robbers", "Instant", gapRobbersOracle, game.CastSpellParams{Modes: []int{1}}); err != nil {
		t.Fatalf("cast aid: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	soldiers := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Soldier" && c.Controller == me.ID && c.Power == 1 && c.Toughness == 1 {
			soldiers++
		}
	}
	if soldiers != 3 {
		t.Errorf("Soldiers = %d, want 3", soldiers)
	}
}

func TestGapAstralConfrontationCostsOneLessPerOpponentYouAreAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, tc := range []struct{ id, def uuid.UUID }{{a, opp.ID}, {b, opp.ID}, {c, other.ID}} {
		if err := g.DeclareAttacker(tc.id, tc.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	// {4}{W} less two attacked opponents is {2}{W}.
	id := effectstest.HandCardFull(me, "Astral Confrontation", "Instant", "{4}{W}", gapAstralConfrontationOracle, []string{"W"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)}); err == nil {
		t.Fatal("cast for two mana")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)}); err != nil {
		t.Fatalf("cast for {2}{W} while attacking two opponents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(target); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the target is in %v, want exile", z)
	}
}

func TestGapAstralConfrontationIsFullPriceWhenNobodyIsAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Bear", 2, 2)
	id := effectstest.HandCardFull(me, "Astral Confrontation", "Instant", "{4}{W}", gapAstralConfrontationOracle, []string{"W"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{G}{G}{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)}); err == nil {
		t.Fatal("cast for four mana with no attackers")
	}
}

func TestGapHeapGateMakesATreasureByTappingAnotherGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	heap := effectstest.PushCatalogPermanent(g, me.ID, "Heap Gate", "Land — Gate", gapHeapGateOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, heap, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("made a Treasure with no other Gate to tap")
	}
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other Gate", "Land — Gate", "", false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, heap, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{other}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, other) {
		t.Error("the other Gate is untapped")
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Treasure" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("Treasures = %d, want 1", n)
	}
}

func TestGapHeapGateTapsForAnyColorForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	heap := effectstest.PushCatalogPermanent(g, me.ID, "Heap Gate", "Land — Gate", gapHeapGateOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, heap, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{1},{T}: add one mana of any color: %v", err)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 1 {
		t.Errorf("pool = %d, want the one mana added", got)
	}
}

func TestGapStarlitSanctumDrainsAndGainsFromACleric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Starlit Sanctum", "Land", gapStarlitSanctumOracle, false)
	cleric := effectstest.B12Creature(g, me.ID, "Cleric", "Creature — Human Cleric", 3, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 5)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a non-Cleric")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{cleric}}); err != nil {
		t.Fatalf("activate lifegain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 5 {
		t.Errorf("life gained = %d, want the Cleric's toughness 5", got)
	}
	cleric2 := effectstest.B12Creature(g, me.ID, "Cleric Two", "Creature — Elf Cleric", 4, 2)
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardByID(g, land); c != nil {
			c.Tapped = false
		}
	})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	oppLife := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, land, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{cleric2}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate drain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 4 {
		t.Errorf("life lost = %d, want the Cleric's power 4", got)
	}
}

func TestGapBaldursGateAddsXManaOfOneColorForTheOtherGates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	gate := effectstest.PushCatalogPermanent(g, me.ID, "Baldur's Gate", "Legendary Land — Gate", gapBaldursGateOracle, false)
	spec, _ := effects.Lookup(gapBaldursGateOracle)
	amount := func() game.ManaAmount {
		var a game.ManaAmount
		g.ReadSnapshot(func() { a = spec.ManaAbilities[1].ProducedAmount(g, me.ID, gate) })
		return a
	}
	if a := amount(); a.N != 0 {
		t.Errorf("alone: %+v, want nothing", a)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Gate One", "Land — Gate", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Gate Two", "Land — Gate", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Gate", "Land — Gate", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	if a := amount(); a.N != 2 || !a.OneColor || len(a.Options) != 5 {
		t.Errorf("with two other Gates: %+v, want two mana of one of five colours", a)
	}
}

func TestGapGondGateMakesGatesEnterUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	gond := effectstest.PushCatalogPermanent(g, me.ID, "Gond Gate", "Land — Gate", gapGondGateOracle, false)
	spec, _ := effects.Lookup(gapGondGateOracle)
	if len(spec.Replacements) != 1 {
		t.Fatalf("Gond Gate has %d replacements", len(spec.Replacements))
	}
	applies := func(owner uuid.UUID, name, typeLine string) bool {
		var out bool
		g.WithWriteLock(func() {
			c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner}
			g.Battlefield.PushTop(c)
			ev := &game.ReplacementEvent{Kind: game.RepEventMove, NewZone: game.ZoneBattlefield, EntersTapped: true, CardID: c.InstanceID}
			src, _ := g.LookupCardForEffect(gond)
			out = spec.Replacements[0].AppliesTo(ev, g, &src)
		})
		return out
	}
	if !applies(me.ID, "Gate", "Land — Gate") {
		t.Error("a Gate you control that would enter tapped is not caught")
	}
	if applies(me.ID, "Forest", "Basic Land — Forest") {
		t.Error("a Forest enters untapped under Gond Gate")
	}
	if applies(opp.ID, "Their Gate", "Land — Gate") {
		t.Error("an opponent's Gate enters untapped under Gond Gate")
	}
}

func TestGapGondGateTapsForAColourAGateYouControlCouldProduce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gond := effectstest.PushCatalogPermanent(g, me.ID, "Gond Gate", "Land — Gate", gapGondGateOracle, false)
	if err := g.ActivateManaAbility(me.ID, gond, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil && effectstest.PoolSize(g, me.ID) != 0 {
		t.Fatal("added mana with no other Gate to read")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Heap Gate", "Land — Gate", gapHeapGateOracle, false)
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardByID(g, gond); c != nil {
			c.Tapped = false
		}
	})
	if err := g.ActivateManaAbility(me.ID, gond, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{T}: add a colour a Gate could produce: %v", err)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 1 {
		t.Errorf("pool = %d, want 1", got)
	}
}

func TestGapXenagosMakesAHastySatyr(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	walker := effectstest.GuffbWalker(g, me.ID, "Xenagos, the Reveler", "Legendary Planeswalker — Xenagos", gapXenagosOracle, 3)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	satyrs := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Satyr" && c.Controller == me.ID && c.Power == 2 && c.Toughness == 2 {
			satyrs++
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "haste") {
				t.Error("the Satyr has no haste")
			}
		}
	}
	if satyrs != 1 {
		t.Fatalf("Satyrs = %d, want 1", satyrs)
	}
}

func TestGapXenagosPlusOneAddsOneManaPerCreature(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	walker := effectstest.GuffbWalker(g, me.ID, "Xenagos, the Reveler", "Legendary Planeswalker — Xenagos", gapXenagosOracle, 3)
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3 && effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana) != nil; i++ {
		effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 3 {
		t.Errorf("pool = %d, want 3 (three creatures)", got)
	}
}

func TestGapXenagosMinusSixExilesSevenAndOffersCreaturesAndLands(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	walker := effectstest.GuffbWalker(g, me.ID, "Xenagos, the Reveler", "Legendary Planeswalker — Xenagos", gapXenagosOracle, 7)
	bear, land, spell := uuid.New(), uuid.New(), uuid.New()
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deep Filler", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Library Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Library Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Library Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if prompt == nil {
		t.Fatal("no creature-and-land pick offered")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.ChooseCards {
		offered[id] = true
	}
	if !offered[bear] || !offered[land] || offered[spell] {
		t.Fatalf("offered %v, want the bear and the land but not the instant", offered)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{bear, land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, bear) && effectstest.FindBattlefieldByName(g, "Library Bear") == uuid.Nil {
		t.Error("the creature card is not on the battlefield")
	}
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the instant is in %v, want exile", z)
	}
}

func TestGapLaezelsAcrobaticsBlinksYourNontokenCreatures(t *testing.T) {
	for _, tc := range []struct {
		natural  int
		backNow  bool
		describe string
	}{{5, false, "1—9 returns them at the next end step only"}, {15, true, "10—20 returns them, exiles them again, and returns them at the next end step"}} {
		acro := uuid.New()
		g := effectstest.D5681Table(t, acro, tc.natural)
		me, opp, _ := effectstest.BlbB1Seats(g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		token := effectstest.B12Push(g, me.ID, "Token Bear", "Token Creature — Bear", "", 2, 2)
		_ = token
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.D5681Spell(g, me, acro, "Lae'zel's Acrobatics", "Instant", "", gapLaezelsAcrobaticsOracle)
		if err := g.CastSpell(me.ID, acro, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, theirs) == false {
			t.Errorf("%s: an opponent's creature was exiled", tc.describe)
		}
		if z := g.FindCardZoneForEffect(mine); tc.backNow == false && (z == nil || z.Kind != game.ZoneExile) {
			t.Errorf("%s: the creature is in %v, want exile until the end step", tc.describe, z)
		}
		if len(g.DelayedTriggers) != 1 {
			t.Errorf("%s: delayed triggers = %d, want 1 (the end-step return)", tc.describe, len(g.DelayedTriggers))
		}
		// Round to the end step: the creature is back as a new object.
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldByName(g, "Mine") == uuid.Nil {
			t.Errorf("%s: the creature did not come back at the end step", tc.describe)
		}
	}
}
