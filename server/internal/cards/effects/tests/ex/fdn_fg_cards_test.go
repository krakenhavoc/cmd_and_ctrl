package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		effectstest.AssertKeywords(t, g, id, c.keyword)
	}
}

// fdnFGSick seeds a catalog creature that arrived this turn.
func fdnFGSick(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner, SummonedThisTurn: true,
	})
}

// --- Fanatical Firebrand --------------------------------------------------------

func TestFdnFGFanaticalFirebrandSacrificesForOneDamageTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	goblin := fdnFGSick(g, me.ID, "Fanatical Firebrand", "Creature — Goblin Pirate", fdnFGFanaticalFirebrandOracle, 1, 1)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, goblin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	if g.Battlefield.Contains(goblin) || !me.Graveyard.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
}

// --- Felidar Cub ----------------------------------------------------------------

func TestFdnFGFelidarCubSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cub := effectstest.B12Push(g, me.ID, "Felidar Cub", "Creature — Cat Beast", fdnFGFelidarCubOracle, 2, 2)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	if err := g.ActivateCatalogAbility(me.ID, cub, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(relic)}); err == nil {
		t.Fatal("an artifact was accepted as the target")
	}
	effectstest.B16Activate(t, g, me.ID, cub, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(aura)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !opp.Graveyard.Contains(aura) {
		t.Error("the enchantment was not destroyed")
	}
	if g.Battlefield.Contains(cub) {
		t.Error("the Cub was not sacrificed")
	}
}

// --- Felidar Savior -------------------------------------------------------------

func TestFdnFGFelidarSaviorCountersUpToTwoOtherCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	savior := effectstest.JmpB4CastCreature(t, g, me, "Felidar Savior", "Creature — Cat Beast", fdnFGFelidarSaviorOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.JmpB4Pick(t, g, me.ID, theirs); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	if err := effectstest.JmpB4Pick(t, g, me.ID, savior); err == nil {
		t.Fatal("the Savior itself was accepted (\"other\")")
	}
	if err := effectstest.JmpB4Pick(t, g, me.ID, bear, elf); err != nil {
		t.Fatalf("two other creatures you control: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.JmpB4PlusOneCounters(t, g, bear) != 1 || effectstest.JmpB4PlusOneCounters(t, g, elf) != 1 {
		t.Error("each target gets a +1/+1 counter")
	}
	if effectstest.JmpB4PlusOneCounters(t, g, savior) != 0 || effectstest.JmpB4PlusOneCounters(t, g, theirs) != 0 {
		t.Error("only the targets get a counter")
	}
}

// --- Felling Blow ---------------------------------------------------------------

func fdnFGFellingBlow(t *testing.T, g *game.Game, mine, theirs uuid.UUID) error {
	t.Helper()
	_, err := effectstest.SocB16TryCast(t, g, game.Card{Name: "Felling Blow", TypeLine: "Sorcery", OracleID: fdnFGFellingBlowOracle},
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
		}})
	return err
}

func TestFdnFGFellingBlowCountersThenDealsTheNewPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	if err := fdnFGFellingBlow(t, g, mine, other); err == nil {
		t.Fatal("your own creature was accepted as the creature an opponent controls")
	}
	if err := fdnFGFellingBlow(t, g, mine, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the counter makes it 3 power, so 3 damage kills the 3/3")
	}
}

// --- Flashfreeze ----------------------------------------------------------------

func TestFdnFGFlashfreezeCountersRedOrGreenOnly(t *testing.T) {
	for _, colors := range [][]string{{"R"}, {"G", "W"}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.ToMainForCost(t, g)
		spell := effectstest.HandCardFull(me, "Coloured Thing", "Sorcery", "", "", colors)
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("%v: cast: %v", colors, err)
		}
		effectstest.CastCatalogSpell(t, g, "Flashfreeze", "Instant", fdnFGFlashfreezeOracle, effectstest.CardRefs(spell))
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(spell) {
			t.Errorf("%v: the spell was not countered", colors)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	white := effectstest.HandCardFull(me, "White Thing", "Sorcery", "", "", []string{"W"})
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast white spell: %v", err)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Flashfreeze", "Instant", fdnFGFlashfreezeOracle, effectstest.CardRefs(white)); err == nil {
		t.Error("a white spell was accepted as the target")
	}
}

// --- Fleeting Distraction -------------------------------------------------------

func TestFdnFGFleetingDistractionShrinksThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Fleeting Distraction", "Instant", fdnFGFleetingDistraction, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 1/2", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (the spell left, a card was drawn)", got, hand+1)
	}
}

// --- Fynn, the Fangbearer -------------------------------------------------------

