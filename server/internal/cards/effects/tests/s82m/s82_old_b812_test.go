package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b812_test.go — Mirage, batch 812 (#5633): counters, phasing, recursion.

const (
	oldB812Martyr      = "122dfe9f-e0e2-4f69-9637-289c956fd54a"
	oldB812Unyaro      = "27d80a75-688e-4f68-9622-54484dd3bd91"
	oldB812Boon        = "6d87415b-0618-48e1-9510-320c4611ecda"
	oldB812SeedCharm   = "b8d66cd6-15db-41ff-bae3-60f1aad3ea23"
	oldB812Illumin     = "cd1a6508-aaba-4515-9496-057915d02040"
	oldB812Taniwha     = "371c993a-69b8-4a51-af89-982b79940b2d"
	oldB812Wurm        = "c54c1517-cf7c-4afb-92f8-70c9d1591b45"
	oldB812Isle        = "ce55657d-d82f-4528-a83e-5cad7de111fd"
	oldB812Imp         = "e4826967-b2f9-4f32-8411-d243edf41082"
	oldB812Harvest     = "0cc38d20-5ae8-413a-9f26-ff13cd92f24c"
	oldB812Mtenda      = "273f0353-4936-491b-82da-6d4d7f3259a6"
	oldB812Hakim       = "43e99cba-a6e3-4abc-a92b-aaf5839c72d2"
	oldB812Naturalize  = "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728"
	oldB812BearType    = "Creature — Bear"
	oldB812GriffinType = "Creature — Griffin"
)

func TestOldB812EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB812Martyr: "Vigilant Martyr", oldB812Unyaro: "Unyaro Griffin", oldB812Boon: "Withering Boon",
		oldB812SeedCharm: "Seedling Charm", oldB812Illumin: "Illumination", oldB812Taniwha: "Taniwha",
		oldB812Wurm: "Warping Wurm", oldB812Isle: "Teferi's Isle", oldB812Imp: "Teferi's Imp",
		oldB812Harvest: "Bone Harvest", oldB812Mtenda: "Mtenda Griffin", oldB812Hakim: "Hakim, Loreweaver",
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

func oldB812Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB812Activate(g *game.Game, who, src uuid.UUID, idx int, p game.ActivateAbilityParams) error {
	p.Strict = who == g.Seats[g.Turn.ActiveSeat].ID
	return g.ActivateCatalogAbility(who, src, idx, p)
}

// oldB812Cast casts a card from `from`'s hand (strict), leaving it on the stack.
func oldB812Cast(g *game.Game, from *game.Player, c game.Card, p game.CastSpellParams) (uuid.UUID, error) {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = from.ID, from.ID
	from.Hand.PushTop(c)
	// Timing is not what these tests are about (several spells sit on the
	// stack at once), so casts are permissive; costs and targets are still
	// enforced.
	p.Strict = false
	return c.InstanceID, g.CastSpell(from.ID, c.InstanceID, p)
}

func oldB812Countered(g *game.Game, spell uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && (ev.CardID == spell || ev.StackItemID == spell) {
			return true
		}
	}
	return false
}

func TestOldB812VigilantMartyrRegeneratesAndCountersASpellTargetingAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB812Seats(g)
	martyr := effectstest.B12Push(g, me.ID, "Vigilant Martyr", "Creature — Human Cleric", oldB812Martyr, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB812BearType, 2, 2)
	if err := oldB812Activate(g, me.ID, martyr, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(martyr) {
		t.Error("the Martyr was not sacrificed")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 5) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the Bear")
	}

	// The counter: legal only against a spell with an enchantment among its targets.
	martyr = effectstest.B12Push(g, me.ID, "Vigilant Martyr", "Creature — Human Cleric", oldB812Martyr, 1, 1)
	ench := effectstest.B12Permanent(g, opp.ID, "Test Enchantment", "Enchantment")
	rock := effectstest.B12Permanent(g, opp.ID, "Test Rock", "Artifact")
	onRock := effectstest.Batch01OpponentCasts(t, g, opp, "Naturalize", oldB812Naturalize, "", effectstest.CardTarget(rock))
	effectstest.FloatMana(t, g, me, "{W}{W}")
	if err := oldB812Activate(g, me.ID, martyr, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(onRock)}); err == nil {
		t.Error("countered a spell that targets only an artifact")
	}
	onEnch := effectstest.Batch01OpponentCasts(t, g, opp, "Naturalize", oldB812Naturalize, "", effectstest.CardTarget(ench))
	if err := oldB812Activate(g, me.ID, martyr, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(onEnch)}); err != nil {
		t.Fatalf("counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB812Countered(g, onEnch) {
		t.Error("the spell targeting an enchantment was not countered")
	}
	if g.Battlefield.Contains(martyr) {
		t.Error("the Martyr was not sacrificed as a cost")
	}
}

