package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b607_test.go — Mirrodin, batch 607 (#5606): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB607SunriseOracle  = "15cef63b-903e-4f1b-a128-e59157e07c36"
	oldB607WrenchOracle   = "484eeb13-fe74-45fe-b089-7beec87cfcdb"
	oldB607BaneOracle     = "3609d300-af9f-4ddf-ad0e-8bebeca15a04"
	oldB607BubbleOracle   = "864c7575-4589-416c-b143-310d5ef238c5"
	oldB607OupheOracle    = "5b42e91a-e3bd-4f69-9abf-ed6273715fff"
	oldB607OverrideOracle = "a2845761-7e43-4a88-8fc2-2d95d471b00c"
	oldB607SifterOracle   = "99f8805f-6618-4838-a4e7-a3660e3b087f"
	oldB607TangleOracle   = "b94efd70-705d-4cdf-9990-3a065a541f28"
	oldB607CoilsOracle    = "bbb16f23-fdb1-444b-bef8-ce3e177c2678"
	oldB607IncubatorOracl = "d0a493af-551e-4475-af5d-ec4e080728cb"
	oldB607HumOracle      = "95803a5a-9f98-4764-b6bd-6cf4cdcb740b"
	oldB607OgreOracle     = "d556c71e-ce16-4229-880b-744790f93797"
)

func TestOldB607CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB607SunriseOracle: "Second Sunrise", oldB607WrenchOracle: "Wrench Mind", oldB607BaneOracle: "Relic Bane",
		oldB607BubbleOracle: "Inertia Bubble", oldB607OupheOracle: "Brown Ouphe", oldB607OverrideOracle: "Override",
		oldB607SifterOracle: "Timesifter", oldB607TangleOracle: "Tangleroot", oldB607CoilsOracle: "Lightning Coils",
		oldB607IncubatorOracl: "Myr Incubator", oldB607HumOracle: "Hum of the Radix",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Second Sunrise ---

func TestOldB607SecondSunriseReturnsOnlyCardsPutThereFromTheBattlefieldThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	for _, id := range []uuid.UUID{bear, rock} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	g.RunStateChecksForTest()
	discarded := effectstest.GraveyardCardOf(g, me, "Discarded Bear", "Creature — Bear", "{1}{G}", "")
	sorcery := effectstest.GraveyardCardOf(g, me, "Dead Sorcery", "Sorcery", "{1}", "")
	effectstest.CastCatalogSpell(t, g, "Second Sunrise", "Instant", oldB607SunriseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("my creature and artifact that died this turn did not return")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature that died this turn did not return")
	}
	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Errorf("the opponent's creature came back under %s, want its owner", got)
	}
	if g.Battlefield.Contains(discarded) || g.Battlefield.Contains(sorcery) {
		t.Error("a card that was not put there from the battlefield returned")
	}
}

// --- Wrench Mind ---

func oldB607WrenchMind(t *testing.T) (g *game.Game, me, opp *game.Player, artifact, bear, bear2 uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for _, c := range append([]uuid.UUID(nil), effectstest.HandIDs(g, opp.ID)...) {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneLibrary, Owner: opp.ID}, c); err != nil {
			t.Fatalf("clear the hand: %v", err)
		}
	}
	artifact = effectstest.PushToHand(g, opp.ID, "Rock", "Artifact")
	bear = effectstest.PushToHand(g, opp.ID, "Bear", "Creature — Bear")
	bear2 = effectstest.PushToHand(g, opp.ID, "Other Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Wrench Mind", "Sorcery", oldB607WrenchOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestOldB607WrenchMindAcceptsOneArtifactCardOrTwoCards(t *testing.T) {
	g, _, opp, artifact, bear, bear2 := oldB607WrenchMind(t)
	prompt := effectstest.DiscardChoiceFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("the target player was not asked to discard")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("a lone nonartifact card paid the clause")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{artifact}); err != nil {
		t.Fatalf("a lone artifact card was refused: %v", err)
	}
	if opp.Hand.Contains(artifact) || !opp.Hand.Contains(bear) || !opp.Hand.Contains(bear2) {
		t.Error("only the artifact card should have been discarded")
	}
}

func TestOldB607WrenchMindAcceptsTwoNonartifactCards(t *testing.T) {
	g, _, opp, artifact, bear, bear2 := oldB607WrenchMind(t)
	effectstest.AnswerDiscard(t, g, opp.ID, bear, bear2)
	if opp.Hand.Contains(bear) || opp.Hand.Contains(bear2) || !opp.Hand.Contains(artifact) {
		t.Error("the two chosen cards should have been discarded and the artifact kept")
	}
}

// --- Relic Bane and Inertia Bubble ---

