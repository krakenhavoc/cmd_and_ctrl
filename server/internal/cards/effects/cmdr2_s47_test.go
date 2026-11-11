package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s47ChatterfangOracle = "f15b3b76-d38a-48db-bb44-3296183c8641"
)

func s47Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}

func TestChatterfangAddsThatManySquirrels(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Chatterfang, Squirrel General", "Legendary Creature — Squirrel Warrior", s47ChatterfangOracle, false)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 2) })
	if got := s47Count(g, me.ID, "Soldier"); got != 2 {
		t.Errorf("Soldiers = %d, want 2", got)
	}
	if got := s47Count(g, me.ID, "Squirrel"); got != 2 {
		t.Errorf("Squirrels = %d, want 2 (that many)", got)
	}
	// An opponent's tokens are untouched.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(them.ID, TokenCard("1/1 white Soldier"), 1) })
	if got := s47Count(g, them.ID, "Squirrel"); got != 0 {
		t.Errorf("opponent got %d Squirrels", got)
	}
}

func TestChatterfangSacrificesXSquirrelsForPlusXMinusX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cf := pushCatalogPermanent(g, me.ID, "Chatterfang, Squirrel General", "Legendary Creature — Squirrel Warrior", s47ChatterfangOracle, false)
	s1 := ccPermanent(g, me.ID, "Squirrel", "Token Creature — Squirrel")
	s2 := ccPermanent(g, me.ID, "Squirrel", "Token Creature — Squirrel")
	victim := ccPermanent(g, me.ID, "Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 4
			}
		}
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, cf, 0, game.ActivateAbilityParams{
		XValue: 2, SacrificeIDs: []uuid.UUID{s1, s2},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	p, tt := effectivePTOf(t, g, victim)
	if p != 4 || tt != 2 {
		t.Errorf("victim = %d/%d, want 4/2", p, tt)
	}
}

const (
	s47NightmareShepherdOracle = "b8ee0dfc-f588-43b6-ab96-f3a3c58a8a3d"
	s47BrenardOracle           = "1a2ab12a-5778-40df-a2ec-2733a7121cb4"
	s47AdephageOracle          = "8e1d261d-b8d3-4fb1-922d-47789854b5b3"
	s47MimicVatOracle          = "6ebdc9f3-6b23-4716-8011-48f00d7e30c8"
	s47DuplicationOracle       = "efe7e212-5c91-42f0-9dff-591131b4f6d8"
	s47FishersTalentOracle     = "44ba2aa5-2bf5-4267-ae82-f0daf8f5e3e8"
	s47OctomancerOracle        = "1cf7e5a0-cc8f-4bf7-9767-2f3bcae26339"
	s47PerchProtectionOracle   = "295fe873-ecb5-4d5e-a73c-3af6c478afc5"
)

func s47Tokens(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

// --- Nightmare Shepherd ---

func TestNightmareShepherdExilesTheDeadAndMakesAOneOneNightmareCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Nightmare Shepherd", "Enchantment Creature — Demon", s47NightmareShepherdOracle, 4, 4)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", socB7GrizzlyBearsOracle, 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if socB7Zone(g, bear) != game.ZoneExile {
		t.Fatalf("the dead Bear is in %q, want exile", socB7Zone(g, bear))
	}
	toks := s47Tokens(g, "Grizzly Bears")
	if len(toks) != 1 {
		t.Fatalf("Bear tokens = %d, want 1", len(toks))
	}
	if !toks[0].HasSubtype("Nightmare") || !toks[0].HasSubtype("Bear") {
		t.Errorf("type line %q, want a Bear Nightmare", toks[0].TypeLine)
	}
	if p, tg := effectivePTOf(t, g, toks[0].InstanceID); p != 1 || tg != 1 {
		t.Errorf("the copy is %d/%d, want 1/1", p, tg)
	}
}

func TestNightmareShepherdDecliningLeavesTheCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Nightmare Shepherd", "Enchantment Creature — Demon", s47NightmareShepherdOracle, 4, 4)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", socB7GrizzlyBearsOracle, 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if socB7Zone(g, bear) != game.ZoneGraveyard {
		t.Errorf("the Bear is in %q, want graveyard", socB7Zone(g, bear))
	}
	if n := len(s47Tokens(g, "Grizzly Bears")); n != 0 {
		t.Errorf("%d tokens, want none", n)
	}
}

func TestNightmareShepherdIgnoresTokensAndOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Nightmare Shepherd", "Enchantment Creature — Demon", s47NightmareShepherdOracle, 4, 4)
	theirs := b12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", socB7GrizzlyBearsOracle, 2, 2)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || len(s47Tokens(g, "Grizzly Bears")) != 0 {
		t.Errorf("an opponent's creature dying triggered the Shepherd: %+v", g.PendingChoices)
	}
}

// --- Brenard, Ginger Sculptor ---

