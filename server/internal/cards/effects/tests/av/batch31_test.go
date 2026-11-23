package testsav

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b31CountNamed counts battlefield permanents `controller` controls
// with the given name.
func b31CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// b31GraveyardCards seeds n filler cards into a player's graveyard.
func b31GraveyardCards(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Goblin King
// was already on main from the S26 tribal lords and is pinned here
// so the table matches the issue's 24 minus the six declared skips.
func TestBatch31CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B31GoblinInstigatorOracle:     "Goblin Instigator",
		effectstest.B31BlossomingTortoiseOracle:   "Blossoming Tortoise",
		effectstest.B31SealOfCleansingOracle:      "Seal of Cleansing",
		effectstest.B31LeoninWarleaderOracle:      "Leonin Warleader",
		effectstest.B31KinjallisSunwingOracle:     "Kinjalli's Sunwing",
		effectstest.B31SunscorchedDesertOracle:    "Sunscorched Desert",
		effectstest.B31TheWarringTriadOracle:      "The Warring Triad",
		effectstest.B31FainTheBrokerOracle:        "Fain, the Broker",
		effectstest.B31SoulsMajestyOracle:         "Soul's Majesty",
		effectstest.B31TillerEngineOracle:         "Tiller Engine",
		effectstest.B31RapaciousGuestOracle:       "Rapacious Guest",
		effectstest.B31NephaliaDrownyardOracle:    "Nephalia Drownyard",
		effectstest.B31TheEndstoneOracle:          "The Endstone",
		effectstest.B31VampireOfTheDireMoonOracle: "Vampire of the Dire Moon",
		effectstest.B31GemstoneMineOracle:         "Gemstone Mine",
		effectstest.B31FlagstonesOfTrokairOracle:  "Flagstones of Trokair",
		effectstest.B31WeaponsManufacturingOracle: "Weapons Manufacturing",
		effectstest.B31MagusOfTheMoonOracle:       "Magus of the Moon",
		effectstest.B31GoblinKingAlreadyOracle:    "Goblin King",
	}
	if len(want) != 19 {
		t.Fatalf("the batch registers 18 cards plus one already on main, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
	// The last declared skip, Tree of Redemption, came off this list with the
	// continuous effect that has no duration (a base-toughness effect pinned to
	// the object, CR 611.2a) and ships in tree_of_redemption.go (#5368).
	// Oviya,
	// Automech Artisan came OFF this list with the put-from-hand prompt
	// (PutFromHandOntoBattlefield) and ships in oviya_automech_artisan.go
	// (#4464). Manaweft Sliver came OFF this list with ADR 0093's granted
	// abilities and ships in manaweft_sliver.go; Duress came off it with
	// ADR 0116's filtered revealed-hand pick and ships in duress.go;
	// Conspicuous Snoop came off it with ADR 0155's library-top borrow
	// and ships in conspicuous_snoop.go.
}

// --- the keywords and statics --------------------------------------

func TestB31VampireOfTheDireMoonHasDeathtouchAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := effectstest.B31Push(g, me.ID, "Vampire of the Dire Moon", "Creature — Vampire", effectstest.B31VampireOfTheDireMoonOracle, "{B}", 1, 1, "B")
	if !effectstest.HasEffectiveKeyword(t, g, vamp, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, vamp, "lifelink") {
		t.Fatal("deathtouch and lifelink are printed")
	}
	before := me.Life
	effectstest.AttackWith(t, g, opp.ID, vamp)
	if me.Life != before+1 {
		t.Errorf("lifelink: life %d, want %d", me.Life, before+1)
	}
}

