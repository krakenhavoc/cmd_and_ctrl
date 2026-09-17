package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_n_cards_test.go — the cards #747's sacrifice-N cost
// enabled: five caveats removed (Savvy Hunter, Samwise Gamgee, Sai,
// both Magdas) and eight cards added. The engine rules are pinned in
// sacrifice_n_cost_test.go; these assert each card pays its printed
// count and does what it prints.

const (
	priestOfForgottenGodsOracle = "2ad8ff62-d090-4835-9274-3b755ba0f8e6"
	kuldothaForgemasterOracle   = "b0f99367-4313-4bb1-a9fd-7711bc4ce40e"
	tamiyosJournalOracle        = "20591a83-7138-4876-b088-035bd3788be3"
	teysaOrzhovScionOracle      = "8191342b-b25e-4c4d-8f69-aee662148ff4"
	hedronDetonatorOracle       = "9221b29b-9351-4f70-9af4-8b9a33b43138"
	oliviaOpulentOutlawOracle   = "4e5e0b22-ee5a-4c19-be95-2f9bd50641bc"
	breyaEtheriumShaperOracle   = "d460a9e2-5a7d-4562-880e-45174be19a9d"
	whisperBloodLiturgistOracle = "496ab82e-24a9-4a74-ba4b-992c7309b44a"
)

func TestSacrificeNCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		priestOfForgottenGodsOracle: "Priest of Forgotten Gods",
		kuldothaForgemasterOracle:   "Kuldotha Forgemaster",
		tamiyosJournalOracle:        "Tamiyo's Journal",
		teysaOrzhovScionOracle:      "Teysa, Orzhov Scion",
		hedronDetonatorOracle:       "Hedron Detonator",
		oliviaOpulentOutlawOracle:   "Olivia, Opulent Outlaw",
		breyaEtheriumShaperOracle:   "Breya, Etherium Shaper",
		whisperBloodLiturgistOracle: "Whisper, Blood Liturgist",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == CompletenessUnreviewed {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
	// Transmutation Font keeps its caveat: "with different names" is a
	// set-level restriction #747 left out of scope.
	if spec, _ := Lookup(b32TransmutationFontOracle); spec.Completeness != CompletenessCaveats || len(spec.Activated) != 3 {
		t.Error("Transmutation Font's tutor stays a declared omission")
	}
}

func pushTokens(g *game.Game, controller uuid.UUID, tmpl game.Card, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = pushToken(g, controller, tmpl)
	}
	return ids
}

// --- the five caveats removed --------------------------------------

func TestSavvyHunterSacrificesTwoFoodsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hunter := b12Push(g, me.ID, "Savvy Hunter", "Creature — Human Warrior", b30SavvyHunterOracle, 3, 3)
	foods := pushTokens(g, me.ID, FoodToken(), 2)
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: foods[:1]}); err != game.ErrInvalidParam {
		t.Fatalf("one Food: got %v, want ErrInvalidParam", err)
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: foods})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if countBattlefieldNamed(g, me.ID, "Food") != 0 {
		t.Error("both Foods are sacrificed")
	}
}

func TestSamwiseGamgeeSacrificesThreeFoodsToReturnAHistoricCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sam := b12Push(g, me.ID, "Samwise Gamgee", "Legendary Creature — Halfling Peasant", b22SamwiseGamgeeOracle, 2, 2)
	relic := pushGraveyardCardTyped(me, "Relic", "Artifact")
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	foods := pushTokens(g, me.ID, FoodToken(), 3)

	if err := g.ActivateCatalogAbility(me.ID, sam, 0, game.ActivateAbilityParams{
		SacrificeIDs: foods, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a nonhistoric creature card is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, sam, 0, game.ActivateAbilityParams{
		SacrificeIDs: foods[:2], Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err != game.ErrInvalidParam {
		t.Fatalf("two Foods: got %v, want ErrInvalidParam", err)
	}
	b16Activate(t, g, me.ID, sam, 0, game.ActivateAbilityParams{
		SacrificeIDs: foods, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	})
	if !me.Hand.Contains(relic) {
		t.Error("the historic card returns to hand")
	}
	if countBattlefieldNamed(g, me.ID, "Food") != 0 {
		t.Error("all three Foods are sacrificed")
	}
}

func TestSaiSacrificesTwoArtifactsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sai := pushCatalogPermanent(g, me.ID, "Sai, Master Thopterist", "Legendary Creature — Human Artificer", b07SaiMasterThopteristOracle, false)
	thopter := TokenCard("1/1 colorless Thopter artifact with flying")
	arts := pushTokens(g, me.ID, thopter, 2)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, sai, 0, game.ActivateAbilityParams{SacrificeIDs: arts})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if countBattlefieldNamed(g, me.ID, "Thopter") != 0 {
		t.Error("both artifacts are sacrificed")
	}
	// Sai is not an artifact, so she can never pay for herself.
	more := pushTokens(g, me.ID, thopter, 1)
	if err := g.ActivateCatalogAbility(me.ID, sai, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{more[0], sai}}); err != game.ErrIllegalTarget {
		t.Errorf("Sai as an artifact: got %v, want ErrIllegalTarget", err)
	}
}

