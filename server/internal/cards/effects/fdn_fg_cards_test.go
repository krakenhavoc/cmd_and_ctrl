package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_fg_cards_test.go — Foundations, F–G (#4465): the cards from that
// stretch of the alphabet that earlier batches left and that existing
// seams express. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	fdnFGFanaticalFirebrandOracle = "d36e11f1-6ab3-4273-8114-a8fbbe21c1c3"
	fdnFGFelidarCubOracle         = "cae60cf8-cd64-4595-a4dd-946694cf2bb1"
	fdnFGFelidarSaviorOracle      = "2bee2508-cd5d-4c62-a205-12d8bbb3143d"
	fdnFGFellingBlowOracle        = "fbfa19ca-98bb-4c7d-a04e-920a8aafb031"
	fdnFGFlashfreezeOracle        = "eaf98e03-729b-4145-b2af-c910c415c15d"
	fdnFGFleetingDistraction      = "b36271cb-cda2-434a-a580-6050dc460409"
	fdnFGFynnOracle               = "c0b1fba1-4338-4671-934b-098689ad2085"
	fdnFGGnarlbackRhinoOracle     = "1ae339fd-8e64-4e3a-8e27-e4993932ba62"
	fdnFGGoblinBoardersOracle     = "4c64feae-95d7-426d-9e42-d3b0841d2054"
	fdnFGGoblinFirebombOracle     = "2625fa9c-638c-46d3-966b-1166f440186f"
	fdnFGGoblinOriflammeOracle    = "836bd011-2da5-443a-a814-19a664b98a1a"
	fdnFGGoblinSmugglerOracle     = "eae15f7d-1f6e-4dfc-afea-7698b93cd145"
	fdnFGGorehornRaiderOracle     = "41699d92-c209-4fcb-bbd6-685b3437d0c2"
	fdnFGGrapplingKrakenOracle    = "1ba99ade-df8d-4654-882a-a5f4256ada3d"

	// Giant Growth is the catalog spell the Rhino tests target it with.
	fdnFGGiantGrowthOracle = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
)

func TestFdnFGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnFGFanaticalFirebrandOracle: "Fanatical Firebrand",
		fdnFGFelidarCubOracle:         "Felidar Cub",
		fdnFGFelidarSaviorOracle:      "Felidar Savior",
		fdnFGFellingBlowOracle:        "Felling Blow",
		fdnFGFlashfreezeOracle:        "Flashfreeze",
		fdnFGFleetingDistraction:      "Fleeting Distraction",
		fdnFGFynnOracle:               "Fynn, the Fangbearer",
		fdnFGGnarlbackRhinoOracle:     "Gnarlback Rhino",
		fdnFGGoblinBoardersOracle:     "Goblin Boarders",
		fdnFGGoblinFirebombOracle:     "Goblin Firebomb",
		fdnFGGoblinOriflammeOracle:    "Goblin Oriflamme",
		fdnFGGoblinSmugglerOracle:     "Goblin Smuggler",
		fdnFGGorehornRaiderOracle:     "Gorehorn Raider",
		fdnFGGrapplingKrakenOracle:    "Grappling Kraken",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestFdnFGKeywordBodiesCarryTheirPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keyword                string
	}{
		{"Fanatical Firebrand", "Creature — Goblin Pirate", fdnFGFanaticalFirebrandOracle, 1, 1, "haste"},
		{"Felidar Savior", "Creature — Cat Beast", fdnFGFelidarSaviorOracle, 2, 3, "lifelink"},
		{"Fynn, the Fangbearer", "Legendary Creature — Human Warrior", fdnFGFynnOracle, 1, 3, "deathtouch"},
		{"Gnarlback Rhino", "Creature — Rhino", fdnFGGnarlbackRhinoOracle, 4, 4, "trample"},
		{"Goblin Smuggler", "Creature — Goblin Rogue", fdnFGGoblinSmugglerOracle, 2, 2, "haste"},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keyword)
	}
}

// fdnFGSick seeds a catalog creature that arrived this turn.
func fdnFGSick(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner, SummonedThisTurn: true,
	})
}

// --- Fanatical Firebrand --------------------------------------------------------

func TestFdnFGFanaticalFirebrandSacrificesForOneDamageTheTurnItArrives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	goblin := fdnFGSick(g, me.ID, "Fanatical Firebrand", "Creature — Goblin Pirate", fdnFGFanaticalFirebrandOracle, 1, 1)
	life := opp.Life
	b16Activate(t, g, me.ID, goblin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	if g.Battlefield.Contains(goblin) || !me.Graveyard.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
}

