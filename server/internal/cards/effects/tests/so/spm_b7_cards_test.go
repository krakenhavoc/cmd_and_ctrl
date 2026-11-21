package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	spmB7BonesplitterOrc = effectstest.BonesplitterOracle
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

func spmB7Target(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// spmB7Sting casts Scorpion's Sting from hand under the strict mana gate and
// returns the cast error.
func spmB7Sting(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, mana string) error {
	t.Helper()
	id := effectstest.HandCardFull(me, "Scorpion's Sting", "Instant", "{1}{B}", spmB7StingOracle, []string{"B"})
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{spmB7Target(target)}})
}

// --- Scorpion's Sting -------------------------------------------------------------------------------

func TestSpmB7StingShrinksTheCreatureUntilEndOfTurnAndKillsASmallOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 3, 3)

	if err := spmB7Sting(t, g, me, big, "{C}{B}"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the 4/4 is %d/%d after -3/-3, want 1/1", p, tough)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want {1}{B} spent", me.ManaPool.Total())
	}

	if err := spmB7Sting(t, g, me, small, "{C}{B}"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 3/3 survived -3/-3")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.EffectivePower(t, g, big); p != 4 {
		t.Errorf("the 4/4 is power %d after the turn ended, want 4", p)
	}
}

func TestSpmB7StingRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Plains", "Basic Land — Plains", "", false)

	if err := spmB7Sting(t, g, me, land, "{C}{B}"); err == nil {
		t.Error("Scorpion's Sting accepted a land as its target")
	}
	_ = bear
}

// --- Scorpion, Seething Striker ----------------------------------------------------------------------

func TestSpmB7ScorpionConnivesAtYourEndStepAfterAnyCreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.B12Push(g, me.ID, "Scorpion, Seething Striker", spmB7ScorpionType, spmB7ScorpionOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, theirs)

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("the connive did not ask for a discard")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[me.Hand.Size()-1].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, bear); got != 1 {
		t.Errorf("the bear has %d counters after a nonland discard, want 1", got)
	}
}

func TestSpmB7ScorpionDoesNothingWhenNoCreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	scorpion := effectstest.B12Push(g, me.ID, "Scorpion, Seething Striker", spmB7ScorpionType, spmB7ScorpionOracle, 3, 3)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToEndStep(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !effectstest.StackFullyEmpty(g) {
		t.Fatal("the end step trigger went on the stack with no creature having died")
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil || effectstest.PlusOneCounters(g, scorpion) != 0 {
		t.Error("Scorpion connived with no creature having died")
	}
}

func TestSpmB7ScorpionHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	scorpion := effectstest.B12Push(g, g.Seats[0].ID, "Scorpion, Seething Striker", spmB7ScorpionType, spmB7ScorpionOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, scorpion, "deathtouch") {
		t.Error("Scorpion lacks deathtouch")
	}
}

// --- Secret Identity ----------------------------------------------------------------------------------

func spmB7Secret(t *testing.T, g *game.Game, mode int, target uuid.UUID) {
	t.Helper()
	effectstest.CastModal(t, g, "Secret Identity", "Instant", spmB7SecretOracle, []int{mode}, []game.TargetRef{spmB7Target(target)})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSpmB7SecretIdentityConcealMakesItAOneByOneCitizenWithHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: bear, Kind: game.CounterPlusOne, N: 1}.Apply(
			effects.NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: bear}))
	})

	spmB7Secret(t, g, 0, bear)
	// Base 1/1 in layer 7b, the counter still lands on top.
	if p, tough := effectstest.HobA4Power(t, g, bear), effectstest.MscM6Toughness(g, bear); p != 2 || tough != 2 {
		t.Errorf("concealed creature is %d/%d, want 2/2 (base 1/1 plus its counter)", p, tough)
	}
	if got := effectstest.EffectiveSubtypes(t, g, bear); len(got) != 1 || got[0] != "Citizen" {
		t.Errorf("subtypes %v, want [Citizen]", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the concealed creature lacks hexproof")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the concealed creature has flying")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.HobA4Power(t, g, bear); p != 5 {
		t.Errorf("after the turn the creature has power %d, want 5", p)
	}
	if got := effectstest.EffectiveSubtypes(t, g, bear); len(got) != 1 || got[0] != "Bear" {
		t.Errorf("after the turn subtypes %v, want [Bear]", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof lasted past the turn")
	}
}

func TestSpmB7SecretIdentityRevealMakesItAThreeByFourHeroWithFlyingAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)

	spmB7Secret(t, g, 1, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("revealed creature is %d/%d, want 3/4", p, tough)
	}
	if got := effectstest.EffectiveSubtypes(t, g, bear); len(got) != 1 || got[0] != "Hero" {
		t.Errorf("subtypes %v, want [Hero]", got)
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the revealed creature lacks %s", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the revealed creature has hexproof")
	}
}

