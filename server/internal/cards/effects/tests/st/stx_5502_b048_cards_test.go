package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stx_5502_b048_cards_test.go — Strixhaven: School of Mages, batch 048
// (#5502). Every test plays the real catalog entry through the engine.

const (
	stxB48ArrogantPoetOracle        = "0b2e9571-0c28-4a8f-9ba7-0147b40290a7"
	stxB48BasicConjurationOracle    = "547ab16c-b8d1-4357-9219-7cabde80f5ba"
	stxB48InfuseWithVitalityOracle  = "6799217a-030c-4c1b-9de3-99725c6ef4cf"
	stxB48ThunderousOratorOracle    = "7716ad72-9e96-4915-bc1e-e12b4d0f6f9c"
	stxB48AccomplishedAlchemistOrcl = "97e18a8d-9068-4782-a744-280d28c68228"
	stxB48HonorTrollOracle          = "98c059ed-d749-4f66-bda1-6352976198f0"
	stxB48FortifyingDraughtOracle   = "a34c78c6-ffbc-4461-a66d-1253a7aa963b"
	stxB48BlexOracle                = "b2c53124-e422-4786-8135-67c350d71a36"
	stxB48MageHunterOracle          = "e3d2df4d-8c79-4391-bc4e-69325958a5be"
)

func stxB48Pass(t *testing.T, g *game.Game) { effectstest.PassPriorityAroundTable(t, g) }

func TestStxB48EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		stxB48ArrogantPoetOracle, stxB48BasicConjurationOracle, stxB48InfuseWithVitalityOracle,
		stxB48ThunderousOratorOracle, stxB48AccomplishedAlchemistOrcl, stxB48HonorTrollOracle,
		stxB48FortifyingDraughtOracle, stxB48BlexOracle, stxB48BlexOracle + "#1", stxB48MageHunterOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- Arrogant Poet -------------------------------------------------------------

func TestStxB48ArrogantPoetPaysTwoLifeForFlying(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		poet := effectstest.B12Push(g, me.ID, "Arrogant Poet", "Creature — Human Warlock", stxB48ArrogantPoetOracle, 2, 1)
		life := me.Life
		effectstest.DeclareAttack(t, g, opp.ID, poet)
		stxB48Pass(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		stxB48Pass(t, g)
		flying := effectstest.EffectiveAbilitiesContain(t, g, poet, "flying")
		if flying != pay {
			t.Errorf("pay=%v: flying = %v", pay, flying)
		}
		wantLife := life
		if pay {
			wantLife = life - 2
		}
		if me.Life != wantLife {
			t.Errorf("pay=%v: life %d, want %d", pay, me.Life, wantLife)
		}
	}
}

func TestStxB48ArrogantPoetDoesNothingUnlessItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	poet := effectstest.B12Push(g, me.ID, "Arrogant Poet", "Creature — Human Warlock", stxB48ArrogantPoetOracle, 2, 1)
	stxB48Pass(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, poet, "flying") {
		t.Error("the Poet has flying without attacking")
	}
}

// --- Thunderous Orator -----------------------------------------------------------

func TestStxB48ThunderousOratorSharesKeywordsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orator := effectstest.B12Push(g, me.ID, "Thunderous Orator", "Creature — Kor Wizard", stxB48ThunderousOratorOracle, 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ally", TypeLine: "Creature — Bear", Power: 1, Toughness: 1,
		Keywords: []string{"flying", "lifelink", "haste"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Deathtoucher", TypeLine: "Creature — Bear", Power: 1, Toughness: 1,
		Keywords: []string{"deathtouch"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AssertKeywords(t, g, orator, "vigilance")
	effectstest.DeclareAttack(t, g, opp.ID, orator)
	stxB48Pass(t, g)
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, orator, kw) {
			t.Errorf("the Orator did not gain %s", kw)
		}
	}
	for _, kw := range []string{"deathtouch", "haste", "trample"} {
		if effectstest.EffectiveAbilitiesContain(t, g, orator, kw) {
			t.Errorf("the Orator gained %s: nobody on its side has it, or it is not on the list", kw)
		}
	}
}

func TestStxB48ThunderousOratorAloneGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orator := effectstest.B12Push(g, me.ID, "Thunderous Orator", "Creature — Kor Wizard", stxB48ThunderousOratorOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, orator)
	stxB48Pass(t, g)
	for _, kw := range []string{"flying", "first strike", "double strike", "deathtouch", "indestructible", "lifelink", "menace", "trample"} {
		if effectstest.EffectiveAbilitiesContain(t, g, orator, kw) {
			t.Errorf("the Orator alone gained %s", kw)
		}
	}
}

// --- Accomplished Alchemist --------------------------------------------------------

