package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b12_cards_test.go — Secrets of Strixhaven Commander, twelfth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	socB12BullOracle       = "61747805-0c34-4beb-8cbf-97bac7a998d0"
	socB12TechniqueOracle  = "163baf58-b35d-4f2b-aa53-8384b8b5800a"
	socB12RoasterOracle    = "86c55fe5-98b5-4e53-8828-a7f04ee3e985"
	socB12RionyaOracle     = "be4f47e9-1bd2-4aa9-bc85-3f203ad21639"
	socB12RoothaOracle     = "348c67ef-9ccc-4651-9038-efdf1ad1b36a"
	socB12RefrainOracle    = "515c522f-51d5-4f6d-b9b1-ea3c7f6a378e"
	socB12ScreamsOracle    = "a3b3b9ae-1e48-481d-80a5-7e2a26cf2f50"
	socB12ScrivOracle      = "64c986f0-e786-419c-9b51-8c316c8fae0e"
	socB12EyesOracle       = "0a7d6489-b239-4ce8-aea3-7b5d0e436f70"
	socB12ParagonOracle    = "ecd8d7f4-7b11-4097-afc0-ef3c81edd90c"
	socB12VanillaOracleTag = "b12-vanilla-test-card"
)

func TestSocB12EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB12BullOracle:      "Renegade Bull",
		socB12TechniqueOracle: "Replication Technique",
		socB12RoasterOracle:   "Ribtruss Roaster",
		socB12RionyaOracle:    "Rionya, Fire Dancer",
		socB12RoothaOracle:    "Rootha, Mastering the Moment",
		socB12RefrainOracle:   "Rousing Refrain",
		socB12ScreamsOracle:   "Screams from Within",
		socB12ScrivOracle:     "Scriv, the Obligator",
		socB12EyesOracle:      "Sentinel's Eyes",
		socB12ParagonOracle:   "Serra Paragon",
	}
	for oracle, name := range want {
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

// --- Renegade Bull ----------------------------------------------------------

func TestSocB12RenegadeBullGetsPlusXForAnInstantOrSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bull := b12Push(g, me.ID, "Renegade Bull", "Creature — Ox", socB12BullOracle, 0, 5)
	if !hasEffectiveKeyword(t, g, bull, "trample") {
		t.Error("the Bull lacks trample")
	}
	socB4Spell(t, g, "Big Sorcery", "Sorcery", "{3}{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bull); p != 4 {
		t.Fatalf("power %d after a mana value 4 sorcery, want 4", p)
	}
	socB4Spell(t, g, "Zap", "Instant", "{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bull); p != 5 {
		t.Fatalf("power %d after a mana value 1 instant too, want 5", p)
	}
	socB4Spell(t, g, "Grizzly", "Creature — Bear", "{1}{G}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bull); p != 5 {
		t.Errorf("power %d after a creature spell, want 5 (it does not trigger)", p)
	}
}

func TestSocB12RenegadeBullCountsTheXOfAnXSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bull := b12Push(g, me.ID, "Renegade Bull", "Creature — Ox", socB12BullOracle, 0, 5)
	socB4Spell(t, g, "X Spell", "Sorcery", "{X}{R}", game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bull); p != 4 {
		t.Errorf("power %d after an X=3 sorcery, want 4 (X counts in its mana value)", p)
	}
}

func TestSocB12RenegadeBullAttackCopiesAndCastsAnInstantForFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bull := b12Push(g, me.ID, "Renegade Bull", "Creature — Ox", socB12BullOracle, 0, 5)
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	life := opp.Life

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bull, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerTriggerTargets(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !inExile(g, bolt) {
		t.Fatal("the targeted instant is not exiled")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy without paying: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 from the free Bolt", life, opp.Life)
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left over after the copy resolved", n)
	}
	if !inExile(g, bolt) {
		t.Error("the original card stays in exile")
	}
}

func TestSocB12RenegadeBullAttackWithNoTargetOrADeclinedCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bull := b12Push(g, me.ID, "Renegade Bull", "Creature — Ox", socB12BullOracle, 0, 5)
	// A creature card is not a legal target.
	gyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	// Declined: the card is exiled, the copy is not cast and ceases to exist.
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	life := opp.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bull, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerTriggerTargets(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	declineResolutionCasts(t, g, me.ID)
	if !inExile(g, bolt) {
		t.Error("the card is exiled whether or not the copy is cast")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after declining", n)
	}
	if opp.Life != life {
		t.Errorf("life moved to %d", opp.Life)
	}
}