func TestSpmB7SecretIdentityRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for mode := 0; mode < 2; mode++ {
		if _, err := effectstest.CastModalErr(t, g, "Secret Identity", "Instant", spmB7SecretOracle, []int{mode},
			[]game.TargetRef{spmB7Target(theirs)}); err == nil {
			t.Errorf("mode %d accepted a creature its caster does not control", mode)
		}
	}
}

// --- Shadow of the Goblin ------------------------------------------------------------------------------

func TestSpmB7GoblinDiscardsThenDrawsAtYourFirstMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.PushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)
	before := me.Hand.Size()
	victim := me.Hand.Cards[0].InstanceID

	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("the first main phase trigger asked for no discard")
	}
	effectstest.AnswerDiscard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(victim) {
		t.Error("the discarded card is still in hand")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size %d, want %d (one discarded, one drawn)", got, before)
	}
}

func TestSpmB7GoblinWithAnEmptyHandDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.PushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatal("asked for a discard with no card in hand")
	}
	if got := me.Hand.Size(); got != 0 {
		t.Errorf("drew %d cards though nothing was discarded", got)
	}
}

func spmB7OpponentLife(g *game.Game) []int {
	return effectstest.LifeOfOpponents(g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)

	before := spmB7OpponentLife(g)
	effectstest.CastFree(t, g, me, "Hand Spell", "Instant")
	spmB7AssertPinged(t, g, before, 0, "a spell cast from hand")

	_, id := effectstest.ImpulseExile(t, g, "Exile Spell", "Instant", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	spmB7AssertPinged(t, g, before, 1, "a spell cast from exile")
}

func TestSpmB7GoblinPingsForALandPlayedFromExileButNotFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Shadow of the Goblin", "Enchantment", spmB7GoblinOracle, false)

	before := spmB7OpponentLife(g)
	land := effectstest.MscM7Hand(me, "Hand Forest", "Basic Land — Forest")
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{}); err != nil {
		t.Fatalf("play the land from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	spmB7AssertPinged(t, g, before, 0, "a land played from hand")

	// A second land drop is needed for the exiled land.
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(me.ID, 1) })
	_, exiled := effectstest.ImpulseExile(t, g, "Exile Forest", "Basic Land — Forest", "")
	if err := g.CastSpell(me.ID, exiled, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	spmB7AssertPinged(t, g, before, 1, "a land played from exile")
}

// --- Shocker, Unshakable -----------------------------------------------------------------------------

func TestSpmB7ShockerHitsTheCreatureAndItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	life := opp.Life

	effectstest.CastHoldCreature(t, g, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	effectstest.PickCard(t, g, g.Seats[0].ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCardPtr(t, g, theirs).DamageMarked; got != 2 {
		t.Errorf("the creature has %d damage, want 2", got)
	}
	if opp.Life != life-2 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-2)
	}
}

func TestSpmB7ShockerStillHitsTheControllerOfACreatureTheDamageKills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	life := opp.Life

	effectstest.CastHoldCreature(t, g, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 1/1 survived 2 damage")
	}
	if opp.Life != life-2 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-2)
	}
}

func TestSpmB7ShockerCanTargetYourOwnCreatureAndHurtsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	life := me.Life

	effectstest.CastHoldCreature(t, g, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("I am at %d, want %d", me.Life, life-2)
	}
}

func TestSpmB7ShockerHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shocker := effectstest.B12Push(g, me.ID, "Shocker, Unshakable", spmB7ShockerType, spmB7ShockerOracle, 5, 5)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, shocker, "first strike") {
		t.Error("Shocker lacks first strike during his controller's turn")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, shocker, "first strike") {
		t.Error("Shocker has first strike during an opponent's turn")
	}
}

// --- Shriek, Treblemaker -----------------------------------------------------------------------------