// --- Felidar Cub ----------------------------------------------------------------

func TestFdnFGFelidarCubSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cub := b12Push(g, me.ID, "Felidar Cub", "Creature — Cat Beast", fdnFGFelidarCubOracle, 2, 2)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	aura := b12Permanent(g, opp.ID, "Aura", "Enchantment")
	if err := g.ActivateCatalogAbility(me.ID, cub, 0, game.ActivateAbilityParams{Targets: b16TargetCard(relic)}); err == nil {
		t.Fatal("an artifact was accepted as the target")
	}
	b16Activate(t, g, me.ID, cub, 0, game.ActivateAbilityParams{Targets: b16TargetCard(aura)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !opp.Graveyard.Contains(aura) {
		t.Error("the enchantment was not destroyed")
	}
	if g.Battlefield.Contains(cub) {
		t.Error("the Cub was not sacrificed")
	}
}

// --- Felidar Savior -------------------------------------------------------------

func TestFdnFGFelidarSaviorCountersUpToTwoOtherCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	savior := jmpB4CastCreature(t, g, me, "Felidar Savior", "Creature — Cat Beast", fdnFGFelidarSaviorOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if err := jmpB4Pick(t, g, me.ID, theirs); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	if err := jmpB4Pick(t, g, me.ID, savior); err == nil {
		t.Fatal("the Savior itself was accepted (\"other\")")
	}
	if err := jmpB4Pick(t, g, me.ID, bear, elf); err != nil {
		t.Fatalf("two other creatures you control: %v", err)
	}
	passPriorityAroundTable(t, g)
	if jmpB4PlusOneCounters(t, g, bear) != 1 || jmpB4PlusOneCounters(t, g, elf) != 1 {
		t.Error("each target gets a +1/+1 counter")
	}
	if jmpB4PlusOneCounters(t, g, savior) != 0 || jmpB4PlusOneCounters(t, g, theirs) != 0 {
		t.Error("only the targets get a counter")
	}
}

// --- Felling Blow ---------------------------------------------------------------

func fdnFGFellingBlow(t *testing.T, g *game.Game, mine, theirs uuid.UUID) error {
	t.Helper()
	_, err := socB16TryCast(t, g, game.Card{Name: "Felling Blow", TypeLine: "Sorcery", OracleID: fdnFGFellingBlowOracle},
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
		}})
	return err
}

func TestFdnFGFellingBlowCountersThenDealsTheNewPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	if err := fdnFGFellingBlow(t, g, mine, other); err == nil {
		t.Fatal("your own creature was accepted as the creature an opponent controls")
	}
	if err := fdnFGFellingBlow(t, g, mine, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the counter makes it 3 power, so 3 damage kills the 3/3")
	}
}

// --- Flashfreeze ----------------------------------------------------------------

func TestFdnFGFlashfreezeCountersRedOrGreenOnly(t *testing.T) {
	for _, colors := range [][]string{{"R"}, {"G", "W"}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		toMainForCost(t, g)
		spell := handCardFull(me, "Coloured Thing", "Sorcery", "", "", colors)
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("%v: cast: %v", colors, err)
		}
		castCatalogSpell(t, g, "Flashfreeze", "Instant", fdnFGFlashfreezeOracle, cardRefs(spell))
		passPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(spell) {
			t.Errorf("%v: the spell was not countered", colors)
		}
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	white := handCardFull(me, "White Thing", "Sorcery", "", "", []string{"W"})
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast white spell: %v", err)
	}
	if err := castCatalogSpellErr(t, g, "Flashfreeze", "Instant", fdnFGFlashfreezeOracle, cardRefs(white)); err == nil {
		t.Error("a white spell was accepted as the target")
	}
}

// --- Fleeting Distraction -------------------------------------------------------

func TestFdnFGFleetingDistractionShrinksThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Fleeting Distraction", "Instant", fdnFGFleetingDistraction, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 1 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 1/2", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (the spell left, a card was drawn)", got, hand+1)
	}
}

// --- Fynn, the Fangbearer -------------------------------------------------------

