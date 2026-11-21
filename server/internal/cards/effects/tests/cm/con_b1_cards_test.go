package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// con_b1_cards_test.go — Conflux, batch 1 (#5587). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	conB1AbsorbVisOracle           = "04ea7e88-7ea5-4b4c-9a7c-2a6ac0bc0682"
	conB1AvenSquireOracle          = "01518a87-6b16-4b7b-805d-b652404c6356"
	conB1AvenTrailblazerOracle     = "518115c9-0f0f-4249-81ff-e79b27dda1be"
	conB1CanyonMinotaurOracle      = "0c7163e6-f5a1-45a5-88c2-19dd9ef0a587"
	conB1CelestialPurgeOracle      = "ec1f6188-2516-46ac-8a03-7b7285b23a62"
	conB1CliffrunnerBehemothOracle = "ea0dde32-1638-47f1-bef6-0813794ba35b"
	conB1ConstrictingOracle        = "9e34ab18-5d39-49d9-9d3d-e1d4a3611b87"
	conB1CourtHomunculusOracle     = "b597e801-2648-48b9-8a21-02baa13e37af"
	conB1DarklitGargoyleOracle     = "e48d9efa-c169-4587-86fc-0021f265d69b"
	conB1DragDownOracle            = "30e397e9-a705-4484-8387-04fb84010b2d"
	conB1EmberWeaverOracle         = "c92221f7-76e8-483a-9d68-b19ff4d0baaf"
	conB1EsperCormorantsOracle     = "aeb6cb00-e8bd-4610-aa39-29663d9c9160"
	conB1FrontlineSageOracle       = "b0f6130c-58f7-4693-894c-598cd9a38ce7"
	conB1GluttonousSlimeOracle     = "5797684a-a022-45d2-aa2e-4d375b33017e"
	conB1GoblinOutlanderOracle     = "3cc99034-6eaf-4e45-9bff-57b74847343e"
	conB1HellsparkOracle           = "a7ba88c6-81e7-4ab7-9c9b-308831e56833"
	conB1InfectiousHorrorOracle    = "0931206b-eed2-40d0-9496-5ecefc7f8f90"
	conB1KraniocerosOracle         = "c66342bb-bbc1-4aaf-8d99-0af68292081e"
	conB1MatcaRiotersOracle        = "b46294cb-9971-415e-90d1-a02c44cd1391"
	conB1MightOfAlaraOracle        = "44570849-b2b0-47a8-9d77-f7c738896a0a"
	conB1MoltenFrameOracle         = "7c72df45-c3bd-4fb9-8365-5a5639566742"
	conB1NacatlOutlanderOracle     = "333555d6-3fcc-4021-ab60-afc36d723877"
	conB1NacatlSavageOracle        = "475412c9-8960-4085-bd8f-5f8a7c7485b9"
	conB1PestilentKathariOracle    = "755ab107-0f65-4536-aa4b-8e9d35a9e524"
	conB1RhoxBodyguardOracle       = "34bd39ed-4058-4c65-ba5f-9ac5cf7c166a"
	conB1RhoxMeditantOracle        = "832c6834-2cc7-41d5-b881-37c24c13ab8a"
	conB1SalvageSlasherOracle      = "196d6ef6-80d9-4bfd-a5ed-d9e8b00faad0"
	conB1ScepterOfInsightOracle    = "e2725411-a600-4ff3-94fb-5d52f5eda414"
	conB1ShamblingRemainsOracle    = "5a29fe9f-89bf-405e-acf6-1b56435cb9ee"
	conB1SporeBurstOracle          = "27a9e0af-76f0-4910-9f93-2a3617b4e79a"
	conB1SylvanBountyOracle        = "c20635d2-20e9-48f7-88f0-6de0f953bb08"
	conB1ThornlingOracle           = "55a59d01-b2cf-4c17-89c5-ce8312a85740"
	conB1ToxicIguanarOracle        = "87e148a0-4299-4771-b68a-6652a4c1c535"
	conB1TraumaticVisionsOracle    = "a30724f0-d174-408e-8baf-4bba3cbc45b5"
	conB1ValeronOutlanderOracle    = "01d70534-7c26-4a54-9eca-d0ebbefc2159"
	conB1ValiantGuardOracle        = "9b00b4c2-d5c4-4cf7-bdcf-cd96297c4350"
	conB1VedalkenOutlanderOracle   = "814a30dc-14e4-45cb-9320-b043e686b8ba"
	conB1VoicesFromTheVoidOracle   = "8c5264d2-8544-4e83-8ebc-7261dde80f01"
	conB1VolcanicFalloutOracle     = "bd6ce147-c991-40ac-b276-65b630fad8b6"
	conB1WanderingGoblinsOracle    = "8046981c-d68c-4fbb-a133-8037996c9d10"
	conB1WildLeotauOracle          = "d5277fe0-b707-41bb-9ff9-62f8be249dd2"
	conB1ZombieOutlanderOracle     = "857524d5-9b5c-42ac-a6cf-2d9bfebd7015"
)

