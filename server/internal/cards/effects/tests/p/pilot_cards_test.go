package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pilot_cards_test.go — ADR 0186 slice 9 (#3551): the two Pilot token
// templates and the four Pilot creatures. The counting rule itself
// (Card.CountedPower, the validator) is pinned in cost_power_test.go; what
// is here is that the printed cards and tokens declare it, and that the
// Captain's anthem and the Ace's untap do what they print.
//
// The crew-3 Vehicle and the Saddle-3 Mount are the test cards of
// cost_power_test.go: every Pilot here is short of 3 on plain power
// (except the Captain, which is checked at the edge too) and long of it
// when counted.

const (
	cloudspireCaptainOracle = "c86bac69-96d6-4b9f-87e4-2d7faecaa112"
	dynamiteDiverOracle     = "5f9ef763-22aa-4d43-b084-ef02465b489c"
	deathlessPilotOracle    = "31284a0b-9bd5-4186-bd27-6b16a3597e27"
	interfaceAceOracle      = "c4272676-a38a-4b23-988e-9cdb2feb46c1"
)

// Each Pilot creature counts for its printed amount when it crews and
// when it saddles, and its real power is untouched.
func TestPilotCreaturesCountForCrewAndSaddle(t *testing.T) {
	cases := []struct {
		name      string
		typeLine  string
		oracle    string
		power     int
		toughness int
		counted   int
	}{
		{"Cloudspire Captain", "Creature — Human Pilot", cloudspireCaptainOracle, 2, 3, 4},
		{"Dynamite Diver", "Creature — Goblin Pilot", dynamiteDiverOracle, 1, 1, 3},
		{"Deathless Pilot", "Creature — Zombie Pilot", deathlessPilotOracle, 2, 2, 4},
		{"Interface Ace", "Artifact Creature — Robot Pilot", interfaceAceOracle, 0, 4, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name+" crews", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.ToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			vehicle := effectstest.PushTestVehicle(g, me.ID, 4, 4)
			pilot := effectstest.PushPilotCard(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)

			if got := effectstest.CountedPowerOf(t, g, pilot, game.CostPowerCrew); got != tc.counted {
				t.Fatalf("CountedPower(crew) = %d, want %d", got, tc.counted)
			}
			if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, pilot); err != nil {
				t.Fatalf("%s could not crew a crew-3 Vehicle alone: %v", tc.name, err)
			}
			if got := effectstest.RealPowerOf(t, g, pilot); got != tc.power {
				t.Errorf("CurrentPower = %d, want %d (the rule must not change power)", got, tc.power)
			}
		})
		t.Run(tc.name+" saddles", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.ToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			mount := effectstest.PushMountForTest(g, me.ID, "Cost Power Mount", effectstest.CostPowerMountOracle, 2)
			pilot := effectstest.PushPilotCard(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)

			if got := effectstest.CountedPowerOf(t, g, pilot, game.CostPowerSaddle); got != tc.counted {
				t.Fatalf("CountedPower(saddle) = %d, want %d", got, tc.counted)
			}
			if err := effectstest.SaddleAbility(t, g, me.ID, mount, pilot); err != nil {
				t.Fatalf("%s could not saddle a Saddle 3 Mount alone: %v", tc.name, err)
			}
		})
	}
}

