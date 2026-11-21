package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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
		g := effectstest.NewCatalogGame(t)
		land := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
		if c := effectstest.FindBattlefieldCardByID(g, land); c == nil || !c.Tapped {
			t.Errorf("%s should enter tapped", tc.name)
		}
		spec, _ := effects.Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced || !spec.ManaAbilities[0].Cost.Tap {
			t.Errorf("%s mana abilities %+v, want {T}: %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

// --- Goobbue Gardener ---------------------------------------------------------------

func TestFinCGGoobbueGardenerTapsForGreenOnceItIsNotSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Goobbue Gardener", "Creature — Plant Beast", finCGGoobbueOracle, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Gardener tapped for mana")
	}
	ready := effectstest.PushCatalogPermanent(g, me.ID, "Goobbue Gardener", "Creature — Plant Beast", finCGGoobbueOracle, false)
	before := me.ManaPool.CountColor("G")
	effectstest.ActivateManaFor(t, g, me.ID, ready, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("G"); got != before+1 {
		t.Errorf("green in the pool %d → %d, want one more", before, got)
	}
	if c := effectstest.FindBattlefieldCardByID(g, ready); c == nil || !c.Tapped {
		t.Error("the Gardener did not tap")
	}
}

// --- Coeurl -------------------------------------------------------------------------

func TestFinCGCoeurlTapsANonenchantmentCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	coeurl := effectstest.PushCatalogPermanent(g, me.ID, "Coeurl", "Creature — Cat Beast", finCGCoeurlOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	god := effectstest.B12Creature(g, opp.ID, "God", "Enchantment Creature — God", 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, coeurl, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, bear); c == nil || !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if c := effectstest.FindBattlefieldCardByID(g, coeurl); c == nil || !c.Tapped {
		t.Error("Coeurl did not tap as its cost")
	}

	// A fresh Coeurl can't name the enchantment creature.
	other := effectstest.PushCatalogPermanent(g, me.ID, "Coeurl", "Creature — Cat Beast", finCGCoeurlOracle, false)
	effectstest.B10AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(god)}); err == nil {
		t.Error("an enchantment creature was a legal target")
	}
}

// --- Dragoon's Wyvern ---------------------------------------------------------------

func TestFinCGDragoonsWyvernFliesAndMakesAHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wyvern := effectstest.CastRingCreature(t, g, "Dragoon's Wyvern", "Creature — Drake", finCGWyvernOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, wyvern, "flying") {
		t.Error("the Wyvern has no flying")
	}
	heroes := effectstest.SocB8Named(g, me.ID, "Hero")
	if len(heroes) != 1 {
		t.Fatalf("%d Heroes, want 1", len(heroes))
	}
	if v := effectstest.S58p6View(t, g, heroes[0].InstanceID); v.CurrentPower() != 1 || v.CurrentToughness() != 1 || len(v.Effective().Colors) != 0 {
		t.Errorf("the Hero is %d/%d %v, want a colorless 1/1", v.CurrentPower(), v.CurrentToughness(), v.Effective().Colors)
	}
}

// --- Dwarven Castle Guard -----------------------------------------------------------

func TestFinCGDwarvenCastleGuardDiesIntoAHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guard := effectstest.B12Push(g, me.ID, "Dwarven Castle Guard", "Creature — Dwarf Soldier", finCGGuardOracle, 2, 1)
	if n := len(effectstest.SocB8Named(g, me.ID, "Hero")); n != 0 {
		t.Fatalf("%d Heroes before it died", n)
	}
	effectstest.B18Kill(t, g, guard)
	if g.Battlefield.Contains(guard) {
		t.Fatal("the Guard is still on the battlefield")
	}
	if n := len(effectstest.SocB8Named(g, me.ID, "Hero")); n != 1 {
		t.Errorf("%d Heroes after it died, want 1", n)
	}
}

// --- Gigantoad ----------------------------------------------------------------------

