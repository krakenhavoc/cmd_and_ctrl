package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hou_5534_b1_cards_test.go — a first batch of Hour of Devastation cards
// (#5534). Every test plays the real catalog entry through the engine.

const (
	houDefiantKhenraOracle         = "0e6fe04f-b20d-4df9-a7ac-edebc3bcc9be"
	houDutifulServantsOracle       = "5837b3d9-b3cb-4333-b0db-ed38e49c997f"
	houHarrierNagaOracle           = "4a1cbac1-ee22-44cf-a37f-a58055f5c9bc"
	houBitterbowSharpshootersOrac  = "76f22670-6806-4124-8bbb-822ed33d2bda"
	houBrambleweftBehemothOracle   = "f5070528-cc4c-4910-8047-1ed17efd6496"
	houCarrionScreecherOracle      = "7b813e46-34f6-4918-97a7-a4b2a2464e86"
	houScroungerOfSoulsOracle      = "23dd6d4a-d933-4328-8d09-75f7908ae347"
	houAngelOfTheGodPharaohOracle  = "701912d5-9995-4ac7-9858-188f78b2b8a3"
	houGraniticTitanOracle         = "d398868b-7574-40d5-8915-6393ee489067"
	houLurchingRotbeastOracle      = "771be8ad-5ab5-40b1-ad9b-e5e834949d32"
	houMoaningWallOracle           = "7bf22390-a6c0-4da6-9f05-e66628f72079"
	houRampagingHippoOracle        = "e48aa13b-7488-4644-9080-6c342a8ac463"
	houStripedRiverwinderOracle    = "725972de-4816-44cf-a2a5-4326c3bc1b4b"
	houAdornedPouncerOracle        = "80b8ee3d-4174-4f47-a533-a3db1e703382"
	houSteadfastSentinelOracle     = "38bfc0bc-66a6-4d9e-9186-5ccaef773014"
	houProvenCombatantOracle       = "24ce3dbe-2b61-43a6-ab16-32b37e385127"
	houKhenraEternalOracle         = "91b46b22-8062-434b-8432-5ed1062ccaf9"
	houSpellweaverEternalOracle    = "71f3d3d3-3f82-43f6-b2a9-d21afbdd7739"
	houManticoreEternalOracle      = "d859e9b3-7804-418d-be95-e03a5de290ca"
	houFrontlineDevastatorOracle   = "5ace15b8-f853-4c44-b9c7-d307367b612c"
	houMercilessEternalOracle      = "55c419e9-43f2-4d2d-8c13-057a48d47b53"
	houBurningFistMinotaurOracle   = "5460c76f-39dc-4db0-acf7-146d533d822b"
	houFrilledSandwallaOracle      = "3315aa42-d3ae-425e-9c12-191d26e67a0a"
	houDevoteeOfStrengthOracle     = "a6d8f944-f516-49d7-9aa8-77b98e862d6a"
	houRiverHoopoeOracle           = "058bbc0a-9c09-4c98-8735-83255a71bbca"
	houSeerOfTheLastTomorrowOracle = "dc5cb189-9eff-4b4e-a46d-7a8e62529c89"
	houThornedMolochOracle         = "d2a49109-6db8-4f08-828c-da7a456a78a7"
	houAvenOfEnduringHopeOracle    = "e8291185-9d87-4aa2-b7ba-d4946e5a082f"
	houFeralProwlerOracle          = "5bebba2d-867f-4ed1-a308-904fac32c8b6"
	houRuinRatOracle               = "f77012cd-170f-40df-816b-330bf85e809d"
	houDisposalMummyOracle         = "002a9cea-8cf7-48ba-83eb-e1c87a7024e5"
	houChaosMawOracle              = "04f0c489-3d1e-4ee4-a0ab-7983ae643ecb"
	houDauntlessAvenOracle         = "7104e471-06aa-426d-af1e-d5f0b1fb18c3"
	houGodPharaohsFaithfulOracle   = "3f86e6c6-fca5-43e9-88cc-4546cdd5f0ec"
	houMummyParamountOracle        = "99cc909d-3cb2-499c-8231-41418745cd71"
	houGrislySurvivorOracle        = "8fe1e619-f827-43f5-acaf-3b79aa671d1d"
	houOminousSphinxOracle         = "2cb968c4-5b50-4242-8db7-dd1fa9ce4948"
	houAmbuscadeOracle             = "5e419975-e054-4c35-8845-a08589ebb08f"
	houBeneathTheSandsOracle       = "27547519-9474-49fa-80e7-e604d61cb3ca"
	houCountervailingWindsOracle   = "2c51e20b-07b2-4c96-8d73-87379aafbc7c"
	houDjerusRenunciationOracle    = "11b21086-0779-4d16-8a56-252c6bee9bed"
	houRazakethsRiteOracle         = "0ab0be8e-d6be-411b-827f-584bb5d03766"
	houWithoutWeaknessOracle       = "fffa9334-3576-4f70-9605-2ad062cdbc69"
	houInfernoJetOracle            = "50c13b38-9865-4c60-b2df-3fd6e3a9c6a2"
	houLifeGoesOnOracle            = "15d2176b-ab3d-4738-bcbc-bcebf891d16e"
	houBlurOfBladesOracle          = "65410f7a-c749-4b23-ad61-8a7136efcad2"

	houBearType = "Creature — Bear"
)