func TestOldB812UnyaroGriffinCountersOnlyARedInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp, me := oldB812Seats(g) // the active seat casts; the other seat answers
	griffin := effectstest.B12Push(g, me.ID, "Unyaro Griffin", oldB812GriffinType, oldB812Unyaro, 2, 2)
	creature, err := oldB812Cast(g, opp, game.Card{Name: "Red Creature", TypeLine: oldB812BearType, Colors: []string{"R"}}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := oldB812Activate(g, me.ID, griffin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(creature)}); err == nil {
		t.Error("countered a red creature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	red, err := oldB812Cast(g, opp, game.Card{Name: "Red Spell", TypeLine: "Sorcery", Colors: []string{"R"}}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	green, err := oldB812Cast(g, opp, game.Card{Name: "Green Spell", TypeLine: "Instant", Colors: []string{"G"}}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := oldB812Activate(g, me.ID, griffin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(green)}); err == nil {
		t.Error("countered a green spell")
	}
	if err := oldB812Activate(g, me.ID, griffin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(red)}); err != nil {
		t.Fatalf("counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB812Countered(g, red) || g.Battlefield.Contains(griffin) {
		t.Errorf("countered %v, Griffin on battlefield %v: want countered and sacrificed", oldB812Countered(g, red), g.Battlefield.Contains(griffin))
	}
}

