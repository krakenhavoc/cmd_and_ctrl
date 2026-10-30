package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m2_cards_test.go — Marvel Super Heroes Commander, M–Z batch 2 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM2SkyriderOracle   = "ea0b0c71-4a1b-49aa-b4cd-db6c5d26b95e"
	mscM2VariantOracle    = "fa4ed027-6e18-4a87-9845-e13f274d9994"
	mscM2StarlingOracle   = "13fd7ca6-c278-409f-a9c4-57f16051ffce"
	mscM2CatOracle        = "eb54d807-624b-471e-8c44-07322003903e"
	mscM2TaxiOracle       = "a17a92f7-e95c-4d0e-a43e-7e6b41a59139"
	mscM2WhooshOracle     = "855353e2-b686-482a-b987-c6ae456f088c"
	mscM2AllianceOracle   = "e59c07ab-08fd-4d5b-978e-1594604ba410"
	mscM2WisecrackOracle  = "e9c9a445-4386-4616-ad00-f08064d73dcd"
	mscM2ZuriOracle       = "aa792980-0811-4c08-8343-00fba0c5cf5e"
	mscM2PatrolOracle     = "9851f207-9f75-4ceb-8024-8f9bf21f90ad"
	mscM2SpearmasterOracl = "9ae33bde-ccc9-40e2-9957-05cd9725f8cf"
	mscM2BotsOracle       = "facc26d5-bb53-4f91-9b8c-73bf4dd44fcd"
)

func TestMscM2EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscM2SkyriderOracle:   "Valkyrior Skyrider",
		mscM2VariantOracle:    "Virtuous Variant",
		mscM2StarlingOracle:   "Starling, Aerial Ally",
		mscM2CatOracle:        "Venomized Cat",
		mscM2TaxiOracle:       "Taxi Driver",
		mscM2WhooshOracle:     "Whoosh!",
		mscM2AllianceOracle:   "Ultimate Alliance",
		mscM2WisecrackOracle:  "Wisecrack",
		mscM2ZuriOracle:       "Zuri, Warrior of Wakanda",
		mscM2PatrolOracle:     "Shipwreck Patrol",
		mscM2SpearmasterOracl: "Sif's Spearmaster",
		mscM2BotsOracle:       "Repulsor Bots",
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

// mscM2PickTargets answers the open pick_target prompt with cards.
func mscM2PickTargets(t *testing.T, g *game.Game, ids ...uuid.UUID) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	refs := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		refs[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func mscM2Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

func mscM2OnBattlefield(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

// --- Valkyrior Skyrider ---

func TestMscM2ValkyriorSkyriderGainsFourOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	id := castHoldCreature(t, g, "Valkyrior Skyrider", "Creature — God Warrior Hero", mscM2SkyriderOracle, 3, 4)
	if me.Life != before+4 {
		t.Errorf("life %d -> %d, want +4", before, me.Life)
	}
	c := mscM2Live(t, g, id)
	if !game.HasKeyword(&c, "flying") {
		t.Errorf("keywords %v, want flying", c.Keywords)
	}
}

// --- Virtuous Variant ---

func TestMscM2VirtuousVariantPutsACounterOnACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := castHoldCreature(t, g, "Virtuous Variant", "Creature — Spider Human Hero", mscM2VariantOracle, 1, 2)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("a creature the opponent controls was a legal target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 || plusOnes(g, theirs) != 0 || plusOnes(g, id) != 0 {
		t.Errorf("counters: mine %d, theirs %d, variant %d; want 1, 0, 0", plusOnes(g, bear), plusOnes(g, theirs), plusOnes(g, id))
	}
	c := mscM2Live(t, g, id)
	if !game.HasKeyword(&c, "flying") {
		t.Errorf("keywords %v, want flying", c.Keywords)
	}
}

// --- Starling, Aerial Ally ---

func TestMscM2StarlingGivesAnotherCreatureFlyingUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := castHoldCreature(t, g, "Starling, Aerial Ally", "Legendary Creature — Human Hero", mscM2StarlingOracle, 3, 4)
	// Starling itself is not a legal pick ("another"), so the bear is the only one.
	if ch := latestPickTarget(g, me.ID); ch != nil {
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
	}
	b := mscM2Live(t, g, bear)
	if !game.HasKeyword(&b, "flying") {
		t.Errorf("the bear's keywords %v, want flying", b.Keywords)
	}
	s := mscM2Live(t, g, id)
	if !game.HasKeyword(&s, "flying") {
		t.Errorf("Starling keywords %v, want her printed flying", s.Keywords)
	}
}

func TestMscM2StarlingAloneHasNothingToTarget(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Starling, Aerial Ally", "Legendary Creature — Human Hero", mscM2StarlingOracle, 3, 4)
	if ch := latestPickTarget(g, g.Seats[0].ID); ch != nil {
		t.Error("Starling alone was offered a target")
	}
}