func TestConB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		conB1AbsorbVisOracle:           "Absorb Vis",
		conB1AvenSquireOracle:          "Aven Squire",
		conB1AvenTrailblazerOracle:     "Aven Trailblazer",
		conB1CanyonMinotaurOracle:      "Canyon Minotaur",
		conB1CelestialPurgeOracle:      "Celestial Purge",
		conB1CliffrunnerBehemothOracle: "Cliffrunner Behemoth",
		conB1ConstrictingOracle:        "Constricting Tendrils",
		conB1CourtHomunculusOracle:     "Court Homunculus",
		conB1DarklitGargoyleOracle:     "Darklit Gargoyle",
		conB1DragDownOracle:            "Drag Down",
		conB1EmberWeaverOracle:         "Ember Weaver",
		conB1EsperCormorantsOracle:     "Esper Cormorants",
		conB1FrontlineSageOracle:       "Frontline Sage",
		conB1GluttonousSlimeOracle:     "Gluttonous Slime",
		conB1GoblinOutlanderOracle:     "Goblin Outlander",
		conB1HellsparkOracle:           "Hellspark Elemental",
		conB1InfectiousHorrorOracle:    "Infectious Horror",
		conB1KraniocerosOracle:         "Kranioceros",
		conB1MatcaRiotersOracle:        "Matca Rioters",
		conB1MightOfAlaraOracle:        "Might of Alara",
		conB1MoltenFrameOracle:         "Molten Frame",
		conB1NacatlOutlanderOracle:     "Nacatl Outlander",
		conB1NacatlSavageOracle:        "Nacatl Savage",
		conB1PestilentKathariOracle:    "Pestilent Kathari",
		conB1RhoxBodyguardOracle:       "Rhox Bodyguard",
		conB1RhoxMeditantOracle:        "Rhox Meditant",
		conB1SalvageSlasherOracle:      "Salvage Slasher",
		conB1ScepterOfInsightOracle:    "Scepter of Insight",
		conB1ShamblingRemainsOracle:    "Shambling Remains",
		conB1SporeBurstOracle:          "Spore Burst",
		conB1SylvanBountyOracle:        "Sylvan Bounty",
		conB1ThornlingOracle:           "Thornling",
		conB1ToxicIguanarOracle:        "Toxic Iguanar",
		conB1TraumaticVisionsOracle:    "Traumatic Visions",
		conB1ValeronOutlanderOracle:    "Valeron Outlander",
		conB1ValiantGuardOracle:        "Valiant Guard",
		conB1VedalkenOutlanderOracle:   "Vedalken Outlander",
		conB1VoicesFromTheVoidOracle:   "Voices from the Void",
		conB1VolcanicFalloutOracle:     "Volcanic Fallout",
		conB1WanderingGoblinsOracle:    "Wandering Goblins",
		conB1WildLeotauOracle:          "Wild Leotau",
		conB1ZombieOutlanderOracle:     "Zombie Outlander",
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

