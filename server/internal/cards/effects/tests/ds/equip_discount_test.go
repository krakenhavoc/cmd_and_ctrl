package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// equip_discount_test.go — #3778: the equip bit on the activation's
// cost query (game.AbilityCostSubject.Equip) and the cards that read
// it. Loxodon Warhammer's equip is {3}.

const (
	helitrooperOracle        = "a2900c79-8e55-49f9-b548-5890a753932e"
	dwarvenMaulerOracle      = "70538676-2ffb-4cad-827e-b55e81088678"
	ferventChampionOracle    = "8de50bda-1ef6-4ccf-9efc-fab10b7eb2e5"
	bureauHeadmasterOracle   = "c022b5df-1d85-44c0-94a1-d01b345b845b"
	cloudPlanetsChampOracle  = "8512bd08-ecba-4e0e-b65c-8209764a65e5"
	sailmongerOracle         = "b15af529-a61a-4228-bf2a-7304c9050720"
	inspiringStatuaryOracle  = "78b2b68d-6e27-49f3-8373-6e19faf6f679"
	summoningMateriaOracle   = "2ca9e6e5-5eec-4593-8816-5c4317c79eb0"
	summoningMateriaTypeLine = "Artifact — Equipment"
)

func TestHelitrooperDiscountsEquipThatTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	heli := effectstest.PushCatalogPermanent(g, me.ID, "Helitrooper", "Creature — Human Soldier", helitrooperOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)

	if got := effectstest.EquipPrice(t, g, me.ID, hammer, heli); got != "{1}" {
		t.Errorf("equip onto Helitrooper: %s, want {1}", got)
	}
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, bear); got != "{3}" {
		t.Errorf("equip onto another creature: %s, want the printed {3}", got)
	}

	// The engine charges the same number it shows.
	if err := effectstest.StrictEquip(t, g, me.ID, hammer, bear, "{C}"); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip a bear for {C}: err = %v, want insufficient mana", err)
	}
	if err := effectstest.StrictEquip(t, g, me.ID, hammer, heli, ""); err != nil {
		t.Fatalf("equip Helitrooper for {C}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, hammer, heli) {
		t.Error("the Warhammer should be attached to Helitrooper")
	}
}

// A non-equip ability aimed at Helitrooper is not an equip ability.
func TestHelitrooperDoesNotDiscountOtherAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	heli := effectstest.PushCatalogPermanent(g, me.ID, "Helitrooper", "Creature — Human Soldier", helitrooperOracle, false)
	sail := effectstest.PushCatalogPermanent(g, me.ID, "Sailmonger", "Creature — Human Monk", sailmongerOracle, false)
	p, err := g.PriceActivation(me.ID, sail, 0, []game.TargetRef{{Kind: game.TargetCard, ID: heli}})
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if got := p.Total.String(); got != "{2}" {
		t.Errorf("Sailmonger's {2} aimed at Helitrooper: %s, want {2}", got)
	}
}

func TestHelitrooperGivesAnotherAttackerFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	heli := effectstest.PushCatalogPermanent(g, me.ID, "Helitrooper", "Creature — Human Soldier", helitrooperOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.DeclareAttack(t, g, opp.ID, heli, bear)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("the other attacker should gain flying")
	}
}

func TestDwarvenMaulerAndFerventChampionDiscountTheirOwnEquip(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, want string
	}{
		{"Dwarven Mauler", dwarvenMaulerOracle, "{1}"},
		{"Fervent Champion", ferventChampionOracle, "{0}"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		host := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Creature — Test", tc.oracle, false)
		bear := effectstest.SeedBear(g, me.ID)
		hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
		if got := effectstest.EquipPrice(t, g, me.ID, hammer, host); got != tc.want {
			t.Errorf("%s: equip onto it %s, want %s", tc.name, got, tc.want)
		}
		if got := effectstest.EquipPrice(t, g, me.ID, hammer, bear); got != "{3}" {
			t.Errorf("%s: equip onto a bear %s, want {3}", tc.name, got)
		}
	}
}

