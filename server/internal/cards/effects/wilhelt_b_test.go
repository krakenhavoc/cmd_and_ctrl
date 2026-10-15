package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wilhelt_b_test.go — the Wilhelt, the Rotcleaver deck request (#3209),
// slice B: the cards that needed a closer look at an existing seam.

const (
	wrLilianaDGOracle      = "5e958212-6a5b-4288-8d31-f1572619d7dc"
	wrNecrodualityOracle   = "d36acdfe-2b1b-4c19-8b81-0b3127e480fe"
	wrNoosegrafOracle      = "f27fb53f-a983-410a-821b-e48cd8c01f2e"
	wrGeralfOracle         = "a62e6cf4-1adb-4c28-b24b-91287bb23e96"
	wrRiseOfDreadMarnOrcl  = "0998768a-1632-4da1-8310-352582501c4b"
	wrCurseRestlessOracle  = "0b5f7a1b-fbb9-4659-8fcb-7822dc0263bc"
	wrEmbalmersToolsOracle = "5bb4ee32-11c6-4ed4-a071-c3cf48070fff"
	wrReflectionsOracle    = "c3fdfb94-2d10-4743-864c-a59fdd57d8b7"
	wrGhoulsNightOutOracle = "cc9f9f14-a11c-4e33-9e11-8bf6ef13cf2b"
	wrUndeadAlchemistOrcl  = "3807e6fe-0555-4b09-aef2-efd58abaf669"
)

func TestWilheltSliceBIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		wrLilianaDGOracle:           "Liliana, Dreadhorde General",
		wrNecrodualityOracle:        "Necroduality",
		wrNoosegrafOracle:           "Noosegraf Mob",
		wrGeralfOracle:              "Geralf, the Fleshwright",
		wrRiseOfDreadMarnOrcl:       "Rise of the Dread Marn",
		wrCurseRestlessOracle:       "Curse of the Restless Dead",
		wrEmbalmersToolsOracle:      "Embalmer's Tools",
		wrReflectionsOracle:         "Reflections of Littjara",
		wrGhoulsNightOutOracle:      "Ghouls' Night Out",
		wrUndeadAlchemistOrcl:       "Undead Alchemist",
		poppetStitcherOracle:        "Poppet Stitcher",
		poppetStitcherOracle + "#1": "Poppet Factory",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestLilianaDreadhordeGeneralDrawsOnDeathsAndMakesZombies(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	lili := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 6)
	advanceToMainOf(t, g, seat)
	b16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	zombie := findBattlefieldByName(g, "Zombie")
	if zombie == uuid.Nil {
		t.Fatal("+1 made no Zombie")
	}
	hand := me.Hand.Size()
	wrDestroy(t, g, zombie)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a creature you control died: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestLilianaDreadhordeGeneralMinusFourEachPlayerSacrificesTwo(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	lili := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 6)
	advanceToMainOf(t, g, seat)
	mine := []uuid.UUID{pushVanillaCreature(g, me.ID, "M1", 1, 1), pushVanillaCreature(g, me.ID, "M2", 1, 1), pushVanillaCreature(g, me.ID, "M3", 1, 1)}
	theirs := []uuid.UUID{pushVanillaCreature(g, opp.ID, "T1", 1, 1), pushVanillaCreature(g, opp.ID, "T2", 1, 1)}

	b16Activate(t, g, me.ID, lili, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if prompt := latestChooseCardsFor(g, me.ID); prompt == nil || prompt.ChooseMin != 2 || prompt.ChooseMax != 2 {
		t.Fatalf("controller's choice = %+v, want exactly two of three", prompt)
	}
	answerChooseCards(t, g, me.ID, mine[0], mine[1])
	if g.Battlefield.Contains(mine[0]) || g.Battlefield.Contains(mine[1]) {
		t.Fatal("the controller's chosen creatures survived after all required choices were collected")
	}
	if prompt := latestChooseCardsFor(g, opp.ID); prompt != nil {
		t.Fatalf("opponent with exactly two creatures should have a forced set, got prompt %+v", prompt)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine[0]) || g.Battlefield.Contains(mine[1]) || !g.Battlefield.Contains(mine[2]) {
		t.Error("the controller did not sacrifice exactly the two chosen")
	}
	for _, id := range theirs {
		if g.Battlefield.Contains(id) {
			t.Error("the opponent kept a creature with only two")
		}
	}
}

func TestLilianaDreadhordeGeneralKeepRule(t *testing.T) {
	artCreature := game.Card{TypeLine: "Artifact Creature — Golem"}
	creature := game.Card{TypeLine: "Creature — Bear"}
	land := game.Card{TypeLine: "Basic Land — Swamp"}
	ench := game.Card{TypeLine: "Enchantment"}
	present := permanentTypesAmong([]game.Card{artCreature, creature, land, ench})
	valid := lilianaKeepsOneOfEachType(present)
	for _, tc := range []struct {
		name string
		keep []game.Card
		want bool
	}{
		{"the artifact creature for both, a land and an enchantment", []game.Card{artCreature, land, ench}, true},
		{"the artifact creature and the bear", []game.Card{artCreature, creature, land, ench}, true},
		{"no enchantment kept", []game.Card{artCreature, land}, false},
		{"two creatures for one type", []game.Card{creature, creature, artCreature, land, ench}, false},
		{"the bear but no artifact", []game.Card{creature, land, ench}, false},
	} {
		if got := valid(tc.keep); got != tc.want {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestLilianaDreadhordeGeneralMinusNineSacrificesTheRest(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	lili := pushWalkerForTest(g, me.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 9)
	advanceToMainOf(t, g, seat)
	keepBear := pushVanillaCreature(g, opp.ID, "Keep Bear", 2, 2)
	loseBear := pushVanillaCreature(g, opp.ID, "Lose Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)

	b16Activate(t, g, me.ID, lili, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, keepBear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(keepBear) || g.Battlefield.Contains(loseBear) {
		t.Error("the opponent did not keep exactly the chosen creature")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the controller's own permanents were touched")
	}
}

func TestNecrodualityCopiesANontokenZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Necroduality", "Enchantment", wrNecrodualityOracle, false)
	castCatalogSpell(t, g, "Gisa and Geralf", "Legendary Creature — Human Wizard", wrGisaAndGeralfOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Gisa and Geralf"); got != 1 {
		t.Fatalf("a non-Zombie was copied: %d", got)
	}
	castCatalogSpell(t, g, "Walker", "Creature — Zombie", "", nil)
	passPriorityAroundTable(t, g)
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Walker" && c.IsToken() {
			tokens++
		}
	}
	if tokens != 1 || b29CountNamed(g, me.ID, "Walker") != 2 {
		t.Errorf("a nontoken Zombie entered: %d token copies, want 1 (and no copy of the copy)", tokens)
	}
}

func TestNoosegrafMobTradesCountersForZombiesOnAnySpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Noosegraf Mob", "Creature — Zombie", wrNoosegrafOracle, nil)
	passPriorityAroundTable(t, g)
	mob := findBattlefieldByName(g, "Noosegraf Mob")
	if got := counterCount(g, mob, game.CounterPlusOne); got != 5 {
		t.Fatalf("%d counters, want 5", got)
	}
	castCatalogSpell(t, g, "Village Rites Stand-in", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mob, game.CounterPlusOne); got != 4 {
		t.Errorf("%d counters after a spell, want 4", got)
	}
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("%d Zombies, want 1", got)
	}
}

func TestGeralfMakesARogueFromTheSecondSpellAndGrowsZombies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Geralf, the Fleshwright", "Legendary Creature — Human Warlock", wrGeralfOracle, false)
	castCatalogSpell(t, g, "First", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Zombie Rogue"); got != 0 {
		t.Fatalf("the first spell made %d Rogues", got)
	}
	castCatalogSpell(t, g, "Second", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	rogue := findBattlefieldByName(g, "Zombie Rogue")
	if rogue == uuid.Nil {
		t.Fatal("the second spell made no Rogue")
	}
	// The Rogue is the first Zombie in this turn: no counter.
	if got := counterCount(g, rogue, game.CounterPlusOne); got != 0 {
		t.Errorf("the first Zombie got %d counters", got)
	}
	castCatalogSpell(t, g, "Third", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	var second uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie Rogue" && c.InstanceID != rogue {
			second = c.InstanceID
		}
	}
	if second == uuid.Nil {
		t.Fatal("the third spell made no Rogue")
	}
	if got := counterCount(g, second, game.CounterPlusOne); got != 1 {
		t.Errorf("the second Zombie this turn got %d counters, want 1", got)
	}
}

