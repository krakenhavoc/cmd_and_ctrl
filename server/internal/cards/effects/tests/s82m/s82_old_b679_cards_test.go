package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b679_cards_test.go — S82 OLD batch 679 (#5618, Nemesis: activated
// abilities, enters and dies triggers, removal, instants and sorceries). Each
// test plays the real catalog entry.

const (
	oldB679BelbesArmor     = "1650d85f-4c65-429d-8952-f68b62935a55"
	oldB679FanaticalDev    = "34e56412-4649-481d-be51-357b331d0b75"
	oldB679HarvestMage     = "6f82b01d-8975-42a1-87e2-99782272bbea"
	oldB679BelbesPortal    = "46e50ee9-91f8-416e-82fa-f1a76e61fb73"
	oldB679OverlaidTerrain = "e5818ab3-8c37-4063-9e9d-cbe4ae0bba14"
	oldB679PhyrexianDriver = "e8d12d8e-afd2-438b-b121-5911fca479f9"
	oldB679Avenger         = "62009b63-ca6c-4fb0-8c9a-eccdec51cb18"
	oldB679Topple          = "75eb30cc-688d-400a-abe3-5319b79fcbae"
	oldB679ReverentSilence = "e670d1c9-5e28-4df8-aad9-d08d19dc74e9"
	oldB679PaleMoon        = "78e3a3c3-3091-46ab-a47a-86b7db68ced4"
	oldB679Gambit          = "797bb53d-b111-41fe-95d7-64cb0f833b22"
	oldB679AnimateLand     = "b2bd58ea-b2c6-47f1-b082-7969f093bbc6"
	oldB679Bear            = "Creature — Bear"
)

func TestOldB679EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB679BelbesArmor:     "Belbe's Armor",
		oldB679FanaticalDev:    "Fanatical Devotion",
		oldB679HarvestMage:     "Harvest Mage",
		oldB679BelbesPortal:    "Belbe's Portal",
		oldB679OverlaidTerrain: "Overlaid Terrain",
		oldB679PhyrexianDriver: "Phyrexian Driver",
		oldB679Avenger:         "Avenger en-Dal",
		oldB679Topple:          "Topple",
		oldB679ReverentSilence: "Reverent Silence",
		oldB679PaleMoon:        "Pale Moon",
		oldB679Gambit:          "Stronghold Gambit",
		oldB679AnimateLand:     "Animate Land",
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

func oldB679ToughnessOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	got := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		got = c.CurrentToughness()
	})
	return got
}

func oldB679Attacking(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = defender
			}
		}
	})
}

func oldB679Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Belbe's Armor ------------------------------------------------------------

func TestOldB679BelbesArmorGivesMinusXPlusX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	armor := effectstest.B12Push(g, me.ID, "Belbe's Armor", "Artifact", oldB679BelbesArmor, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB679Bear, 2, 2)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: oldB679Target(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PowerOf(t, g, bear), oldB679ToughnessOf(t, g, bear); p != -1 || tou != 5 {
		t.Errorf("the bear is %d/%d, want -1/5", p, tou)
	}
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{Strict: true, XValue: 0, Targets: oldB679Target(bear)}); err == nil {
		t.Error("a tapped Armor activated again")
	}
}

func TestOldB679BelbesArmorNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	armor := effectstest.B12Push(g, me.ID, "Belbe's Armor", "Artifact", oldB679BelbesArmor, 0, 0)
	land := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, armor, 0, game.ActivateAbilityParams{Strict: true, XValue: 1, Targets: oldB679Target(land)}); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Fanatical Devotion ---------------------------------------------------------

func TestOldB679FanaticalDevotionSacrificesACreatureToRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dev := effectstest.B12Push(g, me.ID, "Fanatical Devotion", "Enchantment", oldB679FanaticalDev, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB679Bear, 1, 1)
	saved := effectstest.B12Creature(g, me.ID, "Saved", oldB679Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, dev, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB679Target(saved)}); err == nil {
		t.Fatal("activated with no sacrifice named")
	}
	if err := g.ActivateCatalogAbility(me.ID, dev, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}, Targets: oldB679Target(saved)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	// Destroying the target now regenerates it.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(saved); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Battlefield.Contains(saved) {
		t.Error("the regeneration shield did not save the creature")
	}
}

// --- Harvest Mage -----------------------------------------------------------------

func TestOldB679HarvestMageMakesOneManaOfAChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mage := effectstest.B12Push(g, me.ID, "Harvest Mage", "Creature — Human Spellshaper", oldB679HarvestMage, 1, 1)
	pitch := effectstest.PushHandCardOfType(me, "Pitched", "Sorcery")
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandHasID(me, pitch) {
		t.Error("the card was not discarded")
	}
	swamp := effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateManaAbility(me.ID, swamp, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tapping the Swamp for red: %v", err)
	}
	if got := effectstest.PoolColours(me); got["R"] != 1 || len(got) != 1 {
		t.Errorf("pool %v, want one {R}", got)
	}
}