func TestB31TheWarringTriadWakesAtEightCardsInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	b31GraveyardCards(me, 7)
	triad := effectstest.B31Push(g, me.ID, "The Warring Triad", "Legendary Artifact Creature — God", effectstest.B31TheWarringTriadOracle, "{3}", 5, 5)
	if effectstest.ContainsString(effectstest.EffectiveTypes(t, g, triad), "Creature") {
		t.Fatal("seven cards: not a creature")
	}
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, triad), "Artifact") {
		t.Error("it is still an artifact")
	}
	if effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, triad), "God") {
		t.Error("a creature type goes with the creature type (CR 205.1a)")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(triad, g.Seats[1].ID); err == nil {
		t.Error("a noncreature cannot attack")
	}
	b31GraveyardCards(me, 1)
	g.BumpLayerVersionForTest()
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, triad), "Creature") {
		t.Fatal("eight cards: a creature")
	}
	if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, triad), "God") {
		t.Error("the God type is back")
	}
	for _, kw := range []string{"flying", "trample", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, triad, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	if got := effectstest.EffectivePower(t, g, triad); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	// #3845: the tap-and-mill ability is real now, so nothing is
	// declared as missing. The ability's own tests are in
	// mill_cost_test.go.
	if spec, _ := effects.Lookup(effectstest.B31TheWarringTriadOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the mill cost is implemented; the Triad is complete")
	}
}

func TestB31BlossomingTortoisePumpsLandCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B31Push(g, me.ID, "Blossoming Tortoise", "Creature — Turtle", effectstest.B31BlossomingTortoiseOracle, "{2}{G}{G}", 3, 3, "G")
	arbor := effectstest.B31Push(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", "", "", 1, 1, "G")
	bear := effectstest.B31Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	theirs := effectstest.B31Push(g, opp.ID, "Their Arbor", "Land Creature — Forest Dryad", "", "", 1, 1, "G")
	if got := effectstest.EffectivePower(t, g, arbor); got != 2 {
		t.Errorf("a land creature you control: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("a nonland creature: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's land creature: power %d, want 1", got)
	}
}

// --- the replacements ----------------------------------------------

func TestB31KinjallisSunwingTapsOpponentsCreaturesOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wing := effectstest.B31Push(g, me.ID, "Kinjalli's Sunwing", "Creature — Dinosaur", effectstest.B31KinjallisSunwingOracle, "{2}{W}", 2, 3, "W")
	if !effectstest.HasEffectiveKeyword(t, g, wing, "flying") {
		t.Error("flying is printed")
	}
	mine := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	if effectstest.B31Tapped(t, g, mine) {
		t.Error("your own creature enters untapped")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	theirs := effectstest.B20CastCreature(t, g, opp, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
	if n := effectstest.TapEventsFor(g, theirs); n != 0 {
		t.Errorf("a replacement, not a tap: %d tap events", n)
	}
	rock := effectstest.B20CastCreature(t, g, opp, "Their Rock", "Artifact", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B31Tapped(t, g, rock) {
		t.Error("only creatures enter tapped")
	}
}

// --- the triggers --------------------------------------------------

func TestB31GoblinInstigatorBringsAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Goblin Instigator", "Creature — Goblin Rogue", effectstest.B31GoblinInstigatorOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Goblin"); got != 1 {
		t.Fatalf("%d Goblin tokens, want 1", got)
	}
	tok := effectstest.CardByID(g, effectstest.FindBattlefieldByName(g, "Goblin"))
	if tok.Power != 1 || tok.Toughness != 1 || len(tok.Colors) != 1 || tok.Colors[0] != "R" {
		t.Errorf("a 1/1 red Goblin, got %d/%d %v", tok.Power, tok.Toughness, tok.Colors)
	}
}

func TestB31LeoninWarleaderAttacksWithTwoLifelinkCats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warleader := effectstest.B31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", effectstest.B31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
	before := opp.Life
	mine := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, warleader)
	if n := effectstest.TriggersOnStackFrom(g, warleader); n != 1 {
		t.Fatalf("one attack trigger, got %d", n)
	}
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	cats := effectstest.BattlefieldIDsNamed(g, "Cat")
	if len(cats) != 2 {
		t.Fatalf("%d Cats, want 2", len(cats))
	}
	for _, id := range cats {
		c := effectstest.CardByID(g, id)
		if !c.Tapped || c.AttackingTarget != opp.ID {
			t.Errorf("a Cat tapped and attacking the defender: %+v", c)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") || effectstest.EffectivePower(t, g, id) != 1 {
			t.Error("a 1/1 with lifelink")
		}
	}
	// The Cats were put onto the battlefield attacking, never
	// declared: no second Warleader trigger.
	if n := effectstest.TriggersOnStackFrom(g, warleader); n != 0 || len(g.PendingTriggers) != 0 {
		t.Error("an attacking token is not a declared attacker")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("the Warleader and two Cats connect: life %d, want %d", opp.Life, before-6)
	}
	if me.Life != mine+2 {
		t.Errorf("two lifelink Cats: life %d, want %d", me.Life, mine+2)
	}
}