func TestBrenardMakesAFoodGolemCopyAndPumpsFoodAndGolems(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Brenard, Ginger Sculptor", "Legendary Creature — Human Artificer", s47BrenardOracle, 3, 3)
	golem := b12Creature(g, me.ID, "Clay Golem", "Artifact Creature — Golem", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	if p, tg := effectivePTOf(t, g, golem); p != 4 || tg != 4 {
		t.Errorf("Golem is %d/%d, want 4/4", p, tg)
	}
	if !hasEffectiveKeyword(t, g, golem, "trample") || hasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("only the Golem should have trample")
	}
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", socB7GrizzlyBearsOracle, 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	toks := s47Tokens(g, "Grizzly Bears")
	if len(toks) != 1 {
		t.Fatalf("Bear tokens = %d, want 1", len(toks))
	}
	tok := toks[0]
	if !tok.HasSubtype("Food") || !tok.HasSubtype("Golem") || !tok.HasSubtype("Bear") || !tok.IsArtifact() {
		t.Errorf("type line %q, want an artifact Bear Food Golem", tok.TypeLine)
	}
	// 1/1, +2/+2 from Brenard (it is a Food and a Golem) = 3/3 with trample.
	if p, tg := effectivePTOf(t, g, tok.InstanceID); p != 3 || tg != 3 {
		t.Errorf("the copy is %d/%d, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, tok.InstanceID, "trample") {
		t.Error("the copy lacks trample")
	}
	// Its Food ability: {2}, {T}, sacrifice it: gain 3 life.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tok.InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
		_ = g.AddManaForEffect(me.ID, uuid.Nil, "{2}")
	})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, tok.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the granted Food ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if g.Battlefield.Contains(tok.InstanceID) {
		t.Error("the token was not sacrificed")
	}
}

// --- Giant Adephage ---

func TestGiantAdephageCopiesItselfOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ad := b12Push(g, me.ID, "Giant Adephage", "Creature — Insect", s47AdephageOracle, 7, 7)
	dealCombatDamageToPlayer(g, ad, opp.ID, 7)
	passPriorityAroundTable(t, g)
	toks := s47Tokens(g, "Giant Adephage")
	if len(toks) != 1 {
		t.Fatalf("Adephage tokens = %d, want 1", len(toks))
	}
	if p, tg := effectivePTOf(t, g, toks[0].InstanceID); p != 7 || tg != 7 {
		t.Errorf("the copy is %d/%d, want 7/7", p, tg)
	}
	// The token has the trigger too.
	dealCombatDamageToPlayer(g, toks[0].InstanceID, opp.ID, 7)
	passPriorityAroundTable(t, g)
	if n := len(s47Tokens(g, "Giant Adephage")); n != 2 {
		t.Errorf("Adephage tokens = %d, want 2 after the token connected", n)
	}
}

// --- Mimic Vat ---

func TestMimicVatImprintsOneCardAndCopiesItWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vat := pushCatalogPermanent(g, me.ID, "Mimic Vat", "Artifact", s47MimicVatOracle, false)
	first := b12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", socB7GrizzlyBearsOracle, 2, 2)
	killCreature(t, g, opp.ID, first)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if socB7Zone(g, first) != game.ZoneExile {
		t.Fatalf("the first creature is in %q, want exile", socB7Zone(g, first))
	}
	second := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	killCreature(t, g, me.ID, second)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if socB7Zone(g, second) != game.ZoneExile {
		t.Fatalf("the second creature is in %q, want exile", socB7Zone(g, second))
	}
	if socB7Zone(g, first) != game.ZoneGraveyard {
		t.Errorf("the first card is in %q, want its owner's graveyard", socB7Zone(g, first))
	}
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{3}") })
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := s47Tokens(g, "Hill Giant")
	if len(toks) != 1 {
		t.Fatalf("Hill Giant tokens = %d, want 1", len(toks))
	}
	if !hasEffectiveKeyword(t, g, toks[0].InstanceID, "haste") {
		t.Error("the copy lacks haste")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := len(s47Tokens(g, "Hill Giant")); n != 0 {
		t.Errorf("%d tokens after the end step, want 0", n)
	}
}

func TestMimicVatIgnoresTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Mimic Vat", "Artifact", s47MimicVatOracle, false)
	tok := ccPermanent(g, me.ID, "Soldier", "Token Creature — Soldier")
	killCreature(t, g, me.ID, tok)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			t.Errorf("a token's death offered the imprint: %+v", c)
		}
	}
}

// --- Theoretical Duplication ---

func TestTheoreticalDuplicationCopiesAnOpponentsNontokenCreaturesThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Theoretical Duplication", "Instant", s47DuplicationOracle, nil)
	passPriorityAroundTable(t, g)
	// An opponent's creature enters.
	c := game.NewCard("Hill Giant", opp.ID)
	c.TypeLine = "Creature — Giant"
	c.Power, c.Toughness = 3, 3
	opp.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := s47Tokens(g, "Hill Giant")
	if len(toks) != 1 || toks[0].Controller != me.ID {
		t.Fatalf("Hill Giant tokens = %+v, want one under my control", toks)
	}
	// My own creature entering is not copied.
	mine := game.NewCard("Mine", me.ID)
	mine.TypeLine = "Creature — Bear"
	me.Hand.PushTop(mine)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, mine.InstanceID); err != nil {
		t.Fatalf("enter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(s47Tokens(g, "Mine")); n != 0 {
		t.Errorf("my own creature was copied %d times", n)
	}
}

