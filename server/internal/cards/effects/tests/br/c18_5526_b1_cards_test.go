package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c18_5526_b1_cards_test.go — the first batch of Commander 2018 cards
// (#5526). Every test plays the real catalog entry through the engine.

const (
	c18CrashOfRhinoBeetlesOracle    = "61dd63b7-11dc-49db-8106-36d455a182be"
	c18LoyalDrakeOracle             = "daebf011-ec85-4cbd-9445-02ac59675c58"
	c18LoyalSubordinateOracle       = "76f405a1-242a-47e6-9aa6-261e49a85b45"
	c18EmpyrialStormOracle          = "0af5f9f4-04d6-4701-be1f-65d4a0940737"
	c18FuryStormOracle              = "371fa9e3-5432-4f2f-89d4-55061b0b4e57"
	c18GenesisStormOracle           = "80e72437-0c17-44c0-b0a7-eab075a64569"
	c18SkullStormOracle             = "6f34d70e-7a3a-43f7-b2c9-f89519804d3c"
	c18EverWatchingThresholdOracle  = "c6639b52-b51d-46e7-a83c-c4621b1cecaa"
	c18ThantisOracle                = "be543388-1e4f-412f-8f3b-dd826fd8e03d"
	c18TuvasaOracle                 = "6a83d976-8087-48d7-b6f5-63efb8329e99"
	c18NyleasColossusOracle         = "3d73dfd6-fd74-453f-b323-6505825caf4c"
	c18NightIncarnateOracle         = "abd1c415-8fd3-4852-92c6-4557453dfc41"
	c18EntreatTheDeadOracle         = "2de6c3d9-1759-40a2-99c6-8cbe17b4bcdd"
	c18VarinaOracle                 = "9859fbca-60c3-4a57-a377-51b56f03894f"
	c18SaheelisDirectiveOracle      = "1aee3a93-d064-4d10-9fe6-879a1f342c1f"
	c18TawnosOracle                 = "e835da40-f484-4d8b-98de-595cd23b51bc"
	c18TreasureNabberOracle         = "20a6d576-a5b2-43b4-8be9-75be001e567d"
	c18KestiaOracle                 = "21620bfc-8ef0-4e43-b675-d0bbad86e83d"
	c18HeavenlyBlademasterOracle    = "56940b6c-90ab-491b-a88a-4cd4a8adfbdb"
	c18IsolatedWatchtowerOracle     = "3893f320-47fd-49ec-a78c-80bfb607a279"
	c18GyrusOracle                  = "2deae472-3cfe-4d7b-b9cf-a045f07a002e"
	c18VarchildOracle               = "26ae37ab-91f8-4113-beb0-1154f9b62d28"
	c18SolRingOracle                = "6ad8011d-3471-4369-9d68-b264cc027487"
	c18ThresholdType                = "Enchantment"
	c18BearType                     = "Creature — Bear"
	c18ForestType                   = "Basic Land — Forest"
	c18CommanderName                = "My Commander"
	c18CommanderType                = "Legendary Creature — Bear"
	c18LieutenantCommanderPower     = 2
	c18LieutenantCommanderToughness = 2
)

// c18Commander puts a commander the player owns and controls onto the
// battlefield.
func c18Commander(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.M3cb1Push(g, owner, game.Card{Name: c18CommanderName, TypeLine: c18CommanderType,
		Power: c18LieutenantCommanderPower, Toughness: c18LieutenantCommanderToughness, IsCommander: true})
}

