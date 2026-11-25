package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b197_test.go — Core Set 2019 planeswalkers, pump and tokens, S82 set
// OLD batch 197 (#5527). Every test plays the real catalog entry through the
// engine.

const (
	oldB197Ajani     = "341f198b-beb6-4a44-af91-a64a5989c865"
	oldB197Vivien    = "34fe53c2-8909-4786-b5c0-331de997ff81"
	oldB197Tezzeret  = "e5a95145-dc29-4bfc-bb14-b7a269268366"
	oldB197Liliana   = "f3ad60a6-05d0-48a1-a804-7900d306d613"
	oldB197Ogre      = "0156bfbc-26bc-4699-a7ff-437f7c043716"
	oldB197Ursine    = "4f142d99-8146-4b16-a2cc-ac03da19b960"
	oldB197Amulet    = "98709d61-4da4-4d3a-a654-bb2a7a0edb6b"
	oldB197Druid     = "06c12fcc-1c08-40a0-a0da-e139c42f083f"
	oldB197Lieut     = "2b761ce6-4fb1-4f22-b51c-d9f25f62b39a"
	oldB197Bloodlord = "ce9d3382-23ca-4d35-ad30-5e2b75517012"
)

func TestOldB197CardsAreRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB197Ajani: "Ajani, Wise Counselor", oldB197Vivien: "Vivien of the Arkbow",
		oldB197Tezzeret: "Tezzeret, Cruel Machinist", oldB197Liliana: "Liliana, the Necromancer",
		oldB197Ogre: "Brawl-Bash Ogre", oldB197Ursine: "Ursine Champion",
		oldB197Amulet: "Amulet of Safekeeping", oldB197Druid: "Druid of Horns",
		oldB197Lieut: "Thorn Lieutenant", oldB197Bloodlord: "Regal Bloodlord",
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

func oldB197Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB197Tokens(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.IsToken() && c.Name == name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// --- Ajani, Wise Counselor ---------------------------------------------------

func TestOldB197AjaniPlusTwoGainsLifeForEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	aj := effectstest.S82Walker(g, me.ID, "Ajani, Wise Counselor", oldB197Ajani, 5)
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, aj, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d, want 2 (only my creatures count)", got)
	}
	if got := effectstest.LoyaltyOf(g, aj); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestOldB197AjaniMinusThreePumpsMyCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	aj := effectstest.S82Walker(g, me.ID, "Ajani, Wise Counselor", oldB197Ajani, 5)
	mine := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, aj, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePTOf(t, g, mine); p != 4 || tt != 4 {
		t.Errorf("mine %d/%d, want 4/4", p, tt)
	}
	if p, tt := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("theirs %d/%d, want 2/2", p, tt)
	}
}

func TestOldB197AjaniMinusNinePutsCountersEqualToYourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	aj := effectstest.S82Walker(g, me.ID, "Ajani, Wise Counselor", oldB197Ajani, 9)
	bear := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, aj, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != life {
		t.Errorf("%d counters, want %d (my life total)", got, life)
	}
}

// --- Vivien of the Arkbow ----------------------------------------------------

func TestOldB197VivienPlusTwoPutsTwoCountersOnUpToOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien of the Arkbow", oldB197Vivien, 5)
	bear := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
	if got := effectstest.LoyaltyOf(g, viv); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestOldB197VivienPlusTwoWithNoTargetStillAddsLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien of the Arkbow", oldB197Vivien, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, viv); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestOldB197VivienMinusThreeBitesACreatureYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien of the Arkbow", oldB197Vivien, 5)
	biter := effectstest.B12Creature(g, me.ID, "Biter", "Creature — Wolf", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(biter, victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the victim survived 3 damage")
	}
	if c, ok := effectstest.BattlefieldCard(g, biter); !ok || c.DamageMarked != 0 {
		t.Errorf("the biter took damage back: present %v", ok)
	}
}

func TestOldB197VivienMinusThreeRefusesACreatureYouControlAsTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien of the Arkbow", oldB197Vivien, 5)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Wolf", 3, 3)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, viv, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a, b)})
	if err == nil {
		t.Error("targeting my own creature as the victim was accepted")
	}
}

