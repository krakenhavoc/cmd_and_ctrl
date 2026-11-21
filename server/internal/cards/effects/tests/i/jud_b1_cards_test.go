package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jud_b1_cards_test.go — Judgment, batch 1 (#5611). Each test plays the
// real catalog entry and asserts what its text says.

const (
	judB1DefyGravityOracle    = "5a295dc1-7f60-4581-b0f9-61e8b14339bf"
	judB1CanopyClawsOracle    = "195c6453-e5cc-4b3b-8518-eaccf29bc4a6"
	judB1FolkMedicineOracle   = "135eb851-3f2d-45c5-bfb9-49d1e6c98344"
	judB1RayOfRevelationOracl = "4620bfa8-7198-44ae-a162-51a19ca0f555"
	judB1KrosanReclamationOrc = "e69252a6-8c27-4af9-8cbe-11eb416a6a46"
	judB1GrizzlyFateOracle    = "f3073e4f-ca7e-4fde-88a4-bd7b6ebfbc8c"
	judB1BarkripperOracle     = "2cdd85c8-fa60-44ca-bc96-d38f029d2184"
	judB1BattlewiseAvenOracle = "f165e482-5b87-4062-87ac-bbe8303f74c9"
	judB1ToxicStenchOracle    = "9c3be33b-0be9-49ec-8160-898cbcab8814"
	judB1EmberShotOracle      = "1f300320-ad98-4812-b34b-54885a86aee9"
	judB1SuddenStrengthOracle = "8df51012-b334-475f-b70d-af4f2b3840a0"
	judB1MentalNoteOracle     = "e8d5f31c-7abf-4fbb-977e-8353a97daf7a"
	judB1EnvelopOracle        = "30062dd0-c049-4872-8011-8b4810a3fa26"
	judB1ChastiseOracle       = "b7553f3f-5de1-409c-a184-e12e40f017ab"
)

func TestJudB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		judB1DefyGravityOracle:    "Defy Gravity",
		judB1CanopyClawsOracle:    "Canopy Claws",
		judB1FolkMedicineOracle:   "Folk Medicine",
		judB1RayOfRevelationOracl: "Ray of Revelation",
		judB1KrosanReclamationOrc: "Krosan Reclamation",
		judB1GrizzlyFateOracle:    "Grizzly Fate",
		judB1BarkripperOracle:     "Anurid Barkripper",
		judB1BattlewiseAvenOracle: "Battlewise Aven",
		judB1ToxicStenchOracle:    "Toxic Stench",
		judB1EmberShotOracle:      "Ember Shot",
		judB1SuddenStrengthOracle: "Sudden Strength",
		judB1MentalNoteOracle:     "Mental Note",
		judB1EnvelopOracle:        "Envelop",
		judB1ChastiseOracle:       "Chastise",
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

// --- flashback ---------------------------------------------------------------

func TestJudB1FlashbackCardsOfferTheirPrintedPriceFromTheGraveyard(t *testing.T) {
	for _, c := range []struct{ name, oracle, cost string }{
		{"Defy Gravity", judB1DefyGravityOracle, "{U}"},
		{"Canopy Claws", judB1CanopyClawsOracle, "{G}"},
		{"Folk Medicine", judB1FolkMedicineOracle, "{1}{W}"},
		{"Ray of Revelation", judB1RayOfRevelationOracl, "{G}"},
		{"Krosan Reclamation", judB1KrosanReclamationOrc, "{1}{G}"},
		{"Grizzly Fate", judB1GrizzlyFateOracle, "{5}{G}{G}"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not open the graveyard", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "flashback" {
				t.Fatalf("%s graveyard offers = %+v, want one flashback", c.name, offers)
			}
			if offers[0].ManaCost != c.cost || offers[0].Life != 0 || !offers[0].ExileOnLeavingStack {
				t.Errorf("%s flashback pays %q and %d life (exile %v), want %q, no life, and exile",
					c.name, offers[0].ManaCost, offers[0].Life, offers[0].ExileOnLeavingStack, c.cost)
			}
			if got := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneHand); len(got) != 0 {
				t.Errorf("%s offers %+v from hand", c.name, got)
			}
		})
	}
}

func TestJudB1DefyGravityGrantsFlyingAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	ogre := effectstest.BlbB1Push(g, me.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Defy Gravity", "Instant", judB1DefyGravityOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("the target did not gain flying")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, ogre, "flying") {
		t.Fatal("a creature that was not targeted gained flying")
	}

	defy := effectstest.SeedGraveyardCard(t, g, "Defy Gravity", "Instant", judB1DefyGravityOracle)
	effectstest.OdyB1Flashback(t, g, defy, "{U}", effectstest.CardTarget(ogre)...)
	if !effectstest.HasKeywordOnBattlefield(t, g, ogre, "flying") {
		t.Error("the flashed-back Defy Gravity did not grant flying")
	}
}

// Battlewise Aven prints flying, so it is the flier the Claws are aimed at.
func TestJudB1CanopyClawsTakesFlyingAwayAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	aven := effectstest.BlbB1Push(g, opp.ID, "Battlewise Aven", "Creature — Bird Soldier", judB1BattlewiseAvenOracle, 2, 2)
	other := effectstest.BlbB1Push(g, opp.ID, "Battlewise Aven", "Creature — Bird Soldier", judB1BattlewiseAvenOracle, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Canopy Claws", "Instant", judB1CanopyClawsOracle, effectstest.CardTarget(aven))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, aven, "flying") {
		t.Fatal("the target kept flying")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, other, "flying") {
		t.Fatal("a creature that was not targeted lost flying")
	}

	claws := effectstest.SeedGraveyardCard(t, g, "Canopy Claws", "Instant", judB1CanopyClawsOracle)
	effectstest.OdyB1Flashback(t, g, claws, "{G}", effectstest.CardTarget(other)...)
	if effectstest.HasKeywordOnBattlefield(t, g, other, "flying") {
		t.Error("the flashed-back Canopy Claws did not take flying away")
	}
}

func TestJudB1FolkMedicineGainsOnePerCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	for i := 0; i < 3; i++ {
		effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	}
	effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.BlbB1Push(g, me.ID, "Relic", "Artifact", "", 0, 0)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Folk Medicine", "Instant", judB1FolkMedicineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Fatalf("gained %d life with three creatures, want 3", got)
	}

	effectstest.BlbB1Push(g, me.ID, "Fourth Bear", "Creature — Bear", "", 2, 2)
	life = me.Life
	medicine := effectstest.SeedGraveyardCard(t, g, "Folk Medicine", "Instant", judB1FolkMedicineOracle)
	effectstest.OdyB1Flashback(t, g, medicine, "{1}{W}")
	if got := me.Life - life; got != 4 {
		t.Errorf("the flashback gained %d life with four creatures, want 4", got)
	}
}

func TestJudB1RayOfRevelationDestroysAnEnchantmentAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	pact := effectstest.BlbB1Push(g, opp.ID, "Pact", "Enchantment", "", 0, 0)
	ward := effectstest.BlbB1Push(g, opp.ID, "Ward", "Enchantment", "", 0, 0)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Ray of Revelation", "Instant", judB1RayOfRevelationOracl, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the target enchantment")
	}
	effectstest.CastCatalogSpell(t, g, "Ray of Revelation", "Instant", judB1RayOfRevelationOracl, effectstest.CardTarget(pact))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pact) {
		t.Fatal("the enchantment was not destroyed")
	}

	ray := effectstest.SeedGraveyardCard(t, g, "Ray of Revelation", "Instant", judB1RayOfRevelationOracl)
	effectstest.OdyB1Flashback(t, g, ray, "{G}", effectstest.CardTarget(ward)...)
	if g.Battlefield.Contains(ward) {
		t.Error("the flashed-back Ray of Revelation did not destroy the enchantment")
	}
}