// c18GraveyardCreature puts a creature card with a power and toughness into
// p's graveyard.
func c18GraveyardCreature(p *game.Player, name string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: c18BearType,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

// c18LibraryTop puts a card on top of p's library.
func c18LibraryTop(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

// c18TokensWithSubtype counts the tokens `controller` controls with the
// subtype.
func c18TokensWithSubtype(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == controller && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// c18AnswerChooseCards answers the chooser's open choose-cards prompt.
func c18AnswerChooseCards(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, chooser)
	if c == nil {
		t.Fatalf("no choose_cards prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, chooser, ids); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// c18Settle passes priority around the table until the copies a storm
// trigger made have resolved too (none of the storms tested here has a
// target to choose again).
func c18Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// --- every card is in the catalog ----------------------------------------

func TestC18B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		c18CrashOfRhinoBeetlesOracle, c18LoyalDrakeOracle, c18LoyalSubordinateOracle, c18EmpyrialStormOracle,
		c18FuryStormOracle, c18GenesisStormOracle, c18SkullStormOracle, c18EverWatchingThresholdOracle,
		c18ThantisOracle, c18TuvasaOracle, c18NyleasColossusOracle, c18NightIncarnateOracle,
		c18EntreatTheDeadOracle, c18VarinaOracle, c18SaheelisDirectiveOracle, c18TawnosOracle,
		c18TreasureNabberOracle, c18KestiaOracle, c18HeavenlyBlademasterOracle, c18IsolatedWatchtowerOracle,
		c18GyrusOracle, c18VarchildOracle,
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

// --- printed keywords -----------------------------------------------------

func TestC18B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Crash of Rhino Beetles", "Creature — Insect", c18CrashOfRhinoBeetlesOracle, []string{"trample"}},
		{"Loyal Drake", "Creature — Drake", c18LoyalDrakeOracle, []string{"flying"}},
		{"Loyal Subordinate", "Creature — Zombie", c18LoyalSubordinateOracle, []string{"menace"}},
		{"Thantis, the Warweaver", "Legendary Creature — Spider", c18ThantisOracle, []string{"reach", "vigilance"}},
		{"Night Incarnate", "Creature — Elemental", c18NightIncarnateOracle, []string{"deathtouch"}},
		{"Heavenly Blademaster", "Creature — Angel", c18HeavenlyBlademasterOracle, []string{"flying", "double strike"}},
		{"Tawnos, Urza's Apprentice", "Legendary Creature — Human Artificer", c18TawnosOracle, []string{"haste"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- Crash of Rhino Beetles -----------------------------------------------

func TestC18B1CrashOfRhinoBeetlesGrowsAtTenLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beetles := effectstest.B12Push(g, me.ID, "Crash of Rhino Beetles", "Creature — Insect", c18CrashOfRhinoBeetlesOracle, 5, 5)
	for i := 0; i < 9; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", c18ForestType)
	}
	g.RunStateChecksForTest()
	if p, tough := effectstest.EffectivePTOf(t, g, beetles); p != 5 || tough != 5 {
		t.Fatalf("nine lands: %d/%d, want 5/5", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Forest", c18ForestType)
	g.RunStateChecksForTest()
	if p, tough := effectstest.EffectivePTOf(t, g, beetles); p != 15 || tough != 15 {
		t.Errorf("ten lands: %d/%d, want 15/15", p, tough)
	}
}

// --- the Loyal lieutenants ------------------------------------------------

func TestC18B1LoyalDrakeDrawsOnlyWithYourCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Loyal Drake", "Creature — Drake", c18LoyalDrakeOracle, 2, 2)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("without a commander: hand %d -> %d, want no draw", hand, got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B12Push(g2, me2.ID, "Loyal Drake", "Creature — Drake", c18LoyalDrakeOracle, 2, 2)
	c18Commander(g2, me2.ID)
	hand2 := me2.Hand.Size()
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != hand2+1 {
		t.Errorf("with your commander: hand %d -> %d, want one draw", hand2, got)
	}
}

func TestC18B1LoyalSubordinateDrainsEachOpponentWithYourCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Loyal Subordinate", "Creature — Zombie", c18LoyalSubordinateOracle, 3, 1)
	c18Commander(g, me.ID)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 3
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("%s: life %d -> %d, want %d", p.Name, lives[p.ID], p.Life, want)
		}
	}
}

// --- the commander storms ---------------------------------------------------

func TestC18B1EmpyrialStormCopiesPerCommanderCast(t *testing.T) {
	for _, casts := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		me.CommanderCasts[uuid.New()] = casts
		effectstest.CastCatalogSpell(t, g, "Empyrial Storm", "Sorcery", c18EmpyrialStormOracle, nil)
		c18Settle(t, g)
		if got, want := c18TokensWithSubtype(g, me.ID, "Angel"), 1+casts; got != want {
			t.Errorf("with %d commander casts: %d Angels, want %d", casts, got, want)
		}
	}
}

func TestC18B1FuryStormCopiesTheTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	storm := effectstest.CastCatalogSpell(t, g, "Empyrial Storm", "Sorcery", c18EmpyrialStormOracle, nil)
	effectstest.CastCatalogSpell(t, g, "Fury Storm", "Instant", c18FuryStormOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: storm}})
	c18Settle(t, g)
	if got := c18TokensWithSubtype(g, me.ID, "Angel"); got != 2 {
		t.Errorf("%d Angels, want 2 (the Empyrial Storm and Fury Storm's copy of it)", got)
	}
}

