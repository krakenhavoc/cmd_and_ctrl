package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the announcement -------------------------------------------------

// Demand Answers: the sacrifice branch sacrifices the artifact with the
// spell on the stack, records branch 0, and draws two.
func TestDemandAnswersSacrificeBranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.EitherPermanent(g, me.ID, "Relic", "Artifact")
	id, err := effectstest.CastWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", effectstest.DemandAnswersOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(0), SacrificeIDs: []uuid.UUID{relic}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if effects.OnBattlefield(g, relic) {
		t.Fatal("the artifact is still on the battlefield with the spell on the stack")
	}
	item := g.StackMeta[id]
	if item == nil || item.Paid.CostBranch != 1 || item.Paid.Sacrificed != 1 {
		t.Fatalf("Paid = %+v, want CostBranch 1 (branch 0) and one sacrifice", item.Paid)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d, want 2", got)
	}
}

// The discard branch discards the named card, and the record keeps it.
func TestDemandAnswersDiscardBranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pitch := effectstest.HandCard(me, "Pitch", "Instant")
	id, err := effectstest.CastWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", effectstest.DemandAnswersOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(1), DiscardIDs: []uuid.UUID{pitch}})
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.EitherPermanent(g, me.ID, "Relic", "Artifact")
	pitch := effectstest.HandCard(me, "Pitch", "Instant")
	for _, tc := range []struct {
		why    string
		params game.CastSpellParams
	}{
		{"no branch", game.CastSpellParams{DiscardIDs: []uuid.UUID{pitch}}},
		{"branch out of range", game.CastSpellParams{CostBranch: effectstest.Branch(2), DiscardIDs: []uuid.UUID{pitch}}},
		{"negative branch", game.CastSpellParams{CostBranch: effectstest.Branch(-1), DiscardIDs: []uuid.UUID{pitch}}},
		{"discard sent for the sacrifice branch", game.CastSpellParams{CostBranch: effectstest.Branch(0), DiscardIDs: []uuid.UUID{pitch}}},
		{"sacrifice sent for the discard branch", game.CastSpellParams{CostBranch: effectstest.Branch(1), SacrificeIDs: []uuid.UUID{relic}}},
		{"both branches paid", game.CastSpellParams{CostBranch: effectstest.Branch(0), SacrificeIDs: []uuid.UUID{relic}, DiscardIDs: []uuid.UUID{pitch}}},
	} {
		if _, err := effectstest.CastWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", effectstest.DemandAnswersOracle, tc.params); err == nil {
			t.Errorf("%s: cast accepted", tc.why)
		}
		if !effects.OnBattlefield(g, relic) || !me.Hand.Contains(pitch) {
			t.Fatalf("%s: a refused cast paid something", tc.why)
		}
	}
}

// A branch on a card with no either/or cost is refused, not ignored.
func TestCostBranchOnACardWithoutBranchesIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if _, err := effectstest.CastWithTapParams(t, g, "Murderous Cut", "Instant", "{4}{B}", effectstest.MurderousCutOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(0)}); !errors.Is(err, game.ErrCostBranch) {
		t.Fatalf("CastSpell = %v, want ErrCostBranch", err)
	}
}

// CR 601.2h / 118.3: a branch the caster cannot pay is refused even
// when the payload looks complete — Demand Answers' sacrifice branch
// with no artifact to sacrifice.
func TestAnUnpayableBranchIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	creature := effectstest.EitherPermanent(g, me.ID, "Bear", "Creature — Bear")
	if _, err := effectstest.CastWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", effectstest.DemandAnswersOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(0), SacrificeIDs: []uuid.UUID{creature}}); err == nil {
		t.Fatal("a creature paid Demand Answers' 'sacrifice an artifact' branch")
	}
}

// --- the price ---------------------------------------------------------

// Lightning Axe: the mana branch joins the total at CR 601.2f, so the
// one pricer charges {5}{R}; the discard branch is {R}; an unannounced
// branch is quoted without branch mana.
func TestLightningAxePricesTheManaBranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	axe := game.Card{InstanceID: uuid.New(), Name: "Lightning Axe", TypeLine: "Instant",
		ManaCost: "{R}", OracleID: effectstest.LightningAxeOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(axe)
	for _, tc := range []struct {
		branch *int
		want   int
	}{{nil, 1}, {effectstest.Branch(0), 1}, {effectstest.Branch(1), 6}} {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		foe := g.Seats[1]
		target := effectstest.EitherPermanent(g, foe.ID, "Ogre", "Creature — Ogre")
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
			params.CostBranch = effectstest.Branch(1)
		} else {
			params.CostBranch = effectstest.Branch(0)
			params.DiscardIDs = []uuid.UUID{effectstest.HandCard(me, "Pitch", "Instant")}
		}
		me.ManaPool.AddMana(game.ManaToken{Color: "R"})
		if _, err := effectstest.CastWithTapParams(t, g, "Lightning Axe", "Instant", "{R}", effectstest.LightningAxeOracle, params); err != nil {
			t.Fatalf("mana branch %v: CastSpell: %v", useMana, err)
		}
		if n := len(me.ManaPool); n != 0 {
			t.Fatalf("mana branch %v: %d mana left floating", useMana, n)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effects.OnBattlefield(g, target) {
			t.Fatalf("mana branch %v: the 2/2 survived 5 damage", useMana)
		}
	}
}

