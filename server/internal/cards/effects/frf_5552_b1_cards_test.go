package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// frf_5552_b1_cards_test.go — Fate Reforged batch 1 (#5552). Every test
// plays the real catalog entry through the engine.

const (
	frfArchersOfQarsiOracle     = "ad6bdd24-5a32-4046-83ee-2fe51bf8bb7a"
	frfAvenSkirmisherOracle     = "57d8d656-ccbd-4deb-abf1-a7fde4e7d691"
	frfFeralKrushokOracle       = "8c42883d-8b70-4691-a1e7-4659e7954b88"
	frfGoreSwineOracle          = "af4b5853-ae60-48c9-9be1-bf121fa2e03d"
	frfGreatHornKrushokOracle   = "b193a6df-da22-4e3c-a086-5b02c604089e"
	frfPilgrimOfTheFiresOracle  = "53390f91-a159-482f-9f64-f5adc812680b"
	frfDragonBellMonkOracle     = "83f7e8e5-d485-4571-9fa6-47b5e962d221"
	frfLotusPathDjinnOracle     = "84dc0aae-029b-48dc-b8d0-e93c227e32e5"
	frfMarduScoutOracle         = "c18ab011-c5e0-4174-b623-d89e4d745ce6"
	frfAleshasVanguardOracle    = "be662217-41c2-4caa-89da-ae7f6f817c34"
	frfMarduShadowspearOracle   = "b6e972ae-cc2d-4c7e-9f49-c0616dc82064"
	frfGoblinHeelcutterOracle   = "ab52988f-ffd0-4154-9052-593a9d450d98"
	frfVaultbreakerOracle       = "2577860a-663a-4b56-880c-05576522e3b7"
	frfDromokaOracle            = "73520b5a-3d55-40c9-af60-90fdd811934f"
	frfArashinClericOracle      = "69bcdb42-c170-4848-b95b-f4621867d084"
	frfSandsteppeMastodonOracle = "99fa202b-052c-4627-b78c-f0b323f7bdf3"
	frfSibsigHostOracle         = "11776f73-f504-4f19-b5b6-2f58279f8b15"
	frfSultaiSkullkeeperOracle  = "726ba167-9e7b-4d46-a52c-72d5a500aee2"
	frfLotusEyeMysticsOracle    = "efb11270-4b5a-4c52-862f-1d24c6b95d73"
	frfAmbushKrotiqOracle       = "e53f5f82-a8e0-44b1-a3be-162a0724755d"
	frfJeskaiBarricadeOracle    = "7a2fa0a6-1324-4714-8896-6bb97640578a"
	frfDefiantOgreOracle        = "db849660-1ad4-4af4-8822-40ba6c0b6d73"
	frfAvenSurveyorOracle       = "57831c07-6915-4579-b3e1-de93a3664444"
	frfAbzanSkycaptainOracle    = "866fc4b9-f661-444b-b7fd-929814dbc69d"
	frfJeskaiSageOracle         = "b79d4b17-482e-415b-a63b-1f97a55e7583"
	frfSmolderingEfreetOracle   = "789b6a61-656e-40c0-be78-9c89c4ecfb0b"
	frfDestructorDragonOracle   = "b0104d66-5c28-482f-838a-416aabaeaf43"
	frfSultaiEmissaryOracle     = "0274f286-46a1-4028-878a-88c2146a895c"
	frfAbzanKinGuardOracle      = "b292d183-1ba9-40b4-9cba-bd70bf1380ef"
	frfBattleBrawlerOracle      = "d46d8913-7575-4aef-b511-6c0ca9d4b506"
	frfAbzanRunemarkOracle      = "0a9276fc-387c-4bde-a6d6-b0dd7abee94b"
	frfJeskaiRunemarkOracle     = "5fe13418-c63c-4f3c-9a42-7ab06949e3cd"
	frfMarduRunemarkOracle      = "a912834b-e4bc-4ccc-ae13-915ebe5acd00"
	frfSultaiRunemarkOracle     = "59410454-93ff-41f0-ab62-cdaf156163d0"
	frfTemurRunemarkOracle      = "590c2455-90f7-4ff6-a6cf-5d0d25575cda"
	frfFrostWalkerOracle        = "54a22814-c0d5-4637-9aa9-a6bc96401914"
	frfEnhancedAwarenessOracle  = "397964db-465a-407d-9dc7-d0a4944473ac"
	frfPressurePointOracle      = "d4bbdd0c-e578-49ed-b2f7-b1decb04ee87"
	frfRefocusOracle            = "0101f891-e54c-43e5-86f3-263d95060fcd"
	frfNeutralizingBlastOracle  = "e42ca67d-c9c3-499a-a7b7-63dc2dafdeab"
	frfReachOfShadowsOracle     = "e01382e0-a685-41c4-9ab6-727ffdbf2e0a"
	frfReturnToTheEarthOracle   = "722f2965-a1ec-4b97-a4bd-b99364f58b1e"
	frfSandblastOracle          = "41ed9b02-5926-4af0-bcee-7f41bbb3fdd6"
	frfCollateralDamageOracle   = "b89dd47e-0f65-41f0-ad23-cefc61e68978"
	frfCachedDefensesOracle     = "0e564f52-94fb-4c44-926f-0eebedc7d4c0"
	frfHonorsRewardOracle       = "87e6aabb-41f5-41d2-b662-cd09ad61a1c8"
	frfWhiskAwayOracle          = "ac141fe0-328a-4d9b-8048-b0c47f399587"
	frfEtherealAmbushOracle     = "c0604afb-eae2-41c8-b390-dbc0f7e025ed"

	frfBearType  = "Creature — Bear"
	frfAuraType  = "Enchantment — Aura"
	frfDragonTyp = "Creature — Dragon"
)

