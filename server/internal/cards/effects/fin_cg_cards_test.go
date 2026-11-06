package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_cg_cards_test.go — Final Fantasy (FIN), cards C–G (#4462). Each
// test plays the real catalog entry and asserts what its text says.

const (
	finCGGohnOracle       = "13fbd525-c9ee-4bd6-b269-94fb264024b6"
	finCGGongagaOracle    = "ac4e1818-4873-4143-9b99-46f205d006d9"
	finCGGuadosalamOracle = "849c97e4-df15-4ecb-bdf8-283bb497d90c"
	finCGGoobbueOracle    = "b5c64331-2ba5-43ad-b77c-979875cc39e3"
	finCGCoeurlOracle     = "00d1596a-c3e2-4109-86da-388934a0c652"
	finCGWyvernOracle     = "f328012f-4696-47bb-ab43-247bdcb13f03"
	finCGGuardOracle      = "b51e0ae3-126b-4b7e-9dd1-818a5b8e512a"
	finCGGigantoadOracle  = "fa5e4aa5-e2cb-4c8c-91cb-554bec0f21d4"
	finCGGaelicatOracle   = "34bdeca4-b1e0-4fa3-afe4-bec743a080b8"
	finCGEjectOracle      = "9a18fb98-796a-43b1-8157-e3afca53a1c0"
	finCGDreamsOracle     = "e060bf7a-f266-40b4-9772-e243dcbf965b"
	finCGFightOnOracle    = "4ab42daa-8726-4973-947d-1d6c32f0ba61"
	finCGBehemothOracle   = "076cf4f1-b3d4-4175-b976-938ad851cc0e"
)

func TestFinCGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finCGGohnOracle:       "Gohn, Town of Ruin",
		finCGGongagaOracle:    "Gongaga, Reactor Town",
		finCGGuadosalamOracle: "Guadosalam, Farplane Gateway",
		finCGGoobbueOracle:    "Goobbue Gardener",
		finCGCoeurlOracle:     "Coeurl",
		finCGWyvernOracle:     "Dragoon's Wyvern",
		finCGGuardOracle:      "Dwarven Castle Guard",
		finCGGigantoadOracle:  "Gigantoad",
		finCGGaelicatOracle:   "Gaelicat",
		finCGEjectOracle:      "Eject",
		finCGDreamsOracle:     "Dreams of Laguna",
		finCGFightOnOracle:    "Fight On!",
		finCGBehemothOracle:   "Coliseum Behemoth",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Gohn, Gongaga and Guadosalam ---------------------------------------------------

func TestFinCGTownDualsEnterTappedAndTapForTheirTwoColours(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Gohn, Town of Ruin", finCGGohnOracle, "{B|G}"},
		{"Gongaga, Reactor Town", finCGGongagaOracle, "{R|G}"},
		{"Guadosalam, Farplane Gateway", finCGGuadosalamOracle, "{G|U}"},
	} {
		g := newCatalogGame(t)
		land := playLandFromHand(t, g, tc.name, tc.oracle)
		if c := findBattlefieldCardByID(g, land); c == nil || !c.Tapped {
			t.Errorf("%s should enter tapped", tc.name)
		}
		spec, _ := Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced || !spec.ManaAbilities[0].Cost.Tap {
			t.Errorf("%s mana abilities %+v, want {T}: %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

// --- Goobbue Gardener ---------------------------------------------------------------

func TestFinCGGoobbueGardenerTapsForGreenOnceItIsNotSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sick := pushCatalogPermanent(g, me.ID, "Goobbue Gardener", "Creature — Plant Beast", finCGGoobbueOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Gardener tapped for mana")
	}
	ready := pushCatalogPermanent(g, me.ID, "Goobbue Gardener", "Creature — Plant Beast", finCGGoobbueOracle, false)
	before := me.ManaPool.CountColor("G")
	activateManaFor(t, g, me.ID, ready, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("G"); got != before+1 {
		t.Errorf("green in the pool %d → %d, want one more", before, got)
	}
	if c := findBattlefieldCardByID(g, ready); c == nil || !c.Tapped {
		t.Error("the Gardener did not tap")
	}
}

// --- Coeurl -------------------------------------------------------------------------

func TestFinCGCoeurlTapsANonenchantmentCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	coeurl := pushCatalogPermanent(g, me.ID, "Coeurl", "Creature — Cat Beast", finCGCoeurlOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	god := b12Creature(g, opp.ID, "God", "Enchantment Creature — God", 4, 4)
	advanceToMain(t, g)
	b10AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, coeurl, 0, game.ActivateAbilityParams{Targets: sgCards(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, bear); c == nil || !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if c := findBattlefieldCardByID(g, coeurl); c == nil || !c.Tapped {
		t.Error("Coeurl did not tap as its cost")
	}

	// A fresh Coeurl can't name the enchantment creature.
	other := pushCatalogPermanent(g, me.ID, "Coeurl", "Creature — Cat Beast", finCGCoeurlOracle, false)
	b10AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{Targets: sgCards(god)}); err == nil {
		t.Error("an enchantment creature was a legal target")
	}
}

// --- Dragoon's Wyvern ---------------------------------------------------------------

func TestFinCGDragoonsWyvernFliesAndMakesAHero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wyvern := castRingCreature(t, g, "Dragoon's Wyvern", "Creature — Drake", finCGWyvernOracle, 2, 1)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, wyvern, "flying") {
		t.Error("the Wyvern has no flying")
	}
	heroes := socB8Named(g, me.ID, "Hero")
	if len(heroes) != 1 {
		t.Fatalf("%d Heroes, want 1", len(heroes))
	}
	if v := s58p6View(t, g, heroes[0].InstanceID); v.CurrentPower() != 1 || v.CurrentToughness() != 1 || len(v.Effective().Colors) != 0 {
		t.Errorf("the Hero is %d/%d %v, want a colorless 1/1", v.CurrentPower(), v.CurrentToughness(), v.Effective().Colors)
	}
}

// --- Dwarven Castle Guard -----------------------------------------------------------

func TestFinCGDwarvenCastleGuardDiesIntoAHero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guard := b12Push(g, me.ID, "Dwarven Castle Guard", "Creature — Dwarf Soldier", finCGGuardOracle, 2, 1)
	if n := len(socB8Named(g, me.ID, "Hero")); n != 0 {
		t.Fatalf("%d Heroes before it died", n)
	}
	b18Kill(t, g, guard)
	if g.Battlefield.Contains(guard) {
		t.Fatal("the Guard is still on the battlefield")
	}
	if n := len(socB8Named(g, me.ID, "Hero")); n != 1 {
		t.Errorf("%d Heroes after it died, want 1", n)
	}
}

// --- Gigantoad ----------------------------------------------------------------------

func TestFinCGGigantoadGetsTwoTwoWithSevenLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toad := b12Push(g, me.ID, "Gigantoad", "Creature — Frog", finCGGigantoadOracle, 4, 4)
	for i := 0; i < 6; i++ {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	if v := s58p6View(t, g, toad); v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
		t.Fatalf("with six lands it is %d/%d, want 4/4", v.CurrentPower(), v.CurrentToughness())
	}
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if v := s58p6View(t, g, toad); v.CurrentPower() != 6 || v.CurrentToughness() != 6 {
		t.Errorf("with seven lands it is %d/%d, want 6/6", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- Gaelicat -----------------------------------------------------------------------

func TestFinCGGaelicatGetsTwoPowerWithTwoArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := b12Push(g, me.ID, "Gaelicat", "Creature — Cat", finCGGaelicatOracle, 1, 3)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, cat, kw) {
			t.Errorf("Gaelicat has no %s", kw)
		}
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if v := s58p6View(t, g, cat); v.CurrentPower() != 1 || v.CurrentToughness() != 3 {
		t.Fatalf("with one artifact of yours it is %d/%d, want 1/3", v.CurrentPower(), v.CurrentToughness())
	}
	b12Permanent(g, me.ID, "Second Rock", "Artifact")
	if v := s58p6View(t, g, cat); v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("with two artifacts of yours it is %d/%d, want 3/3", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- Eject --------------------------------------------------------------------------

func TestFinCGEjectCantBeCounteredBouncesAndDraws(t *testing.T) {
	spec, _ := Lookup(finCGEjectOracle)
	if !spec.CantBeCountered {
		t.Error("Eject can be countered")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Eject", "Instant", finCGEjectOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the spell, want %d (one drawn)", got, hand+1)
	}
	land := seedPermanentWithOracle(g, opp.ID, "Their Forest", "Basic Land — Forest", "")
	if err := castCatalogSpellErr(t, g, "Eject", "Instant", finCGEjectOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Dreams of Laguna ---------------------------------------------------------------

func TestFinCGDreamsOfLagunaSurveilsThenDrawsAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g) // past the draw step, so the seeded cards stay on top
	ids := seedLibrary(me, "Top", "Second")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Dreams of Laguna", "Instant", finCGDreamsOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %+v, want one card", c)
	}
	// Bin the top card: the draw takes the card under it.
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Errorf("want the top card surveilled away and the second drawn: binned %v, drew %v",
			me.Graveyard.Contains(ids[0]), me.Hand.Contains(ids[1]))
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}

	offers := game.AlternativeCostsOfferedFromZone(finCGDreamsOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "flashback" || offers[0].ManaCost != "{3}{U}" || !offers[0].ExileOnLeavingStack {
		t.Fatalf("graveyard offers %+v, want flashback {3}{U} that exiles", offers)
	}
	g2 := newCatalogGame(t)
	active := g2.Seats[g2.Turn.ActiveSeat]
	seedLibrary(active, "One", "Two")
	id := seedGraveyardCard(t, g2, "Dreams of Laguna", "Instant", finCGDreamsOracle)
	if err := g2.CastSpell(active.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g2)
	fc := surveilChoiceFor(g2, active.ID)
	if fc == nil {
		t.Fatal("the flashed-back spell queued no surveil")
	}
	if err := g2.ResolveSurveil(fc.ID, active.ID, nil, fc.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !g2.Exile.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}
}