// conB1Seats is the active seat and the next one.
func conB1Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestConB1KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Esper Cormorants", "Artifact Creature — Bird", conB1EsperCormorantsOracle, []string{"flying"}},
		{"Goblin Outlander", "Creature — Goblin Scout", conB1GoblinOutlanderOracle, []string{"protection from white"}},
		{"Nacatl Outlander", "Creature — Cat Scout", conB1NacatlOutlanderOracle, []string{"protection from blue"}},
		{"Nacatl Savage", "Creature — Cat Warrior", conB1NacatlSavageOracle, []string{"protection from artifacts"}},
		{"Valeron Outlander", "Creature — Human Scout", conB1ValeronOutlanderOracle, []string{"protection from black"}},
		{"Vedalken Outlander", "Artifact Creature — Vedalken Scout", conB1VedalkenOutlanderOracle, []string{"protection from red"}},
		{"Zombie Outlander", "Creature — Zombie Scout", conB1ZombieOutlanderOracle, []string{"protection from green"}},
		{"Aven Squire", "Creature — Bird Soldier", conB1AvenSquireOracle, []string{"flying", game.KeywordExalted}},
		{"Frontline Sage", "Creature — Human Wizard", conB1FrontlineSageOracle, []string{game.KeywordExalted}},
		{"Rhox Bodyguard", "Creature — Rhino Monk Soldier", conB1RhoxBodyguardOracle, []string{game.KeywordExalted}},
		{"Gluttonous Slime", "Creature — Ooze", conB1GluttonousSlimeOracle, []string{"flash"}},
		{"Hellspark Elemental", "Creature — Elemental", conB1HellsparkOracle, []string{"trample", "haste"}},
		{"Darklit Gargoyle", "Artifact Creature — Gargoyle", conB1DarklitGargoyleOracle, []string{"flying"}},
		{"Pestilent Kathari", "Creature — Bird Warrior", conB1PestilentKathariOracle, []string{"flying", "deathtouch"}},
		{"Ember Weaver", "Creature — Spider", conB1EmberWeaverOracle, []string{"reach"}},
		{"Aven Trailblazer", "Creature — Bird Soldier", conB1AvenTrailblazerOracle, []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Protection from white: a white creature can't block Goblin Outlander
// (CR 702.16a).
func TestConB1GoblinOutlanderCantBeBlockedByWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := effectstest.B12Push(g, me.ID, "Goblin Outlander", "Creature — Goblin Scout", conB1GoblinOutlanderOracle, 2, 2)
	white := effectstest.TmpB1Push(g, opp.ID, "White Knight", "Creature — Human Knight", "", 2, 2, "W")
	if err := effectstest.PorBlock(t, g, goblin, opp.ID, white); err == nil {
		t.Error("a white creature blocked the Goblin Outlander")
	}
}

// --- exalted ----------------------------------------------------------------

func TestConB1AvenSquireExaltsALoneAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	squire := effectstest.B12Push(g, me.ID, "Aven Squire", "Creature — Bird Soldier", conB1AvenSquireOracle, 1, 1)
	if n := game.ExaltedCount(effectstest.ApaLive(g, squire)); n != 1 {
		t.Fatalf("Aven Squire has %d exalted instances, want 1", n)
	}
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AttackWith(t, g, opp.ID, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the lone attacker is %d/%d, want 3/3", p, tough)
	}
}

func TestConB1FrontlineSageLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := conB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	sage := effectstest.B12Push(g, me.ID, "Frontline Sage", "Creature — Human Wizard", conB1FrontlineSageOracle, 0, 1)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Instant"})
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	if err := effectstest.ConB1Activate(t, g, me, sage, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without {U}")
	}
	if err := effectstest.ConB1Activate(t, g, me, sage, 0, "U", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) || me.Hand.Size() != hand-1 || !me.Graveyard.Contains(fodder) {
		t.Errorf("hand %d -> %d; the fodder was not discarded", hand, me.Hand.Size())
	}
	if !effectstest.B13Tapped(t, g, sage) {
		t.Error("the Sage did not tap")
	}
}

func TestConB1RhoxBodyguardGainsThreeOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Rhox Bodyguard", "Creature — Rhino Monk Soldier", conB1RhoxBodyguardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

// --- triggers -----------------------------------------------------------------

func TestConB1InfectiousHorrorDrainsEachOpponentWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horror := effectstest.B12Push(g, me.ID, "Infectious Horror", "Creature — Zombie Horror", conB1InfectiousHorrorOracle, 2, 2)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(horror, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		switch s.ID {
		case me.ID:
			if s.Life != lives[s.ID] {
				t.Errorf("my life %d, want an unchanged %d", s.Life, lives[s.ID])
			}
		case opp.ID:
			// The attacked player may also have taken the combat damage.
			if s.Life > lives[s.ID]-2 {
				t.Errorf("attacked opponent's life %d, want at most %d", s.Life, lives[s.ID]-2)
			}
		default:
			if s.Life != lives[s.ID]-2 {
				t.Errorf("an opponent not attacked has %d life, want %d", s.Life, lives[s.ID]-2)
			}
		}
	}
}

