package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch31_test.go — card-level coverage for the card-coverage
// roadmap's batch 31 (#394, `edhrec_rank` 3245–3345): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, damage, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b31-prefixed.

const (
	b31GoblinInstigatorOracle     = "8b022754-6d16-470e-b754-4df6e4f4709e"
	b31BlossomingTortoiseOracle   = "5e1bf23b-7fb0-45ff-8544-fce9fa3eba00"
	b31SealOfCleansingOracle      = "a75dbe70-7e3e-446f-9a76-9fbb414f2e7c"
	b31LeoninWarleaderOracle      = "8b1351e6-165e-4ca3-96d5-4774b3176362"
	b31KinjallisSunwingOracle     = "b9243163-b726-432e-830f-86132aa7f34a"
	b31SunscorchedDesertOracle    = "256b8c23-589e-429d-9e6e-433d55079eb4"
	b31TheWarringTriadOracle      = "df0b5995-117b-4ba8-964d-ea3c592621ef"
	b31FainTheBrokerOracle        = "31b990e3-9bad-4b0a-a9d5-f5b9ed2ad0b0"
	b31SoulsMajestyOracle         = "50fd4cf3-c347-4eb9-ab15-a9b0c5ea8b0f"
	b31TillerEngineOracle         = "62f6baba-da3b-45b8-a3c1-efb75763cca8"
	b31RapaciousGuestOracle       = "54accd4a-b471-4ae5-b3b2-a5cec44023b7"
	b31NephaliaDrownyardOracle    = "6429b4ed-1845-4643-9a3d-85f7c12f2bba"
	b31TheEndstoneOracle          = "57ee0730-3ac5-4f17-ac82-bd4a6673c2bb"
	b31VampireOfTheDireMoonOracle = "5a938371-8428-48e3-85ed-6a84a39b29c6"
	b31GemstoneMineOracle         = "0c828f10-4775-492f-9224-1e2814ad2cad"
	b31FlagstonesOfTrokairOracle  = "f73979bb-91a5-4388-b70b-0cd7a4e14291"
	b31WeaponsManufacturingOracle = "e48a8160-cffe-4e00-a3e9-a59d7fc7b3a2"
	// Already on main from the S26 tribal work (tribal_lords.go).
	b31GoblinKingAlreadyOracle = "d236b3fc-0d3f-4d99-875d-e32a33fe5767"
)

// b31Push seeds a permanent with a mana cost, colours and P/T on the
// battlefield with a layer timestamp and no summoning sickness.
func b31Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b31PlayLand has the active seat play a land with the given type
// line from hand — playLandFromHand with the type line as a
// parameter, for a Desert.
func b31PlayLand(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	return b13PlayAs(t, g, g.Turn.ActiveSeat, name, typeLine, oracle)
}

