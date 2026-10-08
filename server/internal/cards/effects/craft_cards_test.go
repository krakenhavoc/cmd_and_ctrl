package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// craft_cards_test.go — one test per card #2124 ships on the craft
// seam (ADR 0137). The seam itself is craft_test.go.

// craftInto activates `source`'s craft ability (its ability `index`)
// with `materials`, paying `mana`, and settles the stack. Returns the
// crafted permanent, found by its back face's name.
func craftInto(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, mana, back string, materials ...uuid.UUID) game.Card {
	t.Helper()
	floatMana(t, g, p, mana)
	if err := activateCraft(g, p, source, index, materials...); err != nil {
		t.Fatalf("craft into %s: %v", back, err)
	}
	passPriorityAroundTable(t, g)
	return craftedPermanent(t, g, back)
}

func seedCraftLibrary(p *game.Player, n int) {
	names := make([]string, n)
	for i := range names {
		names[i] = "Library Card"
	}
	seedLibrary(p, names...)
}

// Visage of Dread takes two creatures from both zones at once, and Dread
// Osseosaur's "you may mill two" fires on its entry.
func TestVisageOfDreadCraftsWithTwoCreaturesFromBothZones(t *testing.T) {
	g, me, _ := spendTable(t)
	visage := pushCraftCard(g, me, visageOfDreadRow())
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	dead := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	seedCraftLibrary(me, 4)

	floatMana(t, g, me, "{C}{C}{C}{C}{C}{B}")
	if err := activateCraft(g, me, visage, 0, bear); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("one creature for \"craft with two creatures\": err = %v, want ErrInvalidParam", err)
	}
	if err := activateCraft(g, me, visage, 0, bear, dead); err != nil {
		t.Fatalf("craft with two creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	libBefore := me.Library.Size()
	passPriorityAroundTable(t, g)

	osseosaur := craftedPermanent(t, g, "Dread Osseosaur")
	if !effectiveAbilitiesContain(t, g, osseosaur.InstanceID, "menace") {
		t.Error("Dread Osseosaur has no menace")
	}
	if got := effectivePower(t, g, osseosaur.InstanceID); got != 5 {
		t.Errorf("power = %d, want 5", got)
	}
	if me.Library.Size() != libBefore-2 {
		t.Errorf("library %d -> %d, want the enters trigger to mill two", libBefore, me.Library.Size())
	}
	if len(osseosaur.CraftedWith) != 2 {
		t.Errorf("CraftedWith = %v, want both creatures", osseosaur.CraftedWith)
	}
}

// Consuming Sepulcher drains each opponent at its controller's upkeep.
func TestTithingBladeBecomesConsumingSepulcherWhichDrainsAtUpkeep(t *testing.T) {
	g, me, opp := spendTable(t)
	blade := pushCraftCard(g, me, tithingBladeRow())
	dead := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	craftInto(t, g, me, blade, 0, "{C}{C}{C}{C}{B}", "Consuming Sepulcher", dead)

	life, oppLife := me.Life, opp.Life
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d -> %d, want one lost", oppLife, opp.Life)
	}
	if me.Life != life+1 {
		t.Errorf("my life %d -> %d, want one gained", life, me.Life)
	}
}

// Clay-Fired Bricks eats an artifact card from the graveyard and comes
// back as Cosmium Kiln, whose entry makes two Gnomes its anthem pumps.
func TestClayFiredBricksCraftsIntoCosmiumKilnAndItsGnomes(t *testing.T) {
	g, me, _ := spendTable(t)
	bricks := pushCraftCard(g, me, transformRow(clayFiredBricksOracleID, "Clay-Fired Bricks", "Artifact", "{1}{W}",
		"Cosmium Kiln", "Artifact", "", "", []string{"W"}))
	rock := pushGraveyardCardTyped(me, "Spent Rock", "Artifact")
	craftInto(t, g, me, bricks, 0, "{C}{C}{C}{C}{C}{W}{W}", "Cosmium Kiln", rock)

	gnomes := battlefieldIDsNamed(g, "Gnome")
	if len(gnomes) != 2 {
		t.Fatalf("%d Gnomes, want two from the Kiln's enters trigger", len(gnomes))
	}
	for _, id := range gnomes {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 2 || tough != 2 {
			t.Errorf("Gnome is %d/%d, want 2/2 under the Kiln's anthem", p, tough)
		}
	}
}

