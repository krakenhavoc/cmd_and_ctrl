package protocol

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Secret of Bloodbending's optional waterbend
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_control_view_test.go — #3999: what the view, the log and the
// controller's legal moves carry for a control of one combat phase
// (Secret of Bloodbending), and the waterbend offer's picker.

const secretOfBloodbendingOracle = "743fa824-18cb-4225-ba7b-5b0bbc3c8d41"

func combatControlTable(t *testing.T) (g *game.Game, alice, bob, carol uuid.UUID) {
	t.Helper()
	g = game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Turn.ActiveSeat
	alice, bob, carol = g.Seats[a].ID, g.Seats[(a+1)%3].ID, g.Seats[(a+2)%3].ID
	g.WithWriteLock(func() {
		g.BeginCombatControlForEffect(game.CombatControl{
			Controller: alice, Controlled: bob, Source: uuid.New(), SourceName: "Secret of Bloodbending",
		})
	})
	return g, alice, bob, carol
}

func stepToSeatStep(t *testing.T, g *game.Game, seat uuid.UUID, step game.Step) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if g.Turn.Step == step && g.Seats[g.Turn.ActiveSeat].ID == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached %s", step)
}

// Scheduled, it is public and says it waits for a combat phase; live,
// it is listed with until "combat", and Alice is offered Bob's moves.
func TestCombatControlViewAndTheControllersMoves(t *testing.T) {
	g, alice, bob, carol := combatControlTable(t)
	for _, viewer := range []string{alice.String(), bob.String(), carol.String(), SpectatorViewerID} {
		v := ViewOfGameFor(g, viewer)
		if len(v.ScheduledPlayerControls) != 1 || v.ScheduledPlayerControls[0].Until != "combat" {
			t.Fatalf("viewer %s: scheduled_player_controls = %+v, want one waiting for a combat phase", viewer, v.ScheduledPlayerControls)
		}
	}
	stepToSeatStep(t, g, bob, game.StepPrecombatMain)
	if v := ViewOfGameFor(g, alice.String()); len(v.PlayerControls) != 0 || len(v.ScheduledPlayerControls) != 1 {
		t.Fatalf("Bob's main phase: player_controls %+v, scheduled %+v", v.PlayerControls, v.ScheduledPlayerControls)
	}

	stepToSeatStep(t, g, bob, game.StepBeginCombat)
	full := ViewOfGame(g)
	av := FilterViewFor(full, alice.String())
	if len(av.PlayerControls) != 1 || av.PlayerControls[0].Until != "combat" || av.PlayerControls[0].Controlled != bob.String() {
		t.Fatalf("player_controls = %+v, want Alice controlling Bob for the combat", av.PlayerControls)
	}
	if len(av.ScheduledPlayerControls) != 0 {
		t.Errorf("a begun control is still scheduled: %+v", av.ScheduledPlayerControls)
	}
	// CR 723.4: Alice sees Bob's hand while the control is live.
	for _, c := range seatViewOf(t, av, bob).Hand.Cards {
		if !c.KnownByYou {
			t.Errorf("Alice does not know %s in Bob's hand", c.Name)
		}
	}
	if d := av.ControlledLegalActions[bob.String()]; d == nil || !d.Pass {
		t.Errorf("controlled_legal_actions[bob] = %+v, want a pass", d)
	}
	for _, id := range []uuid.UUID{bob, carol} {
		if v := FilterViewFor(full, id.String()); len(v.LegalMoves) != 0 || v.ControlledLegalActions != nil {
			t.Errorf("%v: legal_moves %d, controlled_legal_actions %v", id, len(v.LegalMoves), v.ControlledLegalActions)
		}
	}

	stepToSeatStep(t, g, bob, game.StepPostcombatMain)
	av = ViewOfGameFor(g, alice.String())
	if len(av.PlayerControls) != 0 {
		t.Errorf("player_controls after the combat phase = %+v", av.PlayerControls)
	}
	for _, c := range seatViewOf(t, av, bob).Hand.Cards {
		if c.KnownByYou {
			t.Fatalf("Alice still knows %s in Bob's hand after the combat phase", c.Name)
		}
	}
}

