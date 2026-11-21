package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b97_cards_test.go — Commander Masters artifacts, GAP batch 97
// (#5948). Each test plays the real catalog entry through the engine.

const (
	gapAssaultSuitOracle      = "c3b6e114-7bc8-4890-97d9-edbeea1f425f"
	gapCampfireOracle         = "011ff7a5-6bce-4784-a19b-e74f7b49b8f1"
	gapDarksteelMonolithOracl = "fdd3482d-7bce-411b-80c4-956a12fc143d"
	gapForebearsBladeOracle   = "f9af02f3-6ab1-469c-9d37-e8f6fa9608c4"
	gapHeartPiercerBowOracle  = "ebe9dad7-d74d-4fa8-b17a-88ff79209379"
	gapJeweledLotusOracle     = "23216e59-f147-4bb7-a698-12777c2f8584"
	gapPillarOfOriginsOracle  = "9e62f30d-3130-416e-b459-48b5f71f8be9"
)

func TestGapB97EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapAssaultSuitOracle:      "Assault Suit",
		gapCampfireOracle:         "Campfire",
		gapDarksteelMonolithOracl: "Darksteel Monolith",
		gapForebearsBladeOracle:   "Forebear's Blade",
		gapHeartPiercerBowOracle:  "Heart-Piercer Bow",
		gapJeweledLotusOracle:     "Jeweled Lotus",
		gapPillarOfOriginsOracle:  "Pillar of Origins",
	}
	for oracle, name := range want {
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

// --- Pillar of Origins ---

func TestGapB97PillarOfOriginsPaysOnlyForCreaturesOfTheChosenType(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	pillar := effectstest.PushNamedTribePermanent(t, g, me.ID, "Pillar of Origins", "Artifact", gapPillarOfOriginsOracle, "Elf")
	if err := g.ActivateManaAbility(me.ID, pillar, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the Pillar: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool %d, want one mana", me.ManaPool.Total())
	}
	bear := effectstest.HandSpell(me, "Grizzly Bear", "Creature — Bear", "{G}")
	artifact := effectstest.HandSpell(me, "Plain Rock", "Artifact", "{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the mana paid for a creature of another type")
	}
	if err := g.CastSpell(me.ID, artifact, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the mana paid for a noncreature spell")
	}
	elf := effectstest.HandSpell(me, "Llanowar Scout", "Creature — Elf", "{G}")
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a creature of the chosen type: %v", err)
	}
}

// --- Jeweled Lotus ---

func TestGapB97JeweledLotusMakesThreeManaForYourCommanderOnly(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	lotus := effectstest.PushCatalogPermanent(g, me.ID, "Jeweled Lotus", "Artifact", gapJeweledLotusOracle, false)
	if err := g.ActivateManaAbility(me.ID, lotus, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap the Lotus: %v", err)
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool %d, want three mana", me.ManaPool.Total())
	}
	if _, ok := effectstest.BattlefieldCard(g, lotus); ok {
		t.Error("the Lotus was not sacrificed")
	}
	cost, _ := game.ParseCost("{2}{R}")
	other := game.Card{InstanceID: uuid.New(), Name: "Not A Commander", TypeLine: "Creature — Goblin", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(other)) {
		t.Error("the mana pays for a spell that is not a commander")
	}
	theirs := game.Card{InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Goblin", ManaCost: "{2}{R}", IsCommander: true, Owner: opp.ID, Controller: me.ID}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(theirs)) {
		t.Error("the mana pays for a commander you don't own")
	}
	cmdr := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cmdr, Name: "Kiki", TypeLine: "Legendary Creature — Goblin", ManaCost: "{2}{R}",
		IsCommander: true, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, cmdr, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("your commander with the Lotus's mana: %v", err)
	}
}

// --- Campfire ---

func TestGapB97CampfireGainsTwoLife(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	fire := effectstest.PushCatalogPermanent(g, me.ID, "Campfire", "Artifact", gapCampfireOracle, false)
	effectstest.B06AddMana(me, "G")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, fire, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestGapB97CampfireReturnsCommandersAndShufflesTheGraveyard(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	fire := effectstest.PushCatalogPermanent(g, me.ID, "Campfire", "Artifact", gapCampfireOracle, false)
	inCommand := uuid.New()
	inYard := uuid.New()
	theirs := uuid.New()
	filler := uuid.New()
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{InstanceID: inCommand, Name: "Cmdr One", TypeLine: "Legendary Creature — Elf",
			Owner: me.ID, Controller: me.ID, IsCommander: true})
		me.Graveyard.PushTop(game.Card{InstanceID: inYard, Name: "Cmdr Two", TypeLine: "Legendary Creature — Elf",
			Owner: me.ID, Controller: me.ID, IsCommander: true})
		me.Graveyard.PushTop(game.Card{InstanceID: filler, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, fire, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Each commander headed for a hand gets the CR 903.9b offer; declining
	// is what the card is for.
	for offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionalReplacement); offer != nil; offer = effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionalReplacement) {
		if err := g.ResolveOptionalReplacement(offer.ID, me.ID, false); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
	}
	if !me.Hand.Contains(inCommand) || !me.Hand.Contains(inYard) {
		t.Errorf("commanders in hand: command-zone one %v, graveyard one %v", me.Hand.Contains(inCommand), me.Hand.Contains(inYard))
	}
	if me.Graveyard.Size() != 0 || !me.Library.Contains(filler) {
		t.Errorf("graveyard %d, filler in library %v; want the graveyard shuffled in", me.Graveyard.Size(), me.Library.Contains(filler))
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("another player's graveyard was shuffled")
	}
	if g.Battlefield.Contains(fire) {
		t.Error("Campfire was not exiled as a cost")
	}
}