// frfPushLibraryTop puts a named creature card on top of p's library.
func frfPushLibraryTop(g *game.Game, p *game.Player, name string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = frfBearType
	c.ManaCost = "{1}{G}"
	c.Power, c.Toughness = 2, 2
	g.WithWriteLock(func() { p.Library.PushTop(c) })
	return c.InstanceID
}

// frfPlusOnes reads a battlefield permanent's +1/+1 counters under the
// read lock.
func frfPlusOnes(g *game.Game, id uuid.UUID) int {
	var n int
	g.ReadSnapshot(func() { n = plusOneCounters(g, id) })
	return n
}

// frfPendingFor reports whether any prompt is open for chooser.
func frfPendingFor(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == chooser {
			return true
		}
	}
	return false
}

// --- every card is in the catalog ----------------------------------------

func TestFrf5552B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		frfArchersOfQarsiOracle, frfAvenSkirmisherOracle, frfFeralKrushokOracle, frfGoreSwineOracle,
		frfGreatHornKrushokOracle, frfPilgrimOfTheFiresOracle, frfDragonBellMonkOracle, frfLotusPathDjinnOracle,
		frfMarduScoutOracle, frfAleshasVanguardOracle, frfMarduShadowspearOracle, frfGoblinHeelcutterOracle,
		frfVaultbreakerOracle, frfDromokaOracle, frfArashinClericOracle, frfSandsteppeMastodonOracle,
		frfSibsigHostOracle, frfSultaiSkullkeeperOracle, frfLotusEyeMysticsOracle, frfAmbushKrotiqOracle,
		frfJeskaiBarricadeOracle, frfDefiantOgreOracle, frfAvenSurveyorOracle, frfAbzanSkycaptainOracle,
		frfJeskaiSageOracle, frfSmolderingEfreetOracle, frfDestructorDragonOracle, frfSultaiEmissaryOracle,
		frfAbzanKinGuardOracle, frfBattleBrawlerOracle, frfAbzanRunemarkOracle, frfJeskaiRunemarkOracle,
		frfMarduRunemarkOracle, frfSultaiRunemarkOracle, frfTemurRunemarkOracle, frfFrostWalkerOracle,
		frfEnhancedAwarenessOracle, frfPressurePointOracle, frfRefocusOracle, frfNeutralizingBlastOracle,
		frfReachOfShadowsOracle, frfReturnToTheEarthOracle, frfSandblastOracle, frfCollateralDamageOracle,
		frfCachedDefensesOracle, frfHonorsRewardOracle, frfWhiskAwayOracle, frfEtherealAmbushOracle,
	} {
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

// --- printed keywords -----------------------------------------------------------

func TestFrf5552B1PrintedKeywords(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Archers of Qarsi", "Creature — Snake Archer", frfArchersOfQarsiOracle, []string{"defender", "reach"}},
		{"Aven Skirmisher", "Creature — Bird Warrior", frfAvenSkirmisherOracle, []string{"flying"}},
		{"Pilgrim of the Fires", "Artifact Creature — Golem", frfPilgrimOfTheFiresOracle, []string{"first strike", "trample"}},
		{"Dragon Bell Monk", "Creature — Human Monk", frfDragonBellMonkOracle, []string{"vigilance", "prowess"}},
		{"Lotus Path Djinn", "Creature — Djinn Monk", frfLotusPathDjinnOracle, []string{"flying", "prowess"}},
		{"Dromoka, the Eternal", "Legendary Creature — Dragon", frfDromokaOracle, []string{"flying"}},
		{"Sandsteppe Mastodon", "Creature — Elephant", frfSandsteppeMastodonOracle, []string{"reach"}},
		{"Lotus-Eye Mystics", "Creature — Human Monk", frfLotusEyeMysticsOracle, []string{"prowess"}},
		{"Ambush Krotiq", "Creature — Insect", frfAmbushKrotiqOracle, []string{"trample"}},
		{"Jeskai Barricade", "Creature — Wall", frfJeskaiBarricadeOracle, []string{"flash", "defender"}},
		{"Aven Surveyor", "Creature — Bird Scout", frfAvenSurveyorOracle, []string{"flying"}},
		{"Abzan Skycaptain", "Creature — Bird Soldier", frfAbzanSkycaptainOracle, []string{"flying"}},
		{"Jeskai Sage", "Creature — Human Monk", frfJeskaiSageOracle, []string{"prowess"}},
		{"Destructor Dragon", frfDragonTyp, frfDestructorDragonOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Prowess (CR 702.108a): a noncreature spell pumps Dragon Bell Monk.
func TestFrf5552B1DragonBellMonkProwess(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	monk := b12Push(g, me.ID, "Dragon Bell Monk", "Creature — Human Monk", frfDragonBellMonkOracle, 2, 2)
	castCatalogSpell(t, g, "Enhanced Awareness", "Instant", frfEnhancedAwarenessOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, monk); p != 3 || tough != 3 {
		t.Errorf("the Monk is %d/%d after a noncreature spell, want 3/3", p, tough)
	}
}

// --- dash ----------------------------------------------------------------------

func TestFrf5552B1DashOffers(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Mardu Scout", frfMarduScoutOracle, "{1}{R}"},
		{"Alesha's Vanguard", frfAleshasVanguardOracle, "{2}{B}"},
		{"Mardu Shadowspear", frfMarduShadowspearOracle, "{1}{B}"},
		{"Goblin Heelcutter", frfGoblinHeelcutterOracle, "{2}{R}"},
		{"Vaultbreaker", frfVaultbreakerOracle, "{2}{R}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := game.AlternativeCostByKey(tc.oracle, "dash")
			if b == nil || !b.Dash || b.ManaCost != tc.cost {
				t.Fatalf("dash offer = %+v, want Dash %s", b, tc.cost)
			}
		})
	}
}

// Dashed for {1}{R}, Mardu Scout has haste, attacks the turn it is cast,
// and goes back to its owner's hand at the beginning of the end step.
func TestFrf5552B1MarduScoutDashes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	id := handCardFull(me, "Mardu Scout", "Creature — Goblin Scout", "{R}{R}", frfMarduScoutOracle, []string{"R"})
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 3, 1
			}
		}
	})
	floatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "dash"}); err != nil {
		t.Fatalf("dash Mardu Scout: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("the dashed Scout has no haste")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("the dashed Scout can't attack: %v", err)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, id) || !me.Hand.Contains(id) {
		t.Error("the dashed Scout did not return to its owner's hand at the end step")
	}
}

