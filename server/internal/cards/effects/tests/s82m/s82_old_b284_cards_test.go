package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b284_cards_test.go — S82 older sets, Shadows over Innistrad
// batch 284 (#5543): card draw and library.

const (
	oldB284Confront  = "221240c5-3c97-438c-b906-2508eacf190f"
	oldB284Fork      = "3bc12623-3852-42f0-b458-5c67fe1b7a07"
	oldB284Pore      = "55fe5ec3-1c81-4e9d-9e4c-f68ce6d7c2de"
	oldB284Sleuth    = "69ca69c7-403b-4b5d-be08-d9dce8ce410e"
	oldB284Hermit    = "6d32f688-a854-4d82-b654-d720b93fb766"
	oldB284Patrol    = "784f0ca3-efe1-4e22-aedd-57cba7996fa5"
	oldB284Missing   = "a31a66ee-5ddf-4d70-9ae9-725fc7327706"
	oldB284Lamp      = "a88c5540-a242-49bc-a725-39b9fda085c3"
	oldB284Expose    = "b5483037-5dd2-4282-a201-de2e08aaa76e"
	oldB284Trail     = "e10065ba-fe24-4974-a500-d4151397a4d9"
	oldB284Merciless = "e9eecc10-8821-4b61-a042-01fef435d37c"
	oldB284Vessel    = "f6894486-0bac-4b72-92b7-5ce99fa426e6"
	oldB284Fleeting  = "fbb0173d-7ec9-4f43-b457-f8de6703497e"
)

func TestOldB284CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB284Confront: "Confront the Unknown", oldB284Fork: "Fork in the Road", oldB284Pore: "Pore Over the Pages",
		oldB284Sleuth: "Daring Sleuth", oldB284Sleuth + "#1": "Bearer of Overwhelming Truths",
		oldB284Hermit: "Hermit of the Natterknolls", oldB284Hermit + "#1": "Lone Wolf of the Natterknolls",
		oldB284Patrol: "Briarbridge Patrol", oldB284Missing: "Gone Missing", oldB284Lamp: "Lamplighter of Selhoff",
		oldB284Expose: "Expose Evil", oldB284Trail: "Trail of Evidence", oldB284Merciless: "Merciless Resolve",
		oldB284Vessel: "Vessel of Paramnesia", oldB284Fleeting: "Fleeting Memories",
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

func oldB284Clues(g *game.Game, p *game.Player) int { return effectstest.TmtB1Count(g, p.ID, "Clue") }

func oldB284SacClue(g *game.Game, p *game.Player) {
	clue := effectstest.PushToken(g, p.ID, effects.ClueToken())
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(clue) })
}

func TestOldB284ConfrontTheUnknownCountsTheNewClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushToken(g, me.ID, effects.ClueToken())
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Confront the Unknown", "Instant", oldB284Confront, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB284Clues(g, me); got != 2 {
		t.Fatalf("%d Clues, want 2", got)
	}
	if p, tou := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tou != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4 (two Clues)", p, tou)
	}
}

func TestOldB284ForkInTheRoadPutsOneInHandAndOneInGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	island := effectstest.SeedLibraryTop(me, "Island", "Basic Land — Island", "")
	nonbasic := effectstest.SeedLibraryTop(me, "Nonbasic", "Land", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAt(t, g, "Fork in the Road", "Sorcery", oldB284Fork)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest, island)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestRevealPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to choose the card for the hand")
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, []uuid.UUID{island}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, island) != game.ZoneHand {
		t.Errorf("the chosen Island is in %v, want hand", effectstest.B12ZoneOf(g, island))
	}
	if effectstest.B12ZoneOf(g, forest) != game.ZoneGraveyard {
		t.Errorf("the other Forest is in %v, want graveyard", effectstest.B12ZoneOf(g, forest))
	}
	if effectstest.B12ZoneOf(g, nonbasic) != game.ZoneLibrary {
		t.Error("a nonbasic land left the library")
	}
}

func TestOldB284PoreOverThePagesDrawsUntapsThenDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Land", "")
	}
	land1 := effectstest.B12Permanent(g, me.ID, "Island One", "Basic Land — Island")
	land2 := effectstest.B12Permanent(g, me.ID, "Island Two", "Basic Land — Island")
	land3 := effectstest.B12Permanent(g, me.ID, "Island Three", "Basic Land — Island")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{land1, land2, land3} {
			effectstest.FindBattlefieldCardForTest(g, id).Tapped = true
		}
	})
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Pore Over the Pages", "Sorcery", oldB284Pore, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand %d, want %d after drawing three", got, before+3)
	}
	// The untap question comes before the discard.
	c := effectstest.LatestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no untap prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land1, land2}); err != nil {
		t.Fatalf("untap two lands: %v", err)
	}
	if effectstest.IsTapped(g, land1) || effectstest.IsTapped(g, land2) || !effectstest.IsTapped(g, land3) {
		t.Error("exactly the two chosen lands should be untapped")
	}
	discard := effectstest.DiscardChoiceFor(g, me.ID)
	if discard == nil {
		t.Fatal("no discard prompt after the untap")
	}
}

