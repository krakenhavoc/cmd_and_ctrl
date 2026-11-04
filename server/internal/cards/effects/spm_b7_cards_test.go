package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b7_cards_test.go — Marvel's Spider-Man, batch 7 (#4460). Each test plays
// the real catalog entry and asserts what its text says.

const (
	spmB7StingOracle     = "7f5e5aa8-3f15-45c9-bfc8-1f4b98d2be38"
	spmB7ScorpionOracle  = "c357230a-afde-4c2a-b390-9f6928599236"
	spmB7SecretOracle    = "e6637f9f-8293-43a5-a134-93799a016ecb"
	spmB7GoblinOracle    = "f343b6f9-41c3-419b-a57c-a7b08603ef5c"
	spmB7ShockerOracle   = "564dac59-97f7-4988-9f30-d82a891e29ef"
	spmB7ShriekOracle    = "5773ae05-44bc-423c-846c-f323dd8a5079"
	spmB7SilkOracle      = "50477dc8-9a2a-41eb-984d-c9e1e78e08a2"
	spmB7SableOracle     = "40b82dee-98c7-4b86-88b8-28388cd0331f"
	spmB7HideoutOracle   = "80f08b47-a237-4efd-8d86-dfe35a816b0e"
	spmB7SkywardOracle   = "d6ba76fa-441c-4a74-bfea-8ae2a2c9e391"
	spmB7ScorpionType    = "Legendary Creature — Scorpion Human Villain"
	spmB7ShockerType     = "Legendary Creature — Human Rogue Villain"
	spmB7ShriekType      = "Legendary Creature — Mutant Villain"
	spmB7SilkType        = "Legendary Creature — Spider Human Hero"
	spmB7SableType       = "Legendary Creature — Human Mercenary Hero"
	spmB7SkywardType     = "Creature — Spider Human Hero"
	spmB7BonesplitterOrc = bonesplitterOracle
)

func TestSpmB7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		spmB7StingOracle:    "Scorpion's Sting",
		spmB7ScorpionOracle: "Scorpion, Seething Striker",
		spmB7SecretOracle:   "Secret Identity",
		spmB7GoblinOracle:   "Shadow of the Goblin",
		spmB7ShockerOracle:  "Shocker, Unshakable",
		spmB7ShriekOracle:   "Shriek, Treblemaker",
		spmB7SilkOracle:     "Silk, Web Weaver",
		spmB7SableOracle:    "Silver Sable, Mercenary Leader",
		spmB7HideoutOracle:  "Sinister Hideout",
		spmB7SkywardOracle:  "Skyward Spider",
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

func spmB7Target(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// spmB7Sting casts Scorpion's Sting from hand under the strict mana gate and
// returns the cast error.
func spmB7Sting(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, mana string) error {
	t.Helper()
	id := handCardFull(me, "Scorpion's Sting", "Instant", "{1}{B}", spmB7StingOracle, []string{"B"})
	if mana != "" {
		floatMana(t, g, me, mana)
	}
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{spmB7Target(target)}})
}

// --- Scorpion's Sting -------------------------------------------------------------------------------

func TestSpmB7StingShrinksTheCreatureUntilEndOfTurnAndKillsASmallOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 3, 3)

	if err := spmB7Sting(t, g, me, big, "{C}{B}"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the 4/4 is %d/%d after -3/-3, want 1/1", p, tough)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want {1}{B} spent", me.ManaPool.Total())
	}

	if err := spmB7Sting(t, g, me, small, "{C}{B}"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 3/3 survived -3/-3")
	}

	advanceToUpkeepOf(t, g, 1)
	if p := effectivePower(t, g, big); p != 4 {
		t.Errorf("the 4/4 is power %d after the turn ended, want 4", p)
	}
}

func TestSpmB7StingRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := pushCatalogPermanent(g, opp.ID, "Plains", "Basic Land — Plains", "", false)

	if err := spmB7Sting(t, g, me, land, "{C}{B}"); err == nil {
		t.Error("Scorpion's Sting accepted a land as its target")
	}
	_ = bear
}

// --- Scorpion, Seething Striker ----------------------------------------------------------------------

func TestSpmB7ScorpionConnivesAtYourEndStepAfterAnyCreatureDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	b12Push(g, me.ID, "Scorpion, Seething Striker", spmB7ScorpionType, spmB7ScorpionOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	killCreature(t, g, opp.ID, theirs)

	advanceToEndStep(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the connive did not ask for a discard")
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[me.Hand.Size()-1].InstanceID)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 1 {
		t.Errorf("the bear has %d counters after a nonland discard, want 1", got)
	}
}