func TestC18B1GenesisStormMayPutTheFirstNonlandPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := c18LibraryTop(me, game.Card{Name: "Library Bear", TypeLine: c18BearType, Power: 2, Toughness: 2})
	landA := c18LibraryTop(me, game.Card{Name: "Top Forest", TypeLine: c18ForestType})
	landB := c18LibraryTop(me, game.Card{Name: "Topmost Forest", TypeLine: c18ForestType})
	effectstest.CastCatalogSpell(t, g, "Genesis Storm", "Sorcery", c18GenesisStormOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c18AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the nonland permanent card did not enter")
	}
	if !me.Library.Contains(landA) || !me.Library.Contains(landB) {
		t.Fatal("the revealed lands left the library")
	}
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	if top == landA || top == landB {
		t.Error("a revealed land stayed on top; they go to the bottom")
	}
}

func TestC18B1SkullStormEdictsAndHalvesTheOpponentsWhoCant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a, b := g.Seats[1], g.Seats[2]
	aBear := effectstest.B12Creature(g, a.ID, "A Bear", c18BearType, 2, 2)
	aLife, bLife, myLife := a.Life, b.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Skull Storm", "Sorcery", c18SkullStormOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, a.ID, aBear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aBear) {
		t.Error("the opponent with a creature did not sacrifice it")
	}
	if a.Life != aLife {
		t.Errorf("the opponent who sacrificed: life %d -> %d, want unchanged", aLife, a.Life)
	}
	if want := bLife - (bLife+1)/2; b.Life != want {
		t.Errorf("an opponent who can't: life %d -> %d, want %d", bLife, b.Life, want)
	}
	if me.Life != myLife {
		t.Errorf("the caster's life %d -> %d, want unchanged", myLife, me.Life)
	}
}

// --- Ever-Watching Threshold ----------------------------------------------

func TestC18B1EverWatchingThresholdDrawsOncePerAttackAtYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B12Push(g, me.ID, "Ever-Watching Threshold", c18ThresholdType, c18EverWatchingThresholdOracle, 0, 0)
	a := effectstest.PushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Raider B", 1, 1)
	c := effectstest.PushVanillaCreature(g, opp.ID, "Raider C", 1, 1)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{a, me.ID}, {b, me.ID}, {c, other.ID}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one draw for the attack", hand, got)
	}
}

func TestC18B1EverWatchingThresholdIgnoresAnAttackOnSomeoneElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B12Push(g, me.ID, "Ever-Watching Threshold", c18ThresholdType, c18EverWatchingThresholdOracle, 0, 0)
	a := effectstest.PushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, other.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no draw when you weren't attacked", hand, got)
	}
}

// --- Thantis, the Warweaver -----------------------------------------------

func TestC18B1ThantisGrowsForEachCreatureAttackingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 1)
	thantis := effectstest.B12Push(g, me.ID, "Thantis, the Warweaver", "Legendary Creature — Spider", c18ThantisOracle, 5, 5)
	a := effectstest.PushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Raider B", 1, 1)
	effectstest.DeclareAttack(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, thantis); got != 2 {
		t.Errorf("Thantis has %d +1/+1 counters, want 2", got)
	}
}

// --- Tuvasa the Sunlit ------------------------------------------------------

func TestC18B1TuvasaCountsEnchantmentsAndDrawsOnTheFirstEnchantmentSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tuvasa := effectstest.B12Push(g, me.ID, "Tuvasa the Sunlit", "Legendary Creature — Merfolk Shaman", c18TuvasaOracle, 1, 1)
	effectstest.B12Permanent(g, me.ID, "Shrine A", c18ThresholdType)
	effectstest.B12Permanent(g, me.ID, "Shrine B", c18ThresholdType)
	g.RunStateChecksForTest()
	if p, tough := effectstest.EffectivePTOf(t, g, tuvasa); p != 3 || tough != 3 {
		t.Fatalf("two enchantments: %d/%d, want 3/3", p, tough)
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Ever-Watching Threshold", c18ThresholdType, c18EverWatchingThresholdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("first enchantment spell: hand %d -> %d, want one draw", hand, got)
	}
	effectstest.CastCatalogSpell(t, g, "Ever-Watching Threshold", c18ThresholdType, c18EverWatchingThresholdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("second enchantment spell: hand %d, want no further draw", got)
	}
}

// --- Nylea's Colossus -------------------------------------------------------

