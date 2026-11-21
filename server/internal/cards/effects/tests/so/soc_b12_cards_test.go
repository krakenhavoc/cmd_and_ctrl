package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSocB12EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.SocB12BullOracle:      "Renegade Bull",
		effectstest.SocB12TechniqueOracle: "Replication Technique",
		effectstest.SocB12RoasterOracle:   "Ribtruss Roaster",
		effectstest.SocB12RionyaOracle:    "Rionya, Fire Dancer",
		effectstest.SocB12RoothaOracle:    "Rootha, Mastering the Moment",
		effectstest.SocB12RefrainOracle:   "Rousing Refrain",
		effectstest.SocB12ScreamsOracle:   "Screams from Within",
		effectstest.SocB12ScrivOracle:     "Scriv, the Obligator",
		effectstest.SocB12EyesOracle:      "Sentinel's Eyes",
		effectstest.SocB12ParagonOracle:   "Serra Paragon",
	}
	for oracle, name := range want {
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

// --- Renegade Bull ----------------------------------------------------------

func TestSocB12RenegadeBullGetsPlusXForAnInstantOrSorceryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bull := effectstest.B12Push(g, me.ID, "Renegade Bull", "Creature — Ox", effectstest.SocB12BullOracle, 0, 5)
	if !effectstest.HasEffectiveKeyword(t, g, bull, "trample") {
		t.Error("the Bull lacks trample")
	}
	effectstest.SocB4Spell(t, g, "Big Sorcery", "Sorcery", "{3}{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bull); p != 4 {
		t.Fatalf("power %d after a mana value 4 sorcery, want 4", p)
	}
	effectstest.SocB4Spell(t, g, "Zap", "Instant", "{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bull); p != 5 {
		t.Fatalf("power %d after a mana value 1 instant too, want 5", p)
	}
	effectstest.SocB4Spell(t, g, "Grizzly", "Creature — Bear", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bull); p != 5 {
		t.Errorf("power %d after a creature spell, want 5 (it does not trigger)", p)
	}
}

func TestSocB12RenegadeBullCountsTheXOfAnXSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bull := effectstest.B12Push(g, me.ID, "Renegade Bull", "Creature — Ox", effectstest.SocB12BullOracle, 0, 5)
	effectstest.SocB4Spell(t, g, "X Spell", "Sorcery", "{X}{R}", game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bull); p != 4 {
		t.Errorf("power %d after an X=3 sorcery, want 4 (X counts in its mana value)", p)
	}
}

func TestSocB12RenegadeBullAttackCopiesAndCastsAnInstantForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bull := effectstest.B12Push(g, me.ID, "Renegade Bull", "Creature — Ox", effectstest.SocB12BullOracle, 0, 5)
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	life := opp.Life

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bull, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the targeted instant is not exiled")
	}
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy without paying: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 from the free Bolt", life, opp.Life)
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left over after the copy resolved", n)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the original card stays in exile")
	}
}

func TestSocB12RenegadeBullAttackWithNoTargetOrADeclinedCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bull := effectstest.B12Push(g, me.ID, "Renegade Bull", "Creature — Ox", effectstest.SocB12BullOracle, 0, 5)
	// A creature card is not a legal target.
	effectstest.GyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	// Declined: the card is exiled, the copy is not cast and ceases to exist.
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bull, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclineResolutionCasts(t, g, me.ID)
	if !effectstest.InExile(g, bolt) {
		t.Error("the card is exiled whether or not the copy is cast")
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left after declining", n)
	}
	if opp.Life != life {
		t.Errorf("life moved to %d", opp.Life)
	}
}

// --- Replication Technique --------------------------------------------------

func socB12CastTechnique(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Replication Technique", "Sorcery", "{4}{U}", effectstest.SocB12TechniqueOracle,
		game.CastSpellParams{Targets: effectstest.CardTargets(target)})
}

