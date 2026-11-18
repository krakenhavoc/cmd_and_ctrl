package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82AfrDevotedPaladinOracle   = "00bcbeb5-196a-4a54-ac8d-dec708e68d01"
	s82AfrAirCultOracle          = "6ec2d957-c6d9-469f-8190-8b621c490238"
	s82AfrElturgardRangerOracle  = "27083627-a5d0-4fc6-9e66-fa44803ef080"
	s82AfrHoardRobberOracle      = "1a04f70c-288d-4010-a6d5-24e9778b9b45"
	s82AfrGuildThiefOracle       = "bf061d7e-46eb-4037-968c-3cfc5ca94552"
	s82AfrHalfElfMonkOracle      = "00d11662-9d10-4915-b80c-c1ad54b4e71b"
	s82AfrCleverConjurerOracle   = "628e59e2-3665-40dd-ba4f-11ab5824de22"
	s82AfrDireWolfProwlerOracle  = "d558b465-75c6-450e-b1f2-5843ad67de71"
	s82AfrSepulcherGhoulOracle   = "04936cc9-c915-4182-b84a-3491d26be548"
	s82AfrRustMonsterOracle      = "3da90ae6-a8cc-4633-bf49-7413ce91b697"
	s82AfrNeverwinterDryadOracle = "a6e654fb-3567-4b92-ae9b-d064bba7112a"
	s82AfrScaledHerbalistOracle  = "6fd9ece7-366f-45c7-a8ea-edb4ca637a06"
	s82AfrGretchenOracle         = "cedeea30-b580-4ea7-bbc3-7358fed8ae61"
)

func s82AfrPay(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func TestS82AfrB02EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82AfrDevotedPaladinOracle:   "Devoted Paladin",
		s82AfrAirCultOracle:          "Air-Cult Elemental",
		s82AfrElturgardRangerOracle:  "Elturgard Ranger",
		s82AfrHoardRobberOracle:      "Hoard Robber",
		s82AfrGuildThiefOracle:       "Guild Thief",
		s82AfrHalfElfMonkOracle:      "Half-Elf Monk",
		s82AfrCleverConjurerOracle:   "Clever Conjurer",
		s82AfrDireWolfProwlerOracle:  "Dire Wolf Prowler",
		s82AfrSepulcherGhoulOracle:   "Sepulcher Ghoul",
		s82AfrRustMonsterOracle:      "Rust Monster",
		s82AfrNeverwinterDryadOracle: "Neverwinter Dryad",
		s82AfrScaledHerbalistOracle:  "Scaled Herbalist",
		s82AfrGretchenOracle:         "Gretchen Titchwillow",
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

func TestS82AfrB02DevotedPaladinPumpsYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	paladin := castAndResolveCreature(t, g, "Devoted Paladin", "Creature — Orc Knight", s82AfrDevotedPaladinOracle)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my Bear is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("my Bear did not gain vigilance")
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("their Ogre is %d/%d, want 3/3", p, tough)
	}
	if hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("their Ogre gained vigilance")
	}
	if !hasEffectiveKeyword(t, g, paladin, "vigilance") {
		t.Error("the Paladin is among your creatures and should gain vigilance")
	}
}

func TestS82AfrB02AirCultElementalBouncesAnotherCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	elemental := castAndResolveCreature(t, g, "Air-Cult Elemental", "Creature — Elemental", s82AfrAirCultOracle)
	if !hasEffectiveKeyword(t, g, elemental, "flying") {
		t.Error("no flying")
	}
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: elemental}); err == nil {
		t.Error("the Elemental itself was a legal target for \"other\"")
	}
	answerPickTargetIfOpen(t, g, ogre)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || !opp.Hand.Contains(ogre) {
		t.Error("the Ogre was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(elemental) {
		t.Error("the Elemental left")
	}
}

