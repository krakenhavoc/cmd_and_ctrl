package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a9_cards_test.go — The Hobbit / The Hobbit Commander, half A, ninth
// batch (#4454): two replacement-effect lords, a Wolf with a can't-attack
// clause, a Goblin lord and two Companys. Each test plays the real catalog
// entry and asserts what the card's text says.

const (
	hobA9BilboOracle       = "d273f87d-1271-49d2-a628-2d1bd888f56c"
	hobA9BardOracle        = "d05db2c1-a19a-4803-8e8a-fa2f9b798181"
	hobA9ChiefWargOracle   = "372d7086-b7dc-4b6d-bdf6-e9ed9c995456"
	hobA9GreatGoblinOracle = "83dcfac0-6efd-4e37-9402-15f9889e84e1"
	hobA9ThranduilOracle   = "2c195485-83f9-424f-854a-50c4f5b04182"
	hobA9DainOracle        = "ae03782f-8f05-4ce2-a049-005b1b688faf"
)

func TestHobA9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA9BilboOracle:       "Bilbo, Fellow Conspirator",
		hobA9BardOracle:        "Bard, King of Dale",
		hobA9ChiefWargOracle:   "Chief Warg's Company",
		hobA9GreatGoblinOracle: "The Great Goblin",
		hobA9ThranduilOracle:   "Thranduil's Company",
		hobA9DainOracle:        "Dáin's Company",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// --- Bilbo, Fellow Conspirator -----------------------------------------------

func TestHobA9BilboMakesAFoodAndATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, hobA9BilboOracle, "Bilbo, Fellow Conspirator", "Legendary Creature — Halfling Citizen", me)

	makeTokens(t, g, me, FoodToken(), 1)

	if got := onBattlefieldNamed(g, "Food"); got != 1 {
		t.Errorf("Foods = %d, want 1", got)
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

// The count rides along: two Foods are two of each.
func TestHobA9BilboKeepsTheCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, hobA9BilboOracle, "Bilbo, Fellow Conspirator", "Legendary Creature — Halfling Citizen", me)

	makeTokens(t, g, me, FoodToken(), 2)

	if food, treasure := onBattlefieldNamed(g, "Food"), onBattlefieldNamed(g, "Treasure"); food != 2 || treasure != 2 {
		t.Errorf("two Foods came out as %d Food and %d Treasure, want 2 and 2", food, treasure)
	}
}

// Only a Food is replaced: a Treasure or a Clue is made as printed.
func TestHobA9BilboLeavesOtherTokensAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushTokenReplacementCard(g, hobA9BilboOracle, "Bilbo, Fellow Conspirator", "Legendary Creature — Halfling Citizen", me)

	makeTokens(t, g, me, ClueToken(), 1)
	makeTokens(t, g, me, TreasureToken(), 1)

	if got := onBattlefieldNamed(g, "Clue"); got != 1 {
		t.Errorf("Clues = %d, want 1", got)
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1 (a Treasure is not a Food)", got)
	}
	if got := onBattlefieldNamed(g, "Food"); got != 0 {
		t.Errorf("Foods = %d, want 0", got)
	}
}

// "If YOU would create": an opponent's Food is just a Food.
func TestHobA9BilboOnlyWatchesItsControllersTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	pushTokenReplacementCard(g, hobA9BilboOracle, "Bilbo, Fellow Conspirator", "Legendary Creature — Halfling Citizen", me)

	makeTokens(t, g, opp, FoodToken(), 1)

	if got := onBattlefieldNamed(g, "Food"); got != 1 {
		t.Errorf("Foods = %d, want 1", got)
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 0 {
		t.Errorf("Treasures = %d, want 0", got)
	}
}