// b31Tapped reads a battlefield card's tapped state.
func b31Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// b31AddMana adds coloured mana to a player's pool.
func b31AddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

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
		b31GoblinInstigatorOracle:     "Goblin Instigator",
		b31BlossomingTortoiseOracle:   "Blossoming Tortoise",
		b31SealOfCleansingOracle:      "Seal of Cleansing",
		b31LeoninWarleaderOracle:      "Leonin Warleader",
		b31KinjallisSunwingOracle:     "Kinjalli's Sunwing",
		b31SunscorchedDesertOracle:    "Sunscorched Desert",
		b31TheWarringTriadOracle:      "The Warring Triad",
		b31FainTheBrokerOracle:        "Fain, the Broker",
		b31SoulsMajestyOracle:         "Soul's Majesty",
		b31TillerEngineOracle:         "Tiller Engine",
		b31RapaciousGuestOracle:       "Rapacious Guest",
		b31NephaliaDrownyardOracle:    "Nephalia Drownyard",
		b31TheEndstoneOracle:          "The Endstone",
		b31VampireOfTheDireMoonOracle: "Vampire of the Dire Moon",
		b31GemstoneMineOracle:         "Gemstone Mine",
		b31FlagstonesOfTrokairOracle:  "Flagstones of Trokair",
		b31WeaponsManufacturingOracle: "Weapons Manufacturing",
		b31GoblinKingAlreadyOracle:    "Goblin King",
	}
	if len(want) != 18 {
		t.Fatalf("the batch registers 17 cards plus one already on main, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
	// The six declared skips must stay out until their seam lands: a
	// mana ability granted to OTHER permanents by a static (Manaweft
	// Sliver), a put-a-card-from-hand-onto-the-battlefield prompt
	// (Oviya), casting from the top of the library (Conspicuous
	// Snoop), a filtered pick from a revealed hand (Duress), a random
	// choice reachable from an effect (Deadbridge Chant), and a
	// continuous effect with no duration (Tree of Redemption).
	for oracle, name := range map[string]string{
		"bd47398d-da35-4a09-8754-771af91b14f4": "Manaweft Sliver",
		"b01649b8-b0b5-4cf6-aea7-291f3407859e": "Oviya, Automech Artisan",
		"f5d1bd3c-0e65-4999-a810-881b2389b40e": "Conspicuous Snoop",
		"33d405ea-7a9a-4970-b70f-9c05d90dd6f0": "Duress",
		"506667b1-7922-4959-a5b3-0f8abe8c3615": "Deadbridge Chant",
		"1372dd15-cdb9-470c-b3e8-3423697e46e8": "Tree of Redemption",
	} {
		if _, ok := Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #394 but is registered — update the issue", name)
		}
	}
}

// --- the keywords and statics --------------------------------------

func TestB31VampireOfTheDireMoonHasDeathtouchAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := b31Push(g, me.ID, "Vampire of the Dire Moon", "Creature — Vampire", b31VampireOfTheDireMoonOracle, "{B}", 1, 1, "B")
	if !hasEffectiveKeyword(t, g, vamp, "deathtouch") || !hasEffectiveKeyword(t, g, vamp, "lifelink") {
		t.Fatal("deathtouch and lifelink are printed")
	}
	before := me.Life
	attackWith(t, g, opp.ID, vamp)
	if me.Life != before+1 {
		t.Errorf("lifelink: life %d, want %d", me.Life, before+1)
	}
}

func TestB31TheWarringTriadWakesAtEightCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b31GraveyardCards(me, 7)
	triad := b31Push(g, me.ID, "The Warring Triad", "Legendary Artifact Creature — God", b31TheWarringTriadOracle, "{3}", 5, 5)
	if containsString(effectiveTypes(t, g, triad), "Creature") {
		t.Fatal("seven cards: not a creature")
	}
	if !containsString(effectiveTypes(t, g, triad), "Artifact") {
		t.Error("it is still an artifact")
	}
	if containsString(effectiveSubtypes(t, g, triad), "God") {
		t.Error("a creature type goes with the creature type (CR 205.1b)")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(triad, g.Seats[1].ID); err == nil {
		t.Error("a noncreature cannot attack")
	}
	b31GraveyardCards(me, 1)
	g.BumpLayerVersionForTest()
	if !containsString(effectiveTypes(t, g, triad), "Creature") {
		t.Fatal("eight cards: a creature")
	}
	if !containsString(effectiveSubtypes(t, g, triad), "God") {
		t.Error("the God type is back")
	}
	for _, kw := range []string{"flying", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, triad, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	if got := effectivePower(t, g, triad); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	if spec, _ := Lookup(b31TheWarringTriadOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the mill-cost gap must be declared")
	}
}

func TestB31BlossomingTortoisePumpsLandCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b31Push(g, me.ID, "Blossoming Tortoise", "Creature — Turtle", b31BlossomingTortoiseOracle, "{2}{G}{G}", 3, 3, "G")
	arbor := b31Push(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", "", "", 1, 1, "G")
	bear := b31Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	theirs := b31Push(g, opp.ID, "Their Arbor", "Land Creature — Forest Dryad", "", "", 1, 1, "G")
	if got := effectivePower(t, g, arbor); got != 2 {
		t.Errorf("a land creature you control: power %d, want 2", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a nonland creature: power %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's land creature: power %d, want 1", got)
	}
}

// --- the replacements ----------------------------------------------

func TestB31KinjallisSunwingTapsOpponentsCreaturesOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wing := b31Push(g, me.ID, "Kinjalli's Sunwing", "Creature — Dinosaur", b31KinjallisSunwingOracle, "{2}{W}", 2, 3, "W")
	if !hasEffectiveKeyword(t, g, wing, "flying") {
		t.Error("flying is printed")
	}
	mine := castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	if b31Tapped(t, g, mine) {
		t.Error("your own creature enters untapped")
	}
	advanceToMainOf(t, g, 1)
	theirs := b20CastCreature(t, g, opp, "Their Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if !b31Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
	if n := tapEventsFor(g, theirs); n != 0 {
		t.Errorf("a replacement, not a tap: %d tap events", n)
	}
	rock := b20CastCreature(t, g, opp, "Their Rock", "Artifact", "", 0, 0)
	passPriorityAroundTable(t, g)
	if b31Tapped(t, g, rock) {
		t.Error("only creatures enter tapped")
	}
}

// --- the triggers --------------------------------------------------

func TestB31GoblinInstigatorBringsAGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Goblin Instigator", "Creature — Goblin Rogue", b31GoblinInstigatorOracle)
	passPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Goblin"); got != 1 {
		t.Fatalf("%d Goblin tokens, want 1", got)
	}
	tok := cardByID(g, findBattlefieldByName(g, "Goblin"))
	if tok.Power != 1 || tok.Toughness != 1 || len(tok.Colors) != 1 || tok.Colors[0] != "R" {
		t.Errorf("a 1/1 red Goblin, got %d/%d %v", tok.Power, tok.Toughness, tok.Colors)
	}
}

