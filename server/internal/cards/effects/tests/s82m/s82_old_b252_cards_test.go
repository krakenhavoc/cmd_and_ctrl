package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b252_cards_test.go — Aether Revolt cards, S82 OLD batch 252
// (#5536). Every test plays the real catalog entry through the engine.

const (
	b252Perilous        = "169728d6-7820-4a6d-a815-70714ad3e887"
	b252ReleaseGremlins = "3b583358-d0d5-4f53-b06c-c6439156e684"
	b252Getaway         = "4b77479e-8d49-46f6-b774-d8caf151c15c"
	b252SramsExpertise  = "5cd17347-5558-46f9-b8bb-ae5deeb1f229"
	b252Obsolescence    = "fbafbe3a-3980-4457-a77b-3f72df7b3329"
	b252Conviction      = "4374b6b1-e691-4823-802d-8de8530aa461"
	b252GremlinInfest   = "94fa21f1-826a-4d51-9ccd-d72e5cf13f3d"
	b252SiegeMod        = "ac6cf9a7-732d-4ff0-9e88-aae97a2ff169"
	b252TezzeretsTouch  = "c6f2c767-3c8e-4113-b271-ea7c19583e6c"
	b252UnbridledGrowth = "d749d075-6135-486a-a94a-36c3142960d7"
	b252Bear            = "Creature — Bear"
	b252AuraLine        = "Enchantment — Aura"
	b252ArtifactCreat   = "Artifact Creature — Construct"
)

func b252Aura(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.B12PlayFromHand(t, g, name, b252AuraLine, oracle, game.CastSpellParams{Targets: effectstest.CardRefs(target)})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestB252PerilousPredicamentTakesOneOfEachKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	art := effectstest.B12Creature(g, opp.ID, "Gearbot", b252ArtifactCreat, 1, 1)
	otherArt := effectstest.B12Creature(g, opp.ID, "Gearbot Two", b252ArtifactCreat, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b252Bear, 2, 2)
	otherBear := effectstest.B12Creature(g, opp.ID, "Bear Two", b252Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Perilous Predicament", "Instant", b252Perilous, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, art)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) || g.Battlefield.Contains(bear) {
		t.Error("the chosen creatures were not sacrificed")
	}
	if !g.Battlefield.Contains(otherArt) || !g.Battlefield.Contains(otherBear) {
		t.Error("more than one creature of a kind was sacrificed")
	}
}

func TestB252ReleaseTheGremlinsDestroysXAndMakesX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Permanent(g, opp.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Rock B", "Artifact")
	c := effectstest.B12Permanent(g, opp.ID, "Rock C", "Artifact")
	if err := effectstest.CastX(t, g, "Release the Gremlins", "Sorcery", b252ReleaseGremlins, 2, effectstest.CardRefs(a, b)); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(c) {
		t.Error("exactly the two targeted artifacts should be destroyed")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Gremlin"); got != 2 {
		t.Errorf("%d Gremlin tokens, want 2", got)
	}
	if err := effectstest.CastX(t, g, "Release the Gremlins", "Sorcery", b252ReleaseGremlins, 2, effectstest.CardRefs(c)); err == nil {
		t.Error("X = 2 was accepted with one target")
	}
}

func TestB252RenegadesGetawayProtectsAndMakesAServo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b252Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Renegade's Getaway", "Instant", b252Getaway, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the target did not gain indestructible")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Servo"); got != 1 {
		t.Errorf("%d Servo tokens, want 1", got)
	}
}

func TestB252SramsExpertiseMakesServosAndOffersACheapFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := uuid.New()
	big := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Bear", TypeLine: b252Bear, ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Bear", TypeLine: b252Bear, ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Sram's Expertise", "Sorcery", b252SramsExpertise, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.OnBattlefieldNamed(g, "Servo"); got != 3 {
		t.Errorf("%d Servo tokens, want 3", got)
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheap)
	for _, id := range offer.DiscardOptions {
		if id == big {
			t.Error("a spell with mana value 4 was offered")
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Cheap Bear") {
		t.Error("the free spell did not resolve")
	}
}

func TestB252NaturalObsolescenceTucksAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	before := opp.Library.Size()
	if err := effectstest.CastCatalogSpellErr(t, g, "Natural Obsolescence", "Instant", b252Obsolescence, effectstest.CardRefs(effectstest.B12Creature(g, opp.ID, "Bear", b252Bear, 2, 2))); err == nil {
		t.Fatal("a non-artifact creature was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Natural Obsolescence", "Instant", b252Obsolescence, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || opp.Library.Size() != before+1 {
		t.Errorf("the artifact is not on the bottom of its owner's library (library %d, was %d)", opp.Library.Size(), before)
	}
	if last := opp.Library.Cards[0]; last.InstanceID != rock {
		t.Error("the artifact is not the bottom card")
	}
}

func TestB252ConvictionPumpsAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b252Bear, 2, 2)
	aura := b252Aura(t, g, "Conviction", b252Conviction, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("the enchanted creature is %d/%d, want 3/5", p, tough)
	}
	effectstest.AddPool(t, g, me.ID, "{W}")
	effectstest.ApaActivate(t, g, me, aura, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(aura) || !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the creature is %d/%d after the Aura left, want 2/2", p, tough)
	}
}

func TestB252GremlinInfestationBurnsAndLeavesAGremlin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	b252Aura(t, g, "Gremlin Infestation", b252GremlinInfest, rock)
	theirLife := opp.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 {
		t.Errorf("the artifact's controller is at %d, want %d", opp.Life, theirLife-2)
	}
	_ = me
}

func TestB252GremlinInfestationMakesAGremlinWhenTheArtifactDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	b252Aura(t, g, "Gremlin Infestation", b252GremlinInfest, rock)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.OnBattlefieldNamed(g, "Gremlin"); got != 1 {
		t.Errorf("%d Gremlin tokens, want 1", got)
	}
}

func TestB252SiegeModificationAnimatesAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.B12Push(g, me.ID, "Test Wagon", "Artifact — Vehicle", "", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", b252Bear, 2, 2)
	b252Aura(t, g, "Siege Modification", b252SiegeMod, vehicle)
	c := effectstest.B12Card(t, g, vehicle)
	if !c.IsCreature() {
		t.Error("the enchanted Vehicle is not a creature")
	}
	if p, _ := effectstest.PtOf(t, g, vehicle); p != 6 {
		t.Errorf("the Vehicle has power %d, want 6", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, vehicle, "first strike") {
		t.Error("the Vehicle has no first strike")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	bear2 := effectstest.B12Creature(g2, me2.ID, "Bear", b252Bear, 2, 2)
	b252Aura(t, g2, "Siege Modification", b252SiegeMod, bear2)
	if p, tough := effectstest.PtOf(t, g2, bear2); p != 5 || tough != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 5/2", p, tough)
	}
	_ = bear
}

func TestB252TezzeretsTouchMakesAFiveFiveAndReturnsTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	b252Aura(t, g, "Tezzeret's Touch", b252TezzeretsTouch, rock)
	c := effectstest.B12Card(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("the artifact is %q, want an artifact creature", c.TypeLine)
	}
	if p, tough := effectstest.PtOf(t, g, rock); p != 5 || tough != 5 {
		t.Errorf("the artifact is %d/%d, want 5/5", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

func TestB252UnbridledGrowthGrantsManaAndCanBeSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	aura := b252Aura(t, g, "Unbridled Growth", b252UnbridledGrowth, land)
	idx, ref, ok := effectstest.GrantedManaAbilityFor(g, land)
	if !ok {
		t.Fatal("the enchanted land has no granted mana ability")
	}
	if err := g.ActivateManaAbility(me.ID, land, idx, game.ManaAbilityParams{Ref: ref, Colors: []string{"U"}}); err != nil {
		t.Fatalf("activate granted ability: %v", err)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}
