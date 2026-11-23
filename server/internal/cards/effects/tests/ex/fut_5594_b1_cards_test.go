package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fut_5594_b1_cards_test.go — Future Sight batch 1 (#5594). Every test
// plays the real catalog entry through the engine.

const (
	futB1BladeOfTheSixthPrideOracle = "b7894010-fdba-498e-97de-22bd70332439"
	futB1BlindPhantasmOracle        = "3b91be0e-0683-48df-91b0-25e90b492796"
	futB1FomoriNomadOracle          = "bfcaa0a4-2fc5-4e27-8b79-802e365c71f2"
	futB1MassOfGhoulsOracle         = "29646827-7b69-48c5-8273-4e66076e2384"
	futB1NessianCourserOracle       = "e876d1fc-3acd-41a3-a34b-2bfa83204393"
	futB1ForceOfSavageryOracle      = "26fbfab5-f479-49f2-b36b-a3ebf964741f"
	futB1LucentLiminidOracle        = "cf02b774-2b3a-4459-bf59-e5d291691e77"
	futB1WhipSpineDrakeOracle       = "70e165a1-494a-4728-8d7e-1a9f306414db"
	futB1KnightOfSursiOracle        = "8dfe5fb5-89f1-42be-9038-40d3c71ce59a"
	futB1InfiltratorIlKorOracle     = "dd491ac7-3c6c-4a47-94af-071d1edc5eb2"
	futB1BogardanLancerOracle       = "67953edb-9681-4339-a5d2-917082925c24"
	futB1CharRumblerOracle          = "758aeca5-6ef3-4463-a65a-66b39289cc86"
	futB1CutthroatIlDalOracle       = "c70526a6-9071-4f34-b5fb-9d075021f667"
	futB1AugurIlVecOracle           = "34e05dc1-35af-402d-96ae-e249b3738493"
	futB1AugurOfSkullsOracle        = "723398f2-eaf2-4615-9078-418b9f540a33"
	futB1AvenAugurOracle            = "e68d1b86-3996-4bbf-9086-a76a7f9ce2a4"
	futB1EmberwildeAugurOracle      = "39bf60b2-eeee-48f3-b7b0-6cc2e038afb1"
	futB1LlanowarAugurOracle        = "b3052531-f7cf-44a7-9492-eb155d75268c"
	futB1SarcomiteMyrOracle         = "e78fd8f6-eca9-4107-aee7-a65689cee2f4"
	futB1WitchsMistOracle           = "4327225e-140f-45a4-a56f-dcf472f6c062"
	futB1ScourgeOfKherRidgesOracle  = "a3ced75d-eb1b-4a56-bbe0-49e8b906f729"
	futB1BloodshotTraineeOracle     = "87ab1c79-79f5-48a9-af1a-7c876684c8d9"
	futB1BondedFetchOracle          = "79d34537-4633-4728-a1b3-5822119f0c84"
	futB1KeldonMegalithsOracle      = "ec0ea7f7-52ce-40d1-b34c-e36dd4b26120"
	futB1TolariaWestOracle          = "46eefc72-2d9e-4389-8ae1-26d9ee472b5c"
	futB1NewBenaliaOracle           = "6e743fbf-b5b6-4176-a4f2-6933f521f2fe"
	futB1ForeseeOracle              = "441de890-7b9f-42f7-859a-a154b8ed74ac"
	futB1SlaughterPactOracle        = "67c4b1d5-757c-46cf-8bb1-9889d9ed1a1f"
	futB1PactOfTheTitanOracle       = "cf4959bf-92ae-4c9b-b631-724ce100cbb8"
	futB1MinionsMurmursOracle       = "fd755ec1-6727-4165-a7fd-5d9b59754ada"
	futB1EvenTheOddsOracle          = "6cd3605c-7885-4d32-8585-3b0b0f733702"
	futB1MysticSpeculationOracle    = "a58adf82-d9e2-4b47-933e-68180f280ba2"
	futB1MarshalingCryOracle        = "0dc30c06-3900-44e8-9e63-c232e42c1b5c"
	futB1RiftsweeperOracle          = "1ba0dce5-e767-47cc-9967-174e52bce120"
	futB1StrongholdRatsOracle       = "e91053ae-682a-4bbe-8394-fd65e626bacc"
	futB1BoundInSilenceOracle       = "b8cee34a-710d-4c09-b475-0e4754eccd66"
	futB1GiftOfGraniteOracle        = "e1abac67-ae07-4bb2-a048-7e41e0fc0371"
	futB1EmblemOfTheWarmindOracle   = "86e88219-0f8e-405d-8439-c3b4acf11330"

	futB1BearType = "Creature — Bear"
	futB1AuraType = "Enchantment — Aura"
)

// futB1Bear seeds a plain 2/2 non-catalog creature.
func futB1Bear(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, futB1BearType, 2, 2)
}