// Beside Academy Manufactor the Food group is one more kind the second
// window can rewrite, and the player orders them (CR 616.1). Bilbo first
// makes a Food and a Treasure, and the Manufactor then makes one of each of
// every group: Clue, Food and Treasure for the Food, and the same for the
// Treasure. Manufactor first makes a Clue, a Food and a Treasure and Bilbo
// then adds a Treasure to the Food.
func TestHobA9BilboAndAcademyManufactorAskForTheOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	bilbo := pushTokenReplacementCard(g, hobA9BilboOracle, "Bilbo, Fellow Conspirator", "Legendary Creature — Halfling Citizen", me)
	am := pushTokenReplacementCard(g, academyManufactorOracle, "Academy Manufactor", "Artifact Creature — Assembly-Worker", me)

	makeTokens(t, g, me, FoodToken(), 1)
	prompt := g.PendingChoices[0]
	if prompt.Kind != game.PendingChoiceReplacementOrder || prompt.Chooser != me {
		t.Fatalf("prompt = %q for %s, want a replacement order for the creator", prompt.Kind, prompt.Chooser)
	}
	first := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, am)
	second := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, bilbo)
	if err := g.ResolveReplacementOrder(prompt.ID, me, []game.ReplacementEffectID{first, second}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if clue, food, treasure := onBattlefieldNamed(g, "Clue"), onBattlefieldNamed(g, "Food"), onBattlefieldNamed(g, "Treasure"); clue != 1 || food != 1 || treasure != 2 {
		t.Errorf("Manufactor then Bilbo: %d Clue, %d Food, %d Treasure, want 1, 1 and 2", clue, food, treasure)
	}
}

// --- Bard, King of Dale ------------------------------------------------------

func TestHobA9BardHasReachAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	bard := b12Push(g, g.Seats[0].ID, "Bard, King of Dale", "Legendary Creature — Human Noble Archer", hobA9BardOracle, 3, 5)
	for _, kw := range []string{"reach", "vigilance"} {
		if !effectiveAbilitiesContain(t, g, bard, kw) {
			t.Errorf("Bard should have %s", kw)
		}
	}
}

func TestHobA9BardDoublesTheTokensYouCreate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Bard, King of Dale", "Legendary Creature — Human Noble Archer", hobA9BardOracle, false)

	makeTokens(t, g, me.ID, RedGoblinToken(), 2)
	if got := onBattlefieldNamed(g, "Goblin"); got != 4 {
		t.Errorf("two Goblins came out as %d, want 4", got)
	}
	makeTokens(t, g, opp.ID, RedGoblinToken(), 1)
	if got := onBattlefieldNamed(g, "Goblin"); got != 5 {
		t.Errorf("after an opponent made one, %d Goblins, want 5 (4 + 1, not doubled)", got)
	}
}

func TestHobA9BardDrawsTwoInsteadOutsideTheDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Bard, King of Dale", "Legendary Creature — Human Noble Archer", hobA9BardOracle, false)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })

	if got := drawFor(t, g, me, 1); got != 2 {
		t.Errorf("one draw drew %d cards, want 2", got)
	}
}

// The turn-based draw of the draw step is the one exempt card; the next draw
// is doubled. An opponent's draws are never doubled.
func TestHobA9BardExemptsOnlyTheFirstDrawOfYourDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Bard, King of Dale", "Legendary Creature — Human Noble Archer", hobA9BardOracle, false)
	before := me.Hand.Size()
	advanceToDrawStepOfSeat(t, g, 1)
	if me.Hand.Size() != before+1 {
		t.Fatalf("the draw step's first draw gave %d cards, want 1", me.Hand.Size()-before)
	}
	if got := drawFor(t, g, me, 1); got != 2 {
		t.Errorf("the next draw in the draw step drew %d cards, want 2", got)
	}
	if got := drawFor(t, g, opp, 1); got != 1 {
		t.Errorf("an opponent's draw drew %d cards, want 1", got)
	}
}

// --- Chief Warg's Company ----------------------------------------------------

func hobA9ChiefWarg(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Chief Warg's Company", "Creature — Wolf", hobA9ChiefWargOracle, 5, 3)
}

// hobA9CanAttack declares id as an attacker at the controller's declare
// attackers step and says whether the engine took it.
func hobA9CanAttack(t *testing.T, g *game.Game, id uuid.UUID) error {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	return g.DeclareAttacker(id, g.Seats[1].ID)
}

func TestHobA9ChiefWargHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	if !effectiveAbilitiesContain(t, g, hobA9ChiefWarg(g, g.Seats[0].ID), "trample") {
		t.Error("Chief Warg's Company should have trample")
	}
}

func TestHobA9ChiefWargCantAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	warg := hobA9ChiefWarg(g, g.Seats[0].ID)
	if err := hobA9CanAttack(t, g, warg); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("alone: err = %v, want ErrCantAttack", err)
	}
}

