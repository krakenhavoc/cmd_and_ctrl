package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_player_skip_cards_test.go — #3995's player-wide "<permanents>
// don't untap during <player>'s next untap step" (CR 502.3, CR 611.2c),
// on every card that prints it.

const (
	upsBlindingBeamOracle   = "0766cf02-ea55-4271-8c73-213c93b0baac"
	upsManaVaporsOracle     = "8e082c95-cdc4-4f29-9915-05f83298f14d"
	upsMisstepOracle        = "4ee994a1-fd3b-43cf-be3c-9c705e60754a"
	upsExhaustionOracle     = "0e7b9caf-8285-4386-98bc-9a809827f447"
	upsBontusOracle         = "c010d833-1873-4132-b4ff-8d78b3d7c2d7"
	upsRhonassOracle        = "01c729d7-9b80-479d-8d30-ae47c76c2c50"
	upsOketrasOracle        = "b5e0aa12-920d-4700-a287-5ca24933abb1"
	upsKefnetsOracle        = "67cf6ba1-55ad-4008-ba1f-96c363a96b16"
	upsIcebreakerOracle     = "01590ff3-34fa-47aa-b5da-f0b8c1b031c6"
	upsIcebreakerKrakenName = "Icebreaker Kraken"
)

func TestPlayerWideUntapSkipCardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Blinding Beam":          upsBlindingBeamOracle,
		"Mana Vapors":            upsManaVaporsOracle,
		"Misstep":                upsMisstepOracle,
		"Exhaustion":             upsExhaustionOracle,
		"Bontu's Last Reckoning": upsBontusOracle,
		"Rhonas's Last Stand":    upsRhonassOracle,
		"Oketra's Last Mercy":    upsOketrasOracle,
		"Kefnet's Last Word":     upsKefnetsOracle,
		upsIcebreakerKrakenName:  upsIcebreakerOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: Lookup = %q, %v, completeness %v", name, spec.Name, ok, spec.Completeness)
		}
	}
}

// upsTapped parks a tapped permanent.
func upsTapped(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Tapped: true,
	})
	return id
}

func upsIsTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

func upsSeatOf(g *game.Game, id uuid.UUID) int {
	for i, p := range g.Seats {
		if p != nil && p.ID == id {
			return i
		}
	}
	return -1
}

