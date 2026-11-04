package effects

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ninja_teen_test.go — #4680: a granted sneak for the creature cards in
// your graveyard (game/sneak.go), and its card, Ninja Teen. Sneak itself
// is sneak_test.go.

// pushNinjaTeen seeds Ninja Teen under p at `level` (CR 716.2).
func pushNinjaTeen(g *game.Game, p *game.Player, level int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ninja Teen", TypeLine: "Enchantment — Class", ManaCost: "{2}{B}",
		OracleID: ninjaTeenOracle, Owner: p.ID, Controller: p.ID, ClassLevel: level,
	})
}

// pushToGraveyard puts a card into p's graveyard.
func pushToGraveyard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	// A graveyard is public (CR 400.2): every seat knows the card.
	c.KnownBy = map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		c.KnownBy[s.ID] = true
	}
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// graveBear is a vanilla creature card with no way out of a graveyard of
// its own.
func graveBear() game.Card {
	return game.Card{Name: "Grave Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
}

// ninjaTeenBoard is sneakBoard with the card in my graveyard: Ninja Teen
// at `level`, my 1/1 attacking seat 1 unblocked, a Grave Bear in my
// graveyard and {B}{B}{B}{B} floating, in my declare blockers step.
func ninjaTeenBoard(t *testing.T, g *game.Game, level int) (bear, attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	pushNinjaTeen(g, me, level)
	attacker = b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	bear = pushToGraveyard(g, me, graveBear())
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("left the declare blockers step (at %s)", g.Turn.Step)
	}
	floatMana(t, g, me, "{B}{B}{B}{B}")
	return bear, attacker, me, opp
}

func grantedSneakCast(g *game.Game, me *game.Player, card, returned uuid.UUID) error {
	return g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedSneak, AltCostIDs: []uuid.UUID{returned}})
}

// graveyardOfferKeys lists the claims CastOffersForLocked offers for a
// card in p's graveyard; "" is its mana cost.
func graveyardOfferKeys(g *game.Game, p *game.Player, card uuid.UUID) []string {
	var keys []string
	g.ReadSnapshot(func() {
		var c game.Card
		for _, gc := range p.Graveyard.Cards {
			if gc.InstanceID == card {
				c = gc
			}
		}
		grant := g.CastPermissionForLocked(p.ID, c, game.ZoneGraveyard)
		for _, ac := range g.CastOffersForLocked(p.ID, c, game.ZoneGraveyard, grant) {
			if ac == nil {
				keys = append(keys, "")
				continue
			}
			keys = append(keys, ac.Key)
		}
	})
	return keys
}

func TestNinjaTeenIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(ninjaTeenOracle)
	if !ok {
		t.Fatal("Ninja Teen is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	if len(spec.GatedCastPermissions) != 1 {
		t.Fatalf("gated permissions %+v, want the level-3 sneak", spec.GatedCastPermissions)
	}
	gp := spec.GatedCastPermissions[0]
	if gp.ActiveWhen != Level(3) || gp.Permission.AltCostKey != game.AltCostKeyGrantedSneak ||
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
	g := newCatalogGame(t)
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
	assistCardSettle(t, g, me.ID, nil, true)

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
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		pushNinjaTeen(g, me, 3)
		rat := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
		bear := pushToGraveyard(g, me, graveBear())
		floatMana(t, g, me, "{G}{G}{B}{B}{B}{B}")
		if err := grantedSneakCast(g, me, bear, rat); err == nil {
			t.Error("granted sneak cast in the main phase was accepted")
		}
		if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
			t.Error("the creature card was cast from the graveyard for its mana cost")
		}
		if keys := graveyardOfferKeys(g, me, bear); slices.Contains(keys, "") {
			t.Errorf("offers %q include the mana cost", keys)
		}
		if !me.Graveyard.Contains(bear) || !g.Battlefield.Contains(rat) {
			t.Error("a refused cast moved something")
		}
	})

	t.Run("combat damage step", func(t *testing.T) {
		g := newCatalogGame(t)
		bear, attacker, me, _ := ninjaTeenBoard(t, g, 3)
		advanceTo(t, g, game.StepCombatDamage)
		if err := grantedSneakCast(g, me, bear, attacker); !errors.Is(err, game.ErrSorcerySpeedRequired) {
			t.Errorf("granted sneak in the damage step: %v, want ErrSorcerySpeedRequired", err)
		}
	})

	t.Run("an opponent's declare blockers step", func(t *testing.T) {
		g := newCatalogGame(t)
		_, _, _, opp := ninjaTeenBoard(t, g, 3)
		pushNinjaTeen(g, opp, 3)
		theirs := pushToGraveyard(g, opp, graveBear())
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushNinjaTeen(g, me, 3)
	blocked := b12Creature(g, me.ID, "Blocked Rat", "Creature — Rat", 1, 1)
	free := b12Creature(g, me.ID, "Free Rat", "Creature — Rat", 1, 1)
	// Level 2 gave my rats menace, so it takes two walls to block one.
	walls := []uuid.UUID{
		b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4),
		b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4),
	}
	bear := pushToGraveyard(g, me, graveBear())
	declareAttack(t, g, opp.ID, blocked, free)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: walls[0], Attacker: blocked}, {Blocker: walls[1], Attacker: blocked}}); err != nil {
		t.Fatalf("DeclareBlockers: %v", err)
	}
	finishBlockDeclarations(t, g)
	floatMana(t, g, me, "{B}{B}{B}{B}")

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
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	pushNinjaTeen(g2, me2, 3)
	b12Creature(g2, me2.ID, "Idle Rat", "Creature — Rat", 1, 1)
	bear2 := pushToGraveyard(g2, me2, graveBear())
	floatMana(t, g2, me2, "{B}{B}{B}{B}")
	if keys := graveyardOfferKeys(g2, me2, bear2); len(keys) != 0 {
		t.Errorf("offers %q with no attacker to return, want none", keys)
	}
}