func TestBureauHeadmasterDiscountsEveryEquipAndEquipmentSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Bureau Headmaster", "Creature — Human Assassin", bureauHeadmasterOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, bear); got != "{2}" {
		t.Errorf("equip under Bureau Headmaster: %s, want {2}", got)
	}

	sword := game.Card{InstanceID: uuid.New(), Name: "Test Sword", TypeLine: "Artifact — Equipment",
		ManaCost: "{3}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(sword)
	price, err := g.PriceCast(me.ID, sword, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if got := price.Total.String(); got != "{2}" {
		t.Errorf("an Equipment spell under Bureau Headmaster: %s, want {2}", got)
	}
}

func TestCloudPlanetsChampionIsArmedOnlyWhenEquippedOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	cloud := effectstest.PushCatalogPermanent(g, me.ID, "Cloud, Planet's Champion", "Legendary Creature — Human Soldier Mercenary",
		cloudPlanetsChampOracle, false)
	hammer := effectstest.SeedEquipment(g, me.ID, "Loxodon Warhammer", effectstest.WarhammerOracle)
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, cloud), "double strike") {
		t.Fatal("unequipped Cloud has double strike")
	}
	if got := effectstest.EquipPrice(t, g, me.ID, hammer, cloud); got != "{1}" {
		t.Errorf("equip onto Cloud: %s, want {1}", got)
	}
	if err := effectstest.StrictEquip(t, g, me.ID, hammer, cloud, "{C}"); err != nil {
		t.Fatalf("equip Cloud: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	abs := effectstest.EffectiveAbilities(t, g, cloud)
	if !effectstest.ContainsString(abs, "double strike") || !effectstest.ContainsString(abs, "indestructible") {
		t.Errorf("equipped Cloud on your turn: %v, want double strike and indestructible", abs)
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, cloud), "double strike") {
		t.Error("Cloud keeps double strike on another player's turn")
	}
}

func TestInspiringStatuaryGivesNonartifactSpellsImprovise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	statuary := effectstest.PushCatalogPermanent(g, me.ID, "Inspiring Statuary", "Artifact", inspiringStatuaryOracle, false)
	rock := effectstest.PushArtifactForTest(g, me.ID, "Rock")
	bears := game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bears)
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, bears.InstanceID, game.CastSpellParams{Strict: true,
		TapIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("Grizzly Bears improvised with an artifact: %v", err)
	}
	if !effectstest.IsTapped(g, rock) {
		t.Error("the improvised artifact should be tapped")
	}
	if effectstest.IsTapped(g, statuary) {
		t.Error("the Statuary was tapped though it was not named")
	}

	// An artifact spell does not have improvise.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Inspiring Statuary", "Artifact", inspiringStatuaryOracle, false)
	rock2 := effectstest.PushArtifactForTest(g2, me2.ID, "Rock")
	trinket := game.Card{InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact",
		ManaCost: "{1}", Owner: me2.ID, Controller: me2.ID}
	me2.Hand.PushTop(trinket)
	if err := g2.CastSpell(me2.ID, trinket.InstanceID, game.CastSpellParams{Strict: true,
		TapIDs: []uuid.UUID{rock2}}); err == nil {
		t.Error("an artifact spell was improvised under Inspiring Statuary")
	}
}

func TestSummoningMateriaCastsCreaturesFromTheTopOnlyWhileAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	materia := effectstest.SeedEquipment(g, me.ID, "Summoning Materia", summoningMateriaOracle)
	bear := effectstest.SeedBear(g, me.ID)
	top := game.Card{InstanceID: uuid.New(), Name: "Top Bear", TypeLine: "Creature — Bear",
		ManaCost: "{G}", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(top)

	if err := g.CastSpell(me.ID, top.InstanceID, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Fatal("cast from the top of the library with the Materia unattached")
	}
	if err := effectstest.StrictEquip(t, g, me.ID, materia, bear, "{C}{C}"); err != nil {
		t.Fatalf("equip the Materia: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("equipped bear: %d/%d, want 4/4", p, tt)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "vigilance") {
		t.Error("equipped bear should have vigilance")
	}
	if err := g.CastSpell(me.ID, top.InstanceID, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast the top creature with the Materia attached: %v", err)
	}
}