func TestSpmB7ScorpionDoesNothingWhenNoCreatureDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	scorpion := b12Push(g, me.ID, "Scorpion, Seething Striker", spmB7ScorpionType, spmB7ScorpionOracle, 3, 3)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	advanceToEndStep(t, g)
	if latestPickTarget(g, me.ID) != nil || !stackFullyEmpty(g) {
		t.Fatal("the end step trigger went on the stack with no creature having died")
	}
	if discardChoiceFor(g, me.ID) != nil || plusOneCounters(g, scorpion) != 0 {
		t.Error("Scorpion connived with no creature having died")
	}
}

func TestSpmB7ScorpionHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	scorpion := b12Push(g, g.Seats[0].ID, "Scorpion, Seething Striker", spmB7ScorpionType, spmB7ScorpionOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, scorpion, "deathtouch") {
		t.Error("Scorpion lacks deathtouch")
	}
}

// --- Secret Identity ----------------------------------------------------------------------------------

func spmB7Secret(t *testing.T, g *game.Game, mode int, target uuid.UUID) {
	t.Helper()
	castModal(t, g, "Secret Identity", "Instant", spmB7SecretOracle, []int{mode}, []game.TargetRef{spmB7Target(target)})
	passPriorityAroundTable(t, g)
}

func TestSpmB7SecretIdentityConcealMakesItAOneByOneCitizenWithHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		_ = AddCounter{Target: bear, Kind: game.CounterPlusOne, N: 1}.Apply(
			NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: bear}))
	})

	spmB7Secret(t, g, 0, bear)
	// Base 1/1 in layer 7b, the counter still lands on top.
	if p, tough := hobA4Power(t, g, bear), mscM6Toughness(g, bear); p != 2 || tough != 2 {
		t.Errorf("concealed creature is %d/%d, want 2/2 (base 1/1 plus its counter)", p, tough)
	}
	if got := effectiveSubtypes(t, g, bear); len(got) != 1 || got[0] != "Citizen" {
		t.Errorf("subtypes %v, want [Citizen]", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the concealed creature lacks hexproof")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the concealed creature has flying")
	}

	advanceToUpkeepOf(t, g, 1)
	if p := hobA4Power(t, g, bear); p != 5 {
		t.Errorf("after the turn the creature has power %d, want 5", p)
	}
	if got := effectiveSubtypes(t, g, bear); len(got) != 1 || got[0] != "Bear" {
		t.Errorf("after the turn subtypes %v, want [Bear]", got)
	}
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof lasted past the turn")
	}
}

func TestSpmB7SecretIdentityRevealMakesItAThreeByFourHeroWithFlyingAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)

	spmB7Secret(t, g, 1, bear)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("revealed creature is %d/%d, want 3/4", p, tough)
	}
	if got := effectiveSubtypes(t, g, bear); len(got) != 1 || got[0] != "Hero" {
		t.Errorf("subtypes %v, want [Hero]", got)
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the revealed creature lacks %s", kw)
		}
	}
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the revealed creature has hexproof")
	}
}

func TestSpmB7SecretIdentityRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for mode := 0; mode < 2; mode++ {
		if _, err := castModalErr(t, g, "Secret Identity", "Instant", spmB7SecretOracle, []int{mode},
			[]game.TargetRef{spmB7Target(theirs)}); err == nil {
			t.Errorf("mode %d accepted a creature its caster does not control", mode)
		}
	}
}

// --- Shadow of the Goblin ------------------------------------------------------------------------------

func TestSpmB7GoblinDiscardsThenDrawsAtYourFirstMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	pushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)
	before := me.Hand.Size()
	victim := me.Hand.Cards[0].InstanceID

	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the first main phase trigger asked for no discard")
	}
	answerDiscard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(victim) {
		t.Error("the discarded card is still in hand")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size %d, want %d (one discarded, one drawn)", got, before)
	}
}

func TestSpmB7GoblinWithAnEmptyHandDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	pushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Fatal("asked for a discard with no card in hand")
	}
	if got := me.Hand.Size(); got != 0 {
		t.Errorf("drew %d cards though nothing was discarded", got)
	}
}

func spmB7OpponentLife(g *game.Game) []int {
	return lifeOfOpponents(g)
}