func TestConB1WildLeotauIsPaidForOrSacrificed(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.B12Push(g, me.ID, "Wild Leotau", "Creature — Cat", conB1WildLeotauOracle, 5, 4)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.DcMana(t, g, me.ID, "{G}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if g.Battlefield.Contains(id) != pay {
			t.Errorf("paid=%v: the Leotau on the battlefield = %v", pay, g.Battlefield.Contains(id))
		}
		if !pay && !me.Graveyard.Contains(id) {
			t.Error("the declined Leotau is not in its owner's graveyard")
		}
	}
}

func TestConB1HellsparkElementalIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Hellspark Elemental", "Creature — Elemental", conB1HellsparkOracle, 3, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Fatal("the Elemental was not sacrificed at the end step")
	}
}

// Unearthed, the sacrifice and the unearth exile both trigger at the end
// step; either way it ends in exile, never the graveyard (CR 702.84a).
func TestConB1HellsparkElementalUnearthsAndEndsInExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Hellspark Elemental", "Creature — Elemental",
		conB1HellsparkOracle, 3, 1, "{C}{R}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the unearthed Elemental is not on the battlefield")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	// Two of its controller's triggers at once: the order prompt first.
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
		t.Errorf("battlefield %v, graveyard %v, exile %v: want exile only",
			g.Battlefield.Contains(id), me.Graveyard.Contains(id), g.Exile.Contains(id))
	}
}

func TestConB1ShamblingRemainsCantBlockAndUnearths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Shambling Remains", "Creature — Zombie Horror", conB1ShamblingRemainsOracle, 4, 3)
	effectstest.AssertRestrictions(t, g, id, game.CantBlock)

	g = effectstest.NewCatalogGame(t)
	id, _ = effectstest.ActivateFromGraveyard(t, g, "Shambling Remains", "Creature — Zombie Horror",
		conB1ShamblingRemainsOracle, 4, 3, "{B}{R}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("the unearthed Remains is not on the battlefield with haste")
	}
}

func TestConB1RhoxMeditantDrawsOnlyWithAGreenPermanent(t *testing.T) {
	for _, tc := range []struct {
		name      string
		greenMine bool
		greenOpp  bool
		drawn     int
	}{
		{"you control a green permanent", true, false, 1},
		{"only an opponent does", false, true, 0},
		{"none", false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := conB1Seats(g)
			effectstest.SeedLibrary(me, "A", "B")
			if tc.greenMine {
				effectstest.TmpB1Push(g, me.ID, "Elf", "Creature — Elf", "", 1, 1, "G")
			}
			if tc.greenOpp {
				effectstest.TmpB1Push(g, opp.ID, "Their Elf", "Creature — Elf", "", 1, 1, "G")
			}
			effectstest.CastCatalogSpell(t, g, "Rhox Meditant", "Creature — Rhino Monk", conB1RhoxMeditantOracle, nil)
			hand := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.drawn {
				t.Errorf("drew %d, want %d", got, tc.drawn)
			}
		})
	}
}

// --- spells -----------------------------------------------------------------

func TestConB1AbsorbVisDrainsFourAndLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	myLife, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Absorb Vis", "Sorcery", conB1AbsorbVisOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 || me.Life != myLife+4 {
		t.Errorf("opponent %d -> %d, me %d -> %d: want -4 and +4", oppLife, opp.Life, myLife, me.Life)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"), game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	id, _ := effectstest.CycleFromHand(t, g, "Absorb Vis", "Sorcery", conB1AbsorbVisOracle, "{B}{B}")
	if !me.Graveyard.Contains(id) {
		t.Error("Absorb Vis was not discarded to landcycle")
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Swamp")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the Swamp", hand, me.Hand.Size())
	}
}

func TestConB1SylvanBountyGivesAnyPlayerEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := conB1Seats(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Sylvan Bounty", "Instant", conB1SylvanBountyOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+8 {
		t.Errorf("target life %d -> %d, want +8", life, opp.Life)
	}
}