func TestOldB812WitheringBoonCostsThreeLifeAndCountersACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp, me := oldB812Seats(g) // the active seat casts; the other seat answers
	creature, err := oldB812Cast(g, opp, game.Card{Name: "Their Bear", TypeLine: oldB812BearType}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	other, err := oldB812Cast(g, opp, game.Card{Name: "Their Instant", TypeLine: "Instant"}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	me.Life = 2
	if _, err := oldB812Cast(g, me, game.Card{Name: "Withering Boon", TypeLine: "Instant", OracleID: oldB812Boon}, game.CastSpellParams{Targets: effectstest.CardTarget(creature)}); err == nil {
		t.Error("cast Withering Boon at 2 life")
	}
	me.Life = 20
	if _, err := oldB812Cast(g, me, game.Card{Name: "Withering Boon", TypeLine: "Instant", OracleID: oldB812Boon}, game.CastSpellParams{Targets: effectstest.CardTarget(other)}); err == nil {
		t.Error("targeted a noncreature spell")
	}
	if _, err := oldB812Cast(g, me, game.Card{Name: "Withering Boon", TypeLine: "Instant", OracleID: oldB812Boon}, game.CastSpellParams{Targets: effectstest.CardTarget(creature)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != 17 {
		t.Errorf("life %d, want 17 after paying 3", me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB812Countered(g, creature) {
		t.Error("the creature spell was not countered")
	}
}

func TestOldB812IlluminationCountersAnArtifactSpellAndItsControllerGainsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp, me := oldB812Seats(g) // the active seat casts; the other seat answers
	creature, err := oldB812Cast(g, opp, game.Card{Name: "Their Bear", TypeLine: oldB812BearType}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if _, err := oldB812Cast(g, me, game.Card{Name: "Illumination", TypeLine: "Instant", OracleID: oldB812Illumin}, game.CastSpellParams{Targets: effectstest.CardTarget(creature)}); err == nil {
		t.Error("targeted a creature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, opp, "{C}{C}{C}")
	rock, err := oldB812Cast(g, opp, game.Card{Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{3}"}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	before := opp.Life
	effectstest.FloatMana(t, g, me, "{W}{W}")
	if _, err := oldB812Cast(g, me, game.Card{Name: "Illumination", TypeLine: "Instant", ManaCost: "{W}{W}", OracleID: oldB812Illumin}, game.CastSpellParams{Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB812Countered(g, rock) {
		t.Error("the artifact spell was not countered")
	}
	if opp.Life != before+3 {
		t.Errorf("life %d→%d, want +3 (its mana value)", before, opp.Life)
	}
}

func TestOldB812SeedlingCharmEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB812Seats(g)
	host := effectstest.B12Creature(g, opp.ID, "Host", oldB812BearType, 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
	landAura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Land Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: land}})
	charm := func(mode int, target uuid.UUID) error {
		_, err := oldB812Cast(g, me, game.Card{Name: "Seedling Charm", TypeLine: "Instant", OracleID: oldB812SeedCharm},
			game.CastSpellParams{Modes: []int{mode}, Targets: effectstest.CardTarget(target)})
		return err
	}
	// Mode 0: only an Aura attached to a creature.
	if err := charm(0, landAura); err == nil {
		t.Error("returned an Aura attached to a land")
	}
	if err := charm(0, aura); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !me.Hand.Contains(aura) {
		t.Error("the Aura was not returned to its owner's hand")
	}
	// Mode 1: only a green creature.
	green := effectstest.B16Creature(g, me.ID, "Green Bear", oldB812BearType, 2, 2, "G")
	if err := charm(1, host); err == nil {
		t.Error("regenerated a colorless creature")
	}
	if err := charm(1, green); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), green, 5) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Error("the green creature did not regenerate")
	}
	// Mode 2: trample until end of turn.
	if err := charm(2, host); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, host, "trample") {
		t.Error("the creature did not gain trample")
	}
}

func TestOldB812TaniwhaPhasesOutYourLandsAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB812Seats(g)
	taniwha := effectstest.B12Push(g, me.ID, "Taniwha", "Legendary Creature — Serpent", oldB812Taniwha, 7, 7)
	mine := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	// Taniwha's own phasing takes it out at my first untap step, where its upkeep
	// trigger cannot fire; it is back for the next upkeep.
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB812SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPhasedOutForEffect(taniwha) || g.IsPhasedOutForEffect(mine) {
		t.Fatalf("first upkeep: Taniwha phased out %v, land phased out %v; want true, false", g.IsPhasedOutForEffect(taniwha), g.IsPhasedOutForEffect(mine))
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB812SeatIndex(g, me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPhasedOutForEffect(taniwha) {
		t.Fatal("Taniwha did not phase back in")
	}
	if !g.IsPhasedOutForEffect(mine) {
		t.Error("my land did not phase out")
	}
	if g.IsPhasedOutForEffect(theirs) {
		t.Error("an opponent's land phased out")
	}
}

func oldB812SeatIndex(g *game.Game, id uuid.UUID) int {
	for i, s := range g.Seats {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func TestOldB812WarpingWurmPhasesOutUnlessPaidAndGrowsWhenItPhasesIn(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB812Seats(g)
		wurm := effectstest.B12Push(g, me.ID, "Warping Wurm", "Creature — Wurm", oldB812Wurm, 1, 1)
		// Its own phasing takes it out at my first untap step and back in at the
		// second, before the upkeep where it asks for {2}{G}{U}.
		effectstest.AdvanceToUpkeepOfSeat(t, g, oldB812SeatIndex(g, me.ID))
		effectstest.PassPriorityAroundTable(t, g)
		if !g.IsPhasedOutForEffect(wurm) {
			t.Fatal("the Wurm did not phase out at its controller's untap step")
		}
		effectstest.AdvanceToUpkeepOfSeat(t, g, oldB812SeatIndex(g, me.ID))
		// Its phase-in trigger and its upkeep trigger go on the stack together.
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if g.IsPhasedOutForEffect(wurm) {
			t.Fatal("the Wurm did not phase back in")
		}
		if n := effectstest.CountersOf(t, g, wurm); n != 1 {
			t.Errorf("+1/+1 counters %d, want 1 after phasing in", n)
		}
		if effectstest.HasPayUnlessFor(g, me.ID) == false {
			t.Fatal("no pay-or-phase-out offer at the upkeep")
		}
		if pay {
			effectstest.FloatMana(t, g, me, "{C}{C}{G}{U}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.IsPhasedOutForEffect(wurm) == pay {
			t.Errorf("paid %v: phased out %v, want phased out only when unpaid", pay, g.IsPhasedOutForEffect(wurm))
		}
	}
}

func TestOldB812TeferisIsleEntersTappedAndTapsForUU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB812Seats(g)
	id := effectstest.PlayLandFromHand(t, g, "Teferi's Isle", oldB812Isle)
	if !effectstest.TappedNow(g, id) {
		t.Fatal("Teferi's Isle did not enter tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, game.KeywordPhasing) {
		t.Error("Teferi's Isle does not have phasing")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a tapped Isle made mana")
	}
	effectstest.Untap(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("pool %v, want {U}{U}", got)
	}
}

func TestOldB812TeferisImpDiscardsWhenItPhasesOutAndDrawsWhenItPhasesIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB812Seats(g)
	imp := effectstest.B12Push(g, me.ID, "Teferi's Imp", "Creature — Imp", oldB812Imp, 1, 1)
	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, imp) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard choice on phasing out")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand-1 {
		t.Errorf("hand %d, want %d after the discard", len(me.Hand.Cards), hand-1)
	}
	g.WithWriteLock(func() { _ = g.PhaseInForEffect(imp) })
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d, want %d after the draw", len(me.Hand.Cards), hand)
	}
}

func TestOldB812BoneHarvestPutsCreatureCardsOnTopAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB812Seats(g)
	a, b := uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: a, Name: "Dead Bear A", TypeLine: oldB812BearType, Owner: me.ID, Controller: me.ID},
		{InstanceID: b, Name: "Dead Bear B", TypeLine: oldB812BearType, Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Dead Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID},
	} {
		me.Graveyard.PushTop(c)
	}
	sorc := me.Graveyard.Cards[2].InstanceID
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bone Harvest", TypeLine: "Instant", OracleID: oldB812Harvest, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, sorc), Strict: true}); err == nil {
		t.Fatal("targeted a sorcery card")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked to order the cards")
	}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("order: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := len(me.Library.Cards)
	top2 := map[uuid.UUID]bool{me.Library.Cards[n-1].InstanceID: true, me.Library.Cards[n-2].InstanceID: true}
	if !top2[a] || !top2[b] {
		t.Error("the two creature cards are not the top two cards of the library")
	}
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) || !me.Graveyard.Contains(sorc) {
		t.Error("the graveyard does not hold exactly the sorcery card")
	}
	hand := len(me.Hand.Cards)
	effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d: the delayed draw did not happen", len(me.Hand.Cards), hand+1)
	}
}