// --- Fight On! ----------------------------------------------------------------------

func TestFinCGFightOnReturnsUpToTwoCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := graveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	b := graveyardCardOf(g, me, "Dead Elf", "Creature — Elf", "{G}", "")
	spell := graveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	theirs := graveyardCardOf(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	for _, bad := range [][]game.TargetRef{cardRefs(spell), cardRefs(theirs), cardRefs(a, b, spell)} {
		if err := castCatalogSpellErr(t, g, "Fight On!", "Instant", finCGFightOnOracle, bad); err == nil {
			t.Fatalf("targets %+v were accepted", bad)
		}
	}
	castCatalogSpell(t, g, "Fight On!", "Instant", finCGFightOnOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not return to hand")
	}
	if !me.Graveyard.Contains(spell) || !opp.Graveyard.Contains(theirs) {
		t.Error("a card that was not targeted left its graveyard")
	}
	// Up to two: no target is a legal cast.
	castCatalogSpell(t, g, "Fight On!", "Instant", finCGFightOnOracle, nil)
	passPriorityAroundTable(t, g)
}

// --- Coliseum Behemoth --------------------------------------------------------------

func TestFinCGColiseumBehemothDrawsOrDestroysAnArtifactOrEnchantment(t *testing.T) {
	t.Run("draw", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		hand := me.Hand.Size()
		behemoth := castRingCreature(t, g, "Coliseum Behemoth", "Creature — Beast", finCGBehemothOracle, 7, 7)
		passPriorityAroundTable(t, g)
		if !hasEffectiveKeyword(t, g, behemoth, "trample") {
			t.Error("the Behemoth has no trample")
		}
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("hand %d → %d, want one card drawn", hand, got)
		}
	})
	t.Run("destroy", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		castRingCreature(t, g, "Coliseum Behemoth", "Creature — Beast", finCGBehemothOracle, 7, 7)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		p := latestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("no target prompt for the destroy mode")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a creature was accepted as a target")
		}
		pickCard(t, g, me.ID, aura)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(aura) {
			t.Error("the enchantment survived")
		}
		if !g.Battlefield.Contains(bear) {
			t.Error("the creature was destroyed")
		}
	})
}
