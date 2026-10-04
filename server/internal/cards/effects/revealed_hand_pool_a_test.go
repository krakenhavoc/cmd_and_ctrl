package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0116 pool A (#2078): the plain revealed-hand pick, alone or
// followed by one independent clause. The pick itself, its reveal, its
// refusal of a bad answer and its restore are Thoughtseize's tests in
// revealed_hand_discard_test.go; these cover what each pool card adds:
// its filter, its target clause and its trailing clause.

const (
	coercionOracle             = "68413337-ddb9-46f8-8c8e-f3d2d672c652"
	darkInquiryOracle          = "fa04f161-4141-491b-ad6f-a16f69c862f3"
	despiseOracle              = "b13b40f1-1b1f-40e8-989b-ef24dc05007a"
	distressOracle             = "7bb1b48b-aa6b-4064-9355-c43b0f509388"
	divestOracle               = "9e0d7408-3827-4a65-9905-9fef3ae23c7e"
	duressOracle               = "33d405ea-7a9a-4970-b70f-9c05d90dd6f0"
	encroachOracle             = "b9959d63-5489-4ed1-a04e-9d6ba4a5f68e"
	inquisitionOfKozilekOracle = "395611fa-aec2-4c9e-92c3-70cf95cc00f4"
	layBareTheHeartOracle      = "910f4339-9125-45b7-929f-57e1adb6ac3c"
	ostracizeOracle            = "43975c07-103c-4664-b7e6-55bb29d85183"
	pilferOracle               = "d13f0907-de39-4a90-940a-831740d7aa9b"
	psychicSpearOracle         = "284b0605-4b33-499e-aff3-e8b9edcce39b"
	shatteredDreamsOracle      = "91c2f488-2b49-4499-8d5d-c1aff73a4b57"
	brainbiteOracle            = "fef94125-aa8d-4147-a609-1e990961bde2"
	harshScrutinyOracle        = "31338e24-a437-4587-acab-cec46be021e7"
	thoughtErasureOracle       = "4bae4e34-fcf4-4aee-8da2-5b3ee41a595a"
	gixsCaressOracle           = "0180bed7-892a-4af9-ac4b-96c9cd4beb42"
	tollOfTheInvasionOracle    = "3a7c2e32-8585-4f7f-8635-199e3c6cd8a9"
	theTormentOfGollumOracle   = "8b70aad1-bcb1-4359-b27e-ba8b40ef2752"
	diplomacyOfTheWastesOracle = "8b4670b9-6701-4ca6-afd7-256742267e86"
)

// poolAHand is one hand with something for every pool A filter, in
// this order:
//
//	0 Forest              basic land
//	1 Snow-Covered Swamp  basic land (snow is a supertype, not "nonbasic")
//	2 Sacred Foundry      nonbasic land with basic land types (CR 205.4c)
//	3 Lightning Bolt      instant, mana value 1
//	4 Grizzly Bears       creature, mana value 2
//	5 Jace                legendary planeswalker, mana value 4
//	6 Sol Ring            artifact, mana value 1
//	7 Venser              legendary creature, mana value 5
//	8 Spirit              creature — Spirit, mana value 2
//	9 Arcane instant      instant — Arcane, mana value 2
//	10 Changeling Outcast creature with changeling, every creature type (CR 702.73a)
//	11 Fireball           sorcery {X}{R}, mana value 1 in a hand (CR 202.3e)
//	12 Rite               sorcery, mana value 4
//	13 Pelakka Predation  MDFC, a nonland sorcery by its front face (CR 712.8a)
func poolAHand() []game.Card {
	return []game.Card{
		rhForest(),
		{Name: "Snow-Covered Swamp", TypeLine: "Basic Snow Land — Swamp"},
		{Name: "Sacred Foundry", TypeLine: "Land — Mountain Plains"},
		rhBolt(),
		{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}},
		{Name: "Jace Beleren", TypeLine: "Legendary Planeswalker — Jace", ManaCost: "{1}{U}{U}{U}", Colors: []string{"U"}},
		{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
		{Name: "Venser, Shaper Savant", TypeLine: "Legendary Creature — Human Wizard", ManaCost: "{2}{U}{U}{U}", Colors: []string{"U"}},
		{Name: "Kami of Ancient Law", TypeLine: "Creature — Spirit", ManaCost: "{1}{W}", Colors: []string{"W"}},
		{Name: "Peer Through Depths", TypeLine: "Instant — Arcane", ManaCost: "{1}{U}", Colors: []string{"U"}},
		{Name: "Changeling Outcast", TypeLine: "Creature — Shapeshifter Rogue", ManaCost: "{B}", Colors: []string{"B"}, Keywords: []string{game.KeywordChangeling}},
		rhFireball(),
		{Name: "Rite of Replication", TypeLine: "Sorcery", ManaCost: "{2}{U}{U}", Colors: []string{"U"}},
		rhPelakka(),
	}
}

// Each card's filter, over poolAHand, offers exactly the cards its
// printed restriction names, under the label it prints.
func TestPoolAFiltersOfferOnlyWhatTheCardNames(t *testing.T) {
	all := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	nonland := []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	creatures := []int{4, 7, 8, 10}
	for _, tc := range []struct {
		name, oracle, label string
		want                []int
	}{
		{"Coercion", coercionOracle, "card", all},
		{"Dark Inquiry", darkInquiryOracle, "nonland card", nonland},
		{"Pilfer", pilferOracle, "nonland card", nonland},
		{"Distress", distressOracle, "nonland card", nonland},
		{"Despise", despiseOracle, "creature or planeswalker card", []int{4, 5, 7, 8, 10}},
		{"Divest", divestOracle, "artifact or creature card", []int{4, 6, 7, 8, 10}},
		{"Duress", duressOracle, "noncreature, nonland card", []int{3, 5, 6, 9, 11, 12, 13}},
		{"Encroach", encroachOracle, "nonbasic land card", []int{2}},
		{"Inquisition of Kozilek", inquisitionOfKozilekOracle, "nonland card with mana value 3 or less", []int{3, 4, 6, 8, 9, 10, 11, 13}},
		{"Lay Bare the Heart", layBareTheHeartOracle, "nonlegendary, nonland card", []int{3, 4, 6, 8, 9, 10, 11, 12, 13}},
		{"Ostracize", ostracizeOracle, "creature card", creatures},
		{"Harsh Scrutiny", harshScrutinyOracle, "creature card", creatures},
		{"Psychic Spear", psychicSpearOracle, "Spirit or Arcane card", []int{8, 9, 10}},
		{"Shattered Dreams", shatteredDreamsOracle, "artifact card", []int{6}},
		{"Brainbite", brainbiteOracle, "card", all},
		{"Thought Erasure", thoughtErasureOracle, "nonland card", nonland},
		{"Gix's Caress", gixsCaressOracle, "nonland card", nonland},
		{"Toll of the Invasion", tollOfTheInvasionOracle, "nonland card", nonland},
		{"The Torment of Gollum", theTormentOfGollumOracle, "nonland card", nonland},
		{"Diplomacy of the Wastes", diplomacyOfTheWastesOracle, "nonland card", nonland},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, victim := g.Seats[0], g.Seats[1]
			ids := revealHand(victim, poolAHand()...)

			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
			passPriorityAroundTable(t, g)

			pick := openPick(t, g)
			if pick.Chooser != caster.ID || pick.FromPlayer != victim.ID || pick.Count != 1 {
				t.Fatalf("prompt = chooser %v from %v count %d", pick.Chooser, pick.FromPlayer, pick.Count)
			}
			want := make([]uuid.UUID, 0, len(tc.want))
			for _, i := range tc.want {
				want = append(want, ids[i])
			}
			if !sameIDs(pick.DiscardOptions, want) {
				t.Errorf("options = %v, want hand positions %v", poolAPositions(ids, pick.DiscardOptions), tc.want)
			}
			if pick.DiscardLabel != tc.label {
				t.Errorf("label = %q, want %q", pick.DiscardLabel, tc.label)
			}
		})
	}
}