// --- Venomized Cat ---

func TestMscM2VenomizedCatMillsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	lib, gy := me.Library.Size(), me.Graveyard.Size()
	id := castHoldCreature(t, g, "Venomized Cat", "Creature — Symbiote Cat Villain", mscM2CatOracle, 2, 3)
	if me.Library.Size() != lib-2 || me.Graveyard.Size() != gy+2 {
		t.Errorf("library %d -> %d, graveyard %d -> %d; want two milled", lib, me.Library.Size(), gy, me.Graveyard.Size())
	}
	c := mscM2Live(t, g, id)
	if !game.HasKeyword(&c, "deathtouch") {
		t.Errorf("keywords %v, want deathtouch", c.Keywords)
	}
}

// --- Taxi Driver ---

func TestMscM2TaxiDriverGivesHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	taxi := pushCatalogPermanent(g, me.ID, "Taxi Driver", "Creature — Human Pilot", mscM2TaxiOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, taxi, 0, game.ActivateAbilityParams{
		Targets: b16TargetCard(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := mscM2Live(t, g, bear)
	if !game.HasKeyword(&c, "haste") {
		t.Errorf("the bear's keywords %v, want haste", c.Keywords)
	}
	if tc := mscM2Live(t, g, taxi); !tc.Tapped {
		t.Error("the {T} cost was not paid")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left floating, the {1} cost should have spent it", me.ManaPool.Total())
	}
}

// --- Whoosh! ---

func TestMscM2WhooshBouncesAndDrawsOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		draws    int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			seedLibrary(me, "A", "B", "C")
			target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			hand := me.Hand.Size()
			if _, err := castWithOptionalCosts(t, g, "Whoosh!", "Instant", mscM2WhooshOracle, b16TargetCard(target), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if mscM2OnBattlefield(g, target) || !opp.Hand.Contains(target) {
				t.Error("the target was not returned to its owner's hand")
			}
			// The spell left the hand when cast.
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("hand changed by %d, want %d draws", got, tc.draws)
			}
		})
	}
}

func TestMscM2WhooshRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	land := b12Push(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	if _, err := castWithOptionalCosts(t, g, "Whoosh!", "Instant", mscM2WhooshOracle, b16TargetCard(land), nil, nil); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Ultimate Alliance ---

func TestMscM2UltimateAllianceDealsDamageEqualToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		b12Creature(g, me.ID, "Ally", "Creature — Hero", 1, 1)
	}
	b12Creature(g, opp.ID, "Their Other", "Creature — Bear", 1, 1)
	victim := b12Creature(g, opp.ID, "Big Victim", "Creature — Giant", 3, 6)
	castCatalogSpell(t, g, "Ultimate Alliance", "Instant", mscM2AllianceOracle, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	c := battlefieldCardPtr(t, g, victim)
	if c.DamageMarked != 3 {
		t.Errorf("damage marked %d, want 3 (the caster's three creatures)", c.DamageMarked)
	}
}

// --- Wisecrack ---

func TestMscM2WisecrackHitsTheCreatureWithItsOwnPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Big Victim", "Creature — Giant", 3, 6)
	life := opp.Life
	castCatalogSpell(t, g, "Wisecrack", "Instant", mscM2WisecrackOracle, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	if c := battlefieldCardPtr(t, g, victim); c.DamageMarked != 3 {
		t.Errorf("damage marked %d, want its power, 3", c.DamageMarked)
	}
	if opp.Life != life {
		t.Errorf("the controller lost %d life though the creature was not attacking", life-opp.Life)
	}
}

func TestMscM2WisecrackAlsoHitsTheControllerOfAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Big Attacker", "Creature — Giant", 3, 6)
	battlefieldCardPtr(t, g, victim).AttackingTarget = me.ID
	life := opp.Life
	castCatalogSpell(t, g, "Wisecrack", "Instant", mscM2WisecrackOracle, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	if c := battlefieldCardPtr(t, g, victim); c.DamageMarked != 3 {
		t.Errorf("damage marked %d, want 3", c.DamageMarked)
	}
	if opp.Life != life-2 {
		t.Errorf("the attacker's controller went %d -> %d, want -2", life, opp.Life)
	}
}

