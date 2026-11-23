package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bl"
	cardsch "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ch"
	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	cardsso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/so"
	cardsw "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/w"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func seedCraftLibrary(p *game.Player, n int) {
	names := make([]string, n)
	for i := range names {
		names[i] = "Library Card"
	}
	effectstest.SeedLibrary(p, names...)
}

// Visage of Dread takes two creatures from both zones at once, and Dread
// Osseosaur's "you may mill two" fires on its entry.
func TestVisageOfDreadCraftsWithTwoCreaturesFromBothZones(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	visage := effectstest.PushCraftCard(g, me, visageOfDreadRow())
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	seedCraftLibrary(me, 4)

	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{B}")
	if err := effectstest.ActivateCraft(g, me, visage, 0, bear); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("one creature for \"craft with two creatures\": err = %v, want ErrInvalidParam", err)
	}
	if err := effectstest.ActivateCraft(g, me, visage, 0, bear, dead); err != nil {
		t.Fatalf("craft with two creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	libBefore := me.Library.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	osseosaur := effectstest.CraftedPermanent(t, g, "Dread Osseosaur")
	if !effectstest.EffectiveAbilitiesContain(t, g, osseosaur.InstanceID, "menace") {
		t.Error("Dread Osseosaur has no menace")
	}
	if got := effectstest.EffectivePower(t, g, osseosaur.InstanceID); got != 5 {
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
	g, me, opp := effectstest.SpendTable(t)
	blade := effectstest.PushCraftCard(g, me, effectstest.TithingBladeRow())
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.CraftInto(t, g, me, blade, 0, "{C}{C}{C}{C}{B}", "Consuming Sepulcher", dead)

	life, oppLife := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
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
	g, me, _ := effectstest.SpendTable(t)
	bricks := effectstest.PushCraftCard(g, me, effectstest.TransformRow(cardsch.ClayFiredBricksOracleID, "Clay-Fired Bricks", "Artifact", "{1}{W}",
		"Cosmium Kiln", "Artifact", "", "", []string{"W"}))
	rock := effectstest.PushGraveyardCardTyped(me, "Spent Rock", "Artifact")
	effectstest.CraftInto(t, g, me, bricks, 0, "{C}{C}{C}{C}{C}{W}{W}", "Cosmium Kiln", rock)

	gnomes := effectstest.BattlefieldIDsNamed(g, "Gnome")
	if len(gnomes) != 2 {
		t.Fatalf("%d Gnomes, want two from the Kiln's enters trigger", len(gnomes))
	}
	for _, id := range gnomes {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tough != 2 {
			t.Errorf("Gnome is %d/%d, want 2/2 under the Kiln's anthem", p, tough)
		}
	}
}

// Braided Net taps a permanent and stops its activated abilities only
// while it stays tapped; it pays a net counter, and crafts away an
// artifact.
func TestBraidedNetLocksAPermanentWhileItRemainsTapped(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	c := deck.ToGameCard(effectstest.TransformRow(cardsbl.BraidedNetOracleID, "Braided Net", "Artifact", "{2}{U}",
		"Braided Quipu", "Artifact", "", "", []string{"U"}), false)
	c.Owner, c.Controller, c.Counters = me.ID, me.ID, map[string]int{"net": 3}
	net := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	seer := effectstest.PushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard", effectstest.VisceraSeerOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)

	if err := g.ActivateCatalogAbility(me.ID, net, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: seer}},
	}); err != nil {
		t.Fatalf("activate the Net: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.BattlefieldCard(g, net); n.Counters["net"] != 2 {
		t.Errorf("net counters = %d, want one removed", n.Counters["net"])
	}
	if s, _ := effectstest.BattlefieldCard(g, seer); !s.Tapped {
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
		return effectstest.TransformRow(effects.JadeSeedstonesOracleID, "Jade Seedstones", "Artifact", "{3}{G}",
			"Jadeheart Attendant", "Artifact Creature — Golem", "7", "7", []string{"G"})
	}
	g, me, _ := effectstest.SpendTable(t)
	seed := effectstest.PushCraftCard(g, me, row())
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}{G}", Owner: me.ID, Controller: me.ID})
	life := me.Life
	effectstest.CraftInto(t, g, me, seed, 0, "{C}{C}{C}{C}{C}{G}{G}", "Jadeheart Attendant", dead)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 for a four-mana creature card", life, me.Life)
	}

	g2, me2, _ := effectstest.SpendTable(t)
	seed2 := effectstest.PushCraftCard(g2, me2, row())
	token := effectstest.PushCatalogPermanent(g2, me2.ID, "Soldier", "Token Creature — Soldier", "", false)
	life = me2.Life
	effectstest.CraftInto(t, g2, me2, seed2, 0, "{C}{C}{C}{C}{C}{G}{G}", "Jadeheart Attendant", token)
	if me2.Life != life {
		t.Errorf("life %d -> %d, want nothing for a token material", life, me2.Life)
	}
}

