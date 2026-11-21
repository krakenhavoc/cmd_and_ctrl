package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: Lookup = %q, %v, completeness %v", name, spec.Name, ok, spec.Completeness)
		}
	}
}

// --- Blinding Beam ----------------------------------------------------

func TestBlindingBeamEntwinedTapsTwoAndHoldsThatPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	a := effectstest.PushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	land := effectstest.UpsTapped(g, opp.ID, "Island", "Basic Land — Island")
	if _, err := effectstest.CastEntwineSpell(g, me, "Blinding Beam", "Instant", "{2}{W}", upsBlindingBeamOracle, []string{"W"},
		"{C}{C}{C}{W}", []int{0, 1}, true,
		[]game.TargetRef{effectstest.EwaTarget(a, 0), effectstest.EwaTarget(b, 0), effectstest.EwaPlayer(opp.ID, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.UpsIsTapped(t, g, a) || !effectstest.UpsIsTapped(t, g, b) {
		t.Fatal("the tap bullet taps both creatures")
	}
	// CR 611.2c: a creature that arrives after the spell resolved is
	// held too.
	later := effectstest.UpsTapped(g, opp.ID, "Later Bear", "Creature — Bear")
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	for _, id := range []uuid.UUID{a, b, later} {
		if !effectstest.UpsIsTapped(t, g, id) {
			t.Errorf("%s untapped in the target player's next untap step", id)
		}
	}
	if effectstest.UpsIsTapped(t, g, land) {
		t.Error("a land is not a creature: it untaps")
	}
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if effectstest.UpsIsTapped(t, g, a) || effectstest.UpsIsTapped(t, g, later) {
		t.Error("the following untap step untaps them")
	}
}

func TestBlindingBeamUnentwinedTapsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	a := effectstest.PushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	if _, err := effectstest.CastEntwineSpell(g, me, "Blinding Beam", "Instant", "{2}{W}", upsBlindingBeamOracle, []string{"W"},
		"{C}{C}{W}", []int{0}, false, []game.TargetRef{effectstest.EwaTarget(a, 0), effectstest.EwaTarget(b, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.UpsIsTapped(t, g, a) || !effectstest.UpsIsTapped(t, g, b) {
		t.Fatal("the tap bullet taps both creatures")
	}
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if effectstest.UpsIsTapped(t, g, a) || effectstest.UpsIsTapped(t, g, b) {
		t.Error("without the second bullet they untap as usual")
	}
}

func TestBlindingBeamSecondBulletAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	theirs := effectstest.UpsTapped(g, opp.ID, "Their Bear", "Creature — Bear")
	mine := effectstest.UpsTapped(g, me.ID, "My Bear", "Creature — Bear")
	if _, err := effectstest.CastEntwineSpell(g, me, "Blinding Beam", "Instant", "{2}{W}", upsBlindingBeamOracle, []string{"W"},
		"{C}{C}{W}", []int{1}, false, []game.TargetRef{effectstest.EwaPlayer(opp.ID, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if !effectstest.UpsIsTapped(t, g, theirs) {
		t.Error("the target player's creature untapped")
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if effectstest.UpsIsTapped(t, g, mine) {
		t.Error("the caster's creature sits out nothing")
	}
}

// --- Mana Vapors, Misstep, Exhaustion -----------------------------------

func TestManaVaporsHoldsThatPlayersLandsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.CmcSeats(t, g)
	land := effectstest.UpsTapped(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.UpsTapped(g, opp.ID, "Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Mana Vapors", "Sorcery", upsManaVaporsOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	later := effectstest.UpsTapped(g, opp.ID, "Later Island", "Basic Land — Island")
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if !effectstest.UpsIsTapped(t, g, land) || !effectstest.UpsIsTapped(t, g, later) {
		t.Error("their lands, the later one included, untapped")
	}
	if effectstest.UpsIsTapped(t, g, bear) {
		t.Error("their creature stayed tapped")
	}
}

func TestMisstepHoldsThatPlayersCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	land := effectstest.UpsTapped(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.UpsTapped(g, opp.ID, "Bear", "Creature — Bear")
	mine := effectstest.UpsTapped(g, me.ID, "My Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Misstep", "Sorcery", upsMisstepOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if !effectstest.UpsIsTapped(t, g, bear) {
		t.Error("their creature untapped")
	}
	if effectstest.UpsIsTapped(t, g, land) {
		t.Error("their land stayed tapped")
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if effectstest.UpsIsTapped(t, g, mine) {
		t.Error("the caster's creature stayed tapped")
	}
}

func TestExhaustionHoldsCreaturesAndLandsOfAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	land := effectstest.UpsTapped(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.UpsTapped(g, opp.ID, "Bear", "Creature — Bear")
	rock := effectstest.UpsTapped(g, opp.ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Exhaustion", "Sorcery", upsExhaustionOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Exhaustion targets an opponent, not its caster")
	}
	effectstest.CastCatalogSpell(t, g, "Exhaustion", "Sorcery", upsExhaustionOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if !effectstest.UpsIsTapped(t, g, land) || !effectstest.UpsIsTapped(t, g, bear) {
		t.Error("their creature or land untapped")
	}
	if effectstest.UpsIsTapped(t, g, rock) {
		t.Error("an artifact is neither: it untaps")
	}
}

// --- the Amonkhet Last cycle ---------------------------------------------

func TestBontusLastReckoningSweepsAndHoldsTheCastersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	myLand := effectstest.UpsTapped(g, me.ID, "Swamp", "Basic Land — Swamp")
	theirLand := effectstest.UpsTapped(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Bontu's Last Reckoning", "Sorcery", upsBontusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mine); ok {
		t.Error("the caster's creature survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the opponent's creature survived")
	}
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if effectstest.UpsIsTapped(t, g, theirLand) {
		t.Error("the opponent's land stayed tapped")
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if !effectstest.UpsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
}

func TestRhonassLastStandMakesASnakeAndHoldsTheCastersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	myLand := effectstest.UpsTapped(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Rhonas's Last Stand", "Sorcery", upsRhonassOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	snakes := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Snake" && c.Controller == me.ID && c.Power == 5 && c.Toughness == 4 {
			snakes++
		}
	}
	if snakes != 1 {
		t.Errorf("%d 5/4 Snakes, want 1", snakes)
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if !effectstest.UpsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
}

func TestOketrasLastMercyResetsLifeAndHoldsTheCastersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	g.WithWriteLock(func() { me.Life = 7 })
	myLand := effectstest.UpsTapped(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.CastCatalogSpell(t, g, "Oketra's Last Mercy", "Sorcery", upsOketrasOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if want := effects.StartingLifeOf(g); me.Life != want {
		t.Errorf("life %d, want the starting %d", me.Life, want)
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if !effectstest.UpsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
}

func TestKefnetsLastWordStealsAndHoldsTheCastersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	myLand := effectstest.UpsTapped(g, me.ID, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Kefnet's Last Word", "Sorcery", upsKefnetsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, theirs); c.Controller != me.ID {
		t.Fatalf("the bear's controller is %s, want the caster", c.Controller)
	}
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
	if !effectstest.UpsIsTapped(t, g, myLand) {
		t.Error("the caster's land untapped in their next untap step")
	}
	if c := effectstest.GaLayered(t, g, theirs); c.Controller != me.ID {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	snow := upsSnowLands(g, me.ID, 10)
	plain := effectstest.UpsTapped(g, me.ID, "Island", "Basic Land — Island")
	theirBear := effectstest.UpsTapped(g, opp.ID, "Their Bear", "Creature — Bear")
	theirRock := effectstest.UpsTapped(g, opp.ID, "Their Rock", "Artifact")
	theirLand := effectstest.UpsTapped(g, opp.ID, "Their Island", "Basic Land — Island")

	// {10}{U}{U} less {1} for each of ten snow lands is {U}{U}.
	id := effectstest.HandCardFull(me, upsIcebreakerKrakenName, "Snow Creature — Kraken", "{10}{U}{U}", upsIcebreakerOracle, []string{"U"})
	me.ManaPool.EmptyPool()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {U}{U} with ten snow lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.UpsAfterTheirUntap(t, g, opp.ID)
	if !effectstest.UpsIsTapped(t, g, theirBear) || !effectstest.UpsIsTapped(t, g, theirRock) {
		t.Error("their artifact or creature untapped")
	}
	if effectstest.UpsIsTapped(t, g, theirLand) {
		t.Error("their land stayed tapped")
	}

	// Return three snow lands: the Kraken goes back to hand.
	effectstest.UpsAfterTheirUntap(t, g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Error("the Kraken did not return to its owner's hand")
	}
}
