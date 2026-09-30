package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// either_cost_test.go — ADR 0100 sub-PR 3: either/or additional costs
// (CR 601.2b, 601.2f–h) and the discard record (owner decision 6). The
// plan arithmetic is tested next to the code in game/; these are the
// board-level rules — what the announcement must name, what each branch
// pays, and what the resolution reads back — and one test per card.

const (
	demandAnswersOracle      = "c11e84a1-dbda-429b-8cd6-fd0deaefc689"
	boneShardsOracle         = "6e760cfe-45b9-4a3f-b2e2-4ca6ec2bd13a"
	lightningAxeOracle       = "81b90905-fbc0-426a-a084-c3300533abb4"
	grabThePrizeOracle       = "a37387f0-9d57-49ff-aa38-2c6b359e8916"
	bitterTriumphOracle      = "776341cb-d2ec-423f-9250-92dc8bd8d503"
	eatenAliveOracle         = "d437ecc2-2fd3-4ad3-b23e-217f55e58dae"
	sparkHarvestOracle       = "25756e33-25e5-4976-8cce-a39fa2c4b007"
	lashOfTheBalrogOracle    = "66499407-3ad2-464b-bbfe-95d860f53616"
	annihilatingGlareOracle  = "a75087aa-41bd-4e0f-a118-61b5209356f5"
	deadlyPrecisionOracle    = "d23ca3d2-0731-4df4-a3c6-69ddf0c58601"
	stirUpTroubleOracle      = "dda607bd-f419-4b7f-b052-a5ce6ce22bfe"
	finalPaymentOracle       = "e47d57dc-2e69-4939-8aec-077595f2ae05"
	pumpkinBombardmentOracle = "37827e84-9f5f-49ed-b939-0cf80dac82e7"
	bogslithersEmbraceOracle = "21695c70-f746-4d77-b6d3-e2713a5c929e"
	wildUnravelingOracle     = "dd1a4219-437b-494f-ae0e-1d3181ffcd63"
	morkrutBehemothOracle    = "76c2571d-7d45-4c7a-99d1-302e2b26f8f6"
	bayouGroffOracle         = "e83b2790-3cec-429c-a65d-6f4c7d025d37"
	minionMissileOracle      = "670cf656-5c72-4d28-b5e6-dc77781dc001"
	louisoixsSacrificeOracle = "42d8ab49-703a-4fd3-84d0-0a3ac1eafca8"
	redirectLightningOracle  = "c62b5c22-e058-436f-9575-a01cb2112829"
	lethalThrowdownOracle    = "4ab565bb-3188-4d23-93d8-5de67fc0d056"
	soulsOfTheLostOracle     = "c05d5a51-8e39-4475-9743-1adc60283c5a"
)

func branch(i int) *int { return &i }

// eitherPermanent puts a permanent the active seat controls onto the
// battlefield and returns its ID.
func eitherPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// --- the announcement -------------------------------------------------

// Demand Answers: the sacrifice branch sacrifices the artifact with the
// spell on the stack, records branch 0, and draws two.
func TestDemandAnswersSacrificeBranch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := eitherPermanent(g, me.ID, "Relic", "Artifact")
	id, err := castWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", demandAnswersOracle,
		game.CastSpellParams{CostBranch: branch(0), SacrificeIDs: []uuid.UUID{relic}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if onBattlefield(g, relic) {
		t.Fatal("the artifact is still on the battlefield with the spell on the stack")
	}
	item := g.StackMeta[id]
	if item == nil || item.Paid.CostBranch != 1 || item.Paid.Sacrificed != 1 {
		t.Fatalf("Paid = %+v, want CostBranch 1 (branch 0) and one sacrifice", item.Paid)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d, want 2", got)
	}
}

// The discard branch discards the named card, and the record keeps it.
func TestDemandAnswersDiscardBranch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pitch := handCard(me, "Pitch", "Instant")
	id, err := castWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", demandAnswersOracle,
		game.CastSpellParams{CostBranch: branch(1), DiscardIDs: []uuid.UUID{pitch}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Fatal("the discarded card is not in the graveyard")
	}
	item := g.StackMeta[id]
	if item.Paid.CostBranch != 2 || len(item.Paid.Discarded) != 1 || item.Paid.Discarded[0] != pitch {
		t.Fatalf("Paid = %+v, want CostBranch 2 and the pitched card recorded", item.Paid)
	}
}