func TestB31SunscorchedDesertPingsOnEntryAndTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	desert := effectstest.B31PlayLand(t, g, "Sunscorched Desert", "Land — Desert", effectstest.B31SunscorchedDesertOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetPlayers, opp.ID) || !effectstest.HasID(p.PickTargetPlayers, me.ID) {
		t.Error("any player is a legal target")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("1 damage: life %d, want %d", opp.Life, before-1)
	}
	if err := g.ActivateManaAbility(me.ID, desert, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if c, _ := effectstest.BattlefieldCard(g, desert); !c.HasSubtype("Desert") {
		t.Error("it is a Desert")
	}
}

func TestB31TheEndstoneDrawsOnLandsAndSpellsAndResetsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B31Push(g, me.ID, "The Endstone", "Legendary Artifact", effectstest.B31TheEndstoneOracle, "{7}", 0, 0)
	// The helpers seed the played card into the hand first, so a
	// draw shows as +1 on the size.
	hand := me.Hand.Size()
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a land play: hand %d → %d, want +1 (the draw)", hand, got)
	}
	hand = me.Hand.Size()
	effectstest.B24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a spell: hand %d → %d, want +1 (the draw)", hand, got)
	}
	// An opponent's spell is not "you cast".
	hand = me.Hand.Size()
	opp := g.Seats[1]
	effectstest.B29CastAs(t, g, opp, "Opt", "Instant", "", "{U}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an opponent's spell: hand %d → %d, want unchanged", hand, got)
	}
	// The end step: life becomes 20, whichever side of it you are on.
	me.Life = 33
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 20 {
		t.Errorf("above 20: life %d, want 20", me.Life)
	}
	// Another player's end step is not yours.
	me.Life = 33
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 33 {
		t.Errorf("an opponent's end step: life %d, want 33", me.Life)
	}
	me.Life = 4
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 20 {
		t.Errorf("below 20: life %d, want 20", me.Life)
	}
}

// #1337: b20LandPlayed still reads every hand-origin entry as a play,
// which stopped being exact once #654 shipped
// PutFromHandOntoBattlefield. A land Eureka Moment PUTS from hand
// (never played) still draws a card here — the declared, stronger-
// than-printed gap tracked as #1326. This test pins the gap so it
// gets noticed and cleared the day #1326 lands, not silently
// widened.
// #1326: Event.Played tells a PUT apart from a PLAY (CR 305.4), so a
// land Eureka Moment puts onto the battlefield draws nothing.
func TestB31TheEndstoneDoesNotDrawOnALandPutFromHandNotPlayed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B31Push(g, me.ID, "The Endstone", "Legendary Artifact", effectstest.B31TheEndstoneOracle, "{7}", 0, 0)
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")

	effectstest.CastCatalogSpell(t, g, "Eureka Moment", "Instant", effectstest.PutFromHandEurekaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	beforePut := me.Hand.Size()
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the chosen land did not reach the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != beforePut-1 {
		t.Errorf("hand %d → %d, want -1 (the land left, no draw — it was PUT, not played)", beforePut, got)
	}
}

