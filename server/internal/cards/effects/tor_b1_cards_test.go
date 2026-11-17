package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tor_b1_cards_test.go — Torment, batch 1 (#5612). Each test plays the
// real catalog entry and asserts what its text says.

const (
	torB1ChainersEdictOracle = "7975cb9d-d37e-44f8-9614-86d2ad453c35"
	torB1CripplingOracle     = "d87190d0-bda3-4ad9-84b2-019f751999ce"
	torB1AcornHarvestOracle  = "6aeba615-2d37-4eb8-a377-41a517e95aa3"
	torB1ObsessiveOracle     = "116ee88e-9e13-4f21-bc6a-efa994b5bf75"
	torB1FranticPurifOracle  = "b4025136-3146-499a-8c17-ed9de548ab8d"
	torB1PsychoticHazeOracle = "4b6aa5e2-b539-46cd-804b-cd9e996a9456"
	torB1ArrogantWurmOracle  = "dc7f0c33-80ea-453e-a237-473f077f038e"
	torB1RootwallaOracle     = "b2ec080f-ad5d-4274-91e3-bc6ac89fc3e4"
	torB1SetonsScoutOracle   = "7b2a32ca-ef5b-40a5-acf7-aefac1d09e5b"
	torB1MysticFamiliarOracl = "9d3931a2-baba-4e6e-b923-44df00900e51"
	torB1TerohsFaithfulOracl = "875c96ce-f6e4-438b-929d-de53afdd99a1"
	torB1AccelerateOracle    = "79189464-9645-4411-8886-68fd40a588ed"
	torB1NantukoShadeOracle  = "40d02b8c-fb0e-4607-855f-49fbcec2e209"
	torB1UnhingeOracle       = "cb1ae4f2-4fec-4cac-8992-616af82e0ade"
)

func TestTorB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		torB1ChainersEdictOracle: "Chainer's Edict",
		torB1CripplingOracle:     "Crippling Fatigue",
		torB1AcornHarvestOracle:  "Acorn Harvest",
		torB1ObsessiveOracle:     "Obsessive Search",
		torB1FranticPurifOracle:  "Frantic Purification",
		torB1PsychoticHazeOracle: "Psychotic Haze",
		torB1ArrogantWurmOracle:  "Arrogant Wurm",
		torB1RootwallaOracle:     "Basking Rootwalla",
		torB1SetonsScoutOracle:   "Seton's Scout",
		torB1MysticFamiliarOracl: "Mystic Familiar",
		torB1TerohsFaithfulOracl: "Teroh's Faithful",
		torB1AccelerateOracle:    "Accelerate",
		torB1NantukoShadeOracle:  "Nantuko Shade",
		torB1UnhingeOracle:       "Unhinge",
	} {
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

// --- flashback ---------------------------------------------------------------

func TestTorB1FlashbackCardsOfferTheirPrintedPriceFromTheGraveyard(t *testing.T) {
	for _, c := range []struct {
		name, oracle, cost, label string
		life                      int
	}{
		{"Chainer's Edict", torB1ChainersEdictOracle, "{5}{B}{B}", "Flashback {5}{B}{B}", 0},
		{"Crippling Fatigue", torB1CripplingOracle, "{1}{B}", "Flashback—{1}{B}, Pay 3 life", 3},
		{"Acorn Harvest", torB1AcornHarvestOracle, "{1}{G}", "Flashback—{1}{G}, Pay 3 life", 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not open the graveyard", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "flashback" {
				t.Fatalf("%s graveyard offers = %+v, want one flashback", c.name, offers)
			}
			o := offers[0]
			if o.ManaCost != c.cost || o.Life != c.life || o.Label != c.label || !o.ExileOnLeavingStack {
				t.Errorf("%s flashback = %q, %d life, label %q, exile %v; want %q, %d life, %q and exile",
					c.name, o.ManaCost, o.Life, o.Label, o.ExileOnLeavingStack, c.cost, c.life, c.label)
			}
			if got := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneHand); len(got) != 0 {
				t.Errorf("%s offers %+v from hand", c.name, got)
			}
		})
	}
}

func TestTorB1ChainersEdictMakesTheTargetSacrificeAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	a := blbB1Push(g, opp.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := blbB1Push(g, opp.ID, "Bear B", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Chainer's Edict", "Sorcery", torB1ChainersEdictOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the target player was not asked to sacrifice a creature")
	}
	answerSacrifice(t, g, opp.ID, a)
	if g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("the chosen creature was not the one sacrificed")
	}

	edict := seedGraveyardCard(t, g, "Chainer's Edict", "Sorcery", torB1ChainersEdictOracle)
	c := blbB1Push(g, opp.ID, "Bear C", "Creature — Bear", "", 2, 2)
	odyB1CastFlashback(t, g, edict, "{5}{B}{B}", playerTarget(opp.ID))
	answerSacrifice(t, g, opp.ID, b)
	g.SettleResolution()
	if g.Battlefield.Contains(b) || !g.Battlefield.Contains(c) {
		t.Error("the flashed-back Chainer's Edict did not make the target sacrifice the chosen creature")
	}
	if !g.Exile.Contains(edict) {
		t.Error("the flashed-back Chainer's Edict was not exiled")
	}
}