// --- attack triggers -------------------------------------------------------------

func TestFrf5552B1MarduShadowspearDrainsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spear := b12Push(g, me.ID, "Mardu Shadowspear", "Creature — Human Warrior", frfMarduShadowspearOracle, 1, 1)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	attackWith(t, g, opp.ID, spear)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		switch s.ID {
		case me.ID:
			if s.Life != before[s.ID] {
				t.Errorf("the attacker's controller lost life: %d -> %d", before[s.ID], s.Life)
			}
		case opp.ID:
			if s.Life > before[s.ID]-1 {
				t.Errorf("the defending opponent did not lose life: %d -> %d", before[s.ID], s.Life)
			}
		default:
			if s.Life != before[s.ID]-1 {
				t.Errorf("an uninvolved opponent life %d -> %d, want -1", before[s.ID], s.Life)
			}
		}
	}
}

func TestFrf5552B1GoblinHeelcutterStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushDiesCreatureForTest(g, me.ID, "Goblin Heelcutter", frfGoblinHeelcutterOracle, "Creature — Goblin Berserker", 3, 2)
	wall := b16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4, "W")
	attackWithTrigger(t, g, id, opp.ID, wall)
	var restricted bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(wall); ok {
			restricted = c.Effective().Restrictions&game.CantBlock != 0
		}
	})
	if !restricted {
		t.Error("the targeted creature can still block")
	}
}