func TestC18B1NyleasColossusDoublesTargetCreatureOnConstellation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Nylea's Colossus", "Enchantment Creature — Giant", c18NyleasColossusOracle, 6, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", c18BearType, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Ever-Watching Threshold", c18ThresholdType, c18EverWatchingThresholdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 6 {
		t.Errorf("the doubled Bear is %d/%d, want 4/6", p, tough)
	}
}

// --- Night Incarnate --------------------------------------------------------

func TestC18B1NightIncarnateShrinksEveryCreatureAsItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	night := effectstest.B12Push(g, me.ID, "Night Incarnate", "Creature — Elemental", c18NightIncarnateOracle, 3, 4)
	mine := effectstest.B12Creature(g, me.ID, "Big Bear", c18BearType, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", c18BearType, 3, 3)
	effectstest.B15Destroy(t, g, night)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3/3 survived -3/-3")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("the 4/4 is %d/%d, want 1/1", p, tough)
	}
	if spec, _ := effects.Lookup(c18NightIncarnateOracle); len(spec.AlternativeCosts) != 1 {
		t.Error("Night Incarnate has no evoke cost")
	}
}

// --- Entreat the Dead -------------------------------------------------------

func TestC18B1EntreatTheDeadReturnsXCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := c18GraveyardCreature(me, "Dead Bear A", 2, 2)
	b := c18GraveyardCreature(me, "Dead Bear B", 3, 3)
	id := effectstest.B13HandCard(me, "Entreat the Dead", "Sorcery", c18EntreatTheDeadOracle)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("the two creature cards did not return to the battlefield")
	}
}

// --- Varina, Lich Queen -----------------------------------------------------

func TestC18B1VarinaLootsAndGainsPerAttackingZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	varina := effectstest.B12Push(g, me.ID, "Varina, Lich Queen", "Legendary Creature — Zombie Wizard", c18VarinaOracle, 4, 4)
	zombie := effectstest.B12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", c18BearType, 2, 2)
	lib, life := me.Library.Size(), me.Life
	effectstest.DeclareAttack(t, g, opp.ID, varina, zombie, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib-2 {
		t.Errorf("library %d -> %d, want two draws for two attacking Zombies", lib, got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestC18B1VarinaExilesTwoGraveyardCardsForATappedZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	varina := effectstest.B12Push(g, me.ID, "Varina, Lich Queen", "Legendary Creature — Zombie Wizard", c18VarinaOracle, 4, 4)
	a := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	b := effectstest.PushGraveyardCardTyped(me, "Old Land", "Land")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, varina, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := c18TokensWithSubtype(g, me.ID, "Zombie"); got != 1 {
		t.Fatalf("%d Zombie tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Zombie") && !c.Tapped {
			t.Error("the Zombie token entered untapped")
		}
	}
}

// --- Saheeli's Directive ----------------------------------------------------

func TestC18B1SaheelisDirectivePutsCheapArtifactsAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := c18LibraryTop(me, game.Card{Name: "Cheap Relic", TypeLine: "Artifact", ManaCost: "{2}"})
	dear := c18LibraryTop(me, game.Card{Name: "Dear Relic", TypeLine: "Artifact", ManaCost: "{5}"})
	land := c18LibraryTop(me, game.Card{Name: "Forest", TypeLine: c18ForestType})
	id := effectstest.B13HandCard(me, "Saheeli's Directive", "Sorcery", c18SaheelisDirectiveOracle)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c18AnswerChooseCards(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the artifact with mana value 2 did not enter")
	}
	if !me.Graveyard.Contains(dear) || !me.Graveyard.Contains(land) {
		t.Error("the revealed cards not put onto the battlefield are not in the graveyard")
	}
}

// --- Tawnos, Urza's Apprentice ----------------------------------------------

func TestC18B1TawnosOffersOnlyAbilitiesFromArtifactSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tawnos, Urza's Apprentice", "Legendary Creature — Human Artificer", c18TawnosOracle, false)
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", c18BearType, "", false)
	if err := g.AnnounceTrigger(me.ID, relic, game.AbilityParams{Label: "an artifact trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.AnnounceTrigger(me.ID, bear, game.AbilityParams{Label: "a creature trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	offered := effectstest.AcLegalAbilityTargets(g, me.ID, c18TawnosOracle, 0)
	if !offered[effectstest.AcItemLabelled(g, "an artifact trigger")] {
		t.Error("an ability from an artifact source is not offered")
	}
	if offered[effectstest.AcItemLabelled(g, "a creature trigger")] {
		t.Error("an ability from a non-artifact source is offered")
	}
}

// --- Treasure Nabber --------------------------------------------------------

func TestC18B1TreasureNabberStealsAnArtifactAnOpponentTapsForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Treasure Nabber", "Creature — Goblin Rogue", c18TreasureNabberOracle, 3, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Sol Ring", "Artifact", c18SolRingOracle, 0, 0)
	mine := effectstest.B12Push(g, me.ID, "Sol Ring", "Artifact", c18SolRingOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("my Sol Ring: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, mine) != me.ID {
		t.Fatal("my own artifact changed hands")
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("their Sol Ring: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, theirs) != me.ID {
		t.Error("the opponent's artifact tapped for mana was not taken")
	}
}

// --- Kestia, the Cultivator -------------------------------------------------

func TestC18B1KestiaDrawsForEnchantedAndEnchantmentAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kestia := effectstest.B12Push(g, me.ID, "Kestia, the Cultivator", "Legendary Enchantment Creature — Nymph", c18KestiaOracle, 4, 4)
	enchanted := effectstest.B12Creature(g, me.ID, "Enchanted Bear", c18BearType, 2, 2)
	effectstest.B15Aura(g, me.ID, enchanted)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", c18BearType, 2, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, kestia, enchanted, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want two draws (Kestia and the enchanted Bear)", hand, got)
	}
}

