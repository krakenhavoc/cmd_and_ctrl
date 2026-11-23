package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b65_test.go — Fallout creatures (GAP batch 65, #5948), played
// through the real catalog entries.

const (
	gapB65Curie     = "48cf6b88-6bca-49d4-9951-36d98befa588"
	gapB65James     = "8c90b558-de07-4689-9427-da68d823c2e8"
	gapB65Jason     = "5c4f1f94-9916-49aa-956a-c1c6b299b24f"
	gapB65Nick      = "5663288a-7300-4e1b-8fb0-43c467709fe5"
	gapB65Piper     = "1ff9598a-0dac-4437-9450-fc8134fa3740"
	gapB65Arcade    = "e001d7e6-71c9-42fa-9d5c-9455da66e4ec"
	gapB65Owyn      = "19788405-7141-475e-94ee-d0c89488f50f"
	gapB65Aradesh   = "32620b26-10b9-47a7-bfa7-ca590ab5fa50"
	gapB65Outcast   = "79d0cdc1-2e27-48c5-8840-430490068f68"
	gapB65Sofia     = "a3fdb633-7fe9-4e72-a3a3-aa17b6f17864"
	gapB65Codsworth = "f496ff42-3644-494a-9f21-27f781ee880b"
)

func TestGapB65CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB65Curie: "Curie, Emergent Intelligence", gapB65James: "James, Wandering Dad", gapB65James + "#1": "Follow Him",
		gapB65Jason: "Jason Bright, Glowing Prophet", gapB65Nick: "Nick Valentine, Private Eye",
		gapB65Piper: "Piper Wright, Publick Reporter", gapB65Arcade: "Arcade Gannon", gapB65Owyn: "Elder Owyn Lyons",
		gapB65Aradesh: "Aradesh, the Founder", gapB65Outcast: "Brotherhood Outcast",
		gapB65Sofia: "Commander Sofia Daguerre", gapB65Codsworth: "Codsworth, Handy Helper",
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

func gapB65Clues(g *game.Game, controller uuid.UUID) int {
	return effectstest.TmtB1Count(g, controller, "Clue")
}

func gapB65PlusOnes(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[game.CounterPlusOne]
}

// --- Piper Wright ---------------------------------------------------------

func TestGapB65PiperInvestigatesOncePerDamageAndGrowsAnyCreatureWhenAClueIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	piper := effectstest.B27Push(g, me.ID, "Piper Wright, Publick Reporter", "Legendary Creature — Human Detective", gapB65Piper, "{1}{U}", 1, 2, "U")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, piper, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB65Clues(g, me.ID); got != 3 {
		t.Fatalf("%d Clues after 3 combat damage, want 3", got)
	}
	clue := effectstest.TmtB1Named(g, "Clue")[0]
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack a Clue: %v", err)
	}
	effectstest.B27Settle(t, g, me.ID, bear)
	if got := gapB65PlusOnes(t, g, bear); got != 1 {
		t.Errorf("the target creature has %d +1/+1 counters, want 1", got)
	}
}

func TestGapB65PiperDoesNothingForNoDamageOrAnOpponentsClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	piper := effectstest.B27Push(g, me.ID, "Piper Wright, Publick Reporter", "Legendary Creature — Human Detective", gapB65Piper, "{1}{U}", 1, 2, "U")
	effectstest.DealCombatDamageToPlayer(g, piper, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB65Clues(g, me.ID); got != 0 {
		t.Errorf("%d Clues for no damage, want 0", got)
	}
	g.WithWriteLock(func() { _ = g.InvestigateForEffect(opp.ID, uuid.Nil, 1) })
	oppClue := effectstest.TmtB1Named(g, "Clue")[0]
	effectstest.FloatMana(t, g, opp, "{C}{C}")
	if err := g.ActivateCatalogAbility(opp.ID, oppClue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the opponent's Clue: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB65PlusOnes(t, g, piper); got != 0 {
		t.Errorf("Piper got %d counters from an opponent's Clue, want 0", got)
	}
}

// --- Nick Valentine -------------------------------------------------------

func TestGapB65NickCantBeBlockedExceptByArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nick := effectstest.B27Push(g, me.ID, "Nick Valentine, Private Eye", "Legendary Artifact Creature — Synth Detective", gapB65Nick, "{2}{U}", 2, 2, "U")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(nick, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, nick); err == nil {
		t.Error("a nonartifact creature blocked Nick Valentine")
	}
	if err := g.DeclareBlocker(golem, nick); err != nil {
		t.Errorf("an artifact creature could not block Nick Valentine: %v", err)
	}
}