// A strict mana-branch cast with only {R} floating is refused for mana,
// and the refusal pays nothing.
func TestLightningAxeManaBranchNeedsTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	target := effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	_, err := effectstest.CastWithTapParams(t, g, "Lightning Axe", "Instant", "{R}", effectstest.LightningAxeOracle,
		game.CastSpellParams{Strict: true, CostBranch: effectstest.Branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("CastSpell = %v, want InsufficientManaError", err)
	}
}

// --- life and blight branches ----------------------------------------

// Bitter Triumph's life branch pays 3 life (CR 119.4) and is refused
// below 3; the discard branch is still open then.
func TestBitterTriumphLifeBranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	life := me.Life
	if _, err := effectstest.CastWithTapParams(t, g, "Bitter Triumph", "Instant", "{1}{B}", effectstest.BitterTriumphOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d", me.Life, life-3)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, target) {
		t.Fatal("the target survived")
	}

	me.Life = 2
	other := effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre 2", "Creature — Ogre")
	if _, err := effectstest.CastWithTapParams(t, g, "Bitter Triumph", "Instant", "{1}{B}", effectstest.BitterTriumphOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Fatal("paid 3 life at 2")
	}
	if me.Life != 2 {
		t.Fatalf("a refused cast changed life to %d", me.Life)
	}
}

// Bogslither's Embrace's blight branch puts the -1/-1 counter on the
// named creature as a cost (CR 701.68a) and exiles the target.
func TestBogslithersEmbraceBlightBranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.EitherPermanent(g, me.ID, "Bear", "Creature — Bear")
	target := effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	if _, err := effectstest.CastWithTapParams(t, g, "Bogslither's Embrace", "Sorcery", "{1}{B}", effectstest.BogslithersEmbraceOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(0), BlightIDs: []uuid.UUID{mine},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	c, ok := effectstest.BattlefieldCard(g, mine)
	if !ok || c.Counters[game.CounterMinusOne] != 1 {
		t.Fatalf("blighted creature counters = %v, want one -1/-1", c.Counters)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, target) || !g.Exile.Contains(target) {
		t.Fatal("the target was not exiled")
	}
	// The mana branch names no creature, and one named is refused.
	other := effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre 2", "Creature — Ogre")
	if _, err := effectstest.CastWithTapParams(t, g, "Bogslither's Embrace", "Sorcery", "{1}{B}", effectstest.BogslithersEmbraceOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(1), BlightIDs: []uuid.UUID{mine},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Fatal("blight_ids on the mana branch were accepted")
	}
}

// Wild Unraveling's blight branch with no creature to blight is refused
// (CR 701.68b); the {1} branch counters the spell.
func TestWildUnravelingBranches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	victim := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationDelveOracle, nil)
	if _, err := effectstest.CastWithTapParams(t, g, "Wild Unraveling", "Instant", "{U}{U}", effectstest.WildUnravelingOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(0), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Fatal("blight with no creature was accepted")
	}
	if _, err := effectstest.CastWithTapParams(t, g, "Wild Unraveling", "Instant", "{U}{U}", effectstest.WildUnravelingOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(1), Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Fatal("the countered spell is not in its owner's graveyard")
	}
}

// --- the records --------------------------------------------------------

// Grab the Prize: a nonland discard deals 2 to each opponent; a land
// discard does not. Both draw two. The record names the discarded card.
func TestGrabThePrizeReadsTheDiscardedCard(t *testing.T) {
	for _, land := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		typeLine := "Instant"
		if land {
			typeLine = "Basic Land — Mountain"
		}
		pitch := effectstest.HandCard(me, "Pitch", typeLine)
		id, err := effectstest.CastWithTapParams(t, g, "Grab the Prize", "Sorcery", "{1}{R}", effectstest.GrabThePrizeOracle,
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
		effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		fodder := effectstest.EitherPermanent(g, me.ID, "Bear", "Creature — Bear")
		if tc.modified {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == fodder {
					g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
				}
			}
		}
		target := effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
		_, err := effectstest.CastWithTapParams(t, g, "Lethal Throwdown", "Sorcery", "{B}", effectstest.LethalThrowdownOracle,
			game.CastSpellParams{CostBranch: effectstest.Branch(tc.branch), SacrificeIDs: []uuid.UUID{fodder},
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
		if (err == nil) != tc.accepted {
			t.Fatalf("%+v: CastSpell = %v", tc, err)
		}
		if err != nil {
			continue
		}
		hand := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("%+v: drew %d", tc, got)
		}
		if effects.OnBattlefield(g, target) {
			t.Errorf("%+v: the target survived", tc)
		}
	}
}