func TestOldB197VivienMinusNineGivesPlusFourAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien of the Arkbow", oldB197Vivien, 9)
	mine := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePTOf(t, g, mine); p != 6 || tt != 6 {
		t.Errorf("mine %d/%d, want 6/6", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("my creature has no trample")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature was affected")
	}
}

// --- Tezzeret, Cruel Machinist -----------------------------------------------

func TestOldB197TezzeretPlusOneDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	tez := effectstest.S82Walker(g, me.ID, "Tezzeret, Cruel Machinist", oldB197Tezzeret, 4)
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.HandSize(me)
	effectstest.S82Activate(t, g, me.ID, tez, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestOldB197TezzeretZeroMakesAnArtifactAFiveFiveCreatureUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	tez := effectstest.S82Walker(g, me.ID, "Tezzeret, Cruel Machinist", oldB197Tezzeret, 4)
	rock := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Rock", "Artifact", "")
	land := effectstest.B12Permanent(g, me.ID, "Plain Land", "Land")
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, tez, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)})
	if err == nil {
		t.Error("a nonartifact was accepted as the target")
	}
	effectstest.S82Activate(t, g, me.ID, tez, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveType(t, g, rock, "Creature") || !effectstest.HasEffectiveType(t, g, rock, "Artifact") {
		t.Errorf("types %v, want an artifact creature", effectstest.EffectiveTypes(t, g, rock))
	}
	if p, tt := effectstest.EffectivePTOf(t, g, rock); p != 5 || tt != 5 {
		t.Errorf("rock %d/%d, want 5/5", p, tt)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	if effectstest.HasEffectiveType(t, g, rock, "Creature") {
		t.Error("the artifact is still a creature on my next turn")
	}
}

func TestOldB197TezzeretMinusSevenPutsChosenHandCardsOntoTheBattlefieldFaceDownAsFiveFives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	tez := effectstest.S82Walker(g, me.ID, "Tezzeret, Cruel Machinist", oldB197Tezzeret, 7)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushHandCardWithManaCost(me, "Hand Bear", "Creature — Bear", "{1}{G}")
	bolt := effectstest.PushHandCardWithManaCost(me, "Hand Bolt", "Instant", "{R}")
	kept := effectstest.PushHandCardWithManaCost(me, "Kept Card", "Sorcery", "{2}")
	effectstest.S82Activate(t, g, me.ID, tez, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick over the hand")
	}
	if c.ChooseMin != 0 {
		t.Errorf("pick floor %d, want 0 (any number)", c.ChooseMin)
	}
	for _, id := range []uuid.UUID{bear, bolt, kept} {
		if !slices.Contains(c.ChooseCards, id) {
			t.Errorf("a hand card %s is not offered (a nonpermanent is legal face down)", id)
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear, bolt}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range []uuid.UUID{bear, bolt} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("a picked card did not reach the battlefield")
		}
		if !card.FaceDown {
			t.Error("a picked card entered face up")
		}
		if p, tt := effectstest.EffectivePTOf(t, g, id); p != 5 || tt != 5 {
			t.Errorf("a face-down card is %d/%d, want 5/5", p, tt)
		}
		if !effectstest.HasEffectiveType(t, g, id, "Artifact") || !effectstest.HasEffectiveType(t, g, id, "Creature") {
			t.Errorf("types %v, want artifact creature", effectstest.EffectiveTypes(t, g, id))
		}
	}
	if !me.Hand.Contains(kept) {
		t.Error("the unpicked card left the hand")
	}
}

func TestOldB197TezzeretMinusSevenMayPutNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	tez := effectstest.S82Walker(g, me.ID, "Tezzeret, Cruel Machinist", oldB197Tezzeret, 7)
	effectstest.AdvanceToMain(t, g)
	kept := effectstest.PushHandCardWithManaCost(me, "Kept Card", "Creature — Bear", "{2}")
	effectstest.S82Activate(t, g, me.ID, tez, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick over the hand")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(kept) || g.Battlefield.Contains(kept) {
		t.Error("choosing none moved a card")
	}
}

// --- Liliana, the Necromancer ------------------------------------------------

func TestOldB197LilianaPlusOneMakesAPlayerLoseTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	lil := effectstest.S82Walker(g, me.ID, "Liliana, the Necromancer", oldB197Liliana, 4)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.S82Activate(t, g, me.ID, lil, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("lost %d, want 2", got)
	}
}

