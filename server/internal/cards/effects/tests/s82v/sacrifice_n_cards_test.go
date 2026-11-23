package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSacrificeNCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.PriestOfForgottenGodsOracle: "Priest of Forgotten Gods",
		effectstest.KuldothaForgemasterOracle:   "Kuldotha Forgemaster",
		effectstest.TamiyosJournalOracle:        "Tamiyo's Journal",
		effectstest.TeysaOrzhovScionOracle:      "Teysa, Orzhov Scion",
		effectstest.HedronDetonatorOracle:       "Hedron Detonator",
		effectstest.OliviaOpulentOutlawOracle:   "Olivia, Opulent Outlaw",
		effectstest.BreyaEtheriumShaperOracle:   "Breya, Etherium Shaper",
		effectstest.WhisperBloodLiturgistOracle: "Whisper, Blood Liturgist",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == effects.CompletenessUnreviewed {
			t.Errorf("%s ships without a Completeness declaration", name)
		}
	}
	// Transmutation Font's "with different names" is a set rule on the
	// clause (#3940), so the tutor ships with it.
	if spec, _ := effects.Lookup(effectstest.B32TransmutationFontOracle); spec.Completeness != effects.CompletenessFull || len(spec.Activated) != 4 {
		t.Error("Transmutation Font ships its tutor, Full")
	}
}

// --- the five caveats removed --------------------------------------

func TestSavvyHunterSacrificesTwoFoodsToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hunter := effectstest.B12Push(g, me.ID, "Savvy Hunter", "Creature — Human Warrior", effectstest.B30SavvyHunterOracle, 3, 3)
	foods := effectstest.PushTokens(g, me.ID, effects.FoodToken(), 2)
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: foods[:1]}); err != game.ErrInvalidParam {
		t.Fatalf("one Food: got %v, want ErrInvalidParam", err)
	}
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: foods})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 0 {
		t.Error("both Foods are sacrificed")
	}
}

func TestSamwiseGamgeeSacrificesThreeFoodsToReturnAHistoricCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sam := effectstest.B12Push(g, me.ID, "Samwise Gamgee", "Legendary Creature — Halfling Peasant", effectstest.B22SamwiseGamgeeOracle, 2, 2)
	relic := effectstest.PushGraveyardCardTyped(me, "Relic", "Artifact")
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	foods := effectstest.PushTokens(g, me.ID, effects.FoodToken(), 3)

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
	effectstest.B16Activate(t, g, me.ID, sam, 0, game.ActivateAbilityParams{
		SacrificeIDs: foods, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	})
	if !me.Hand.Contains(relic) {
		t.Error("the historic card returns to hand")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 0 {
		t.Error("all three Foods are sacrificed")
	}
}

func TestSaiSacrificesTwoArtifactsToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sai := effectstest.PushCatalogPermanent(g, me.ID, "Sai, Master Thopterist", "Legendary Creature — Human Artificer", effectstest.B07SaiMasterThopteristOracle, false)
	thopter := effects.TokenCard("1/1 colorless Thopter artifact with flying")
	arts := effectstest.PushTokens(g, me.ID, thopter, 2)
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, sai, 0, game.ActivateAbilityParams{SacrificeIDs: arts})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Thopter") != 0 {
		t.Error("both artifacts are sacrificed")
	}
	// Sai is not an artifact, so she can never pay for herself.
	more := effectstest.PushTokens(g, me.ID, thopter, 1)
	if err := g.ActivateCatalogAbility(me.ID, sai, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{more[0], sai}}); err != game.ErrIllegalTarget {
		t.Errorf("Sai as an artifact: got %v, want ErrIllegalTarget", err)
	}
}