func TestB31TillerEngineUntapsOrTapsWhenALandEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.B31Push(g, me.ID, "Tiller Engine", "Artifact Creature — Construct", effectstest.B31TillerEngineOracle, "{2}", 1, 3)
	// No opponent nonland permanent: only the untap is offered.
	mire := effectstest.PlayLandFromHand(t, g, "Haunted Mire", effectstest.B11HauntedMireOracle)
	if !effectstest.B31Tapped(t, g, mire) {
		t.Fatal("Haunted Mire enters tapped")
	}
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		if slices.Contains(c.ModeOptionIndex, 1) {
			t.Fatalf("nothing to tap: the tap mode is not offered: %v", c.ModeOptionIndex)
		}
		effectstest.AnswerMode(t, g, c, me.ID, 0)
	}
	if effectstest.TriggerOnStack(g, engine) == nil {
		t.Fatal("the choice is a trigger with a response window")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B31Tapped(t, g, mire) {
		t.Error("no nonland permanent to tap: the land is untapped")
	}
	// An opponent's nonland permanent: choose the tap mode and pick it.
	theirs := effectstest.B31Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	theirLand := effectstest.B31Push(g, opp.ID, "Their Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.B12ToMyNextUpkeep(t, g)
	mire2 := effectstest.PlayLandFromHand(t, g, "Haunted Mire", effectstest.B11HauntedMireOracle)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1}) {
		t.Fatalf("both modes are offered: %+v", c)
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, theirs) || effectstest.HasID(p.PickTargetCards, theirLand) || effectstest.HasID(p.PickTargetCards, engine) {
		t.Error("only nonland permanents an opponent controls are offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, theirs) {
		t.Error("the chosen permanent is tapped")
	}
	if !effectstest.B31Tapped(t, g, mire2) {
		t.Error("the tap mode was chosen: the land stays tapped")
	}
	// The untap mode, with a target available: asks for none.
	effectstest.B12ToMyNextUpkeep(t, g)
	mire3 := effectstest.PlayLandFromHand(t, g, "Haunted Mire", effectstest.B11HauntedMireOracle)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 0)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the untap mode targets nothing")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B31Tapped(t, g, mire3) {
		t.Error("untap mode: the land is untapped")
	}
	// A land that enters untapped is not a trigger.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	if effectstest.TriggerOnStack(g, engine) != nil || len(g.PendingTriggers) != 0 || effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.ModePickChoiceFor(g, me.ID) != nil {
		t.Error("an untapped entry: no trigger")
	}
}

func TestB31TillerEngineIgnoresOpponentsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.B31Push(g, me.ID, "Tiller Engine", "Artifact Creature — Construct", effectstest.B31TillerEngineOracle, "{2}", 1, 3)
	mire := effectstest.B13PlayAs(t, g, 1, "Haunted Mire", "Land", effectstest.B11HauntedMireOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOnStack(g, engine) != nil || !effectstest.B31Tapped(t, g, mire) {
		t.Error("an opponent's land is not \"a land you control\"")
	}
	_ = opp
}

func TestB31WeaponsManufacturingArmsEveryRealArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B31Push(g, me.ID, "Weapons Manufacturing", "Enchantment", effectstest.B31WeaponsManufacturingOracle, "{1}{R}", 0, 0, "R")
	effectstest.B24CastNoncreature(t, g, "Sol Ring", "Artifact", "{1}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Munitions"); got != 1 {
		t.Fatalf("an artifact spell: %d Munitions, want 1", got)
	}
	munitions := effectstest.FindBattlefieldByName(g, "Munitions")
	m := effectstest.CardByID(g, munitions)
	if !m.IsArtifact() || m.IsCreature() || len(m.Colors) != 0 || !effects.IsToken(m) {
		t.Errorf("a colorless artifact token, got %s %v", m.TypeLine, m.Colors)
	}
	// A token artifact does not chain, and an opponent's artifact is
	// not yours.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Munitions"); got != 1 {
		t.Errorf("a Treasure: %d Munitions, want still 1", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B20CastCreature(t, g, opp, "Their Rock", "Artifact", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	if got := b31CountNamed(g, me.ID, "Munitions"); got != 1 {
		t.Errorf("an opponent's artifact: %d Munitions, want still 1", got)
	}
	// The Munitions leaving deals 2 to any target — from the token.
	before := opp.Life
	effectstest.B27Kill(g, munitions)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("a Munitions leaving: life %d, want %d", opp.Life, before-2)
	}
	var dealt game.Event
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventDealDamage && g.Events[i].Target == opp.ID {
			dealt = g.Events[i]
			break
		}
	}
	if dealt.Source != munitions {
		t.Errorf("the damage is dealt by the token, got source %s", dealt.Source)
	}
	if spec, _ := effects.Lookup(effectstest.B31WeaponsManufacturingOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the token carries its own ability now: the card is complete")
	}
}