func TestSocB12ReplicationTechniqueCopiesATargetPermanentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	socB12CastTechnique(t, g, me, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Bear", me.ID); n != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	socB12CastTechnique(t, g, me, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 12; i++ {
		switch {
		case effectstest.CopyPromptFor(g, me.ID) != nil:
			c := effectstest.CopyPromptFor(g, me.ID)
			if err := g.ResolvePickTargets(c.ID, me.ID, effectstest.CardTargets(mine)); err != nil {
				t.Fatalf("my copy keeps its target: %v", err)
			}
		case effectstest.CopyPromptFor(g, opp.ID) != nil:
			c := effectstest.CopyPromptFor(g, opp.ID)
			if err := g.ResolvePickTargets(c.ID, opp.ID, effectstest.CardTargets(theirs)); err != nil {
				t.Fatalf("the opponent's copy picks their creature: %v", err)
			}
		case effectstest.LatestOptionPickFor(g, me.ID) != nil:
			effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
		default:
			i = 12
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Bear", me.ID); n != 3 {
		t.Errorf("%d Bears for me, want the original and two token copies (the original and my copy)", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Wolf", opp.ID); n != 2 {
		t.Errorf("%d Wolves for the opponent, want theirs and one token copy from their copy of the spell", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Wolf", me.ID); n != 0 {
		t.Errorf("the opponent's copy made a Wolf for me")
	}
}

// --- Ribtruss Roaster -------------------------------------------------------

func TestSocB12RibtrussRoasterDevoursAndMakesAPestPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	effectstest.PushDevourFodder(g, me.ID, "Goblin B")
	id := effectstest.CastDevourer(t, g, "Ribtruss Roaster", effectstest.SocB12RoasterOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a)
	if got := effectstest.DevourPlusOnes(g, id); got != 1 {
		t.Fatalf("counters = %d, want 1 (devour 1, one creature)", got)
	}
	// Another counter from elsewhere counts too: "equal to the number of
	// +1/+1 counters on this creature".
	socB9Counter(t, g, id, game.CounterPlusOne, 1)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Pest", me.ID); got != 2 {
		t.Fatalf("%d Pests, want 2 (one per +1/+1 counter)", got)
	}
	// Each Pest gains 1 life when it dies.
	pest := effectstest.FindBattlefieldByName(g, "Pest")
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, pest)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 from the Pest dying", life, me.Life)
	}
}

