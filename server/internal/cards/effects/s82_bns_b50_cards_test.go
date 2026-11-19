package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsB50MemorialGeniusOracle = "81763d7d-3897-4be9-bbf6-f6f5dee366ff"
	bnsB50MemorialGloryOracle  = "eb8ec34c-ae07-4a09-940f-ee965146a787"
	bnsB50IpnuRivuletOracle    = "c17d799f-adc9-4c41-87cf-b243b5ea3be1"
	bnsB50GodsendOracle        = "dd91f4f2-ea46-48c5-8a80-0dc4ae5b79c4"
	bnsB50ApheliaOracle        = "191e003f-1a78-4d85-9c70-2ae132a5f9cc"
	bnsB50AvernaOracle         = "23695bc2-86d3-4d78-b374-9e83c4cc9cc4"
	bnsB50BrauliosOracle       = "70e247e5-6ca4-458c-9519-d19b400955f6"
	bnsB50YamazakiOracle       = "5c820811-1775-4ea4-b003-721750a4d90f"
	bnsB50CleonOracle          = "5e0e5060-2be0-476b-8108-c8c7cd76942f"
	bnsB50CynetteOracle        = "8170ddd6-c054-4851-bdc2-4d319b60ba41"
	bnsB50EverethOracle        = "55b8c81b-f26b-4ae1-94ba-30b19e6a944e"
)

func TestBnsB50EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB50MemorialGeniusOracle: "Memorial to Genius",
		bnsB50MemorialGloryOracle:  "Memorial to Glory",
		bnsB50IpnuRivuletOracle:    "Ipnu Rivulet",
		bnsB50GodsendOracle:        "Godsend",
		bnsB50ApheliaOracle:        "Aphelia, Viper Whisperer",
		bnsB50AvernaOracle:         "Averna, the Chaos Bloom",
		bnsB50BrauliosOracle:       "Braulios of Pheres Band",
		bnsB50YamazakiOracle:       "Brothers Yamazaki",
		bnsB50CleonOracle:          "Cleon, Merry Champion",
		bnsB50CynetteOracle:        "Cynette, Jelly Drover",
		bnsB50EverethOracle:        "Evereth, Viceroy of Plunder",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s is not registered under %s", name, oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) > 0 {
			t.Errorf("%s must be Full with no caveats", name)
		}
	}
}

func TestBnsB50MemorialToGeniusEntersTappedAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Memorial to Genius", bnsB50MemorialGeniusOracle)
	if !b27Tapped(t, g, land) {
		t.Error("it enters tapped")
	}
	b22Untap(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want U", got)
	}
	me.ManaPool = nil
	b22Untap(g, land)
	// A pool that can't pay is refused.
	b06AddMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with 2 mana, wants 5")
	}
	b06AddMana(me, "C", "C", "C")
	hand := len(me.Hand.Cards)
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	if len(me.Hand.Cards) != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, len(me.Hand.Cards))
	}
	if g.Battlefield.Contains(land) {
		t.Error("it is sacrificed")
	}
}

func TestBnsB50MemorialToGloryMakesTwoSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := playLandFromHand(t, g, "Memorial to Glory", bnsB50MemorialGloryOracle)
	if !b27Tapped(t, g, land) {
		t.Error("it enters tapped")
	}
	b22Untap(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want W", got)
	}
	me.ManaPool = nil
	b22Untap(g, land)
	b06AddMana(me, "C", "C", "C", "W")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && c.IsToken() && c.IsCreature() && c.Controller == me.ID {
			n++
			if p, tt := effectivePTOf(t, g, c.InstanceID); p != 1 || tt != 1 {
				t.Errorf("Soldier %d/%d, want 1/1", p, tt)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Soldiers, want 2", n)
	}
	if g.Battlefield.Contains(land) {
		t.Error("it is sacrificed")
	}
}

func TestBnsB50IpnuRivuletMillsFourBySacrificingADesert(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	rivulet := b12Push(g, me.ID, "Ipnu Rivulet", "Land — Desert", bnsB50IpnuRivuletOracle, 0, 0)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, rivulet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" || me.Life != before {
		t.Errorf("pool %v life %d", got, me.Life)
	}
	me.ManaPool = nil
	b22Untap(g, rivulet)
	if err := g.ActivateManaAbility(me.ID, rivulet, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d, want %d", me.Life, before-1)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want U", got)
	}
	me.ManaPool = nil
	b22Untap(g, rivulet)
	seedLibrary(opp, "A", "B", "C", "D", "E", "F")
	lib := len(opp.Library.Cards)
	gy := len(opp.Graveyard.Cards)
	// A non-Desert can't be sacrificed to it.
	forest := b12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	b06AddMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, rivulet, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{forest}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("sacrificed a Forest as a Desert")
	}
	b16Activate(t, g, me.ID, rivulet, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rivulet}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if len(opp.Library.Cards) != lib-4 || len(opp.Graveyard.Cards) != gy+4 {
		t.Errorf("library %d→%d graveyard %d→%d, want 4 milled", lib, len(opp.Library.Cards), gy, len(opp.Graveyard.Cards))
	}
	if g.Battlefield.Contains(rivulet) {
		t.Error("the Rivulet is sacrificed")
	}
}