func TestB31WeaponsManufacturingMunitionsBouncedStillShoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B31Push(g, me.ID, "Weapons Manufacturing", "Enchantment", effectstest.B31WeaponsManufacturingOracle, "{1}{R}", 0, 0, "R")
	bear := effectstest.B31Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.MunitionsToken(), 1) })
	munitions := effectstest.FindBattlefieldByName(g, "Munitions")
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(munitions) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("2 damage to a 2/2: it dies")
	}
}

// The ability is the token's own, so it outlives the enchantment that
// made it (the caveat this card shipped with).
func TestB31MunitionsShootsAfterWeaponsManufacturingIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wm := effectstest.B31Push(g, me.ID, "Weapons Manufacturing", "Enchantment", effectstest.B31WeaponsManufacturingOracle, "{1}{R}", 0, 0, "R")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.MunitionsToken(), 1) })
	munitions := effectstest.FindBattlefieldByName(g, "Munitions")
	effectstest.B27Kill(g, wm)
	if g.Battlefield.Contains(wm) {
		t.Fatal("setup: the enchantment should be gone")
	}
	before := opp.Life
	effectstest.B27Kill(g, munitions)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("a Munitions with no Manufacturing: life %d, want %d", opp.Life, before-2)
	}
}

// A stolen Munitions shoots for whoever controls it as it leaves.
func TestB31MunitionsShootsForItsNewController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B31Push(g, me.ID, "Weapons Manufacturing", "Enchantment", effectstest.B31WeaponsManufacturingOracle, "{1}{R}", 0, 0, "R")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.MunitionsToken(), 1) })
	munitions := effectstest.FindBattlefieldByName(g, "Munitions")
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), munitions, opp.ID, game.IndefiniteDuration(), "test — steal") {
			t.Fatal("setup: control change refused")
		}
	})
	before := me.Life
	effectstest.B27Kill(g, munitions)
	effectstest.B04WaitForPick(t, g, opp.ID)
	effectstest.PickPlayer(t, g, opp.ID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-2 {
		t.Errorf("the thief's Munitions: life %d, want %d", me.Life, before-2)
	}
}

func TestB31RapaciousGuestFeedsGrowsAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guest := effectstest.B31Push(g, me.ID, "Rapacious Guest", "Creature — Halfling Citizen", effectstest.B31RapaciousGuestOracle, "{2}{B}", 2, 2, "B")
	if !effectstest.HasEffectiveKeyword(t, g, guest, "menace") {
		t.Error("menace is printed")
	}
	bear := effectstest.B31Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	effectstest.AttackWith(t, g, opp.ID, guest, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("two creatures connect: %d Foods, want 1 (\"one or more\")", got)
	}
	// Eating the Food grows the Guest.
	food := effectstest.FindBattlefieldByName(g, "Food")
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B31AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterCount(g, guest, game.CounterPlusOne); got != 1 {
		t.Errorf("a Food sacrificed: %d +1/+1 counters, want 1", got)
	}
	// A non-Food sacrifice is not a trigger.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, guest, game.CounterPlusOne); got != 1 {
		t.Errorf("a creature sacrificed: %d counters, want still 1", got)
	}
	// Leaving drains for its power — the counter included.
	before := opp.Life
	effectstest.B27Kill(g, guest)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) {
		t.Error("only an opponent is a legal target")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a 2/2 with a +1/+1 counter leaving: life %d, want %d", opp.Life, before-3)
	}
}

