package testsms

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsn "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/n"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ninja_teen_test.go — #4680: a granted sneak for the creature cards in
// your graveyard (game/sneak.go), and its card, Ninja Teen. Sneak itself
// is sneak_test.go.

// pushNinjaTeen seeds Ninja Teen under p at `level` (CR 716.2).
func pushNinjaTeen(g *game.Game, p *game.Player, level int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ninja Teen", TypeLine: "Enchantment — Class", ManaCost: "{2}{B}",
		OracleID: cardsn.NinjaTeenOracle, Owner: p.ID, Controller: p.ID, ClassLevel: level,
	})
}

// ninjaTeenBoard is sneakBoard with the card in my graveyard: Ninja Teen
// at `level`, my 1/1 attacking seat 1 unblocked, a Grave Bear in my
// graveyard and {B}{B}{B}{B} floating, in my declare blockers step.
func ninjaTeenBoard(t *testing.T, g *game.Game, level int) (bear, attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	pushNinjaTeen(g, me, level)
	attacker = effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	bear = effectstest.PushToGraveyard(g, me, effectstest.GraveBear())
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.FinishBlockDeclarations(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("left the declare blockers step (at %s)", g.Turn.Step)
	}
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}")
	return bear, attacker, me, opp
}

func grantedSneakCast(g *game.Game, me *game.Player, card, returned uuid.UUID) error {
	return g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedSneak, AltCostIDs: []uuid.UUID{returned}})
}

func TestNinjaTeenIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(cardsn.NinjaTeenOracle)
	if !ok {
		t.Fatal("Ninja Teen is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	if len(spec.GatedCastPermissions) != 1 {
		t.Fatalf("gated permissions %+v, want the level-3 sneak", spec.GatedCastPermissions)
	}
	gp := spec.GatedCastPermissions[0]
	if gp.ActiveWhen != effects.Level(3) || gp.Permission.AltCostKey != game.AltCostKeyGrantedSneak ||
		gp.Permission.SneakCost != "{3}{B}" || gp.Permission.Zone != game.ZoneGraveyard {
		t.Errorf("level-3 permission %+v gated %+v, want granted sneak {3}{B} over the graveyard at level 3", gp.Permission, gp.ActiveWhen)
	}
}

// The headline: at level 3 a creature card in my graveyard is cast for
// {3}{B} and an unblocked attacker in my declare blockers step. It leaves
// the graveyard as it is cast and enters tapped and attacking the player
// the returned creature attacked (CR 702.190b). The returned creature
// leaving the battlefield is level 1's trigger.
func TestNinjaTeenSneaksACreatureCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear, attacker, me, opp := ninjaTeenBoard(t, g, 3)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	if err := grantedSneakCast(g, me, bear, attacker); err != nil {
		t.Fatalf("granted sneak cast: %v", err)
	}
	if me.Graveyard.Contains(bear) {
		t.Fatal("the card is still in the graveyard after it was cast")
	}
	if !me.Hand.Contains(attacker) {
		t.Fatal("the unblocked attacker was not returned to hand as the cost")
	}
	if item := g.StackMeta[bear]; item == nil || item.AltCost != game.AltCostKeyGrantedSneak {
		t.Fatalf("stack item %+v, want the granted sneak claim recorded", item)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left in the pool, want 0: {3}{B} was the price", n)
	}
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)

	c, ok := g.LookupCardForEffect(bear)
	if !ok || !g.Battlefield.Contains(bear) {
		t.Fatal("the bear is not on the battlefield")
	}
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("bear tapped=%v attacking %s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, opp.ID)
	}
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d: the returned creature left the battlefield", p.Name, p.Life, want)
		}
	}
}