func TestOldB607RelicBaneDrainsTheArtifactsControllerEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.Enchant(t, g, "Relic Bane", oldB607BaneOracle, rock)
	myLife, theirLife := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 {
		t.Errorf("the artifact's controller has %d life, want %d", opp.Life, theirLife-2)
	}
	if me.Life != myLife {
		t.Errorf("the Aura's controller lost life: %d", me.Life)
	}
}

func TestOldB607RelicBaneCannotEnchantANonartifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Relic Bane", effectstest.AuraTypeLine, oldB607BaneOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("Relic Bane was cast on a nonartifact creature")
	}
}

func TestOldB607InertiaBubbleKeepsTheArtifactTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.Enchant(t, g, "Inertia Bubble", oldB607BubbleOracle, rock)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(rock) })
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.IsTapped(g, rock) {
		t.Error("the enchanted artifact untapped")
	}
}

// --- Brown Ouphe ---

func TestOldB607BrownOupheCountersAnArtifactsAbilityButNotAnotherKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ouphe := effectstest.B12Push(g, me.ID, "Brown Ouphe", "Creature — Ouphe", oldB607OupheOracle, 1, 1)
	ogre := effectstest.B12Push(g, me.ID, "Barrage Ogre", "Artifact Creature — Ogre Shaman", oldB607OgreOracle, 3, 2)
	fodder := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	jailer := effectstest.B12Push(g, me.ID, "Akroan Jailer", "Creature — Human Soldier", "e3ad2c77-2a41-4a4a-a6bf-e542d0162b88", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, jailer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate the Jailer: %v", err)
	}
	notArtifact := oldB607ActivatedItemOf(g, jailer)
	if notArtifact == uuid.Nil {
		t.Fatal("the Jailer's ability is not on the stack")
	}
	effectstest.AddPool(t, g, me.ID, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, ouphe, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(notArtifact)}); err == nil {
		t.Fatal("Brown Ouphe targeted an ability from a source that is not an artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, ogre, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate Barrage Ogre: %v", err)
	}
	ability := oldB607ActivatedItemOf(g, ogre)
	if ability == uuid.Nil {
		t.Fatal("Barrage Ogre's ability is not on the stack")
	}
	if err := g.ActivateCatalogAbility(me.ID, ouphe, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(ability)}); err != nil {
		t.Fatalf("counter the artifact's ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the countered ability still dealt its damage")
	}
}

// --- Override ---

func oldB607Override(t *testing.T, responderArtifacts int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	caster := g.Seats[g.Turn.ActiveSeat]
	responder := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < responderArtifacts; i++ {
		effectstest.B12Permanent(g, responder.ID, "Rock", "Artifact")
	}
	effectstest.B12Permanent(g, caster.ID, "Their Rock", "Artifact")
	victim := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: victim, Name: "Their Spell", TypeLine: "Creature — Bear", Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, responder, "Override", "Instant", oldB607OverrideOracle, effectstest.CardTargets(victim))
	effectstest.PassUntilTaxed(t, g, caster.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the tax was unanswered")
	}
	return g, caster, victim
}

func TestOldB607OverrideTaxesOnePerArtifactItsControllerControls(t *testing.T) {
	g, caster, victim := oldB607Override(t, 2)
	effectstest.FloatMana(t, g, caster, "{C}{C}")
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := caster.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, want 0 (two artifacts, so {2} was paid)", got)
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the spell was countered though its controller paid")
	}
}

func TestOldB607OverrideCountersWhenTheTaxIsDeclined(t *testing.T) {
	g, caster, victim := oldB607Override(t, 2)
	effectstest.AnswerPayUnless(t, g, caster.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !caster.Graveyard.Contains(victim) {
		t.Error("the spell was not countered")
	}
}

// --- Timesifter ---

func TestOldB607TimesifterGivesTheExtraTurnToTheGreatestManaValueAfterATieBreaks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Timesifter", "Artifact", oldB607SifterOracle, false)
	push := func(seat int, cost string) {
		effectstest.PushLibraryCard(g.Seats[seat], game.Card{Name: "Top " + cost, TypeLine: "Sorcery", ManaCost: cost})
	}
	// Second round for the two tied players first (it sits underneath), then the first round.
	push(1, "{2}")
	push(2, "{4}")
	push(0, "{1}{1}")
	push(1, "{5}")
	push(2, "{5}")
	push(3, "{1}")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	queued := g.ExtraTurnsQueued()
	if len(queued) != 1 {
		t.Fatalf("%d extra turns queued, want 1", len(queued))
	}
	if queued[0].Seat != 2 {
		t.Errorf("the extra turn goes to seat %d, want 2 (tied at 5, then 4 beat 2)", queued[0].Seat)
	}
}

// --- Tangleroot ---

