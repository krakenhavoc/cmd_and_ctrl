package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b227_cards_test.go — Commander Legends: Battle for Baldur's Gate
// Backgrounds and enchantments, GAP batch 227 (#5948). Each test plays the
// real catalog entry through the engine.

const (
	gap227HermitOracle       = "f898bf43-6a9c-46eb-b805-63cb12bab273"
	gap227OutlanderOracle    = "d0a33e9b-dc1e-4643-b179-851f3c71a99d"
	gap227JourneyOracle      = "a5f5e662-b32c-44f1-848d-0228f0fe1eb6"
	gap227ChefOracle         = "9eab87ff-82a0-4655-b6ea-00b27e882449"
	gap227DragonCultOracle   = "d1d7c5fa-a92a-4815-a558-6d3daf37f4c4"
	gap227ArchaeoOracle      = "6299d18e-a6a7-414c-85c0-d57a8e7d2627"
	gap227EntertainerOracle  = "702d7b7b-7935-4cb2-9cd7-06f638f5405f"
	gap227UrchinOracle       = "f52759f9-7aa5-4493-8fe3-7be9ec340dde"
	gap227BrawlerOracle      = "7c5cd912-eddb-47c8-a31b-8ad2af46abd3"
	gap227VisitorOracle      = "cbc93912-2c36-4554-9b42-1dcca059fbb5"
	gap227FontOracle         = "755230c3-0b89-4123-bed4-2f7729daa859"
	gap227CharlatanOracle    = "a9b08567-398a-46ba-b0e8-ade65fec83c6"
	gap227BackgroundTypeLine = "Legendary Enchantment — Background"
)

func TestGapB227EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap227HermitOracle:      "Cloakwood Hermit",
		gap227OutlanderOracle:   "Hardy Outlander",
		gap227JourneyOracle:     "Journey to the Lost City",
		gap227ChefOracle:        "Master Chef",
		gap227DragonCultOracle:  "Dragon Cultist",
		gap227ArchaeoOracle:     "Passionate Archaeologist",
		gap227EntertainerOracle: "Popular Entertainer",
		gap227UrchinOracle:      "Street Urchin",
		gap227BrawlerOracle:     "Tavern Brawler",
		gap227VisitorOracle:     "Feywild Visitor",
		gap227FontOracle:        "Font of Magic",
		gap227CharlatanOracle:   "Shameless Charlatan",
	}
	for oracle, name := range want {
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

func gap227Background(g *game.Game, owner uuid.UUID, name, oracle string) {
	effectstest.B21Push(g, owner, name, gap227BackgroundTypeLine, oracle, 0, 0)
}

// --- Cloakwood Hermit ---

func TestGapB227CloakwoodHermitMakesTwoTappedSquirrelsWhenACreatureCardWentToYourGraveyard(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		gap227Background(g, me.ID, "Cloakwood Hermit", gap227HermitOracle)
		effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if died {
			effectstest.KillCreature(t, g, me.ID, bear)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		n := effectstest.CountOnBattlefield(g, "Squirrel", me.ID)
		want := 0
		if died {
			want = 2
		}
		if n != want {
			t.Errorf("died=%v: %d Squirrels, want %d", died, n, want)
		}
		if died {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Squirrel" && !c.Tapped {
					t.Error("a Squirrel entered untapped")
				}
			}
		}
	}
}

// --- Dragon Cultist ---

func TestGapB227DragonCultistMakesADragonAfterFiveDamageFromOneSource(t *testing.T) {
	for _, dmg := range []int{4, 5} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		if me == opp {
			opp = g.Seats[2]
		}
		gap227Background(g, me.ID, "Dragon Cultist", gap227DragonCultOracle)
		effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
		src := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Bear", 5, 5)
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 1, 9)
		effectstest.Pr6Damage(t, g, src, victim, dmg)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		n := effectstest.CountOnBattlefield(g, "Dragon", me.ID)
		want := 0
		if dmg >= 5 {
			want = 1
		}
		if n != want {
			t.Errorf("%d damage: %d Dragons, want %d", dmg, n, want)
		}
	}
}

// --- Hardy Outlander ---

func TestGapB227HardyOutlanderPumpsAnotherCreatureByTheCommandersPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap227Background(g, me.ID, "Hardy Outlander", gap227OutlanderOracle)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(commander, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, commander, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, commander); p != 3 {
		t.Errorf("the commander is %d power, want 3", p)
	}
}

// --- Street Urchin ---