func TestFrf5552B1VaultbreakerOffersARummage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d2dSetHand(me, 2)
	id := pushDiesCreatureForTest(g, me.ID, "Vaultbreaker", frfVaultbreakerOracle, "Creature — Orc Rogue", 4, 2)
	attackWithTrigger(t, g, id, opp.ID, uuid.Nil)
	if !frfPendingFor(g, me.ID) {
		t.Error("attacking did not offer the discard")
	}
}

func TestFrf5552B1DromokaBolstersWhenADragonAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dromoka := pushDiesCreatureForTest(g, me.ID, "Dromoka, the Eternal", frfDromokaOracle, "Legendary Creature — Dragon", 5, 5)
	weak := b12Creature(g, me.ID, "Weak", frfBearType, 1, 1)
	attackWithTrigger(t, g, dromoka, opp.ID, uuid.Nil)
	if n := frfPlusOnes(g, weak); n != 2 {
		t.Errorf("the least-toughness creature has %d +1/+1 counters, want 2", n)
	}
}

// --- enters triggers -------------------------------------------------------------

func TestFrf5552B1ArashinClericGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Arashin Cleric", "Creature — Human Cleric", frfArashinClericOracle, 1, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestFrf5552B1SandsteppeMastodonBolstersFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weak := b12Creature(g, me.ID, "Weak", frfBearType, 1, 1)
	castHoldCreature(t, g, "Sandsteppe Mastodon", "Creature — Elephant", frfSandsteppeMastodonOracle, 5, 5)
	passPriorityAroundTable(t, g)
	if n := frfPlusOnes(g, weak); n != 5 {
		t.Errorf("the least-toughness creature has %d +1/+1 counters, want 5", n)
	}
}

func TestFrf5552B1SibsigHostMillsEachPlayerThree(t *testing.T) {
	g := newCatalogGame(t)
	for _, s := range g.Seats {
		for i := 0; i < 4; i++ {
			frfPushLibraryTop(g, s, "Filler")
		}
	}
	libs := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		libs[s.ID] = s.Library.Size()
	}
	castHoldCreature(t, g, "Sibsig Host", "Creature — Zombie", frfSibsigHostOracle, 2, 6)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if got := libs[s.ID] - s.Library.Size(); got != 3 {
			t.Errorf("%s milled %d, want 3", s.Name, got)
		}
	}
}

func TestFrf5552B1SultaiSkullkeeperMillsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		frfPushLibraryTop(g, me, "Filler")
	}
	library := me.Library.Size()
	castHoldCreature(t, g, "Sultai Skullkeeper", "Creature — Snake Shaman", frfSultaiSkullkeeperOracle, 2, 1)
	passPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestFrf5552B1LotusEyeMysticsReturnsAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aura := wthGraveyardCard(me, "Old Aura", frfAuraType)
	castHoldCreature(t, g, "Lotus-Eye Mystics", "Creature — Human Monk", frfLotusEyeMysticsOracle, 3, 2)
	answerPickTarget(t, g, aura)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) || me.Graveyard.Contains(aura) {
		t.Error("the enchantment card did not return to hand")
	}
}

func TestFrf5552B1AmbushKrotiqReturnsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := b12Creature(g, me.ID, "Other", frfBearType, 2, 2)
	krotiq := castHoldCreature(t, g, "Ambush Krotiq", "Creature — Insect", frfAmbushKrotiqOracle, 5, 5)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return a creature")
	}
	if hasID(pick.ChooseCards, krotiq) {
		t.Error("the Krotiq itself is offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{other}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) || !me.Hand.Contains(other) {
		t.Error("the other creature did not return to hand")
	}
}