func TestOldB679HarvestMageNeedsADiscardAndGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mage := effectstest.B12Push(g, me.ID, "Harvest Mage", "Creature — Human Spellshaper", oldB679HarvestMage, 1, 1)
	pitch := effectstest.PushHandCardOfType(me, "Pitched", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err == nil {
		t.Error("activated without {G}")
	}
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without discarding a card")
	}
}

// --- Pale Moon -------------------------------------------------------------------

func TestOldB679PaleMoonMakesNonbasicLandsProduceColourless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastCatalogSpell(t, g, "Pale Moon", "Instant", oldB679PaleMoon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	nonbasic := effectstest.FreeSpellPermanent(g, me.ID, "Test Forest", "Land — Forest")
	basic := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.FreeSpellPermanent(g, opp.ID, "Test Mountain", "Land — Mountain")
	if err := g.ActivateManaAbility(me.ID, nonbasic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the nonbasic land: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, basic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the basic land: %v", err)
	}
	if got := effectstest.PoolColours(me); got["C"] != 1 || got["G"] != 1 || len(got) != 2 {
		t.Errorf("pool %v, want one {C} from the nonbasic land and one {G} from the basic one", got)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the opponent's nonbasic land: %v", err)
	}
	if got := effectstest.PoolColours(opp); got["C"] != 1 || len(got) != 1 {
		t.Errorf("opponent pool %v, want one {C}", got)
	}
}

func TestOldB679PaleMoonEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Pale Moon", "Instant", oldB679PaleMoon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	nonbasic := effectstest.FreeSpellPermanent(g, me.ID, "Test Forest", "Land — Forest")
	if err := g.ActivateManaAbility(me.ID, nonbasic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping: %v", err)
	}
	if got := effectstest.PoolColours(me); got["G"] != 1 {
		t.Errorf("pool %v after the turn ended, want {G}", got)
	}
}

// --- Animate Land --------------------------------------------------------------

func TestOldB679AnimateLandMakesATargetLandAThreeThreeCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Animate Land", "Instant", oldB679AnimateLand, oldB679Target(land))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsCreatureNow(g, land) {
		t.Fatal("the land is not a creature")
	}
	if p, tou := effectstest.PowerOf(t, g, land), oldB679ToughnessOf(t, g, land); p != 3 || tou != 3 {
		t.Errorf("the land is %d/%d, want 3/3", p, tou)
	}
	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(land)
		if !c.IsLand() {
			t.Error("the creature land stopped being a land")
		}
	})
}

func TestOldB679AnimateLandTargetsOnlyALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB679Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Animate Land", "Instant", oldB679AnimateLand, oldB679Target(bear)); err == nil {
		t.Error("a nonland creature was a legal target")
	}
}

// --- Phyrexian Driver ------------------------------------------------------------

func TestOldB679PhyrexianDriverPumpsEveryOtherMercenary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merc := effectstest.B12Creature(g, me.ID, "Merc", "Creature — Human Mercenary", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Merc", "Creature — Human Mercenary", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB679Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Phyrexian Driver", "Creature — Phyrexian Zombie Mercenary", oldB679PhyrexianDriver, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	driver := effectstest.FindBattlefieldByName(g, "Phyrexian Driver")
	if driver == uuid.Nil {
		t.Fatal("the Driver did not enter")
	}
	for id, want := range map[uuid.UUID]int{merc: 3, theirs: 3, bear: 2} {
		if got := effectstest.PowerOf(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if tou := oldB679ToughnessOf(t, g, merc); tou != 3 {
		t.Errorf("Mercenary toughness %d, want 3", tou)
	}
}

// --- Topple ----------------------------------------------------------------------

func TestOldB679ToppleExilesOnlyAGreatestPowerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, me.ID, "Small", oldB679Bear, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB679Bear, 5, 5)
	tied := effectstest.B12Creature(g, me.ID, "Tied", oldB679Bear, 5, 1)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Topple", "Sorcery", oldB679Topple, oldB679Target(small)); err == nil {
		t.Fatal("a creature without the greatest power was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Topple", "Sorcery", oldB679Topple, oldB679Target(tied))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tied) {
		t.Error("the tied creature was not exiled")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) {
		t.Error("another creature was exiled")
	}
	found := false
	for _, c := range g.Exile.Cards {
		if c.Name == "Tied" {
			found = true
		}
	}
	if !found {
		t.Error("the creature went somewhere other than exile")
	}
}