// Braided Net taps a permanent and stops its activated abilities only
// while it stays tapped; it pays a net counter, and crafts away an
// artifact.
func TestBraidedNetLocksAPermanentWhileItRemainsTapped(t *testing.T) {
	g, me, _ := spendTable(t)
	c := deck.ToGameCard(transformRow(braidedNetOracleID, "Braided Net", "Artifact", "{2}{U}",
		"Braided Quipu", "Artifact", "", "", []string{"U"}), false)
	c.Owner, c.Controller, c.Counters = me.ID, me.ID, map[string]int{"net": 3}
	net := pushBattlefieldCardWithTimestamp(g, c)
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard", visceraSeerOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)

	if err := g.ActivateCatalogAbility(me.ID, net, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: seer}},
	}); err != nil {
		t.Fatalf("activate the Net: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := battlefieldCard(g, net); n.Counters["net"] != 2 {
		t.Errorf("net counters = %d, want one removed", n.Counters["net"])
	}
	if s, _ := battlefieldCard(g, seer); !s.Tapped {
		t.Fatal("the Seer was not tapped")
	}
	sac := game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}
	if err := g.ActivateCatalogAbility(me.ID, seer, 0, sac); !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("the locked Seer's ability: err = %v, want ErrCantActivate", err)
	}
	g.WithWriteLock(func() {
		if err := g.UntapTargetForEffect(seer); err != nil {
			t.Fatalf("untap: %v", err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, seer, 0, sac); errors.Is(err, game.ErrCantActivate) {
		t.Error("the Seer is still locked after it untapped — the lock lasts only while it remains tapped")
	}
}

// Jadeheart Attendant gains life equal to the mana value of the card
// that crafted it, and nothing for a token, which ceased to exist in
// exile and is no exiled card (CR 702.167c).
func TestJadeheartAttendantGainsTheCraftedCardsManaValue(t *testing.T) {
	row := func() cards.Card {
		return transformRow(jadeSeedstonesOracleID, "Jade Seedstones", "Artifact", "{3}{G}",
			"Jadeheart Attendant", "Artifact Creature — Golem", "7", "7", []string{"G"})
	}
	g, me, _ := spendTable(t)
	seed := pushCraftCard(g, me, row())
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}{G}", Owner: me.ID, Controller: me.ID})
	life := me.Life
	craftInto(t, g, me, seed, 0, "{C}{C}{C}{C}{C}{G}{G}", "Jadeheart Attendant", dead)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 for a four-mana creature card", life, me.Life)
	}

	g2, me2, _ := spendTable(t)
	seed2 := pushCraftCard(g2, me2, row())
	token := pushCatalogPermanent(g2, me2.ID, "Soldier", "Token Creature — Soldier", "", false)
	life = me2.Life
	craftInto(t, g2, me2, seed2, 0, "{C}{C}{C}{C}{C}{G}{G}", "Jadeheart Attendant", token)
	if me2.Life != life {
		t.Errorf("life %d -> %d, want nothing for a token material", life, me2.Life)
	}
}

// Sovereign's Macuahuitl attaches itself as it enters and pumps.
func TestIdolOfTheDeepKingCraftsIntoAnEquipmentThatAttachesItself(t *testing.T) {
	g, me, _ := spendTable(t)
	idol := pushCraftCard(g, me, transformRow(idolOfTheDeepKingOracleID, "Idol of the Deep King", "Artifact", "{2}{R}",
		"Sovereign's Macuahuitl", "Artifact — Equipment", "", "", []string{"R"}))
	rock := pushCatalogPermanent(g, me.ID, "Mind Stone", "Artifact", "", false)
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	floatMana(t, g, me, "{C}{C}{R}")
	if err := activateCraft(g, me, idol, 0, rock); err != nil {
		t.Fatalf("craft: %v", err)
	}
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	club := craftedPermanent(t, g, "Sovereign's Macuahuitl")
	if club.AttachedTo.ID != bear {
		t.Errorf("Macuahuitl attached to %v, want the bear", club.AttachedTo)
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("bear power = %d, want 1 + 2", got)
	}
}

// Oteclan Levitator is a 1/4 flier.
func TestOteclanLandmarkCraftsIntoAFlyingLevitator(t *testing.T) {
	g, me, _ := spendTable(t)
	mark := pushCraftCard(g, me, transformRow(oteclanLandmarkOracleID, "Oteclan Landmark", "Artifact", "{W}",
		"Oteclan Levitator", "Artifact Creature — Golem", "1", "4", []string{"W"}))
	rock := pushCatalogPermanent(g, me.ID, "Mind Stone", "Artifact", "", false)
	lev := craftInto(t, g, me, mark, 0, "{C}{C}{W}", "Oteclan Levitator", rock)
	if !effectiveAbilitiesContain(t, g, lev.InstanceID, "flying") || effectiveToughness(t, g, lev.InstanceID) != 4 {
		t.Error("Oteclan Levitator is not a 1/4 flier")
	}
}