// --- Darksteel Monolith ---

func TestGapB97DarksteelMonolithMakesOneColorlessSpellFromHandFreeEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMainForTest(t, g)
	effectstest.PushPermanentForTest(g, me.ID, "Darksteel Monolith", gapDarksteelMonolithOracl, "Artifact")
	key := "granted-zero-colorless-spell-from-hand"
	rock := effectstest.HandSpell(me, "Colorless Rock", "Artifact", "{6}")
	red := me.Hand.Cards
	_ = red
	coloured := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: coloured, Name: "Red Rock", TypeLine: "Artifact", ManaCost: "{4}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, coloured, game.ZoneHand), key) != nil {
		t.Error("offered a free cast of a coloured spell")
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, rock, game.ZoneHand), key) == nil {
		t.Fatal("no free cast offered for a colorless spell in hand")
	}
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	again := effectstest.HandSpell(me, "Another Rock", "Artifact", "{5}")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, again, game.ZoneHand), key) != nil {
		t.Error("a second free cast in one turn")
	}
}

func TestGapB97DarksteelMonolithIsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushPermanentForTest(g, me.ID, "Darksteel Monolith", gapDarksteelMonolithOracl, "Artifact")
	effectstest.AssertKeywords(t, g, id, "indestructible")
}

// --- Forebear's Blade ---

func TestGapB97ForebearsBladePumpsAndMovesOnWhenItsCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	blade := effectstest.PushCatalogPermanent(g, me.ID, "Forebear's Blade", "Artifact — Equipment", gapForebearsBladeOracle, false)
	first := effectstest.B12Creature(g, me.ID, "First Bear", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "G", "G", "G")
	effectstest.EquipTo(t, g, me.ID, blade, first)
	if p, tough := effectstest.EffectivePTOf(t, g, first); p != 5 || tough != 2 {
		t.Errorf("equipped P/T %d/%d, want 5/2", p, tough)
	}
	effectstest.AssertKeywords(t, g, first, "vigilance", "trample")
	_ = g.DestroyPermanentForEffect(first)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, second)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != second {
		t.Errorf("the Blade is attached to %v, want the second Bear", host.ID)
	}
}

// --- Heart-Piercer Bow ---

func TestGapB97HeartPiercerBowPingsACreatureTheDefenderControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	bow := effectstest.PushCatalogPermanent(g, me.ID, "Heart-Piercer Bow", "Artifact — Equipment", gapHeartPiercerBowOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	bystander := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	effectstest.EquipTo(t, g, me.ID, bow, bear)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, bystander) || effectstest.HasID(p.PickTargetCards, bear) {
			t.Errorf("offered a creature the defending player does not control: %v", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, theirs); c == nil || c.DamageMarked != 1 {
		t.Errorf("the defender's creature: %+v, want 1 damage", c)
	}
	if c := effectstest.FindBattlefieldCardForTest(g, bystander); c == nil || c.DamageMarked != 0 {
		t.Errorf("a bystander took damage: %+v", c)
	}
}

// --- Assault Suit ---

func TestGapB97AssaultSuitPumpsAndLendsItsCreatureToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	suit := effectstest.PushCatalogPermanent(g, me.ID, "Assault Suit", "Artifact — Equipment", gapAssaultSuitOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.B06AddMana(me, "G", "G", "G")
	effectstest.EquipTo(t, g, me.ID, suit, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped P/T %d/%d, want 4/4", p, tough)
	}
	effectstest.AssertKeywords(t, g, bear, "haste")
	if !effectstest.LayeredCard(t, g, bear).Effective().CantBeSacrificed {
		t.Error("the equipped creature can be sacrificed")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, bear)
	if c.Controller != opp.ID {
		t.Fatalf("controller %v, want the opponent whose upkeep it is", c.Controller)
	}
	if c.Tapped {
		t.Error("the borrowed creature was not untapped")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Error("the equipped creature attacked its Equipment's controller")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Errorf("attacking another player: %v", err)
	}
}