// Minion Missile destroys the creature and deals 2 to its controller.
func TestMinionMissileHitsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	foe := g.Seats[2]
	target := effectstest.EitherPermanent(g, foe.ID, "Ogre", "Creature — Ogre")
	life := foe.Life
	if _, err := effectstest.CastWithTapParams(t, g, "Minion Missile", "Sorcery", "{1}{B}", effectstest.MinionMissileOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(1), DiscardIDs: []uuid.UUID{effectstest.HandCard(me, "Pitch", "Instant")},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, target) || foe.Life != life-2 {
		t.Fatalf("target on battlefield %v, controller life %d (was %d)", effects.OnBattlefield(g, target), foe.Life, life)
	}
}

// Souls of the Lost counts the permanent cards in its controller's
// graveyard — including the one its own cost put there.
func TestSoulsOfTheLostCountsPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Spell", TypeLine: "Instant", Owner: me.ID})
	fodder := effectstest.EitherPermanent(g, me.ID, "Relic", "Artifact")
	id, err := effectstest.CastWithTapParams(t, g, "Souls of the Lost", "Creature — Spirit", "{1}{B}", effectstest.SoulsOfTheLostOracle,
		game.CastSpellParams{CostBranch: effectstest.Branch(1), SacrificeIDs: []uuid.UUID{fodder}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardFor(g, id)
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
		{"Demand Answers", "Instant", "{1}{R}", effectstest.DemandAnswersOracle, "none"},
		{"Bone Shards", "Sorcery", "{B}", effectstest.BoneShardsOracle, "creature"},
		{"Lightning Axe", "Instant", "{R}", effectstest.LightningAxeOracle, "creature"},
		{"Bitter Triumph", "Instant", "{1}{B}", effectstest.BitterTriumphOracle, "creature"},
		{"Eaten Alive", "Sorcery", "{B}", effectstest.EatenAliveOracle, "creature"},
		{"Spark Harvest", "Sorcery", "{B}", effectstest.SparkHarvestOracle, "creature"},
		{"Lash of the Balrog", "Sorcery", "{B}", effectstest.LashOfTheBalrogOracle, "creature"},
		{"Annihilating Glare", "Sorcery", "{B}", effectstest.AnnihilatingGlareOracle, "creature"},
		{"Deadly Precision", "Sorcery", "{B}", effectstest.DeadlyPrecisionOracle, "creature"},
		{"Stir Up Trouble", "Sorcery", "{B}", effectstest.StirUpTroubleOracle, "creature"},
		{"Final Payment", "Instant", "{W}{B}", effectstest.FinalPaymentOracle, "creature"},
		{"Pumpkin Bombardment", "Sorcery", "{B/R}", effectstest.PumpkinBombardmentOracle, "creature"},
		{"Bogslither's Embrace", "Sorcery", "{1}{B}", effectstest.BogslithersEmbraceOracle, "creature"},
		{"Wild Unraveling", "Instant", "{U}{U}", effectstest.WildUnravelingOracle, "spell"},
		{"Morkrut Behemoth", "Creature — Zombie Giant", "{4}{B}", effectstest.MorkrutBehemothOracle, "none"},
		{"Bayou Groff", "Creature — Plant Dog", "{1}{G}", effectstest.BayouGroffOracle, "none"},
		{"Minion Missile", "Sorcery", "{1}{B}", effectstest.MinionMissileOracle, "creature"},
		{"Louisoix's Sacrifice", "Instant", "{U}", effectstest.LouisoixsSacrificeOracle, "spell"},
		{"Lethal Throwdown", "Sorcery", "{B}", effectstest.LethalThrowdownOracle, "creature"},
		{"Souls of the Lost", "Creature — Spirit", "{1}{B}", effectstest.SoulsOfTheLostOracle, "none"},
	} {
		ac := game.AdditionalCostFor(c.oracle)
		if !ac.Branched() {
			t.Fatalf("%s: no either/or cost registered", c.name)
		}
		for i, b := range ac.Either {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			params := game.CastSpellParams{CostBranch: effectstest.Branch(i)}
			if b.DiscardCards > 0 {
				params.DiscardIDs = []uuid.UUID{effectstest.HandCard(me, "Pitch", "Instant")}
			}
			if b.Sacrifice != nil {
				// A legendary, modified artifact creature enchantment pays
				// every sacrifice clause in the set.
				fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Legendary Artifact Enchantment Creature — Golem",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
					Counters: map[string]int{game.CounterPlusOne: 1},
				})
				params.SacrificeIDs = []uuid.UUID{fodder}
			}
			if b.Blight > 0 {
				params.BlightIDs = []uuid.UUID{effectstest.EitherPermanent(g, me.ID, "Bear", "Creature — Bear")}
			}
			switch c.target {
			case "creature":
				params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")}}
			case "spell":
				victim := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationDelveOracle, nil)
				params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
			}
			life := me.Life
			if _, err := effectstest.CastWithTapParams(t, g, c.name, c.typeLine, c.cost, c.oracle, params); err != nil {
				t.Fatalf("%s branch %d (%s): CastSpell: %v", c.name, i, b.Key, err)
			}
			if me.Life != life-b.PayLife {
				t.Errorf("%s branch %d: paid %d life, want %d", c.name, i, life-me.Life, b.PayLife)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if c.target == "creature" && effects.OnBattlefield(g, params.Targets[0].ID) {
				t.Errorf("%s branch %d: the target is still on the battlefield", c.name, i)
			}
		}
	}
}

