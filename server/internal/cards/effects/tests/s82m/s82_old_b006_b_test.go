package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 old sets, batch 006, group B: Chandra, Flame's Catalyst; Estrid, the
// Masked; Tezzeret, Agent of Bolas (#5462).

const (
	oldB006ChandraOracle  = "7721e800-fba6-4ae7-855e-631b2ecc8d6b"
	oldB006EstridOracle   = "7cf1fa2e-1572-4d9d-8e93-e49016b34458"
	oldB006TezzeretOracle = "8ab7358e-a0f5-4220-8769-25177237bcdf"
)

func TestS82OldB006BCardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB006ChandraOracle:  "Chandra, Flame's Catalyst",
		oldB006EstridOracle:   "Estrid, the Masked",
		oldB006TezzeretOracle: "Tezzeret, Agent of Bolas",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v", name, spec, ok)
		}
	}
}

// --- Tezzeret ---------------------------------------------------------------

func TestS82OldB006BTezzeretPlusOneTakesAnArtifactAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tez := effectstest.S82Walker(g, me.ID, "Tezzeret, Agent of Bolas", oldB006TezzeretOracle, 3)
	bolt := effectstest.PutOnTopForTest(me, "Filler", "Instant")
	relic := effectstest.PutOnTopForTest(me, "Relic", "Artifact")
	deep := effectstest.PutOnTopForTest(me, "Deep Filler", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, tez, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, relic)
	effectstest.AnswerAllPrompts(t, g, me.ID)
	if !me.Hand.Contains(relic) || me.Hand.Contains(deep) || me.Hand.Contains(bolt) {
		t.Errorf("hand has Relic=%v Deep=%v Filler=%v, want only the artifact", me.Hand.Contains(relic), me.Hand.Contains(deep), me.Hand.Contains(bolt))
	}
	if me.Library.Contains(deep) && me.Library.Cards[0].InstanceID == deep {
		t.Error("the rest stays on top, want the bottom")
	}
	if got := effectstest.LoyaltyOf(g, tez); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestS82OldB006BTezzeretMinusOneMakesAFiveFiveArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tez := effectstest.S82Walker(g, me.ID, "Tezzeret, Agent of Bolas", oldB006TezzeretOracle, 3)
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tez, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(land)}); err == nil {
		t.Fatal("a Forest is not an artifact: the target must be refused")
	}
	effectstest.S82Activate(t, g, me.ID, tez, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(relic)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, relic)
	if !c.IsCreature() || !c.IsArtifact() || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("Relic is creature=%v artifact=%v %d/%d, want a 5/5 artifact creature", c.IsCreature(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%4)
	c, _ = effectstest.BattlefieldCard(g, relic)
	if !c.IsCreature() {
		t.Error("no duration is printed: it stays a creature on later turns")
	}
}

func TestS82OldB006BTezzeretMinusFourDrainsTwiceTheArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	tez := effectstest.S82Walker(g, me.ID, "Tezzeret, Agent of Bolas", oldB006TezzeretOracle, 4)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Relic", "Artifact", "", false) // not mine
	effectstest.AdvanceToMain(t, g)
	mine, theirs := me.Life, opp.Life
	effectstest.S82Activate(t, g, me.ID, tez, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-6 || me.Life != mine+6 {
		t.Errorf("life me %d->%d, opponent %d->%d, want +6 and -6 (two per artifact I control)", mine, me.Life, theirs, opp.Life)
	}
}

// --- Chandra, Flame's Catalyst -----------------------------------------------

func TestS82OldB006BChandraPlusOneHitsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ch := effectstest.S82Walker(g, me.ID, "Chandra, Flame's Catalyst", oldB006ChandraOracle, 5)
	effectstest.AdvanceToMain(t, g)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.S82Activate(t, g, me.ID, ch, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want -= 3
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, want)
		}
	}
	if got := effectstest.LoyaltyOf(g, ch); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func oldB006BGraveSpell(me *game.Player, name, typeLine string, colors ...string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{R}", Colors: colors,
		Owner: me.ID, Controller: me.ID})
	return id
}

func TestS82OldB006BChandraMinusTwoOnlyTargetsRedInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ch := effectstest.S82Walker(g, me.ID, "Chandra, Flame's Catalyst", oldB006ChandraOracle, 5)
	blue := oldB006BGraveSpell(me, "Blue Instant", "Instant", "U")
	red := oldB006BGraveSpell(me, "Red Creature", "Creature — Goblin", "R")
	effectstest.AdvanceToMain(t, g)
	for _, bad := range []uuid.UUID{blue, red} {
		if err := g.ActivateCatalogAbility(me.ID, ch, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bad)}); err == nil {
			t.Fatal("a card that is not a red instant or sorcery was accepted as a target")
		}
	}
}