func TestStxB48AccomplishedAlchemistAddsOneManaOfAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	alch := effectstest.B12Push(g, me.ID, "Accomplished Alchemist", "Creature — Elf Druid", stxB48AccomplishedAlchemistOrcl, 2, 5)
	effectstest.UntapForTest(g, alch)
	effectstest.ActivateManaFor(t, g, me.ID, alch, 0, game.ManaAbilityParams{})
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no colour pick for the one-mana ability")
	}
}

func TestStxB48AccomplishedAlchemistAddsManaEqualToLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	alch := effectstest.B12Push(g, me.ID, "Accomplished Alchemist", "Creature — Elf Druid", stxB48AccomplishedAlchemistOrcl, 2, 5)
	effectstest.UntapForTest(g, alch)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.ActivateManaFor(t, g, me.ID, alch, 1, game.ManaAbilityParams{})
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["U"] != 3 {
		t.Fatalf("pick = %+v, want amounts of three", pick)
	}
}

func TestStxB48AccomplishedAlchemistMakesNothingWithoutLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	alch := effectstest.B12Push(g, me.ID, "Accomplished Alchemist", "Creature — Elf Druid", stxB48AccomplishedAlchemistOrcl, 2, 5)
	effectstest.UntapForTest(g, alch)
	// A loss is not a gain.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -4) })
	_ = g.ActivateManaAbility(me.ID, alch, 1, game.ManaAbilityParams{})
	if pick := effectstest.PendingOfKind(g, game.PendingChoiceMana); pick != nil && pick.ManaAmounts["U"] > 0 {
		t.Errorf("a turn with no life gained produced mana: %+v", pick)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want empty", me.ManaPool)
	}
}

// --- Honor Troll ----------------------------------------------------------------------

func TestStxB48HonorTrollAddsOneToEachGainAndIsBigAtTwentyFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.B12Push(g, me.ID, "Honor Troll", "Creature — Troll Druid", stxB48HonorTrollOracle, 2, 3)
	effectstest.AssertKeywords(t, g, troll, "vigilance")
	if me.Life < 25 {
		t.Fatalf("the table starts at %d life", me.Life)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, troll); p != 4 || tough != 4 {
		t.Errorf("at %d life the Troll is %d/%d, want 4/4", me.Life, p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 24-me.Life) })
	if me.Life != 24 {
		t.Fatalf("life %d, want 24 (a loss is not replaced)", me.Life)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, troll); p != 2 || tough != 3 {
		t.Errorf("at 24 life the Troll is %d/%d, want 2/3", p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	if me.Life != 27 {
		t.Errorf("gaining 2 from 24: life %d, want 27", me.Life)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, troll); p != 4 || tough != 4 {
		t.Errorf("at %d life the Troll is %d/%d, want 4/4", me.Life, p, tough)
	}
}

func TestStxB48HonorTrollIgnoresAnOpponentsLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Honor Troll", "Creature — Troll Druid", stxB48HonorTrollOracle, 2, 3)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 2) })
	if opp.Life != life+2 {
		t.Errorf("an opponent's gain of 2 became %d", opp.Life-life)
	}
}

// --- Fortifying Draught ----------------------------------------------------------------

func TestStxB48FortifyingDraughtCountsTheLifeGainedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Fortifying Draught", "Instant", stxB48FortifyingDraughtOracle, effectstest.CardTarget(bear))
	stxB48Pass(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the Bear is %d/%d, want 7/7 (3 earlier plus its own 2)", p, tough)
	}
}

func TestStxB48FortifyingDraughtReadsLifeAfterAReplacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Honor Troll", "Creature — Troll Druid", stxB48HonorTrollOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fortifying Draught", "Instant", stxB48FortifyingDraughtOracle, effectstest.CardTarget(bear))
	stxB48Pass(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5 (the gain of 2 became 3)", p, tough)
	}
}

func TestStxB48FortifyingDraughtNeedsALegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Fortifying Draught", "Instant", stxB48FortifyingDraughtOracle,
		[]game.TargetRef{effectstest.PlayerTarget(me.ID)})
	if err == nil {
		t.Error("the Draught was cast at a player")
	}
	if me.Life != life {
		t.Errorf("life changed to %d without a resolved spell", me.Life)
	}
}

// --- Infuse with Vitality ---------------------------------------------------------------

func TestStxB48InfuseWithVitalityGivesDeathtouchAndAOneShotReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Infuse with Vitality", "Instant", stxB48InfuseWithVitalityOracle, effectstest.CardTarget(bear))
	stxB48Pass(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the target has no deathtouch")
	}
	effectstest.B18Kill(t, g, bear)
	back := effectstest.FindBattlefieldCardByID(g, bear)
	if back == nil {
		t.Fatal("the creature died and did not come back")
	}
	if !back.Tapped {
		t.Error("it returns tapped")
	}
	if effectstest.CounterCount(g, bear, game.CounterPlusOne) != 0 {
		t.Error("it returned with a counter the card does not give")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the returned creature is a new object and keeps deathtouch")
	}
	effectstest.B18Kill(t, g, bear)
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("the second death is final")
	}
}

func TestStxB48InfuseWithVitalityEndsAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Infuse with Vitality", "Instant", stxB48InfuseWithVitalityOracle, effectstest.CardTarget(bear))
	stxB48Pass(t, g)
	effectstest.AdvancePastCleanupForTest(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("deathtouch outlived the turn")
	}
	effectstest.B18Kill(t, g, bear)
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("after the turn, the creature stays dead")
	}
}

func TestStxB48InfuseWithVitalityNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Infuse with Vitality", "Instant", stxB48InfuseWithVitalityOracle,
		[]game.TargetRef{effectstest.PlayerTarget(me.ID)})
	if err == nil {
		t.Error("Infuse with Vitality was cast at a player")
	}
}

// --- Mage Hunter ------------------------------------------------------------------------------

func TestStxB48MageHunterDrainsAnOpponentWhoCastsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mage Hunter", "Creature — Horror", stxB48MageHunterOracle, 3, 4)
	life, mine := opp.Life, me.Life
	effectstest.B09OpponentCastsCard(t, g, opp, game.Card{Name: "Their Instant", TypeLine: "Instant"})
	stxB48Pass(t, g)
	if opp.Life != life-1 {
		t.Errorf("after an instant the caster is at %d, want %d", opp.Life, life-1)
	}
	effectstest.B09OpponentCastsCard(t, g, opp, game.Card{Name: "Their Sorcery", TypeLine: "Sorcery", Keywords: []string{"flash"}})
	stxB48Pass(t, g)
	if opp.Life != life-2 {
		t.Errorf("after a sorcery the caster is at %d, want %d", opp.Life, life-2)
	}
	effectstest.B09OpponentCastsCard(t, g, opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"flash"}})
	stxB48Pass(t, g)
	if opp.Life != life-2 {
		t.Errorf("a creature spell cost the caster life: %d, want %d", opp.Life, life-2)
	}
	if me.Life != mine {
		t.Errorf("the Mage Hunter's controller is at %d, want %d", me.Life, mine)
	}
}

func TestStxB48MageHunterIgnoresYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mage Hunter", "Creature — Horror", stxB48MageHunterOracle, 3, 4)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{effectstest.PlayerTarget(g.Seats[1].ID)})
	stxB48Pass(t, g)
	if me.Life != life {
		t.Errorf("your own instant cost you life: %d, want %d", me.Life, life)
	}
}

func TestStxB48MageHunterTriggersOnAnOpponentsCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mage Hunter", "Creature — Horror", stxB48MageHunterOracle, 3, 4)
	hunter := g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID
	spell := uuid.New()
	g.Stack.PushTop(game.Card{InstanceID: spell, Name: "Copied Sorcery", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	life := opp.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSpellCopied, CardID: spell, Actor: opp.ID})
	})
	stxB48Pass(t, g)
	_ = hunter
	if opp.Life != life-1 {
		t.Errorf("an opponent's copy cost them %d life, want 1", life-opp.Life)
	}
}

// --- Basic Conjuration -------------------------------------------------------------------------

func TestStxB48BasicConjurationTakesACreatureAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Wurm", TypeLine: "Artifact Creature — Wurm"},
		game.Card{Name: "E"}, game.Card{Name: "F"},
		game.Card{Name: "Seventh", TypeLine: "Instant"},
	)
	life, lib := me.Life, me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Basic Conjuration", "Sorcery — Lesson", stxB48BasicConjurationOracle, nil)
	stxB48Pass(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature-card prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("prompt bounds = %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	for i, want := range []bool{false, true, false, true, false, false} {
		if offered[ids[i]] != want {
			t.Errorf("card %d offered = %v, want %v (creature cards only)", i, offered[ids[i]], want)
		}
	}
	if offered[ids[6]] {
		t.Error("the seventh card is beyond the top six")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen creature did not reach the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library = %d, want %d", me.Library.Size(), lib-1)
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[6] {
		t.Error("the seventh card must now be on top; the other five went under it")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestStxB48BasicConjurationGainsLifeEvenIfYouTakeNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "A", TypeLine: "Instant"}, game.Card{Name: "B", TypeLine: "Instant"}, game.Card{Name: "C", TypeLine: "Instant"},
		game.Card{Name: "D", TypeLine: "Instant"}, game.Card{Name: "E", TypeLine: "Instant"}, game.Card{Name: "F", TypeLine: "Instant"},
	)
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Basic Conjuration", "Sorcery — Lesson", stxB48BasicConjurationOracle, nil)
	stxB48Pass(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("a prompt was offered with no creature among the six")
	}
	if me.Hand.Size() != hand || me.Life != life+3 {
		t.Errorf("hand %d -> %d, life %d -> %d: want no card and +3 life", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- Blex, Vexing Pest // Search for Blex ---------------------------------------------------------

func stxB48BlexCard(owner uuid.UUID) game.Card {
	return effectstest.MdfcCard(owner, stxB48BlexOracle, game.LayoutModalDFC,
		game.Face{Name: "Blex, Vexing Pest", TypeLine: "Legendary Creature — Pest", ManaCost: "{2}{G}",
			Power: 3, Toughness: 2, Colors: []string{"G"}},
		game.Face{Name: "Search for Blex", TypeLine: "Sorcery", ManaCost: "{2}{B}{B}", Colors: []string{"B"}},
	)
}

func TestStxB48BlexPumpsOtherPestsBatsInsectsSnakesAndSpiders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blex := stxB48BlexCard(me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, blex)
	byType := map[string]uuid.UUID{}
	for _, sub := range []string{"Pest", "Bat", "Insect", "Snake", "Spider", "Bear"} {
		byType[sub] = effectstest.B12Creature(g, me.ID, sub, "Creature — "+sub, 2, 2)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Pest", "Creature — Pest", 2, 2)
	if p, tough := effectstest.PtOf(t, g, blex.InstanceID); p != 3 || tough != 2 {
		t.Errorf("Blex is %d/%d, want 3/2 (the lord says other)", p, tough)
	}
	for sub, id := range byType {
		want := 3
		if sub == "Bear" {
			want = 2
		}
		if p, tough := effectstest.PtOf(t, g, id); p != want || tough != want {
			t.Errorf("%s is %d/%d, want %d/%d", sub, p, tough, want, want)
		}
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Pest is %d/%d, want 2/2", p, tough)
	}
}

func TestStxB48BlexGainsFourLifeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blex := stxB48BlexCard(me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, blex)
	life := me.Life
	effectstest.B18Kill(t, g, blex.InstanceID)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
	if !effectstest.InGraveyardOf(g, me.ID, blex.InstanceID) {
		t.Error("Blex is not in the graveyard")
	}
}

func TestStxB48BlexExileDoesNotGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blex := stxB48BlexCard(me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, blex)
	life := me.Life
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(blex.InstanceID) })
	stxB48Pass(t, g)
	if me.Life != life {
		t.Errorf("an exiled Blex gained life: %d, want %d", me.Life, life)
	}
}

func stxB48CastSearchForBlex(t *testing.T, g *game.Game, me *game.Player, strict bool, mana int) error {
	t.Helper()
	card := stxB48BlexCard(me.ID)
	card.SetFace(1)
	me.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < mana; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	}
	return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Strict: strict})
}

func TestStxB48SearchForBlexTakesAnyNumberAndLosesThreeEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "A"}, game.Card{Name: "B"}, game.Card{Name: "C"}, game.Card{Name: "D"}, game.Card{Name: "E"},
		game.Card{Name: "Sixth"},
	)
	life := me.Life
	if err := stxB48CastSearchForBlex(t, g, me, true, 4); err != nil {
		t.Fatalf("cast Search for Blex: %v", err)
	}
	stxB48Pass(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 5 || len(pick.ChooseCards) != 5 {
		t.Errorf("prompt = %d..%d over %d cards, want 0..5 over 5", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) {
		t.Error("the chosen cards did not reach the hand")
	}
	for _, i := range []int{1, 3, 4} {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("card %d did not go to the graveyard", i)
		}
	}
	if me.Library.Size() == 0 || me.Library.Cards[me.Library.Size()-1].InstanceID != ids[5] {
		t.Error("the sixth card must still be on top of the library")
	}
	if me.Life != life-6 {
		t.Errorf("life %d, want %d (3 for each of two cards)", me.Life, life-6)
	}
}

func TestStxB48SearchForBlexTakingNothingCostsNoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "A"}, game.Card{Name: "B"}, game.Card{Name: "C"}, game.Card{Name: "D"}, game.Card{Name: "E"},
	)
	life := me.Life
	if err := stxB48CastSearchForBlex(t, g, me, true, 4); err != nil {
		t.Fatalf("cast Search for Blex: %v", err)
	}
	stxB48Pass(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for i, id := range ids {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %d did not go to the graveyard", i)
		}
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

func TestStxB48SearchForBlexCantBeCastWithoutItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := stxB48CastSearchForBlex(t, g, me, true, 3); err == nil {
		t.Error("Search for Blex was cast with three mana")
	}
}