// houCards is every card of the batch, for the registry check.
var houCards = []string{
	houDefiantKhenraOracle, houDutifulServantsOracle, houHarrierNagaOracle, houBitterbowSharpshootersOrac,
	houBrambleweftBehemothOracle, houCarrionScreecherOracle, houScroungerOfSoulsOracle, houAngelOfTheGodPharaohOracle,
	houGraniticTitanOracle, houLurchingRotbeastOracle, houMoaningWallOracle, houRampagingHippoOracle,
	houStripedRiverwinderOracle, houAdornedPouncerOracle, houSteadfastSentinelOracle, houProvenCombatantOracle,
	houKhenraEternalOracle, houSpellweaverEternalOracle, houManticoreEternalOracle, houFrontlineDevastatorOracle,
	houMercilessEternalOracle, houBurningFistMinotaurOracle, houFrilledSandwallaOracle, houDevoteeOfStrengthOracle,
	houRiverHoopoeOracle, houSeerOfTheLastTomorrowOracle, houThornedMolochOracle, houAvenOfEnduringHopeOracle,
	houFeralProwlerOracle, houRuinRatOracle, houDisposalMummyOracle, houChaosMawOracle,
	houDauntlessAvenOracle, houGodPharaohsFaithfulOracle, houMummyParamountOracle, houGrislySurvivorOracle,
	houOminousSphinxOracle, houAmbuscadeOracle, houBeneathTheSandsOracle, houCountervailingWindsOracle,
	houDjerusRenunciationOracle, houRazakethsRiteOracle, houWithoutWeaknessOracle, houInfernoJetOracle,
	houLifeGoesOnOracle, houBlurOfBladesOracle,
}

// houUnsick clears summoning sickness on a pushed permanent, so it can
// attack this turn.
func houUnsick(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, id); c != nil {
			c.SummonedThisTurn = false
		}
	})
}

// houToken is the token named `name` that `owner` controls, or nil.
func houToken(g *game.Game, owner uuid.UUID, name string) *game.Card {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == name && c.Controller == owner && c.IsToken() {
			return c
		}
	}
	return nil
}

// houFodder puts a plain card in p's hand to pay a discard cost.
func houFodder(p *game.Player) uuid.UUID {
	return handCardFull(p, "Fodder", "Sorcery", "{1}", "", nil)
}

// --- every card is in the catalog ----------------------------------------