// poolAPositions turns option IDs back into hand positions, for a
// readable failure.
func poolAPositions(ids, options []uuid.UUID) []int {
	at := map[uuid.UUID]int{}
	for i, id := range ids {
		at[id] = i
	}
	out := make([]int, 0, len(options))
	for _, id := range options {
		out = append(out, at[id])
	}
	return out
}

// "Target opponent" cannot be the caster; "target player" can, and
// the caster then reveals to the table and picks from their own hand.
func TestPoolATargetClauses(t *testing.T) {
	for _, tc := range []struct {
		name, oracle  string
		mayTargetSelf bool
	}{
		{"Coercion", coercionOracle, false},
		{"Dark Inquiry", darkInquiryOracle, false},
		{"Despise", despiseOracle, false},
		{"Distress", distressOracle, true},
		{"Divest", divestOracle, true},
		{"Duress", duressOracle, false},
		{"Encroach", encroachOracle, true},
		{"Inquisition of Kozilek", inquisitionOfKozilekOracle, true},
		{"Lay Bare the Heart", layBareTheHeartOracle, false},
		{"Ostracize", ostracizeOracle, false},
		{"Pilfer", pilferOracle, false},
		{"Psychic Spear", psychicSpearOracle, true},
		{"Shattered Dreams", shatteredDreamsOracle, false},
		{"Brainbite", brainbiteOracle, false},
		{"Harsh Scrutiny", harshScrutinyOracle, false},
		{"Thought Erasure", thoughtErasureOracle, false},
		{"Gix's Caress", gixsCaressOracle, false},
		{"Toll of the Invasion", tollOfTheInvasionOracle, false},
		{"The Torment of Gollum", theTormentOfGollumOracle, false},
		{"Diplomacy of the Wastes", diplomacyOfTheWastesOracle, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			err := castCatalogSpellErr(t, g, tc.name, "Sorcery", tc.oracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
			if tc.mayTargetSelf && err != nil {
				t.Errorf("targeting its caster: %v", err)
			}
			if !tc.mayTargetSelf && err == nil {
				t.Error("a target-opponent spell targeted its caster")
			}
		})
	}
}