// futB1AnswerEveryDiscard answers each open discard prompt with the
// chooser's first card in hand until none is left.
func futB1AnswerEveryDiscard(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		answered := false
		for _, p := range g.Seats {
			if c := effectstest.DiscardChoiceFor(g, p.ID); c != nil && p.Hand.Size() > 0 {
				effectstest.AnswerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
				answered = true
			}
		}
		if !answered {
			return
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// --- every card is in the catalog ----------------------------------------

func TestFut5594B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		futB1BladeOfTheSixthPrideOracle, futB1BlindPhantasmOracle, futB1FomoriNomadOracle, futB1MassOfGhoulsOracle,
		futB1NessianCourserOracle, futB1ForceOfSavageryOracle, futB1LucentLiminidOracle, futB1WhipSpineDrakeOracle,
		futB1KnightOfSursiOracle, futB1InfiltratorIlKorOracle, futB1BogardanLancerOracle, futB1CharRumblerOracle,
		futB1CutthroatIlDalOracle, futB1AugurIlVecOracle, futB1AugurOfSkullsOracle,
		futB1AvenAugurOracle, futB1EmberwildeAugurOracle, futB1LlanowarAugurOracle, futB1SarcomiteMyrOracle,
		futB1WitchsMistOracle, futB1ScourgeOfKherRidgesOracle, futB1BloodshotTraineeOracle, futB1BondedFetchOracle,
		futB1KeldonMegalithsOracle, futB1TolariaWestOracle, futB1NewBenaliaOracle, futB1ForeseeOracle,
		futB1SlaughterPactOracle, futB1PactOfTheTitanOracle, futB1MinionsMurmursOracle, futB1EvenTheOddsOracle,
		futB1MysticSpeculationOracle, futB1MarshalingCryOracle, futB1RiftsweeperOracle,
		futB1StrongholdRatsOracle, futB1BoundInSilenceOracle, futB1GiftOfGraniteOracle, futB1EmblemOfTheWarmindOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords ------------------------------------------------------

func TestFut5594B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Force of Savagery", "Creature — Elemental", futB1ForceOfSavageryOracle, []string{"trample"}},
		{"Lucent Liminid", "Enchantment Creature — Elemental", futB1LucentLiminidOracle, []string{"flying"}},
		{"Whip-Spine Drake", "Creature — Drake", futB1WhipSpineDrakeOracle, []string{"flying"}},
		{"Knight of Sursi", "Creature — Human Knight", futB1KnightOfSursiOracle, []string{"flying", "flanking"}},
		{"Infiltrator il-Kor", "Creature — Kor Rogue", futB1InfiltratorIlKorOracle, []string{"shadow"}},
		{"Bogardan Lancer", "Creature — Human Knight", futB1BogardanLancerOracle, []string{"flanking"}},
		{"Char-Rumbler", "Creature — Elemental", futB1CharRumblerOracle, []string{"double strike"}},
		{"Augur il-Vec", "Creature — Human Cleric", futB1AugurIlVecOracle, []string{"shadow"}},
		{"Aven Augur", "Creature — Bird Wizard", futB1AvenAugurOracle, []string{"flying"}},
		{"Scourge of Kher Ridges", "Creature — Dragon", futB1ScourgeOfKherRidgesOracle, []string{"flying"}},
		{"Bonded Fetch", "Creature — Homunculus", futB1BondedFetchOracle, []string{"defender", "haste"}},
		{"Stronghold Rats", "Creature — Rat", futB1StrongholdRatsOracle, []string{"shadow"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// The vanilla bodies resolve as plain creatures.
func TestFut5594B1VanillaCreaturesResolve(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Blade of the Sixth Pride", "Creature — Cat Rebel", futB1BladeOfTheSixthPrideOracle, 3, 1},
		{"Blind Phantasm", "Creature — Illusion", futB1BlindPhantasmOracle, 2, 3},
		{"Fomori Nomad", "Creature — Nomad Giant", futB1FomoriNomadOracle, 4, 4},
		{"Mass of Ghouls", "Creature — Zombie Warrior", futB1MassOfGhoulsOracle, 5, 3},
		{"Nessian Courser", "Creature — Centaur Warrior", futB1NessianCourserOracle, 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.toughness)
			}
		})
	}
}