func spmB7AssertPinged(t *testing.T, g *game.Game, before []int, by int, why string) {
	t.Helper()
	after := spmB7OpponentLife(g)
	for i := range after {
		if before[i]-after[i] != by {
			t.Errorf("%s: opponent %d lost %d life, want %d", why, i+1, before[i]-after[i], by)
		}
	}
}

func TestSpmB7GoblinPingsForASpellCastFromExileButNotFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)

	before := spmB7OpponentLife(g)
	castFree(t, g, me, "Hand Spell", "Instant")
	spmB7AssertPinged(t, g, before, 0, "a spell cast from hand")

	_, id := impulseExile(t, g, "Exile Spell", "Instant", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	spmB7AssertPinged(t, g, before, 1, "a spell cast from exile")
}

func TestSpmB7GoblinPingsForALandPlayedFromExileButNotFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)

	before := spmB7OpponentLife(g)
	land := mscM7Hand(me, "Hand Forest", "Basic Land — Forest")
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{}); err != nil {
		t.Fatalf("play the land from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	spmB7AssertPinged(t, g, before, 0, "a land played from hand")

	// A second land drop is needed for the exiled land.
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(me.ID, 1) })
	_, exiled := impulseExile(t, g, "Exile Forest", "Basic Land — Forest", "")
	if err := g.CastSpell(me.ID, exiled, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	spmB7AssertPinged(t, g, before, 1, "a land played from exile")
}

// --- Shocker, Unshakable -----------------------------------------------------------------------------

func TestSpmB7ShockerHitsTheCreatureAndItsController(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	life := opp.Life

	castHoldCreature(t, g, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	pickCard(t, g, g.Seats[0].ID, theirs)
	passPriorityAroundTable(t, g)
	if got := battlefieldCardPtr(t, g, theirs).DamageMarked; got != 2 {
		t.Errorf("the creature has %d damage, want 2", got)
	}
	if opp.Life != life-2 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-2)
	}
}

func TestSpmB7ShockerStillHitsTheControllerOfACreatureTheDamageKills(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	life := opp.Life

	castHoldCreature(t, g, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 1/1 survived 2 damage")
	}
	if opp.Life != life-2 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-2)
	}
}

func TestSpmB7ShockerCanTargetYourOwnCreatureAndHurtsYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	life := me.Life

	castHoldCreature(t, g, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("I am at %d, want %d", me.Life, life-2)
	}
}

func TestSpmB7ShockerHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shocker := b12Push(g, me.ID, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, shocker, "first strike") {
		t.Error("Shocker lacks first strike during his controller's turn")
	}
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, shocker, "first strike") {
		t.Error("Shocker has first strike during an opponent's turn")
	}
}

// --- Shriek, Treblemaker -----------------------------------------------------------------------------

// spmB7ShriekPrompt walks priority until the discard question is open.
func spmB7ShriekPrompt(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == me.ID {
			return c
		}
	}
	return nil
}

func TestSpmB7ShriekDiscardMakesACreatureUnableToBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	card := me.Hand.Cards[0].InstanceID

	advanceToMain(t, g)
	c := spmB7ShriekPrompt(t, g, me)
	if c == nil {
		t.Fatal("no discard question at the first main phase")
	}
	if err := g.ResolvePayUnlessWithOption(c.ID, me.ID, true, nil, []uuid.UUID{card}); err != nil {
		t.Fatalf("discarding: %v", err)
	}
	if !me.Graveyard.Contains(card) {
		t.Fatal("the card was not discarded")
	}
	pick := spmB7Target(theirs)
	assistCardSettle(t, g, me.ID, &pick, false)
	assertRestrictions(t, g, theirs, game.CantBlock)
}

func TestSpmB7ShriekDecliningDiscardsNothingAndRestrictsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	before := me.Hand.Size()

	advanceToMain(t, g)
	if spmB7ShriekPrompt(t, g, me) == nil {
		t.Fatal("no discard question at the first main phase")
	}
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("hand is %d cards, want %d", me.Hand.Size(), before)
	}
	if restrictionsOf(t, g, theirs)&game.CantBlock != 0 {
		t.Error("a creature can't block though no card was discarded")
	}
}

func TestSpmB7ShriekAsksNothingWithAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("asked to discard with no card in hand")
	}
}

func TestSpmB7ShriekSonicBlastHitsThePlayerWhoLostTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life

	killCreature(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("my creature dying cost lives: me %d->%d, opponent %d->%d", myLife, me.Life, oppLife, opp.Life)
	}
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-1)
	}
	if me.Life != myLife {
		t.Errorf("I lost life: %d", me.Life)
	}
}