// upsAfterTheirUntap walks to the upkeep of `player`'s next turn, so
// their untap step has just happened.
func upsAfterTheirUntap(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	seat := upsSeatOf(g, player)
	if g.Turn.ActiveSeat == seat {
		// Walk off this turn first: its untap step is behind it.
		for g.Turn.ActiveSeat == seat {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
	}
	advanceToUpkeepOf(t, g, seat)
}

// --- Blinding Beam ----------------------------------------------------

func TestBlindingBeamEntwinedTapsTwoAndHoldsThatPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	land := upsTapped(g, opp.ID, "Island", "Basic Land — Island")
	if _, err := castEntwineSpell(g, me, "Blinding Beam", "Instant", "{2}{W}", upsBlindingBeamOracle, []string{"W"},
		"{C}{C}{C}{W}", []int{0, 1}, true,
		[]game.TargetRef{ewaTarget(a, 0), ewaTarget(b, 0), ewaPlayer(opp.ID, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !upsIsTapped(t, g, a) || !upsIsTapped(t, g, b) {
		t.Fatal("the tap bullet taps both creatures")
	}
	// CR 611.2c: a creature that arrives after the spell resolved is
	// held too.
	later := upsTapped(g, opp.ID, "Later Bear", "Creature — Bear")
	upsAfterTheirUntap(t, g, opp.ID)
	for _, id := range []uuid.UUID{a, b, later} {
		if !upsIsTapped(t, g, id) {
			t.Errorf("%s untapped in the target player's next untap step", id)
		}
	}
	if upsIsTapped(t, g, land) {
		t.Error("a land is not a creature: it untaps")
	}
	upsAfterTheirUntap(t, g, opp.ID)
	if upsIsTapped(t, g, a) || upsIsTapped(t, g, later) {
		t.Error("the following untap step untaps them")
	}
}

func TestBlindingBeamUnentwinedTapsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	if _, err := castEntwineSpell(g, me, "Blinding Beam", "Instant", "{2}{W}", upsBlindingBeamOracle, []string{"W"},
		"{C}{C}{W}", []int{0}, false, []game.TargetRef{ewaTarget(a, 0), ewaTarget(b, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !upsIsTapped(t, g, a) || !upsIsTapped(t, g, b) {
		t.Fatal("the tap bullet taps both creatures")
	}
	upsAfterTheirUntap(t, g, opp.ID)
	if upsIsTapped(t, g, a) || upsIsTapped(t, g, b) {
		t.Error("without the second bullet they untap as usual")
	}
}

func TestBlindingBeamSecondBulletAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	theirs := upsTapped(g, opp.ID, "Their Bear", "Creature — Bear")
	mine := upsTapped(g, me.ID, "My Bear", "Creature — Bear")
	if _, err := castEntwineSpell(g, me, "Blinding Beam", "Instant", "{2}{W}", upsBlindingBeamOracle, []string{"W"},
		"{C}{C}{W}", []int{1}, false, []game.TargetRef{ewaPlayer(opp.ID, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	upsAfterTheirUntap(t, g, opp.ID)
	if !upsIsTapped(t, g, theirs) {
		t.Error("the target player's creature untapped")
	}
	upsAfterTheirUntap(t, g, me.ID)
	if upsIsTapped(t, g, mine) {
		t.Error("the caster's creature sits out nothing")
	}
}

// --- Mana Vapors, Misstep, Exhaustion -----------------------------------

func TestManaVaporsHoldsThatPlayersLandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := cmcSeats(t, g)
	land := upsTapped(g, opp.ID, "Island", "Basic Land — Island")
	bear := upsTapped(g, opp.ID, "Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Mana Vapors", "Sorcery", upsManaVaporsOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	later := upsTapped(g, opp.ID, "Later Island", "Basic Land — Island")
	upsAfterTheirUntap(t, g, opp.ID)
	if !upsIsTapped(t, g, land) || !upsIsTapped(t, g, later) {
		t.Error("their lands, the later one included, untapped")
	}
	if upsIsTapped(t, g, bear) {
		t.Error("their creature stayed tapped")
	}
}

func TestMisstepHoldsThatPlayersCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	land := upsTapped(g, opp.ID, "Island", "Basic Land — Island")
	bear := upsTapped(g, opp.ID, "Bear", "Creature — Bear")
	mine := upsTapped(g, me.ID, "My Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Misstep", "Sorcery", upsMisstepOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	upsAfterTheirUntap(t, g, opp.ID)
	if !upsIsTapped(t, g, bear) {
		t.Error("their creature untapped")
	}
	if upsIsTapped(t, g, land) {
		t.Error("their land stayed tapped")
	}
	upsAfterTheirUntap(t, g, me.ID)
	if upsIsTapped(t, g, mine) {
		t.Error("the caster's creature stayed tapped")
	}
}

func TestExhaustionHoldsCreaturesAndLandsOfAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	land := upsTapped(g, opp.ID, "Island", "Basic Land — Island")
	bear := upsTapped(g, opp.ID, "Bear", "Creature — Bear")
	rock := upsTapped(g, opp.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Exhaustion", "Sorcery", upsExhaustionOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Exhaustion targets an opponent, not its caster")
	}
	castCatalogSpell(t, g, "Exhaustion", "Sorcery", upsExhaustionOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	upsAfterTheirUntap(t, g, opp.ID)
	if !upsIsTapped(t, g, land) || !upsIsTapped(t, g, bear) {
		t.Error("their creature or land untapped")
	}
	if upsIsTapped(t, g, rock) {
		t.Error("an artifact is neither: it untaps")
	}
}

// --- the Amonkhet Last cycle ---------------------------------------------

func TestBontusLastReckoningSweepsAndHoldsTheCastersLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	myLand := upsTapped(g, me.ID, "Swamp", "Basic Land — Swamp")
	theirLand := upsTapped(g, opp.ID, "Island", "Basic Land — Island")
	castCatalogSpell(t, g, "Bontu's Last Reckoning", "Sorcery", upsBontusOracle, nil)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, mine); ok {
		t.Error("the caster's creature survived")
	}
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the opponent's creature survived")
	}
	upsAfterTheirUntap(t, g, opp.ID)
	if upsIsTapped(t, g, theirLand) {
		t.Error("the opponent's land stayed tapped")
	}
	upsAfterTheirUntap(t, g, me.ID)
	if !upsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
}

func TestRhonassLastStandMakesASnakeAndHoldsTheCastersLands(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	myLand := upsTapped(g, me.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Rhonas's Last Stand", "Sorcery", upsRhonassOracle, nil)
	passPriorityAroundTable(t, g)
	snakes := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Snake" && c.Controller == me.ID && c.Power == 5 && c.Toughness == 4 {
			snakes++
		}
	}
	if snakes != 1 {
		t.Errorf("%d 5/4 Snakes, want 1", snakes)
	}
	upsAfterTheirUntap(t, g, me.ID)
	if !upsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
}

func TestOketrasLastMercyResetsLifeAndHoldsTheCastersLands(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	g.WithWriteLock(func() { me.Life = 7 })
	myLand := upsTapped(g, me.ID, "Plains", "Basic Land — Plains")
	castCatalogSpell(t, g, "Oketra's Last Mercy", "Sorcery", upsOketrasOracle, nil)
	passPriorityAroundTable(t, g)
	if want := startingLifeOf(g); me.Life != want {
		t.Errorf("life %d, want the starting %d", me.Life, want)
	}
	upsAfterTheirUntap(t, g, me.ID)
	if !upsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
}

func TestKefnetsLastWordStealsAndHoldsTheCastersLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	myLand := upsTapped(g, me.ID, "Island", "Basic Land — Island")
	castCatalogSpell(t, g, "Kefnet's Last Word", "Sorcery", upsKefnetsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, theirs); c.Controller != me.ID {
		t.Fatalf("the bear's controller is %s, want the caster", c.Controller)
	}
	upsAfterTheirUntap(t, g, me.ID)
	if !upsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
	if c := gaLayered(t, g, theirs); c.Controller != me.ID {
		t.Error("the control change has no duration: it lasts")
	}
}

// --- Icebreaker Kraken --------------------------------------------------

func upsSnowLands(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island",
			Owner: owner, Controller: owner,
		})
		out = append(out, id)
	}
	return out
}