func TestMagdaTheHoardmasterSacrificesThreeTreasuresForADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magda := b12Push(g, me.ID, "Magda, the Hoardmaster", "Legendary Creature — Dwarf Berserker", b18MagdaTheHoardmasterOrcl, 2, 2)
	treasures := pushTokens(g, me.ID, TreasureToken(), 3)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, magda, 0, game.ActivateAbilityParams{SacrificeIDs: treasures})
	dragon := findBattlefieldByName(g, "Scorpion Dragon")
	if dragon == uuid.Nil {
		t.Fatal("no Scorpion Dragon")
	}
	card, _ := battlefieldCard(g, dragon)
	if card.Power != 4 || card.Toughness != 4 || !card.HasColor("R") {
		t.Errorf("the Dragon is a 4/4 red: %+v", card)
	}
	abilities := effectiveAbilities(t, g, dragon)
	if !eotHasAbility(abilities, "flying") || !eotHasAbility(abilities, "haste") {
		t.Errorf("the Dragon has flying and haste: %v", abilities)
	}
	if countBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("all three Treasures are sacrificed")
	}
}

func TestMagdaBrazenOutlawSacrificesFiveTreasuresToTutorOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magda := b11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", b11MagdaOracle, 2, 1)
	pushLibraryCardForTest(me, game.Card{Name: "Wurmcoil", TypeLine: "Artifact Creature — Phyrexian Wurm"})
	treasures := pushTokens(g, me.ID, TreasureToken(), 5)
	if err := g.ActivateCatalogAbility(me.ID, magda, 0, game.ActivateAbilityParams{SacrificeIDs: treasures[:4]}); err != game.ErrInvalidParam {
		t.Fatalf("four Treasures: got %v, want ErrInvalidParam", err)
	}
	b16Activate(t, g, me.ID, magda, 0, game.ActivateAbilityParams{SacrificeIDs: treasures})
	if findBattlefieldByName(g, "Wurmcoil") == uuid.Nil {
		t.Error("the artifact is put onto the battlefield")
	}
	if countBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("all five Treasures are sacrificed")
	}
}

// --- the cards added -----------------------------------------------

func TestPriestOfForgottenGodsEdictsDrainsAddsManaAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	priest := b12Push(g, me.ID, "Priest of Forgotten Gods", "Creature — Human Cleric", priestOfForgottenGodsOracle, 1, 2)
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")
	victim := sacNCreature(g, opp.ID, "Their Bear")
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	// "Two OTHER creatures": the Priest is not one of them.
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, priest}, Targets: target,
	}); err != game.ErrIllegalTarget {
		t.Fatalf("the Priest itself: got %v, want ErrIllegalTarget", err)
	}
	life, hand := opp.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b}, Targets: target,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("target lost %d life, want 2", life-opp.Life)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := poolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want B B", got)
	}
	answerSacrifice(t, g, opp.ID, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the target player sacrifices a creature")
	}
	if !b16Tapped(t, g, priest) {
		t.Error("the Priest taps")
	}
}