func TestOldB284DaringSleuthTransformsWhenYouSacrificeAClueAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oldB284Sleuth, Layout: game.LayoutTransform, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Daring Sleuth", TypeLine: "Creature — Human Rogue", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
			{Name: "Bearer of Overwhelming Truths", TypeLine: "Creature — Human Wizard", Colors: []string{"U"}, Power: 3, Toughness: 2},
		},
	}
	c.SetFace(0)
	sleuth := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	oldB284SacClue(g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.TransformedPermanent(effectstest.B12Card(t, g, sleuth)) {
		t.Fatal("an opponent's Clue sacrifice transformed the Sleuth")
	}
	oldB284SacClue(g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.TransformedPermanent(effectstest.B12Card(t, g, sleuth)) {
		t.Fatal("the Sleuth did not transform")
	}
	if !effectstest.HasEffectiveKeyword(t, g, sleuth, "prowess") {
		t.Error("the Bearer has no prowess")
	}
	effectstest.DeclareAttack(t, g, opp.ID, sleuth)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB284Clues(g, me); got != 1 {
		t.Errorf("%d Clues after combat damage to a player, want 1", got)
	}
}

func TestOldB284HermitDrawsOnOpponentCastsDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 4; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Land", "")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Hermit of the Natterknolls", "Creature — Human Werewolf", oldB284Hermit, 2, 3)
	before := me.Hand.Size()
	spell := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("the opponent's cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d after an opponent's spell on my turn", got, before+1)
	}
	// My own spell does not draw.
	before = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Filler Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Error("my own spell drew a card")
	}
}

func TestOldB284HermitTransformsOnAQuietTurnAndLoneWolfDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oldB284Hermit, Layout: game.LayoutTransform, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Hermit of the Natterknolls", TypeLine: "Creature — Human Werewolf", ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 2, Toughness: 3},
			{Name: "Lone Wolf of the Natterknolls", TypeLine: "Creature — Werewolf", Colors: []string{"G"}, Power: 3, Toughness: 5},
		},
	}
	c.SetFace(0)
	hermit := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	// No spell is cast on this turn; the next upkeep transforms it.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.TransformedPermanent(effectstest.B12Card(t, g, hermit)) {
		t.Fatal("the Hermit did not transform after a turn with no spells")
	}
	for i := 0; i < 4; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Land", "")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	before := me.Hand.Size()
	spell := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("the opponent's cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d, want %d (Lone Wolf draws two)", got, before+2)
	}
}