func TestB31RapaciousGuestDrainsWhenBouncedToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guest := effectstest.B31Push(g, me.ID, "Rapacious Guest", "Creature — Halfling Citizen", effectstest.B31RapaciousGuestOracle, "{2}{B}", 2, 2, "B")
	before := opp.Life
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(guest) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("a bounce is a leave: life %d, want %d", opp.Life, before-2)
	}
}

func TestB31FlagstonesOfTrokairFetchesAPlainsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	flag := effectstest.B31Push(g, me.ID, "Flagstones of Trokair", "Legendary Land", effectstest.B31FlagstonesOfTrokairOracle, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, flag, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
	plains := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Snow-Covered Plains", TypeLine: "Basic Snow Land — Plains"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.B27Kill(g, flag)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Plains") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Snow-Covered Plains") == uuid.Nil {
		t.Error("every land with the Plains type is offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest is not a Plains card")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, plains)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(plains) || !effectstest.B31Tapped(t, g, plains) {
		t.Error("the Plains is put onto the battlefield tapped")
	}
}

func TestB31FlagstonesOfTrokairBouncedOrDeclinedSearchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	flag := effectstest.B31Push(g, me.ID, "Flagstones of Trokair", "Legendary Land", effectstest.B31FlagstonesOfTrokairOracle, "", 0, 0)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(flag) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("a bounce is not \"put into a graveyard from the battlefield\"")
	}
	flag2 := effectstest.B31Push(g, me.ID, "Flagstones of Trokair", "Legendary Land", effectstest.B31FlagstonesOfTrokairOracle, "", 0, 0)
	effectstest.B27Kill(g, flag2)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil || effectstest.FindBattlefieldByName(g, "Plains") != uuid.Nil {
		t.Error("declined: no search")
	}
}

func TestB31BlossomingTortoiseMillsAndReturnsALandOnEntryAndAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swamp := effectstest.SeedGraveyardCard(t, g, "Swamp", "Basic Land — Swamp", "")
	effectstest.SeedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	library := me.Library.Size()
	tortoise := effectstest.CastAndResolveCreature(t, g, "Blossoming Tortoise", "Creature — Turtle", effectstest.B31BlossomingTortoiseOracle)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if p == nil || !effectstest.HasID(p.ChooseCards, swamp) || len(p.ChooseCards) != 1 || p.ChooseMin != 1 {
		t.Fatalf("only land cards in your graveyard are offered, and the return is not optional: %+v", p)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, swamp)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library-3 {
		t.Errorf("mill three: library %d → %d", library, got)
	}
	if !g.Battlefield.Contains(swamp) || !effectstest.B31Tapped(t, g, swamp) {
		t.Error("the chosen land returns tapped")
	}
	// The attack does it again — with an empty land pile, the mill
	// alone.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	library = me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, tortoise)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no land card: no pick")
	}
	if got := me.Library.Size(); got != library-3 {
		t.Errorf("the attack mills three: library %d → %d", library, got)
	}
	if spec, _ := effects.Lookup(effectstest.B31BlossomingTortoiseOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the land is chosen after the mill and the discount is in: nothing is left to declare")
	}
}

// The land is chosen after the mill, so one the trigger itself milled
// is the only legal pick when the graveyard held none before.
func TestB31BlossomingTortoiseCanReturnALandItJustMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	milled := effectstest.B10LibraryTop(me, "Milled Swamp", "Basic Land — Swamp", "", 0, 0)
	effectstest.B10LibraryTop(me, "Filler A", "Instant", "{R}", 0, 0)
	effectstest.B10LibraryTop(me, "Filler B", "Instant", "{R}", 0, 0)
	effectstest.CastAndResolveCreature(t, g, "Blossoming Tortoise", "Creature — Turtle", effectstest.B31BlossomingTortoiseOracle)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if p == nil || !effectstest.HasID(p.ChooseCards, milled) {
		t.Fatal("a land the trigger milled is offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	if !g.Battlefield.Contains(milled) || !effectstest.B31Tapped(t, g, milled) {
		t.Error("the milled land returns tapped")
	}
}