// The branch is required on a branched card, an out-of-range branch is
// refused, and a payload for the branch NOT announced fails the
// exact-consumption check. Nothing moves on a refusal.
func TestEitherCostAnnouncementRefusals(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := eitherPermanent(g, me.ID, "Relic", "Artifact")
	pitch := handCard(me, "Pitch", "Instant")
	for _, tc := range []struct {
		why    string
		params game.CastSpellParams
	}{
		{"no branch", game.CastSpellParams{DiscardIDs: []uuid.UUID{pitch}}},
		{"branch out of range", game.CastSpellParams{CostBranch: branch(2), DiscardIDs: []uuid.UUID{pitch}}},
		{"negative branch", game.CastSpellParams{CostBranch: branch(-1), DiscardIDs: []uuid.UUID{pitch}}},
		{"discard sent for the sacrifice branch", game.CastSpellParams{CostBranch: branch(0), DiscardIDs: []uuid.UUID{pitch}}},
		{"sacrifice sent for the discard branch", game.CastSpellParams{CostBranch: branch(1), SacrificeIDs: []uuid.UUID{relic}}},
		{"both branches paid", game.CastSpellParams{CostBranch: branch(0), SacrificeIDs: []uuid.UUID{relic}, DiscardIDs: []uuid.UUID{pitch}}},
	} {
		if _, err := castWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", demandAnswersOracle, tc.params); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
		if !onBattlefield(g, relic) || !me.Hand.Contains(pitch) {
			t.Fatalf("%s: a refused cast paid something", tc.why)
		}
	}
}

// A branch on a card with no either/or cost is refused, not ignored.
func TestCostBranchOnACardWithoutBranchesIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := castWithTapParams(t, g, "Murderous Cut", "Instant", "{4}{B}", murderousCutOracle,
		game.CastSpellParams{CostBranch: branch(0)}); !errors.Is(err, game.ErrCostBranch) {
		t.Fatalf("CastSpell = %v, want ErrCostBranch", err)
	}
}

// CR 601.2h / 118.3: a branch the caster cannot pay is refused even
// when the payload looks complete — Demand Answers' sacrifice branch
// with no artifact to sacrifice.
func TestAnUnpayableBranchIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	creature := eitherPermanent(g, me.ID, "Bear", "Creature — Bear")
	if _, err := castWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", demandAnswersOracle,
		game.CastSpellParams{CostBranch: branch(0), SacrificeIDs: []uuid.UUID{creature}}); err == nil {
		t.Fatal("a creature paid Demand Answers' 'sacrifice an artifact' branch")
	}
}

// --- the price ---------------------------------------------------------

// Lightning Axe: the mana branch joins the total at CR 601.2f, so the
// one pricer charges {5}{R}; the discard branch is {R}; an unannounced
// branch is quoted without branch mana.
func TestLightningAxePricesTheManaBranch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	axe := game.Card{InstanceID: uuid.New(), Name: "Lightning Axe", TypeLine: "Instant",
		ManaCost: "{R}", OracleID: lightningAxeOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(axe)
	for _, tc := range []struct {
		branch *int
		want   int
	}{{nil, 1}, {branch(0), 1}, {branch(1), 6}} {
		price, err := g.PriceCast(me.ID, axe, game.CastSpellParams{CostBranch: tc.branch})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		if got := price.Total.ManaValue(); got != tc.want {
			t.Errorf("branch %v: total %s (%d), want mana value %d", tc.branch, price.Total.String(), got, tc.want)
		}
	}
}