// --- Replication Technique --------------------------------------------------

func socB12CastTechnique(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Replication Technique", "Sorcery", "{4}{U}", socB12TechniqueOracle,
		game.CastSpellParams{Targets: cardTargets(target)})
}

func TestSocB12ReplicationTechniqueCopiesATargetPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	socB12CastTechnique(t, g, me, bear)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Bear", me.ID); n != 2 {
		t.Fatalf("%d Bears, want the original and one token copy", n)
	}
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.IsToken() {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d token Bears, want 1", tokens)
	}
}

func TestSocB12ReplicationTechniqueDemonstrateCopiesForYouAndAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	socB12CastTechnique(t, g, me, mine)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	for i := 0; i < 12; i++ {
		switch {
		case copyPromptFor(g, me.ID) != nil:
			c := copyPromptFor(g, me.ID)
			if err := g.ResolvePickTargets(c.ID, me.ID, cardTargets(mine)); err != nil {
				t.Fatalf("my copy keeps its target: %v", err)
			}
		case copyPromptFor(g, opp.ID) != nil:
			c := copyPromptFor(g, opp.ID)
			if err := g.ResolvePickTargets(c.ID, opp.ID, cardTargets(theirs)); err != nil {
				t.Fatalf("the opponent's copy picks their creature: %v", err)
			}
		case latestOptionPickFor(g, me.ID) != nil:
			answerChoosePlayer(t, g, me.ID, opp)
		default:
			i = 12
		}
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Bear", me.ID); n != 3 {
		t.Errorf("%d Bears for me, want the original and two token copies (the original and my copy)", n)
	}
	if n := countOnBattlefield(g, "Wolf", opp.ID); n != 2 {
		t.Errorf("%d Wolves for the opponent, want theirs and one token copy from their copy of the spell", n)
	}
	if n := countOnBattlefield(g, "Wolf", me.ID); n != 0 {
		t.Errorf("the opponent's copy made a Wolf for me")
	}
}

// --- Ribtruss Roaster -------------------------------------------------------

func TestSocB12RibtrussRoasterDevoursAndMakesAPestPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushDevourFodder(g, me.ID, "Goblin A")
	pushDevourFodder(g, me.ID, "Goblin B")
	id := castDevourer(t, g, "Ribtruss Roaster", socB12RoasterOracle)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a)
	if got := devourPlusOnes(g, id); got != 1 {
		t.Fatalf("counters = %d, want 1 (devour 1, one creature)", got)
	}
	// Another counter from elsewhere counts too: "equal to the number of
	// +1/+1 counters on this creature".
	socB9Counter(t, g, id, game.CounterPlusOne, 1)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Pest", me.ID); got != 2 {
		t.Fatalf("%d Pests, want 2 (one per +1/+1 counter)", got)
	}
	// Each Pest gains 1 life when it dies.
	pest := findBattlefieldByName(g, "Pest")
	life := me.Life
	killCreature(t, g, me.ID, pest)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 from the Pest dying", life, me.Life)
	}
}

func TestSocB12RibtrussRoasterWithoutCountersMakesNothingAndOtherEndStepsDont(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Ribtruss Roaster", "Creature — Troll Druid", socB12RoasterOracle, 3, 3)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Pest", me.ID); got != 0 {
		t.Errorf("%d Pests with no counters, want 0", got)
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	id := b12Push(g2, me2.ID, "Ribtruss Roaster", "Creature — Troll Druid", socB12RoasterOracle, 3, 3)
	socB9Counter(t, g2, id, game.CounterPlusOne, 2)
	advanceToEndStep(t, g2)
	passPriorityAroundTable(t, g2)
	if got := countOnBattlefield(g2, "Pest", me2.ID); got != 2 {
		t.Fatalf("%d Pests at my end step, want 2", got)
	}
	advanceToEndStepOf(t, g2, 1)
	passPriorityAroundTable(t, g2)
	if got := countOnBattlefield(g2, "Pest", me2.ID); got != 2 {
		t.Errorf("%d Pests after another player's end step, want still 2", got)
	}
}