// Activated abilities of lands you control cost {1} less: Mount Doom's
// {1}{B}{R} is {B}{R} with the Tortoise out. From a pool of exactly three,
// one is left with the Tortoise and none without it; the discount is only
// generic mana.
func TestB31BlossomingTortoiseDiscountsLandActivations(t *testing.T) {
	for _, withTortoise := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withTortoise {
			effectstest.B31Push(g, me.ID, "Blossoming Tortoise", "Creature — Turtle", effectstest.B31BlossomingTortoiseOracle, "{2}{G}{G}", 3, 3, "G")
		}
		doom := effectstest.PushCatalogPermanent(g, me.ID, "Mount Doom", "Legendary Land", "995c8dac-fd27-468a-abd4-02372cf0c850", false)
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatForTest(g, me, "BRC")
		lifeBefore := opp.Life
		if err := g.ActivateCatalogAbility(me.ID, doom, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("Mount Doom (Tortoise %v): %v", withTortoise, err)
		}
		want := 0
		if withTortoise {
			want = 1
		}
		if me.ManaPool.Total() != want {
			t.Errorf("Tortoise %v: %d mana left, want %d", withTortoise, me.ManaPool.Total(), want)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != lifeBefore-1 {
			t.Errorf("Tortoise %v: the ability resolved, opponent life %d want %d", withTortoise, opp.Life, lifeBefore-1)
		}
	}
}

func TestB31BlossomingTortoiseMillPastTheLibraryNeverLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for me.Library.Size() > 1 {
		me.Library.Cards = me.Library.Cards[1:]
	}
	effectstest.CastAndResolveCreature(t, g, "Blossoming Tortoise", "Creature — Turtle", effectstest.B31BlossomingTortoiseOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 0 {
		t.Error("the one card is milled")
	}
	if me.AttemptedEmptyDraw {
		t.Error("a mill does not lose the game")
	}
}

// --- the activations -----------------------------------------------

func TestB31SealOfCleansingSacrificesToDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seal := effectstest.B31Push(g, me.ID, "Seal of Cleansing", "Enchantment", effectstest.B31SealOfCleansingOracle, "{1}{W}", 0, 0, "W")
	rock := effectstest.B31Push(g, opp.ID, "Their Rock", "Artifact", "", "{1}", 0, 0)
	bear := effectstest.B31Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	effectstest.B16Activate(t, g, me.ID, seal, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)})
	if g.Battlefield.Contains(seal) {
		t.Error("the Seal is sacrificed to pay")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
}

func TestB31NephaliaDrownyardMillsATargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	yard := effectstest.B31Push(g, me.ID, "Nephalia Drownyard", "Land", effectstest.B31NephaliaDrownyardOracle, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, yard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool = nil
	effectstest.B22Untap(g, yard)
	library := opp.Library.Size()
	effectstest.B31AddMana(me, "C", "U", "B")
	effectstest.B16Activate(t, g, me.ID, yard, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID)})
	if got := opp.Library.Size(); got != library-3 {
		t.Errorf("target player mills three: library %d → %d", library, got)
	}
	if !effectstest.B31Tapped(t, g, yard) {
		t.Error("the ability taps")
	}
	// A near-empty library is milled out without losing the game.
	effectstest.B22Untap(g, yard)
	for opp.Library.Size() > 2 {
		opp.Library.Cards = opp.Library.Cards[1:]
	}
	effectstest.B31AddMana(me, "C", "U", "B")
	effectstest.B16Activate(t, g, me.ID, yard, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID)})
	if opp.Library.Size() != 0 || opp.AttemptedEmptyDraw {
		t.Error("two cards milled, no loss flagged")
	}
}