// Traumatic Visions counters a spell; Volcanic Fallout can't be
// countered (CR 101.2), so the same counter leaves it to resolve.
func TestConB1TraumaticVisionsCountersButNotVolcanicFallout(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Traumatic Visions", "Instant", conB1TraumaticVisionsOracle, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Error("the Bolt was not countered")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	fallout := effectstest.Batch01OpponentCasts(t, g, opp, "Volcanic Fallout", conB1VolcanicFalloutOracle, "", nil)
	effectstest.CastCatalogSpell(t, g, "Traumatic Visions", "Instant", conB1TraumaticVisionsOracle, effectstest.B16TargetCard(fallout))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	for _, s := range g.Seats {
		if s.Life != lives[s.ID]-2 {
			t.Errorf("a player has %d life, want %d: the Fallout was countered or missed a player", s.Life, lives[s.ID]-2)
		}
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 2 {
		t.Errorf("the 3/3 should survive with 2 damage, marked %d", effectstest.DamageMarkedOn(g, big))
	}
}

func TestConB1TraumaticVisionsLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Island", "Basic Land — Island"))
	effectstest.CycleFromHand(t, g, "Traumatic Visions", "Instant", conB1TraumaticVisionsOracle, "{U}{U}")
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the Island", hand, me.Hand.Size())
	}
}

func TestConB1ConstrictingTendrilsShrinksAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := conB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Constricting Tendrils", "Instant", conB1ConstrictingOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 4 {
		t.Errorf("the Ogre is %d/%d, want 1/4", p, tough)
	}
	g = effectstest.NewCatalogGame(t)
	effectstest.SeedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
	effectstest.Ons1Cycle(t, g, "Constricting Tendrils", "Instant", conB1ConstrictingOracle, "{C}{C}")
}

func TestConB1MoltenFrameDestroysOnlyAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := conB1Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Molten Frame", "Instant", conB1MoltenFrameOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("a nonartifact creature was a legal target")
	}
	g = effectstest.NewCatalogGame(t)
	_, opp = conB1Seats(g)
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Molten Frame", "Instant", conB1MoltenFrameOracle, effectstest.CardTarget(golem))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("the artifact creature was not destroyed")
	}
	g = effectstest.NewCatalogGame(t)
	effectstest.SeedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
	effectstest.Ons1Cycle(t, g, "Molten Frame", "Instant", conB1MoltenFrameOracle, "{C}{C}")
}

func TestConB1CelestialPurgeExilesBlackOrRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := conB1Seats(g)
	green := effectstest.TmpB1Push(g, opp.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Celestial Purge", "Instant", conB1CelestialPurgeOracle, effectstest.CardTarget(green)); err == nil {
		t.Fatal("a green creature was a legal target")
	}
	for _, color := range []string{"B", "R"} {
		g := effectstest.NewCatalogGame(t)
		_, opp := conB1Seats(g)
		target := effectstest.TmpB1Push(g, opp.ID, "Target", "Creature — Bear", "", 2, 2, color)
		effectstest.CastCatalogSpell(t, g, "Celestial Purge", "Instant", conB1CelestialPurgeOracle, effectstest.CardTarget(target))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) || !g.Exile.Contains(target) {
			t.Errorf("the %s permanent was not exiled", color)
		}
	}
}

// --- domain -------------------------------------------------------------------

// domainLands(…, "Forest", "Island") gives me three basic land types
// (it adds a Plains) and the opponent a Swamp that does not count.
func TestConB1MightOfAlaraAndDragDownCountBasicLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Might of Alara", "Instant", conB1MightOfAlaraOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Might of Alara: the Bear is %d/%d, want 5/5", p, tough)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = conB1Seats(g)
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Drag Down", "Instant", conB1DragDownOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("Drag Down: the Ogre is %d/%d, want 1/1", p, tough)
	}
}

func TestConB1SporeBurstMakesASaprolingPerBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	effectstest.DomainLands(g, me, opp, "Forest", "Island", "Mountain")
	effectstest.CastCatalogSpell(t, g, "Spore Burst", "Sorcery", conB1SporeBurstOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Saproling" {
			n++
		}
	}
	if n != 4 {
		t.Errorf("%d Saprolings, want 4 (Forest, Island, Mountain, Plains)", n)
	}
}

func TestConB1VoicesFromTheVoidDiscardsPerBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	cards := []uuid.UUID{
		effectstest.HandCard(opp, "One", "Instant"), effectstest.HandCard(opp, "Two", "Instant"),
		effectstest.HandCard(opp, "Three", "Instant"), effectstest.HandCard(opp, "Four", "Instant"),
	}
	hand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Voices from the Void", "Sorcery", conB1VoicesFromTheVoidOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, cards[:3]...)
	if opp.Hand.Size() != hand-3 {
		t.Errorf("opponent's hand %d -> %d, want three discarded", hand, opp.Hand.Size())
	}
}