// The window is my own declare blockers step (CR 702.190a), and the card's
// mana cost is never a way out of the graveyard, even in a main phase.
func TestNinjaTeenGrantedSneakOnlyInYourDeclareBlockersStep(t *testing.T) {
	t.Run("main phase, and never for its mana cost", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		pushNinjaTeen(g, me, 3)
		rat := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
		bear := effectstest.PushToGraveyard(g, me, effectstest.GraveBear())
		effectstest.FloatMana(t, g, me, "{G}{G}{B}{B}{B}{B}")
		if err := grantedSneakCast(g, me, bear, rat); err == nil {
			t.Error("granted sneak cast in the main phase was accepted")
		}
		if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
			t.Error("the creature card was cast from the graveyard for its mana cost")
		}
		if keys := effectstest.GraveyardOfferKeys(g, me, bear); slices.Contains(keys, "") {
			t.Errorf("offers %q include the mana cost", keys)
		}
		if !me.Graveyard.Contains(bear) || !g.Battlefield.Contains(rat) {
			t.Error("a refused cast moved something")
		}
	})

	t.Run("combat damage step", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		bear, attacker, me, _ := ninjaTeenBoard(t, g, 3)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		if err := grantedSneakCast(g, me, bear, attacker); !errors.Is(err, game.ErrSorcerySpeedRequired) {
			t.Errorf("granted sneak in the damage step: %v, want ErrSorcerySpeedRequired", err)
		}
	})

	t.Run("an opponent's declare blockers step", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		_, _, _, opp := ninjaTeenBoard(t, g, 3)
		pushNinjaTeen(g, opp, 3)
		theirs := effectstest.PushToGraveyard(g, opp, effectstest.GraveBear())
		var offer *game.AlternativeCost
		var open bool
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(theirs)
			grant := g.CastPermissionForClaimLocked(opp.ID, c, game.ZoneGraveyard, game.AltCostKeyGrantedSneak)
			offer = grant.AlternativeCostFor(c)
			open = g.CastTimingForOfferOpenLocked(opp.ID, c, game.ZoneGraveyard, grant, offer)
		})
		if offer == nil || !offer.Sneak {
			t.Fatalf("the defending player's Ninja Teen offers %+v, want its granted sneak", offer)
		}
		if open {
			t.Error("the granted sneak window is open for the defending player")
		}
	})
}