// Force of Savagery's printed toughness of 0 puts it into the graveyard
// as it enters (CR 704.5f).
func TestFut5594B1ForceOfSavageryDiesWithZeroToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastHoldCreature(t, g, "Force of Savagery", "Creature — Elemental", futB1ForceOfSavageryOracle, 8, 0)
	// The fixture has no printing behind it, so mark its 0 as printed
	// (ToughnessIsKnown, #691): a real Force of Savagery's is.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].PrintedPTKnown = true
			}
		}
	})
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("an 8/0 stayed on the battlefield")
	}
}

func TestFut5594B1SuspendCardsDeclareTheirPrintedNumbers(t *testing.T) {
	effectstest.S40SuspendAction(t, futB1KnightOfSursiOracle, 3, "{W}")
	effectstest.S40SuspendAction(t, futB1InfiltratorIlKorOracle, 2, "{1}{U}")
}

func TestFut5594B1WhipSpineDrakeHasMorph(t *testing.T) {
	spec, _ := effects.Lookup(futB1WhipSpineDrakeOracle)
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].FaceDown == nil {
		t.Fatalf("alternative costs %+v, want one morph", spec.AlternativeCosts)
	}
	ac := spec.AlternativeCosts[0]
	if ac.ManaCost != effects.MorphCastCost || ac.FaceDown.FaceUpCost != "{2}{W}" {
		t.Errorf("morph cast %s, face up %s; want %s and {2}{W}", ac.ManaCost, ac.FaceDown.FaceUpCost, effects.MorphCastCost)
	}
}

// Bloodthirst 1: the counter only when an opponent was dealt damage this
// turn.
func TestFut5594B1BogardanLancerBloodthirst(t *testing.T) {
	for _, tc := range []struct {
		name     string
		damaged  bool
		counters int
	}{
		{"no damage", false, 0},
		{"an opponent was dealt damage", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			if tc.damaged {
				pinger := futB1Bear(g, me.ID, "Pinger")
				g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, opp.ID, 1) })
			}
			id := effectstest.CastHoldCreature(t, g, "Bogardan Lancer", "Creature — Human Knight", futB1BogardanLancerOracle, 1, 1)
			c, ok := effectstest.BattlefieldCard(g, id)
			if !ok {
				t.Fatal("the Lancer did not resolve")
			}
			if got := c.Counters[game.CounterPlusOne]; got != tc.counters {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.counters)
			}
		})
	}
}

func TestFut5594B1CharRumblerFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.B12Push(g, me.ID, "Char-Rumbler", "Creature — Elemental", futB1CharRumblerOracle, -1, 3)
	for i := 0; i < 2; i++ {
		effectstest.ApaMana(me, "R")
		effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	}
	if got := effectstest.EffectivePower(t, g, id); got != 1 {
		t.Errorf("power %d after two activations, want 1", got)
	}
}