func TestConB1DomainCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	rioters := effectstest.B12Push(g, me.ID, "Matca Rioters", "Creature — Human Warrior", conB1MatcaRiotersOracle, 0, 0)
	aven := effectstest.B12Push(g, me.ID, "Aven Trailblazer", "Creature — Bird Soldier", conB1AvenTrailblazerOracle, 2, 0)
	if p, tough := effectstest.PtOf(t, g, rioters); p != 3 || tough != 3 {
		t.Errorf("Matca Rioters is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, aven); p != 2 || tough != 3 {
		t.Errorf("Aven Trailblazer is %d/%d, want 2/3", p, tough)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if p, tough := effectstest.PtOf(t, g, rioters); p != 4 || tough != 4 {
		t.Errorf("with a fourth type Matca Rioters is %d/%d, want 4/4", p, tough)
	}
}

func TestConB1WanderingGoblinsPumpsPerBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	goblins := effectstest.B12Push(g, me.ID, "Wandering Goblins", "Creature — Goblin Warrior", conB1WanderingGoblinsOracle, 0, 3)
	if err := effectstest.ConB1Activate(t, g, me, goblins, 0, "CC", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated for {2}; it costs {3}")
	}
	if err := effectstest.ConB1Activate(t, g, me, goblins, 0, "C", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, goblins); p != 3 || tough != 3 {
		t.Errorf("the Goblins are %d/%d, want 3/3", p, tough)
	}
}

// --- activated rows -----------------------------------------------------------