func TestTorB1CripplingFatigueShrinksAndFlashesBackForManaAndLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	ogre := blbB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	castCatalogSpell(t, g, "Crippling Fatigue", "Sorcery", torB1CripplingOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, ogre); p != 1 || tough != 1 {
		t.Fatalf("the 3/3 is %d/%d after Crippling Fatigue, want 1/1", p, tough)
	}

	life := me.Life
	fatigue := seedGraveyardCard(t, g, "Crippling Fatigue", "Sorcery", torB1CripplingOracle)
	odyB1Flashback(t, g, fatigue, "{1}{B}", cardTarget(bear)...)
	if got := life - me.Life; got != 3 {
		t.Errorf("the flashback cost %d life, want 3", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the flashed-back Crippling Fatigue did not kill the 2/2")
	}
}

// The life is a cost: a caster below 3 life can't claim the flashback.
func TestTorB1PayLifeFlashbackNeedsTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	harvest := seedGraveyardCard(t, g, "Acorn Harvest", "Sorcery", torB1AcornHarvestOracle)
	g.WithWriteLock(func() { me.Life = 2 })
	dcMana(t, g, me.ID, "{C}{G}")
	if err := g.CastSpell(me.ID, harvest, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true,
	}); err == nil {
		t.Fatal("flashback was cast at 2 life")
	}
	if !me.Graveyard.Contains(harvest) || me.Life != 2 {
		t.Error("a refused flashback moved the card or took the life")
	}
}

func TestTorB1AcornHarvestMakesTwoSquirrelsFromHandAndGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Acorn Harvest", "Sorcery", torB1AcornHarvestOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Squirrel"); got != 2 {
		t.Fatalf("%d Squirrel tokens after the hand cast, want 2", got)
	}
	life := me.Life
	harvest := seedGraveyardCard(t, g, "Acorn Harvest", "Sorcery", torB1AcornHarvestOracle)
	odyB1Flashback(t, g, harvest, "{1}{G}")
	if got := b43TokensNamed(g, me.ID, "Squirrel"); got != 4 {
		t.Fatalf("%d Squirrel tokens after the flashback, want 4", got)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("the flashback cost %d life, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Squirrel" {
			if p, tough := ptOf(t, g, c.InstanceID); p != 1 || tough != 1 || !c.HasColor("G") {
				t.Errorf("a Squirrel is %d/%d green=%v, want a green 1/1", p, tough, c.HasColor("G"))
			}
		}
	}
}

// --- madness -----------------------------------------------------------------

func TestTorB1MadnessCardsGrowBothHalves(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Obsessive Search", torB1ObsessiveOracle, "{U}"},
		{"Frantic Purification", torB1FranticPurifOracle, "{W}"},
		{"Psychotic Haze", torB1PsychoticHazeOracle, "{1}{B}"},
		{"Arrogant Wurm", torB1ArrogantWurmOracle, "{2}{G}"},
		{"Basking Rootwalla", torB1RootwallaOracle, "{0}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := Lookup(tc.oracle)
			if spec.Madness != tc.cost {
				t.Errorf("madness cost %q, want %q", spec.Madness, tc.cost)
			}
			reps, trigs := madnessAbilitiesFor(tc.oracle)
			if reps != 1 || trigs != 1 {
				t.Errorf("madness halves: %d replacements, %d triggers; want 1 and 1 (CR 702.35a)", reps, trigs)
			}
		})
	}
}

// torB1LootIntoExile puts the card in an emptied hand beside a spare,
// casts Faithless Looting and discards both, and checks the madness
// card went to exile with a cast offered.
func torB1LootIntoExile(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := pushCatalogHandCard(me, name, typeLine, oracle)
	spare := pushCatalogHandCard(me, "Spare", "Sorcery", "")
	castCatalogSpell(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracl, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, card, spare)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(card) || me.Graveyard.Contains(card) {
		t.Fatalf("the discarded %s is not in exile (CR 702.35a)", name)
	}
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatalf("the madness trigger offered no cast of %s", name)
	}
	return card
}

func TestTorB1ObsessiveSearchIsCastForUAfterADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	search := torB1LootIntoExile(t, g, "Obsessive Search", "Instant", torB1ObsessiveOracle)
	hand := me.Hand.Size()
	dcMana(t, g, me.ID, "{U}")
	if err := g.CastSpell(me.ID, search, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness,
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want {U} to be the whole price", n)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after Obsessive Search, want %d", got, hand+1)
	}
	if !me.Graveyard.Contains(search) {
		t.Error("the resolved Obsessive Search is not in its owner's graveyard")
	}
}