func TestMscM2WisecrackKillsAFragileCreature(t *testing.T) {
	g := newCatalogGame(t)
	victim := b12Creature(g, g.Seats[1].ID, "Glass Cannon", "Creature — Giant", 4, 4)
	castCatalogSpell(t, g, "Wisecrack", "Instant", mscM2WisecrackOracle, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	if mscM2OnBattlefield(g, victim) {
		t.Error("a 4/4 dealt 4 damage by itself should die")
	}
}

// --- Zuri, Warrior of Wakanda ---

func mscM2CastArtifact(t *testing.T, g *game.Game, cost, mana string) {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Gizmo", TypeLine: "Artifact", ManaCost: cost,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	floatMana(t, g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestMscM2ZuriCountersEveryCreatureOnABigArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zuri := b12Push(g, me.ID, "Zuri, Warrior of Wakanda", "Legendary Creature — Human Warrior Hero", mscM2ZuriOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mscM2CastArtifact(t, g, "{4}", "{C}{C}{C}{C}")
	if plusOnes(g, zuri) != 1 || plusOnes(g, bear) != 1 || plusOnes(g, theirs) != 0 {
		t.Errorf("counters: Zuri %d, bear %d, theirs %d; want 1, 1, 0", plusOnes(g, zuri), plusOnes(g, bear), plusOnes(g, theirs))
	}
}

func TestMscM2ZuriIgnoresASmallArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	zuri := b12Push(g, g.Seats[0].ID, "Zuri, Warrior of Wakanda", "Legendary Creature — Human Warrior Hero", mscM2ZuriOracle, 2, 2)
	mscM2CastArtifact(t, g, "{3}", "{C}{C}{C}")
	if plusOnes(g, zuri) != 0 {
		t.Errorf("a mana value 3 artifact gave Zuri %d counters", plusOnes(g, zuri))
	}
}

func TestMscM2ZuriIgnoresABigNonartifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zuri := b12Push(g, me.ID, "Zuri, Warrior of Wakanda", "Legendary Creature — Human Warrior Hero", mscM2ZuriOracle, 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{4}",
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, zuri) != 0 {
		t.Errorf("a sorcery gave Zuri %d counters", plusOnes(g, zuri))
	}
}

// --- Shipwreck Patrol ---

func TestMscM2ShipwreckPatrolTapsAndStunsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Shipwreck Patrol", "Creature — Merfolk Warrior", mscM2PatrolOracle, 3, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("a creature the caster controls was a legal target")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	c := mscM2Live(t, g, theirs)
	if !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("their bear: tapped %v, stun counters %d; want tapped with one", c.Tapped, c.Counters[game.CounterStun])
	}
	if m := mscM2Live(t, g, mine); m.Tapped || m.Counters[game.CounterStun] != 0 {
		t.Error("the caster's own creature was affected")
	}
}

// --- Sif's Spearmaster ---

func TestMscM2SifsSpearmasterDealsItsPowerToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sif := b12Push(g, me.ID, "Sif's Spearmaster", "Creature — God Warrior Hero", mscM2SpearmasterOracl, 3, 4)
	battlefieldCardPtr(t, g, sif).SummonedThisTurn = false
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, sif, 0, game.ActivateAbilityParams{
		Targets: b16TargetPlayer(opp.ID),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d, want the Spearmaster's power, 3", got)
	}
}

func TestMscM2SifsSpearmasterCannotTargetYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sif := pushCatalogPermanent(g, me.ID, "Sif's Spearmaster", "Creature — God Warrior Hero", mscM2SpearmasterOracl, false)
	if err := g.ActivateCatalogAbility(me.ID, sif, 0, game.ActivateAbilityParams{
		Targets: b16TargetPlayer(me.ID),
	}); err == nil {
		t.Error("the Spearmaster targeted its own controller")
	}
}

// --- Repulsor Bots ---

func TestMscM2RepulsorBotsReturnsUpToTwoOtherArtifactsOrCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	spare := b12Creature(g, opp.ID, "Spare Bear", "Creature — Bear", 2, 2)
	id := castHoldCreature(t, g, "Repulsor Bots", "Artifact Creature — Robot Hero", mscM2BotsOracle, 4, 4)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: spare},
	}); err == nil {
		t.Fatal("three targets were accepted for an up-to-two clause")
	}
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, []game.TargetRef{
		{Kind: game.TargetCard, ID: id},
	}); err == nil {
		t.Fatal("the Bots targeted themselves")
	}
	mscM2PickTargets(t, g, bear, rock)
	if mscM2OnBattlefield(g, bear) || mscM2OnBattlefield(g, rock) {
		t.Error("a chosen permanent stayed on the battlefield")
	}
	if !opp.Hand.Contains(bear) || !me.Hand.Contains(rock) {
		t.Error("a chosen permanent did not return to its owner's hand")
	}
	if !mscM2OnBattlefield(g, spare) || !mscM2OnBattlefield(g, id) {
		t.Error("an unchosen permanent left the battlefield")
	}
}

func TestMscM2RepulsorBotsMayChooseNone(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Repulsor Bots", "Artifact Creature — Robot Hero", mscM2BotsOracle, 4, 4)
	mscM2PickTargets(t, g)
	if !mscM2OnBattlefield(g, bear) {
		t.Error("a creature was returned though no target was chosen")
	}
}