func TestOldB197LilianaMinusOneReturnsACreatureCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	lil := effectstest.S82Walker(g, me.ID, "Liliana, the Necromancer", oldB197Liliana, 4)
	mine := effectstest.PushGraveyardCardTyped(me, "My Bear", "Creature — Bear")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, lil, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)})
	if err == nil {
		t.Error("a card in an opponent's graveyard was accepted")
	}
	effectstest.S82Activate(t, g, me.ID, lil, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("the creature card did not return to hand")
	}
}

func TestOldB197LilianaMinusSevenDestroysTwoAndReanimatesTwoFromGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	lil := effectstest.S82Walker(g, me.ID, "Liliana, the Necromancer", oldB197Liliana, 7)
	a := effectstest.B12Creature(g, opp.ID, "Victim A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Victim B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	oldCard := effectstest.PushGraveyardCardTyped(opp, "Old Dragon", "Creature — Dragon")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, lil, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a target survived")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an untargeted creature was destroyed")
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick over the graveyards")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("pick %d..%d, want 0..2", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range []uuid.UUID{a, b, oldCard} {
		if !slices.Contains(pick.ChooseCards, id) {
			t.Errorf("creature card %s from a graveyard (the destroyed ones included) is not offered", id)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, oldCard}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range []uuid.UUID{a, oldCard} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok || card.Controller != me.ID {
			t.Errorf("a picked card is not on my side of the battlefield (present %v)", ok)
		}
	}
	if g.Battlefield.Contains(b) {
		t.Error("the card not picked came back")
	}
}

// --- Brawl-Bash Ogre ----------------------------------------------------------

func TestOldB197OgreSacrificesAnotherCreatureToGainTwoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	ogre := effectstest.B12Push(g, me.ID, "Brawl-Bash Ogre", "Creature — Ogre Warrior", oldB197Ogre, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{ogre}); err == nil {
		t.Fatal("sacrificed the Ogre itself")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still there")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, ogre); p != 5 || tt != 5 {
		t.Errorf("Ogre %d/%d, want 5/5", p, tt)
	}
}

func TestOldB197OgreDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	ogre := effectstest.B12Push(g, me.ID, "Brawl-Bash Ogre", "Creature — Ogre Warrior", oldB197Ogre, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) {
		t.Error("declining still sacrificed")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, ogre); p != 3 || tt != 3 {
		t.Errorf("Ogre %d/%d, want 3/3", p, tt)
	}
}

// --- Ursine Champion ----------------------------------------------------------

func TestOldB197UrsineChampionPumpsAndBecomesABearBerserkerOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	champ := effectstest.B12Push(g, me.ID, "Ursine Champion", "Creature — Human Berserker", oldB197Ursine, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePTOf(t, g, champ); p != 5 || tt != 5 {
		t.Errorf("champion %d/%d, want 5/5", p, tt)
	}
	subs := slices.Clone(effectstest.EffectiveSubtypes(t, g, champ))
	slices.Sort(subs)
	if !slices.Equal(subs, []string{"Bear", "Berserker"}) {
		t.Errorf("subtypes %v, want Bear and Berserker only (Human is replaced)", subs)
	}
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation in the same turn was accepted")
	}
}

// --- Amulet of Safekeeping ----------------------------------------------------

func TestOldB197AmuletShrinksEveryCreatureToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	effectstest.PushPermanentForTest(g, me.ID, "Amulet of Safekeeping", oldB197Amulet, "Artifact")
	mine := effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Beast"))
	theirs := effectstest.PushToken(g, opp.ID, effects.TokenCard("1/1 green Elf Warrior"))
	card := effectstest.B12Creature(g, opp.ID, "Real Bear", "Creature — Bear", 2, 2)
	if p, tt := effectstest.EffectivePTOf(t, g, mine); p != 2 || tt != 3 {
		t.Errorf("my token %d/%d, want 2/3", p, tt)
	}
	if p, tt := effectstest.EffectivePTOf(t, g, theirs); p != 0 || tt != 1 {
		t.Errorf("their token %d/%d, want 0/1", p, tt)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, card); p != 2 {
		t.Errorf("a nontoken creature has power %d, want 2", p)
	}
}

