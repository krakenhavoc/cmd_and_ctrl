package effects

import (
	"testing"

	"github.com/google/uuid"

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
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	ogre := blbB1Push(g, me.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	castCatalogSpell(t, g, "Defy Gravity", "Instant", judB1DefyGravityOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("the target did not gain flying")
	}
	if hasKeywordOnBattlefield(t, g, ogre, "flying") {
		t.Fatal("a creature that was not targeted gained flying")
	}

	defy := seedGraveyardCard(t, g, "Defy Gravity", "Instant", judB1DefyGravityOracle)
	odyB1Flashback(t, g, defy, "{U}", cardTarget(ogre)...)
	if !hasKeywordOnBattlefield(t, g, ogre, "flying") {
		t.Error("the flashed-back Defy Gravity did not grant flying")
	}
}

// Battlewise Aven prints flying, so it is the flier the Claws are aimed at.
func TestJudB1CanopyClawsTakesFlyingAwayAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	aven := blbB1Push(g, opp.ID, "Battlewise Aven", "Creature — Bird Soldier", judB1BattlewiseAvenOracle, 2, 2)
	other := blbB1Push(g, opp.ID, "Battlewise Aven", "Creature — Bird Soldier", judB1BattlewiseAvenOracle, 2, 2)
	castCatalogSpell(t, g, "Canopy Claws", "Instant", judB1CanopyClawsOracle, cardTarget(aven))
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, aven, "flying") {
		t.Fatal("the target kept flying")
	}
	if !hasKeywordOnBattlefield(t, g, other, "flying") {
		t.Fatal("a creature that was not targeted lost flying")
	}

	claws := seedGraveyardCard(t, g, "Canopy Claws", "Instant", judB1CanopyClawsOracle)
	odyB1Flashback(t, g, claws, "{G}", cardTarget(other)...)
	if hasKeywordOnBattlefield(t, g, other, "flying") {
		t.Error("the flashed-back Canopy Claws did not take flying away")
	}
}

func TestJudB1FolkMedicineGainsOnePerCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	for i := 0; i < 3; i++ {
		blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	}
	blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, me.ID, "Relic", "Artifact", "", 0, 0)
	life := me.Life
	castCatalogSpell(t, g, "Folk Medicine", "Instant", judB1FolkMedicineOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Fatalf("gained %d life with three creatures, want 3", got)
	}

	blbB1Push(g, me.ID, "Fourth Bear", "Creature — Bear", "", 2, 2)
	life = me.Life
	medicine := seedGraveyardCard(t, g, "Folk Medicine", "Instant", judB1FolkMedicineOracle)
	odyB1Flashback(t, g, medicine, "{1}{W}")
	if got := me.Life - life; got != 4 {
		t.Errorf("the flashback gained %d life with four creatures, want 4", got)
	}
}

func TestJudB1RayOfRevelationDestroysAnEnchantmentAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	pact := blbB1Push(g, opp.ID, "Pact", "Enchantment", "", 0, 0)
	ward := blbB1Push(g, opp.ID, "Ward", "Enchantment", "", 0, 0)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErr(t, g, "Ray of Revelation", "Instant", judB1RayOfRevelationOracl, cardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the target enchantment")
	}
	castCatalogSpell(t, g, "Ray of Revelation", "Instant", judB1RayOfRevelationOracl, cardTarget(pact))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pact) {
		t.Fatal("the enchantment was not destroyed")
	}

	ray := seedGraveyardCard(t, g, "Ray of Revelation", "Instant", judB1RayOfRevelationOracl)
	odyB1Flashback(t, g, ray, "{G}", cardTarget(ward)...)
	if g.Battlefield.Contains(ward) {
		t.Error("the flashed-back Ray of Revelation did not destroy the enchantment")
	}
}

func TestJudB1KrosanReclamationShufflesUpToTwoCardsAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	var theirs []uuid.UUID
	for _, n := range []string{"T1", "T2", "T3"} {
		theirs = append(theirs, pushGraveyardCardTyped(opp, n, "Creature — Bear"))
	}
	lib := opp.Library.Size()
	castCatalogSpell(t, g, "Krosan Reclamation", "Instant", judB1KrosanReclamationOrc, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: theirs[0], Slot: 1}, {Kind: game.TargetCard, ID: theirs[1], Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != lib+2 || opp.Graveyard.Contains(theirs[0]) || opp.Graveyard.Contains(theirs[1]) || !opp.Graveyard.Contains(theirs[2]) {
		t.Fatalf("library %d (want %d); the two picked cards should be in it and the third still in the graveyard", opp.Library.Size(), lib+2)
	}

	mine := pushGraveyardCardTyped(me, "Mine", "Creature — Bear")
	myLib := me.Library.Size()
	reclamation := seedGraveyardCard(t, g, "Krosan Reclamation", "Instant", judB1KrosanReclamationOrc)
	odyB1Flashback(t, g, reclamation, "{1}{G}",
		game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 1})
	if me.Graveyard.Contains(mine) || me.Library.Size() != myLib+1 {
		t.Error("the flashed-back Krosan Reclamation did not shuffle the card into its owner's library")
	}
}