func TestS82OldB006BChandraMinusTwoCastsTheSpellAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ch := effectstest.S82Walker(g, me.ID, "Chandra, Flame's Catalyst", oldB006ChandraOracle, 5)
	spell := oldB006BGraveSpell(me, "Red Sorcery", "Sorcery", "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ch, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(spell)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, spell)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err == nil {
		t.Fatal("the cast is at its mana cost: no mana was available")
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err != nil {
		t.Fatalf("cast for {R}: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(spell) || !effectstest.InExile(g, spell) {
		t.Error("the spell is exiled instead of going to the graveyard")
	}
	if got := effectstest.LoyaltyOf(g, ch); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestS82OldB006BChandraMinusTwoDeclinedLeavesTheCardInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ch := effectstest.S82Walker(g, me.ID, "Chandra, Flame's Catalyst", oldB006ChandraOracle, 5)
	spell := oldB006BGraveSpell(me, "Red Instant", "Instant", "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ch, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(spell)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, spell)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(spell) {
		t.Error("a declined card stays in the graveyard")
	}
}

func TestS82OldB006BChandraUltimateDiscardsDrawsSevenAndCastsFromHandForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ch := effectstest.S82Walker(g, me.ID, "Chandra, Flame's Catalyst", oldB006ChandraOracle, 8)
	old := effectstest.HandCardFull(me, "Old Card", "Sorcery", "{3}", "", nil)
	for i := 0; i < 8; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Instant", "{1}")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ch, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(old) || !me.Graveyard.Contains(old) {
		t.Error("the old hand is discarded")
	}
	if me.Hand.Size() != 7 {
		t.Errorf("hand has %d cards, want 7", me.Hand.Size())
	}
	var spell, gy uuid.UUID
	for _, c := range me.Hand.Cards {
		spell = c.InstanceID
		break
	}
	gy = effectstest.GyCard(me, "Graveyard Card", "Sorcery", "{3}", "")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, spell, game.ZoneHand), game.GrantedAltCostNextFree) == nil {
		t.Error("a spell in hand is offered the free cast")
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, gy, game.ZoneGraveyard), game.GrantedAltCostNextFree) != nil {
		t.Error("only spells cast from the hand are free: a graveyard cast is not")
	}
	// It is never spent: the next spell is offered the same.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostNextFree}); err != nil {
		t.Fatalf("free cast from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range me.Hand.Cards {
		if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, c.InstanceID, game.ZoneHand), game.GrantedAltCostNextFree) == nil {
			t.Error("the promise is spent by one cast, want it to last the turn")
		}
		break
	}
}

// --- Estrid, the Masked --------------------------------------------------------

func oldB006BAuraOn(g *game.Game, me *game.Player, host uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Pacifism", TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
	return id
}

func TestS82OldB006BEstridPlusTwoUntapsEnchantedPermanentsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	est := effectstest.S82Walker(g, me.ID, "Estrid, the Masked", oldB006EstridOracle, 3)
	enchanted := effectstest.PushCatalogPermanent(g, me.ID, "Enchanted Bear", "Creature — Bear", "", false)
	plain := effectstest.PushCatalogPermanent(g, me.ID, "Plain Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	oldB006BAuraOn(g, me, enchanted)
	oldB006BAuraOn(g, opp, theirs)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case enchanted, plain, theirs:
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, est, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for id, wantTapped := range map[uuid.UUID]bool{enchanted: false, plain: true, theirs: true} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, id); tapped != wantTapped {
			t.Errorf("tapped=%v for %s, want %v", tapped, id, wantTapped)
		}
	}
	if got := effectstest.LoyaltyOf(g, est); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB006BEstridMinusOneMakesAMaskThatSavesTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	est := effectstest.S82Walker(g, me.ID, "Estrid, the Masked", oldB006EstridOracle, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, est, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(est)}); err == nil {
		t.Fatal("Estrid can't target herself: it says another target permanent")
	}
	effectstest.S82Activate(t, g, me.ID, est, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	var mask *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Mask" {
			mask = c
		}
	}
	if mask == nil {
		t.Fatal("no Mask token")
	}
	if !mask.IsAura() || !mask.IsAttachedTo(bear) || len(mask.Colors) != 1 || mask.Colors[0] != "W" {
		t.Errorf("Mask: aura=%v attached=%v colors=%v, want a white Aura on the Bear", mask.IsAura(), mask.IsAttachedTo(bear), mask.Colors)
	}
	// Umbra armor: destroying the Bear destroys the Mask instead.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the Bear should have survived through umbra armor")
	}
	if effectstest.CountBattlefieldByName(g, "Mask") != 0 {
		t.Error("the Mask is destroyed in its place")
	}
}

func TestS82OldB006BEstridMinusSevenMillsThenReturnsEnchantmentsThenAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	est := effectstest.S82Walker(g, me.ID, "Estrid, the Masked", oldB006EstridOracle, 7)
	ench := effectstest.GyCard(me, "Oblivion Ring", "Enchantment", "{2}{W}", "")
	creature := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	aura := effectstest.GyCard(me, "Pacifism", "Enchantment — Aura", "{1}{W}", "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// A milled enchantment returns too.
	milled := effectstest.SeedLibraryTop(me, "Milled Enchantment", "Enchantment", "{1}")
	for i := 0; i < 6; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Land", "")
	}
	// SeedLibraryTop pushes on top: put the enchantment among the seven.
	_ = milled
	effectstest.AdvanceToMain(t, g)
	libBefore := me.Library.Size()
	effectstest.S82Activate(t, g, me.ID, est, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		effectstest.AnswerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := libBefore - me.Library.Size(); got != 7 {
		t.Errorf("library fell by %d, want 7", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, ench); !ok {
		t.Error("the non-Aura enchantment did not return")
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("a creature card must stay in the graveyard")
	}
	if _, ok := effectstest.BattlefieldCard(g, aura); !ok {
		t.Error("the Aura did not return")
	} else if a, _ := effectstest.BattlefieldCard(g, aura); a.AttachedTo.ID == uuid.Nil {
		t.Error("the returned Aura is attached to nothing")
	}
}