func TestBnsB50CynetteMakesAFlyingJellyfishOnEnteringAndDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b06AddMana(me, "C", "C", "C", "U")
	cynette := castWithCost(t, g, "Cynette, Jelly Drover", "Legendary Creature — Human Wizard", "{3}{U}", bnsB50CynetteOracle)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cynette) {
		t.Fatal("Cynette did not resolve")
	}
	jelly := findBattlefieldByName(g, "Jellyfish")
	if jelly == uuid.Nil {
		t.Fatal("no Jellyfish on entering")
	}
	// The Jellyfish flies, so Cynette's anthem makes it 2/2.
	if p, tt := effectivePTOf(t, g, jelly); p != 2 || tt != 2 {
		t.Errorf("Jellyfish %d/%d, want 2/2", p, tt)
	}
	if !hasKeywordOf(g, jelly, "flying") {
		t.Error("the Jellyfish has flying")
	}
	// A non-flyer is untouched.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tt := effectivePTOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("Bear %d/%d, want 2/2", p, tt)
	}
	b15Destroy(t, g, cynette)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Jellyfish" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d Jellyfish after dying, want 2", n)
	}
	if p, tt := effectivePTOf(t, g, jelly); p != 1 || tt != 1 {
		t.Errorf("Jellyfish %d/%d once Cynette is gone, want 1/1", p, tt)
	}
}

func TestBnsB50BrothersYamazakiLegendRuleExemptionIsExactlyTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	a := b12Push(g, me.ID, "Brothers Yamazaki", "Legendary Creature — Human Samurai", bnsB50YamazakiOracle, 2, 1)
	b := b12Push(g, me.ID, "Brothers Yamazaki", "Legendary Creature — Human Samurai", bnsB50YamazakiOracle, 2, 1)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("exactly two Brothers Yamazaki: the legend rule doesn't apply")
	}
	// Each is another Yamazaki to the other: 4/3 with haste.
	for _, id := range []uuid.UUID{a, b} {
		if p, tt := effectivePTOf(t, g, id); p != 4 || tt != 3 {
			t.Errorf("Yamazaki %d/%d, want 4/3", p, tt)
		}
		if !hasKeywordOf(g, id, "haste") {
			t.Error("each has haste from the other")
		}
	}
	// An opponent's copy makes it three: the rule applies to mine.
	b12Push(g, opp.ID, "Brothers Yamazaki", "Legendary Creature — Human Samurai", bnsB50YamazakiOracle, 2, 1)
	g.RunStateChecksForTest()
	asked := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceLegendRule && c.Chooser == me.ID {
			asked = true
		}
	}
	if !asked {
		t.Error("with three on the battlefield the legend rule applies: the controller of two must choose which to keep")
	}
}

func TestBnsB50BrothersYamazakiAloneIsPlain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Push(g, me.ID, "Brothers Yamazaki", "Legendary Creature — Human Samurai", bnsB50YamazakiOracle, 2, 1)
	if p, tt := effectivePTOf(t, g, a); p != 2 || tt != 1 {
		t.Errorf("a lone Yamazaki is %d/%d, want 2/1", p, tt)
	}
	if hasKeywordOf(g, a, "haste") {
		t.Error("a lone Yamazaki has no haste")
	}
}

func TestBnsB50BrauliosCountsLandsAndRamps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	braulios := b12Push(g, me.ID, "Braulios of Pheres Band", "Legendary Creature — Centaur Scout", bnsB50BrauliosOracle, 0, 0)
	for i := 0; i < 3; i++ {
		b12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	}
	if p, tt := effectivePTOf(t, g, braulios); p != 3 || tt != 3 {
		t.Errorf("Braulios %d/%d, want 3/3 with three lands", p, tt)
	}
	b12Push(g, g.Seats[1].ID, "Island", "Basic Land — Island", "", 0, 0)
	if p, _ := effectivePTOf(t, g, braulios); p != 3 {
		t.Errorf("an opponent's land doesn't count: %d", p)
	}
}

func TestBnsB50CleonHeroicExilesTheTopCardAndMayPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cleon := b12Push(g, me.ID, "Cleon, Merry Champion", "Legendary Creature — Human Soldier", bnsB50CleonOracle, 2, 2)
	if !hasKeywordOf(g, cleon, "double strike") {
		t.Error("Cleon has double strike")
	}
	other := pr7Creature(g, me.ID, "Other", 2, "W")
	top := seedLibraryTop(me, "Top Card", "Creature — Bear", "{1}{G}")
	exiled := func() bool {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == top {
				return true
			}
		}
		return false
	}
	// A spell that targets another creature is not heroic.
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(other)})
	passPriorityAroundTable(t, g)
	if exiled() {
		t.Fatal("a spell targeting another creature triggered heroic")
	}
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(cleon)})
	passPriorityAroundTable(t, g)
	if !exiled() {
		t.Fatal("the top card is exiled by heroic")
	}
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("the exiled card is playable until the end of your next turn")
	}
}