// --- Silk, Web Weaver ---------------------------------------------------------------------------------

func TestSpmB7SilkWebSlingsForOneGreenWhiteAndATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	silk := handCardFull(me, "Silk, Web Weaver", spmB7SilkType, "{2}{G}{W}", spmB7SilkOracle, []string{"G", "W"})
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	floatMana(t, g, me, "{C}{G}{W}")

	if err := g.CastSpell(me.ID, silk, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(silk) || !me.Hand.Contains(tapped) {
		t.Error("Silk did not enter with the tapped creature returned")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want {1}{G}{W} spent", me.ManaPool.Total())
	}
}

func TestSpmB7SilkRefusesAnUntappedCreatureForWebSlinging(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	silk := handCardFull(me, "Silk, Web Weaver", spmB7SilkType, "{2}{G}{W}", spmB7SilkOracle, []string{"G", "W"})
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	floatMana(t, g, me, "{C}{G}{W}")

	if err := g.CastSpell(me.ID, silk, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{untapped},
	}); err == nil {
		t.Error("web-slinging accepted an untapped creature")
	}
	if err := g.CastSpell(me.ID, silk, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the full {2}{G}{W} price was accepted with three mana")
	}
}

func TestSpmB7SilkMakesACitizenForEachCreatureSpellYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Silk, Web Weaver", spmB7SilkType, spmB7SilkOracle, 3, 5)

	castFree(t, g, me, "Instant Spell", "Instant")
	if got := mshB1Tokens(g, me.ID, "Citizen"); got != 0 {
		t.Fatalf("a noncreature spell made %d Citizens", got)
	}
	castFree(t, g, me, "My Creature", "Creature — Bear")
	if got := mshB1Tokens(g, me.ID, "Citizen"); got != 1 {
		t.Fatalf("my creature spell made %d Citizens, want 1", got)
	}
	castFree(t, g, me, "Another Creature", "Creature — Bear")
	if got := mshB1Tokens(g, me.ID, "Citizen"); got != 2 {
		t.Errorf("two creature spells made %d Citizens, want 2", got)
	}
	advanceToMainOf(t, g, 1)
	castFree(t, g, opp, "Their Creature", "Creature — Bear")
	if got := mshB1Tokens(g, me.ID, "Citizen"); got != 2 {
		t.Errorf("an opponent's creature spell changed my Citizens to %d, want 2", got)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name != "Human Citizen" {
				continue
			}
			if c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Human") || !c.HasSubtype("Citizen") || len(c.Colors) != 2 {
				t.Errorf("token is %d/%d %q colours %v", c.Power, c.Toughness, c.TypeLine, c.Colors)
			}
		}
	})
}

func TestSpmB7SilkPumpsEveryCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	silk := b12Push(g, me.ID, "Silk, Web Weaver", spmB7SilkType, spmB7SilkOracle, 3, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, silk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump was activated with no mana under the strict gate")
	}
	floatMana(t, g, me, "{C}{C}{C}{G}{W}")
	if err := g.ActivateCatalogAbility(me.ID, silk, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{silk, bear} {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p < 4 || tough < 4 {
			t.Errorf("%s is %d/%d after +2/+2", id, p, tough)
		}
		if !hasEffectiveKeyword(t, g, id, "vigilance") {
			t.Errorf("%s lacks vigilance", id)
		}
	}
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("the bear is power %d, want 4", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 || hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Errorf("an opponent's creature was pumped: power %d", p)
	}
	advanceToUpkeepOf(t, g, 1)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the bear is power %d after the turn, want 2", p)
	}
}

// --- Silver Sable, Mercenary Leader ---------------------------------------------------------------------

func TestSpmB7SableEntersAndPutsACounterOnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	sable := castHoldCreature(t, g, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, spmB7Target(sable)); err == nil {
		t.Fatal("Silver Sable was allowed to target herself with \"another target creature\"")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, theirs); got != 1 {
		t.Errorf("their creature has %d counters, want 1", got)
	}
	if got := plusOneCounters(g, sable); got != 0 {
		t.Errorf("Silver Sable has %d counters, want 0", got)
	}
}