// The cost is an unblocked attacker I control: a blocked one is refused,
// and the bot's move list offers only the unblocked one, as a move the
// engine accepts.
func TestNinjaTeenGrantedSneakNeedsAnUnblockedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushNinjaTeen(g, me, 3)
	blocked := effectstest.B12Creature(g, me.ID, "Blocked Rat", "Creature — Rat", 1, 1)
	free := effectstest.B12Creature(g, me.ID, "Free Rat", "Creature — Rat", 1, 1)
	// Level 2 gave my rats menace, so it takes two walls to block one.
	walls := []uuid.UUID{
		effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4),
		effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4),
	}
	bear := effectstest.PushToGraveyard(g, me, effectstest.GraveBear())
	effectstest.DeclareAttack(t, g, opp.ID, blocked, free)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: walls[0], Attacker: blocked}, {Blocker: walls[1], Attacker: blocked}}); err != nil {
		t.Fatalf("DeclareBlockers: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}")

	var returned []uuid.UUID
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != bear {
			continue
		}
		var p struct {
			AlternativeCost string   `json:"alternative_cost"`
			AltCostIDs      []string `json:"alt_cost_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if p.AlternativeCost != game.AltCostKeyGrantedSneak {
			t.Errorf("enumerator offers %q for the graveyard bear, want only the granted sneak", p.AlternativeCost)
			continue
		}
		for _, s := range p.AltCostIDs {
			returned = append(returned, uuid.MustParse(s))
		}
	}
	if len(returned) != 1 || returned[0] != free {
		t.Errorf("enumerator returns %v, want only the unblocked %s", returned, free)
	}
	if err := grantedSneakCast(g, me, bear, blocked); err == nil {
		t.Fatal("granted sneak paid with a blocked attacker")
	}
	if !g.Battlefield.Contains(blocked) || !me.Graveyard.Contains(bear) {
		t.Fatal("the refused cast moved something")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedSneak, AltCostIDs: []uuid.UUID{free}, AutoTap: true, Strict: true,
	}); err != nil {
		t.Fatalf("the enumerated move is refused: %v", err)
	}

	// No attacker at all: nothing to return, nothing offered (the offer
	// list is not filtered by timing, so a main phase asks the same).
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	pushNinjaTeen(g2, me2, 3)
	effectstest.B12Creature(g2, me2.ID, "Idle Rat", "Creature — Rat", 1, 1)
	bear2 := effectstest.PushToGraveyard(g2, me2, effectstest.GraveBear())
	effectstest.FloatMana(t, g2, me2, "{B}{B}{B}{B}")
	if keys := effectstest.GraveyardOfferKeys(g2, me2, bear2); len(keys) != 0 {
		t.Errorf("offers %q with no attacker to return, want none", keys)
	}
}

// A creature card that prints its own sneak may use either (CR 113.2c):
// Elektra's {1}{B}{B} or the granted {3}{B}. Without Ninja Teen she is
// not castable from the graveyard at all.
func TestNinjaTeenACardWithSneakMayUseEither(t *testing.T) {
	elektraCard := game.Card{
		Name: effectstest.ElektraDaughterName, TypeLine: effectstest.ElektraDaughterType, ManaCost: "{2}{B}{B}",
		OracleID: effectstest.ElektraDaughterOracle, Power: 3, Toughness: 3,
	}
	for _, claim := range []string{"sneak", game.AltCostKeyGrantedSneak} {
		t.Run(claim, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, attacker, me, opp := ninjaTeenBoard(t, g, 3)
			elektra := effectstest.PushToGraveyard(g, me, elektraCard)
			keys := effectstest.GraveyardOfferKeys(g, me, elektra)
			if !slices.Contains(keys, "sneak") || !slices.Contains(keys, game.AltCostKeyGrantedSneak) || slices.Contains(keys, "") {
				t.Fatalf("offers %q, want her own sneak and the granted one, and not her mana cost", keys)
			}
			if err := g.CastSpell(me.ID, elektra, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: claim, AltCostIDs: []uuid.UUID{attacker}}); err != nil {
				t.Fatalf("%s cast from the graveyard: %v", claim, err)
			}
			want := 0 // {B}{B}{B}{B} less {3}{B}
			if claim == "sneak" {
				want = 1 // less {1}{B}{B}
			}
			if n := me.ManaPool.Total(); n != want {
				t.Errorf("%d mana left, want %d", n, want)
			}
			effectstest.AssistCardSettle(t, g, me.ID, nil, true)
			c, ok := g.LookupCardForEffect(elektra)
			if !ok || !g.Battlefield.Contains(elektra) || !c.Tapped || c.AttackingTarget != opp.ID {
				t.Errorf("Elektra on battlefield=%v tapped=%v attacking %s, want tapped and attacking %s",
					g.Battlefield.Contains(elektra), c.Tapped, c.AttackingTarget, opp.ID)
			}
		})
	}

	t.Run("without Ninja Teen", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
		elektra := effectstest.PushToGraveyard(g, me, elektraCard)
		effectstest.DeclareAttack(t, g, opp.ID, attacker)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.FinishBlockDeclarations(t, g)
		effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}")
		if err := g.CastSpell(me.ID, elektra, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "sneak", AltCostIDs: []uuid.UUID{attacker}}); err == nil {
			t.Error("Elektra sneaked out of a graveyard with no Ninja Teen")
		}
	})
}

// "Using their sneak abilities": a creature card's other alternative cost
// is not a way out of the graveyard, nor is any noncreature card.
func TestNinjaTeenOpensTheGraveyardOnlyToSneak(t *testing.T) {
	const otherAltOracle = "test-other-alt-cost-4680"
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: otherAltOracle, Name: "Cheap Brute", Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{Key: "test-cheap", Label: "Pay {B}", ManaCost: "{B}"}},
	})
	g := effectstest.NewCatalogGame(t)
	_, attacker, me, _ := ninjaTeenBoard(t, g, 3)
	brute := effectstest.PushToGraveyard(g, me, game.Card{
		Name: "Cheap Brute", TypeLine: "Creature — Ogre", ManaCost: "{5}{B}", OracleID: otherAltOracle, Power: 5, Toughness: 5,
	})
	if err := g.CastSpell(me.ID, brute, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "test-cheap"}); err == nil {
		t.Error("a creature card's non-sneak alternative cost cast it from the graveyard")
	}
	if keys := effectstest.GraveyardOfferKeys(g, me, brute); !slices.Equal(keys, []string{game.AltCostKeyGrantedSneak}) {
		t.Errorf("offers %q, want only the granted sneak", keys)
	}
	bolt := effectstest.PushToGraveyard(g, me, game.Card{Name: "Grave Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	if err := grantedSneakCast(g, me, bolt, attacker); err == nil {
		t.Error("an instant card was sneaked out of the graveyard")
	}
	if keys := effectstest.GraveyardOfferKeys(g, me, bolt); len(keys) != 0 {
		t.Errorf("instant offers %q, want none", keys)
	}
}

// Level 2 or lower grants nothing (CR 716.2a); level 2's anthem and menace
// are on from level 2 and stay at level 3.
func TestNinjaTeenLevelsGateTheirLines(t *testing.T) {
	for _, level := range []int{1, 2, 3} {
		g := effectstest.NewCatalogGame(t)
		bear, attacker, me, _ := ninjaTeenBoard(t, g, level)
		keys := effectstest.GraveyardOfferKeys(g, me, bear)
		if level < 3 {
			if len(keys) != 0 {
				t.Errorf("level %d: offers %q, want none", level, keys)
			}
			if err := grantedSneakCast(g, me, bear, attacker); err == nil {
				t.Errorf("level %d: granted sneak accepted", level)
			}
			if !me.Graveyard.Contains(bear) {
				t.Errorf("level %d: the bear left the graveyard", level)
			}
		} else if !slices.Equal(keys, []string{game.AltCostKeyGrantedSneak}) {
			t.Errorf("level 3: offers %q, want the granted sneak", keys)
		}

		rat := effectstest.LayeredCard(t, g, attacker)
		wantPower, wantMenace := 1, false
		if level >= 2 {
			wantPower, wantMenace = 2, true
		}
		if got := rat.Effective().Power; got != wantPower {
			t.Errorf("level %d: rat power %d, want %d", level, got, wantPower)
		}
		if got := game.HasKeyword(&rat, "menace"); got != wantMenace {
			t.Errorf("level %d: rat menace %v, want %v", level, got, wantMenace)
		}
	}
}

// Level 1: any creature I control leaving, a token or a card, by any route,
// drains each opponent; an opponent's creature does not.
func TestNinjaTeenCreatureLeavingDrainsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pushNinjaTeen(g, me, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, theirs); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, mine); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.Name, p.Life, want)
		}
	}
}

// The claim rides the stack item into an undo clone and a snapshot round
// trip, and the restored spell still enters tapped and attacking.
func TestNinjaTeenGrantedSneakSurvivesUndoAndRestore(t *testing.T) {
	for _, tc := range []struct {
		name    string
		reshape func(t *testing.T, g *game.Game) *game.Game
	}{
		{"clone", func(_ *testing.T, g *game.Game) *game.Game { return g.Clone() }},
		{"snapshot", func(t *testing.T, g *game.Game) *game.Game { return effectstest.RestoreRoundTrip(t, g, true) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			bear, attacker, me, opp := ninjaTeenBoard(t, g, 3)
			if err := grantedSneakCast(g, me, bear, attacker); err != nil {
				t.Fatalf("granted sneak cast: %v", err)
			}
			restored := tc.reshape(t, g)
			effectstest.AssistCardSettle(t, restored, me.ID, nil, true)
			c, ok := restored.LookupCardForEffect(bear)
			if !ok || !c.Tapped || c.AttackingTarget != opp.ID {
				t.Fatalf("restored: tapped=%v attacking %s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, opp.ID)
			}
		})
	}
}

// The client's picker: the graveyard card is castable here, with the
// granted sneak on the wire and the unblocked attacker as its payment.
func TestNinjaTeenGrantedSneakOnTheWire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear, attacker, me, _ := ninjaTeenBoard(t, g, 3)
	v := protocol.ViewOfGameFor(g, me.ID.String())
	var card *protocol.CardView
	for _, s := range v.Seats {
		for i := range s.Graveyard.Cards {
			if s.Graveyard.Cards[i].InstanceID == bear.String() {
				card = &s.Graveyard.Cards[i]
			}
		}
	}
	if card == nil {
		t.Fatal("the bear is not in the graveyard view")
	}
	if !card.CastableHere {
		t.Error("the bear is not castable_here in the declare blockers step")
	}
	var offer *protocol.AlternativeCostView
	for i := range card.AlternativeCosts {
		if card.AlternativeCosts[i].Key == game.AltCostKeyGrantedSneak {
			offer = &card.AlternativeCosts[i]
		}
	}
	if offer == nil {
		t.Fatalf("no granted sneak on the wire: %+v", card.AlternativeCosts)
	}
	if offer.TimingClosed || offer.PayOptions == nil || len(offer.PayOptions.Cards) != 1 || offer.PayOptions.Cards[0] != attacker.String() {
		t.Errorf("offer %+v, want open and paid with %s", offer, attacker)
	}
}
