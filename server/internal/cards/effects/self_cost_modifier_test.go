package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// self_cost_modifier_test.go — #746: the cards that change their OWN
// cost (ADR 0048 addendum). The engine's ordering, floor, zone, face
// and target rules are pinned in game/self_cost_modifier_test.go; these
// tests pin the DECLARATIONS — that each card counts the right things
// for the right player — and the effects that ship with them.

const (
	thoughtMonitorOracle    = "9deded8b-cec4-4ede-a50b-131404d456d4"
	myrEnforcerOracle       = "2d8e1054-654f-42e8-8c29-12c3cf13a3eb"
	thoughtcastOracle       = "cce9bbff-82dc-4b2f-addd-d6715588de20"
	voyageHomeOracle        = "5c0cbc44-6c31-44fe-a3da-a97a37a01726"
	ghaltaPrimalOracle      = "b0b6be0c-41cf-4757-9f0e-87227b6ba6b3"
	vanquishHordeOracle     = "a332e80a-dc51-4dc6-bc85-e114a1c6fdb8"
	fireballOracle          = "aa7714b0-2bfb-458a-8ebf-37ec2c53383e"
	callCoppercoatsOracle   = "73d8c33d-916a-4220-96ae-9622aad36210"
	wingedWordsOracle       = "c623aeb1-e6d4-48fe-bd2a-a7a6729aa4df"
	wizardsRetortOracle     = "b828251c-86a9-454f-9852-d0876d0f5153"
	intoTheStoryOracle      = "f290c2e4-ab52-44d4-bdeb-31aeb835b18d"
	hourOfRevelationOracle  = "920c1dd4-c3f6-4020-a6f0-2e8acad2c212"
	mortalitySpearOracle    = "544acde0-850d-4c4f-8389-e22930d87345"
	curtainsCallOracle      = "cfb1b64b-bf78-4015-ac29-cdcdf2adaa02"
	priceOfFameOracle       = "45d5cd4c-7285-4507-8cbf-eace7a734f41"
	shadowOfMortalityOracle = "45378874-aa51-4bb6-a161-be336c164778"
	witherbloomOracle       = "fbb04a21-e513-4317-aac8-fa6df91c3438"
	mycosynthGolemOracle    = "ebcd864a-b7dd-4330-89e1-80576a9437ec"
	hamzaOracle             = "0cde6a29-d597-4feb-8b2b-528d2d57a1f3"
	ancientStoneIdolOracle  = "5f981cca-5cd9-49e4-ab1a-bbbf6fc7e737"
	hagraMaulingOracle      = "37783ce6-af58-4ef6-8ab4-587079970307"
)

// selfPriced is what the board charges `seat` to cast a card carrying
// `oracle` from `from` with `targets`, through the public pricing
// surface the auto-tap preview uses.
func selfPriced(t *testing.T, g *game.Game, seat *game.Player, oracle, typeLine, manaCost string, from game.ZoneKind, targets []game.TargetRef) game.ParsedCost {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", manaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: "priced", TypeLine: typeLine, ManaCost: manaCost,
			OracleID: oracle, Owner: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   from,
		Targets:    targets,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers(%s): %v", oracle, err)
	}
	return out
}

// selfPricedMV is selfPriced from hand with no targets, as a mana value.
func selfPricedMV(t *testing.T, g *game.Game, seat *game.Player, oracle, typeLine, manaCost string) int {
	t.Helper()
	return selfPriced(t, g, seat, oracle, typeLine, manaCost, game.ZoneHand, nil).ManaValue()
}