func TestHou5534EveryCardIsRegisteredFull(t *testing.T) {
	if len(houCards) != 46 {
		t.Fatalf("the batch lists %d cards, want 46", len(houCards))
	}
	for _, oracle := range houCards {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords ------------------------------------------------------

func TestHou5534KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Bitterbow Sharpshooters", "Creature — Jackal Archer", houBitterbowSharpshootersOrac, []string{"reach", "vigilance"}},
		{"Brambleweft Behemoth", "Creature — Elemental", houBrambleweftBehemothOracle, []string{"trample"}},
		{"Carrion Screecher", "Creature — Zombie Bird", houCarrionScreecherOracle, []string{"flying"}},
		{"Scrounger of Souls", "Creature — Horror", houScroungerOfSoulsOracle, []string{"lifelink"}},
		{"Angel of the God-Pharaoh", "Creature — Angel", houAngelOfTheGodPharaohOracle, []string{"flying"}},
		{"Granitic Titan", "Creature — Elemental", houGraniticTitanOracle, []string{"menace"}},
		{"Moaning Wall", "Creature — Zombie Wall", houMoaningWallOracle, []string{"defender"}},
		{"Rampaging Hippo", "Creature — Hippo", houRampagingHippoOracle, []string{"trample"}},
		{"Striped Riverwinder", "Creature — Serpent", houStripedRiverwinderOracle, []string{"hexproof"}},
		{"Adorned Pouncer", "Creature — Cat", houAdornedPouncerOracle, []string{"double strike"}},
		{"Steadfast Sentinel", "Creature — Human Cleric", houSteadfastSentinelOracle, []string{"vigilance"}},
		{"Spellweaver Eternal", "Creature — Zombie Snake Wizard", houSpellweaverEternalOracle, []string{"prowess"}},
		{"Burning-Fist Minotaur", "Creature — Minotaur Wizard", houBurningFistMinotaurOracle, []string{"first strike"}},
		{"River Hoopoe", "Creature — Bird", houRiverHoopoeOracle, []string{"flying"}},
		{"Thorned Moloch", "Creature — Lizard", houThornedMolochOracle, []string{"prowess"}},
		{"Aven of Enduring Hope", "Creature — Bird Cleric", houAvenOfEnduringHopeOracle, []string{"flying"}},
		{"Ruin Rat", "Creature — Rat", houRuinRatOracle, []string{"deathtouch"}},
		{"Dauntless Aven", "Creature — Bird Warrior", houDauntlessAvenOracle, []string{"flying"}},
		{"Ominous Sphinx", "Creature — Sphinx", houOminousSphinxOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// The vanilla bodies are registered and bring nothing with them.
func TestHou5534VanillaBodiesHaveNoAbilities(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Defiant Khenra", "Creature — Jackal Warrior", houDefiantKhenraOracle},
		{"Dutiful Servants", "Creature — Zombie", houDutifulServantsOracle},
		{"Harrier Naga", "Creature — Snake Warrior", houHarrierNagaOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := Lookup(tc.oracle)
			if !ok {
				t.Fatalf("%s is not registered", tc.name)
			}
			if len(spec.PrintedKeywords) != 0 || len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 {
				t.Errorf("%s registers abilities it doesn't print", tc.name)
			}
			g := newCatalogGame(t)
			advanceToMain(t, g)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !g.Battlefield.Contains(id) {
				t.Errorf("%s did not resolve", tc.name)
			}
		})
	}
}

// Thorned Moloch has first strike only while it is attacking.
func TestHou5534ThornedMolochHasFirstStrikeWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	moloch := b12Push(g, me.ID, "Thorned Moloch", "Creature — Lizard", houThornedMolochOracle, 2, 2)
	if hasEffectiveKeyword(t, g, moloch, "first strike") {
		t.Fatal("it has first strike at home")
	}
	brAttack(t, g, moloch)
	if !hasEffectiveKeyword(t, g, moloch, "first strike") {
		t.Error("it has no first strike while attacking")
	}
}

// --- cycling -----------------------------------------------------------------