// Inverted Iceberg's entry mills a card and then draws one.
func TestInvertedIcebergMillsThenDrawsOnEntry(t *testing.T) {
	g, me, _ := spendTable(t)
	seedCraftLibrary(me, 3)
	hand, grave, lib := me.Hand.Size(), me.Graveyard.Size(), me.Library.Size()
	importToBattlefield(t, g, transformRow(invertedIcebergOracleID, "Inverted Iceberg", "Artifact", "{1}{U}",
		"Iceberg Titan", "Artifact Creature — Golem", "6", "6", []string{"U"}), me)
	passPriorityAroundTable(t, g)
	// The Iceberg went through the hand on its way in, so the hand is
	// up by the one card drawn.
	if me.Graveyard.Size() != grave+1 || me.Hand.Size() != hand+1 || me.Library.Size() != lib-2 {
		t.Errorf("graveyard %d->%d, hand %d->%d, library %d->%d — want one milled and one drawn",
			grave, me.Graveyard.Size(), hand, me.Hand.Size(), lib, me.Library.Size())
	}
}

// Craft with Cave takes a Cave of any card type — here a land card in
// the graveyard — and refuses a land that is not a Cave.
func TestKaslemsStonetreeCraftsWithACave(t *testing.T) {
	g, me, _ := spendTable(t)
	tree := pushCraftCard(g, me, transformRow(kaslemsStonetreeOracleID, "Kaslem's Stonetree", "Artifact", "{2}{G}",
		"Kaslem's Strider", "Artifact Creature — Golem", "5", "5", []string{"G"}))
	forest := pushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	cave := pushGraveyardCardTyped(me, "Cavernous Maw", "Land — Cave")
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{G}")
	if err := activateCraft(g, me, tree, 0, forest); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a Forest for \"craft with Cave\": err = %v, want ErrIllegalTarget", err)
	}
	if err := activateCraft(g, me, tree, 0, cave); err != nil {
		t.Fatalf("craft with Cave: %v", err)
	}
	passPriorityAroundTable(t, g)
	if s := craftedPermanent(t, g, "Kaslem's Strider"); effectivePower(t, g, s.InstanceID) != 5 {
		t.Error("Kaslem's Strider is not a 5/5")
	}
}

// Craft with Island takes an Island you control; the Hulk mills for {T}.
func TestWaterloggedHulkMillsAndCraftsWithAnIsland(t *testing.T) {
	g, me, _ := spendTable(t)
	hulk := pushCraftCard(g, me, transformRow(waterloggedHulkOracleID, "Waterlogged Hulk", "Artifact", "{U}",
		"Watertight Gondola", "Artifact — Vehicle", "4", "4", []string{"U"}))
	seedCraftLibrary(me, 2)
	grave := me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{T}: mill a card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != grave+1 {
		t.Fatalf("graveyard %d -> %d, want one milled", grave, me.Graveyard.Size())
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(hulk) })
	island := pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	gondola := craftInto(t, g, me, hulk, 1, "{C}{C}{C}{U}", "Watertight Gondola", island)
	if !effectiveAbilitiesContain(t, g, gondola.InstanceID, "vigilance") {
		t.Error("Watertight Gondola has no vigilance")
	}
}

// Bladewheel Chariot becomes a creature by tapping two other artifacts.
func TestSpringLoadedSawbladesCraftsIntoAChariotTwoArtifactsAnimate(t *testing.T) {
	g, me, _ := spendTable(t)
	saw := pushCraftCard(g, me, transformRow(springLoadedSawbladesOracleID, "Spring-Loaded Sawblades", "Artifact", "{1}{W}",
		"Bladewheel Chariot", "Artifact — Vehicle", "5", "5", []string{"W"}))
	fuel := pushGraveyardCardTyped(me, "Spent Rock", "Artifact")
	chariot := craftInto(t, g, me, saw, 0, "{C}{C}{C}{W}", "Bladewheel Chariot", fuel)
	a := pushCatalogPermanent(g, me.ID, "Rock A", "Artifact", "", false)
	b := pushCatalogPermanent(g, me.ID, "Rock B", "Artifact", "", false)
	if err := g.ActivateCatalogAbility(me.ID, chariot.InstanceID, 0, game.ActivateAbilityParams{
		Strict: true, TapIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("tap two artifacts: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !containsString(effectiveTypes(t, g, chariot.InstanceID), "Creature") {
		t.Error("Bladewheel Chariot did not become a creature")
	}
}