func TestRiseOfTheDreadMarnCountsNontokenDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wrDestroy(t, g, pushVanillaCreature(g, me.ID, "Bear", 2, 2))
	wrDestroy(t, g, pushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Bear", 2, 2))
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 black Zombie"), 1) })
	wrDestroy(t, g, findBattlefieldByName(g, "Zombie"))
	passPriorityAroundTable(t, g)

	castCatalogSpell(t, g, "Rise of the Dread Marn", "Instant", wrRiseOfDreadMarnOrcl, nil)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Zombie Berserker"); got != 2 {
		t.Errorf("%d Berserkers, want 2 (the token's death does not count)", got)
	}
}

func TestCurseOfTheRestlessDeadMakesAZombieOnTheirLandfall(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	cursed := (seat + 1) % 4
	curse := pushCatalogPermanent(g, me.ID, "Curse of the Restless Dead", "Enchantment — Aura Curse", wrCurseRestlessOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == curse {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: g.Seats[cursed].ID}
			}
		}
	})
	b13PlayAs(t, g, cursed, "Swamp", "Basic Land — Swamp", "")
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("the enchanted player's land: %d Zombies, want 1", got)
	}
	b13PlayAs(t, g, (seat+2)%4, "Swamp", "Basic Land — Swamp", "")
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("another player's land: %d Zombies, want still 1", got)
	}
}

func TestEmbalmersToolsMillsByTappingAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tools := pushCatalogPermanent(g, me.ID, "Embalmer's Tools", "Artifact", wrEmbalmersToolsOracle, false)
	walker := wrZombie(g, me.ID, "Walker", 2, 2)
	library := them.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, tools, 0, game.ActivateAbilityParams{
		TapIDs:  []uuid.UUID{walker},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if them.Library.Size() != library-1 {
		t.Errorf("milled %d, want 1", library-them.Library.Size())
	}
	if !b16Tapped(t, g, walker) {
		t.Error("the Zombie was not tapped")
	}
}

func TestEmbalmersToolsDiscountsOnlyGraveyardCreatureAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tools := pushCatalogPermanent(g, me.ID, "Embalmer's Tools", "Artifact", wrEmbalmersToolsOracle, false)
	var src game.Card
	g.ReadSnapshot(func() { src, _ = g.LookupCardForEffect(tools) })
	pred := ofACreatureCardInYourGraveyard()
	card := game.Card{TypeLine: "Creature — Zombie", Owner: me.ID}
	if !pred(game.CostQuery{Card: card, Source: src, FromZone: game.ZoneGraveyard}) {
		t.Error("a creature card in your graveyard is not discounted")
	}
	if pred(game.CostQuery{Card: card, Source: src, FromZone: game.ZoneBattlefield}) {
		t.Error("a creature on the battlefield is discounted")
	}
	if pred(game.CostQuery{Card: game.Card{TypeLine: "Creature — Zombie", Owner: g.Seats[1].ID}, Source: src, FromZone: game.ZoneGraveyard}) {
		t.Error("a creature card in an opponent's graveyard is discounted")
	}
}