func TestHou5534CyclingCardsCycle(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana string }{
		{"Angel of the God-Pharaoh", "Creature — Angel", houAngelOfTheGodPharaohOracle, "{C}{C}"},
		{"Granitic Titan", "Creature — Elemental", houGraniticTitanOracle, "{C}{C}"},
		{"Lurching Rotbeast", "Creature — Zombie Beast", houLurchingRotbeastOracle, "{B}"},
		{"Moaning Wall", "Creature — Zombie Wall", houMoaningWallOracle, "{C}{C}"},
		{"Rampaging Hippo", "Creature — Hippo", houRampagingHippoOracle, "{C}{C}"},
		{"Striped Riverwinder", "Creature — Serpent", houStripedRiverwinderOracle, "{U}"},
		{"Beneath the Sands", "Sorcery", houBeneathTheSandsOracle, "{C}{C}"},
		{"Countervailing Winds", "Instant", houCountervailingWindsOracle, "{C}{C}"},
		{"Djeru's Renunciation", "Instant", houDjerusRenunciationOracle, "{W}"},
		{"Razaketh's Rite", "Sorcery", houRazakethsRiteOracle, "{B}"},
		{"Without Weakness", "Instant", houWithoutWeaknessOracle, "{C}{C}"},
		{"Inferno Jet", "Sorcery", houInfernoJetOracle, "{C}{C}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
			ons1Cycle(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
		})
	}
}

// --- eternalize --------------------------------------------------------------

func TestHou5534EternalizeMakesAFourFourBlackZombieCopy(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		power, toughness             int
		keyword                      string
	}{
		{"Adorned Pouncer", "Creature — Cat", houAdornedPouncerOracle, "{C}{C}{C}{W}{W}", 1, 1, "double strike"},
		{"Steadfast Sentinel", "Creature — Human Cleric", houSteadfastSentinelOracle, "{C}{C}{C}{C}{W}{W}", 2, 3, "vigilance"},
		{"Proven Combatant", "Creature — Human Warrior", houProvenCombatantOracle, "{C}{C}{C}{C}{U}{U}", 1, 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id, me := activateFromGraveyard(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, tc.mana, game.ActivateAbilityParams{})
			if !g.Exile.Contains(id) {
				t.Fatal("the card was not exiled to pay the cost")
			}
			passPriorityAroundTable(t, g)
			token := houToken(g, me.ID, tc.name)
			if token == nil {
				t.Fatal("no eternalize token was created")
			}
			if p, tough := ptOf(t, g, token.InstanceID); p != 4 || tough != 4 {
				t.Errorf("the token is %d/%d, want 4/4", p, tough)
			}
			if !token.HasSubtype("Zombie") {
				t.Errorf("the token's type line %q has no Zombie", token.TypeLine)
			}
			if len(token.Colors) != 1 || token.Colors[0] != "B" {
				t.Errorf("token colours %v, want [B]", token.Colors)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, token.InstanceID, tc.keyword) {
				t.Errorf("the token lost %s", tc.keyword)
			}
		})
	}
}

// --- afflict -----------------------------------------------------------------

func TestHou5534AfflictWhenBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Khenra Eternal", "Creature — Zombie Jackal Warrior", houKhenraEternalOracle, 1},
		{"Spellweaver Eternal", "Creature — Zombie Snake Wizard", houSpellweaverEternalOracle, 2},
		{"Frontline Devastator", "Creature — Zombie Minotaur Warrior", houFrontlineDevastatorOracle, 2},
		{"Merciless Eternal", "Creature — Zombie Cleric", houMercilessEternalOracle, 2},
		{"Manticore Eternal", "Creature — Zombie Manticore", houManticoreEternalOracle, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
			life := opp.Life
			brAttack(t, g, attacker)
			if err := g.DeclareBlocker(wall, attacker); err != nil {
				t.Fatalf("block: %v", err)
			}
			lockInBlocks(t, g)
			passPriorityAroundTable(t, g)
			if opp.Life != life-tc.n {
				t.Errorf("the defending player is at %d, want %d", opp.Life, life-tc.n)
			}
		})
	}
}

func TestHou5534ManticoreEternalMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	manticore := b12Push(g, me.ID, "Manticore Eternal", "Creature — Zombie Manticore", houManticoreEternalOracle, 5, 4)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != manticore {
		t.Fatalf("pass with the Manticore home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(manticore, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Manticore attacking: %v", err)
	}
}

// --- activated abilities ---------------------------------------------------

func TestHou5534FrontlineDevastatorPumpsPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := b12Push(g, me.ID, "Frontline Devastator", "Creature — Zombie Minotaur Warrior", houFrontlineDevastatorOracle, 3, 3)
	dcMana(t, g, me.ID, "{C}{R}")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, id); p != 4 || tough != 3 {
		t.Errorf("it is %d/%d, want 4/3", p, tough)
	}
}

func TestHou5534DiscardPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		power, toughness             int
		wantP, wantT                 int
	}{
		{"Merciless Eternal", "Creature — Zombie Cleric", houMercilessEternalOracle, "{C}{C}{B}", 2, 2, 4, 4},
		{"Burning-Fist Minotaur", "Creature — Minotaur Wizard", houBurningFistMinotaurOracle, "{C}{R}", 2, 1, 4, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if err := mmqB2Shape(t, g, me, id, tc.mana, uuid.Nil, nil); err == nil {
				t.Fatal("activated without discarding a card")
			}
			fodder := houFodder(me)
			if err := mmqB2Shape(t, g, me, id, "", fodder, nil); err != nil {
				t.Fatalf("activation: %v", err)
			}
			if !me.Graveyard.Contains(fodder) {
				t.Fatal("the card was not discarded as the cost")
			}
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("it is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestHou5534FrilledSandwallaPumpsOnceATurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := b12Push(g, me.ID, "Frilled Sandwalla", "Creature — Lizard", houFrilledSandwallaOracle, 1, 1)
	dcMana(t, g, me.ID, "{C}{G}")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, id); p != 3 || tough != 3 {
		t.Fatalf("it is %d/%d, want 3/3", p, tough)
	}
	dcMana(t, g, me.ID, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the pump was activated a second time this turn")
	}
}

func TestHou5534DevoteeOfStrengthPumpsTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	devotee := b12Push(g, me.ID, "Devotee of Strength", "Creature — Snake Wizard", houDevoteeOfStrengthOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", houBearType, 2, 2)
	dcMana(t, g, me.ID, "{C}{C}{C}{C}{G}")
	apaActivate(t, g, me, devotee, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestHou5534RiverHoopoeGainsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	seedLibrary(me, "A", "B")
	hoopoe := b12Push(g, me.ID, "River Hoopoe", "Creature — Bird", houRiverHoopoeOracle, 1, 3)
	life, hand := me.Life, me.Hand.Size()
	dcMana(t, g, me.ID, "{C}{C}{C}{G}{U}")
	apaActivate(t, g, me, hoopoe, 0, game.ActivateAbilityParams{})
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestHou5534SeerOfTheLastTomorrowMillsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	seedLibrary(opp, "A", "B", "C", "D")
	seer := b12Push(g, me.ID, "Seer of the Last Tomorrow", "Creature — Snake Cleric", houSeerOfTheLastTomorrowOracle, 1, 4)
	library := opp.Library.Size()
	mmqB2Discards(t, g, me, seer, "{U}", []game.TargetRef{playerTarget(opp.ID)})
	if got := library - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
}

// --- triggered abilities ---------------------------------------------------

func TestHou5534AvenOfEnduringHopeGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Aven of Enduring Hope", "Creature — Bird Cleric", houAvenOfEnduringHopeOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestHou5534FeralProwlerDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	prowler := b12Push(g, me.ID, "Feral Prowler", "Creature — Cat", houFeralProwlerOracle, 1, 3)
	hand := me.Hand.Size()
	b27Kill(g, prowler)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestHou5534RuinRatExilesAnOpponentsCardWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := b12Push(g, me.ID, "Ruin Rat", "Creature — Rat", houRuinRatOracle, 1, 1)
	theirs := s82b10Grave(opp, "Their Card", "Sorcery", "{1}")
	mine := s82b10Grave(me, "My Card", "Sorcery", "{1}")
	s82b10Kill(g, rat)
	s82b10Settle(t, g, me.ID, s82b10Target(theirs))
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
}

func TestHou5534DisposalMummyExilesAnOpponentsCardWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := s82b10Grave(opp, "Their Card", "Sorcery", "{1}")
	mine := s82b10Grave(me, "My Card", "Sorcery", "{1}")
	castHoldCreature(t, g, "Disposal Mummy", "Creature — Zombie Jackal", houDisposalMummyOracle, 2, 3)
	s82b10Settle(t, g, me.ID, s82b10Target(theirs))
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
}