func TestOldB284BriarbridgePatrolInvestigatesOncePerBatchAndPutsACreatureWithThreeClues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	patrol := effectstest.B12Push(g, me.ID, "Briarbridge Patrol", "Creature — Human Warrior", oldB284Patrol, 3, 3)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 3)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 3)
	effectstest.DeclareAttack(t, g, opp.ID, patrol)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, blocker := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(blocker, patrol); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInCombatDeclaration(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceDamageAssignment && pc.DamageAssignment != nil {
			var entries []game.DamageAssignmentEntry
			for _, blocker := range pc.DamageAssignment.BlockerIDs {
				entries = append(entries, game.DamageAssignmentEntry{BlockerID: blocker, Amount: pc.DamageAssignment.AttackerPower / len(pc.DamageAssignment.BlockerIDs)})
			}
			entries[0].Amount += pc.DamageAssignment.AttackerPower - 2*(pc.DamageAssignment.AttackerPower/2)
			if err := g.ResolveDamageAssignment(pc.ID, pc.Chooser, entries, 0); err != nil {
				t.Fatalf("ResolveDamageAssignment: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB284Clues(g, me); got != 1 {
		t.Errorf("%d Clues after damaging two creatures at once, want 1", got)
	}
}

func TestOldB284BriarbridgePatrolPutsACreatureFromHandAfterThreeClueSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Briarbridge Patrol", "Creature — Human Warrior", oldB284Patrol, 3, 3)
	big := effectstest.B13HandCard(me, "Hand Giant", "Creature — Giant", "")
	oldB284SacClue(g, me)
	oldB284SacClue(g, me)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, big) != game.ZoneHand {
		t.Fatal("two sacrificed Clues were enough")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	effectstest.B12Push(g2, me2.ID, "Briarbridge Patrol", "Creature — Human Warrior", oldB284Patrol, 3, 3)
	big2 := effectstest.B13HandCard(me2, "Hand Giant", "Creature — Giant", "")
	for i := 0; i < 3; i++ {
		oldB284SacClue(g2, me2)
	}
	effectstest.AdvanceToEndStep(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if c := effectstest.LatestChooseCards(g2, me2.ID); c != nil {
		if err := g2.ResolveChooseCards(c.ID, me2.ID, []uuid.UUID{big2}); err != nil {
			t.Fatalf("pick the creature: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.B12ZoneOf(g2, big2) != game.ZoneBattlefield {
		t.Errorf("the creature is in %v, want the battlefield", effectstest.B12ZoneOf(g2, big2))
	}
}

func TestOldB284GoneMissingTucksAnyPermanentAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Gone Missing", "Sorcery", oldB284Missing, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneLibrary {
		t.Errorf("the target is in %v, want its owner's library", effectstest.B12ZoneOf(g, theirs))
	}
	if opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != theirs && opp.Library.Cards[0].InstanceID != theirs {
		t.Error("the target is not on top of its owner's library")
	}
	if got := oldB284Clues(g, me); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestOldB284LamplighterLootsOnlyWithAnotherZombie(t *testing.T) {
	for _, zombie := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < 3; i++ {
			effectstest.SeedLibraryTop(me, "Filler", "Land", "")
		}
		effectstest.AdvanceToMain(t, g)
		if zombie {
			effectstest.B12Creature(g, me.ID, "Other Zombie", "Creature — Zombie", 2, 2)
		}
		before := me.Hand.Size()
		effectstest.CastAndResolveCreature(t, g, "Lamplighter of Selhoff", "Creature — Zombie Horror", oldB284Lamp)
		effectstest.PassPriorityAroundTable(t, g)
		prompted := oldB284AnswerConfirmIfOpen(t, g, me)
		if !zombie && prompted {
			t.Error("no other Zombie: the loot was offered")
		}
		if zombie && !prompted {
			t.Error("another Zombie: the loot was not offered")
		}
		_ = before
	}
}

func TestOldB284ExposeEvilTapsUpToTwoAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Expose Evil", "Instant", oldB284Expose, effectstest.CardTargets(a, b, c)); err == nil {
		t.Error("three targets were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Expose Evil", "Instant", oldB284Expose, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, a) || !effectstest.IsTapped(g, b) || effectstest.IsTapped(g, c) {
		t.Error("exactly the two targets should be tapped")
	}
	if got := oldB284Clues(g, me); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestOldB284TrailOfEvidenceInvestigatesOnInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Trail of Evidence", "Enchantment", oldB284Trail, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB284Clues(g, me); got != 0 {
		t.Fatalf("a creature spell made %d Clues", got)
	}
	effectstest.CastCatalogSpell(t, g, "Filler Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Filler Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB284Clues(g, me); got != 2 {
		t.Errorf("%d Clues after a sorcery and an instant, want 2", got)
	}
}

func TestOldB284MercilessResolveNeedsASacrificeAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Land", "")
	}
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B13HandCard(me, "Merciless Resolve", "Instant", oldB284Merciless)
	effectstest.ApaMana(me, "B", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with nothing to sacrifice")
	}
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	before := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("cast sacrificing a land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if got := me.Hand.Size(); got != before-1+2 {
		t.Errorf("hand %d, want %d", got, before-1+2)
	}
}

func TestOldB284VesselOfParamnesiaMillsTargetAndYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 5; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Land", "")
		effectstest.SeedLibraryTop(opp, "Filler", "Land", "")
	}
	effectstest.AdvanceToMain(t, g)
	vessel := effectstest.SeedPermanentWithOracle(g, me.ID, "Vessel of Paramnesia", "Enchantment", oldB284Vessel)
	grave, hand := opp.Graveyard.Size(), me.Hand.Size()
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, vessel, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.AcPlayerRef(opp.ID)})
	if g.Battlefield.Contains(vessel) {
		t.Error("the Vessel was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != grave+3 {
		t.Errorf("the target's graveyard is %d, want %d", opp.Graveyard.Size(), grave+3)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the controller draws)", me.Hand.Size(), hand+1)
	}
}

func TestOldB284FleetingMemoriesInvestigatesAndMillsOnClueSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 8; i++ {
		effectstest.SeedLibraryTop(opp, "Filler", "Land", "")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAndResolveCreature(t, g, "Fleeting Memories", "Enchantment", oldB284Fleeting)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB284Clues(g, me); got != 1 {
		t.Fatalf("%d Clues on entering, want 1", got)
	}
	grave := opp.Graveyard.Size()
	oldB284SacClue(g, me)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != grave+3 {
		t.Errorf("the opponent's graveyard is %d, want %d", got, grave+3)
	}
}

// oldB284AnswerConfirmIfOpen says yes to an open yes/no prompt and reports
// whether there was one.
func oldB284AnswerConfirmIfOpen(t *testing.T, g *game.Game, me *game.Player) bool {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == me.ID {
			if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return true
		}
	}
	return false
}