// Hellbent: shadow only while its controller's hand is empty.
func TestFut5594B1CutthroatIlDalHasShadowWhileHellbent(t *testing.T) {
	for _, tc := range []struct {
		hand   int
		shadow bool
	}{{0, true}, {2, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.D2dSetHand(me, tc.hand)
		id := effectstest.B12Push(g, me.ID, "Cutthroat il-Dal", "Creature — Human Rogue", futB1CutthroatIlDalOracle, 4, 1)
		if got := effectstest.HasEffectiveKeyword(t, g, id, "shadow"); got != tc.shadow {
			t.Errorf("with %d cards in hand: shadow %v, want %v", tc.hand, got, tc.shadow)
		}
	}
}

// --- the augurs: upkeep-only sacrifices -----------------------------------

// futB1AugurAtUpkeep seeds an augur for seat 0, proves it can't be
// activated in a main phase, and walks to seat 0's next upkeep.
func futB1AugurAtUpkeep(t *testing.T, g *game.Game, name, typeLine, oracle string, row int, params game.ActivateAbilityParams) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, name, typeLine, oracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	params.Strict = true
	if err := g.ActivateCatalogAbility(me.ID, id, row, params); err == nil {
		t.Fatalf("%s was activated in a main phase", name)
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	return id
}

func TestFut5594B1AugurIlVecGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := futB1AugurAtUpkeep(t, g, "Augur il-Vec", "Creature — Human Cleric", futB1AugurIlVecOracle, 0, game.ActivateAbilityParams{})
	life := me.Life
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Augur was not sacrificed")
	}
}

func TestFut5594B1AugurOfSkullsMakesATargetDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.D2dSetHand(opp, 3)
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}
	id := futB1AugurAtUpkeep(t, g, "Augur of Skulls", "Creature — Skeleton Wizard", futB1AugurOfSkullsOracle, 1, params)
	effectstest.ApaActivate(t, g, me, id, 1, params)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Errorf("the target owes %d discards, want 2", n)
	}
}

func TestFut5594B1AugurOfSkullsRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.B12Push(g, me.ID, "Augur of Skulls", "Creature — Skeleton Wizard", futB1AugurOfSkullsOracle, 1, 1)
	effectstest.ApaMana(me, "B", "C")
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	effectstest.B27Kill(g, id)
	if !g.Battlefield.Contains(id) {
		t.Error("the shielded Augur was destroyed")
	}
}

func TestFut5594B1AvenAugurBouncesUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := futB1Bear(g, opp.ID, "First")
	b := futB1Bear(g, opp.ID, "Second")
	params := game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)}
	id := futB1AugurAtUpkeep(t, g, "Aven Augur", "Creature — Bird Wizard", futB1AvenAugurOracle, 0, params)
	effectstest.ApaActivate(t, g, me, id, 0, params)
	for _, c := range []uuid.UUID{a, b} {
		if g.Battlefield.Contains(c) || !opp.Hand.Contains(c) {
			t.Error("a target did not return to its owner's hand")
		}
	}
}

func TestFut5594B1EmberwildeAugurDealsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}
	id := futB1AugurAtUpkeep(t, g, "Emberwilde Augur", "Creature — Goblin Shaman", futB1EmberwildeAugurOracle, 0, params)
	life := opp.Life
	effectstest.ApaActivate(t, g, me, id, 0, params)
	if opp.Life != life-3 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-3)
	}
}

func TestFut5594B1LlanowarAugurPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := futB1Bear(g, me.ID, "Bear")
	params := game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}
	id := futB1AugurAtUpkeep(t, g, "Llanowar Augur", "Creature — Elf Shaman", futB1LlanowarAugurOracle, 0, params)
	effectstest.ApaActivate(t, g, me, id, 0, params)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear has no trample")
	}
}

// --- activated abilities ---------------------------------------------------