func TestHou5534ChaosMawDamagesEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", houBearType, 3, 3)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	maw := castHoldCreature(t, g, "Chaos Maw", "Creature — Hellion", houChaosMawOracle, 6, 6)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("your 3/3 survived 3 damage")
	}
	if got := findCardAnywhere(t, g, big).DamageMarked; got != 3 {
		t.Errorf("the 4/4 has %d damage, want 3", got)
	}
	if got := findCardAnywhere(t, g, maw).DamageMarked; got != 0 {
		t.Errorf("Chaos Maw dealt itself %d damage", got)
	}
}

func TestHou5534DauntlessAvenUntapsACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aven := b12Push(g, me.ID, "Dauntless Aven", "Creature — Bird Warrior", houDauntlessAvenOracle, 2, 1)
	bear := b12Creature(g, me.ID, "Bear", houBearType, 2, 2)
	houUnsick(g, aven)
	tapForTest(t, g, bear)
	advanceTo(t, g, game.StepDeclareAttackers)
	declareAttack(t, g, opp.ID, aven)
	lockInAttacks(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Error("the Bear is still tapped")
	}
}

func TestHou5534GodPharaohsFaithfulGainsOnBlueBlackOrRed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	b12Push(g, me.ID, "God-Pharaoh's Faithful", "Creature — Human Wizard", houGodPharaohsFaithfulOracle, 0, 4)
	for _, tc := range []struct {
		color string
		gain  int
	}{{"U", 1}, {"B", 1}, {"R", 1}, {"G", 0}} {
		life := me.Life
		spell := handCardFull(me, "Spell", "Sorcery", "{"+tc.color+"}", "", []string{tc.color})
		dcMana(t, g, me.ID, "{"+tc.color+"}")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast the %s spell: %v", tc.color, err)
		}
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
		if me.Life != life+tc.gain {
			t.Errorf("a %s spell: life %d, want %d", tc.color, me.Life, life+tc.gain)
		}
	}
}

func TestHou5534MummyParamountGrowsWhenAZombieEnters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	paramount := b12Push(g, me.ID, "Mummy Paramount", "Creature — Zombie", houMummyParamountOracle, 2, 2)
	castHoldCreature(t, g, "Bear", houBearType, "", 2, 2)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, paramount); p != 2 || tough != 2 {
		t.Fatalf("a Bear entering made it %d/%d, want 2/2", p, tough)
	}
	castHoldCreature(t, g, "Walker", "Creature — Zombie", "", 2, 2)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, paramount); p != 3 || tough != 3 {
		t.Errorf("a Zombie entering made it %d/%d, want 3/3", p, tough)
	}
}

func TestHou5534GrislySurvivorGrowsOnADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	survivor := b12Push(g, me.ID, "Grisly Survivor", "Creature — Minotaur Warrior", houGrislySurvivorOracle, 2, 3)
	s58HandCards(me, 2)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, survivor); p != 4 || tough != 3 {
		t.Errorf("it is %d/%d, want 4/3", p, tough)
	}
}

func TestHou5534OminousSphinxShrinksAnOpponentsCreatureOnADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ominous Sphinx", "Creature — Sphinx", houOminousSphinxOracle, 4, 4)
	bear := b12Creature(g, opp.ID, "Their Bear", houBearType, 3, 3)
	s58HandCards(me, 2)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	g.RunStateChecksForTest()
	s82b10Settle(t, g, me.ID, s82b10Target(bear))
	if p, tough := ptOf(t, g, bear); p != 1 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 1/3", p, tough)
	}
}

// --- the spells ----------------------------------------------------------

func TestHou5534AmbuscadeBoostsThenBites(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "My Bear", houBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Ambuscade", "Instant", houAmbuscadeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, mine); p != 3 {
		t.Errorf("your creature has power %d, want 3", p)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 3/3 survived 3 damage")
	}
	if got := findCardAnywhere(t, g, mine).DamageMarked; got != 0 {
		t.Errorf("your creature was dealt %d damage; the bite is one-sided", got)
	}
}