func TestMagdaTheHoardmasterSacrificesThreeTreasuresForADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	magda := effectstest.B12Push(g, me.ID, "Magda, the Hoardmaster", "Legendary Creature — Dwarf Berserker", effectstest.B18MagdaTheHoardmasterOrcl, 2, 2)
	treasures := effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, magda, 0, game.ActivateAbilityParams{SacrificeIDs: treasures})
	dragon := effectstest.FindBattlefieldByName(g, "Scorpion Dragon")
	if dragon == uuid.Nil {
		t.Fatal("no Scorpion Dragon")
	}
	card, _ := effectstest.BattlefieldCard(g, dragon)
	if card.Power != 4 || card.Toughness != 4 || !card.HasColor("R") {
		t.Errorf("the Dragon is a 4/4 red: %+v", card)
	}
	abilities := effectstest.EffectiveAbilities(t, g, dragon)
	if !effects.EotHasAbility(abilities, "flying") || !effects.EotHasAbility(abilities, "haste") {
		t.Errorf("the Dragon has flying and haste: %v", abilities)
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("all three Treasures are sacrificed")
	}
}

func TestMagdaBrazenOutlawSacrificesFiveTreasuresToTutorOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	magda := effectstest.B11Push(g, me.ID, "Magda, Brazen Outlaw", "Legendary Creature — Dwarf Berserker", effectstest.B11MagdaOracle, 2, 1)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Wurmcoil", TypeLine: "Artifact Creature — Phyrexian Wurm"})
	treasures := effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 5)
	if err := g.ActivateCatalogAbility(me.ID, magda, 0, game.ActivateAbilityParams{SacrificeIDs: treasures[:4]}); err != game.ErrInvalidParam {
		t.Fatalf("four Treasures: got %v, want ErrInvalidParam", err)
	}
	effectstest.B16Activate(t, g, me.ID, magda, 0, game.ActivateAbilityParams{SacrificeIDs: treasures})
	effectstest.TakeSearchPrompts(t, g)
	if effectstest.FindBattlefieldByName(g, "Wurmcoil") == uuid.Nil {
		t.Error("the artifact is put onto the battlefield")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("all five Treasures are sacrificed")
	}
}

// --- the cards added -----------------------------------------------

func TestPriestOfForgottenGodsEdictsDrainsAddsManaAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	priest := effectstest.B12Push(g, me.ID, "Priest of Forgotten Gods", "Creature — Human Cleric", effectstest.PriestOfForgottenGodsOracle, 1, 2)
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")
	victim := sacNCreature(g, opp.ID, "Their Bear")
	// A second creature, so the target has a real choice to be asked.
	sacNCreature(g, opp.ID, "Their Spare")
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
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("target lost %d life, want 2", life-opp.Life)
	}
	// #1019: "you add {B}{B} and draw a card" is the run's
	// continuation, so it lands after the last target has answered —
	// the printed order.
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d before the sacrifice was answered, want 0", got)
	}
	if got := effectstest.PoolColors(me); len(got) != 0 {
		t.Errorf("pool %v before the sacrifice was answered, want empty", got)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the target player sacrifices a creature")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want B B", got)
	}
	if !effectstest.B16Tapped(t, g, priest) {
		t.Error("the Priest taps")
	}
}

func TestKuldothaForgemasterMayBeOneOfItsOwnThreeArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forge := effectstest.PushCatalogPermanent(g, me.ID, "Kuldotha Forgemaster", "Artifact Creature — Construct", effectstest.KuldothaForgemasterOracle, true)
	arts := effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 2)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Blightsteel", TypeLine: "Legendary Artifact Creature — Phyrexian Golem"})
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
	effectstest.B16Activate(t, g, me.ID, forge, 0, game.ActivateAbilityParams{SacrificeIDs: pay})
	effectstest.TakeSearchPrompts(t, g)
	if g.Battlefield.Contains(forge) {
		t.Error("the Forgemaster paid for itself")
	}
	if effectstest.FindBattlefieldByName(g, "Blightsteel") == uuid.Nil {
		t.Error("the artifact is put onto the battlefield")
	}
}