func TestOldB607TanglerootAddsGForAnyPlayersCreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tangleroot", "Artifact", oldB607TangleOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Not a Creature", "Sorcery", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 0 {
		t.Fatalf("a noncreature spell made %d mana", got)
	}
	effectstest.CastFromHandForTest(t, g, me, "My Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("my creature spell made %d mana, want 1", got)
	}
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.AdvanceToMain(t, g)
	before := me.ManaPool.Total()
	effectstest.CastFromHandForTest(t, g, opp, "Their Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.ManaPool.Total(); got != 1 {
		t.Errorf("the opponent's creature spell gave them %d mana, want 1", got)
	}
	if me.ManaPool.Total() != before {
		t.Error("Tangleroot's controller got the opponent's mana")
	}
}

// --- Lightning Coils ---

func TestOldB607LightningCoilsCountsDeathsAndMakesElementalsAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	coils := effectstest.PushCatalogPermanent(g, me.ID, "Lightning Coils", "Artifact", oldB607CoilsOracle, false)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 5; i++ {
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		if i < 4 && oldB607Charge(g, coils) != i+1 {
			t.Fatalf("after %d deaths the Coils have %d charge counters", i+1, oldB607Charge(g, coils))
		}
	}
	if got := oldB607Charge(g, coils); got != 5 {
		t.Fatalf("%d charge counters, want 5", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB607Charge(g, coils); got != 0 {
		t.Errorf("%d charge counters left, want 0", got)
	}
	n := effectstest.CountBattlefieldNamed(g, me.ID, "Elemental")
	if n != 5 {
		t.Fatalf("%d Elemental tokens, want 5", n)
	}
	tok := effectstest.BattlefieldIDNamed(g, me.ID, "Elemental")
	if p, tough := effectstest.EffectivePTOf(t, g, tok); p != 3 || tough != 1 {
		t.Errorf("the Elemental is %d/%d, want 3/1", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, tok, "haste") {
		t.Error("the Elemental lacks haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Elemental"); got != 0 {
		t.Errorf("%d Elementals survived the end step", got)
	}
}

func TestOldB607LightningCoilsIgnoresTokensAndAnOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	coils := effectstest.PushCatalogPermanent(g, me.ID, "Lightning Coils", "Artifact", oldB607CoilsOracle, false)
	effectstest.AdvanceToMain(t, g)
	token := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	for _, id := range []uuid.UUID{token, theirs} {
		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := oldB607Charge(g, coils); got != 0 {
		t.Errorf("%d charge counters after a token and an opponent's creature died, want 0", got)
	}
}

// --- Myr Incubator ---

func TestOldB607MyrIncubatorMakesAMyrForEachArtifactCardFound(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inc := effectstest.PushCatalogPermanent(g, me.ID, "Myr Incubator", "Artifact", oldB607IncubatorOracl, false)
	a := effectstest.PushLibraryCard(me, game.Card{Name: "Rock A", TypeLine: "Artifact"})
	b := effectstest.PushLibraryCard(me, game.Card{Name: "Rock B", TypeLine: "Artifact Creature — Golem"})
	forest := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, inc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(inc) {
		t.Error("the Incubator was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		for _, id := range c.SearchCards {
			if id == forest {
				t.Fatal("a land was offered to a search for artifact cards")
			}
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Myr"); got != 2 {
		t.Errorf("%d Myr tokens, want 2", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a found card is in %v, want exile", z)
		}
	}
}

// --- Hum of the Radix ---

func TestOldB607HumOfTheRadixTaxesArtifactSpellsByTheirCastersArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Hum of the Radix", "Enchantment", oldB607HumOracle, false)
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	}
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if got := effectstest.SelfPricedMV(t, g, me, oldB607TangleOracle, "Artifact", "{3}"); got != 5 {
		t.Errorf("an artifact spell costs %d for the player with two artifacts, want 5", got)
	}
	if got := effectstest.SelfPricedMV(t, g, opp, oldB607TangleOracle, "Artifact", "{3}"); got != 4 {
		t.Errorf("an artifact spell costs %d for the player with one artifact, want 4", got)
	}
	if got := effectstest.SelfPricedMV(t, g, me, "", "Creature — Bear", "{3}"); got != 3 {
		t.Errorf("a nonartifact spell costs %d, want 3", got)
	}
}

// oldB607Charge is the number of charge counters on a permanent.
func oldB607Charge(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			n = c.Counters[game.CounterCharge]
		}
	})
	return n
}

// oldB607ActivatedItemOf is the id of the activated ability of `source` waiting
// on the stack.
func oldB607ActivatedItemOf(g *game.Game, source uuid.UUID) uuid.UUID {
	for id, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.SourceCardID == source {
			return id
		}
	}
	return uuid.Nil
}