func TestIcebreakerKrakenCostsLessPerSnowLandAndHoldsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	snow := upsSnowLands(g, me.ID, 10)
	plain := upsTapped(g, me.ID, "Island", "Basic Land — Island")
	theirBear := upsTapped(g, opp.ID, "Their Bear", "Creature — Bear")
	theirRock := upsTapped(g, opp.ID, "Their Rock", "Artifact")
	theirLand := upsTapped(g, opp.ID, "Their Island", "Basic Land — Island")

	// {10}{U}{U} less {1} for each of ten snow lands is {U}{U}.
	id := handCardFull(me, upsIcebreakerKrakenName, "Snow Creature — Kraken", "{10}{U}{U}", upsIcebreakerOracle, []string{"U"})
	me.ManaPool.EmptyPool()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {U}{U} with ten snow lands: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	upsAfterTheirUntap(t, g, opp.ID)
	if !upsIsTapped(t, g, theirBear) || !upsIsTapped(t, g, theirRock) {
		t.Error("their artifact or creature untapped")
	}
	if upsIsTapped(t, g, theirLand) {
		t.Error("their land stayed tapped")
	}

	// Return three snow lands: the Kraken goes back to hand.
	upsAfterTheirUntap(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{snow[0], snow[1], plain},
	}); err == nil {
		t.Fatal("a non-snow land paid the return cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		ReturnIDs: snow[:3],
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, l := range snow[:3] {
		if !me.Hand.Contains(l) {
			t.Error("a snow land did not go back to hand")
		}
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Error("the Kraken did not return to its owner's hand")
	}
}