func TestConB1SelfPumpRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough, row      int
		mana                   string
		wantP, wantT           int
	}{
		{"Darklit Gargoyle", "Artifact Creature — Gargoyle", conB1DarklitGargoyleOracle, 1, 2, 0, "B", 3, 1},
		{"Kranioceros", "Creature — Beast", conB1KraniocerosOracle, 5, 2, 0, "CW", 5, 5},
		{"Thornling +1/-1", "Creature — Elemental Shapeshifter", conB1ThornlingOracle, 4, 4, 3, "C", 5, 3},
		{"Thornling -1/+1", "Creature — Elemental Shapeshifter", conB1ThornlingOracle, 4, 4, 4, "C", 3, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := conB1Seats(g)
			effectstest.AdvanceToMain(t, g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			if err := effectstest.ConB1Activate(t, g, me, id, tc.row, "", game.ActivateAbilityParams{}); err == nil {
				t.Fatal("activated with no mana")
			}
			if err := effectstest.ConB1Activate(t, g, me, id, tc.row, tc.mana, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestConB1KeywordRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		row                    int
		mana, keyword          string
	}{
		{"Pestilent Kathari", "Creature — Bird Warrior", conB1PestilentKathariOracle, 0, "CCR", "first strike"},
		{"Thornling", "Creature — Elemental Shapeshifter", conB1ThornlingOracle, 0, "G", "haste"},
		{"Thornling", "Creature — Elemental Shapeshifter", conB1ThornlingOracle, 1, "G", "trample"},
		{"Thornling", "Creature — Elemental Shapeshifter", conB1ThornlingOracle, 2, "G", "indestructible"},
	} {
		t.Run(tc.name+" "+tc.keyword, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := conB1Seats(g)
			effectstest.AdvanceToMain(t, g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
			if effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
				t.Fatalf("has %s before the activation", tc.keyword)
			}
			if err := effectstest.ConB1Activate(t, g, me, id, tc.row, tc.mana, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
				t.Errorf("did not gain %s", tc.keyword)
			}
		})
	}
}

func TestConB1ScepterOfInsightDrawsTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := conB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B")
	scepter := effectstest.PushCatalogPermanent(g, me.ID, "Scepter of Insight", "Artifact", conB1ScepterOfInsightOracle, true)
	if err := effectstest.ConB1Activate(t, g, me, scepter, 0, "CCC", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without {U}")
	}
	if err := effectstest.ConB1Activate(t, g, me, scepter, 0, "U", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if !effectstest.B13Tapped(t, g, scepter) {
		t.Error("the Scepter did not tap")
	}
}

// --- statics --------------------------------------------------------------------

func TestConB1CourtHomunculusNeedsAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	hom := effectstest.B12Push(g, me.ID, "Court Homunculus", "Artifact Creature — Homunculus", conB1CourtHomunculusOracle, 1, 1)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p, tough := effectstest.PtOf(t, g, hom); p != 1 || tough != 1 {
		t.Errorf("alone (an opponent's artifact only) it is %d/%d, want 1/1", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	if p, tough := effectstest.PtOf(t, g, hom); p != 2 || tough != 2 {
		t.Errorf("with another artifact it is %d/%d, want 2/2", p, tough)
	}
}

func TestConB1SalvageSlasherCountsArtifactCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	slasher := effectstest.B12Push(g, me.ID, "Salvage Slasher", "Artifact Creature — Human Rogue", conB1SalvageSlasherOracle, 1, 1)
	effectstest.Batch01GraveyardCard(me, "Relic", "Artifact")
	effectstest.Batch01GraveyardCard(me, "Golem", "Artifact Creature — Golem")
	effectstest.Batch01GraveyardCard(me, "Bear", "Creature — Bear")
	effectstest.Batch01GraveyardCard(opp, "Their Relic", "Artifact")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if p, tough := effectstest.PtOf(t, g, slasher); p != 3 || tough != 1 {
		t.Errorf("the Slasher is %d/%d, want 3/1 (two artifact cards in your graveyard)", p, tough)
	}
}

func TestConB1ColouredPermanentStatics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := conB1Seats(g)
	iguanar := effectstest.B12Push(g, me.ID, "Toxic Iguanar", "Creature — Lizard", conB1ToxicIguanarOracle, 1, 1)
	weaver := effectstest.B12Push(g, me.ID, "Ember Weaver", "Creature — Spider", conB1EmberWeaverOracle, 2, 3)
	behemoth := effectstest.B12Push(g, me.ID, "Cliffrunner Behemoth", "Creature — Rhino Beast", conB1CliffrunnerBehemothOracle, 5, 3)
	// Every colour on the opponent's side: none of it counts.
	effectstest.TmpB1Push(g, opp.ID, "Their Rainbow", "Creature — Elemental", "", 1, 1, "W", "R", "G")
	if effectstest.HasKeywordOnBattlefield(t, g, iguanar, "deathtouch") || effectstest.HasKeywordOnBattlefield(t, g, weaver, "first strike") ||
		effectstest.HasKeywordOnBattlefield(t, g, behemoth, "haste") || effectstest.HasKeywordOnBattlefield(t, g, behemoth, "lifelink") {
		t.Fatal("an opponent's coloured permanent turned a static on")
	}
	if p, _ := effectstest.PtOf(t, g, weaver); p != 2 {
		t.Errorf("the Weaver's power is %d with no red permanent, want 2", p)
	}

	effectstest.TmpB1Push(g, me.ID, "My Elf", "Creature — Elf", "", 1, 1, "G")
	if !effectstest.HasKeywordOnBattlefield(t, g, iguanar, "deathtouch") {
		t.Error("the Iguanar has no deathtouch with a green permanent")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, behemoth, "haste") {
		t.Error("the Behemoth has haste with only a green permanent")
	}
	effectstest.TmpB1Push(g, me.ID, "My Goblin", "Creature — Goblin", "", 1, 1, "R")
	if p, tough := effectstest.PtOf(t, g, weaver); p != 3 || tough != 3 || !effectstest.HasKeywordOnBattlefield(t, g, weaver, "first strike") {
		t.Errorf("with a red permanent the Weaver is %d/%d (first strike %v), want 3/3 with first strike",
			p, tough, effectstest.HasKeywordOnBattlefield(t, g, weaver, "first strike"))
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, behemoth, "haste") || effectstest.HasKeywordOnBattlefield(t, g, behemoth, "lifelink") {
		t.Error("with red and no white the Behemoth wants haste and no lifelink")
	}
	effectstest.TmpB1Push(g, me.ID, "My Soldier", "Creature — Human Soldier", "", 1, 1, "W")
	if !effectstest.HasKeywordOnBattlefield(t, g, behemoth, "lifelink") {
		t.Error("the Behemoth has no lifelink with a white permanent")
	}
}

func TestConB1GluttonousSlimeDevoursOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	b := effectstest.PushDevourFodder(g, me.ID, "Goblin B")
	id := effectstest.CastDevourer(t, g, "Gluttonous Slime", conB1GluttonousSlimeOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	if got := effectstest.DevourPlusOnes(g, id); got != 2 {
		t.Errorf("counters = %d, want 2 (devour 1 x 2)", got)
	}
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the devoured creatures are not in the graveyard")
	}
}