func TestSocB12RibtrussRoasterWithoutCountersMakesNothingAndOtherEndStepsDont(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ribtruss Roaster", "Creature — Troll Druid", effectstest.SocB12RoasterOracle, 3, 3)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Pest", me.ID); got != 0 {
		t.Errorf("%d Pests with no counters, want 0", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	id := effectstest.B12Push(g2, me2.ID, "Ribtruss Roaster", "Creature — Troll Druid", effectstest.SocB12RoasterOracle, 3, 3)
	socB9Counter(t, g2, id, game.CounterPlusOne, 2)
	effectstest.AdvanceToEndStep(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CountOnBattlefield(g2, "Pest", me2.ID); got != 2 {
		t.Fatalf("%d Pests at my end step, want 2", got)
	}
	effectstest.AdvanceToEndStepOf(t, g2, 1)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CountOnBattlefield(g2, "Pest", me2.ID); got != 2 {
		t.Errorf("%d Pests after another player's end step, want still 2", got)
	}
}

// --- Rionya, Fire Dancer ----------------------------------------------------

func TestSocB12RionyaCopiesAnotherCreatureOncePlusOnePerInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rionya := effectstest.B12Push(g, me.ID, "Rionya, Fire Dancer", "Legendary Creature — Human Wizard", effectstest.SocB12RionyaOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SocB4Spell(t, g, "Zap", "Instant", "{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SocB4Spell(t, g, "Bolt", "Sorcery", "{1}{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// A creature spell does not count.
	effectstest.SocB4Spell(t, g, "Grizzly", "Creature — Bear", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if c := effectstest.LatestPickTarget(g, me.ID); c == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardTargets(rionya)); err == nil {
		t.Fatal("Rionya was a legal target for her own \"another target creature\"")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
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
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Error("a copy lacks haste")
		}
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Bear", me.ID); n != 1 {
		t.Errorf("%d Bears after the end step, want only the original (the copies are exiled)", n)
	}
}

func TestSocB12RionyaMakesOneCopyWithNoSpellsCastAndOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Rionya, Fire Dancer", "Legendary Creature — Human Wizard", effectstest.SocB12RionyaOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Bear", me.ID); n != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Rootha, Mastering the Moment", "Legendary Creature — Orc Sorcerer", effectstest.SocB12RoothaOracle, 3, 4)
	effectstest.SocB4Spell(t, g, "Zap", "Instant", "{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SocB4Spell(t, g, "Four", "Sorcery", "{3}{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SocB4Spell(t, g, "Big Creature", "Creature — Dragon", "{5}{R}{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	els := socB12Elementals(g, me.ID)
	if len(els) != 1 {
		t.Fatalf("%d Elementals, want 1", len(els))
	}
	el := els[0]
	if p, tg := effectstest.EffectivePower(t, g, el), effectstest.EffectiveToughness(t, g, el); p != 4 || tg != 4 {
		t.Errorf("Elemental is %d/%d, want 4/4 (the sorcery's mana value, the creature spell does not count)", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, el, "flying") || !effectstest.HasEffectiveKeyword(t, g, el, "haste") {
		t.Error("the Elemental lacks flying or haste")
	}
	if c, _ := effectstest.BattlefieldCard(g, el); len(c.Colors) != 2 {
		t.Errorf("colours %v, want blue and red", c.Colors)
	}
}

func TestSocB12RoothaCountsXAndNeedsASpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Rootha, Mastering the Moment", "Legendary Creature — Orc Sorcerer", effectstest.SocB12RoothaOracle, 3, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB12Elementals(g, me.ID)); n != 0 {
		t.Fatalf("%d Elementals with no instant or sorcery cast, want 0", n)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	effectstest.B12Push(g2, me2.ID, "Rootha, Mastering the Moment", "Legendary Creature — Orc Sorcerer", effectstest.SocB12RoothaOracle, 3, 4)
	effectstest.SocB4Spell(t, g2, "X Spell", "Sorcery", "{X}{R}", game.CastSpellParams{XValue: 4})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	els := socB12Elementals(g2, me2.ID)
	if len(els) != 1 {
		t.Fatalf("%d Elementals, want 1", len(els))
	}
	if p := effectstest.EffectivePower(t, g2, els[0]); p != 5 {
		t.Errorf("Elemental power %d, want 5 (X=4 plus the {R} in its mana value)", p)
	}
}

// --- Rousing Refrain --------------------------------------------------------

func TestSocB12RousingRefrainAddsRedPerCardInHandKeepsItAndSuspendsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cards := opp.Hand.Size()
	if cards == 0 {
		t.Fatal("the opponent has an empty hand")
	}
	id := effectstest.CastFromHandForTest(t, g, me, "Rousing Refrain", "Sorcery", "{3}{R}{R}", effectstest.SocB12RefrainOracle,
		game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != cards {
		t.Fatalf("mana pool holds %d, want %d (one {R} per card in the opponent's hand)", got, cards)
	}
	if !effectstest.InExile(g, id) {
		t.Fatal("Rousing Refrain is not exiled")
	}
	if n := effectstest.CounterCountInExile(g, id, game.CounterTime); n != 3 {
		t.Fatalf("%d time counters, want 3", n)
	}
	// The mana is not lost as steps and phases end.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if got := me.ManaPool.Total(); got != cards {
		t.Errorf("mana pool holds %d at the beginning of combat, want %d kept", got, cards)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCountInExile(g, id, game.CounterTime); n != 2 {
		t.Errorf("%d time counters after one upkeep, want 2 (suspend counts down)", n)
	}
}

func TestSocB12RousingRefrainCanBeSuspendedFromHand(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.SocB12RefrainOracle)
	if len(spec.SpecialActions) != 1 {
		t.Fatalf("%d special actions, want suspend", len(spec.SpecialActions))
	}
	sa := spec.SpecialActions[0]
	if sa.Kind != game.SpecialActionSuspend || sa.Counters != 3 || sa.Cost != "{1}{R}" {
		t.Errorf("special action %+v, want suspend 3—{1}{R}", sa)
	}
}

// --- Screams from Within ----------------------------------------------------

func TestSocB12ScreamsFromWithinShrinksAndReturnsWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Screams from Within", effectstest.SocB12ScreamsOracle, wolf)
	if p, tg := effectstest.EffectivePower(t, g, wolf), effectstest.EffectiveToughness(t, g, wolf); p != 1 || tg != 1 {
		t.Fatalf("enchanted Wolf is %d/%d, want 1/1", p, tg)
	}
	effectstest.KillCreature(t, g, opp.ID, wolf)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		effectstest.AnswerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("Screams from Within did not return to the battlefield")
	}
	if host := effectstest.AttachmentHostOf(t, g, aura); host.ID != bear {
		t.Errorf("it enchants %v, want the creature its controller chose", host.ID)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 1 {
		t.Errorf("the new host has power %d, want 1", p)
	}
}

func TestSocB12ScreamsFromWithinStaysWhenNothingCanBeEnchantedOrItLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Screams from Within", effectstest.SocB12ScreamsOracle, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !me.Graveyard.Contains(aura) {
		t.Error("with no creature left to enchant the Aura stays in the graveyard")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	bear2 := effectstest.AuraBear(g2, me2.ID)
	effectstest.AuraBear(g2, me2.ID)
	aura2 := effectstest.AuraCast(t, g2, "Screams from Within", effectstest.SocB12ScreamsOracle, bear2)
	effectstest.KillCreature(t, g2, me2.ID, bear2)
	// Exile the Aura in response to the trigger: it is no longer in the graveyard.
	if err := g2.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me2.ID}, game.ZoneRef{Kind: game.ZoneExile}, aura2); err != nil {
		t.Fatalf("exile the Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(aura2) || effectstest.AuraHostChoiceFor(g2, me2.ID) != nil {
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
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Scriv, the Obligator", "Legendary Creature — Inkling Bird", "{2}{W}{B}", effectstest.SocB12ScrivOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, target)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB12ScrivMakesAContractAttachedToAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	socB12ScrivEnters(t, g, me, wolf)
	contracts := socB12Contracts(g)
	if len(contracts) != 1 {
		t.Fatalf("%d Contracts, want 1", len(contracts))
	}
	c := contracts[0]
	if !c.IsToken() || !c.IsAura() || c.Controller != me.ID || len(c.Colors) != 1 || c.Colors[0] != "W" {
		t.Errorf("Contract %+v, want a white Aura token under Scriv's controller", c)
	}
	if host := effectstest.AttachmentHostOf(t, g, c.InstanceID); host.ID != wolf {
		t.Errorf("Contract is attached to %v, want the opponent's Wolf", host.ID)
	}
	if text := game.TokenTextForCard(c); text == "" {
		t.Error("the Contract has no printed text")
	}
	_ = mine
}

func TestSocB12ScrivContractPumpsAnAttackOnAnotherOpponentAndCostsLifeOtherwise(t *testing.T) {
	// An attack on one of Scriv's controller's OTHER opponents: +2/+0.
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 3, 3)
	socB12ScrivEnters(t, g, me, wolf)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolf, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, wolf); p != 5 {
		t.Errorf("Wolf power %d attacking another of Scriv's controller's opponents, want 5", p)
	}
	if opp.Life != 40 {
		t.Errorf("opponent life %d, want 40 untouched", opp.Life)
	}

	// An attack on Scriv's controller: its controller loses 2 life.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	wolf2 := effectstest.B12Creature(g2, opp2.ID, "Wolf", "Creature — Wolf", 3, 3)
	socB12ScrivEnters(t, g2, me2, wolf2)
	life := opp2.Life
	effectstest.AdvanceToNextSeatsTurn(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(wolf2, me2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if opp2.Life != life-2 {
		t.Errorf("attacker's controller life %d -> %d, want -2", life, opp2.Life)
	}
	if p := effectstest.EffectivePower(t, g2, wolf2); p != 3 {
		t.Errorf("Wolf power %d attacking Scriv's controller, want 3 (no pump)", p)
	}
}

// --- Sentinel's Eyes --------------------------------------------------------

func TestSocB12SentinelsEyesPumpsGrantsVigilanceAndEscapes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Sentinel's Eyes", effectstest.SocB12EyesOracle, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted Bear is %d/%d, want 3/3", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the enchanted creature lacks vigilance")
	}
	// Escape: from the graveyard for {W} and two other graveyard cards.
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
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
	target := effectstest.AuraBear(g, me.ID)
	a := effectstest.GyCard(me, "Fodder A", "Creature — Bear", "{1}", "")
	b := effectstest.GyCard(me, "Fodder B", "Creature — Bear", "{1}", "")
	// Without {W} the escape cast is refused.
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{FromZone: "graveyard", Strict: true, Targets: effectstest.CardTargets(target)}); err == nil {
		t.Fatal("escaped for no mana")
	}
	effectstest.FloatForTest(g, me, "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		FromZone: "graveyard", Strict: true, Targets: effectstest.CardTargets(target),
		AlternativeCost: "escape", AltCostIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the escaped Aura did not enter the battlefield")
	}
	if !effectstest.InExile(g, a) || !effectstest.InExile(g, b) {
		t.Error("the two other graveyard cards were not exiled as the escape cost")
	}
}