func TestFdnFGFynnPoisonsForEachDeathtouchCreatureThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fynn := effectstest.B12Push(g, me.ID, "Fynn, the Fangbearer", "Legendary Creature — Human Warrior", fdnFGFynnOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DealNoncombatDamage(g, me.ID, fynn, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamage(g, me.ID, bear, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 0 {
		t.Fatalf("poison %d after noncombat damage and a creature without deathtouch, want 0", got)
	}
	effectstest.DealCombatDamage(g, me.ID, fynn, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("poison %d after Fynn connected, want 2", got)
	}
	if got := me.Counters[game.CounterPoison]; got != 0 {
		t.Errorf("the controller got %d poison counters", got)
	}
}

// --- Gnarlback Rhino ------------------------------------------------------------

func TestFdnFGGnarlbackRhinoDrawsWhenYourSpellTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rhino := effectstest.B12Push(g, me.ID, "Gnarlback Rhino", "Creature — Rhino", fdnFGGnarlbackRhinoOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", fdnFGGiantGrowthOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d after a spell targeting another creature, want %d", got, hand)
	}
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", fdnFGGiantGrowthOracle, effectstest.CardRefs(rhino))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after a spell targeting the Rhino, want %d", got, hand+1)
	}
}

// --- Goblin Boarders ------------------------------------------------------------

func TestFdnFGGoblinBoardersEntersWithACounterOnlyAfterAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.AvB16CastCreature(t, g, "Goblin Boarders", "Creature — Goblin Pirate", fdnFGGoblinBoardersOracle, 3, 2)
	if n := effectstest.CountersOn(g, first, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters without an attack this turn, want 0", n)
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	second := effectstest.AvB16CastCreature(t, g, "Goblin Boarders", "Creature — Goblin Pirate", fdnFGGoblinBoardersOracle, 3, 2)
	if n := effectstest.CountersOn(g, second, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after attacking this turn, want 1", n)
	}
}

// --- Goblin Firebomb ------------------------------------------------------------

func TestFdnFGGoblinFirebombSacrificesToDestroyAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if spec, _ := effects.Lookup(fdnFGGoblinFirebombOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want [flash]", spec.PrintedKeywords)
	}
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Firebomb", "Artifact", fdnFGGoblinFirebombOracle, false)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.P1g3Activate(t, g, bomb, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)})
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("the permanent was not destroyed")
	}
	if g.Battlefield.Contains(bomb) || !me.Graveyard.Contains(bomb) {
		t.Error("the Firebomb was not sacrificed")
	}
}

// --- Goblin Oriflamme -----------------------------------------------------------

func TestFdnFGGoblinOriflammePumpsYourAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Goblin Oriflamme", "Enchantment", fdnFGGoblinOriflammeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("a creature not attacking has power %d, want 2", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("an attacking Bear is %d/%d, want 3/2", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, home); p != 2 {
		t.Errorf("a creature that stayed home has power %d, want 2", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

// --- Goblin Smuggler ------------------------------------------------------------

func TestFdnFGGoblinSmugglerMakesAnotherSmallCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	smuggler := fdnFGSick(g, me.ID, "Goblin Smuggler", "Creature — Goblin Rogue", fdnFGGoblinSmugglerOracle, 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, bad := range []uuid.UUID{smuggler, big} {
		if err := g.ActivateCatalogAbility(me.ID, smuggler, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bad)}); err == nil {
			t.Fatalf("activated targeting %s", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, smuggler, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("activate the turn it arrived, at another player's creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, theirs)&game.CantBeBlocked == 0 {
		t.Error("the target can be blocked")
	}
}

// --- Gorehorn Raider ------------------------------------------------------------

func TestFdnFGGorehornRaiderDealsTwoOnlyAfterAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AvB16CastCreature(t, g, "Gorehorn Raider", "Creature — Minotaur Pirate", fdnFGGorehornRaiderOracle, 4, 4)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the raid trigger fired without an attack this turn")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.AvB16CastCreature(t, g, "Gorehorn Raider", "Creature — Minotaur Pirate", fdnFGGorehornRaiderOracle, 4, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 2 damage")
	}
}

// --- Grappling Kraken -----------------------------------------------------------

func TestFdnFGGrapplingKrakenTapsAndStunsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Grappling Kraken", "Creature — Kraken", fdnFGGrapplingKrakenOracle, 5, 6)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("your own creature was accepted")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, victim) {
		t.Error("the target was not tapped")
	}
	if got := effectstest.CountersOn(g, victim, game.CounterStun); got != 1 {
		t.Errorf("stun counters %d, want 1", got)
	}
}