func TestGapB227StreetUrchinCommanderSacrificesAnotherCreatureToPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gap227Background(g, me.ID, "Street Urchin", gap227UrchinOracle)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "R")
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{commander}, Strict: true}); err == nil {
		t.Error("the commander sacrificed itself to its own cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if opp.Life != before-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, before-1)
	}
}

// --- Popular Entertainer, Feywild Visitor ---

func TestGapB227PopularEntertainerGoadsACreatureOfThePlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gap227Background(g, me.ID, "Popular Entertainer", gap227EntertainerOracle)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	other := effectstest.PushVanillaCreature(g, g.Seats[2].ID, "Other", 2, 2)
	gap227AttackUnblocked(t, g, commander, opp.ID)
	effectstest.HobAnswerTargets(t, g, me.ID, theirs, other)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, theirs); len(c.Goads) == 0 {
		t.Error("the creature of the player who was hit is not goaded")
	}
	if c := effectstest.B12Card(t, g, other); len(c.Goads) != 0 {
		t.Error("another player's creature was goaded")
	}
}

func TestGapB227FeywildVisitorMakesAFaerieDragonForNontokenDamageOnly(t *testing.T) {
	for _, token := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		gap227Background(g, me.ID, "Feywild Visitor", gap227VisitorOracle)
		attacker := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
		if token {
			attacker = effectstest.PushToken(g, me.ID, game.Card{Name: "Token Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2})
		}
		gap227AttackUnblocked(t, g, attacker, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if token {
			want = 0
		}
		if n := effectstest.CountOnBattlefield(g, "Faerie Dragon", me.ID); n != want {
			t.Errorf("token attacker %v: %d Faerie Dragons, want %d", token, n, want)
		}
	}
}

// gap227AttackUnblocked attacks the defender with the creature and runs the
// combat damage step with no blockers.
func gap227AttackUnblocked(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
}

// --- Tavern Brawler ---

func TestGapB227TavernBrawlerExilesAndPumpsByTheCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap227Background(g, me.ID, "Tavern Brawler", gap227BrawlerOracle)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for i := 0; i < 4 && g.Turn.ActiveSeat != 0; i++ {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", effectstest.B12ZoneOf(g, top))
	}
	if p, _ := effectstest.EffectivePTOf(t, g, commander); p != 6 {
		t.Errorf("the commander has power %d, want 6 (3, +3 for a mana value 3 card)", p)
	}
}

// --- Font of Magic ---

func TestGapB227FontOfMagicDiscountsInstantsAndSorceriesPerCommanderCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Font of Magic", "Enchantment", gap227FontOracle, false)
	g.WithWriteLock(func() { me.CommanderCasts = map[uuid.UUID]int{uuid.New(): 2} })
	effectstest.B06AddMana(me, "R", "R")
	if _, err := effectstest.CastFromHandErr(g, me, "Big Sorcery", "Sorcery", "{2}{R}{R}", "",
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("two commander casts should make {2}{R}{R} cost {R}{R}... got: %v", err)
	}
}

// --- Shameless Charlatan ---

func TestGapB227ShamelessCharlatanCommanderBecomesACopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gap227Background(g, me.ID, "Shameless Charlatan", gap227CharlatanOracle)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Big Theirs", 6, 6)
	effectstest.B06AddMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(commander), Strict: true}); err == nil {
		t.Error("the commander targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, commander); got.Name != "Big Theirs" {
		t.Errorf("the commander is %q, want a copy of Big Theirs", got.Name)
	}
}

// --- Passionate Archaeologist ---

func TestGapB227PassionateArchaeologistDamagesAnOpponentForASpellCastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if me == opp {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	gap227Background(g, me.ID, "Passionate Archaeologist", gap227ArchaeoOracle)
	effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Exiled Bolt", TypeLine: "Instant", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{}); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.B06AddMana(me, "R", "R")
	before := opp.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the Archaeologist's trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent at %d, want %d (2 damage for a mana value 2 spell)", opp.Life, before-2)
	}
}

// --- Master Chef ---

func TestGapB227MasterChefAddsACounterToOtherCreaturesAndTheCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap227Background(g, me.ID, "Master Chef", gap227ChefOracle)
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	// Read the commander through the layers so its grant is in force.
	effectstest.LayeredCard(t, g, commander)
	token := effectstest.B06GreenCreature(t, g, me.ID, "Token Bear")
	if n := effectstest.CounterOn(g, token, game.CounterPlusOne); n != 1 {
		t.Errorf("another creature entered with %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.CounterOn(g, commander, game.CounterPlusOne); n != 0 {
		t.Errorf("the commander already on the battlefield got %d counters", n)
	}
	cast := effectstest.B12PlayFromHand(t, g, "Hand Commander", "Legendary Creature — Elf", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Hand Commander" {
				_ = cast
			}
		}
	})
}