func TestFrf5552B1JeskaiBarricadeMayReturnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0]
	bear := pushCatalogPermanent(g, p0.ID, "Bears", frfBearType, "", false)
	castCatalogSpell(t, g, "Jeskai Barricade", "Creature — Wall", frfJeskaiBarricadeOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, p0.ID, bear)
	answerLatestTriggerPrompt(t, g, p0.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !p0.Hand.Contains(bear) {
		t.Error("the chosen creature did not return to its owner's hand")
	}
}

func TestFrf5552B1DefiantOgreDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castHoldCreature(t, g, "Defiant Ogre", "Creature — Ogre Warrior", frfDefiantOgreOracle, 3, 5)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatal(err)
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestFrf5552B1AvenSurveyorBouncesOrGrows(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		surveyor := castHoldCreature(t, g, "Aven Surveyor", "Creature — Bird Scout", frfAvenSurveyorOracle, 2, 2)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if n := frfPlusOnes(g, surveyor); n != 1 {
			t.Errorf("%d counters on the Surveyor, want 1", n)
		}
	})
	t.Run("bounce", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, opp.ID, "Bear", frfBearType, 2, 2)
		castHoldCreature(t, g, "Aven Surveyor", "Creature — Bird Scout", frfAvenSurveyorOracle, 2, 2)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
			t.Error("the target did not return to its owner's hand")
		}
	})
}

// --- dies triggers ---------------------------------------------------------------

func TestFrf5552B1AbzanSkycaptainBolstersTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	captain := b12Push(g, me.ID, "Abzan Skycaptain", "Creature — Bird Soldier", frfAbzanSkycaptainOracle, 2, 2)
	weak := b12Creature(g, me.ID, "Weak", frfBearType, 1, 1)
	b12Creature(g, me.ID, "Strong", frfBearType, 4, 4)
	b27Kill(g, captain)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := frfPlusOnes(g, weak); n != 2 {
		t.Errorf("the least-toughness creature has %d +1/+1 counters, want 2", n)
	}
}

func TestFrf5552B1JeskaiSageDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sage := b12Push(g, me.ID, "Jeskai Sage", "Creature — Human Monk", frfJeskaiSageOracle, 1, 1)
	hand := me.Hand.Size()
	b27Kill(g, sage)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestFrf5552B1SmolderingEfreetDealsTwoToYouWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	efreet := b12Push(g, me.ID, "Smoldering Efreet", "Creature — Efreet Monk", frfSmolderingEfreetOracle, 2, 2)
	life := me.Life
	b27Kill(g, efreet)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestFrf5552B1DestructorDragonDestroysANoncreaturePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := b12Push(g, me.ID, "Destructor Dragon", frfDragonTyp, frfDestructorDragonOracle, 4, 4)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	b27Kill(g, dragon)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the noncreature permanent survived")
	}
}

func TestFrf5552B1SultaiEmissaryManifestsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emissary := b12Push(g, me.ID, "Sultai Emissary", "Creature — Zombie Warrior", frfSultaiEmissaryOracle, 1, 1)
	top := frfPushLibraryTop(g, me, "Hidden")
	b27Kill(g, emissary)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	c := findCardAnywhere(t, g, top)
	if !c.FaceDownIsPermanent() || c.FaceDownKind != game.FaceDownManifested || c.Controller != me.ID {
		t.Errorf("the top card is not a manifest you control: face-down %v (%q), controller %s", c.FaceDown, c.FaceDownKind, c.Controller)
	}
}

// --- statics ----------------------------------------------------------------------