// Madness {0} is a real free cast: no mana at all.
func TestTorB1BaskingRootwallaIsCastForNothingAfterADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walla := torB1LootIntoExile(t, g, "Basking Rootwalla", "Creature — Lizard", torB1RootwallaOracle)
	if err := g.CastSpell(me.ID, walla, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness,
	}); err != nil {
		t.Fatalf("the madness cast for {0}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(walla) {
		t.Error("the madness-cast Rootwalla is not on the battlefield")
	}
}

func TestTorB1BaskingRootwallaPumpsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walla := blbB1Push(g, me.ID, "Basking Rootwalla", "Creature — Lizard", torB1RootwallaOracle, 1, 1)
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}{C}")
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, walla); p != 3 || tough != 3 {
		t.Fatalf("after one pump %d/%d, want 3/3", p, tough)
	}
	floatMana(t, g, me, "{G}{C}")
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a second activation in one turn was accepted")
	}
}

func TestTorB1FranticPurificationDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	aura := blbB1Push(g, opp.ID, "Pact", "Enchantment", "", 0, 0)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErr(t, g, "Frantic Purification", "Instant", torB1FranticPurifOracle, cardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the target enchantment")
	}
	castCatalogSpell(t, g, "Frantic Purification", "Instant", torB1FranticPurifOracle, cardTarget(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment was not destroyed")
	}
}

func TestTorB1PsychoticHazeHitsEachCreatureAndEachPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "My Elf", "Creature — Elf", "", 1, 1)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Psychotic Haze", "Instant", torB1PsychoticHazeOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := lives[p.ID] - p.Life; got != 1 {
			t.Errorf("a player lost %d life, want 1", got)
		}
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the caster's own 1/1 survived")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 died to 1 damage")
	}
}

func TestTorB1ArrogantWurmHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	wurm := blbB1Push(g, g.Seats[0].ID, "Arrogant Wurm", "Creature — Wurm", torB1ArrogantWurmOracle, 4, 4)
	if !hasKeywordOnBattlefield(t, g, wurm, "trample") {
		t.Error("Arrogant Wurm does not have trample")
	}
}

// --- threshold ---------------------------------------------------------------

func TestTorB1ThresholdCreaturesGrowAtSevenCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, printed, gained string
		power, toughness, bigP, bigT            int
	}{
		{"Seton's Scout", "Creature — Centaur Druid Scout Archer", torB1SetonsScoutOracle, "reach", "", 2, 1, 4, 3},
		{"Mystic Familiar", "Creature — Bird", torB1MysticFamiliarOracl, "flying", "protection from black", 1, 2, 2, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			id := blbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if !hasKeywordOnBattlefield(t, g, id, tc.printed) {
				t.Fatalf("%s does not have %s", tc.name, tc.printed)
			}
			// Seven cards in an OPPONENT's graveyard are not threshold.
			blbMPFillGraveyard(g, opp, 7)
			blbMPFillGraveyard(g, me, 6)
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Fatalf("six cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			if tc.gained != "" && hasKeywordOnBattlefield(t, g, id, tc.gained) {
				t.Fatalf("%s before threshold", tc.gained)
			}
			blbMPFillGraveyard(g, me, 1)
			if p, tough := ptOf(t, g, id); p != tc.bigP || tough != tc.bigT {
				t.Errorf("seven cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.bigP, tc.bigT)
			}
			if tc.gained != "" && !hasKeywordOnBattlefield(t, g, id, tc.gained) {
				t.Errorf("no %s at threshold", tc.gained)
			}
		})
	}
}

// --- the rest ----------------------------------------------------------------

func TestTorB1TerohsFaithfulGainsFourOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Teroh's Faithful", "Creature — Human Cleric", torB1TerohsFaithfulOracl, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 4 {
		t.Errorf("gained %d life, want 4", got)
	}
}

func TestTorB1AccelerateGrantsHasteAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Accelerate", "Instant", torB1AccelerateOracle, cardTarget(bear))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after Accelerate, want %d", got, hand+1)
	}
}

func TestTorB1NantukoShadePumpsEachActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shade := blbB1Push(g, me.ID, "Nantuko Shade", "Creature — Insect Shade", torB1NantukoShadeOracle, 2, 1)
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		floatMana(t, g, me, "{B}")
		if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tough := ptOf(t, g, shade); p != 4 || tough != 3 {
		t.Errorf("after two pumps %d/%d, want 4/3", p, tough)
	}
}

func TestTorB1UnhingeMakesTheTargetDiscardThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	victim := pushCatalogHandCard(opp, "Victim", "Instant", "")
	castCatalogSpell(t, g, "Unhinge", "Sorcery", torB1UnhingeOracle, []game.TargetRef{playerTarget(opp.ID)})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, opp.ID); got != 1 {
		t.Fatalf("the target owes %d discards, want 1", got)
	}
	answerDiscard(t, g, opp.ID, victim)
	g.SettleResolution()
	if !opp.Graveyard.Contains(victim) {
		t.Fatal("the chosen card was not discarded")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("caster's hand %d after Unhinge, want %d", got, hand+1)
	}
}