func TestB31LeoninWarleaderAttacksWithTwoLifelinkCats(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warleader := b31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", b31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
	before := opp.Life
	mine := me.Life
	declareAttack(t, g, opp.ID, warleader)
	if n := triggersOnStackFrom(g, warleader); n != 1 {
		t.Fatalf("one attack trigger, got %d", n)
	}
	passPriorityAroundTable(t, g)
	cats := battlefieldIDsNamed(g, "Cat")
	if len(cats) != 2 {
		t.Fatalf("%d Cats, want 2", len(cats))
	}
	for _, id := range cats {
		c := cardByID(g, id)
		if !c.Tapped || c.AttackingTarget != opp.ID {
			t.Errorf("a Cat tapped and attacking the defender: %+v", c)
		}
		if !hasEffectiveKeyword(t, g, id, "lifelink") || effectivePower(t, g, id) != 1 {
			t.Error("a 1/1 with lifelink")
		}
	}
	// The Cats were put onto the battlefield attacking, never
	// declared: no second Warleader trigger.
	if n := triggersOnStackFrom(g, warleader); n != 0 || len(g.PendingTriggers) != 0 {
		t.Error("an attacking token is not a declared attacker")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("the Warleader and two Cats connect: life %d, want %d", opp.Life, before-6)
	}
	if me.Life != mine+2 {
		t.Errorf("two lifelink Cats: life %d, want %d", me.Life, mine+2)
	}
}

func TestB31SunscorchedDesertPingsOnEntryAndTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	desert := b31PlayLand(t, g, "Sunscorched Desert", "Land — Desert", b31SunscorchedDesertOracle)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetPlayers, opp.ID) || !hasID(p.PickTargetPlayers, me.ID) {
		t.Error("any player is a legal target")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("1 damage: life %d, want %d", opp.Life, before-1)
	}
	if err := g.ActivateManaAbility(me.ID, desert, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if c, _ := battlefieldCard(g, desert); !c.HasSubtype("Desert") {
		t.Error("it is a Desert")
	}
}