func TestHou5534BeneathTheSandsFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: houBearType},
	)
	castCatalogSpell(t, g, "Beneath the Sands", "Sorcery", houBeneathTheSandsOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Forest")
	land, ok := visBattlefieldNamed(g, me.ID, "Forest")
	if !ok {
		t.Fatal("the basic land is not on the battlefield")
	}
	if !land.Tapped {
		t.Error("the land entered untapped")
	}
}

func TestHou5534CountervailingWindsTaxesByYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	for i := 0; i < 2; i++ {
		pushGraveyardCardTyped(me, "Old Card", "Sorcery")
	}
	for i := 0; i < 4; i++ {
		pushGraveyardCardTyped(opp, "Their Card", "Sorcery")
	}
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	castCatalogSpell(t, g, "Countervailing Winds", "Instant", houCountervailingWindsOracle, cardTarget(victim))
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatalf("no pay-unless prompt: %+v", g.PendingChoices)
	}
	if c.PayCost != "{2}" {
		t.Errorf("the prompt asks for %q, want {2}: one for each card in the caster's graveyard", c.PayCost)
	}
}

func TestHou5534DjerusRenunciationTapsUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	a := b12Creature(g, opp.ID, "A", houBearType, 2, 2)
	b := b12Creature(g, opp.ID, "B", houBearType, 2, 2)
	castCatalogSpell(t, g, "Djeru's Renunciation", "Instant", houDjerusRenunciationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	if !isTapped(g, a) || !isTapped(g, b) {
		t.Error("both targets should be tapped")
	}
}

func TestHou5534RazakethsRiteTutorsAnyCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: houBearType},
		game.Card{Name: "Wanted Spell", TypeLine: "Sorcery"},
	)
	castCatalogSpell(t, g, "Razaketh's Rite", "Sorcery", houRazakethsRiteOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Wanted Spell")
	if !b02bHandHasNamed(me, "Wanted Spell") {
		t.Error("the chosen card did not reach the hand")
	}
}

func TestHou5534WithoutWeaknessGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "My Bear", houBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", houBearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Without Weakness", "Instant", houWithoutWeaknessOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Without Weakness targeted a creature you don't control")
	}
	castCatalogSpell(t, g, "Without Weakness", "Instant", houWithoutWeaknessOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Fatal("the creature has no indestructible")
	}
	b27Kill(g, mine)
	if !g.Battlefield.Contains(mine) {
		t.Error("the indestructible creature was destroyed")
	}
}

func TestHou5534InfernoJetHitsAnOpponentForSix(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	if err := castCatalogSpellErr(t, g, "Inferno Jet", "Sorcery", houInfernoJetOracle, []game.TargetRef{playerTarget(me.ID)}); err == nil {
		t.Fatal("Inferno Jet targeted its own caster")
	}
	life := opp.Life
	castCatalogSpell(t, g, "Inferno Jet", "Sorcery", houInfernoJetOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-6)
	}
}

func TestHou5534LifeGoesOnGainsFourOrEight(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	life := me.Life
	castCatalogSpell(t, g, "Life Goes On", "Instant", houLifeGoesOnOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("no creature died: life %d, want %d", me.Life, life+4)
	}
	bear := b12Creature(g, me.ID, "Bear", houBearType, 2, 2)
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	life = me.Life
	castCatalogSpell(t, g, "Life Goes On", "Instant", houLifeGoesOnOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Errorf("a creature died: life %d, want %d", me.Life, life+8)
	}
}

func TestHou5534BlurOfBladesShrinksAndHitsTheController(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Their Bear", houBearType, 2, 2)
	life := opp.Life
	castCatalogSpell(t, g, "Blur of Blades", "Instant", houBlurOfBladesOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterMinusOne); got != 1 {
		t.Errorf("the Bear has %d -1/-1 counters, want 1", got)
	}
	if opp.Life != life-2 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-2)
	}
}