// spmB7ShriekPrompt walks priority until the discard question is open.
func spmB7ShriekPrompt(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	card := me.Hand.Cards[0].InstanceID

	effectstest.AdvanceToMain(t, g)
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
	effectstest.AssistCardSettle(t, g, me.ID, &pick, false)
	effectstest.AssertRestrictions(t, g, theirs, game.CantBlock)
}

func TestSpmB7ShriekDecliningDiscardsNothingAndRestrictsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	before := me.Hand.Size()

	effectstest.AdvanceToMain(t, g)
	if spmB7ShriekPrompt(t, g, me) == nil {
		t.Fatal("no discard question at the first main phase")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("hand is %d cards, want %d", me.Hand.Size(), before)
	}
	if effectstest.RestrictionsOf(t, g, theirs)&game.CantBlock != 0 {
		t.Error("a creature can't block though no card was discarded")
	}
}

func TestSpmB7ShriekAsksNothingWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("asked to discard with no card in hand")
	}
}

func TestSpmB7ShriekSonicBlastHitsThePlayerWhoLostTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Shriek, Treblemaker", spmB7ShriekType, spmB7ShriekOracle, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life

	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("my creature dying cost lives: me %d->%d, opponent %d->%d", myLife, me.Life, oppLife, opp.Life)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-1)
	}
	if me.Life != myLife {
		t.Errorf("I lost life: %d", me.Life)
	}
}

// --- Silk, Web Weaver ---------------------------------------------------------------------------------

func TestSpmB7SilkWebSlingsForOneGreenWhiteAndATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	silk := effectstest.HandCardFull(me, "Silk, Web Weaver", spmB7SilkType, "{2}{G}{W}", spmB7SilkOracle, []string{"G", "W"})
	tapped := effectstest.PushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{G}{W}")

	if err := g.CastSpell(me.ID, silk, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(silk) || !me.Hand.Contains(tapped) {
		t.Error("Silk did not enter with the tapped creature returned")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want {1}{G}{W} spent", me.ManaPool.Total())
	}
}

func TestSpmB7SilkRefusesAnUntappedCreatureForWebSlinging(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	silk := effectstest.HandCardFull(me, "Silk, Web Weaver", spmB7SilkType, "{2}{G}{W}", spmB7SilkOracle, []string{"G", "W"})
	untapped := effectstest.PushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{G}{W}")

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Silk, Web Weaver", spmB7SilkType, spmB7SilkOracle, 3, 5)

	effectstest.CastFree(t, g, me, "Instant Spell", "Instant")
	if got := effectstest.MshB1Tokens(g, me.ID, "Citizen"); got != 0 {
		t.Fatalf("a noncreature spell made %d Citizens", got)
	}
	effectstest.CastFree(t, g, me, "My Creature", "Creature — Bear")
	if got := effectstest.MshB1Tokens(g, me.ID, "Citizen"); got != 1 {
		t.Fatalf("my creature spell made %d Citizens, want 1", got)
	}
	effectstest.CastFree(t, g, me, "Another Creature", "Creature — Bear")
	if got := effectstest.MshB1Tokens(g, me.ID, "Citizen"); got != 2 {
		t.Errorf("two creature spells made %d Citizens, want 2", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.CastFree(t, g, opp, "Their Creature", "Creature — Bear")
	if got := effectstest.MshB1Tokens(g, me.ID, "Citizen"); got != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	silk := effectstest.B12Push(g, me.ID, "Silk, Web Weaver", spmB7SilkType, spmB7SilkOracle, 3, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, silk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump was activated with no mana under the strict gate")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{G}{W}")
	if err := g.ActivateCatalogAbility(me.ID, silk, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{silk, bear} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p < 4 || tough < 4 {
			t.Errorf("%s is %d/%d after +2/+2", id, p, tough)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Errorf("%s lacks vigilance", id)
		}
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("the bear is power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Errorf("an opponent's creature was pumped: power %d", p)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the bear is power %d after the turn, want 2", p)
	}
}

// --- Silver Sable, Mercenary Leader ---------------------------------------------------------------------

func TestSpmB7SableEntersAndPutsACounterOnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	sable := effectstest.CastHoldCreature(t, g, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, spmB7Target(sable)); err == nil {
		t.Fatal("Silver Sable was allowed to target herself with \"another target creature\"")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, theirs); got != 1 {
		t.Errorf("their creature has %d counters, want 1", got)
	}
	if got := effectstest.PlusOneCounters(g, sable); got != 0 {
		t.Errorf("Silver Sable has %d counters, want 0", got)
	}
}

func TestSpmB7SableAttackGivesALegalModifiedCreatureLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sable := effectstest.B12Push(g, me.ID, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	modified := effectstest.B12Creature(g, me.ID, "Modified Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Modified Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{sable, modified, theirs} {
			_ = effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(
				effects.NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: sable}))
		}
	})

	effectstest.DeclareAttack(t, g, opp.ID, sable)
	ch := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, modified, "lifelink") {
		t.Error("the modified creature did not gain lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, plain, "lifelink") || effectstest.HasEffectiveKeyword(t, g, sable, "lifelink") {
		t.Error("another creature gained lifelink")
	}
}