// Sovereign's Macuahuitl attaches itself as it enters and pumps.
func TestIdolOfTheDeepKingCraftsIntoAnEquipmentThatAttachesItself(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	idol := effectstest.PushCraftCard(g, me, effectstest.TransformRow(cardshu.IdolOfTheDeepKingOracleID, "Idol of the Deep King", "Artifact", "{2}{R}",
		"Sovereign's Macuahuitl", "Artifact — Equipment", "", "", []string{"R"}))
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Mind Stone", "Artifact", "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	if err := effectstest.ActivateCraft(g, me, idol, 0, rock); err != nil {
		t.Fatalf("craft: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	club := effectstest.CraftedPermanent(t, g, "Sovereign's Macuahuitl")
	if club.AttachedTo.ID != bear {
		t.Errorf("Macuahuitl attached to %v, want the bear", club.AttachedTo)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("bear power = %d, want 1 + 2", got)
	}
}

// Oteclan Levitator is a 1/4 flier.
func TestOteclanLandmarkCraftsIntoAFlyingLevitator(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	mark := effectstest.PushCraftCard(g, me, effectstest.TransformRow(cardso.OteclanLandmarkOracleID, "Oteclan Landmark", "Artifact", "{W}",
		"Oteclan Levitator", "Artifact Creature — Golem", "1", "4", []string{"W"}))
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Mind Stone", "Artifact", "", false)
	lev := effectstest.CraftInto(t, g, me, mark, 0, "{C}{C}{W}", "Oteclan Levitator", rock)
	if !effectstest.EffectiveAbilitiesContain(t, g, lev.InstanceID, "flying") || effectstest.EffectiveToughness(t, g, lev.InstanceID) != 4 {
		t.Error("Oteclan Levitator is not a 1/4 flier")
	}
}

// Inverted Iceberg's entry mills a card and then draws one.
func TestInvertedIcebergMillsThenDrawsOnEntry(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	seedCraftLibrary(me, 3)
	hand, grave, lib := me.Hand.Size(), me.Graveyard.Size(), me.Library.Size()
	effectstest.ImportToBattlefield(t, g, effectstest.TransformRow(cardshu.InvertedIcebergOracleID, "Inverted Iceberg", "Artifact", "{1}{U}",
		"Iceberg Titan", "Artifact Creature — Golem", "6", "6", []string{"U"}), me)
	effectstest.PassPriorityAroundTable(t, g)
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
	g, me, _ := effectstest.SpendTable(t)
	tree := effectstest.PushCraftCard(g, me, effectstest.TransformRow(cardsj.KaslemsStonetreeOracleID, "Kaslem's Stonetree", "Artifact", "{2}{G}",
		"Kaslem's Strider", "Artifact Creature — Golem", "5", "5", []string{"G"}))
	forest := effectstest.PushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	cave := effectstest.PushGraveyardCardTyped(me, "Cavernous Maw", "Land — Cave")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{G}")
	if err := effectstest.ActivateCraft(g, me, tree, 0, forest); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a Forest for \"craft with Cave\": err = %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.ActivateCraft(g, me, tree, 0, cave); err != nil {
		t.Fatalf("craft with Cave: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if s := effectstest.CraftedPermanent(t, g, "Kaslem's Strider"); effectstest.EffectivePower(t, g, s.InstanceID) != 5 {
		t.Error("Kaslem's Strider is not a 5/5")
	}
}

// Craft with Island takes an Island you control; the Hulk mills for {T}.
func TestWaterloggedHulkMillsAndCraftsWithAnIsland(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	hulk := effectstest.PushCraftCard(g, me, effectstest.TransformRow(cardsw.WaterloggedHulkOracleID, "Waterlogged Hulk", "Artifact", "{U}",
		"Watertight Gondola", "Artifact — Vehicle", "4", "4", []string{"U"}))
	seedCraftLibrary(me, 2)
	grave := me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{T}: mill a card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != grave+1 {
		t.Fatalf("graveyard %d -> %d, want one milled", grave, me.Graveyard.Size())
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(hulk) })
	island := effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	gondola := effectstest.CraftInto(t, g, me, hulk, 1, "{C}{C}{C}{U}", "Watertight Gondola", island)
	if !effectstest.EffectiveAbilitiesContain(t, g, gondola.InstanceID, "vigilance") {
		t.Error("Watertight Gondola has no vigilance")
	}
}

// Bladewheel Chariot becomes a creature by tapping two other artifacts.
func TestSpringLoadedSawbladesCraftsIntoAChariotTwoArtifactsAnimate(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	saw := effectstest.PushCraftCard(g, me, effectstest.TransformRow(cardsso.SpringLoadedSawbladesOracleID, "Spring-Loaded Sawblades", "Artifact", "{1}{W}",
		"Bladewheel Chariot", "Artifact — Vehicle", "5", "5", []string{"W"}))
	fuel := effectstest.PushGraveyardCardTyped(me, "Spent Rock", "Artifact")
	chariot := effectstest.CraftInto(t, g, me, saw, 0, "{C}{C}{C}{W}", "Bladewheel Chariot", fuel)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Rock A", "Artifact", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Rock B", "Artifact", "", false)
	if err := g.ActivateCatalogAbility(me.ID, chariot.InstanceID, 0, game.ActivateAbilityParams{
		Strict: true, TapIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("tap two artifacts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, chariot.InstanceID), "Creature") {
		t.Error("Bladewheel Chariot did not become a creature")
	}
}