// Redirect Lightning: the life branch pays 5 life and the mana branch
// is priced {2}{R}; either way it targets a spell with a single target.
func TestRedirectLightningBranches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ogre := effectstest.EitherPermanent(g, g.Seats[1].ID, "Ogre", "Creature — Ogre")
	cut := effectstest.CastCatalogSpell(t, g, "Murderous Cut", "Instant", effectstest.MurderousCutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	redirect := game.Card{InstanceID: uuid.New(), Name: "Redirect Lightning", TypeLine: "Instant — Lesson",
		ManaCost: "{R}", OracleID: effectstest.RedirectLightningOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(redirect)
	price, err := g.PriceCast(me.ID, redirect, game.CastSpellParams{CostBranch: effectstest.Branch(1)})
	if err != nil || price.Total.ManaValue() != 3 {
		t.Fatalf("mana branch price = %s (%v), want {2}{R}", price.Total.String(), err)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, redirect.InstanceID, game.CastSpellParams{CostBranch: effectstest.Branch(0),
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
		{"branch(es)", effects.EitherCost(effects.DiscardCost(1).Keyed("discard"))},
		{"no Key", effects.EitherCost(effects.DiscardCost(1).Keyed("discard"), effects.ManaAdditionalCost("{2}"))},
		{"two either/or branches keyed", effects.EitherCost(effects.DiscardCost(1).Keyed("x"), effects.ManaAdditionalCost("{2}").Keyed("x"))},
		{"components of its own", func() *game.AdditionalCost {
			c := effects.EitherCost(effects.DiscardCost(1).Keyed("discard"), effects.ManaAdditionalCost("{2}").Keyed("mana"))
			c.DiscardCards = 1
			return c
		}()},
		{"is Optional", func() *game.AdditionalCost {
			k := effects.Kicker("{1}")
			return effects.EitherCost(&k, effects.DiscardCost(1).Keyed("discard"))
		}()},
		{"is itself branched", effects.EitherCost(
			effects.EitherCost(effects.DiscardCost(1).Keyed("a"), effects.ManaAdditionalCost("{1}").Keyed("b")).Keyed("nested"),
			effects.DiscardCost(1).Keyed("discard"))},
		{"sacrifices 1 to 0", effects.EitherCost(
			(&game.AdditionalCost{Sacrifice: effects.TargetPermanent("creatures", effects.Creature()).WithCount(1, 0), Label: "Sacrifice one or more"}).Keyed("sac"),
			effects.DiscardCost(1).Keyed("discard"))},
		{"unparseable mana cost", effects.EitherCost(effects.ManaAdditionalCost("{Q}").Keyed("mana"), effects.DiscardCost(1).Keyed("discard"))},
		{"demands nothing", effects.EitherCost((&game.AdditionalCost{Label: "Nothing"}).Keyed("nothing"), effects.DiscardCost(1).Keyed("discard"))},
	} {
		effectstest.MustPanic(t, tc.want, func() {
			effects.Register(effects.Spec{OracleID: "either-register-test-" + tc.want, Name: "Bad Either", AdditionalCost: tc.cost})
		})
	}
	effectstest.MustPanic(t, "mandatory additional MANA cost", func() {
		effects.Register(effects.Spec{OracleID: "either-register-test-mana", Name: "Bad Mana", AdditionalCost: effects.ManaAdditionalCost("{2}")})
	})
	effectstest.MustPanic(t, "pays fixed life", func() {
		pl := effects.PayLifeCost(2)
		pl.Optional, pl.Key = true, "life"
		effects.Register(effects.Spec{OracleID: "either-register-test-life", Name: "Bad Life", OptionalCosts: []game.AdditionalCost{*pl}})
	})
}