// One other Wolf is not enough, and neither is the Company itself nor a
// Wolf an opponent controls nor a creature that is not a Wolf.
func TestHobA9ChiefWargNeedsTwoOtherWolvesOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	warg := hobA9ChiefWarg(g, me)
	b12Creature(g, me, "Pack Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, me, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp, "Their Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, opp, "Their Other Wolf", "Creature — Wolf", 2, 2)
	if err := hobA9CanAttack(t, g, warg); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("one other Wolf of mine: err = %v, want ErrCantAttack", err)
	}
}

func TestHobA9ChiefWargAttacksWithTwoOtherWolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	warg := hobA9ChiefWarg(g, me)
	b12Creature(g, me, "Pack Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, me, "Other Pack Wolf", "Creature — Wolf", 2, 2)
	if err := hobA9CanAttack(t, g, warg); err != nil {
		t.Fatalf("with two other Wolves: %v", err)
	}
}

// The restriction is read live: the second Wolf leaving takes the attack
// away again.
func TestHobA9ChiefWargRestrictionFollowsTheWolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	warg := hobA9ChiefWarg(g, me)
	b12Creature(g, me, "Pack Wolf", "Creature — Wolf", 2, 2)
	second := b12Creature(g, me, "Other Pack Wolf", "Creature — Wolf", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(second) })
	g.RunStateChecksForTest()
	if err := hobA9CanAttack(t, g, warg); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("after a Wolf died: err = %v, want ErrCantAttack", err)
	}
}

// Two Companys are Wolves for each other, so two of them still lack the
// second other Wolf each needs.
func TestHobA9TwoChiefWargsDoNotEnableEachOther(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	one := hobA9ChiefWarg(g, me)
	hobA9ChiefWarg(g, me)
	if err := hobA9CanAttack(t, g, one); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("two Companys: err = %v, want ErrCantAttack", err)
	}
}

func TestHobA9ChiefWargMakesAWolfEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobA9ChiefWarg(g, me.ID)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := namedOnBattlefield(g, "Wolf", me.ID); got != 1 {
		t.Fatalf("after one upkeep: %d Wolf tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wolf" {
			if !c.IsToken() || c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 1 || c.Colors[0] != "G" {
				t.Errorf("the Wolf is %+v, want a 2/2 green token", c)
			}
		}
	}
}

// Only on its controller's upkeep.
func TestHobA9ChiefWargMakesNoWolfOnAnOpponentsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	hobA9ChiefWarg(g, g.Seats[0].ID)
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := namedOnBattlefield(g, "Wolf", g.Seats[0].ID); got != 0 {
		t.Errorf("%d Wolf tokens on an opponent's upkeep, want 0", got)
	}
}

// --- The Great Goblin --------------------------------------------------------

func hobA9GreatGoblinTable(t *testing.T) (g *game.Game, me, opp *game.Player, goblin uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	goblin = b12Push(g, me.ID, "The Great Goblin", "Legendary Creature — Goblin Noble", hobA9GreatGoblinOracle, 3, 2)
	return g, me, opp, goblin
}

// hobA9SettleAgainst answers the trigger's target prompt with `opp` and lets
// the stack empty.
func hobA9SettleAgainst(t *testing.T, g *game.Game, me, opp uuid.UUID) {
	t.Helper()
	pick := hobA7PlayerRef(opp)
	assistCardSettle(t, g, me, &pick, true)
}