func TestGapB65NickInvestigatesWhenYourArtifactCreatureDiesButNotOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nick := effectstest.B27Push(g, me.ID, "Nick Valentine, Private Eye", "Legendary Artifact Creature — Synth Detective", gapB65Nick, "{2}{U}", 2, 2, "U")
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)

	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB65Clues(g, me.ID); got != 0 {
		t.Fatalf("%d Clues from a nonartifact creature and an opponent's artifact creature, want 0", got)
	}
	effectstest.B27Kill(g, golem)
	effectstest.B27Settle(t, g, me.ID, uuid.Nil)
	if got := gapB65Clues(g, me.ID); got != 1 {
		t.Fatalf("%d Clues after my artifact creature died, want 1", got)
	}
	effectstest.B27Kill(g, nick)
	effectstest.B27Settle(t, g, me.ID, uuid.Nil)
	if got := gapB65Clues(g, me.ID); got != 2 {
		t.Errorf("%d Clues after Nick himself died, want 2", got)
	}
}

// --- Elder Owyn Lyons -----------------------------------------------------

func TestGapB65OwynReturnsAnArtifactCardOnEntryAndOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushGraveyardPermanent(me, "Rock", "Artifact", "{1}")
	bear := effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}")
	effectstest.CastCatalogSpell(t, g, "Elder Owyn Lyons", "Legendary Creature — Human Knight", gapB65Owyn, nil)
	effectstest.B27Settle(t, g, me.ID, rock)
	if !me.Hand.Contains(rock) {
		t.Fatal("the artifact card did not return to hand on entry")
	}
	if me.Hand.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("a nonartifact card was returned")
	}
	rock2 := effectstest.PushGraveyardPermanent(me, "Second Rock", "Artifact", "{2}")
	var owyn uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elder Owyn Lyons" {
			owyn = c.InstanceID
		}
	}
	effectstest.B27Kill(g, owyn)
	effectstest.B27Settle(t, g, me.ID, rock2)
	if !me.Hand.Contains(rock2) {
		t.Error("the artifact card did not return to hand when Owyn died")
	}
}

func TestGapB65OwynGivesYourArtifactsWardOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Elder Owyn Lyons", "Legendary Creature — Human Knight", gapB65Owyn, false)
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	blade := effectstest.CastAtWardedCreature(t, g, opp, golem)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for a spell targeting my artifact")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) || !opp.Graveyard.Contains(blade) {
		t.Error("a declined ward must counter the spell")
	}
}

// --- Brotherhood Outcast --------------------------------------------------

func gapB65OutcastMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	effectstest.CastRingCreature(t, g, "Brotherhood Outcast", "Creature — Human Soldier", gapB65Outcast, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestGapB65OutcastPutsAShieldCounterOnATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	gapB65OutcastMode(t, g, me, 1)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, bear).Counters[game.CounterShield]; got != 1 {
		t.Errorf("the target has %d shield counters, want 1", got)
	}
}

func TestGapB65OutcastReturnsACheapEquipmentFromTheGraveyardButNotAnExpensiveOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.PushGraveyardPermanent(me, "Cheap Sword", "Artifact — Equipment", "{3}")
	pricey := effectstest.PushGraveyardPermanent(me, "Pricey Sword", "Artifact — Equipment", "{4}")
	gapB65OutcastMode(t, g, me, 0)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the first mode")
	}
	if effectstest.HasID(pick.PickTargetCards, pricey) {
		t.Error("an Equipment card with mana value 4 is offered")
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the Equipment card did not return to the battlefield")
	}
	if g.Battlefield.Contains(pricey) {
		t.Error("the expensive Equipment card returned")
	}
}

// --- Commander Sofia Daguerre ---------------------------------------------

func TestGapB65SofiaDestroysALegendaryPermanentAndItsControllerGetsJunk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := effectstest.B27Push(g, opp.ID, "Their Legend", "Legendary Creature — Elf", "gap-b65-legend", "{1}{G}", 2, 2, "G")
	plain := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Commander Sofia Daguerre", "Legendary Creature — Human Pilot", gapB65Sofia, nil)
	pick := effectstest.LatestPickTarget(g, me.ID)
	for i := 0; pick == nil && i < 6; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		pick = effectstest.LatestPickTarget(g, me.ID)
	}
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, plain) {
		t.Error("a nonlegendary permanent is offered")
	}
	effectstest.PickCard(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(legend) {
		t.Error("the legendary permanent survived")
	}
	if got := effectstest.TmtB1Count(g, opp.ID, "Junk"); got != 1 {
		t.Errorf("its controller has %d Junk tokens, want 1", got)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Junk"); got != 0 {
		t.Errorf("I have %d Junk tokens, want 0", got)
	}
}

// --- Codsworth, Handy Helper ----------------------------------------------