func socB12ParagonTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	paragon := effectstest.B12Push(g, me.ID, "Serra Paragon", "Creature — Angel", effectstest.SocB12ParagonOracle, 3, 4)
	return g, me, paragon
}

func TestSocB12SerraParagonOpensOneCheapPermanentFromTheGraveyardPerTurn(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	spell := effectstest.SocB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{2}{G}")
	second := effectstest.SocB12GraveCard(me, "Second Bear", "Creature — Bear", "{1}")
	land := effectstest.SocB12GraveCard(me, "Forest", "Basic Land — Forest", "")
	costly := effectstest.SocB12GraveCard(me, "Costly Bear", "Creature — Bear", "{3}{G}")
	instant := effectstest.SocB12GraveCard(me, "Shock", "Instant", "{R}")

	if err := effectstest.MuldrothaCast(g, me, costly, ""); err == nil {
		t.Error("a mana value 4 permanent was cast from the graveyard")
	}
	if err := effectstest.MuldrothaCast(g, me, instant, ""); err == nil {
		t.Error("an instant was cast from the graveyard")
	}
	if err := effectstest.MuldrothaCast(g, me, spell, ""); err != nil {
		t.Fatalf("a mana value 3 creature from the graveyard: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	// Once during each of your turns: the land and a second spell are refused.
	if err := effectstest.MuldrothaCast(g, me, land, ""); err == nil {
		t.Error("a land was played after the turn's one use was spent on a spell")
	}
	if err := effectstest.MuldrothaCast(g, me, second, ""); err == nil {
		t.Error("a second permanent spell was cast from the graveyard")
	}
	if effectstest.EnumeratedFor(g, me.ID, second) || effectstest.EnumeratedFor(g, me.ID, land) {
		t.Error("the enumerator still offers a second use")
	}
}

func TestSocB12SerraParagonLandPlaySpendsTheSameUse(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	land := effectstest.SocB12GraveCard(me, "Forest", "Basic Land — Forest", "")
	spell := effectstest.SocB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{1}")
	if err := effectstest.MuldrothaCast(g, me, land, ""); err != nil {
		t.Fatalf("a land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land is not on the battlefield")
	}
	if err := effectstest.MuldrothaCast(g, me, spell, ""); err == nil {
		t.Error("a permanent spell was cast after the turn's use went to a land")
	}
}

func TestSocB12SerraParagonUseRenewsEachOfYourTurnsAndNotOnOthers(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	first := effectstest.SocB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{1}")
	if err := effectstest.MuldrothaCast(g, me, first, ""); err != nil {
		t.Fatalf("first use: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	later := effectstest.SocB12GraveCard(me, "Later Bear", "Creature — Bear", "{1}")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.MuldrothaCast(g, me, later, ""); err == nil {
		t.Error("cast from the graveyard on another player's turn")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	if err := effectstest.MuldrothaCast(g, me, later, ""); err != nil {
		t.Errorf("the use did not renew on my next turn: %v", err)
	}
}

func TestSocB12SerraParagonSpellGainsTheExileAndGainLifeRider(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	bear := effectstest.SocB12GraveCard(me, "Cheap Bear", "Creature — Bear", "{1}")
	if err := effectstest.MuldrothaCast(g, me, bear, ""); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not enter")
	}
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bear) {
		t.Error("the permanent was not exiled when it was put into a graveyard")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestSocB12SerraParagonLandGainsTheRiderToo(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	land := effectstest.SocB12GraveCard(me, "Forest", "Basic Land — Forest", "")
	if err := effectstest.MuldrothaCast(g, me, land, ""); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, land) {
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
		InstanceID: crypt, Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: effectstest.BloodCryptOracle,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	if err := effectstest.MuldrothaCast(g, me, crypt, ""); err != nil {
		t.Fatalf("play the shockland: %v", err)
	}
	if effectstest.EntryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the shockland did not ask for its 2 life")
	}
	effectstest.AnswerEntryPayLife(t, g, me.ID, true)
	if !g.Battlefield.Contains(crypt) {
		t.Fatal("the shockland did not enter")
	}
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, crypt)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, crypt) || me.Life != life+2 {
		t.Errorf("exiled %v, life %d -> %d: a land that waited on its own entry prompt still gains the rider", effectstest.InExile(g, crypt), life, me.Life)
	}
}

func TestSocB12SerraParagonPermanentNotCastWithItIsUnchanged(t *testing.T) {
	g, me, _ := socB12ParagonTable(t)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", OracleID: effectstest.SocB12VanillaOracleTag,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InExile(g, bear) || me.Life != life {
		t.Error("a permanent cast from hand gained Serra Paragon's rider")
	}
}