func TestHobA9GreatGoblinDealsTwoWhenYouPutCountersOnAGoblin(t *testing.T) {
	g, me, opp, goblin := hobA9GreatGoblinTable(t)
	other := b12Creature(g, me.ID, "Other Goblin", "Creature — Goblin Scout", 1, 1)
	for _, id := range []uuid.UUID{goblin, other} {
		life := opp.Life
		if err := g.AddCounterBy(me.ID, id, "+1/+1", 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
		hobA9SettleAgainst(t, g, me.ID, opp.ID)
		if opp.Life != life-2 {
			t.Errorf("opponent life = %d, want %d", opp.Life, life-2)
		}
	}
}

func TestHobA9GreatGoblinCountsOrcsAndArmiesAndAnyKindOfCounter(t *testing.T) {
	g, me, opp, _ := hobA9GreatGoblinTable(t)
	orc := b12Creature(g, me.ID, "An Orc", "Creature — Orc Warrior", 2, 2)
	army := b12Creature(g, me.ID, "An Army", "Creature — Army", 0, 0)
	for _, c := range []struct {
		id   uuid.UUID
		kind string
	}{{orc, "+1/+1"}, {army, "+1/+1"}, {orc, "shield"}} {
		life := opp.Life
		if err := g.AddCounterBy(me.ID, c.id, c.kind, 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
		hobA9SettleAgainst(t, g, me.ID, opp.ID)
		if opp.Life != life-2 {
			t.Errorf("a %s counter: opponent life = %d, want %d", c.kind, opp.Life, life-2)
		}
	}
}

func TestHobA9GreatGoblinIgnoresOtherCreaturesAndOtherPlayers(t *testing.T) {
	g, me, opp, _ := hobA9GreatGoblinTable(t)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin Scout", 1, 1)
	mine := b12Creature(g, me.ID, "My Goblin", "Creature — Goblin Scout", 1, 1)
	life := opp.Life

	// A Bear is not a Goblin, an Orc or an Army.
	if err := g.AddCounterBy(me.ID, bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	// "You control": a counter on the opponent's Goblin.
	if err := g.AddCounterBy(me.ID, theirs, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	// "You put": the opponent puts one on my Goblin.
	if err := g.AddCounterBy(opp.ID, mine, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || opp.Life != life {
		t.Errorf("a trigger fired for none of these: life %d -> %d, %d prompts", life, opp.Life, len(g.PendingChoices))
	}
}

// One proliferate puts a counter of each kind on the same Goblin: that is
// one "one or more counters", so one trigger and 2 damage, not 4.
func TestHobA9GreatGoblinTriggersOncePerPermanentForOneProliferate(t *testing.T) {
	g, me, opp, goblin := hobA9GreatGoblinTable(t)
	for _, kind := range []string{"+1/+1", "shield"} {
		if err := g.AddCounterBy(me.ID, goblin, kind, 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
		hobA9SettleAgainst(t, g, me.ID, opp.ID)
	}
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{goblin}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	hobA9SettleAgainst(t, g, me.ID, opp.ID)
	if opp.Life != life-2 {
		t.Errorf("opponent life = %d, want %d (one trigger)", opp.Life, life-2)
	}
}

func TestHobA9GreatGoblinExilesTheTopCardWhenAnotherGoblinDies(t *testing.T) {
	g, me, _, _ := hobA9GreatGoblinTable(t)
	top := seedTopOfLibrary(me, game.Card{Name: "Impulse Land", TypeLine: "Basic Land — Forest"})
	victim := b12Creature(g, me.ID, "Doomed Goblin", "Creature — Goblin Scout", 1, 1)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	g.RunStateChecksForTest()
	assistCardSettle(t, g, me.ID, nil, true)

	if me.Library.Contains(top) || !g.Exile.Contains(top) {
		t.Fatal("the top card should be exiled")
	}
	if grantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("the exiled card carries no permission to play it")
	}
}

func TestHobA9GreatGoblinCountsOrcsAndArmiesDying(t *testing.T) {
	g, me, _, _ := hobA9GreatGoblinTable(t)
	for _, typeLine := range []string{"Creature — Orc Warrior", "Creature — Army"} {
		top := seedTopOfLibrary(me, game.Card{Name: "Impulse Card", TypeLine: "Instant"})
		victim := b12Creature(g, me.ID, "Doomed", typeLine, 1, 1)
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
		g.RunStateChecksForTest()
		assistCardSettle(t, g, me.ID, nil, true)
		if !g.Exile.Contains(top) {
			t.Errorf("a dying %q did not exile the top card", typeLine)
		}
	}
}

// Not for itself, not for another creature type, not for an opponent's
// Goblin, and not when a Goblin leaves without dying.
func TestHobA9GreatGoblinExilesNothingForTheWrongDeaths(t *testing.T) {
	g, me, opp, goblin := hobA9GreatGoblinTable(t)
	top := seedTopOfLibrary(me, game.Card{Name: "Impulse Card", TypeLine: "Instant"})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin Scout", 1, 1)
	exiled := b12Creature(g, me.ID, "Exiled Goblin", "Creature — Goblin Scout", 1, 1)

	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.DestroyPermanentForEffect(theirs)
		_ = g.ExileCardForEffect(exiled)
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !me.Library.Contains(top) || g.Exile.Contains(top) {
		t.Fatal("the top card was exiled for a death that is not a Goblin, Orc or Army of mine")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(goblin) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("The Great Goblin's own death exiled the top card; it says another")
	}
}

// --- Thranduil's Company -----------------------------------------------------

func hobA9Thranduil(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Thranduil's Company", "Creature — Elf Soldier", hobA9ThranduilOracle, 3, 4)
}

func TestHobA9ThranduilExtraLandNeedsAnotherElf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	base := g.LandDropsRemainingFor(me.ID)
	hobA9Thranduil(g, me.ID)
	if got := g.LandDropsRemainingFor(me.ID); got != base {
		t.Fatalf("Thranduil's Company alone: %d land drops, want %d (it is not its own other Elf)", got, base)
	}
	b12Creature(g, opp.ID, "Their Elf", "Creature — Elf Druid", 1, 1)
	if got := g.LandDropsRemainingFor(me.ID); got != base {
		t.Fatalf("with only an opponent's Elf: %d land drops, want %d", got, base)
	}
	b12Creature(g, me.ID, "Not An Elf", "Creature — Bear", 1, 1)
	if got := g.LandDropsRemainingFor(me.ID); got != base {
		t.Fatalf("with a Bear: %d land drops, want %d", got, base)
	}
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf Druid", 1, 1)
	if got := g.LandDropsRemainingFor(me.ID); got != base+1 {
		t.Fatalf("with another Elf: %d land drops, want %d", got, base+1)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(elf) })
	g.RunStateChecksForTest()
	if got := g.LandDropsRemainingFor(me.ID); got != base {
		t.Errorf("after the other Elf died: %d land drops, want %d", got, base)
	}
}

// The tile says the condition, and does not also print the unconditional
// Exploration sentence.
func TestHobA9ThranduilTileStatesTheCondition(t *testing.T) {
	g := newCatalogGame(t)
	c := findBattlefieldCardByID(g, hobA9Thranduil(g, g.Seats[0].ID))
	want := "As long as you control another Elf, you may play an additional land on each of your turns."
	var statics []string
	for _, row := range game.AbilityRowsOnBattlefield(*c) {
		if row.Kind == game.AbilityRowStatic {
			statics = append(statics, row.Label)
		}
	}
	if len(statics) != 1 || statics[0] != want {
		t.Errorf("static rows = %q, want just %q", statics, want)
	}
}

// A changeling is an Elf, and two Companys are each other's other Elf.
func TestHobA9ThranduilCountsChangelingsAndOtherCompanys(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	base := g.LandDropsRemainingFor(me.ID)
	hobA9Thranduil(g, me.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Changeling", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{"changeling"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if got := g.LandDropsRemainingFor(me.ID); got != base+1 {
		t.Fatalf("with a changeling: %d land drops, want %d", got, base+1)
	}
	hobA9Thranduil(g, me.ID)
	if got := g.LandDropsRemainingFor(me.ID); got != base+2 {
		t.Errorf("with two Companys and a changeling: %d land drops, want %d", got, base+2)
	}
}

// The extra drop is real: a second land can be played with another Elf and
// is refused without one.
func TestHobA9ThranduilLetsYouPlayASecondLand(t *testing.T) {
	for _, withElf := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		thranduil := hobA9Thranduil(g, me.ID)
		if withElf {
			b12Creature(g, me.ID, "My Elf", "Creature — Elf Druid", 1, 1)
		}
		var lands []uuid.UUID
		for i := 0; i < 2; i++ {
			id := uuid.New()
			lands = append(lands, id)
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		}
		advanceToMain(t, g)
		if err := g.CastSpell(me.ID, lands[0], game.CastSpellParams{}); err != nil {
			t.Fatalf("first land: %v", err)
		}
		onThranduil := game.TargetRef{Kind: game.TargetCard, ID: thranduil}
		assistCardSettle(t, g, me.ID, &onThranduil, true)
		err := g.CastSpell(me.ID, lands[1], game.CastSpellParams{})
		if withElf && err != nil {
			t.Errorf("with another Elf, the second land: %v", err)
		}
		if !withElf && err == nil {
			t.Error("without another Elf, a second land was played")
		}
	}
}

func TestHobA9ThranduilLandfallPutsTwoCountersAndGivesVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thranduil := hobA9Thranduil(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the landfall trigger asked for no target")
	}
	if !hasID(p.PickTargetCards, thranduil) || !hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, theirs) {
		t.Errorf("legal targets %v: creatures I control only", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if got := hobA7Counters(t, g, bear); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", got)
	}
	if !effectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("the Bear should have vigilance")
	}
	if got := hobA7Counters(t, g, thranduil); got != 0 {
		t.Errorf("Thranduil's Company has %d counters, want 0", got)
	}
}