func TestB31TheEndstoneDrawsOnLandsAndSpellsAndResetsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b31Push(g, me.ID, "The Endstone", "Legendary Artifact", b31TheEndstoneOracle, "{7}", 0, 0)
	// The helpers seed the played card into the hand first, so a
	// draw shows as +1 on the size.
	hand := me.Hand.Size()
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a land play: hand %d → %d, want +1 (the draw)", hand, got)
	}
	hand = me.Hand.Size()
	b24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("a spell: hand %d → %d, want +1 (the draw)", hand, got)
	}
	// An opponent's spell is not "you cast".
	hand = me.Hand.Size()
	opp := g.Seats[1]
	b29CastAs(t, g, opp, "Opt", "Instant", "", "{U}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an opponent's spell: hand %d → %d, want unchanged", hand, got)
	}
	// The end step: life becomes 20, whichever side of it you are on.
	me.Life = 33
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != 20 {
		t.Errorf("above 20: life %d, want 20", me.Life)
	}
	// Another player's end step is not yours.
	me.Life = 33
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Life != 33 {
		t.Errorf("an opponent's end step: life %d, want 33", me.Life)
	}
	me.Life = 4
	b12ToMyNextUpkeep(t, g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != 20 {
		t.Errorf("below 20: life %d, want 20", me.Life)
	}
}

func TestB31TillerEngineUntapsOrTapsWhenALandEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := b31Push(g, me.ID, "Tiller Engine", "Artifact Creature — Construct", b31TillerEngineOracle, "{2}", 1, 3)
	// No opponent nonland permanent: the trigger is the untap alone.
	mire := playLandFromHand(t, g, "Haunted Mire", b11HauntedMireOracle)
	if !b31Tapped(t, g, mire) {
		t.Fatal("Haunted Mire enters tapped")
	}
	if triggerOnStack(g, engine) == nil {
		t.Fatal("the choice is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	if b31Tapped(t, g, mire) {
		t.Error("no nonland permanent to tap: the land is untapped")
	}
	// An opponent's nonland permanent: pick it to tap it instead.
	theirs := b31Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	theirLand := b31Push(g, opp.ID, "Their Forest", "Basic Land — Forest", "", "", 0, 0)
	b12ToMyNextUpkeep(t, g)
	mire2 := playLandFromHand(t, g, "Haunted Mire", b11HauntedMireOracle)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, theirs) || hasID(p.PickTargetCards, theirLand) || hasID(p.PickTargetCards, engine) {
		t.Error("only nonland permanents an opponent controls are offered")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !b31Tapped(t, g, theirs) {
		t.Error("the chosen permanent is tapped")
	}
	if !b31Tapped(t, g, mire2) {
		t.Error("the tap mode was chosen: the land stays tapped")
	}
	// Decline the pick: the untap mode.
	b12ToMyNextUpkeep(t, g)
	mire3 := playLandFromHand(t, g, "Haunted Mire", b11HauntedMireOracle)
	b04WaitForPick(t, g, me.ID)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if b31Tapped(t, g, mire3) {
		t.Error("no target chosen: the land is untapped")
	}
	// A land that enters untapped is not a trigger.
	b12ToMyNextUpkeep(t, g)
	playLandFromHand(t, g, "Forest", "")
	if triggerOnStack(g, engine) != nil || len(g.PendingTriggers) != 0 || latestPickTarget(g, me.ID) != nil {
		t.Error("an untapped entry: no trigger")
	}
}

func TestB31TillerEngineIgnoresOpponentsLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := b31Push(g, me.ID, "Tiller Engine", "Artifact Creature — Construct", b31TillerEngineOracle, "{2}", 1, 3)
	mire := b13PlayAs(t, g, 1, "Haunted Mire", "Land", b11HauntedMireOracle)
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, engine) != nil || !b31Tapped(t, g, mire) {
		t.Error("an opponent's land is not \"a land you control\"")
	}
	_ = opp
}