func TestFrf5552B1AbzanKinGuardLifelinkWithAWhiteOrBlackPermanent(t *testing.T) {
	for _, tc := range []struct {
		colour string
		want   bool
	}{{"", false}, {"R", false}, {"W", true}, {"B", true}} {
		t.Run("colour "+tc.colour, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			if tc.colour != "" {
				tmpB1Push(g, me.ID, "Other", frfBearType, "", 1, 1, tc.colour)
			}
			guard := b12Push(g, me.ID, "Abzan Kin-Guard", "Creature — Human Warrior", frfAbzanKinGuardOracle, 3, 3)
			if got := hasEffectiveKeyword(t, g, guard, "lifelink"); got != tc.want {
				t.Errorf("lifelink = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFrf5552B1BattleBrawlerWithARedOrWhitePermanent(t *testing.T) {
	for _, tc := range []struct {
		colour string
		want   bool
	}{{"", false}, {"G", false}, {"R", true}, {"W", true}} {
		t.Run("colour "+tc.colour, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			if tc.colour != "" {
				tmpB1Push(g, me.ID, "Other", frfBearType, "", 1, 1, tc.colour)
			}
			brawler := b12Push(g, me.ID, "Battle Brawler", "Creature — Orc Warrior", frfBattleBrawlerOracle, 2, 2)
			wantPower := 2
			if tc.want {
				wantPower = 3
			}
			if p, tough := ptOf(t, g, brawler); p != wantPower || tough != 2 {
				t.Errorf("the Brawler is %d/%d, want %d/2", p, tough, wantPower)
			}
			if got := hasEffectiveKeyword(t, g, brawler, "first strike"); got != tc.want {
				t.Errorf("first strike = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFrf5552B1Runemarks(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword, yes, no string
	}{
		{"Abzan Runemark", frfAbzanRunemarkOracle, "vigilance", "G", "W"},
		{"Jeskai Runemark", frfJeskaiRunemarkOracle, "flying", "R", "U"},
		{"Mardu Runemark", frfMarduRunemarkOracle, "first strike", "B", "R"},
		{"Sultai Runemark", frfSultaiRunemarkOracle, "deathtouch", "U", "B"},
		{"Temur Runemark", frfTemurRunemarkOracle, "trample", "U", "G"},
	} {
		for _, colour := range []string{tc.yes, tc.no} {
			t.Run(tc.name+" with "+colour, func(t *testing.T) {
				g := newCatalogGame(t)
				advanceToMain(t, g)
				me := top100ActiveSeat(g)
				tmpB1Push(g, me.ID, "Coloured", "Artifact", "", 0, 0, colour)
				bear := b12Creature(g, me.ID, "Bear", frfBearType, 2, 2)
				castCatalogSpell(t, g, tc.name, frfAuraType, tc.oracle, cardTarget(bear))
				passPriorityAroundTable(t, g)
				if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
					t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tough)
				}
				want := colour == tc.yes
				if got := hasEffectiveKeyword(t, g, bear, tc.keyword); got != want {
					t.Errorf("%s = %v, want %v", tc.keyword, got, want)
				}
			})
		}
	}
}

// Targeted by any spell, Frost Walker is sacrificed.
func TestFrf5552B1FrostWalkerIsSacrificedWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	walker := b12Push(g, opp.ID, "Frost Walker", "Creature — Elemental", frfFrostWalkerOracle, 4, 1)
	castCatalogSpell(t, g, "Pressure Point", "Instant", frfPressurePointOracle, cardTarget(walker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) || !opp.Graveyard.Contains(walker) {
		t.Error("Frost Walker was not sacrificed")
	}
}

// --- the spells --------------------------------------------------------------------

func TestFrf5552B1EnhancedAwarenessDrawsThreeThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Enhanced Awareness", "Instant", frfEnhancedAwarenessOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+3)
	}
	if n := discardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestFrf5552B1PressurePointTapsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", frfBearType, 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Pressure Point", "Instant", frfPressurePointOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestFrf5552B1RefocusUntapsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", frfBearType, 2, 2)
	g.WithWriteLock(func() { b16Tap(g, bear) })
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Refocus", "Instant", frfRefocusOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if hmlTapped(t, g, bear) {
		t.Error("the target is still tapped")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestFrf5552B1NeutralizingBlastCountersOnlyAMulticoloredSpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mono := b14CastAt(t, g, opp, "Mono Spell", "Instant", "", []string{"U"}, game.CastSpellParams{})
	if err := g.CastSpell(me.ID, b14HandCard(me, "Neutralizing Blast", "Instant", frfNeutralizingBlastOracle),
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mono}}}); err == nil {
		t.Fatal("a mono-coloured spell is a legal target")
	}
	gold := b14CastAt(t, g, opp, "Gold Spell", "Instant", "", []string{"U", "R"}, game.CastSpellParams{})
	castCatalogSpell(t, g, "Neutralizing Blast", "Instant", frfNeutralizingBlastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: gold}})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, gold) != game.ZoneGraveyard || g.Stack.Contains(gold) {
		t.Error("the multicolored spell was not countered")
	}
}

func TestFrf5552B1ReachOfShadowsNeedsAColouredCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	elf := tmpB1Push(g, opp.ID, "Elf", "Creature — Elf", "", 1, 1, "G")
	if err := castCatalogSpellErr(t, g, "Reach of Shadows", "Instant", frfReachOfShadowsOracle, cardTarget(golem)); err == nil {
		t.Fatal("Reach of Shadows targeted a colorless creature")
	}
	castCatalogSpell(t, g, "Reach of Shadows", "Instant", frfReachOfShadowsOracle, cardTarget(elf))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elf) {
		t.Error("the coloured creature survived")
	}
}