// --- Fisher's Talent ---

func TestFishersTalentReplacesFishWithSharksThenOctopuses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	class := pushClass(g, me.ID, "Fisher's Talent", "Enchantment — Class", s47FishersTalentOracle)
	mk := func() {
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 blue Fish"), 2) })
	}
	mk()
	if n := len(s47Tokens(g, "Fish")); n != 2 {
		t.Fatalf("level 1: Fish = %d, want 2", n)
	}
	if err := levelUpTo(t, g, me.ID, class, 0, "{G}{U}"); err != nil {
		t.Fatal(err)
	}
	mk()
	if n := len(s47Tokens(g, "Shark")); n != 2 {
		t.Errorf("level 2: Sharks = %d, want 2", n)
	}
	if err := levelUpTo(t, g, me.ID, class, 1, "{2}{G}{U}"); err != nil {
		t.Fatal(err)
	}
	mk()
	if n := len(s47Tokens(g, "Octopus")); n != 2 {
		t.Errorf("level 3: Octopuses = %d, want 2 (Fish → Shark → Octopus)", n)
	}
	if n := len(s47Tokens(g, "Shark")); n != 2 {
		t.Errorf("level 3: Sharks = %d, want the 2 already made", n)
	}
}

func TestFishersTalentUpkeepRevealsALandForAFishThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushClass(g, me.ID, "Fisher's Talent", "Enchantment — Class", s47FishersTalentOracle)
	land := game.NewCard("Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	me.Library.PushTop(land)
	hand := me.Hand.Size()
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, true)
	passPriorityAroundTable(t, g)
	if n := len(s47Tokens(g, "Fish")); n != 1 {
		t.Errorf("Fish = %d, want 1", n)
	}
	if me.Hand.Size() <= hand {
		t.Errorf("hand %d → %d, want the draw", hand, me.Hand.Size())
	}
}

// --- Octomancer ---

func TestOctomancerGiftGivesAnOctopusAndCopiesATokenAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mustCastWithGift(t, g, "Octomancer", "Creature — Frog Druid", s47OctomancerOracle, nil, opp.ID)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, opp.ID, "Octopus"); n != 1 {
		t.Fatalf("opponent's Octopuses = %d, want 1", n)
	}
	advanceToEndStep(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		var refs []game.TargetRef
		for _, id := range p.PickTargetCards {
			if len(refs) < 1 {
				refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
			t.Fatalf("pick the token: %v", err)
		}
	} else {
		t.Fatal("no target pick for the end-step trigger")
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Octopus"); n != 1 {
		t.Errorf("my Octopuses = %d, want a copy of the gifted token", n)
	}
}

func TestOctomancerUnpromisedGivesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	mustCastWithGift(t, g, "Octomancer", "Creature — Frog Druid", s47OctomancerOracle, nil, uuid.Nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, opp.ID, "Octopus"); n != 0 {
		t.Errorf("opponent's Octopuses = %d, want 0", n)
	}
}

// --- Perch Protection ---

func TestPerchProtectionPromisedGivesATurnAndShelters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	turns := len(g.ExtraTurns)
	id := mustCastWithGift(t, g, "Perch Protection", "Instant", s47PerchProtectionOracle, nil, opp.ID)
	passPriorityAroundTable(t, g)
	if len(g.ExtraTurns) != turns+1 || g.ExtraTurns[len(g.ExtraTurns)-1].Seat != 1 {
		t.Errorf("extra turns %+v, want one for the opponent", g.ExtraTurns)
	}
	if n, _ := countTokensNamed(g, me.ID, "Bird"); n != 0 {
		t.Errorf("%d Birds on the battlefield, want 0 (they phased out)", n)
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	if me.Life != life {
		t.Errorf("life %d → %d, want no change", life, me.Life)
	}
	if socB7Zone(g, id) != game.ZoneExile {
		t.Errorf("Perch Protection is in %q, want exile", socB7Zone(g, id))
	}
}

func TestPerchProtectionUnpromisedJustMakesFourBirds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	turns := len(g.ExtraTurns)
	id := mustCastWithGift(t, g, "Perch Protection", "Instant", s47PerchProtectionOracle, nil, uuid.Nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Bird"); n != 4 {
		t.Errorf("Birds = %d, want 4", n)
	}
	if len(g.ExtraTurns) != turns {
		t.Error("an unpromised Perch Protection queued an extra turn")
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, me.Life)
	}
	if socB7Zone(g, id) != game.ZoneExile {
		t.Errorf("Perch Protection is in %q, want exile", socB7Zone(g, id))
	}
}