// Vigilance only lasts until end of turn.
func TestHobA9ThranduilVigilanceEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thranduil := hobA9Thranduil(g, me.ID)

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, thranduil)
	}
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, thranduil, "vigilance") {
		t.Fatal("no vigilance this turn")
	}
	advanceToNextSeatsTurn(t, g)
	if effectiveAbilitiesContain(t, g, thranduil, "vigilance") {
		t.Error("vigilance lasted past the turn")
	}
	if got := hobA7Counters(t, g, thranduil); got != 2 {
		t.Errorf("the counters are permanent: %d, want 2", got)
	}
}

// --- Dáin's Company ----------------------------------------------------------

func hobA9Dain(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Dáin's Company", "Creature — Dwarf Warrior", hobA9DainOracle, 2, 2)
}

func TestHobA9DainHasLifelinkOnlyWhileYouControlAnotherDwarf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	dain := hobA9Dain(g, me)
	if effectiveAbilitiesContain(t, g, dain, "lifelink") {
		t.Fatal("alone, Dáin's Company has lifelink")
	}
	b12Creature(g, opp, "Their Dwarf", "Creature — Dwarf Warrior", 1, 1)
	b12Creature(g, me, "My Bear", "Creature — Bear", 1, 1)
	if effectiveAbilitiesContain(t, g, dain, "lifelink") {
		t.Fatal("an opponent's Dwarf and a Bear gave it lifelink")
	}
	friend := b12Creature(g, me, "My Dwarf", "Creature — Dwarf Warrior", 1, 1)
	if !effectiveAbilitiesContain(t, g, dain, "lifelink") {
		t.Fatal("with another Dwarf it has no lifelink")
	}
	if effectiveAbilitiesContain(t, g, friend, "lifelink") {
		t.Error("the other Dwarf gained lifelink; only Dáin's Company has it")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(friend) })
	g.RunStateChecksForTest()
	if effectiveAbilitiesContain(t, g, dain, "lifelink") {
		t.Error("lifelink stayed after the other Dwarf died")
	}
}