func TestGapB65CodsworthManaPaysOnlyForAuraAndEquipmentSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	cod := effectstest.PushCatalogPermanent(g, me.ID, "Codsworth, Handy Helper", "Legendary Artifact Creature — Robot", gapB65Codsworth, false)
	if err := g.ActivateManaAbility(me.ID, cod, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	creature := effectstest.PushCatalogHandCard(me, "Test Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == creature {
				me.Hand.Cards[i].ManaCost = "{W}{W}"
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 2, 2
			}
		}
	})
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Codsworth's mana paid for a creature spell")
	}
	sword := effectstest.PushCatalogHandCard(me, "Test Sword", "Artifact — Equipment", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == sword {
				me.Hand.Cards[i].ManaCost = "{W}{W}"
			}
		}
	})
	if err := g.CastSpell(me.ID, sword, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Codsworth's mana did not pay for an Equipment spell: %v", err)
	}
}

func TestGapB65CodsworthAttachesAnEquipmentAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	cod := effectstest.PushCatalogPermanent(g, me.ID, "Codsworth, Handy Helper", "Legendary Artifact Creature — Robot", gapB65Codsworth, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Test Sword", "Artifact — Equipment", "", false)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: sword}, {Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, cod, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("attach: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, sword); c.AttachedTo.ID != bear {
		t.Errorf("the Equipment is attached to %v, want the bear", c.AttachedTo)
	}
}

// --- Aradesh, the Founder -------------------------------------------------

func gapB65AradeshEnlist(t *testing.T, first string) (g *game.Game, me *game.Player, aradesh uuid.UUID, drew int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	aradesh = effectstest.B12Push(g, me.ID, "Aradesh, the Founder", "Legendary Creature — Human Soldier", gapB65Aradesh, 1, 4)
	mate := effectstest.B12Push(g, me.ID, "Mate", "Creature — Bear", "", 3, 3)
	hand := me.Hand.Size()
	r35EnlistDeclare(t, g, aradesh, mate)
	r35EnlistSettle(t, g, first)
	return g, me, aradesh, me.Hand.Size() - hand
}

// Aradesh is 1/4; the enlist trigger adds the enlisted creature's power (3).
// Power is read as Aradesh's ability resolves, so it draws only when the
// enlist trigger resolved first. The attacker has double strike either way.
func TestGapB65AradeshGivesAnEnlistingAttackerDoubleStrikeAndDrawsAtFourPower(t *testing.T) {
	var drawsByOrder [2]int
	for i, first := range []string{"Aradesh", "Enlist"} {
		g, _, aradesh, drew := gapB65AradeshEnlist(t, first)
		if !effectstest.HasEffectiveKeyword(t, g, aradesh, "double strike") {
			t.Errorf("order %q: the enlisting attacker did not gain double strike", first)
		}
		drawsByOrder[i] = drew
	}
	if drawsByOrder[0]+drawsByOrder[1] != 1 {
		t.Errorf("draws by trigger order %v: want exactly one order (the enlist pump first) to draw one card", drawsByOrder)
	}
}

func TestGapB65AradeshDoesNothingWhenNoCreatureEnlisted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aradesh := effectstest.B12Push(g, me.ID, "Aradesh, the Founder", "Legendary Creature — Human Soldier", gapB65Aradesh, 1, 4)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(aradesh, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	r35EnlistSettle(t, g, "")
	if effectstest.HasEffectiveKeyword(t, g, aradesh, "double strike") {
		t.Error("an attacker that enlisted nothing gained double strike")
	}
	if me.Hand.Size() != hand {
		t.Error("a card was drawn")
	}
}

// --- Jason Bright ---------------------------------------------------------

func TestGapB65JasonDrawsOnlyWhenTheDyingZombieOrMutantHadOddPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.B27Push(g, me.ID, "Jason Bright, Glowing Prophet", "Legendary Creature — Zombie Mutant Advisor", gapB65Jason, "{2}{U}", 2, 3, "U")
	plain := effectstest.B12Creature(g, me.ID, "Plain Zombie", "Creature — Zombie", 2, 2)
	pumped := effectstest.B12Creature(g, me.ID, "Pumped Mutant", "Creature — Mutant", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Pumped Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(pumped, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
	})
	hand := me.Hand.Size()
	effectstest.B27Kill(g, plain)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("drew %d from a plain Zombie and a pumped Bear, want 0", me.Hand.Size()-hand)
	}
	effectstest.B27Kill(g, pumped)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d from a pumped Mutant, want 1", got)
	}
}

func TestGapB65JasonSacrificesACreatureForACounterAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	jason := effectstest.PushCatalogPermanent(g, me.ID, "Jason Bright, Glowing Prophet", "Legendary Creature — Zombie Mutant Advisor", gapB65Jason, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, jason, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := gapB65PlusOnes(t, g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
}

// --- Arcade Gannon --------------------------------------------------------

func TestGapB65ArcadeLootsAndGainsAQuestCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	arcade := effectstest.PushCatalogPermanent(g, me.ID, "Arcade Gannon", "Legendary Creature — Human Doctor", gapB65Arcade, false)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, arcade, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("loot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d before the discard, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := effectstest.LayeredCard(t, g, arcade).Counters["quest"]; got != 1 {
		t.Errorf("%d quest counters, want 1", got)
	}
}