func TestOldB812MtendaGriffinReturnsItselfAndAGriffinOnlyInYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB812Seats(g)
	mtenda := effectstest.B12Push(g, me.ID, "Mtenda Griffin", oldB812GriffinType, oldB812Mtenda, 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Griffin", TypeLine: oldB812GriffinType, Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: oldB812BearType, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{W}")
	if err := oldB812Activate(g, me.ID, mtenda, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(dead)}); err == nil {
		t.Fatal("activated outside the upkeep")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB812SeatIndex(g, me.ID))
	effectstest.Untap(g, mtenda)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := oldB812Activate(g, me.ID, mtenda, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("targeted a non-Griffin card")
	}
	if err := oldB812Activate(g, me.ID, mtenda, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(dead)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mtenda) || !me.Hand.Contains(mtenda) {
		t.Error("the Griffin did not return to its owner's hand")
	}
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the Griffin card did not return from the graveyard to the hand")
	}
}

func TestOldB812HakimReturnsAnAuraAttachedToItselfAndDestroysThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB812Seats(g)
	hakim := effectstest.B12Push(g, me.ID, "Hakim, Loreweaver", "Legendary Creature — Human Wizard", oldB812Hakim, 2, 4)
	auraCard := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: auraCard, Name: "Dead Aura", TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{U}{U}")
	if err := oldB812Activate(g, me.ID, hakim, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(auraCard)}); err == nil {
		t.Fatal("activated outside the upkeep")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, oldB812SeatIndex(g, me.ID))
	effectstest.FloatMana(t, g, me, "{U}{U}")
	if err := oldB812Activate(g, me.ID, hakim, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(auraCard)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, auraCard)
	if !ok {
		t.Fatal("the Aura did not enter the battlefield")
	}
	if card.AttachedTo.ID != hakim {
		t.Errorf("the Aura is attached to %s, want Hakim", card.AttachedTo.ID)
	}
	// Hakim is now enchanted: the first ability can't be activated again.
	other := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Another Aura", TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{U}{U}")
	if err := oldB812Activate(g, me.ID, hakim, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(other)}); err == nil {
		t.Error("activated while Hakim was enchanted")
	}
	// The second ability destroys every Aura attached to Hakim.
	effectstest.FloatMana(t, g, me, "{U}{U}")
	if err := oldB812Activate(g, me.ID, hakim, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(auraCard) {
		t.Error("the Aura attached to Hakim was not destroyed")
	}
	if !g.Battlefield.Contains(hakim) {
		t.Error("Hakim was destroyed")
	}
}