// hobA9DainLook seeds a library whose top four are, top first: a Dwarf, an
// Equipment, a Bear and a land, over a card that stays put, then casts
// Dáin's Company and returns the pieces.
func hobA9DainLook(t *testing.T) (g *game.Game, me *game.Player, dwarf, equipment, bear, land, deep uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	deep, dwarf, equipment, bear, land = uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, c := range []struct {
		id             uuid.UUID
		name, typeLine string
	}{
		{deep, "Deep Dwarf", "Creature — Dwarf Scout"}, // the fifth card: never looked at
		{land, "A Land", "Basic Land — Mountain"},
		{bear, "A Bear", "Creature — Bear"},
		{equipment, "A Sword", "Artifact — Equipment"},
		{dwarf, "A Dwarf", "Creature — Dwarf Warrior"},
	} {
		me.Library.PushTop(game.Card{InstanceID: c.id, Name: c.name, TypeLine: c.typeLine, Owner: me.ID, Controller: me.ID})
	}
	castHoldCreature(t, g, "Dáin's Company", "Creature — Dwarf Warrior", hobA9DainOracle, 2, 2)
	return g, me, dwarf, equipment, bear, land, deep
}

func TestHobA9DainTakesADwarfAndBottomsTheRest(t *testing.T) {
	g, me, dwarf, equipment, bear, land, deep := hobA9DainLook(t)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the look")
	}
	if len(c.ChooseCards) != 2 {
		t.Errorf("%d cards offered, want the Dwarf and the Equipment only", len(c.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, dwarf)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(dwarf) {
		t.Error("the Dwarf is not in hand")
	}
	for _, id := range []uuid.UUID{equipment, bear, land} {
		if !me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Errorf("card %s should be in the library, not in hand", id)
		}
	}
	// The rest went UNDER the fifth card, in some order.
	if n := len(me.Library.Cards); n != 4 || me.Library.Cards[n-1].InstanceID != deep {
		t.Errorf("the library should be the untouched fifth card (the top) over the three bottomed ones: %d cards", n)
	}
}

func TestHobA9DainCanTakeTheEquipmentInstead(t *testing.T) {
	g, me, dwarf, equipment, _, _, _ := hobA9DainLook(t)
	answerChooseCards(t, g, me.ID, equipment)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(equipment) || me.Hand.Contains(dwarf) {
		t.Error("only the Equipment should have been taken")
	}
}