func TestFrf5552B1ReturnToTheEarthDestroysAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	walker := b12Creature(g, opp.ID, "Walker", frfBearType, 2, 2)
	flyer := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}})
	if err := castCatalogSpellErr(t, g, "Return to the Earth", "Instant", frfReturnToTheEarthOracle, cardTarget(walker)); err == nil {
		t.Fatal("Return to the Earth targeted a creature without flying")
	}
	castCatalogSpell(t, g, "Return to the Earth", "Instant", frfReturnToTheEarthOracle, cardTarget(flyer))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer survived")
	}
}

// frfCastInCombat casts a catalog instant from the active player's hand
// without leaving the current step.
func frfCastInCombat(g *game.Game, me *game.Player, name, oracle string, targets []game.TargetRef) error {
	return g.CastSpell(me.ID, b13HandCard(me, name, "Instant", oracle), game.CastSpellParams{Targets: targets})
}

func TestFrf5552B1SandblastHitsOnlyACombatant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raider := b12Creature(g, me.ID, "Raider", frfBearType, 2, 6)
	idle := b12Creature(g, me.ID, "Idle", frfBearType, 2, 6)
	brAttack(t, g, raider)
	if err := frfCastInCombat(g, me, "Sandblast", frfSandblastOracle, cardTarget(idle)); err == nil {
		t.Fatal("Sandblast targeted a creature that isn't in combat")
	}
	if err := frfCastInCombat(g, me, "Sandblast", frfSandblastOracle, cardTarget(raider)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, raider).DamageMarked; got != 5 {
		t.Errorf("the attacker has %d damage, want 5", got)
	}
}

func TestFrf5552B1WhiskAwayTucksACombatant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	raider := b12Creature(g, me.ID, "Raider", frfBearType, 2, 2)
	idle := b12Creature(g, me.ID, "Idle", frfBearType, 2, 2)
	brAttack(t, g, raider)
	if err := frfCastInCombat(g, me, "Whisk Away", frfWhiskAwayOracle, cardTarget(idle)); err == nil {
		t.Fatal("Whisk Away targeted a creature that isn't in combat")
	}
	if err := frfCastInCombat(g, me, "Whisk Away", frfWhiskAwayOracle, cardTarget(raider)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := me.Library.Size(); n == 0 || me.Library.Cards[n-1].InstanceID != raider {
		t.Error("the attacker is not on top of its owner's library")
	}
}

func TestFrf5552B1CollateralDamageSacrificesACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	victim := b12Creature(g, me.ID, "Victim", frfBearType, 2, 2)
	id := b13HandCard(me, "Collateral Damage", "Instant", frfCollateralDamageOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if opp.Life != life-3 {
		t.Errorf("the target player is at %d, want %d", opp.Life, life-3)
	}
}

func TestFrf5552B1BolsterSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		counters, life         int
	}{
		{"Cached Defenses", "Sorcery", frfCachedDefensesOracle, 3, 0},
		{"Honor's Reward", "Instant", frfHonorsRewardOracle, 2, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			weak := b12Creature(g, me.ID, "Weak", frfBearType, 1, 1)
			b12Creature(g, me.ID, "Strong", frfBearType, 3, 3)
			life := me.Life
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			if n := frfPlusOnes(g, weak); n != tc.counters {
				t.Errorf("the least-toughness creature has %d +1/+1 counters, want %d", n, tc.counters)
			}
			if me.Life != life+tc.life {
				t.Errorf("life %d, want %d", me.Life, life+tc.life)
			}
		})
	}
}

func TestFrf5552B1EtherealAmbushManifestsTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	first := frfPushLibraryTop(g, me, "Second")
	second := frfPushLibraryTop(g, me, "First")
	castCatalogSpell(t, g, "Ethereal Ambush", "Instant", frfEtherealAmbushOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second} {
		c := findCardAnywhere(t, g, id)
		if !c.FaceDownIsPermanent() || c.FaceDownKind != game.FaceDownManifested {
			t.Errorf("%s is not a manifest: face-down %v (%q)", id, c.FaceDown, c.FaceDownKind)
		}
	}
}