// --- Journey to the Lost City ---

// gap227Journey runs Journey to the Lost City's upkeep trigger on a table
// whose d20 shows `natural`. The roll is not the first draw from the
// source's stream (the exile before it draws too), so the table is built
// under each of 256 RNG keys until the roll the trigger actually made is the
// wanted one.
func gap227Journey(t *testing.T, natural int, types [4]string) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	for k := 0; k < 256; k++ {
		source := uuid.New()
		g := effectstest.NewCatalogGame(t)
		g.SetRNGKeyForTest(effectstest.RandomModesKey(byte(k)))
		me := g.Seats[0]
		effectstest.PushCatalogPermanentWithID(g, source, me.ID, "Journey to the Lost City", "Enchantment", gap227JourneyOracle, false)
		var top []uuid.UUID
		for i := 3; i >= 0; i-- {
			id := uuid.New()
			tl := types[i]
			cost := "{1}"
			if tl == "Land" {
				cost = ""
			}
			me.Library.PushTop(game.Card{InstanceID: id, Name: "Card " + tl, TypeLine: tl, ManaCost: cost, Owner: me.ID, Controller: me.ID})
			top = append(top, id)
		}
		effectstest.AdvanceToNextSeatsTurn(t, g)
		for i := 0; i < 4 && g.Turn.ActiveSeat != 0; i++ {
			effectstest.AdvanceToNextSeatsTurn(t, g)
		}
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		effectstest.PassPriorityAroundTable(t, g)
		rolls := effectstest.RollEventsOf(g)
		if len(rolls) == 1 && rolls[0].Amount == natural {
			return g, me, source, top
		}
	}
	t.Fatalf("no RNG key makes the upkeep roll a %d", natural)
	return nil, nil, uuid.Nil, nil
}

func gap227OnBattlefieldNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

func gap227InExile(g *game.Game, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if effectstest.B12ZoneOf(g, id) == game.ZoneExile {
			n++
		}
	}
	return n
}

func TestGapB227JourneyLowRollMayPutALandFromAmongTheExiledCards(t *testing.T) {
	g, me, _, top := gap227Journey(t, 5, [4]string{"Creature — Bear", "Land", "Sorcery", "Land"})
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no land choice after a roll of 5")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d lands offered, want 2", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pick.ChooseCards[0]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gap227OnBattlefieldNamed(g, me.ID, "Card Land"); n != 1 {
		t.Errorf("%d lands on the battlefield, want 1", n)
	}
	if n := gap227InExile(g, top); n != 3 {
		t.Errorf("%d of the four cards in exile, want 3", n)
	}
}

func TestGapB227JourneyMidRollMakesAWolfWithACounterPerCreatureCard(t *testing.T) {
	g, me, _, _ := gap227Journey(t, 15, [4]string{"Creature — Bear", "Creature — Elf", "Sorcery", "Land"})
	effectstest.PassPriorityAroundTable(t, g)
	wolf := effectstest.BattlefieldIDNamed(g, me.ID, "Wolf")
	if wolf == uuid.Nil {
		t.Fatal("no Wolf")
	}
	if n := effectstest.CounterOn(g, wolf, game.CounterPlusOne); n != 2 {
		t.Errorf("the Wolf has %d +1/+1 counters, want 2 (two creature cards)", n)
	}
	if c := effectstest.LayeredCard(t, g, wolf); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the Wolf is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestGapB227JourneyNaturalTwentyPutsEveryExiledPermanentCardOntoTheBattlefieldAndSacrificesIt(t *testing.T) {
	g, me, source, top := gap227Journey(t, 20, [4]string{"Creature — Bear", "Land", "Sorcery", "Artifact"})
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Card Creature — Bear", "Card Land", "Card Artifact"} {
		if n := gap227OnBattlefieldNamed(g, me.ID, name); n != 1 {
			t.Errorf("%d x %s on the battlefield, want 1", n, name)
		}
	}
	if n := gap227InExile(g, top); n != 1 {
		t.Errorf("%d of the four cards in exile, want 1 (the Sorcery stays)", n)
	}
	if g.Battlefield.Contains(source) {
		t.Error("Journey to the Lost City was not sacrificed")
	}
}