func TestHobA9DainMayTakeNothing(t *testing.T) {
	g, me, dwarf, equipment, bear, land, deep := hobA9DainLook(t)
	hand := me.Hand.Size()
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand)
	}
	if n := len(me.Library.Cards); n != 5 || me.Library.Cards[n-1].InstanceID != deep {
		t.Error("the four looked-at cards should all be on the bottom, under the untouched fifth card")
	}
	for _, id := range []uuid.UUID{dwarf, equipment, bear, land} {
		if !me.Library.Contains(id) {
			t.Errorf("card %s is not in the library", id)
		}
	}
}

// --- The two Sagas -----------------------------------------------------------

const (
	hobA9BurnOracle  = "a97d6c5c-1cff-442d-b535-fc8389160b0b"
	hobA9RoadsOracle = "3cd3cdc9-2ed2-4cdf-b58a-12c3a2ef218b"
)

func TestHobA9SagasAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA9BurnOracle:  "Burn, Burn, Tree and Fern",
		hobA9RoadsOracle: "Roads Go Ever, Ever On",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered %v, name %q, completeness %s", name, ok, spec.Name, spec.Completeness)
		}
	}
}

// hobA9NextChapter walks to the active seat's next precombat main, where the
// Saga's lore counter arrives, and settles the chapter that reaches.
func hobA9NextChapter(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
}

func TestHobA9BurnBurnWalksAllFourChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Big Beast", "Creature — Beast", 7, 7)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirArtifact := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	myArtifact := pushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)

	// I — six damage to a creature an opponent controls.
	saga := castCatalogSpell(t, g, "Burn, Burn, Tree and Fern", "Enchantment — Saga", hobA9BurnOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter I asked for no target")
	}
	if hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, big) {
		t.Errorf("chapter I targets %v: a creature an opponent controls only", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, big); got != 6 {
		t.Fatalf("chapter I marked %d damage, want 6", got)
	}

	// II — destroy an artifact an opponent controls.
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter II asked for no target")
	}
	if hasID(p.PickTargetCards, myArtifact) || !hasID(p.PickTargetCards, theirArtifact) {
		t.Errorf("chapter II targets %v: an artifact an opponent controls only", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirArtifact)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirArtifact) || !g.Battlefield.Contains(myArtifact) {
		t.Fatal("chapter II should destroy only the opponent's artifact")
	}

	// III and IV — add {R} each.
	for chapter := 3; chapter <= 4; chapter++ {
		me.ManaPool = game.ManaPool{}
		hobA9NextChapter(t, g)
		if got := poolColors(me); len(got) != 1 || got[0] != "R" {
			t.Fatalf("chapter %d left %v in the pool, want one R", chapter, got)
		}
	}
	// Sacrificed after IV.
	if !inGraveyardOf(g, me.ID, saga) {
		t.Error("the Saga should be sacrificed after chapter IV")
	}
}

// A chapter with nothing to target is dropped and the Saga carries on.
func TestHobA9BurnBurnWithNothingToTargetStillAdvances(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	saga := castCatalogSpell(t, g, "Burn, Burn, Tree and Fern", "Enchantment — Saga", hobA9BurnOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("chapter I asked for a target with no creature to hit")
	}
	hobA9NextChapter(t, g)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Errorf("lore counters = %d, want 2", got)
	}
}

// hobA9RoadsLibrary seeds the library: three basic Plains, a nonbasic
// Plains and a Mountain, with the basics in front.
func hobA9RoadsLibrary(me *game.Player) {
	seedSearchLibrary(me,
		searchTestLand("Plains One", "Basic Land — Plains"),
		searchTestLand("Plains Two", "Basic Land — Plains"),
		searchTestLand("Plains Three", "Basic Land — Plains"),
		searchTestLand("Hallowed Nonbasic", "Land — Plains Island"),
		searchTestLand("A Mountain", "Basic Land — Mountain"),
	)
}

