package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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

// equipPrice is what equipping `equipment` onto `target` costs `who`.
func equipPrice(t *testing.T, g *game.Game, who, equipment, target uuid.UUID) string {
	t.Helper()
	p, err := g.PriceActivation(who, equipment, 0, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if p.Total.ManaValue() == 0 {
		return "{0}"
	}
	return p.Total.String()
}

func TestHelitrooperDiscountsEquipThatTargetsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	heli := pushCatalogPermanent(g, me.ID, "Helitrooper", "Creature — Human Soldier", helitrooperOracle, false)
	bear := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)

	if got := equipPrice(t, g, me.ID, hammer, heli); got != "{1}" {
		t.Errorf("equip onto Helitrooper: %s, want {1}", got)
	}
	if got := equipPrice(t, g, me.ID, hammer, bear); got != "{3}" {
		t.Errorf("equip onto another creature: %s, want the printed {3}", got)
	}

	// The engine charges the same number it shows.
	if err := strictEquip(t, g, me.ID, hammer, bear, "{C}"); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip a bear for {C}: err = %v, want insufficient mana", err)
	}
	if err := strictEquip(t, g, me.ID, hammer, heli, ""); err != nil {
		t.Fatalf("equip Helitrooper for {C}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !attachedTo(t, g, hammer, heli) {
		t.Error("the Warhammer should be attached to Helitrooper")
	}
}

// A non-equip ability aimed at Helitrooper is not an equip ability.
func TestHelitrooperDoesNotDiscountOtherAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	heli := pushCatalogPermanent(g, me.ID, "Helitrooper", "Creature — Human Soldier", helitrooperOracle, false)
	sail := pushCatalogPermanent(g, me.ID, "Sailmonger", "Creature — Human Monk", sailmongerOracle, false)
	p, err := g.PriceActivation(me.ID, sail, 0, []game.TargetRef{{Kind: game.TargetCard, ID: heli}})
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if got := p.Total.String(); got != "{2}" {
		t.Errorf("Sailmonger's {2} aimed at Helitrooper: %s, want {2}", got)
	}
}

func TestHelitrooperGivesAnotherAttackerFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	heli := pushCatalogPermanent(g, me.ID, "Helitrooper", "Creature — Human Soldier", helitrooperOracle, false)
	bear := seedBear(g, me.ID)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	declareAttack(t, g, opp.ID, heli, bear)
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !containsString(effectiveAbilities(t, g, bear), "flying") {
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		host := pushCatalogPermanent(g, me.ID, tc.name, "Creature — Test", tc.oracle, false)
		bear := seedBear(g, me.ID)
		hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)
		if got := equipPrice(t, g, me.ID, hammer, host); got != tc.want {
			t.Errorf("%s: equip onto it %s, want %s", tc.name, got, tc.want)
		}
		if got := equipPrice(t, g, me.ID, hammer, bear); got != "{3}" {
			t.Errorf("%s: equip onto a bear %s, want {3}", tc.name, got)
		}
	}
}

func TestBureauHeadmasterDiscountsEveryEquipAndEquipmentSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Bureau Headmaster", "Creature — Human Assassin", bureauHeadmasterOracle, false)
	bear := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)
	if got := equipPrice(t, g, me.ID, hammer, bear); got != "{2}" {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cloud := pushCatalogPermanent(g, me.ID, "Cloud, Planet's Champion", "Legendary Creature — Human Soldier Mercenary",
		cloudPlanetsChampOracle, false)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)
	if containsString(effectiveAbilities(t, g, cloud), "double strike") {
		t.Fatal("unequipped Cloud has double strike")
	}
	if got := equipPrice(t, g, me.ID, hammer, cloud); got != "{1}" {
		t.Errorf("equip onto Cloud: %s, want {1}", got)
	}
	if err := strictEquip(t, g, me.ID, hammer, cloud, "{C}"); err != nil {
		t.Fatalf("equip Cloud: %v", err)
	}
	passPriorityAroundTable(t, g)
	abs := effectiveAbilities(t, g, cloud)
	if !containsString(abs, "double strike") || !containsString(abs, "indestructible") {
		t.Errorf("equipped Cloud on your turn: %v, want double strike and indestructible", abs)
	}
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if containsString(effectiveAbilities(t, g, cloud), "double strike") {
		t.Error("Cloud keeps double strike on another player's turn")
	}
}

func TestInspiringStatuaryGivesNonartifactSpellsImprovise(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	statuary := pushCatalogPermanent(g, me.ID, "Inspiring Statuary", "Artifact", inspiringStatuaryOracle, false)
	rock := pushArtifactForTest(g, me.ID, "Rock")
	bears := game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bears)
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, bears.InstanceID, game.CastSpellParams{Strict: true,
		TapIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("Grizzly Bears improvised with an artifact: %v", err)
	}
	if !isTapped(g, rock) {
		t.Error("the improvised artifact should be tapped")
	}
	if isTapped(g, statuary) {
		t.Error("the Statuary was tapped though it was not named")
	}

	// An artifact spell does not have improvise.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	pushCatalogPermanent(g2, me2.ID, "Inspiring Statuary", "Artifact", inspiringStatuaryOracle, false)
	rock2 := pushArtifactForTest(g2, me2.ID, "Rock")
	trinket := game.Card{InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact",
		ManaCost: "{1}", Owner: me2.ID, Controller: me2.ID}
	me2.Hand.PushTop(trinket)
	if err := g2.CastSpell(me2.ID, trinket.InstanceID, game.CastSpellParams{Strict: true,
		TapIDs: []uuid.UUID{rock2}}); err == nil {
		t.Error("an artifact spell was improvised under Inspiring Statuary")
	}
}

func TestSummoningMateriaCastsCreaturesFromTheTopOnlyWhileAttached(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	materia := seedEquipment(g, me.ID, "Summoning Materia", summoningMateriaOracle)
	bear := seedBear(g, me.ID)
	top := game.Card{InstanceID: uuid.New(), Name: "Top Bear", TypeLine: "Creature — Bear",
		ManaCost: "{G}", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(top)

	if err := g.CastSpell(me.ID, top.InstanceID, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Fatal("cast from the top of the library with the Materia unattached")
	}
	if err := strictEquip(t, g, me.ID, materia, bear, "{C}{C}"); err != nil {
		t.Fatalf("equip the Materia: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("equipped bear: %d/%d, want 4/4", p, tt)
	}
	if !containsString(effectiveAbilities(t, g, bear), "vigilance") {
		t.Error("equipped bear should have vigilance")
	}
	if err := g.CastSpell(me.ID, top.InstanceID, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast the top creature with the Materia attached: %v", err)
	}
}