// pushPermanent drops a plain permanent on the battlefield.
func pushPermanent(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func TestSelfCostModifiersAreWired(t *testing.T) {
	for _, oracle := range []string{
		thoughtMonitorOracle, myrEnforcerOracle, thoughtcastOracle, voyageHomeOracle,
		ghaltaPrimalOracle, vanquishHordeOracle, blasphemousActOracle, fireballOracle,
		callCoppercoatsOracle, wingedWordsOracle, wizardsRetortOracle, intoTheStoryOracle,
		hourOfRevelationOracle, mortalitySpearOracle, curtainsCallOracle, priceOfFameOracle,
		shadowOfMortalityOracle, witherbloomOracle, mycosynthGolemOracle, hamzaOracle,
		ancientStoneIdolOracle, hagraMaulingOracle,
	} {
		if len(game.SelfCostModifiersFor(game.Card{OracleID: oracle})) == 0 {
			t.Errorf("%s: no self cost modifier reached the engine", oracle)
		}
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull {
			t.Errorf("%s: want a registered, complete card", oracle)
		}
	}
	// The overwhelming majority of cards declare none.
	if game.SelfCostModifiersFor(game.Card{OracleID: faithlessLootingOracle}) != nil {
		t.Error("Faithless Looting has no self cost modifier")
	}
	// Only the per-target ones carry a note for the X picker.
	if got := game.TargetPricedCostClauses(fireballOracle); len(got) != 1 {
		t.Errorf("Fireball's X-picker note = %v, want its surcharge clause", got)
	}
	if got := game.TargetPricedCostClauses(ghaltaPrimalOracle); got != nil {
		t.Errorf("Ghalta's price does not depend on targets, note = %v", got)
	}
}

// Affinity for artifacts counts the caster's artifacts and nobody
// else's (CR 702.41a), across the four affinity cards.
func TestAffinityForArtifactsCountsYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	pushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})

	for _, tc := range []struct {
		oracle, typeLine, cost string
		want                   int
	}{
		{thoughtMonitorOracle, "Artifact Creature — Construct", "{6}{U}", 4},
		{myrEnforcerOracle, "Artifact Creature — Myr", "{7}", 4},
		{thoughtcastOracle, "Sorcery", "{4}{U}", 2},
		{voyageHomeOracle, "Sorcery", "{5}{W}{U}", 4},
		{mycosynthGolemOracle, "Artifact Creature — Golem", "{11}", 8},
	} {
		if got := selfPricedMV(t, g, me, tc.oracle, tc.typeLine, tc.cost); got != tc.want {
			t.Errorf("%s at %s with three artifacts: %d, want %d", tc.oracle, tc.cost, got, tc.want)
		}
	}
	// The opponent has one artifact: Thought Monitor costs them six.
	if got := selfPricedMV(t, g, opp, thoughtMonitorOracle, "Artifact Creature — Construct", "{6}{U}"); got != 6 {
		t.Errorf("Thought Monitor for the opponent: %d, want 6", got)
	}
}

// Thought Monitor still draws two when it enters.
func TestThoughtMonitorDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Thought Monitor", "Artifact Creature — Construct", thoughtMonitorOracle, nil)
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("Thought Monitor ETB: hand %d → %d, want +2", before, got)
	}
}