func findPilotTokenByName(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

// The saddle-and-crew Pilot counts 3 for both, and keeps the ability
// across a snapshot round trip as the token key alone.
func TestPilotTokenCountsThreeAndSurvivesARoundTrip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushTestVehicle(g, me.ID, 4, 4)
	mount := effectstest.PushMountForTest(g, me.ID, "Cost Power Mount", effectstest.CostPowerMountOracle, 2)
	effectstest.MakeTokens(t, g, me.ID, effects.PilotToken(), 1)
	tok, ok := findPilotTokenByName(g, "Pilot")
	if !ok {
		t.Fatal("no Pilot token on the battlefield")
	}

	if tok.TokenKey != game.TokenKey("pilot-saddles-and-crews") {
		t.Fatalf("TokenKey = %q", tok.TokenKey)
	}
	if got := game.TokenTextForCard(tok); got != "This token saddles Mounts and crews Vehicles as though its power were 2 greater." {
		t.Errorf("token text = %q", got)
	}
	for _, use := range []game.CostPowerUse{game.CostPowerCrew, game.CostPowerSaddle} {
		if got := effectstest.CountedPowerOf(t, g, tok.InstanceID, use); got != 3 {
			t.Errorf("CountedPower(%d) = %d, want 3", use, got)
		}
	}
	if got := effectstest.RealPowerOf(t, g, tok.InstanceID); got != 1 {
		t.Errorf("CurrentPower = %d, want 1", got)
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a table holding a Pilot token is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	rc := effectstest.FindRestoredCard(t, restored, tok.InstanceID)
	if rc.TokenKey != tok.TokenKey {
		t.Fatalf("restored token key = %q, want %q", rc.TokenKey, tok.TokenKey)
	}
	for _, use := range []game.CostPowerUse{game.CostPowerCrew, game.CostPowerSaddle} {
		if got := effectstest.CountedPowerOf(t, restored, tok.InstanceID, use); got != 3 {
			t.Errorf("restored CountedPower(%d) = %d, want 3", use, got)
		}
	}
	// And it still pays the real cost on the other side of the restore.
	if err := effectstest.CrewCostPowerVehicle(t, restored, me.ID, vehicle, tok.InstanceID); err != nil {
		t.Fatalf("the restored Pilot could not crew a crew-3 Vehicle: %v", err)
	}
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, tok.InstanceID); err != nil {
		t.Fatalf("the Pilot could not saddle a Saddle 3 Mount: %v", err)
	}
}

// The crew-only Pilot counts 3 for a crew cost and only 1 for a saddle.
func TestCrewPilotTokenCountsOneForSaddle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushTestVehicle(g, me.ID, 4, 4)
	mount := effectstest.PushMountForTest(g, me.ID, "Cost Power Mount", effectstest.CostPowerMountOracle, 2)
	effectstest.MakeTokens(t, g, me.ID, effects.CrewPilotToken(), 1)
	tok, ok := findPilotTokenByName(g, "Pilot")
	if !ok {
		t.Fatal("no Pilot token on the battlefield")
	}

	if tok.TokenKey != game.TokenKey("pilot-crews") {
		t.Fatalf("TokenKey = %q", tok.TokenKey)
	}
	if got := effectstest.CountedPowerOf(t, g, tok.InstanceID, game.CostPowerCrew); got != 3 {
		t.Errorf("CountedPower(crew) = %d, want 3", got)
	}
	if got := effectstest.CountedPowerOf(t, g, tok.InstanceID, game.CostPowerSaddle); got != 1 {
		t.Errorf("CountedPower(saddle) = %d, want 1", got)
	}
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, tok.InstanceID); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("a crew-only Pilot saddling a Saddle 3 Mount: err = %v, want ErrInsufficientCrew", err)
	}
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, tok.InstanceID); err != nil {
		t.Fatalf("a crew-only Pilot could not crew a crew-3 Vehicle: %v", err)
	}
}

// Interface Ace taps to crew on its controller's turn and untaps once;
// a second tap the same turn stays tapped.
func TestInterfaceAceUntapsOncePerTurnWhenCrewing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	first := effectstest.PushTestVehicle(g, me.ID, 4, 4)
	second := effectstest.PushTestVehicle(g, me.ID, 4, 4)
	ace := effectstest.PushPilotCard(g, me.ID, "Interface Ace", "Artifact Creature — Robot Pilot", interfaceAceOracle, 0, 4)

	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, first, ace); err != nil {
		t.Fatalf("the Ace could not crew on its controller's turn: %v", err)
	}
	if c, _ := effectstest.BattlefieldCardByID(g, ace); !c.Tapped {
		t.Fatal("the Ace did not tap to pay the crew cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, ace); c.Tapped {
		t.Fatal("the Ace did not untap after becoming tapped on its controller's turn")
	}
	if c, _ := effectstest.BattlefieldCardByID(g, first); !c.IsCreature() {
		t.Error("the first Vehicle was not crewed")
	}

	// A second tap this turn: crew another Vehicle with the Ace again.
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, second, ace); err != nil {
		t.Fatalf("the untapped Ace could not crew a second Vehicle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, ace); !c.Tapped {
		t.Error("the Ace untapped a second time in one turn")
	}
	if c, _ := effectstest.BattlefieldCardByID(g, second); !c.IsCreature() {
		t.Error("the second Vehicle was not crewed")
	}
}