func TestReflectionsOfLittjaraCopiesSpellsOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	refl := pushCatalogPermanent(g, me.ID, "Reflections of Littjara", "Enchantment", wrReflectionsOracle, false)
	g.WithWriteLock(func() { g.SetNamedTribeForEffect(refl, "Zombie") })
	castCatalogSpell(t, g, "Walker", "Creature — Zombie", "", nil)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Walker"); got != 2 {
		t.Errorf("%d Walkers, want 2 (the card and the token copy)", got)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := b29CountNamed(g, me.ID, "Bear"); got != 1 {
		t.Errorf("%d Bears, want 1", got)
	}
}

func TestGhoulsNightOutTakesOneCreatureFromEachGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	mine := ptGraveCard(me, "My Bear", "Creature — Bear")
	theirs := ptGraveCard(opp, "Their Elf", "Creature — Elf")

	castCatalogSpell(t, g, "Ghouls' Night Out", "Sorcery", wrGhoulsNightOutOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, mine)
	answerChooseCards(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatal("a chosen card did not enter")
		}
		if c.Controller != me.ID {
			t.Error("a returned creature is not under the caster's control")
		}
		eff := c.Effective()
		if !c.HasSubtype("Zombie") || !c.HasColor("B") {
			t.Errorf("%s is not a black Zombie: %+v", c.Name, eff.Subtypes)
		}
		if !hasEffectiveKeyword(t, g, id, game.KeywordDecayed) {
			t.Errorf("%s has no decayed", c.Name)
		}
	}
}

func TestUndeadAlchemistTurnsZombieCombatDamageIntoMill(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Undead Alchemist", "Creature — Zombie", wrUndeadAlchemistOrcl, false)
	walker := wrZombie(g, me.ID, "Walker", 3, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	life, library := them.Life, them.Library.Size()
	pr7bAttack(t, g, them.ID, walker, bear)
	pr7bDamageStep(t, g)
	if them.Life != life-2 {
		t.Errorf("life %d, want %d (only the Bear's damage is dealt)", them.Life, life-2)
	}
	if them.Library.Size() != library-3 {
		t.Errorf("milled %d, want 3", library-them.Library.Size())
	}
}

func TestUndeadAlchemistExilesMilledCreaturesForZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Undead Alchemist", "Creature — Zombie", wrUndeadAlchemistOrcl, false)
	beast := uuid.New()
	them.Library.PushTop(game.Card{InstanceID: beast, Name: "Beast", TypeLine: "Creature — Beast", Owner: them.ID, Controller: them.ID})
	g.WithWriteLock(func() { _ = g.MillNForEffect(them.ID, 1) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(beast) {
		t.Error("the milled creature card was not exiled")
	}
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("%d Zombies, want 1", got)
	}
}

func TestPoppetStitcherMakesDecayedZombiesAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: poppetStitcherOracle, Layout: game.LayoutTransform,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Poppet Stitcher", TypeLine: "Creature — Human Wizard", ManaCost: "{2}{U}", Power: 2, Toughness: 3},
			{Name: "Poppet Factory", TypeLine: "Artifact"},
		},
	}
	card.SetFace(0)
	stitcher := pushBattlefieldCardWithTimestamp(g, card)
	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		castCatalogSpell(t, g, "Ponder", "Sorcery", "", nil)
		passPriorityAroundTable(t, g)
	}
	if got := b29CountNamed(g, me.ID, "Zombie"); got != 3 {
		t.Fatalf("%d Zombies, want 3", got)
	}
	b12ToMyNextUpkeep(t, g)
	prompt := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, stitcher)
	if !ok || c.ActiveFace != 1 {
		t.Fatalf("Poppet Stitcher did not transform: %+v", c.Name)
	}
	token := findBattlefieldByName(g, "Zombie")
	if p, tough := effectivePower(t, g, token), effectiveToughness(t, g, token); p != 3 || tough != 3 {
		t.Errorf("a Zombie under the Factory is %d/%d, want 3/3", p, tough)
	}
	if hasEffectiveKeyword(t, g, token, game.KeywordDecayed) {
		t.Error("the Zombie kept decayed under the Factory")
	}
}