func TestS82AfrB02ElturgardRangerMakesAWolf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ranger := castAndResolveCreature(t, g, "Elturgard Ranger", "Creature — Human Elf Ranger", s82AfrElturgardRangerOracle)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Wolf", me.ID); n != 1 {
		t.Errorf("%d Wolf tokens, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, ranger, "reach") {
		t.Error("no reach")
	}
}

func TestS82AfrB02HoardRobberMakesATreasureOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robber := b12Push(g, me.ID, "Hoard Robber", "Creature — Tiefling Rogue", s82AfrHoardRobberOracle, 1, 3)
	dealCombatDamageToPlayer(g, robber, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasures after combat damage, want 1", n)
	}
	dealNoncombatDamage(g, me.ID, robber, opp.ID)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasures after noncombat damage, want still 1", n)
	}
}

func TestS82AfrB02GuildThiefGrowsAndCanBecomeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thief := b12Push(g, me.ID, "Guild Thief", "Creature — Orc Rogue", s82AfrGuildThiefOracle, 1, 1)
	dealCombatDamageToPlayer(g, thief, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, thief, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {3}{U} with an empty pool")
	}
	s82AfrPay(me, "U", "C", "C", "C")
	b16Activate(t, g, me.ID, thief, 0, game.ActivateAbilityParams{Strict: true})
	assertRestrictions(t, g, thief, game.CantBeBlocked)
}

func TestS82AfrB02HalfElfMonkTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monk := b12Push(g, me.ID, "Half-Elf Monk", "Creature — Human Elf Monk", s82AfrHalfElfMonkOracle, 1, 4)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, monk, "vigilance") {
		t.Error("no vigilance")
	}
	if err := g.ActivateCatalogAbility(me.ID, monk, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(ogre)}); err == nil {
		t.Fatal("activated {1}{W},{T} with an empty pool")
	}
	s82AfrPay(me, "W", "C")
	b16Activate(t, g, me.ID, monk, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(ogre)})
	if !b16Tapped(t, g, ogre) {
		t.Error("the Ogre was not tapped")
	}
	if !b16Tapped(t, g, monk) {
		t.Error("the {T} cost was not paid")
	}
}

func TestS82AfrB02CleverConjurerUntapsAnythingButAConjurerAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	conj := b12Push(g, me.ID, "Clever Conjurer", "Creature — Gnome Wizard", s82AfrCleverConjurerOracle, 2, 3)
	other := b12Push(g, me.ID, "Clever Conjurer", "Creature — Gnome Wizard", s82AfrCleverConjurerOracle, 2, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, conj, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(other)}); err == nil {
		t.Fatal("a permanent named Clever Conjurer was a legal target")
	}
	b16Activate(t, g, me.ID, conj, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(rock)})
	if b16Tapped(t, g, rock) {
		t.Error("the Rock stayed tapped")
	}
	if !b16Tapped(t, g, conj) {
		t.Error("the {T} cost was not paid")
	}
	// Sorcery speed: not in the beginning of combat step.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	c2 := b12Push(g2, me2.ID, "Clever Conjurer", "Creature — Gnome Wizard", s82AfrCleverConjurerOracle, 2, 3)
	r2 := b12Permanent(g2, me2.ID, "Rock", "Artifact")
	advanceToMain(t, g2)
	for g2.Turn.Step != game.StepBeginCombat {
		if _, err := g2.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g2.ActivateCatalogAbility(me2.ID, c2, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(r2)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestS82AfrB02DireWolfProwlerPumpsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wolf := b12Push(g, me.ID, "Dire Wolf Prowler", "Creature — Wolf", s82AfrDireWolfProwlerOracle, 2, 2)
	advanceToMain(t, g)
	s82AfrPay(me, "G", "G", "G", "G")
	b16Activate(t, g, me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true})
	if p, tough := effectivePTOf(t, g, wolf); p != 4 || tough != 4 {
		t.Errorf("the Prowler is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, wolf, "haste") {
		t.Error("no haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation in the same turn was allowed")
	}
}

func TestS82AfrB02SepulcherGhoulSacrificesAnotherCreatureOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ghoul := b12Push(g, me.ID, "Sepulcher Ghoul", "Creature — Zombie", s82AfrSepulcherGhoulOracle, 2, 1)
	a := b12Creature(g, me.ID, "Fodder A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "Fodder B", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{ghoul}}); err == nil {
		t.Fatal("the Ghoul sacrificed itself for \"another creature\"")
	}
	b16Activate(t, g, me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a}})
	if g.Battlefield.Contains(a) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if p, tough := effectivePTOf(t, g, ghoul); p != 4 || tough != 3 {
		t.Errorf("the Ghoul is %d/%d, want 4/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{b}}); err == nil {
		t.Error("a second activation in the same turn was allowed")
	}
}