// --- Rionya, Fire Dancer ----------------------------------------------------

func TestSocB12RionyaCopiesAnotherCreatureOncePlusOnePerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rionya := b12Push(g, me.ID, "Rionya, Fire Dancer", "Legendary Creature — Human Wizard", socB12RionyaOracle, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	socB4Spell(t, g, "Zap", "Instant", "{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	socB4Spell(t, g, "Bolt", "Sorcery", "{1}{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// A creature spell does not count.
	socB4Spell(t, g, "Grizzly", "Creature — Bear", "{1}{G}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	advanceTo(t, g, game.StepBeginCombat)
	if c := latestPickTarget(g, me.ID); c == nil {
		passPriorityAroundTable(t, g)
	}
	if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardTargets(rionya)); err == nil {
		t.Fatal("Rionya was a legal target for her own \"another target creature\"")
	}
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	var copies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.IsToken() {
			copies = append(copies, c.InstanceID)
		}
	}
	if len(copies) != 3 {
		t.Fatalf("%d token Bears, want 3 (one plus the two instants and sorceries cast)", len(copies))
	}
	for _, id := range copies {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Error("a copy lacks haste")
		}
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Errorf("%d Bears after the end step, want only the original (the copies are exiled)", n)
	}
}

func TestSocB12RionyaMakesOneCopyWithNoSpellsCastAndOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Rionya, Fire Dancer", "Legendary Creature — Human Wizard", socB12RionyaOracle, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if latestPickTarget(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Bear", me.ID); n != 2 {
		t.Errorf("%d Bears, want the original and exactly one copy", n)
	}
}

// --- Rootha, Mastering the Moment ------------------------------------------