func TestB31WeaponsManufacturingArmsEveryRealArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b31Push(g, me.ID, "Weapons Manufacturing", "Enchantment", b31WeaponsManufacturingOracle, "{1}{R}", 0, 0, "R")
	b24CastNoncreature(t, g, "Sol Ring", "Artifact", "{1}")
	passPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Munitions"); got != 1 {
		t.Fatalf("an artifact spell: %d Munitions, want 1", got)
	}
	munitions := findBattlefieldByName(g, "Munitions")
	m := cardByID(g, munitions)
	if !m.IsArtifact() || m.IsCreature() || len(m.Colors) != 0 || !IsToken(m) {
		t.Errorf("a colorless artifact token, got %s %v", m.TypeLine, m.Colors)
	}
	// A token artifact does not chain, and an opponent's artifact is
	// not yours.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Munitions"); got != 1 {
		t.Errorf("a Treasure: %d Munitions, want still 1", got)
	}
	advanceToMainOf(t, g, 1)
	b20CastCreature(t, g, opp, "Their Rock", "Artifact", "", 0, 0)
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 0)
	if got := b31CountNamed(g, me.ID, "Munitions"); got != 1 {
		t.Errorf("an opponent's artifact: %d Munitions, want still 1", got)
	}
	// The Munitions leaving deals 2 to any target — from the token.
	before := opp.Life
	b27Kill(g, munitions)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
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
	if spec, _ := Lookup(b31WeaponsManufacturingOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the carried-trigger gap must be declared")
	}
}

func TestB31WeaponsManufacturingMunitionsBouncedStillShoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b31Push(g, me.ID, "Weapons Manufacturing", "Enchantment", b31WeaponsManufacturingOracle, "{1}{R}", 0, 0, "R")
	bear := b31Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("Munitions"), 1) })
	munitions := findBattlefieldByName(g, "Munitions")
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(munitions) })
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("2 damage to a 2/2: it dies")
	}
}

func TestB31RapaciousGuestFeedsGrowsAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guest := b31Push(g, me.ID, "Rapacious Guest", "Creature — Halfling Citizen", b31RapaciousGuestOracle, "{2}{B}", 2, 2, "B")
	if !hasEffectiveKeyword(t, g, guest, "menace") {
		t.Error("menace is printed")
	}
	bear := b31Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	attackWith(t, g, opp.ID, guest, bear)
	passPriorityAroundTable(t, g)
	if got := b31CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("two creatures connect: %d Foods, want 1 (\"one or more\")", got)
	}
	// Eating the Food grows the Guest.
	food := findBattlefieldByName(g, "Food")
	advanceToMainOf(t, g, 0)
	b31AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, guest, game.CounterPlusOne); got != 1 {
		t.Errorf("a Food sacrificed: %d +1/+1 counters, want 1", got)
	}
	// A non-Food sacrifice is not a trigger.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, guest, game.CounterPlusOne); got != 1 {
		t.Errorf("a creature sacrificed: %d counters, want still 1", got)
	}
	// Leaving drains for its power — the counter included.
	before := opp.Life
	b27Kill(g, guest)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) {
		t.Error("only an opponent is a legal target")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a 2/2 with a +1/+1 counter leaving: life %d, want %d", opp.Life, before-3)
	}
}

func TestB31RapaciousGuestDrainsWhenBouncedToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guest := b31Push(g, me.ID, "Rapacious Guest", "Creature — Halfling Citizen", b31RapaciousGuestOracle, "{2}{B}", 2, 2, "B")
	before := opp.Life
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(guest) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("a bounce is a leave: life %d, want %d", opp.Life, before-2)
	}
}

func TestB31FlagstonesOfTrokairFetchesAPlainsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	flag := b31Push(g, me.ID, "Flagstones of Trokair", "Legendary Land", b31FlagstonesOfTrokairOracle, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, flag, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
	plains := pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	pushLibraryCardForTest(me, game.Card{Name: "Snow-Covered Plains", TypeLine: "Basic Snow Land — Plains"})
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	b27Kill(g, flag)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Plains") == uuid.Nil || searchOptionNamed(g, c, "Snow-Covered Plains") == uuid.Nil {
		t.Error("every land with the Plains type is offered")
	}
	if searchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest is not a Plains card")
	}
	answerSearchByID(t, g, me.ID, plains)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(plains) || !b31Tapped(t, g, plains) {
		t.Error("the Plains is put onto the battlefield tapped")
	}
}