func TestFut5594B1SarcomiteMyrFliesAndSacrificesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	myr := effectstest.B12Push(g, me.ID, "Sarcomite Myr", "Artifact Creature — Phyrexian Myr", futB1SarcomiteMyrOracle, 2, 1)
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, myr, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, myr, "flying") {
		t.Error("the Myr has no flying")
	}
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, myr, 1, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(myr) {
		t.Error("the Myr was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestFut5594B1WitchsMistDestroysOnlyADamagedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mist := effectstest.PushCatalogPermanent(g, me.ID, "Witch's Mist", "Enchantment", futB1WitchsMistOracle, false)
	pinger := futB1Bear(g, me.ID, "Pinger")
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh", futB1BearType, 3, 3)
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt", futB1BearType, 3, 3)
	effectstest.B27Damage(g, pinger, hurt, 1)
	g.RunStateChecksForTest()
	effectstest.ApaMana(me, "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, mist, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(fresh)}); err == nil {
		t.Fatal("Witch's Mist targeted a creature dealt no damage this turn")
	}
	effectstest.ApaActivate(t, g, me, mist, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(hurt)})
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature survived")
	}
	if !effectstest.HmlTapped(t, g, mist) {
		t.Error("Witch's Mist did not tap")
	}
}

func TestFut5594B1ScourgeOfKherRidgesSweeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	scourge := effectstest.B12Push(g, me.ID, "Scourge of Kher Ridges", "Creature — Dragon", futB1ScourgeOfKherRidgesOracle, 6, 6)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", futB1BearType, 5, 5)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 7, Toughness: 7, Keywords: []string{"flying"}})

	effectstest.ApaMana(me, "R", "C")
	effectstest.ApaActivate(t, g, me, scourge, 0, game.ActivateAbilityParams{})
	if got := effectstest.FindCardAnywhere(t, g, walker).DamageMarked; got != 2 {
		t.Errorf("the creature without flying has %d damage, want 2", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 0 {
		t.Errorf("the flyer has %d damage from the first sweep, want 0", got)
	}

	effectstest.ApaMana(me, "R", "C", "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, scourge, 1, game.ActivateAbilityParams{})
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 6 {
		t.Errorf("the flyer has %d damage, want 6", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, scourge).DamageMarked; got != 0 {
		t.Errorf("the Scourge dealt %d damage to itself", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, walker).DamageMarked; got != 2 {
		t.Errorf("the second sweep hit a creature without flying: %d damage", got)
	}
}

func TestFut5594B1BloodshotTraineeNeedsPowerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	small := effectstest.B12Push(g, me.ID, "Bloodshot Trainee", "Creature — Goblin Warrior", futB1BloodshotTraineeOracle, 2, 3)
	big := effectstest.B12Push(g, me.ID, "Bloodshot Trainee", "Creature — Goblin Warrior", futB1BloodshotTraineeOracle, 4, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", futB1BearType, 5, 5)
	if err := g.ActivateCatalogAbility(me.ID, small, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Fatal("a Trainee with power 2 was activated")
	}
	effectstest.ApaActivate(t, g, me, big, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim)})
	if got := effectstest.FindCardAnywhere(t, g, victim).DamageMarked; got != 4 {
		t.Errorf("the target has %d damage, want 4", got)
	}
}

func TestFut5594B1BondedFetchLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	fetch := effectstest.B12Push(g, me.ID, "Bonded Fetch", "Creature — Homunculus", futB1BondedFetchOracle, 0, 2)
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, fetch, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d before the discard, want %d", got, hand+1)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

// --- lands -----------------------------------------------------------------

func TestFut5594B1LandsEnterTapped(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Keldon Megaliths", futB1KeldonMegalithsOracle},
		{"Tolaria West", futB1TolariaWestOracle},
		{"New Benalia", futB1NewBenaliaOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s is not on the battlefield", tc.name)
			}
			if !effectstest.HmlTapped(t, g, id) {
				t.Errorf("%s entered untapped", tc.name)
			}
		})
	}
}

func TestFut5594B1NewBenaliaScriesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12PlayFromHand(t, g, "New Benalia", "Land", futB1NewBenaliaOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scrying %d cards, want 1", len(c.ScryCards))
	}
}