func TestSpmB7SableAttackWithNoModifiedCreatureHasNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sable := effectstest.B12Push(g, me.ID, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, sable)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the attack trigger asked for a target with no modified creature")
	}
	if effectstest.HasEffectiveKeyword(t, g, sable, "lifelink") {
		t.Error("Silver Sable gained lifelink while unmodified")
	}
}

func TestSpmB7SableEquipmentMakesACreatureModified(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sable := effectstest.B12Push(g, me.ID, "Silver Sable, Mercenary Leader", spmB7SableType, spmB7SableOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", spmB7BonesplitterOrc)
	effectstest.EquipTo(t, g, me.ID, sword, bear)

	effectstest.DeclareAttack(t, g, opp.ID, sable)
	// The equipped bear is the only modified creature: it is picked unasked
	// or by prompt, either way it ends with lifelink.
	if ch := effectstest.LatestPickTarget(g, me.ID); ch != nil {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, spmB7Target(bear)); err != nil {
			t.Fatalf("picking the equipped bear: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the equipped creature did not gain lifelink")
	}
}

// --- Sinister Hideout -----------------------------------------------------------------------------------

func TestSpmB7HideoutEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.EnterFromHand(t, g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle)
	if !effectstest.BattlefieldCardPtr(t, g, land).Tapped {
		t.Error("Sinister Hideout entered untapped")
	}
}

func TestSpmB7HideoutTapsForBlueOrBlackOnly(t *testing.T) {
	for _, colour := range []string{"U", "B"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.PushCatalogPermanent(g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle, false)
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("tap for %s: %v", colour, err)
		}
		if got := me.ManaPool.CountColor(colour); got != 1 || me.ManaPool.Total() != 1 {
			t.Errorf("tapped for %s: %d of it, %d mana in all", colour, got, me.ManaPool.Total())
		}
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil && me.ManaPool.CountColor("R") > 0 {
		t.Error("Sinister Hideout produced red")
	}
}

func TestSpmB7HideoutSurveilsForFourAndATap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Top", "Second", "Third")
	land := effectstest.PushCatalogPermanent(g, me.ID, "Sinister Hideout", "Land", spmB7HideoutOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the surveil was activated with no mana under the strict gate")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	graveyard := me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.BattlefieldCardPtr(t, g, land).Tapped {
		t.Error("the land is untapped after paying {T}")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spider := effectstest.B12Push(g, me.ID, "Skyward Spider", spmB7SkywardType, spmB7SkywardOracle, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, spider, "flying") {
		t.Fatal("an unmodified Skyward Spider flies")
	}
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: spider, Kind: game.CounterPlusOne, N: 1}.Apply(
			effects.NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: spider}))
	})
	if !effectstest.HasEffectiveKeyword(t, g, spider, "flying") {
		t.Error("Skyward Spider with a counter does not fly")
	}
}

func TestSpmB7SkywardSpiderFliesWhileEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spider := effectstest.B12Push(g, me.ID, "Skyward Spider", spmB7SkywardType, spmB7SkywardOracle, 2, 2)
	sword := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", spmB7BonesplitterOrc)
	effectstest.EquipTo(t, g, me.ID, sword, spider)
	if !effectstest.HasEffectiveKeyword(t, g, spider, "flying") {
		t.Error("an equipped Skyward Spider does not fly")
	}
}

func TestSpmB7SkywardSpiderHasWardTwo(t *testing.T) {
	effectstest.MscM7WardDeclined(t, "Skyward Spider", spmB7SkywardType, spmB7SkywardOracle)
}