func TestGapB65ArcadeCastsAnArtifactOrHumanFromTheGraveyardOncePerTurnUpToItsQuestCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	arcade := effectstest.PushCatalogPermanent(g, me.ID, "Arcade Gannon", "Legendary Creature — Human Doctor", gapB65Arcade, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(arcade, "quest", 2) })
	human := effectstest.PushGraveyardPermanent(me, "Cheap Human", "Creature — Human", "{2}")
	rock := effectstest.PushGraveyardPermanent(me, "Cheap Rock", "Artifact", "{1}")
	pricey := effectstest.PushGraveyardPermanent(me, "Pricey Human", "Creature — Human", "{3}")
	bear := effectstest.PushGraveyardPermanent(me, "Cheap Bear", "Creature — Bear", "{1}")
	for name, id := range map[string]uuid.UUID{"a mana value 3 Human": pricey, "a non-Human nonartifact": bear} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
			t.Errorf("cast %s from the graveyard", name)
		}
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.CastSpell(me.ID, human, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("a mana value 2 Human was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(human) {
		t.Fatal("the Human did not resolve")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a second spell was cast from the graveyard in the same turn")
	}
}

// --- James, Wandering Dad // Follow Him -----------------------------------

func TestGapB65JamesManaPaysOnlyForAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	james := effectstest.PushCatalogPermanent(g, me.ID, "James, Wandering Dad", "Legendary Creature — Human Scientist", gapB65James, false)
	if err := g.ActivateManaAbility(me.ID, james, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	spell := effectstest.PushCatalogHandCard(me, "Test Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == spell {
				me.Hand.Cards[i].ManaCost = "{C}{C}"
			}
		}
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("James's mana paid for a spell")
	}
}

func TestGapB65FollowHimInvestigatesXTimesThenJamesCanBeCastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: gapB65James, Layout: game.LayoutAdventure, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "James, Wandering Dad", TypeLine: "Legendary Creature — Human Scientist", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 4},
			{Name: "Follow Him", TypeLine: "Instant — Adventure", ManaCost: "{X}{U}{U}", Colors: []string{"U"}},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.ToMain(t, g)
	effectstest.FloatMana(t, g, me, "{U}{U}{C}{C}")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, XValue: 2, Strict: true}); err != nil {
		t.Fatalf("cast Follow Him: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB65Clues(g, me.ID); got != 2 {
		t.Errorf("%d Clues from Follow Him with X=2, want 2", got)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card was not exiled")
	}
}

// --- Curie, Emergent Intelligence -----------------------------------------

func TestGapB65CurieDrawsCardsEqualToItsBasePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	curie := effectstest.B27Push(g, me.ID, "Curie, Emergent Intelligence", "Legendary Artifact Creature — Robot", gapB65Curie, "{1}{U}", 1, 3, "U")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(curie, game.CounterPlusOne, 2) })
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, curie, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards with base power 1 and two counters, want 1", got)
	}
}

func TestGapB65CurieExilesAnotherNontokenArtifactCreatureToBecomeACopyThatStillDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.ToMain(t, g)
	curie := effectstest.B27Push(g, me.ID, "Curie, Emergent Intelligence", "Legendary Artifact Creature — Robot", gapB65Curie, "{1}{U}", 1, 3, "U")
	golem := effectstest.B12Creature(g, me.ID, "Big Golem", "Artifact Creature — Golem", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for name, bad := range map[string]uuid.UUID{"a nonartifact creature": bear, "Curie herself": curie} {
		effectstest.FloatMana(t, g, me, "{U}{C}")
		if err := g.ActivateCatalogAbility(me.ID, curie, 0, game.ActivateAbilityParams{Strict: true, ExilePermanentIDs: []uuid.UUID{bad}}); err == nil {
			t.Errorf("activated exiling %s", name)
		}
	}
	effectstest.B16Activate(t, g, me.ID, curie, 0, game.ActivateAbilityParams{Strict: true, ExilePermanentIDs: []uuid.UUID{golem}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(golem) {
		t.Fatal("the Golem was not exiled")
	}
	if p, tt := effectstest.PtOf(t, g, curie); p != 4 || tt != 4 {
		t.Fatalf("Curie is %d/%d after copying a 4/4, want 4/4", p, tt)
	}
	if c := effectstest.LayeredCard(t, g, curie); c.Name != "Big Golem" {
		t.Errorf("Curie is named %q, want Big Golem", c.Name)
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, curie, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("the copy drew %d cards, want 4 (its base power)", got)
	}
}