// The mana branch is really charged: a strict cast with {5}{R} floating
// pays it and deals 5; the discard branch deals 5 too.
func TestLightningAxeBothBranchesDealFive(t *testing.T) {
	for _, useMana := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		foe := g.Seats[1]
		target := eitherPermanent(g, foe.ID, "Ogre", "Creature — Ogre")
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		params := game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}
		if useMana {
			for i := 0; i < 5; i++ {
				me.ManaPool.AddMana(game.ManaToken{Color: "C"})
			}
			params.CostBranch = branch(1)
		} else {
			params.CostBranch = branch(0)
			params.DiscardIDs = []uuid.UUID{handCard(me, "Pitch", "Instant")}
		}
		me.ManaPool.AddMana(game.ManaToken{Color: "R"})
		if _, err := castWithTapParams(t, g, "Lightning Axe", "Instant", "{R}", lightningAxeOracle, params); err != nil {
			t.Fatalf("mana branch %v: CastSpell: %v", useMana, err)
		}
		if n := len(me.ManaPool); n != 0 {
			t.Fatalf("mana branch %v: %d mana left floating", useMana, n)
		}
		passPriorityAroundTable(t, g)
		if onBattlefield(g, target) {
			t.Fatalf("mana branch %v: the 2/2 survived 5 damage", useMana)
		}
	}
}

// A strict mana-branch cast with only {R} floating is refused for mana,
// and the refusal pays nothing.
func TestLightningAxeManaBranchNeedsTheMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	target := eitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	_, err := castWithTapParams(t, g, "Lightning Axe", "Instant", "{R}", lightningAxeOracle,
		game.CastSpellParams{Strict: true, CostBranch: branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("CastSpell = %v, want InsufficientManaError", err)
	}
}

// --- life and blight branches ----------------------------------------