func TestTamiyosJournalInvestigatesAndTutorsForThreeClues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	journal := effectstest.PushCatalogPermanent(g, me.ID, "Tamiyo's Journal", "Legendary Artifact — Book", effectstest.TamiyosJournalOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Clue") != 1 {
		t.Fatalf("your upkeep investigates: %d Clues", effectstest.CountBattlefieldNamed(g, me.ID, "Clue"))
	}
	clues := append(effectstest.PushTokens(g, me.ID, effects.ClueToken(), 2), effectstest.FindBattlefieldByName(g, "Clue"))
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Sorcery"})
	if err := g.ActivateCatalogAbility(me.ID, journal, 0, game.ActivateAbilityParams{SacrificeIDs: clues}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Needle")
	if effectstest.CountBattlefieldNamed(g, me.ID, "Clue") != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teysa := effectstest.B12Push(g, me.ID, "Teysa, Orzhov Scion", "Legendary Creature — Human Advisor", effectstest.TeysaOrzhovScionOracle, 2, 3)
	w1 := effectstest.B16Creature(g, me.ID, "White A", "Creature — Soldier", 1, 1, "W")
	w2 := effectstest.B16Creature(g, me.ID, "White B", "Creature — Soldier", 1, 1, "W")
	black := effectstest.B16Creature(g, me.ID, "Black", "Creature — Zombie", 1, 1, "B")
	w3 := effectstest.B16Creature(g, me.ID, "White C", "Creature — Soldier", 1, 1, "W")
	threat := effectstest.B16Creature(g, opp.ID, "Threat", "Creature — Dragon", 6, 6, "R")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: threat}}

	if err := g.ActivateCatalogAbility(me.ID, teysa, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{w1, w2, black}, Targets: target,
	}); err != game.ErrIllegalTarget {
		t.Fatalf("a black creature among the three: got %v, want ErrIllegalTarget", err)
	}
	effectstest.B16Activate(t, g, me.ID, teysa, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{w1, w2, w3}, Targets: target})
	if g.Battlefield.Contains(threat) || !containsCard(g.Exile, threat) {
		t.Error("the target creature is exiled")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Spirit") != 0 {
		t.Error("white creatures dying make no Spirit")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(black) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Spirit") != 1 {
		t.Errorf("a black creature dying makes a Spirit: %d", effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"))
	}
}

func containsCard(z *game.Zone, id uuid.UUID) bool {
	return z != nil && z.Contains(id)
}

func TestHedronDetonatorPingsOnArtifactsAndImpulsesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	det := effectstest.B12Push(g, me.ID, "Hedron Detonator", "Creature — Goblin Artificer", effectstest.HedronDetonatorOracle, 2, 3)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("an artifact entering pings the target opponent: %d -> %d", life, opp.Life)
	}
	arts := append(effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 1), effectstest.FindBattlefieldByName(g, "Treasure"))
	top := pushTopLibraryCard(me, "Impulse")
	effectstest.B16Activate(t, g, me.ID, det, 0, game.ActivateAbilityParams{SacrificeIDs: arts})
	if !containsCard(g.Exile, top) {
		t.Error("the top card of the library is exiled")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("both artifacts are sacrificed")
	}
}

func pushTopLibraryCard(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	return id
}

func TestOliviaSacrificesTwoTreasuresForCountersOnEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	olivia := effectstest.B12Push(g, me.ID, "Olivia, Opulent Outlaw", "Legendary Creature — Vampire Assassin", effectstest.OliviaOpulentOutlawOracle, 3, 3)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	treasures := effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, olivia, 0, game.ActivateAbilityParams{SacrificeIDs: treasures})
	for _, id := range []uuid.UUID{olivia, bear} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Counters["+1/+1"] != 2 {
			t.Errorf("%s has %d +1/+1 counters, want 2", c.Name, c.Counters["+1/+1"])
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature gets none")
	}
}