// "During your turn": crewing on another player's turn leaves the Ace
// tapped.
func TestInterfaceAceDoesNotUntapOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := effectstest.PushTestVehicle(g, other.ID, 4, 4)
	ace := effectstest.PushPilotCard(g, other.ID, "Interface Ace", "Artifact Creature — Robot Pilot", interfaceAceOracle, 0, 4)

	if err := effectstest.CrewCostPowerVehicle(t, g, other.ID, vehicle, ace); err != nil {
		t.Fatalf("crew on another player's turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, ace); !c.Tapped {
		t.Error("the Ace untapped on another player's turn")
	}
}

// The anthem reaches Mounts and Vehicles its controller controls, crewed
// or not, and nothing else.
func TestCloudspireCaptainBuffsMountsAndVehiclesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := effectstest.PushTestVehicle(g, me.ID, 4, 4)
	mount := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cost Power Mount", OracleID: effectstest.CostPowerMountOracle,
		TypeLine: "Creature — Horse Mount", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushPilotCard(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	foeVehicle := effectstest.PushTestVehicle(g, foe.ID, 4, 4)
	foeMount := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Foe Mount", TypeLine: "Creature — Horse Mount",
		Power: 3, Toughness: 3, Owner: foe.ID, Controller: foe.ID,
	})
	captain := effectstest.PushPilotCard(g, me.ID, "Cloudspire Captain", "Creature — Human Pilot", cloudspireCaptainOracle, 2, 3)

	pt := func(id uuid.UUID) [2]int {
		return [2]int{effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)}
	}
	if got := pt(vehicle); got != [2]int{5, 5} {
		t.Errorf("an uncrewed Vehicle is %v, want 5/5", got)
	}
	if got := pt(mount); got != [2]int{4, 4} {
		t.Errorf("a Mount is %v, want 4/4", got)
	}
	if got := pt(bear); got != [2]int{2, 2} {
		t.Errorf("a Bear is %v, want 2/2", got)
	}
	if got := pt(captain); got != [2]int{2, 3} {
		t.Errorf("the Captain is %v, want 2/3", got)
	}
	if got := pt(foeVehicle); got != [2]int{4, 4} {
		t.Errorf("an opponent's Vehicle is %v, want 4/4", got)
	}
	if got := pt(foeMount); got != [2]int{3, 3} {
		t.Errorf("an opponent's Mount is %v, want 3/3", got)
	}

	// Crewed, the Vehicle is a creature and keeps the bonus. The Captain
	// crews it alone (counted 4 against crew 3).
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, captain); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, vehicle); !c.IsCreature() {
		t.Fatal("the Vehicle did not become a creature")
	}
	if got := pt(vehicle); got != [2]int{5, 5} {
		t.Errorf("a crewed Vehicle is %v, want 5/5", got)
	}
	// The Captain's own power is still 2: it crewed for 4.
	if got := effectstest.RealPowerOf(t, g, captain); got != 2 {
		t.Errorf("Captain power = %d, want 2", got)
	}
}

// Dynamite Diver deals 1 damage to any target when it dies.
func TestDynamiteDiverDiesForOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	diver := effectstest.PushPilotCard(g, me.ID, "Dynamite Diver", "Creature — Goblin Pilot", dynamiteDiverOracle, 1, 1)
	before := foe.Life

	effectstest.KillCreature(t, g, me.ID, diver)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("the dies trigger did not ask for a target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: foe.ID}); err != nil {
		t.Fatalf("pick target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != before-1 {
		t.Errorf("life %d -> %d, want -1", before, foe.Life)
	}
}

// Deathless Pilot returns from the graveyard to hand for {3}{B}.
func TestDeathlessPilotReturnsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Deathless Pilot", TypeLine: "Creature — Zombie Pilot",
		OracleID: deathlessPilotOracle, Owner: me.ID, Controller: me.ID})
	effectstest.RfCrAAddMana(t, g, me, "{C}{C}{C}{B}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(id) {
		t.Fatal("Deathless Pilot did not return to hand")
	}
}