func TestB31FlagstonesOfTrokairBouncedOrDeclinedSearchesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	flag := b31Push(g, me.ID, "Flagstones of Trokair", "Legendary Land", b31FlagstonesOfTrokairOracle, "", 0, 0)
	pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(flag) })
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a bounce is not \"put into a graveyard from the battlefield\"")
	}
	flag2 := b31Push(g, me.ID, "Flagstones of Trokair", "Legendary Land", b31FlagstonesOfTrokairOracle, "", 0, 0)
	b27Kill(g, flag2)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil || findBattlefieldByName(g, "Plains") != uuid.Nil {
		t.Error("declined: no search")
	}
}

func TestB31BlossomingTortoiseMillsAndReturnsALandOnEntryAndAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swamp := seedGraveyardCard(t, g, "Swamp", "Basic Land — Swamp", "")
	seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	library := me.Library.Size()
	tortoise := castAndResolveCreature(t, g, "Blossoming Tortoise", "Creature — Turtle", b31BlossomingTortoiseOracle)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, swamp) || len(p.PickTargetCards) != 1 {
		t.Errorf("only land cards in your graveyard are offered: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, swamp)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library-3 {
		t.Errorf("mill three: library %d → %d", library, got)
	}
	if !g.Battlefield.Contains(swamp) || !b31Tapped(t, g, swamp) {
		t.Error("the chosen land returns tapped")
	}
	// The attack does it again — with an empty land pile, the mill
	// alone.
	b12ToMyNextUpkeep(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	library = me.Library.Size()
	declareAttack(t, g, opp.ID, tortoise)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no land card: no pick (the untargeted declaration)")
	}
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library-3 {
		t.Errorf("the attack mills three: library %d → %d", library, got)
	}
	if spec, _ := Lookup(b31BlossomingTortoiseOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the pick-before-mill and discount gaps must be declared")
	}
}

func TestB31BlossomingTortoiseMillPastTheLibraryNeverLoses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for me.Library.Size() > 1 {
		me.Library.Cards = me.Library.Cards[1:]
	}
	castAndResolveCreature(t, g, "Blossoming Tortoise", "Creature — Turtle", b31BlossomingTortoiseOracle)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 0 {
		t.Error("the one card is milled")
	}
	if me.AttemptedEmptyDraw {
		t.Error("a mill does not lose the game")
	}
}

// --- the activations -----------------------------------------------