func TestB31GemstoneMineTapsThreeTimesThenIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.PlayLandFromHand(t, g, "Gemstone Mine", effectstest.B31GemstoneMineOracle)
	if got := effectstest.CounterCount(g, mine, "mining"); got != 3 {
		t.Fatalf("enters with %d mining counters, want 3", got)
	}
	if effectstest.B31Tapped(t, g, mine) {
		t.Error("it enters untapped")
	}
	for i := 3; i >= 1; i-- {
		if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("activation with %d counters: %v", i, err)
		}
		if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "G"); n != 1 {
			t.Fatalf("any colour is a pick: %d prompts", n)
		}
		if i > 1 {
			if got := effectstest.CounterCount(g, mine, "mining"); got != i-1 {
				t.Errorf("after an activation: %d counters, want %d", got, i-1)
			}
			if !effectstest.B31Tapped(t, g, mine) {
				t.Error("the ability taps")
			}
			effectstest.B22Untap(g, mine)
		}
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 {
		t.Errorf("three activations: pool %v", got)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("with no mining counters left the Mine is sacrificed")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("it went to the graveyard")
	}
}

func TestB31GemstoneMineWithNoCountersCannotActivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B31Push(g, me.ID, "Gemstone Mine", "Land", effectstest.B31GemstoneMineOracle, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("no mining counter: the cost cannot be paid")
	}
	if effectstest.B31Tapped(t, g, mine) || !g.Battlefield.Contains(mine) {
		t.Error("a refused activation touches nothing")
	}
}

func TestB31SoulsMajestyDrawsPowerCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.PushCounterCreature(g, me.ID, "Big", game.CounterPlusOne, 2)
	theirs := effectstest.B31Push(g, opp.ID, "Their Big", "Creature — Bear", "", "{1}{G}", 5, 5, "G")
	if effectstest.B03CastRefused(t, g, "Soul's Majesty", "Sorcery", effectstest.B31SoulsMajestyOracle, game.CastSpellParams{Targets: effectstest.B16TargetCard(theirs)}) == false {
		t.Fatal("an opponent's creature is not \"you control\"")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Soul's Majesty", "Sorcery", effectstest.B31SoulsMajestyOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+6 {
		t.Errorf("a 4/4 with two +1/+1 counters: hand %d → %d, want +6", hand, got)
	}
}

func TestB31FainTheBrokerTradesCreaturesAndArtifactsAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fain := effectstest.B31Push(g, me.ID, "Fain, the Broker", "Legendary Creature — Human Warlock", effectstest.B31FainTheBrokerOracle, "{2}{B}", 3, 3, "B")
	fodder := effectstest.B31Push(g, me.ID, "Fodder", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	target := effectstest.B31Push(g, me.ID, "Target", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	rock := effectstest.B31Push(g, me.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, fain, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(target)})
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature is sacrificed to pay")
	}
	if got := effectstest.CounterCount(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if !effectstest.B31Tapped(t, g, fain) {
		t.Fatal("the ability taps")
	}
	// {3}{B}: untap.
	effectstest.B31AddMana(me, "C", "C", "C", "B")
	effectstest.B16Activate(t, g, me.ID, fain, 3, game.ActivateAbilityParams{})
	if effectstest.B31Tapped(t, g, fain) {
		t.Fatal("Fain untaps")
	}
	// The artifact half.
	effectstest.B16Activate(t, g, me.ID, fain, 2, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed to pay")
	}
	inkling := effectstest.FindBattlefieldByName(g, "Inkling")
	if inkling == uuid.Nil {
		t.Fatal("an Inkling token")
	}
	c := effectstest.CardByID(g, inkling)
	if c.Power != 2 || c.Toughness != 1 || len(c.Colors) != 2 || !effectstest.HasEffectiveKeyword(t, g, inkling, "flying") {
		t.Errorf("a 2/1 white and black flier, got %d/%d %v", c.Power, c.Toughness, c.Colors)
	}
	// #625: the Treasure ability — the fourth — is covered in
	// counter_cost_cards_test.go.
	if abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.B31FainTheBrokerOracle}); len(abilities) != 4 {
		t.Errorf("all four abilities are wired, got %d", len(abilities))
	}
	if spec, _ := effects.Lookup(effectstest.B31FainTheBrokerOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Fain has no remaining gap and should declare CompletenessFull")
	}
}