func TestFdnFGFynnPoisonsForEachDeathtouchCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fynn := b12Push(g, me.ID, "Fynn, the Fangbearer", "Legendary Creature — Human Warrior", fdnFGFynnOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dealNoncombatDamage(g, me.ID, fynn, opp.ID)
	passPriorityAroundTable(t, g)
	dealCombatDamage(g, me.ID, bear, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 0 {
		t.Fatalf("poison %d after noncombat damage and a creature without deathtouch, want 0", got)
	}
	dealCombatDamage(g, me.ID, fynn, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("poison %d after Fynn connected, want 2", got)
	}
	if got := me.Counters[game.CounterPoison]; got != 0 {
		t.Errorf("the controller got %d poison counters", got)
	}
}

// --- Gnarlback Rhino ------------------------------------------------------------

func TestFdnFGGnarlbackRhinoDrawsWhenYourSpellTargetsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rhino := b12Push(g, me.ID, "Gnarlback Rhino", "Creature — Rhino", fdnFGGnarlbackRhinoOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Giant Growth", "Instant", fdnFGGiantGrowthOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d after a spell targeting another creature, want %d", got, hand)
	}
	castCatalogSpell(t, g, "Giant Growth", "Instant", fdnFGGiantGrowthOracle, cardRefs(rhino))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after a spell targeting the Rhino, want %d", got, hand+1)
	}
}

// --- Goblin Boarders ------------------------------------------------------------

func TestFdnFGGoblinBoardersEntersWithACounterOnlyAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := avB16CastCreature(t, g, "Goblin Boarders", "Creature — Goblin Pirate", fdnFGGoblinBoardersOracle, 3, 2)
	if n := countersOn(g, first, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters without an attack this turn, want 0", n)
	}
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	second := avB16CastCreature(t, g, "Goblin Boarders", "Creature — Goblin Pirate", fdnFGGoblinBoardersOracle, 3, 2)
	if n := countersOn(g, second, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after attacking this turn, want 1", n)
	}
}

// --- Goblin Firebomb ------------------------------------------------------------

func TestFdnFGGoblinFirebombSacrificesToDestroyAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if spec, _ := Lookup(fdnFGGoblinFirebombOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want [flash]", spec.PrintedKeywords)
	}
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Firebomb", "Artifact", fdnFGGoblinFirebombOracle, false)
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	p1g3Activate(t, g, bomb, 0, game.ActivateAbilityParams{Targets: cardRefs(land)})
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("the permanent was not destroyed")
	}
	if g.Battlefield.Contains(bomb) || !me.Graveyard.Contains(bomb) {
		t.Error("the Firebomb was not sacrificed")
	}
}

// --- Goblin Oriflamme -----------------------------------------------------------

func TestFdnFGGoblinOriflammePumpsYourAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Goblin Oriflamme", "Enchantment", fdnFGGoblinOriflammeOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("a creature not attacking has power %d, want 2", p)
	}
	declareAttack(t, g, opp.ID, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("an attacking Bear is %d/%d, want 3/2", p, tough)
	}
	if p, _ := effectivePTOf(t, g, home); p != 2 {
		t.Errorf("a creature that stayed home has power %d, want 2", p)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

// --- Goblin Smuggler ------------------------------------------------------------

func TestFdnFGGoblinSmugglerMakesAnotherSmallCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	smuggler := fdnFGSick(g, me.ID, "Goblin Smuggler", "Creature — Goblin Rogue", fdnFGGoblinSmugglerOracle, 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, bad := range []uuid.UUID{smuggler, big} {
		if err := g.ActivateCatalogAbility(me.ID, smuggler, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bad)}); err == nil {
			t.Fatalf("activated targeting %s", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, smuggler, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(theirs)}); err != nil {
		t.Fatalf("activate the turn it arrived, at another player's creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, theirs)&game.CantBeBlocked == 0 {
		t.Error("the target can be blocked")
	}
}

// --- Gorehorn Raider ------------------------------------------------------------

func TestFdnFGGorehornRaiderDealsTwoOnlyAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	avB16CastCreature(t, g, "Gorehorn Raider", "Creature — Minotaur Pirate", fdnFGGorehornRaiderOracle, 4, 4)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the raid trigger fired without an attack this turn")
	}
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	avB16CastCreature(t, g, "Gorehorn Raider", "Creature — Minotaur Pirate", fdnFGGorehornRaiderOracle, 4, 4)
	b04WaitForPick(t, g, me.ID)
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 2 damage")
	}
}

// --- Grappling Kraken -----------------------------------------------------------

func TestFdnFGGrapplingKrakenTapsAndStunsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Grappling Kraken", "Creature — Kraken", fdnFGGrapplingKrakenOracle, 5, 6)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	playLandFromHand(t, g, "Forest", "")
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("your own creature was accepted")
	}
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !isTapped(g, victim) {
		t.Error("the target was not tapped")
	}
	if got := countersOn(g, victim, game.CounterStun); got != 1 {
		t.Errorf("stun counters %d, want 1", got)
	}
}