func TestHobA9RoadsChapterOneExilesTwoBasicPlainsAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobA9RoadsLibrary(me)
	life := me.Life
	saga := castCatalogSpell(t, g, "Roads Go Ever, Ever On", "Enchantment — Saga", hobA9RoadsOracle, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Hallowed Nonbasic") != uuid.Nil || searchOptionNamed(g, c, "A Mountain") != uuid.Nil {
		t.Error("only basic Plains cards may be found")
	}
	if me.Life != life {
		t.Error("the life was gained before the search was answered")
	}
	answerSearchNamed(t, g, me.ID, "Plains One", "Plains Two")
	passPriorityAroundTable(t, g)

	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	ref := game.PermissionCardRef(hobA9SagaRef(t, g, saga))
	var linked int
	g.ReadSnapshot(func() { linked = len(g.ExiledWithForEffect(ref)) })
	if linked != 2 {
		t.Errorf("%d cards exiled with the Saga, want 2", linked)
	}
	if len(me.Library.Cards) != 3 {
		t.Errorf("library = %d cards, want 3", len(me.Library.Cards))
	}
}

// hobA9SagaRef is the object reference of a battlefield Saga.
func hobA9SagaRef(t *testing.T, g *game.Game, id uuid.UUID) game.ObjectRef {
	t.Helper()
	var ref game.ObjectRef
	var ok bool
	g.ReadSnapshot(func() { ref, ok = g.PermanentRefForEffect(id) })
	if !ok {
		t.Fatal("the Saga is not on the battlefield")
	}
	return ref
}

// A search may come up short: "up to two".
func TestHobA9RoadsMayTakeFewerThanTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobA9RoadsLibrary(me)
	life := me.Life
	castCatalogSpell(t, g, "Roads Go Ever, Ever On", "Enchantment — Saga", hobA9RoadsOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Plains One")
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d: the life comes whatever was found", me.Life, life+2)
	}
	if !g.Exile.Contains(searchedIDByName(me, g, "Plains One")) {
		t.Error("the Plains was not exiled")
	}
}

// searchedIDByName finds a card by name in exile.
func searchedIDByName(_ *game.Player, g *game.Game, name string) uuid.UUID {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestHobA9RoadsWalksAllFourChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobA9RoadsLibrary(me)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	pushCatalogPermanent(g, me.ID, "Plains A", "Basic Land — Plains", "", false)
	pushCatalogPermanent(g, me.ID, "Plains B", "Basic Land — Plains", "", false)
	pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)

	saga := castCatalogSpell(t, g, "Roads Go Ever, Ever On", "Enchantment — Saga", hobA9RoadsOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Plains One", "Plains Two")
	passPriorityAroundTable(t, g)
	one, two := searchedIDByName(me, g, "Plains One"), searchedIDByName(me, g, "Plains Two")

	// II — two cards to choose from; the controller picks one.
	advanceToPrecombatMainOf(t, g, 0)
	hand := me.Hand.Size() // after the turn's draw
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, two)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(two) || !g.Exile.Contains(one) || me.Hand.Size() != hand+1 {
		t.Fatal("chapter II should put exactly the chosen card into hand")
	}

	// III — one card left: it is taken.
	advanceToPrecombatMainOf(t, g, 0)
	hand = me.Hand.Size() // after the next draw
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(one) || me.Hand.Size() != hand+1 {
		t.Fatal("chapter III should put the last exiled card into hand")
	}

	// IV — whenever you attack this turn, +1/+1 for each Plains.
	hobA9NextChapter(t, g)
	if !inGraveyardOf(g, me.ID, saga) {
		t.Error("the Saga should be sacrificed after chapter IV")
	}
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, attacker)
		passPriorityAroundTable(t, g)
	}
	c := s58p6View(t, g, attacker)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the attacker is %d/%d, want 3/3 (two Plains)", c.CurrentPower(), c.CurrentToughness())
	}
}

// The pump counts Plains when it resolves, not when the chapter did, and
// ends with the turn.
func TestHobA9RoadsPumpEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobA9RoadsLibrary(me)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	pushCatalogPermanent(g, me.ID, "Plains A", "Basic Land — Plains", "", false)
	castCatalogSpell(t, g, "Roads Go Ever, Ever On", "Enchantment — Saga", hobA9RoadsOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID)
	passPriorityAroundTable(t, g)
	hobA9NextChapter(t, g)
	hobA9NextChapter(t, g)
	hobA9NextChapter(t, g)
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, attacker)
		passPriorityAroundTable(t, g)
	}
	if c := s58p6View(t, g, attacker); c.CurrentPower() != 2 {
		t.Fatalf("with one Plains the attacker is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	advanceToNextSeatsTurn(t, g)
	if c := s58p6View(t, g, attacker); c.CurrentPower() != 1 {
		t.Errorf("the pump outlived the turn: power %d", c.CurrentPower())
	}
}