func TestKuldothaForgemasterMayBeOneOfItsOwnThreeArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forge := pushCatalogPermanent(g, me.ID, "Kuldotha Forgemaster", "Artifact Creature — Construct", kuldothaForgemasterOracle, true)
	arts := pushTokens(g, me.ID, TreasureToken(), 2)
	pushLibraryCardForTest(me, game.Card{Name: "Blightsteel", TypeLine: "Legendary Artifact Creature — Phyrexian Golem"})
	pay := append(append([]uuid.UUID(nil), arts...), forge)

	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{SacrificeIDs: pay}); err != game.ErrSummoningSick {
		t.Fatalf("summoning sick: got %v, want ErrSummoningSick", err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == forge {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	b16Activate(t, g, me.ID, forge, 0, game.ActivateAbilityParams{SacrificeIDs: pay})
	if g.Battlefield.Contains(forge) {
		t.Error("the Forgemaster paid for itself")
	}
	if findBattlefieldByName(g, "Blightsteel") == uuid.Nil {
		t.Error("the artifact is put onto the battlefield")
	}
}

func TestTamiyosJournalInvestigatesAndTutorsForThreeClues(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	journal := pushCatalogPermanent(g, me.ID, "Tamiyo's Journal", "Legendary Artifact — Book", tamiyosJournalOracle, false)
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Clue") != 1 {
		t.Fatalf("your upkeep investigates: %d Clues", countBattlefieldNamed(g, me.ID, "Clue"))
	}
	clues := append(pushTokens(g, me.ID, ClueToken(), 2), findBattlefieldByName(g, "Clue"))
	pushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Sorcery"})
	if err := g.ActivateCatalogAbility(me.ID, journal, 0, game.ActivateAbilityParams{SacrificeIDs: clues}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Needle")
	if countBattlefieldNamed(g, me.ID, "Clue") != 0 {
		t.Error("all three Clues are sacrificed")
	}
	found := false
	for _, c := range me.Hand.Cards {
		found = found || c.Name == "Needle"
	}
	if !found {
		t.Error("the searched card goes to hand")
	}
}

func TestTeysaExilesForThreeWhiteCreaturesAndMakesSpiritsForBlackDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teysa := b12Push(g, me.ID, "Teysa, Orzhov Scion", "Legendary Creature — Human Advisor", teysaOrzhovScionOracle, 2, 3)
	w1 := b16Creature(g, me.ID, "White A", "Creature — Soldier", 1, 1, "W")
	w2 := b16Creature(g, me.ID, "White B", "Creature — Soldier", 1, 1, "W")
	black := b16Creature(g, me.ID, "Black", "Creature — Zombie", 1, 1, "B")
	w3 := b16Creature(g, me.ID, "White C", "Creature — Soldier", 1, 1, "W")
	threat := b16Creature(g, opp.ID, "Threat", "Creature — Dragon", 6, 6, "R")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: threat}}

	if err := g.ActivateCatalogAbility(me.ID, teysa, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{w1, w2, black}, Targets: target,
	}); err != game.ErrIllegalTarget {
		t.Fatalf("a black creature among the three: got %v, want ErrIllegalTarget", err)
	}
	b16Activate(t, g, me.ID, teysa, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{w1, w2, w3}, Targets: target})
	if g.Battlefield.Contains(threat) || !containsCard(g.Exile, threat) {
		t.Error("the target creature is exiled")
	}
	if countBattlefieldNamed(g, me.ID, "Spirit") != 0 {
		t.Error("white creatures dying make no Spirit")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(black) })
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Spirit") != 1 {
		t.Errorf("a black creature dying makes a Spirit: %d", countBattlefieldNamed(g, me.ID, "Spirit"))
	}
}

func containsCard(z *game.Zone, id uuid.UUID) bool {
	return z != nil && z.Contains(id)
}

func TestHedronDetonatorPingsOnArtifactsAndImpulsesForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	det := b12Push(g, me.ID, "Hedron Detonator", "Creature — Goblin Artificer", hedronDetonatorOracle, 2, 3)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("an artifact entering pings the target opponent: %d -> %d", life, opp.Life)
	}
	arts := append(pushTokens(g, me.ID, TreasureToken(), 1), findBattlefieldByName(g, "Treasure"))
	top := pushTopLibraryCard(me, "Impulse")
	b16Activate(t, g, me.ID, det, 0, game.ActivateAbilityParams{SacrificeIDs: arts})
	if !containsCard(g.Exile, top) {
		t.Error("the top card of the library is exiled")
	}
	if countBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("both artifacts are sacrificed")
	}
}