// Bitter Triumph's life branch pays 3 life (CR 119.4) and is refused
// below 3; the discard branch is still open then.
func TestBitterTriumphLifeBranch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := eitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	life := me.Life
	if _, err := castWithTapParams(t, g, "Bitter Triumph", "Instant", "{1}{B}", bitterTriumphOracle,
		game.CastSpellParams{CostBranch: branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d", me.Life, life-3)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, target) {
		t.Fatal("the target survived")
	}

	me.Life = 2
	other := eitherPermanent(g, g.Seats[1].ID, "Ogre 2", "Creature — Ogre")
	if _, err := castWithTapParams(t, g, "Bitter Triumph", "Instant", "{1}{B}", bitterTriumphOracle,
		game.CastSpellParams{CostBranch: branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Fatal("paid 3 life at 2")
	}
	if me.Life != 2 {
		t.Fatalf("a refused cast changed life to %d", me.Life)
	}
}

// Bogslither's Embrace's blight branch puts the -1/-1 counter on the
// named creature as a cost (CR 701.68a) and exiles the target.
func TestBogslithersEmbraceBlightBranch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := eitherPermanent(g, me.ID, "Bear", "Creature — Bear")
	target := eitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	if _, err := castWithTapParams(t, g, "Bogslither's Embrace", "Sorcery", "{1}{B}", bogslithersEmbraceOracle,
		game.CastSpellParams{CostBranch: branch(0), BlightIDs: []uuid.UUID{mine},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	c, ok := battlefieldCard(g, mine)
	if !ok || c.Counters[game.CounterMinusOne] != 1 {
		t.Fatalf("blighted creature counters = %v, want one -1/-1", c.Counters)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, target) || !g.Exile.Contains(target) {
		t.Fatal("the target was not exiled")
	}
	// The mana branch names no creature, and one named is refused.
	other := eitherPermanent(g, g.Seats[1].ID, "Ogre 2", "Creature — Ogre")
	if _, err := castWithTapParams(t, g, "Bogslither's Embrace", "Sorcery", "{1}{B}", bogslithersEmbraceOracle,
		game.CastSpellParams{CostBranch: branch(1), BlightIDs: []uuid.UUID{mine},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Fatal("blight_ids on the mana branch were accepted")
	}
}

// Wild Unraveling's blight branch with no creature to blight is refused
// (CR 701.68b); the {1} branch counters the spell.
func TestWildUnravelingBranches(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	if _, err := castWithTapParams(t, g, "Wild Unraveling", "Instant", "{U}{U}", wildUnravelingOracle,
		game.CastSpellParams{CostBranch: branch(0), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Fatal("blight with no creature was accepted")
	}
	if _, err := castWithTapParams(t, g, "Wild Unraveling", "Instant", "{U}{U}", wildUnravelingOracle,
		game.CastSpellParams{CostBranch: branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Fatal("the countered spell is not in its owner's graveyard")
	}
}

// --- the records --------------------------------------------------------

// Grab the Prize: a nonland discard deals 2 to each opponent; a land
// discard does not. Both draw two. The record names the discarded card.
func TestGrabThePrizeReadsTheDiscardedCard(t *testing.T) {
	for _, land := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		typeLine := "Instant"
		if land {
			typeLine = "Basic Land — Mountain"
		}
		pitch := handCard(me, "Pitch", typeLine)
		id, err := castWithTapParams(t, g, "Grab the Prize", "Sorcery", "{1}{R}", grabThePrizeOracle,
			game.CastSpellParams{DiscardIDs: []uuid.UUID{pitch}})
		if err != nil {
			t.Fatalf("land %v: CastSpell: %v", land, err)
		}
		if got := g.StackMeta[id].Paid.Discarded; len(got) != 1 || got[0] != pitch {
			t.Fatalf("land %v: Paid.Discarded = %v", land, got)
		}
		lives := make([]int, len(g.Seats))
		for i, p := range g.Seats {
			lives[i] = p.Life
		}
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != 2 {
			t.Fatalf("land %v: drew %d, want 2", land, got)
		}
		for i, p := range g.Seats {
			want := lives[i]
			if i != 0 && !land {
				want -= 2
			}
			if p.Life != want {
				t.Errorf("land %v: seat %d life %d, want %d", land, i, p.Life, want)
			}
		}
	}
}

// Lethal Throwdown draws only when the "modified" branch was paid —
// the announcement, not the board, answers "the modified creature".
func TestLethalThrowdownReadsTheBranch(t *testing.T) {
	for _, tc := range []struct {
		branch   int
		modified bool
		draws    int
		accepted bool
	}{
		{1, true, 1, true},
		{0, true, 0, true},
		{0, false, 0, true},
		{1, false, 0, false},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		fodder := eitherPermanent(g, me.ID, "Bear", "Creature — Bear")
		if tc.modified {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == fodder {
					g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
				}
			}
		}
		target := eitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
		_, err := castWithTapParams(t, g, "Lethal Throwdown", "Sorcery", "{B}", lethalThrowdownOracle,
			game.CastSpellParams{CostBranch: branch(tc.branch), SacrificeIDs: []uuid.UUID{fodder},
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
		if (err == nil) != tc.accepted {
			t.Fatalf("%+v: CastSpell = %v", tc, err)
		}
		if err != nil {
			continue
		}
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("%+v: drew %d", tc, got)
		}
		if onBattlefield(g, target) {
			t.Errorf("%+v: the target survived", tc)
		}
	}
}

// Minion Missile destroys the creature and deals 2 to its controller.
func TestMinionMissileHitsTheController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	foe := g.Seats[2]
	target := eitherPermanent(g, foe.ID, "Ogre", "Creature — Ogre")
	life := foe.Life
	if _, err := castWithTapParams(t, g, "Minion Missile", "Sorcery", "{1}{B}", minionMissileOracle,
		game.CastSpellParams{CostBranch: branch(1), DiscardIDs: []uuid.UUID{handCard(me, "Pitch", "Instant")},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, target) || foe.Life != life-2 {
		t.Fatalf("target on battlefield %v, controller life %d (was %d)", onBattlefield(g, target), foe.Life, life)
	}
}

// Souls of the Lost counts the permanent cards in its controller's
// graveyard — including the one its own cost put there.
func TestSoulsOfTheLostCountsPermanentCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Spell", TypeLine: "Instant", Owner: me.ID})
	fodder := eitherPermanent(g, me.ID, "Relic", "Artifact")
	id, err := castWithTapParams(t, g, "Souls of the Lost", "Creature — Spirit", "{1}{B}", soulsOfTheLostOracle,
		game.CastSpellParams{CostBranch: branch(1), SacrificeIDs: []uuid.UUID{fodder}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := battlefieldCardFor(g, id)
	if c == nil {
		t.Fatal("Souls of the Lost did not enter")
	}
	// Old Bear and the sacrificed Relic: two permanent cards.
	if c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Fatalf("Souls of the Lost is %d/%d, want 2/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// Every card of the set registers with a branched cost, and casts for
// each of its branches when the board can pay it — the per-card smoke
// test for the ones whose resolution is an existing body.
func TestEitherCostCardsCastForEachBranch(t *testing.T) {
	type card struct {
		name, typeLine, cost, oracle string
		target                       string // "creature", "spell", "none"
	}
	for _, c := range []card{
		{"Demand Answers", "Instant", "{1}{R}", demandAnswersOracle, "none"},
		{"Bone Shards", "Sorcery", "{B}", boneShardsOracle, "creature"},
		{"Lightning Axe", "Instant", "{R}", lightningAxeOracle, "creature"},
		{"Bitter Triumph", "Instant", "{1}{B}", bitterTriumphOracle, "creature"},
		{"Eaten Alive", "Sorcery", "{B}", eatenAliveOracle, "creature"},
		{"Spark Harvest", "Sorcery", "{B}", sparkHarvestOracle, "creature"},
		{"Lash of the Balrog", "Sorcery", "{B}", lashOfTheBalrogOracle, "creature"},
		{"Annihilating Glare", "Sorcery", "{B}", annihilatingGlareOracle, "creature"},
		{"Deadly Precision", "Sorcery", "{B}", deadlyPrecisionOracle, "creature"},
		{"Stir Up Trouble", "Sorcery", "{B}", stirUpTroubleOracle, "creature"},
		{"Final Payment", "Instant", "{W}{B}", finalPaymentOracle, "creature"},
		{"Pumpkin Bombardment", "Sorcery", "{B/R}", pumpkinBombardmentOracle, "creature"},
		{"Bogslither's Embrace", "Sorcery", "{1}{B}", bogslithersEmbraceOracle, "creature"},
		{"Wild Unraveling", "Instant", "{U}{U}", wildUnravelingOracle, "spell"},
		{"Morkrut Behemoth", "Creature — Zombie Giant", "{4}{B}", morkrutBehemothOracle, "none"},
		{"Bayou Groff", "Creature — Plant Dog", "{1}{G}", bayouGroffOracle, "none"},
		{"Minion Missile", "Sorcery", "{1}{B}", minionMissileOracle, "creature"},
		{"Louisoix's Sacrifice", "Instant", "{U}", louisoixsSacrificeOracle, "spell"},
		{"Lethal Throwdown", "Sorcery", "{B}", lethalThrowdownOracle, "creature"},
		{"Souls of the Lost", "Creature — Spirit", "{1}{B}", soulsOfTheLostOracle, "none"},
	} {
		ac := game.AdditionalCostFor(c.oracle)
		if !ac.Branched() {
			t.Fatalf("%s: no either/or cost registered", c.name)
		}
		for i, b := range ac.Either {
			g := newCatalogGame(t)
			me := g.Seats[0]
			params := game.CastSpellParams{CostBranch: branch(i)}
			if b.DiscardCards > 0 {
				params.DiscardIDs = []uuid.UUID{handCard(me, "Pitch", "Instant")}
			}
			if b.Sacrifice != nil {
				// A legendary, modified artifact creature enchantment pays
				// every sacrifice clause in the set.
				fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Legendary Artifact Enchantment Creature — Golem",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
					Counters: map[string]int{game.CounterPlusOne: 1},
				})
				params.SacrificeIDs = []uuid.UUID{fodder}
			}
			if b.Blight > 0 {
				params.BlightIDs = []uuid.UUID{eitherPermanent(g, me.ID, "Bear", "Creature — Bear")}
			}
			switch c.target {
			case "creature":
				params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: eitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")}}
			case "spell":
				victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
				params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
			}
			life := me.Life
			if _, err := castWithTapParams(t, g, c.name, c.typeLine, c.cost, c.oracle, params); err != nil {
				t.Fatalf("%s branch %d (%s): CastSpell: %v", c.name, i, b.Key, err)
			}
			if me.Life != life-b.PayLife {
				t.Errorf("%s branch %d: paid %d life, want %d", c.name, i, life-me.Life, b.PayLife)
			}
			passPriorityAroundTable(t, g)
			if c.target == "creature" && onBattlefield(g, params.Targets[0].ID) {
				t.Errorf("%s branch %d: the target is still on the battlefield", c.name, i)
			}
		}
	}
}

// Redirect Lightning: the life branch pays 5 life and the mana branch
// is priced {2}{R}; either way it targets a spell with a single target.
func TestRedirectLightningBranches(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ogre := eitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	cut := castCatalogSpell(t, g, "Murderous Cut", "Instant", murderousCutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	redirect := game.Card{InstanceID: uuid.New(), Name: "Redirect Lightning", TypeLine: "Instant — Lesson",
		ManaCost: "{R}", OracleID: redirectLightningOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(redirect)
	price, err := g.PriceCast(me.ID, redirect, game.CastSpellParams{CostBranch: branch(1)})
	if err != nil || price.Total.ManaValue() != 3 {
		t.Fatalf("mana branch price = %s (%v), want {2}{R}", price.Total.String(), err)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, redirect.InstanceID, game.CastSpellParams{CostBranch: branch(0),
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: cut}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if me.Life != life-5 {
		t.Fatalf("life %d, want %d", me.Life, life-5)
	}
	if g.StackMeta[redirect.InstanceID] == nil {
		t.Fatal("Redirect Lightning is not on the stack")
	}
}

// --- registration --------------------------------------------------------

func TestRegisterEitherCostShapesPanic(t *testing.T) {
	for _, tc := range []struct {
		want string
		cost *game.AdditionalCost
	}{
		{"branch(es)", EitherCost(DiscardCost(1).Keyed("discard"))},
		{"no Key", EitherCost(DiscardCost(1).Keyed("discard"), ManaAdditionalCost("{2}"))},
		{"two either/or branches keyed", EitherCost(DiscardCost(1).Keyed("x"), ManaAdditionalCost("{2}").Keyed("x"))},
		{"components of its own", func() *game.AdditionalCost {
			c := EitherCost(DiscardCost(1).Keyed("discard"), ManaAdditionalCost("{2}").Keyed("mana"))
			c.DiscardCards = 1
			return c
		}()},
		{"is Optional", func() *game.AdditionalCost {
			k := Kicker("{1}")
			return EitherCost(&k, DiscardCost(1).Keyed("discard"))
		}()},
		{"is itself branched", EitherCost(
			EitherCost(DiscardCost(1).Keyed("a"), ManaAdditionalCost("{1}").Keyed("b")).Keyed("nested"),
			DiscardCost(1).Keyed("discard"))},
		{"sacrifices 1 to 0", EitherCost(
			(&game.AdditionalCost{Sacrifice: TargetPermanent("creatures", Creature()).WithCount(1, 0), Label: "Sacrifice one or more"}).Keyed("sac"),
			DiscardCost(1).Keyed("discard"))},
		{"unparseable mana cost", EitherCost(ManaAdditionalCost("{Q}").Keyed("mana"), DiscardCost(1).Keyed("discard"))},
		{"demands nothing", EitherCost((&game.AdditionalCost{Label: "Nothing"}).Keyed("nothing"), DiscardCost(1).Keyed("discard"))},
	} {
		mustPanic(t, tc.want, func() {
			Register(Spec{OracleID: "either-register-test-" + tc.want, Name: "Bad Either", AdditionalCost: tc.cost})
		})
	}
	mustPanic(t, "mandatory additional MANA cost", func() {
		Register(Spec{OracleID: "either-register-test-mana", Name: "Bad Mana", AdditionalCost: ManaAdditionalCost("{2}")})
	})
	mustPanic(t, "pays fixed life", func() {
		pl := PayLifeCost(2)
		pl.Optional, pl.Key = true, "life"
		Register(Spec{OracleID: "either-register-test-life", Name: "Bad Life", OptionalCosts: []game.AdditionalCost{*pl}})
	})
}