func TestOldB679ToppleFizzlesWhenAPumpTakesTheTargetOutOfTheSet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, me.ID, "Target", oldB679Bear, 5, 5)
	effectstest.B12Creature(g, opp.ID, "Other", oldB679Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Topple", "Sorcery", oldB679Topple, oldB679Target(target))
	effectstest.B12Creature(g, opp.ID, "Late Giant", oldB679Bear, 9, 9)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(target) {
		t.Error("a target that no longer has the greatest power was exiled")
	}
}

// --- Avenger en-Dal ---------------------------------------------------------------

func TestOldB679AvengerExilesAnAttackerAndItsControllerGainsItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	avenger := effectstest.B12Push(g, me.ID, "Avenger en-Dal", "Creature — Human Spellshaper", oldB679Avenger, 1, 1)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", oldB679Bear, 3, 4)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", oldB679Bear, 3, 4)
	oldB679Attacking(g, attacker, me.ID)
	pitch := effectstest.PushHandCardOfType(me, "Pitched", "Sorcery")
	effectstest.FloatForTest(g, me, "WWW")
	if err := g.ActivateCatalogAbility(me.ID, avenger, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}, Targets: oldB679Target(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was a legal target")
	}
	before, mine := opp.Life, me.Life
	if err := g.ActivateCatalogAbility(me.ID, avenger, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}, Targets: oldB679Target(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacker was not exiled")
	}
	if opp.Life != before+4 || me.Life != mine {
		t.Errorf("life: its controller %d (was %d), the Avenger's controller %d (was %d); want +4 for its controller only", opp.Life, before, me.Life, mine)
	}
}

// --- Belbe's Portal ---------------------------------------------------------------

func oldB679Portal(t *testing.T, g *game.Game, me *game.Player, tribe string) uuid.UUID {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Belbe's Portal", "Artifact", oldB679BelbesPortal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, tribe)
	portal := effectstest.FindBattlefieldByName(g, "Belbe's Portal")
	if portal == uuid.Nil {
		t.Fatal("the Portal did not enter")
	}
	return portal
}

func TestOldB679BelbesPortalPutsACreatureOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	portal := oldB679Portal(t, g, me, "Elf")
	elf := effectstest.PushHandCardOfType(me, "Elf Scout", "Creature — Elf Scout")
	human := effectstest.PushHandCardOfType(me, "Human Guard", "Creature — Human Soldier")
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to put a creature card")
	}
	for _, c := range pick.ChooseCards {
		if c == human {
			t.Error("a creature of another type was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, elf)
	if !g.Battlefield.Contains(elf) {
		t.Error("the Elf did not enter the battlefield")
	}
	if g.Battlefield.Contains(human) {
		t.Error("the Human entered")
	}
}

func TestOldB679BelbesPortalDoesNothingWithoutAMatchingCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	portal := oldB679Portal(t, g, me, "Elf")
	human := effectstest.PushHandCardOfType(me, "Human Guard", "Creature — Human Soldier")
	land := effectstest.PushHandCardOfType(me, "Elven Land", "Land — Elf")
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("prompted with no creature card of the type in hand")
	}
	if g.Battlefield.Contains(human) || g.Battlefield.Contains(land) {
		t.Error("a card of the wrong type or kind entered")
	}
}

func TestOldB679BelbesPortalMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	portal := oldB679Portal(t, g, me, "Elf")
	elf := effectstest.PushHandCardOfType(me, "Elf Scout", "Creature — Elf Scout")
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(elf) {
		t.Error("declining still put the Elf onto the battlefield")
	}
}

// --- Overlaid Terrain ------------------------------------------------------------

func TestOldB679OverlaidTerrainSacrificesYourLandsAndGivesTwoManaOfOneColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine1 := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	mine2 := effectstest.FreeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.FreeSpellPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Overlaid Terrain", "Enchantment", oldB679OverlaidTerrain, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine1) || g.Battlefield.Contains(mine2) {
		t.Error("a land its controller controlled was not sacrificed")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's land was sacrificed")
	}
	if effectstest.FindBattlefieldByName(g, "Overlaid Terrain") == uuid.Nil {
		t.Fatal("Overlaid Terrain is not on the battlefield")
	}
	// A land that arrives afterwards has the grant.
	fresh := effectstest.B12Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tapping the new land for the granted ability: %v", err)
	}
	if got := effectstest.PoolColours(me); got["U"] != 2 || len(got) != 1 {
		t.Errorf("pool %v, want {U}{U}", got)
	}
}

// --- Reverent Silence -------------------------------------------------------------

func TestOldB679ReverentSilenceDestroysEveryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "My Aura", "Enchantment — Aura", "", 0, 0)
	theirs := effectstest.B12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB679Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Reverent Silence", "Sorcery", oldB679ReverentSilence, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

func TestOldB679ReverentSilenceAlternativeCostGivesEachOtherPlayerSixLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	spell := effectstest.HandCardFull(me, "Reverent Silence", "Sorcery", "{3}{G}", oldB679ReverentSilence, []string{"G"})
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "reverent-silence"}); err != nil {
		t.Fatalf("cast for the alternative cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] + 6
		if p.ID == me.ID {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
}

func TestOldB679ReverentSilenceAlternativeCostNeedsAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.FreeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	spell := effectstest.HandCardFull(me, "Reverent Silence", "Sorcery", "{3}{G}", oldB679ReverentSilence, []string{"G"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "reverent-silence"}); err == nil {
		t.Error("the alternative cost was offered without a Forest")
	}
}

// --- Stronghold Gambit -------------------------------------------------------------

func oldB679GambitTable(t *testing.T, hands ...[]game.Card) (*game.Game, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	var ids []uuid.UUID
	for i, p := range g.Seats {
		p.Hand.Cards = nil
		if i < len(hands) {
			for _, c := range hands[i] {
				ids = append(ids, effectstest.HandCardFull(p, c.Name, c.TypeLine, c.ManaCost, "", nil))
			}
		}
	}
	return g, ids
}

func TestOldB679StrongholdGambitPutsTheLowestCreatureCardsOntoTheBattlefield(t *testing.T) {
	g, ids := oldB679GambitTable(t,
		[]game.Card{{Name: "Cheap Bear", TypeLine: oldB679Bear, ManaCost: "{1}{G}"}},
		[]game.Card{{Name: "Pricey Ogre", TypeLine: "Creature — Ogre", ManaCost: "{4}{R}"}},
		[]game.Card{{Name: "Other Cheap Bear", TypeLine: oldB679Bear, ManaCost: "{1}{G}"}},
		[]game.Card{{Name: "Cheap Sorcery", TypeLine: "Sorcery", ManaCost: "{R}"}},
	)
	cheap, ogre, tied, spell := ids[0], ids[1], ids[2], ids[3]
	effectstest.CastCatalogSpell(t, g, "Stronghold Gambit", "Sorcery", oldB679Gambit, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range []uuid.UUID{cheap, ogre, tied, spell} {
		effectstest.AnswerChooseCards(t, g, g.Seats[i].ID, id)
	}
	if !g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(tied) {
		t.Error("a creature card of the lowest mana value did not enter")
	}
	if g.Battlefield.Contains(ogre) || g.Battlefield.Contains(spell) {
		t.Error("a costlier creature or a noncreature card entered")
	}
	for i, id := range []uuid.UUID{cheap, ogre, tied, spell} {
		if want := i != 0 && i != 2; effectstest.HandHasID(g.Seats[i], id) != want {
			t.Errorf("seat %d: card in hand = %v, want %v", i, !want, want)
		}
	}
	// Each enters under its owner's control.
	for i, id := range []uuid.UUID{cheap, tied} {
		owner := g.Seats[i*2].ID
		g.WithWriteLock(func() {
			if c, ok := g.LookupCardForEffect(id); !ok || c.Controller != owner {
				t.Errorf("card %d is not under its owner's control", i)
			}
		})
	}
}

func TestOldB679StrongholdGambitANoncreatureLowestMeansNoCreatureMissesOut(t *testing.T) {
	g, ids := oldB679GambitTable(t,
		[]game.Card{{Name: "Pricey Ogre", TypeLine: "Creature — Ogre", ManaCost: "{4}{R}"}},
		[]game.Card{{Name: "Cheap Sorcery", TypeLine: "Sorcery", ManaCost: "{R}"}},
	)
	effectstest.CastCatalogSpell(t, g, "Stronghold Gambit", "Sorcery", oldB679Gambit, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, g.Seats[0].ID, ids[0])
	effectstest.AnswerChooseCards(t, g, g.Seats[1].ID, ids[1])
	if !g.Battlefield.Contains(ids[0]) {
		t.Error("the only creature card revealed did not enter")
	}
	if g.Battlefield.Contains(ids[1]) {
		t.Error("a noncreature card entered")
	}
}

func TestOldB679StrongholdGambitWithNoCardsInAnyHandDoesNothing(t *testing.T) {
	g, _ := oldB679GambitTable(t)
	effectstest.CastCatalogSpell(t, g, "Stronghold Gambit", "Sorcery", oldB679Gambit, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, g.Seats[0].ID) != nil {
		t.Error("a player with an empty hand was asked to choose")
	}
}