func TestJudB1KrosanReclamationShufflesUpToTwoCardsAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	var theirs []uuid.UUID
	for _, n := range []string{"T1", "T2", "T3"} {
		theirs = append(theirs, effectstest.PushGraveyardCardTyped(opp, n, "Creature — Bear"))
	}
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Krosan Reclamation", "Instant", judB1KrosanReclamationOrc, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: theirs[0], Slot: 1}, {Kind: game.TargetCard, ID: theirs[1], Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib+2 || opp.Graveyard.Contains(theirs[0]) || opp.Graveyard.Contains(theirs[1]) || !opp.Graveyard.Contains(theirs[2]) {
		t.Fatalf("library %d (want %d); the two picked cards should be in it and the third still in the graveyard", opp.Library.Size(), lib+2)
	}

	mine := effectstest.PushGraveyardCardTyped(me, "Mine", "Creature — Bear")
	myLib := me.Library.Size()
	reclamation := effectstest.SeedGraveyardCard(t, g, "Krosan Reclamation", "Instant", judB1KrosanReclamationOrc)
	effectstest.OdyB1Flashback(t, g, reclamation, "{1}{G}",
		game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 1})
	if me.Graveyard.Contains(mine) || me.Library.Size() != myLib+1 {
		t.Error("the flashed-back Krosan Reclamation did not shuffle the card into its owner's library")
	}
}

func TestJudB1KrosanReclamationCardsMustBeInTheTargetPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, other := effectstest.BlbB1Seats(g)
	elsewhere := effectstest.PushGraveyardCardTyped(other, "Elsewhere", "Creature — Bear")
	if err := effectstest.CastCatalogSpellErr(t, g, "Krosan Reclamation", "Instant", judB1KrosanReclamationOrc, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: elsewhere, Slot: 1},
	}); err == nil {
		t.Error("a card from another player's graveyard was accepted")
	}
}

func TestJudB1GrizzlyFateMakesTwoBearsOrFourAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.CastCatalogSpell(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Bear"); got != 2 {
		t.Fatalf("%d Bear tokens without threshold, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Bear" {
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 2 || tough != 2 || !c.HasColor("G") {
				t.Errorf("a Bear is %d/%d green=%v, want a green 2/2", p, tough, c.HasColor("G"))
			}
		}
	}

	// Grizzly Fate itself is in the graveyard now: six more is seven.
	effectstest.BlbMPFillGraveyard(g, me, 6)
	fate := effectstest.SeedGraveyardCard(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle)
	effectstest.OdyB1Flashback(t, g, fate, "{5}{G}{G}")
	if got := effectstest.B43TokensNamed(g, me.ID, "Bear"); got != 6 {
		t.Errorf("%d Bear tokens after the flashback at threshold, want 2 + 4", got)
	}
}

// The spell on the stack is not in the graveyard: six other cards and the
// flashed-back Fate itself are not threshold.
func TestJudB1GrizzlyFateDoesNotCountItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbMPFillGraveyard(g, me, 6)
	fate := effectstest.SeedGraveyardCard(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle)
	effectstest.OdyB1Flashback(t, g, fate, "{5}{G}{G}")
	if got := effectstest.B43TokensNamed(g, me.ID, "Bear"); got != 2 {
		t.Errorf("%d Bear tokens, want 2", got)
	}
}

// --- threshold ---------------------------------------------------------------

func TestJudB1ThresholdCreaturesGrowAtSevenCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, printed, gained string
		power, toughness, bigP, bigT            int
	}{
		{"Anurid Barkripper", "Creature — Frog Beast", judB1BarkripperOracle, "", "", 2, 2, 4, 4},
		{"Battlewise Aven", "Creature — Bird Soldier", judB1BattlewiseAvenOracle, "flying", "first strike", 2, 2, 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, _ := effectstest.BlbB1Seats(g)
			id := effectstest.BlbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if tc.printed != "" && !effectstest.HasKeywordOnBattlefield(t, g, id, tc.printed) {
				t.Fatalf("%s does not have %s", tc.name, tc.printed)
			}
			// Seven cards in an OPPONENT's graveyard are not threshold.
			effectstest.BlbMPFillGraveyard(g, opp, 7)
			effectstest.BlbMPFillGraveyard(g, me, 6)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Fatalf("six cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			if tc.gained != "" && effectstest.HasKeywordOnBattlefield(t, g, id, tc.gained) {
				t.Fatalf("%s before threshold", tc.gained)
			}
			effectstest.BlbMPFillGraveyard(g, me, 1)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.bigP || tough != tc.bigT {
				t.Errorf("seven cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.bigP, tc.bigT)
			}
			if tc.gained != "" && !effectstest.HasKeywordOnBattlefield(t, g, id, tc.gained) {
				t.Errorf("no %s at threshold", tc.gained)
			}
		})
	}
}