// "Put two +1/+1 counters on each creature" is one placement of two,
// so Hardened Scales adds one counter per creature (2+1 = 3), not one
// per counter placed (two placements of one would give 4).
func TestOliviaCountersAreOnePlacementUnderHardenedScales(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	olivia := effectstest.B12Push(g, me.ID, "Olivia, Opulent Outlaw", "Legendary Creature — Vampire Assassin", effectstest.OliviaOpulentOutlawOracle, 3, 3)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	_ = effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	treasures := effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, olivia, 0, game.ActivateAbilityParams{SacrificeIDs: treasures})
	for _, id := range []uuid.UUID{olivia, bear} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Counters["+1/+1"] != 3 {
			t.Errorf("%s has %d +1/+1 counters under Hardened Scales, want 3", c.Name, c.Counters["+1/+1"])
		}
	}
}

func TestOliviaMakesOneTreasureWhenOutlawsConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Olivia, Opulent Outlaw", "Legendary Creature — Vampire Assassin", effectstest.OliviaOpulentOutlawOracle, 3, 3)
	rogue := effectstest.B16Creature(g, me.ID, "Rogue A", "Creature — Human Rogue", 2, 2, "B")
	pirate := effectstest.B16Creature(g, me.ID, "Pirate B", "Creature — Human Pirate", 2, 2, "R")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{rogue, pirate} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("two outlaws connecting make %d Treasures, want 1", got)
	}
}

func TestBreyaMakesThoptersAndEatsThemForEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Breya, Etherium Shaper", "Legendary Artifact Creature — Human", effectstest.BreyaEtheriumShaperOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	breya := effectstest.FindBattlefieldByName(g, "Breya, Etherium Shaper")
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Thopter"); got != 2 {
		t.Fatalf("Breya makes two Thopters: %d", got)
	}
	var thopters []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" {
			thopters = append(thopters, c.InstanceID)
		}
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, breya, 0, game.ActivateAbilityParams{
		SacrificeIDs: thopters, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if opp.Life != life-3 {
		t.Errorf("the damage mode deals 3: %d -> %d", life, opp.Life)
	}
	arts := effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 2)
	mine := me.Life
	effectstest.B16Activate(t, g, me.ID, breya, 2, game.ActivateAbilityParams{SacrificeIDs: arts})
	if me.Life != mine+5 {
		t.Errorf("the life mode gains 5: %d -> %d", mine, me.Life)
	}
	bear := effectstest.B16Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5, "G")
	arts = effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 2)
	effectstest.B16Activate(t, g, me.ID, breya, 1, game.ActivateAbilityParams{
		SacrificeIDs: arts, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 1 || tg != 1 {
		t.Errorf("the -4/-4 mode shrinks a 5/5 to 1/1, got %d/%d", p, tg)
	}
}

func TestWhisperSacrificesTwoCreaturesToReanimate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	whisper := effectstest.B12Push(g, me.ID, "Whisper, Blood Liturgist", "Legendary Creature — Human Cleric", effectstest.WhisperBloodLiturgistOracle, 2, 2)
	a := sacNCreature(g, me.ID, "Bear A")
	big := effectstest.PushGraveyardCardTyped(me, "Griselbrand", "Legendary Creature — Demon")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: big}}
	if err := g.ActivateCatalogAbility(me.ID, whisper, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a}, Targets: target}); err != game.ErrInvalidParam {
		t.Fatalf("one creature: got %v, want ErrInvalidParam", err)
	}
	// "Two creatures", not "two other creatures": Whisper may be one.
	effectstest.B16Activate(t, g, me.ID, whisper, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, whisper}, Targets: target})
	if !g.Battlefield.Contains(big) {
		t.Error("the creature card returns to the battlefield")
	}
	if g.Battlefield.Contains(whisper) || g.Battlefield.Contains(a) {
		t.Error("both creatures are sacrificed")
	}
}