func TestOldB197AmuletCountersAnOpponentsSpellAimedAtYouUnlessTheyPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	effectstest.PushPermanentForTest(g, me.ID, "Amulet of Safekeeping", oldB197Amulet, "Artifact")
	life := me.Life
	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no payment prompt for the spell's controller")
	}
	if c := effectstest.PendingOfKind(g, game.PendingChoicePayUnless); c == nil || c.PayCost != "{1}" {
		t.Errorf("the tax is %+v, want {1}", c)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Errorf("declining must counter the Bolt: in graveyard %v, life %d (was %d)", opp.Graveyard.Contains(bolt), me.Life, life)
	}
}

func TestOldB197AmuletIgnoresMyOwnSpellsAndSpellsAimedAtOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	effectstest.PushPermanentForTest(g, me.ID, "Amulet of Safekeeping", oldB197Amulet, "Artifact")
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	life := other.Life
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: other.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("a spell aimed at another player asked for a payment")
	}
	if other.Life != life-3 {
		t.Errorf("the Bolt did not resolve: life %d, want %d", other.Life, life-3)
	}
}

// --- Druid of Horns -----------------------------------------------------------

func TestOldB197DruidMakesABeastWhenYouCastAnAuraAtIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	druid := effectstest.B12Push(g, me.ID, "Druid of Horns", "Creature — Human Druid", oldB197Druid, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", effectstest.PacifismOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: druid}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB197Tokens(g, me.ID, "Beast")); got != 1 {
		t.Errorf("%d Beast tokens, want 1", got)
	}
}

func TestOldB197DruidIgnoresASpellThatIsNotAnAuraAndAnAuraAtAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	druid := effectstest.B12Push(g, me.ID, "Druid of Horns", "Creature — Human Druid", oldB197Druid, 2, 3)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", effectstest.PacifismOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: other}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: druid})
	if got := len(oldB197Tokens(g, me.ID, "Beast")); got != 0 {
		t.Errorf("%d Beast tokens, want 0", got)
	}
}

// --- Thorn Lieutenant ---------------------------------------------------------

func TestOldB197LieutenantMakesAnElfWhenAnOpponentTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB197Seats(g)
	lt := effectstest.B12Push(g, me.ID, "Thorn Lieutenant", "Creature — Elf Warrior", oldB197Lieut, 2, 3)
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: lt})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB197Tokens(g, me.ID, "Elf Warrior")); got != 1 {
		t.Errorf("%d Elf Warrior tokens, want 1", got)
	}
}

func TestOldB197LieutenantIgnoresMyOwnTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	lt := effectstest.B12Push(g, me.ID, "Thorn Lieutenant", "Creature — Elf Warrior", oldB197Lieut, 2, 3)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: lt})
	if got := len(oldB197Tokens(g, me.ID, "Elf Warrior")); got != 0 {
		t.Errorf("%d Elf Warrior tokens from my own spell, want 0", got)
	}
}

func TestOldB197LieutenantPumpsForSixMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	lt := effectstest.B12Push(g, me.ID, "Thorn Lieutenant", "Creature — Elf Warrior", oldB197Lieut, 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, lt, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePTOf(t, g, lt); p != 6 || tt != 7 {
		t.Errorf("Lieutenant %d/%d, want 6/7", p, tt)
	}
}

// --- Regal Bloodlord ----------------------------------------------------------

func TestOldB197BloodlordMakesABatAtEachEndStepIfYouGainedLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	effectstest.B12Push(g, me.ID, "Regal Bloodlord", "Creature — Vampire Soldier", oldB197Bloodlord, 2, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	bats := oldB197Tokens(g, me.ID, "Bat")
	if len(bats) != 1 {
		t.Fatalf("%d Bat tokens, want 1", len(bats))
	}
	if !effectstest.HasEffectiveKeyword(t, g, bats[0], "flying") {
		t.Error("the Bat has no flying")
	}
}

func TestOldB197BloodlordMakesNoBatWithoutLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB197Seats(g)
	effectstest.B12Push(g, me.ID, "Regal Bloodlord", "Creature — Vampire Soldier", oldB197Bloodlord, 2, 4)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB197Tokens(g, me.ID, "Bat")); got != 0 {
		t.Errorf("%d Bat tokens, want 0", got)
	}
}