// --- Heavenly Blademaster ---------------------------------------------------

func TestC18B1HeavenlyBlademasterAttachesAndAnthems(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	bear := effectstest.B12Creature(g, me.ID, "Bear", c18BearType, 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heavenly Blademaster", TypeLine: "Creature — Angel",
		OracleID: c18HeavenlyBlademasterOracle, Power: 3, Toughness: 6, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c18AnswerChooseCards(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	attached := false
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == sword && c.IsAttachedTo(id) {
			attached = true
		}
	}
	if !attached {
		t.Fatal("the Equipment was not attached to the Blademaster")
	}
	g.RunStateChecksForTest()
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the other creature is %d/%d, want 3/3 with one attachment", p, tough)
	}
}

// --- Isolated Watchtower ----------------------------------------------------

func TestC18B1IsolatedWatchtowerNeedsTwoMoreLandsAndFetchesABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tower := effectstest.PushCatalogPermanent(g, me.ID, "Isolated Watchtower", "Land", c18IsolatedWatchtowerOracle, false)
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tower, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated while the opponent has only one more land")
	}
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	forest := c18LibraryTop(me, game.Card{Name: "Forest", TypeLine: c18ForestType})
	if err := g.ActivateCatalogAbility(me.ID, tower, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with the opponent two lands ahead: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the revealed basic land did not enter")
	}
	if !effectstest.BattlefieldTapped(g, forest) {
		t.Error("the basic land entered untapped")
	}
}

// --- Gyrus, Waker of Corpses ------------------------------------------------

func TestC18B1GyrusExilesALesserCreatureCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gyrus := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gyrus, Waker of Corpses",
		TypeLine: "Legendary Creature — Hydra", OracleID: c18GyrusOracle, Power: 0, Toughness: 0,
		Counters: map[string]int{game.CounterPlusOne: 4}, Owner: me.ID, Controller: me.ID})
	small := c18GraveyardCreature(me, "Small Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, gyrus)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(small); z == nil || z.Kind != game.ZoneExile {
		t.Error("the creature card with lesser power was not exiled")
	}
}

// --- Varchild, Betrayer of Kjeldor ------------------------------------------

func TestC18B1VarchildGivesSurvivorsThatCantBlockAndTakesThemBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	varchild := effectstest.B12Push(g, me.ID, "Varchild, Betrayer of Kjeldor", "Legendary Creature — Human Knight", c18VarchildOracle, 3, 3)
	effectstest.AttackWith(t, g, opp.ID, varchild)
	effectstest.PassPriorityAroundTable(t, g)
	if got := c18TokensWithSubtype(g, opp.ID, "Survivor"); got != 3 {
		t.Fatalf("the damaged player has %d Survivors, want 3", got)
	}
	var survivor uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Survivor") {
			survivor = c.InstanceID
		}
	}
	cantBlock := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == survivor {
				cantBlock = c.Effective().Restrictions&game.CantBlock != 0
			}
		}
	})
	if !cantBlock {
		t.Error("an opponent's Survivor can block")
	}
	effectstest.B15Destroy(t, g, varchild)
	if effectstest.ControllerOf(t, g, survivor) != me.ID {
		t.Error("Varchild leaving did not take the Survivors")
	}
}