func socB12Elementals(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Elemental" && c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestSocB12RoothaMakesAnElementalAsBigAsTheGreatestInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Rootha, Mastering the Moment", "Legendary Creature — Orc Sorcerer", socB12RoothaOracle, 3, 4)
	socB4Spell(t, g, "Zap", "Instant", "{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	socB4Spell(t, g, "Four", "Sorcery", "{3}{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	socB4Spell(t, g, "Big Creature", "Creature — Dragon", "{5}{R}{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	els := socB12Elementals(g, me.ID)
	if len(els) != 1 {
		t.Fatalf("%d Elementals, want 1", len(els))
	}
	el := els[0]
	if p, tg := effectivePower(t, g, el), effectiveToughness(t, g, el); p != 4 || tg != 4 {
		t.Errorf("Elemental is %d/%d, want 4/4 (the sorcery's mana value, the creature spell does not count)", p, tg)
	}
	if !hasEffectiveKeyword(t, g, el, "flying") || !hasEffectiveKeyword(t, g, el, "haste") {
		t.Error("the Elemental lacks flying or haste")
	}
	if c, _ := battlefieldCard(g, el); len(c.Colors) != 2 {
		t.Errorf("colours %v, want blue and red", c.Colors)
	}
}

func TestSocB12RoothaCountsXAndNeedsASpellThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Rootha, Mastering the Moment", "Legendary Creature — Orc Sorcerer", socB12RoothaOracle, 3, 4)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := len(socB12Elementals(g, me.ID)); n != 0 {
		t.Fatalf("%d Elementals with no instant or sorcery cast, want 0", n)
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	b12Push(g2, me2.ID, "Rootha, Mastering the Moment", "Legendary Creature — Orc Sorcerer", socB12RoothaOracle, 3, 4)
	socB4Spell(t, g2, "X Spell", "Sorcery", "{X}{R}", game.CastSpellParams{XValue: 4})
	passPriorityAroundTable(t, g2)
	advanceTo(t, g2, game.StepBeginCombat)
	passPriorityAroundTable(t, g2)
	els := socB12Elementals(g2, me2.ID)
	if len(els) != 1 {
		t.Fatalf("%d Elementals, want 1", len(els))
	}
	if p := effectivePower(t, g2, els[0]); p != 5 {
		t.Errorf("Elemental power %d, want 5 (X=4 plus the {R} in its mana value)", p)
	}
}

// --- Rousing Refrain --------------------------------------------------------

func TestSocB12RousingRefrainAddsRedPerCardInHandKeepsItAndSuspendsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cards := opp.Hand.Size()
	if cards == 0 {
		t.Fatal("the opponent has an empty hand")
	}
	id := castFromHandForTest(t, g, me, "Rousing Refrain", "Sorcery", "{3}{R}{R}", socB12RefrainOracle,
		game.CastSpellParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != cards {
		t.Fatalf("mana pool holds %d, want %d (one {R} per card in the opponent's hand)", got, cards)
	}
	if !inExile(g, id) {
		t.Fatal("Rousing Refrain is not exiled")
	}
	if n := counterCountInExile(g, id, game.CounterTime); n != 3 {
		t.Fatalf("%d time counters, want 3", n)
	}
	// The mana is not lost as steps and phases end.
	advanceTo(t, g, game.StepBeginCombat)
	if got := me.ManaPool.Total(); got != cards {
		t.Errorf("mana pool holds %d at the beginning of combat, want %d kept", got, cards)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := counterCountInExile(g, id, game.CounterTime); n != 2 {
		t.Errorf("%d time counters after one upkeep, want 2 (suspend counts down)", n)
	}
}

func TestSocB12RousingRefrainCanBeSuspendedFromHand(t *testing.T) {
	spec, _ := Lookup(socB12RefrainOracle)
	if len(spec.SpecialActions) != 1 {
		t.Fatalf("%d special actions, want suspend", len(spec.SpecialActions))
	}
	sa := spec.SpecialActions[0]
	if sa.Kind != game.SpecialActionSuspend || sa.Counters != 3 || sa.Cost != "{1}{R}" {
		t.Errorf("special action %+v, want suspend 3—{1}{R}", sa)
	}
}

// counterCountInExile reads a counter on a card sitting in exile.
func counterCountInExile(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return -1
}

// --- Screams from Within ----------------------------------------------------

func TestSocB12ScreamsFromWithinShrinksAndReturnsWhenTheCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wolf := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Screams from Within", socB12ScreamsOracle, wolf)
	if p, tg := effectivePower(t, g, wolf), effectiveToughness(t, g, wolf); p != 1 || tg != 1 {
		t.Fatalf("enchanted Wolf is %d/%d, want 1/1", p, tg)
	}
	killCreature(t, g, opp.ID, wolf)
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil {
		answerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
		passPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("Screams from Within did not return to the battlefield")
	}
	if host := attachmentHostOf(t, g, aura); host.ID != bear {
		t.Errorf("it enchants %v, want the creature its controller chose", host.ID)
	}
	if p := effectivePower(t, g, bear); p != 1 {
		t.Errorf("the new host has power %d, want 1", p)
	}
}

func TestSocB12ScreamsFromWithinStaysWhenNothingCanBeEnchantedOrItLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Screams from Within", socB12ScreamsOracle, bear)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !me.Graveyard.Contains(aura) {
		t.Error("with no creature left to enchant the Aura stays in the graveyard")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	bear2 := auraBear(g2, me2.ID)
	auraBear(g2, me2.ID)
	aura2 := auraCast(t, g2, "Screams from Within", socB12ScreamsOracle, bear2)
	killCreature(t, g2, me2.ID, bear2)
	// Exile the Aura in response to the trigger: it is no longer in the graveyard.
	if err := g2.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me2.ID}, game.ZoneRef{Kind: game.ZoneExile}, aura2); err != nil {
		t.Fatalf("exile the Aura: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(aura2) || auraHostChoiceFor(g2, me2.ID) != nil {
		t.Error("an Aura exiled from the graveyard came back")
	}
}

// --- Scriv, the Obligator ---------------------------------------------------

func socB12Contracts(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Contract" {
			out = append(out, c)
		}
	}
	return out
}

func socB12ScrivEnters(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Scriv, the Obligator", "Legendary Creature — Inkling Bird", "{2}{W}{B}", socB12ScrivOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		answerTriggerTargets(t, g, me.ID, target)
	}
	passPriorityAroundTable(t, g)
}

func TestSocB12ScrivMakesAContractAttachedToAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wolf := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	socB12ScrivEnters(t, g, me, wolf)
	contracts := socB12Contracts(g)
	if len(contracts) != 1 {
		t.Fatalf("%d Contracts, want 1", len(contracts))
	}
	c := contracts[0]
	if !c.IsToken() || !c.IsAura() || c.Controller != me.ID || len(c.Colors) != 1 || c.Colors[0] != "W" {
		t.Errorf("Contract %+v, want a white Aura token under Scriv's controller", c)
	}
	if host := attachmentHostOf(t, g, c.InstanceID); host.ID != wolf {
		t.Errorf("Contract is attached to %v, want the opponent's Wolf", host.ID)
	}
	if text := game.TokenTextForCard(c); text == "" {
		t.Error("the Contract has no printed text")
	}
	_ = mine
}

func TestSocB12ScrivContractPumpsAnAttackOnAnotherOpponentAndCostsLifeOtherwise(t *testing.T) {
	// An attack on one of Scriv's controller's OTHER opponents: +2/+0.
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	wolf := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	socB12ScrivEnters(t, g, me, wolf)
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolf, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, wolf); p != 5 {
		t.Errorf("Wolf power %d attacking another of Scriv's controller's opponents, want 5", p)
	}
	if opp.Life != 40 {
		t.Errorf("opponent life %d, want 40 untouched", opp.Life)
	}

	// An attack on Scriv's controller: its controller loses 2 life.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	wolf2 := b12Creature(g2, opp2.ID, "Wolf", "Creature — Wolf", 3, 3)
	socB12ScrivEnters(t, g2, me2, wolf2)
	life := opp2.Life
	advanceToNextSeatsTurn(t, g2)
	advanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(wolf2, me2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g2)
	passPriorityAroundTable(t, g2)
	if opp2.Life != life-2 {
		t.Errorf("attacker's controller life %d -> %d, want -2", life, opp2.Life)
	}
	if p := effectivePower(t, g2, wolf2); p != 3 {
		t.Errorf("Wolf power %d attacking Scriv's controller, want 3 (no pump)", p)
	}
}

// --- Sentinel's Eyes --------------------------------------------------------

func TestSocB12SentinelsEyesPumpsGrantsVigilanceAndEscapes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Sentinel's Eyes", socB12EyesOracle, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted Bear is %d/%d, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the enchanted creature lacks vigilance")
	}
	// Escape: from the graveyard for {W} and two other graveyard cards.
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Fatal("the Aura is still on the battlefield")
	}
	if !me.Graveyard.Contains(aura) {
		g.WithWriteLock(func() {
			_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, aura)
		})
	}
	if !me.Graveyard.Contains(aura) {
		t.Fatal("the Aura is not in the graveyard")
	}
	target := auraBear(g, me.ID)
	a := gyCard(me, "Fodder A", "Creature — Bear", "{1}", "")
	b := gyCard(me, "Fodder B", "Creature — Bear", "{1}", "")
	// Without {W} the escape cast is refused.
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{FromZone: "graveyard", Strict: true, Targets: cardTargets(target)}); err == nil {
		t.Fatal("escaped for no mana")
	}
	floatForTest(g, me, "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		FromZone: "graveyard", Strict: true, Targets: cardTargets(target),
		AlternativeCost: "escape", AltCostIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the escaped Aura did not enter the battlefield")
	}
	if !inExile(g, a) || !inExile(g, b) {
		t.Error("the two other graveyard cards were not exiled as the escape cost")
	}
}

// --- Serra Paragon ----------------------------------------------------------

func socB12GraveCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: socB12VanillaOracleTag,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

func socB12ParagonTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	paragon := b12Push(g, me.ID, "Serra Paragon", "Creature — Angel", socB12ParagonOracle, 3, 4)
	return g, me, paragon
}