func pushTopLibraryCard(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	return id
}

func TestOliviaSacrificesTwoTreasuresForCountersOnEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	olivia := b12Push(g, me.ID, "Olivia, Opulent Outlaw", "Legendary Creature — Vampire Assassin", oliviaOpulentOutlawOracle, 3, 3)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	treasures := pushTokens(g, me.ID, TreasureToken(), 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, olivia, 0, game.ActivateAbilityParams{SacrificeIDs: treasures})
	for _, id := range []uuid.UUID{olivia, bear} {
		if c, _ := battlefieldCard(g, id); c.Counters["+1/+1"] != 2 {
			t.Errorf("%s has %d +1/+1 counters, want 2", c.Name, c.Counters["+1/+1"])
		}
	}
	if c, _ := battlefieldCard(g, theirs); c.Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature gets none")
	}
}

func TestOliviaMakesOneTreasureWhenOutlawsConnect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Olivia, Opulent Outlaw", "Legendary Creature — Vampire Assassin", oliviaOpulentOutlawOracle, 3, 3)
	rogue := b16Creature(g, me.ID, "Rogue A", "Creature — Human Rogue", 2, 2, "B")
	pirate := b16Creature(g, me.ID, "Pirate B", "Creature — Human Pirate", 2, 2, "R")
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{rogue, pirate} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("two outlaws connecting make %d Treasures, want 1", got)
	}
}

func TestBreyaMakesThoptersAndEatsThemForEachMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Breya, Etherium Shaper", "Legendary Artifact Creature — Human", breyaEtheriumShaperOracle, nil)
	passPriorityAroundTable(t, g)
	breya := findBattlefieldByName(g, "Breya, Etherium Shaper")
	if got := countBattlefieldNamed(g, me.ID, "Thopter"); got != 2 {
		t.Fatalf("Breya makes two Thopters: %d", got)
	}
	var thopters []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" {
			thopters = append(thopters, c.InstanceID)
		}
	}
	life := opp.Life
	b16Activate(t, g, me.ID, breya, 0, game.ActivateAbilityParams{
		SacrificeIDs: thopters, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if opp.Life != life-3 {
		t.Errorf("the damage mode deals 3: %d -> %d", life, opp.Life)
	}
	arts := pushTokens(g, me.ID, TreasureToken(), 2)
	mine := me.Life
	b16Activate(t, g, me.ID, breya, 2, game.ActivateAbilityParams{SacrificeIDs: arts})
	if me.Life != mine+5 {
		t.Errorf("the life mode gains 5: %d -> %d", mine, me.Life)
	}
	bear := b16Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5, "G")
	arts = pushTokens(g, me.ID, TreasureToken(), 2)
	b16Activate(t, g, me.ID, breya, 1, game.ActivateAbilityParams{
		SacrificeIDs: arts, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 1 || tg != 1 {
		t.Errorf("the -4/-4 mode shrinks a 5/5 to 1/1, got %d/%d", p, tg)
	}
}

func TestWhisperSacrificesTwoCreaturesToReanimate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	whisper := b12Push(g, me.ID, "Whisper, Blood Liturgist", "Legendary Creature — Human Cleric", whisperBloodLiturgistOracle, 2, 2)
	a := sacNCreature(g, me.ID, "Bear A")
	big := pushGraveyardCardTyped(me, "Griselbrand", "Legendary Creature — Demon")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: big}}
	if err := g.ActivateCatalogAbility(me.ID, whisper, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a}, Targets: target}); err != game.ErrInvalidParam {
		t.Fatalf("one creature: got %v, want ErrInvalidParam", err)
	}
	// "Two creatures", not "two other creatures": Whisper may be one.
	b16Activate(t, g, me.ID, whisper, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, whisper}, Targets: target})
	if !g.Battlefield.Contains(big) {
		t.Error("the creature card returns to the battlefield")
	}
	if g.Battlefield.Contains(whisper) || g.Battlefield.Contains(a) {
		t.Error("both creatures are sacrificed")
	}
}