func TestJudB1ToxicStenchShrinksOrAtThresholdDestroys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	ogre := effectstest.BlbB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Toxic Stench", "Instant", judB1ToxicStenchOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 1 {
		t.Fatalf("the 2/2 is %d/%d without threshold, want 1/1", p, tough)
	}

	// Toxic Stench is in the graveyard now: six more is seven.
	effectstest.BlbMPFillGraveyard(g, me, 6)
	effectstest.CastCatalogSpell(t, g, "Toxic Stench", "Instant", judB1ToxicStenchOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 3/3 survived Toxic Stench at threshold")
	}
}

func TestJudB1ToxicStenchRefusesABlackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	zombie := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Toxic Stench", "Instant", judB1ToxicStenchOracle, effectstest.CardTarget(zombie)); err == nil {
		t.Error("a black creature was accepted as the target")
	}
}

// --- the rest ----------------------------------------------------------------

func TestJudB1EmberShotDealsThreeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Ember Shot", "Instant", judB1EmberShotOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("Ember Shot dealt %d, want 3", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after Ember Shot, want %d", got, hand+1)
	}
}

func TestJudB1SuddenStrengthPumpsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sudden Strength", "Instant", judB1SuddenStrengthOracle, effectstest.CardTarget(bear))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the 2/2 is %d/%d, want 5/5", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after Sudden Strength, want %d", got, hand+1)
	}
}

func TestJudB1MentalNoteMillsTwoThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	id := effectstest.CastCatalogSpell(t, g, "Mental Note", "Instant", judB1MentalNoteOracle, nil)
	hand, grave, lib := me.Hand.Size(), me.Graveyard.Size(), me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after Mental Note, want %d", got, hand+1)
	}
	if got := me.Library.Size(); got != lib-3 {
		t.Errorf("library %d after Mental Note, want %d (two milled, one drawn)", got, lib-3)
	}
	// The two milled cards and Mental Note itself.
	if got := me.Graveyard.Size(); got != grave+3 || !me.Graveyard.Contains(id) {
		t.Errorf("graveyard %d after Mental Note, want %d with the spell in it", got, grave+3)
	}
}

func TestJudB1EnvelopCountersOnlyASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	strength := effectstest.CastCatalogSpell(t, g, "Sudden Strength", "Instant", judB1SuddenStrengthOracle, effectstest.CardTarget(bear))
	envelop := effectstest.PushCatalogHandCard(opp, "Envelop", "Instant", judB1EnvelopOracle)
	if err := g.CastSpell(opp.ID, envelop, game.CastSpellParams{Targets: effectstest.CardTarget(strength)}); err == nil {
		t.Fatal("an instant was accepted as the target sorcery spell")
	}
	effectstest.PassPriorityAroundTable(t, g)

	fate := effectstest.CastCatalogSpell(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle, nil)
	if err := g.CastSpell(opp.ID, envelop, game.CastSpellParams{Targets: effectstest.CardTarget(fate)}); err != nil {
		t.Fatalf("cast Envelop at a sorcery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Bear"); got != 0 {
		t.Errorf("the countered Grizzly Fate made %d Bears", got)
	}
	if !me.Graveyard.Contains(fate) {
		t.Error("the countered sorcery is not in its owner's graveyard")
	}
}

func TestJudB1ChastiseDestroysAnAttackerAndGainsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	attacker := effectstest.BlbB1Push(g, me.ID, "Attacking Ogre", "Creature — Ogre", "", 3, 3)
	idle := effectstest.BlbB1Push(g, me.ID, "Idle Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)

	chastise := effectstest.PushCatalogHandCard(opp, "Chastise", "Instant", judB1ChastiseOracle)
	if err := g.CastSpell(opp.ID, chastise, game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	life := opp.Life
	if err := g.CastSpell(opp.ID, chastise, game.CastSpellParams{Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("cast Chastise: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature was not destroyed")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the other creature left the battlefield")
	}
	if got := opp.Life - life; got != 3 {
		t.Errorf("the caster gained %d life, want the attacker's power, 3", got)
	}
}