// Brainbite draws even when the opponent's hand is empty (its
// 2009-05-01 ruling): nothing is revealed or chosen, and the draw
// still happens.
func TestBrainbiteDrawsWithNothingToTake(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	revealHand(victim)

	castCatalogSpell(t, g, "Brainbite", "Sorcery", brainbiteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	hand := caster.Hand.Size()
	passPriorityAroundTable(t, g)

	poolANoPick(t, g)
	if caster.Hand.Size() != hand+1 {
		t.Errorf("caster hand = %d, want %d", caster.Hand.Size(), hand+1)
	}
}

// Harsh Scrutiny's scry is queued behind the pick, so it is answered
// after the discard, as printed; with no creature card it still
// happens (the 2016-09-20 ruling).
func TestHarshScrutinyScriesAfterThePick(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	revealHand(victim, rhBolt(), poolAHand()[4])

	castCatalogSpell(t, g, "Harsh Scrutiny", "Sorcery", harshScrutinyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

	poolAPickThen(t, g, caster.ID, game.PendingChoiceScry)

	g2 := newCatalogGame(t)
	revealHand(g2.Seats[1], rhBolt(), rhForest())
	castCatalogSpell(t, g2, "Harsh Scrutiny", "Sorcery", harshScrutinyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g2.Seats[1].ID}})
	passPriorityAroundTable(t, g2)
	poolANoPick(t, g2)
	if !poolAHasChoice(g2, g2.Seats[0].ID, game.PendingChoiceScry) {
		t.Error("no scry with no creature card to take")
	}
}