func TestFut5594B1KeldonMegalithsPingsOnlyWhileHellbent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Keldon Megaliths", "Land", futB1KeldonMegalithsOracle, false)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}
	effectstest.D2dSetHand(me, 1)
	effectstest.ApaMana(me, "R", "C")
	strict := params
	strict.Strict = true
	if err := g.ActivateCatalogAbility(me.ID, land, 0, strict); err == nil {
		t.Fatal("the ping was activated with a card in hand")
	}
	effectstest.D2dSetHand(me, 0)
	life := opp.Life
	effectstest.ApaActivate(t, g, me, land, 0, params)
	if opp.Life != life-1 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-1)
	}
}

func TestFut5594B1TolariaWestTransmutesFromHand(t *testing.T) {
	var transmute *game.ActivatedAbilityShape
	rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: futB1TolariaWestOracle})
	for i := range rows {
		if rows[i].Cost.DiscardSelf {
			transmute = &rows[i]
		}
	}
	if transmute == nil {
		t.Fatal("Tolaria West has no transmute row")
	}
	if !transmute.SorcerySpeed || transmute.Cost.Mana != "{1}{U}{U}" {
		t.Errorf("transmute row %+v, want {1}{U}{U} at sorcery speed", transmute)
	}
	inHand := false
	for _, z := range transmute.Zones {
		inHand = inHand || z == game.ZoneHand
	}
	if !inHand {
		t.Error("transmute does not work from the hand")
	}
}

// --- spells ----------------------------------------------------------------

func TestFut5594B1ForeseeScriesFourThenDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Foresee", "Sorcery", futB1ForeseeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Fatalf("scrying %d cards, want 4", len(c.ScryCards))
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("drew before the scry was answered: hand %d, want %d", got, hand)
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
}

func TestFut5594B1SlaughterPactDestroysAndIsOwedNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	black := futB1Bear(g, opp.ID, "Black")
	effectstest.HmlSetColors(g, black, "B")
	green := futB1Bear(g, opp.ID, "Green")
	effectstest.HmlSetColors(g, green, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Slaughter Pact", "Instant", futB1SlaughterPactOracle, effectstest.CardTarget(black)); err == nil {
		t.Fatal("Slaughter Pact targeted a black creature")
	}
	effectstest.CastCatalogSpell(t, g, "Slaughter Pact", "Instant", futB1SlaughterPactOracle, effectstest.CardTarget(green))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Fatal("the nonblack creature survived")
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if !me.Eliminated {
		t.Error("an unpaid Pact did not lose its controller the game")
	}
}

func TestFut5594B1PactOfTheTitanMakesAGiantAndIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Pact of the Titan", "Instant", futB1PactOfTheTitanOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	giants := effectstest.TokensNamed(g, "Giant")
	if len(giants) != 1 {
		t.Fatalf("%d Giant tokens, want 1", len(giants))
	}
	if gi := giants[0]; gi.Controller != me.ID || gi.Power != 4 || gi.Toughness != 4 || !gi.HasColorWord("R") {
		t.Errorf("token %+v, want a red 4/4 Giant of yours", gi)
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{R}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if me.Eliminated {
		t.Error("paying {4}{R} still lost the game")
	}
}