// The three log lines say "combat phase" and "this combat".
func TestCombatControlLogLines(t *testing.T) {
	g, _, bob, _ := combatControlTable(t)
	stepToSeatStep(t, g, bob, game.StepPostcombatMain)
	want := map[LogKind]string{
		LogPlayerControlScheduled: "will control P",
		LogPlayerControlBegan:     "this combat",
		LogPlayerControlEnded:     "no longer controls",
	}
	for kind, frag := range want {
		lines := logLinesOfKind(g, kind)
		if len(lines) != 1 || lines[0].Label != "combat" {
			t.Fatalf("%s lines = %+v, want one labelled combat", kind, lines)
		}
		if !strings.Contains(lines[0].Text, frag) {
			t.Errorf("%s text = %q, want it to contain %q", kind, lines[0].Text, frag)
		}
	}
	if s := logLinesOfKind(g, LogPlayerControlScheduled)[0].Text; !strings.Contains(s, "next combat phase") {
		t.Errorf("scheduled text = %q, want \"next combat phase\"", s)
	}
}

// The waterbend offer ships its picker: the caster's untapped artifacts
// and creatures, at most ten of them.
func TestOptionalWaterbendOfferShipsItsTapPicker(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	bear := seatCreature(g, me, "Plain Bear", 2, false, false)
	secret := game.NewCard("Secret of Bloodbending", me)
	secret.OracleID = secretOfBloodbendingOracle
	secret.ManaCost = "{U}{U}{U}{U}"
	secret.TypeLine = "Sorcery — Lesson"
	secret.KnownBy = map[uuid.UUID]bool{me: true}
	g.Seats[0].Hand.PushTop(secret)

	v := ViewOfGameFor(g, me.String())
	hand := handCardView(t, v, 0, secret.InstanceID)
	if hand.TapCost != nil {
		t.Errorf("the card's own tap_cost = %+v, want none (the waterbend is optional)", hand.TapCost)
	}
	if len(hand.OptionalCosts) != 1 {
		t.Fatalf("optional_costs = %+v, want the waterbend", hand.OptionalCosts)
	}
	oc := hand.OptionalCosts[0]
	if oc.Key != "waterbend" || oc.ManaCost != "{10}" || oc.TapCost == nil {
		t.Fatalf("optional cost = %+v, want waterbend {10} with a picker", oc)
	}
	if oc.TapCost.Max != 10 || oc.TapCost.Key != "waterbend" || !containsID(oc.TapCost.Options, bear) {
		t.Errorf("tap_cost = %+v, want ten from the caster's untapped creatures", oc.TapCost)
	}
}

// #4000: The Dominion Bracelet's granted row on the equipped creature
// says its cost exiles the grantor, and names the grantor.
func TestExileGrantorRowIsMarked(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	known := map[uuid.UUID]bool{me: true}
	bearCard := game.NewCard("Plain Bear", me)
	bearCard.TypeLine = "Creature — Bear"
	bearCard.Power, bearCard.Toughness = 3, 3
	bearCard.KnownBy = known
	bear := bearCard.InstanceID
	g.Battlefield.PushTop(bearCard)
	bracelet := game.NewCard("The Dominion Bracelet", me)
	bracelet.OracleID = "c017f8ae-9103-4965-a07a-f259f44a157d"
	bracelet.TypeLine = "Legendary Artifact — Equipment"
	bracelet.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
	bracelet.KnownBy = known
	g.Battlefield.PushTop(bracelet)
	// The entry event stamps the Bracelet's timestamp, which its layer-6
	// grant is ordered by.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: bracelet.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.BumpLayerVersionForTest()

	v := ViewOfGameFor(g, me.String())
	c := battlefieldCardView(t, v, bear)
	var row *ActivatedAbilityView
	for i := range c.ActivatedAbilities {
		if c.ActivatedAbilities[i].ExileGrantor {
			row = &c.ActivatedAbilities[i]
		}
	}
	if row == nil {
		t.Fatalf("no exile_grantor row on the equipped creature: %+v", c.ActivatedAbilities)
	}
	if row.GrantedBy == nil || row.GrantedBy.ID != bracelet.InstanceID.String() || row.ExileSelf {
		t.Errorf("row = %+v, want granted by the Bracelet and not exile_self", row)
	}
}