// Thought Erasure's surveil is queued behind the pick.
func TestThoughtErasureSurveilsAfterThePick(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	revealHand(victim, rhBolt())

	castCatalogSpell(t, g, "Thought Erasure", "Sorcery", thoughtErasureOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

	poolAPickThen(t, g, caster.ID, game.PendingChoiceSurveil)
}

// Gix's Caress makes one tapped Powerstone for its caster, even with
// nothing to take.
func TestGixsCaressMakesATappedPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	revealHand(victim, rhForest())

	castCatalogSpell(t, g, "Gix's Caress", "Sorcery", gixsCaressOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

	poolANoPick(t, g)
	var stones []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Powerstone") {
			stones = append(stones, c)
		}
	}
	if len(stones) != 1 {
		t.Fatalf("%d Powerstones, want 1", len(stones))
	}
	if stones[0].Controller != caster.ID || !stones[0].Tapped {
		t.Errorf("Powerstone controller %v tapped %v, want the caster's and tapped", stones[0].Controller, stones[0].Tapped)
	}
}

// Toll of the Invasion amasses Zombies 1 and The Torment of Gollum
// amasses Orcs 2, each even with nothing to take (Toll's 2019-05-03
// ruling).
func TestPoolAAmassAfterThePick(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, subtype string
		n                     int
	}{
		{"Toll of the Invasion", tollOfTheInvasionOracle, "Zombie", 1},
		{"The Torment of Gollum", theTormentOfGollumOracle, "Orc", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, victim := g.Seats[0], g.Seats[1]
			revealHand(victim, rhForest())

			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
			passPriorityAroundTable(t, g)

			poolANoPick(t, g)
			army := armyOf(t, g, caster.ID)
			if !army.HasSubtype(tc.subtype) {
				t.Errorf("the Army is %q, want a %s Army", army.TypeLine, tc.subtype)
			}
			if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != tc.n {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.n)
			}
		})
	}
}

// Diplomacy of the Wastes: the target loses 2 life only if the caster
// controls a Warrior, and then even with nothing discarded (the
// 2014-11-24 ruling).
func TestDiplomacyOfTheWastesDrainsOnlyWithAWarrior(t *testing.T) {
	for _, warrior := range []bool{false, true} {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		revealHand(victim)
		if warrior {
			pushPermanentForTest(g, caster.ID, "Mardu Hordechief", "", "Creature — Human Warrior")
		}
		// An opponent's Warrior is not "a Warrior you control".
		pushPermanentForTest(g, victim.ID, "Goblin Warrior", "", "Creature — Goblin Warrior")
		life := victim.Life

		castCatalogSpell(t, g, "Diplomacy of the Wastes", "Sorcery", diplomacyOfTheWastesOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
		passPriorityAroundTable(t, g)

		want := life
		if warrior {
			want = life - 2
		}
		if victim.Life != want {
			t.Errorf("warrior=%v: target life = %d, want %d", warrior, victim.Life, want)
		}
	}
}

// poolANoPick fails if a revealed-hand prompt is open.
func poolANoPick(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
			t.Fatalf("a revealed-hand prompt is open: %+v", c)
		}
	}
}

func poolAHasChoice(g *game.Game, chooser uuid.UUID, kind game.PendingChoiceKind) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			return true
		}
	}
	return false
}

// poolAPickThen checks the pick is open and a `then` prompt for the
// caster is queued after it, then answers the pick and checks the
// `then` prompt is still waiting.
func poolAPickThen(t *testing.T, g *game.Game, caster uuid.UUID, then game.PendingChoiceKind) {
	t.Helper()
	pick := openPick(t, g)
	pickAt, thenAt := -1, -1
	for i, c := range g.PendingChoices {
		switch {
		case c == nil:
		case c.ID == pick.ID:
			pickAt = i
		case c.Kind == then && c.Chooser == caster:
			thenAt = i
		}
	}
	if thenAt < 0 || thenAt < pickAt {
		t.Fatalf("%s prompt at %d, pick at %d: want it queued after the pick", then, thenAt, pickAt)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster, pick.DiscardOptions[:1]); err != nil {
		t.Fatalf("answering the pick: %v", err)
	}
	if !poolAHasChoice(g, caster, then) {
		t.Errorf("the %s prompt went away with the pick", then)
	}
}