// Ghalta: total power, floored at {G}{G}; zero power reduces nothing.
func TestGhaltaCostsTotalPowerLessDownToGG(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := selfPricedMV(t, g, me, ghaltaPrimalOracle, "Legendary Creature — Elder Dinosaur", "{10}{G}{G}"); got != 12 {
		t.Errorf("Ghalta with no creatures: %d, want 12", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Beast", Power: 5, Toughness: 5})
	pushPermanent(g, me.ID, game.Card{Name: "Medium", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	pushPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Beast", Power: 9, Toughness: 9})
	got := selfPriced(t, g, me, ghaltaPrimalOracle, "Legendary Creature — Elder Dinosaur", "{10}{G}{G}", game.ZoneHand, nil)
	if got.Generic != 2 || len(got.Required) != 2 {
		t.Errorf("Ghalta with 8 power: %+v, want {2}{G}{G}", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Huge", TypeLine: "Creature — Beast", Power: 20, Toughness: 20})
	got = selfPriced(t, g, me, ghaltaPrimalOracle, "Legendary Creature — Elder Dinosaur", "{10}{G}{G}", game.ZoneHand, nil)
	if got.Generic != 0 || len(got.Required) != 2 {
		t.Errorf("Ghalta with 28 power: %+v, want exactly {G}{G}", got)
	}
}

// Ghalta under Sphere of Resistance: the Sphere's {1} lands first and
// the reduction eats it (CR 601.2f).
func TestGhaltaUnderSphereOfResistance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sphere of Resistance", "Artifact", sphereOfResistanceOracle, false)
	pushPermanent(g, me.ID, game.Card{Name: "Huge", TypeLine: "Creature — Beast", Power: 11, Toughness: 11})
	got := selfPriced(t, g, me, ghaltaPrimalOracle, "Legendary Creature — Elder Dinosaur", "{10}{G}{G}", game.ZoneHand, nil)
	if got.Generic != 0 || len(got.Required) != 2 {
		t.Errorf("Ghalta, 11 power, under a Sphere: %+v, want {G}{G}", got)
	}
}

// Blasphemous Act and Vanquish the Horde count every creature on the
// battlefield, whoever controls it.
func TestSweepersCostLessPerCreatureOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		seedCreature(g, "Mine", me.ID)
		seedCreature(g, "Theirs", opp.ID)
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	if got := selfPricedMV(t, g, me, vanquishHordeOracle, "Sorcery", "{6}{W}{W}"); got != 2 {
		t.Errorf("Vanquish the Horde with six creatures: %d, want 2 ({W}{W})", got)
	}
	if got := selfPricedMV(t, g, me, blasphemousActOracle, "Sorcery", "{8}{R}"); got != 3 {
		t.Errorf("Blasphemous Act with six creatures: %d, want 3 ({2}{R})", got)
	}
	for i := 0; i < 10; i++ {
		seedCreature(g, "More", opp.ID)
	}
	got := selfPriced(t, g, me, blasphemousActOracle, "Sorcery", "{8}{R}", game.ZoneHand, nil)
	if got.Generic != 0 || len(got.Required) != 1 {
		t.Errorf("Blasphemous Act with sixteen creatures: %+v, want {R}", got)
	}
}

// Fireball's price is per target; its damage is divided evenly among
// the targets still legal, rounded down.
func TestFireballPricesPerTargetAndDividesDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	refs := []game.TargetRef{{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID}, {Kind: game.TargetPlayer, ID: c.ID}}
	for n, want := range []int{0, 0, 1, 2} {
		got := selfPriced(t, g, me, fireballOracle, "Sorcery", "{X}{R}", game.ZoneHand, refs[:n])
		if got.Generic != want || got.XSlots != 1 {
			t.Errorf("Fireball at %d targets: %+v, want {X}{%d}{R}", n, got, want)
		}
	}

	beforeA, beforeB := a.Life, b.Life
	castXSpell(t, g, "Fireball", "Sorcery", fireballOracle, "{X}{R}", 7, refs[:2])
	passPriorityAroundTable(t, g)
	if a.Life != beforeA-3 || b.Life != beforeB-3 {
		t.Errorf("Fireball X=7 at two players: %d→%d and %d→%d, want 3 each (rounded down)", beforeA, a.Life, beforeB, b.Life)
	}

	beforeA, beforeB, beforeC := a.Life, b.Life, c.Life
	castXSpell(t, g, "Fireball", "Sorcery", fireballOracle, "{X}{R}", 2, refs)
	passPriorityAroundTable(t, g)
	if a.Life != beforeA || b.Life != beforeB || c.Life != beforeC {
		t.Error("Fireball X=2 at three targets: each share rounds to zero and nothing is dealt")
	}

	// Zero targets is a legal cast that does nothing.
	beforeMe := me.Life
	beforeA, beforeB, beforeC = a.Life, b.Life, c.Life
	castXSpell(t, g, "Fireball", "Sorcery", fireballOracle, "{X}{R}", 3, nil)
	passPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 {
		t.Errorf("Fireball with no targets did not resolve: stack size %d", g.Stack.Size())
	}
	if me.Life != beforeMe || a.Life != beforeA || b.Life != beforeB || c.Life != beforeC {
		t.Error("Fireball with no targets changed a life total")
	}
}

// Call the Coppercoats: strive's {1}{W} per extra target, and a token
// for each creature the targeted opponents control.
func TestCallTheCoppercoatsStriveAndTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	two := []game.TargetRef{{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID}}
	got := selfPriced(t, g, me, callCoppercoatsOracle, "Instant", "{2}{W}", game.ZoneHand, two)
	whites := 0
	for _, r := range got.Required {
		if len(r.Options) == 1 && r.Options[0] == "W" {
			whites++
		}
	}
	if got.Generic != 3 || whites != 2 {
		t.Errorf("Call the Coppercoats at two targets: %+v, want {3}{W}{W}", got)
	}

	seedCreature(g, "A1", a.ID)
	seedCreature(g, "A2", a.ID)
	seedCreature(g, "B1", b.ID)
	seedCreature(g, "C1", c.ID) // not targeted
	castCatalogSpell(t, g, "Call the Coppercoats", "Instant", callCoppercoatsOracle, two)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefieldByName(g, "Human Soldier", me.ID); n != 3 {
		t.Errorf("Call the Coppercoats at two opponents with three creatures: %d tokens, want 3", n)
	}
}

// The conditional reductions, each on and off.
func TestConditionalSelfReductions(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	if got := selfPricedMV(t, g, me, wingedWordsOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("Winged Words with no flyer: %d, want 3", got)
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Their Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"}})
	if got := selfPricedMV(t, g, me, wingedWordsOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("Winged Words with only an opponent's flyer: %d, want 3", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "My Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"}})
	if got := selfPricedMV(t, g, me, wingedWordsOracle, "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("Winged Words with a flyer: %d, want 2", got)
	}

	if got := selfPricedMV(t, g, me, wizardsRetortOracle, "Instant", "{1}{U}{U}"); got != 3 {
		t.Errorf("Wizard's Retort with no Wizard: %d, want 3", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Wizard", TypeLine: "Creature — Human Wizard"})
	if got := selfPricedMV(t, g, me, wizardsRetortOracle, "Instant", "{1}{U}{U}"); got != 2 {
		t.Errorf("Wizard's Retort with a Wizard: %d, want 2", got)
	}

	if got := selfPricedMV(t, g, me, intoTheStoryOracle, "Instant", "{5}{U}{U}"); got != 7 {
		t.Errorf("Into the Story with empty graveyards: %d, want 7", got)
	}
	for i := 0; i < 7; i++ {
		me.Graveyard.PushTop(game.NewCard("Mine", me.ID)) // the caster's own do not count
	}
	if got := selfPricedMV(t, g, me, intoTheStoryOracle, "Instant", "{5}{U}{U}"); got != 7 {
		t.Errorf("Into the Story with only its caster's graveyard full: %d, want 7", got)
	}
	for i := 0; i < 7; i++ {
		opp.Graveyard.PushTop(game.NewCard("Theirs", opp.ID))
	}
	if got := selfPricedMV(t, g, me, intoTheStoryOracle, "Instant", "{5}{U}{U}"); got != 4 {
		t.Errorf("Into the Story with an opponent at seven: %d, want 4", got)
	}

	for i := 0; i < 9; i++ {
		pushPermanent(g, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	// Nine rocks + three creatures above = twelve nonland permanents.
	if got := selfPricedMV(t, g, me, hourOfRevelationOracle, "Sorcery", "{3}{W}{W}{W}"); got != 3 {
		t.Errorf("Hour of Revelation with ten or more nonland permanents: %d, want 3", got)
	}

	if got := selfPricedMV(t, g, me, mortalitySpearOracle, "Instant", "{2}{B}{G}"); got != 4 {
		t.Errorf("Mortality Spear before gaining life: %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	if got := selfPricedMV(t, g, me, mortalitySpearOracle, "Instant", "{2}{B}{G}"); got != 2 {
		t.Errorf("Mortality Spear after gaining life: %d, want 2", got)
	}

	// Undaunted: three opponents at a four-player table.
	if got := selfPricedMV(t, g, me, curtainsCallOracle, "Instant", "{5}{B}"); got != 3 {
		t.Errorf("Curtains' Call with three opponents: %d, want 3", got)
	}

	if got := selfPricedMV(t, g, me, shadowOfMortalityOracle, "Creature — Avatar", "{13}{B}{B}"); got != 15 {
		t.Errorf("Shadow of Mortality above starting life: %d, want 15", got)
	}
	me.Life = game.StartingLife - 10
	if got := selfPricedMV(t, g, me, shadowOfMortalityOracle, "Creature — Avatar", "{13}{B}{B}"); got != 5 {
		t.Errorf("Shadow of Mortality ten life down: %d, want 5", got)
	}
}

// Price of Fame reads its target: legendary is {1}{B}, anything else —
// and no target at all — is the printed {3}{B}.
func TestPriceOfFameReadsItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := pushPermanent(g, opp.ID, game.Card{Name: "Legend", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	bear := pushPermanent(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }
	if got := selfPriced(t, g, me, priceOfFameOracle, "Instant", "{3}{B}", game.ZoneHand, at(legend)).ManaValue(); got != 2 {
		t.Errorf("Price of Fame at a legend: %d, want 2", got)
	}
	if got := selfPriced(t, g, me, priceOfFameOracle, "Instant", "{3}{B}", game.ZoneHand, at(bear)).ManaValue(); got != 4 {
		t.Errorf("Price of Fame at a bear: %d, want 4", got)
	}
	if got := selfPricedMV(t, g, me, priceOfFameOracle, "Instant", "{3}{B}"); got != 4 {
		t.Errorf("Price of Fame with no target: %d, want 4", got)
	}
}

// Witherbloom, Mycosynth Golem and Hamza use both slots: their own
// reduction, and a battlefield grant to other spells.
func TestBothSlotCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Counters: map[string]int{"+1/+1": 1}})
	}
	pushPermanent(g, me.ID, game.Card{Name: "Myr", TypeLine: "Artifact Creature — Myr", Power: 1, Toughness: 1})

	if got := selfPricedMV(t, g, me, witherbloomOracle, "Legendary Creature — Elder Dragon", "{6}{B}{G}"); got != 5 {
		t.Errorf("Witherbloom with three creatures: %d, want 5", got)
	}
	if got := selfPricedMV(t, g, me, hamzaOracle, "Legendary Creature — Elephant Warrior", "{4}{G}{W}"); got != 4 {
		t.Errorf("Hamza with two countered creatures: %d, want 4", got)
	}

	// The grants do nothing from a hand...
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("sorcery with Witherbloom not in play: %d, want 3", got)
	}
	pushCatalogPermanent(g, me.ID, "Witherbloom, the Balancer", "Legendary Creature — Elder Dragon", witherbloomOracle, false)
	pushCatalogPermanent(g, me.ID, "Mycosynth Golem", "Artifact Creature — Golem", mycosynthGolemOracle, false)
	pushCatalogPermanent(g, me.ID, "Hamza, Guardian of Arashin", "Legendary Creature — Elephant Warrior", hamzaOracle, false)
	// ...and from the battlefield: six creatures now (the three seeded
	// plus Witherbloom, Golem and Hamza, which are creatures too), and
	// two artifacts (Myr and Golem; Hamza is not one).
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{4}{U}"); got != 1 {
		t.Errorf("five-mana sorcery with Witherbloom and six creatures: %d, want 1 ({U})", got)
	}
	if got := priceInHand(t, g, opp, "Divination", "Sorcery", "{4}{U}"); got != 5 {
		t.Errorf("an opponent's sorcery under my Witherbloom: %d, want 5", got)
	}
	// An artifact creature spell: Golem grants affinity for my two
	// artifacts, Hamza {2} for two countered creatures.
	if got := priceInHand(t, g, me, "Construct", "Artifact Creature — Construct", "{6}"); got != 2 {
		t.Errorf("artifact creature spell under Golem and Hamza: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Signet", "Artifact", "{4}"); got != 4 {
		t.Errorf("a noncreature artifact spell gets neither grant: %d, want 4", got)
	}
}

// Ancient Stone Idol counts attacking creatures on both sides.
func TestAncientStoneIdolCountsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := selfPricedMV(t, g, me, ancientStoneIdolOracle, "Artifact Creature — Golem", "{10}"); got != 10 {
		t.Errorf("Ancient Stone Idol outside combat: %d, want 10", got)
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Raider", TypeLine: "Creature — Orc", AttackingTarget: me.ID})
	pushPermanent(g, opp.ID, game.Card{Name: "Raider", TypeLine: "Creature — Orc", AttackingTarget: me.ID})
	pushPermanent(g, opp.ID, game.Card{Name: "Home", TypeLine: "Creature — Orc"})
	if got := selfPricedMV(t, g, me, ancientStoneIdolOracle, "Artifact Creature — Golem", "{10}"); got != 8 {
		t.Errorf("Ancient Stone Idol with two attackers: %d, want 8", got)
	}
}

// Hagra Mauling: the reduction is on the front face only, and needs an
// opponent with no basic land.
func TestHagraMaulingFrontFaceOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, hagraMaulingOracle, "Instant", "{2}{B}{B}"); got != 3 {
		t.Errorf("Hagra Mauling with basicless opponents: %d, want 3", got)
	}
	for _, opp := range g.Seats[1:] {
		pushPermanent(g, opp.ID, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	}
	pushPermanent(g, me.ID, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	if got := selfPricedMV(t, g, me, hagraMaulingOracle, "Instant", "{2}{B}{B}"); got != 4 {
		t.Errorf("Hagra Mauling when every opponent has a basic: %d, want 4", got)
	}
	back := game.Card{OracleID: hagraMaulingOracle, ActiveFace: 1}
	if mods := game.SelfCostModifiersFor(back); mods != nil {
		t.Errorf("Hagra Broodpit (face 1) must not carry the front's reduction: %d modifiers", len(mods))
	}
}

// "For each card in your graveyard" read during a graveyard cast does
// not count the card being cast (ADR 0048 addendum §12).
func TestGraveyardCountExcludesTheCardBeingCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	self := game.NewCard("Self", me.ID)
	me.Graveyard.PushTop(self)
	me.Graveyard.PushTop(game.NewCard("Other", me.ID))
	count := CountCardsInYourGraveyard(nil)
	q := game.CostQuery{Game: g, Card: self, Controller: me.ID, FromZone: game.ZoneGraveyard}
	if got := count(q); got != 1 {
		t.Errorf("graveyard count during its own cast: %d, want 1", got)
	}
}