func TestS82AfrB02RustMonsterEatsArtifactsForPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	monster := b12Push(g, me.ID, "Rust Monster", "Creature — Beast", s82AfrRustMonsterOracle, 2, 1)
	first := b12Permanent(g, me.ID, "Rock", "Artifact")
	second := b12Permanent(g, me.ID, "Stone", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, monster, "first strike") {
		t.Error("no first strike")
	}
	if err := g.ActivateCatalogAbility(me.ID, monster, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a nonartifact creature")
	}
	b16Activate(t, g, me.ID, monster, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{first}})
	b16Activate(t, g, me.ID, monster, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{second}})
	if p, tough := effectivePTOf(t, g, monster); p != 6 || tough != 1 {
		t.Errorf("the Monster is %d/%d, want 6/1", p, tough)
	}
}

func TestS82AfrB02NeverwinterDryadFetchesATappedForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dryad := b12Push(g, me.ID, "Neverwinter Dryad", "Creature — Dryad", s82AfrNeverwinterDryadOracle, 1, 1)
	seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Forest", "Basic Land — Forest"))
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dryad, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {2} with an empty pool")
	}
	s82AfrPay(me, "C", "C")
	b16Activate(t, g, me.ID, dryad, 0, game.ActivateAbilityParams{Strict: true})
	takeSearchPrompts(t, g)
	if g.Battlefield.Contains(dryad) {
		t.Error("the Dryad was not sacrificed")
	}
	card, ok := searchFetchedCard(g, "Forest")
	if !ok {
		t.Fatal("no Forest reached the battlefield")
	}
	if !card.Tapped {
		t.Error("the Forest entered untapped")
	}
	if _, ok := searchFetchedCard(g, "Island"); ok {
		t.Error("an Island was fetched")
	}
}

func TestS82AfrB02ScaledHerbalistPutsALandFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	herb := b12Push(g, me.ID, "Scaled Herbalist", "Creature — Lizard Druid", s82AfrScaledHerbalistOracle, 1, 3)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	rock := handCardForTest(me, "Sol Ring", "Artifact", "")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, herb, 0, game.ActivateAbilityParams{Strict: true})
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land-drop prompt")
	}
	if hasID(pick.ChooseCards, rock) || !hasID(pick.ChooseCards, land) {
		t.Errorf("candidates %v: want only the land", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not reach the battlefield")
	}
	if !b16Tapped(t, g, herb) {
		t.Error("the {T} cost was not paid")
	}
}

func TestS82AfrB02GretchenDrawsThenMayPutTheDrawnLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gretchen := b12Push(g, me.ID, "Gretchen Titchwillow", "Legendary Creature — Halfling Druid", s82AfrGretchenOracle, 0, 4)
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	s82AfrPay(me, "C", "C", "G", "U")
	b16Activate(t, g, me.ID, gretchen, 0, game.ActivateAbilityParams{Strict: true})
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) == 0 {
		t.Fatal("the drawn land was not offered")
	}
	answerChooseCards(t, g, me.ID, pick.ChooseCards[0])
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d after putting the land", me.Hand.Size(), hand)
	}
}