func TestSpmB7SableAttackGivesALegalModifiedCreatureLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sable := b12Push(g, me.ID, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	modified := b12Creature(g, me.ID, "Modified Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Modified Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{sable, modified, theirs} {
			_ = AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(
				NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: sable}))
		}
	})

	declareAttack(t, g, opp.ID, sable)
	ch := latestPickTarget(g, me.ID)
	if ch == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	for _, bad := range []uuid.UUID{plain, theirs} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, spmB7Target(bad)); err == nil {
			t.Errorf("target %s was accepted though it is unmodified or not mine", bad)
		}
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, spmB7Target(modified)); err != nil {
		t.Fatalf("picking the modified creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, modified, "lifelink") {
		t.Error("the modified creature did not gain lifelink")
	}
	if hasEffectiveKeyword(t, g, plain, "lifelink") || hasEffectiveKeyword(t, g, sable, "lifelink") {
		t.Error("another creature gained lifelink")
	}
}

func TestSpmB7SableAttackWithNoModifiedCreatureHasNoTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sable := b12Push(g, me.ID, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, sable)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the attack trigger asked for a target with no modified creature")
	}
	if hasEffectiveKeyword(t, g, sable, "lifelink") {
		t.Error("Silver Sable gained lifelink while unmodified")
	}
}

func TestSpmB7SableEquipmentMakesACreatureModified(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sable := b12Push(g, me.ID, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := seedEquipment(g, me.ID, "Bonesplitter", spmB7BonesplitterOrc)
	equipTo(t, g, me.ID, sword, bear)

	declareAttack(t, g, opp.ID, sable)
	// The equipped bear is the only modified creature: it is picked unasked
	// or by prompt, either way it ends with lifelink.
	if ch := latestPickTarget(g, me.ID); ch != nil {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, spmB7Target(bear)); err != nil {
			t.Fatalf("picking the equipped bear: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the equipped creature did not gain lifelink")
	}
}

// --- Sinister Hideout -----------------------------------------------------------------------------------

func TestSpmB7HideoutEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := enterFromHand(t, g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle)
	if !battlefieldCardPtr(t, g, land).Tapped {
		t.Error("Sinister Hideout entered untapped")
	}
}

func TestSpmB7HideoutTapsForBlueOrBlackOnly(t *testing.T) {
	for _, colour := range []string{"U", "B"} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := pushCatalogPermanent(g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle, false)
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("tap for %s: %v", colour, err)
		}
		if got := me.ManaPool.CountColor(colour); got != 1 || me.ManaPool.Total() != 1 {
			t.Errorf("tapped for %s: %d of it, %d mana in all", colour, got, me.ManaPool.Total())
		}
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil && me.ManaPool.CountColor("R") > 0 {
		t.Error("Sinister Hideout produced red")
	}
}

func TestSpmB7HideoutSurveilsForFourAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "Top", "Second", "Third")
	land := pushCatalogPermanent(g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the surveil was activated with no mana under the strict gate")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	graveyard := me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !battlefieldCardPtr(t, g, land).Tapped {
		t.Error("the land is untapped after paying {T}")
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looks at %d cards, want 1", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := me.Graveyard.Size(); got != graveyard+1 {
		t.Errorf("graveyard %d, want %d after surveilling a card away", got, graveyard+1)
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped land activated its {T} ability again")
	}
}

// --- Skyward Spider -----------------------------------------------------------------------------------

func TestSpmB7SkywardSpiderFliesOnlyWhileModified(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spider := b12Push(g, me.ID, "Skyward Spider", spmB7SkywardType, spmB7SkywardOracle, 2, 2)
	if hasEffectiveKeyword(t, g, spider, "flying") {
		t.Fatal("an unmodified Skyward Spider flies")
	}
	g.WithWriteLock(func() {
		_ = AddCounter{Target: spider, Kind: game.CounterPlusOne, N: 1}.Apply(
			NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: spider}))
	})
	if !hasEffectiveKeyword(t, g, spider, "flying") {
		t.Error("Skyward Spider with a counter does not fly")
	}
}

func TestSpmB7SkywardSpiderFliesWhileEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spider := b12Push(g, me.ID, "Skyward Spider", spmB7SkywardType, spmB7SkywardOracle, 2, 2)
	sword := seedEquipment(g, me.ID, "Bonesplitter", spmB7BonesplitterOrc)
	equipTo(t, g, me.ID, sword, spider)
	if !hasEffectiveKeyword(t, g, spider, "flying") {
		t.Error("an equipped Skyward Spider does not fly")
	}
}

func TestSpmB7SkywardSpiderHasWardTwo(t *testing.T) {
	mscM7WardDeclined(t, "Skyward Spider", spmB7SkywardType, spmB7SkywardOracle)
}