func TestJudB1KrosanReclamationCardsMustBeInTheTargetPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, other := blbB1Seats(g)
	elsewhere := pushGraveyardCardTyped(other, "Elsewhere", "Creature — Bear")
	if err := castCatalogSpellErr(t, g, "Krosan Reclamation", "Instant", judB1KrosanReclamationOrc, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: elsewhere, Slot: 1},
	}); err == nil {
		t.Error("a card from another player's graveyard was accepted")
	}
}

func TestJudB1GrizzlyFateMakesTwoBearsOrFourAtThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	castCatalogSpell(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Bear"); got != 2 {
		t.Fatalf("%d Bear tokens without threshold, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Bear" {
			if p, tough := ptOf(t, g, c.InstanceID); p != 2 || tough != 2 || !c.HasColor("G") {
				t.Errorf("a Bear is %d/%d green=%v, want a green 2/2", p, tough, c.HasColor("G"))
			}
		}
	}

	// Grizzly Fate itself is in the graveyard now: six more is seven.
	blbMPFillGraveyard(g, me, 6)
	fate := seedGraveyardCard(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle)
	odyB1Flashback(t, g, fate, "{5}{G}{G}")
	if got := b43TokensNamed(g, me.ID, "Bear"); got != 6 {
		t.Errorf("%d Bear tokens after the flashback at threshold, want 2 + 4", got)
	}
}

// The spell on the stack is not in the graveyard: six other cards and the
// flashed-back Fate itself are not threshold.
func TestJudB1GrizzlyFateDoesNotCountItself(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbMPFillGraveyard(g, me, 6)
	fate := seedGraveyardCard(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle)
	odyB1Flashback(t, g, fate, "{5}{G}{G}")
	if got := b43TokensNamed(g, me.ID, "Bear"); got != 2 {
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
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			id := blbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if tc.printed != "" && !hasKeywordOnBattlefield(t, g, id, tc.printed) {
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

func TestJudB1ToxicStenchShrinksOrAtThresholdDestroys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	ogre := blbB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	castCatalogSpell(t, g, "Toxic Stench", "Instant", judB1ToxicStenchOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 1 || tough != 1 {
		t.Fatalf("the 2/2 is %d/%d without threshold, want 1/1", p, tough)
	}

	// Toxic Stench is in the graveyard now: six more is seven.
	blbMPFillGraveyard(g, me, 6)
	castCatalogSpell(t, g, "Toxic Stench", "Instant", judB1ToxicStenchOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 3/3 survived Toxic Stench at threshold")
	}
}

func TestJudB1ToxicStenchRefusesABlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	zombie := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if err := castCatalogSpellErr(t, g, "Toxic Stench", "Instant", judB1ToxicStenchOracle, cardTarget(zombie)); err == nil {
		t.Error("a black creature was accepted as the target")
	}
}

// --- the rest ----------------------------------------------------------------

func TestJudB1EmberShotDealsThreeAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	life := opp.Life
	castCatalogSpell(t, g, "Ember Shot", "Instant", judB1EmberShotOracle, []game.TargetRef{playerTarget(opp.ID)})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("Ember Shot dealt %d, want 3", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after Ember Shot, want %d", got, hand+1)
	}
}

func TestJudB1SuddenStrengthPumpsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Sudden Strength", "Instant", judB1SuddenStrengthOracle, cardTarget(bear))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the 2/2 is %d/%d, want 5/5", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after Sudden Strength, want %d", got, hand+1)
	}
}

func TestJudB1MentalNoteMillsTwoThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	id := castCatalogSpell(t, g, "Mental Note", "Instant", judB1MentalNoteOracle, nil)
	hand, grave, lib := me.Hand.Size(), me.Graveyard.Size(), me.Library.Size()
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)

	strength := castCatalogSpell(t, g, "Sudden Strength", "Instant", judB1SuddenStrengthOracle, cardTarget(bear))
	envelop := pushCatalogHandCard(opp, "Envelop", "Instant", judB1EnvelopOracle)
	if err := g.CastSpell(opp.ID, envelop, game.CastSpellParams{Targets: cardTarget(strength)}); err == nil {
		t.Fatal("an instant was accepted as the target sorcery spell")
	}
	passPriorityAroundTable(t, g)

	fate := castCatalogSpell(t, g, "Grizzly Fate", "Sorcery", judB1GrizzlyFateOracle, nil)
	if err := g.CastSpell(opp.ID, envelop, game.CastSpellParams{Targets: cardTarget(fate)}); err != nil {
		t.Fatalf("cast Envelop at a sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Bear"); got != 0 {
		t.Errorf("the countered Grizzly Fate made %d Bears", got)
	}
	if !me.Graveyard.Contains(fate) {
		t.Error("the countered sorcery is not in its owner's graveyard")
	}
}

func TestJudB1ChastiseDestroysAnAttackerAndGainsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	attacker := blbB1Push(g, me.ID, "Attacking Ogre", "Creature — Ogre", "", 3, 3)
	idle := blbB1Push(g, me.ID, "Idle Bear", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, attacker)

	chastise := pushCatalogHandCard(opp, "Chastise", "Instant", judB1ChastiseOracle)
	if err := g.CastSpell(opp.ID, chastise, game.CastSpellParams{Targets: cardTarget(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	life := opp.Life
	if err := g.CastSpell(opp.ID, chastise, game.CastSpellParams{Targets: cardTarget(attacker)}); err != nil {
		t.Fatalf("cast Chastise: %v", err)
	}
	passPriorityAroundTable(t, g)
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