// A creature card that prints its own sneak may use either (CR 113.2c):
// Elektra's {1}{B}{B} or the granted {3}{B}. Without Ninja Teen she is
// not castable from the graveyard at all.
func TestNinjaTeenACardWithSneakMayUseEither(t *testing.T) {
	elektraCard := game.Card{
		Name: elektraDaughterName, TypeLine: elektraDaughterType, ManaCost: "{2}{B}{B}",
		OracleID: elektraDaughterOracle, Power: 3, Toughness: 3,
	}
	for _, claim := range []string{"sneak", game.AltCostKeyGrantedSneak} {
		t.Run(claim, func(t *testing.T) {
			g := newCatalogGame(t)
			_, attacker, me, opp := ninjaTeenBoard(t, g, 3)
			elektra := pushToGraveyard(g, me, elektraCard)
			keys := graveyardOfferKeys(g, me, elektra)
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
			assistCardSettle(t, g, me.ID, nil, true)
			c, ok := g.LookupCardForEffect(elektra)
			if !ok || !g.Battlefield.Contains(elektra) || !c.Tapped || c.AttackingTarget != opp.ID {
				t.Errorf("Elektra on battlefield=%v tapped=%v attacking %s, want tapped and attacking %s",
					g.Battlefield.Contains(elektra), c.Tapped, c.AttackingTarget, opp.ID)
			}
		})
	}

	t.Run("without Ninja Teen", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
		elektra := pushToGraveyard(g, me, elektraCard)
		declareAttack(t, g, opp.ID, attacker)
		advanceTo(t, g, game.StepDeclareBlockers)
		finishBlockDeclarations(t, g)
		floatMana(t, g, me, "{B}{B}{B}{B}")
		if err := g.CastSpell(me.ID, elektra, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "sneak", AltCostIDs: []uuid.UUID{attacker}}); err == nil {
			t.Error("Elektra sneaked out of a graveyard with no Ninja Teen")
		}
	})
}

// "Using their sneak abilities": a creature card's other alternative cost
// is not a way out of the graveyard, nor is any noncreature card.
func TestNinjaTeenOpensTheGraveyardOnlyToSneak(t *testing.T) {
	const otherAltOracle = "test-other-alt-cost-4680"
	registerForTest(t, Spec{
		OracleID: otherAltOracle, Name: "Cheap Brute", Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{Key: "test-cheap", Label: "Pay {B}", ManaCost: "{B}"}},
	})
	g := newCatalogGame(t)
	_, attacker, me, _ := ninjaTeenBoard(t, g, 3)
	brute := pushToGraveyard(g, me, game.Card{
		Name: "Cheap Brute", TypeLine: "Creature — Ogre", ManaCost: "{5}{B}", OracleID: otherAltOracle, Power: 5, Toughness: 5,
	})
	if err := g.CastSpell(me.ID, brute, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "test-cheap"}); err == nil {
		t.Error("a creature card's non-sneak alternative cost cast it from the graveyard")
	}
	if keys := graveyardOfferKeys(g, me, brute); !slices.Equal(keys, []string{game.AltCostKeyGrantedSneak}) {
		t.Errorf("offers %q, want only the granted sneak", keys)
	}
	bolt := pushToGraveyard(g, me, game.Card{Name: "Grave Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	if err := grantedSneakCast(g, me, bolt, attacker); err == nil {
		t.Error("an instant card was sneaked out of the graveyard")
	}
	if keys := graveyardOfferKeys(g, me, bolt); len(keys) != 0 {
		t.Errorf("instant offers %q, want none", keys)
	}
}

// Level 2 or lower grants nothing (CR 716.2a); level 2's anthem and menace
// are on from level 2 and stay at level 3.
func TestNinjaTeenLevelsGateTheirLines(t *testing.T) {
	for _, level := range []int{1, 2, 3} {
		g := newCatalogGame(t)
		bear, attacker, me, _ := ninjaTeenBoard(t, g, level)
		keys := graveyardOfferKeys(g, me, bear)
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

		rat := layeredCard(t, g, attacker)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushNinjaTeen(g, me, 1)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
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
	passPriorityAroundTable(t, g)
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
		{"snapshot", func(t *testing.T, g *game.Game) *game.Game { return restoreRoundTrip(t, g, true) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			bear, attacker, me, opp := ninjaTeenBoard(t, g, 3)
			if err := grantedSneakCast(g, me, bear, attacker); err != nil {
				t.Fatalf("granted sneak cast: %v", err)
			}
			restored := tc.reshape(t, g)
			assistCardSettle(t, restored, me.ID, nil, true)
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
	g := newCatalogGame(t)
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