func TestB31SealOfCleansingSacrificesToDestroy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seal := b31Push(g, me.ID, "Seal of Cleansing", "Enchantment", b31SealOfCleansingOracle, "{1}{W}", 0, 0, "W")
	rock := b31Push(g, opp.ID, "Their Rock", "Artifact", "", "{1}", 0, 0)
	bear := b31Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	b16Activate(t, g, me.ID, seal, 0, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	if g.Battlefield.Contains(seal) {
		t.Error("the Seal is sacrificed to pay")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
}

func TestB31NephaliaDrownyardMillsATargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	yard := b31Push(g, me.ID, "Nephalia Drownyard", "Land", b31NephaliaDrownyardOracle, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, yard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	me.ManaPool = nil
	b22Untap(g, yard)
	library := opp.Library.Size()
	b31AddMana(me, "C", "U", "B")
	b16Activate(t, g, me.ID, yard, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
	if got := opp.Library.Size(); got != library-3 {
		t.Errorf("target player mills three: library %d → %d", library, got)
	}
	if !b31Tapped(t, g, yard) {
		t.Error("the ability taps")
	}
	// A near-empty library is milled out without losing the game.
	b22Untap(g, yard)
	for opp.Library.Size() > 2 {
		opp.Library.Cards = opp.Library.Cards[1:]
	}
	b31AddMana(me, "C", "U", "B")
	b16Activate(t, g, me.ID, yard, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
	if opp.Library.Size() != 0 || opp.AttemptedEmptyDraw {
		t.Error("two cards milled, no loss flagged")
	}
}

func TestB31GemstoneMineTapsThreeTimesThenIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := playLandFromHand(t, g, "Gemstone Mine", b31GemstoneMineOracle)
	if got := counterCount(g, mine, "mining"); got != 3 {
		t.Fatalf("enters with %d mining counters, want 3", got)
	}
	if b31Tapped(t, g, mine) {
		t.Error("it enters untapped")
	}
	for i := 3; i >= 1; i-- {
		if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("activation with %d counters: %v", i, err)
		}
		if n := b10ResolveAllManaPicks(t, g, me.ID, "G"); n != 1 {
			t.Fatalf("any colour is a pick: %d prompts", n)
		}
		if i > 1 {
			if got := counterCount(g, mine, "mining"); got != i-1 {
				t.Errorf("after an activation: %d counters, want %d", got, i-1)
			}
			if !b31Tapped(t, g, mine) {
				t.Error("the ability taps")
			}
			b22Untap(g, mine)
		}
	}
	if got := batch01PoolColors(me); len(got) != 3 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b31Push(g, me.ID, "Gemstone Mine", "Land", b31GemstoneMineOracle, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("no mining counter: the cost cannot be paid")
	}
	if b31Tapped(t, g, mine) || !g.Battlefield.Contains(mine) {
		t.Error("a refused activation touches nothing")
	}
}

func TestB31SoulsMajestyDrawsPowerCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushCounterCreature(g, me.ID, "Big", game.CounterPlusOne, 2)
	theirs := b31Push(g, opp.ID, "Their Big", "Creature — Bear", "", "{1}{G}", 5, 5, "G")
	if b03CastRefused(t, g, "Soul's Majesty", "Sorcery", b31SoulsMajestyOracle, game.CastSpellParams{Targets: b16TargetCard(theirs)}) == false {
		t.Fatal("an opponent's creature is not \"you control\"")
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Soul's Majesty", "Sorcery", b31SoulsMajestyOracle, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+6 {
		t.Errorf("a 4/4 with two +1/+1 counters: hand %d → %d, want +6", hand, got)
	}
}

func TestB31FainTheBrokerTradesCreaturesAndArtifactsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fain := b31Push(g, me.ID, "Fain, the Broker", "Legendary Creature — Human Warlock", b31FainTheBrokerOracle, "{2}{B}", 3, 3, "B")
	fodder := b31Push(g, me.ID, "Fodder", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	target := b31Push(g, me.ID, "Target", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	rock := b31Push(g, me.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, fain, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Targets: b16TargetCard(target)})
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature is sacrificed to pay")
	}
	if got := counterCount(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if !b31Tapped(t, g, fain) {
		t.Fatal("the ability taps")
	}
	// {3}{B}: untap.
	b31AddMana(me, "C", "C", "C", "B")
	b16Activate(t, g, me.ID, fain, 3, game.ActivateAbilityParams{})
	if b31Tapped(t, g, fain) {
		t.Fatal("Fain untaps")
	}
	// The artifact half.
	b16Activate(t, g, me.ID, fain, 2, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed to pay")
	}
	inkling := findBattlefieldByName(g, "Inkling")
	if inkling == uuid.Nil {
		t.Fatal("an Inkling token")
	}
	c := cardByID(g, inkling)
	if c.Power != 2 || c.Toughness != 1 || len(c.Colors) != 2 || !hasEffectiveKeyword(t, g, inkling, "flying") {
		t.Errorf("a 2/1 white and black flier, got %d/%d %v", c.Power, c.Toughness, c.Colors)
	}
	// #625: the Treasure ability — the fourth — is covered in
	// counter_cost_cards_test.go.
	if abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: b31FainTheBrokerOracle}); len(abilities) != 4 {
		t.Errorf("all four abilities are wired, got %d", len(abilities))
	}
	if spec, _ := Lookup(b31FainTheBrokerOracle); spec.Completeness != CompletenessFull {
		t.Error("Fain has no remaining gap and should declare CompletenessFull")
	}
}