func TestFut5594B1MinionsMurmursCountsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	futB1Bear(g, me.ID, "One")
	futB1Bear(g, me.ID, "Two")
	futB1Bear(g, opp.ID, "Theirs")
	hand, life := me.Hand.Size(), me.Life
	effectstest.CastCatalogSpell(t, g, "Minions' Murmurs", "Sorcery", futB1MinionsMurmursOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestFut5594B1EvenTheOddsNeedsFewerCreaturesThanEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for _, p := range g.Seats[1:] {
		futB1Bear(g, p.ID, "Theirs")
	}
	mine := futB1Bear(g, me.ID, "Mine")
	if err := effectstest.CastCatalogSpellErr(t, g, "Even the Odds", "Instant", futB1EvenTheOddsOracle, nil); err == nil {
		t.Fatal("cast while controlling as many creatures as an opponent")
	}
	effectstest.B27Kill(g, mine)
	g.RunStateChecksForTest()
	effectstest.CastCatalogSpell(t, g, "Even the Odds", "Instant", futB1EvenTheOddsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range effectstest.TokensNamed(g, "Soldier") {
		if c.Controller == me.ID {
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d Soldier tokens, want 3", n)
	}
}

func TestFut5594B1MysticSpeculationScriesThreeAndBuysBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id, err := effectstest.CastWithOptionalCosts(t, g, "Mystic Speculation", "Sorcery", futB1MysticSpeculationOracle, nil, []int{0}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 3 {
		t.Fatalf("scrying %d cards, want 3", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Error("the bought-back spell did not return to hand")
	}
}

func TestFut5594B1MarshalingCryPumpsYourTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := futB1Bear(g, me.ID, "Mine")
	theirs := futB1Bear(g, opp.ID, "Theirs")
	effectstest.CastCatalogSpell(t, g, "Marshaling Cry", "Sorcery", futB1MarshalingCryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("your creature has no vigilance")
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestFut5594B1MarshalingCryFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := futB1Bear(g, me.ID, "Mine")
	id := effectstest.SeedGraveyardCard(t, g, "Marshaling Cry", "Sorcery", futB1MarshalingCryOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("your creature has power %d, want 3", p)
	}
	if !g.Exile.Contains(id) {
		t.Error("a flashback spell is exiled as it resolves (CR 702.34a)")
	}
}

// --- triggers --------------------------------------------------------------

func TestFut5594B1RiftsweeperShufflesAnExiledCardIntoItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	exiled := effectstest.PushExiledCard(g, opp.ID, "Exiled Land", "Land", false)
	effectstest.PushExiledCard(g, opp.ID, "Other Exiled Card", "Sorcery", false)
	library := opp.Library.Size()
	effectstest.CastHoldCreature(t, g, "Riftsweeper", "Creature — Elf Shaman", futB1RiftsweeperOracle, 2, 2)
	effectstest.AnswerPickTarget(t, g, exiled)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(exiled) {
		t.Error("the card is still in exile")
	}
	if !opp.Library.Contains(exiled) || opp.Library.Size() != library+1 {
		t.Error("the card did not go into its owner's library")
	}
}

func TestFut5594B1StrongholdRatsMakesEachPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.B12Push(g, me.ID, "Stronghold Rats", "Creature — Rat", futB1StrongholdRatsOracle, 2, 1)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		effectstest.D2dSetHand(p, 2)
		before[p.ID] = p.Graveyard.Size()
	}
	effectstest.DealCombatDamage(g, me.ID, rats, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	futB1AnswerEveryDiscard(t, g)
	for _, p := range g.Seats {
		if got := p.Graveyard.Size(); got != before[p.ID]+1 {
			t.Errorf("%s's graveyard %d -> %d, want one discarded card", p.Name, before[p.ID], got)
		}
	}
}

// --- auras -----------------------------------------------------------------

func TestFut5594B1BoundInSilenceStopsAttacksAndBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushRestrictionBear(g, me, "Grizzly Bears")
	effectstest.CastCatalogSpell(t, g, "Bound in Silence", "Kindred Enchantment — Rebel Aura", futB1BoundInSilenceOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, bear, game.CantAttackOrBlock)
}

func TestFut5594B1GiftOfGraniteIsAFlashToughnessAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := futB1Bear(g, me.ID, "Bear")
	id := effectstest.B13HandCard(me, "Gift of Granite", futB1AuraType, futB1GiftOfGraniteOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("Gift of Granite at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 2/4", p, tough)
	}
}

func TestFut5594B1EmblemOfTheWarmindHastesYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	host := futB1Bear(g, me.ID, "Host")
	theirs := futB1Bear(g, opp.ID, "Theirs")
	if err := effectstest.CastCatalogSpellErr(t, g, "Emblem of the Warmind", futB1AuraType, futB1EmblemOfTheWarmindOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("the Aura enchanted a creature you don't control")
	}
	effectstest.CastCatalogSpell(t, g, "Emblem of the Warmind", futB1AuraType, futB1EmblemOfTheWarmindOracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	other := futB1Bear(g, me.ID, "Other")
	if !effectstest.HasEffectiveKeyword(t, g, other, "haste") {
		t.Error("another creature of yours has no haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}