func TestSocB12SerraParagonOpensOneCheapPermanentFromTheGraveyardPerTurn(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	spell := socB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{2}{G}")
	second := socB12GraveCard(me, "Second Bear", "Creature — Bear", "{1}")
	land := socB12GraveCard(me, "Forest", "Basic Land — Forest", "")
	costly := socB12GraveCard(me, "Costly Bear", "Creature — Bear", "{3}{G}")
	instant := socB12GraveCard(me, "Shock", "Instant", "{R}")

	if err := muldrothaCast(g, me, costly, ""); err == nil {
		t.Error("a mana value 4 permanent was cast from the graveyard")
	}
	if err := muldrothaCast(g, me, instant, ""); err == nil {
		t.Error("an instant was cast from the graveyard")
	}
	if err := muldrothaCast(g, me, spell, ""); err != nil {
		t.Fatalf("a mana value 3 creature from the graveyard: %v", err)
	}
	resolveAllPT(t, g)
	// Once during each of your turns: the land and a second spell are refused.
	if err := muldrothaCast(g, me, land, ""); err == nil {
		t.Error("a land was played after the turn's one use was spent on a spell")
	}
	if err := muldrothaCast(g, me, second, ""); err == nil {
		t.Error("a second permanent spell was cast from the graveyard")
	}
	if enumeratedFor(g, me.ID, second) || enumeratedFor(g, me.ID, land) {
		t.Error("the enumerator still offers a second use")
	}
}

func TestSocB12SerraParagonLandPlaySpendsTheSameUse(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	land := socB12GraveCard(me, "Forest", "Basic Land — Forest", "")
	spell := socB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{1}")
	if err := muldrothaCast(g, me, land, ""); err != nil {
		t.Fatalf("a land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land is not on the battlefield")
	}
	if err := muldrothaCast(g, me, spell, ""); err == nil {
		t.Error("a permanent spell was cast after the turn's use went to a land")
	}
}

func TestSocB12SerraParagonUseRenewsEachOfYourTurnsAndNotOnOthers(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	first := socB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{1}")
	if err := muldrothaCast(g, me, first, ""); err != nil {
		t.Fatalf("first use: %v", err)
	}
	resolveAllPT(t, g)
	later := socB12GraveCard(me, "Later Bear", "Creature — Bear", "{1}")
	advanceToNextSeatsTurn(t, g)
	advanceToMain(t, g)
	if err := muldrothaCast(g, me, later, ""); err == nil {
		t.Error("cast from the graveyard on another player's turn")
	}
	advanceToMainOf(t, g, 0)
	if err := muldrothaCast(g, me, later, ""); err != nil {
		t.Errorf("the use did not renew on my next turn: %v", err)
	}
}

func TestSocB12SerraParagonSpellGainsTheExileAndGainLifeRider(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	bear := socB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{1}")
	if err := muldrothaCast(g, me, bear, ""); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveAllPT(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not enter")
	}
	life := me.Life
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !inExile(g, bear) {
		t.Error("the permanent was not exiled when it was put into a graveyard")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestSocB12SerraParagonLandGainsTheRiderToo(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	land := socB12GraveCard(me, "Forest", "Basic Land — Forest", "")
	if err := muldrothaCast(g, me, land, ""); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	life := me.Life
	killCreature(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if !inExile(g, land) {
		t.Error("the land was not exiled when it was put into a graveyard")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestSocB12SerraParagonShocklandGainsTheRiderAfterThePayLifePrompt(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	crypt := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: crypt, Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: bloodCryptOracle,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	if err := muldrothaCast(g, me, crypt, ""); err != nil {
		t.Fatalf("play the shockland: %v", err)
	}
	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the shockland did not ask for its 2 life")
	}
	answerEntryPayLife(t, g, me.ID, true)
	if !g.Battlefield.Contains(crypt) {
		t.Fatal("the shockland did not enter")
	}
	life := me.Life
	killCreature(t, g, me.ID, crypt)
	passPriorityAroundTable(t, g)
	if !inExile(g, crypt) || me.Life != life+2 {
		t.Errorf("exiled %v, life %d -> %d: a land that waited on its own entry prompt still gains the rider", inExile(g, crypt), life, me.Life)
	}
}

func TestSocB12SerraParagonPermanentNotCastWithItIsUnchanged(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", OracleID: socB12VanillaOracleTag,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	resolveAllPT(t, g)
	life := me.Life
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if inExile(g, bear) || me.Life != life {
		t.Error("a permanent cast from hand gained Serra Paragon's rider")
	}
}