func TestFinCGGigantoadGetsTwoTwoWithSevenLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	toad := effectstest.B12Push(g, me.ID, "Gigantoad", "Creature — Frog", finCGGigantoadOracle, 4, 4)
	for i := 0; i < 6; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	if v := effectstest.S58p6View(t, g, toad); v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
		t.Fatalf("with six lands it is %d/%d, want 4/4", v.CurrentPower(), v.CurrentToughness())
	}
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if v := effectstest.S58p6View(t, g, toad); v.CurrentPower() != 6 || v.CurrentToughness() != 6 {
		t.Errorf("with seven lands it is %d/%d, want 6/6", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- Gaelicat -----------------------------------------------------------------------

func TestFinCGGaelicatGetsTwoPowerWithTwoArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := effectstest.B12Push(g, me.ID, "Gaelicat", "Creature — Cat", finCGGaelicatOracle, 1, 3)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, cat, kw) {
			t.Errorf("Gaelicat has no %s", kw)
		}
	}
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if v := effectstest.S58p6View(t, g, cat); v.CurrentPower() != 1 || v.CurrentToughness() != 3 {
		t.Fatalf("with one artifact of yours it is %d/%d, want 1/3", v.CurrentPower(), v.CurrentToughness())
	}
	effectstest.B12Permanent(g, me.ID, "Second Rock", "Artifact")
	if v := effectstest.S58p6View(t, g, cat); v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Errorf("with two artifacts of yours it is %d/%d, want 3/3", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- Eject --------------------------------------------------------------------------

func TestFinCGEjectCantBeCounteredBouncesAndDraws(t *testing.T) {
	spec, _ := effects.Lookup(finCGEjectOracle)
	if !spec.CantBeCountered {
		t.Error("Eject can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Eject", "Instant", finCGEjectOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the spell, want %d (one drawn)", got, hand+1)
	}
	land := effectstest.SeedPermanentWithOracle(g, opp.ID, "Their Forest", "Basic Land — Forest", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Eject", "Instant", finCGEjectOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Dreams of Laguna ---------------------------------------------------------------

func TestFinCGDreamsOfLagunaSurveilsThenDrawsAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g) // past the draw step, so the seeded cards stay on top
	ids := effectstest.SeedLibrary(me, "Top", "Second")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Dreams of Laguna", "Instant", finCGDreamsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %+v, want one card", c)
	}
	// Bin the top card: the draw takes the card under it.
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g2 := effectstest.NewCatalogGame(t)
	active := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.SeedLibrary(active, "One", "Two")
	id := effectstest.SeedGraveyardCard(t, g2, "Dreams of Laguna", "Instant", finCGDreamsOracle)
	if err := g2.CastSpell(active.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	fc := effectstest.SurveilChoiceFor(g2, active.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	b := effectstest.GraveyardCardOf(g, me, "Dead Elf", "Creature — Elf", "{G}", "")
	spell := effectstest.GraveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	for _, bad := range [][]game.TargetRef{effectstest.CardRefs(spell), effectstest.CardRefs(theirs), effectstest.CardRefs(a, b, spell)} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Fight On!", "Instant", finCGFightOnOracle, bad); err == nil {
			t.Fatalf("targets %+v were accepted", bad)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Fight On!", "Instant", finCGFightOnOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not return to hand")
	}
	if !me.Graveyard.Contains(spell) || !opp.Graveyard.Contains(theirs) {
		t.Error("a card that was not targeted left its graveyard")
	}
	// Up to two: no target is a legal cast.
	effectstest.CastCatalogSpell(t, g, "Fight On!", "Instant", finCGFightOnOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Coliseum Behemoth --------------------------------------------------------------

func TestFinCGColiseumBehemothDrawsOrDestroysAnArtifactOrEnchantment(t *testing.T) {
	t.Run("draw", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		hand := me.Hand.Size()
		behemoth := effectstest.CastRingCreature(t, g, "Coliseum Behemoth", "Creature — Beast", finCGBehemothOracle, 7, 7)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, behemoth, "trample") {
			t.Error("the Behemoth has no trample")
		}
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("hand %d → %d, want one card drawn", hand, got)
		}
	})
	t.Run("destroy", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		aura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.CastRingCreature(t, g, "Coliseum Behemoth", "Creature — Beast", finCGBehemothOracle, 7, 7)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("no target prompt for the destroy mode")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a creature was accepted as a target")
		}
		effectstest.PickCard(t, g, me.ID, aura)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(aura) {
			t.Error("the enchantment survived")
		}
		if !g.Battlefield.Contains(bear) {
			t.Error("the creature was destroyed")
		}
	})
}
