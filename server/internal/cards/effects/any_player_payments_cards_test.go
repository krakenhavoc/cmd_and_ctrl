package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_payments_cards_test.go — the fourteen cards of ADR 0150
// PR 2 (#1903): any-player offers paid with life, damage, a sacrifice,
// a discard, or nothing at all.

const (
	appAetherRift        = "2577ef8e-d85e-47af-ac34-629f31992751"
	appDashHopes         = "6a4ace6b-ef3a-4d11-89d7-958f98dcc445"
	appTemporalExtortion = "13da5255-2fc5-43a4-a44c-13a5d9ff77c3"
	appBookBurning       = "644a8946-4890-4c2b-978f-02eedc95b5ad"
	appBreakingPoint     = "a648b9ed-810a-41db-af40-3b5b4658db90"
	appBrowbeat          = "11cd0c42-4148-4d82-9172-b1a1967fe319"
	appBarbarianBully    = "5a88f616-986d-4e97-9457-26c783489123"
	appArgothianWurm     = "f8aa1893-60e5-439e-9316-872a9f6b6c94"
	appShivanWumpus      = "4b78222f-6973-42a3-987e-eb8cc8fd6c76"
	appBrainGorgers      = "3b9369ae-8827-45d1-af85-c777cde8e1bc"
	appProwlingPangolin  = "92042cad-09f5-464f-85ad-baad27e6d1eb"
	appPhantasmagorian   = "248e0dfa-0bcb-402c-badf-ca59f02792f7"
	appDistantMemories   = "12c78cc4-7fbe-4639-bef7-bf959a0d6d68"
)

// appOpenLink is the open link of an any-player offer, or nil.
func appOpenLink(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.AnyPlayerOffer != nil {
			return c
		}
	}
	return nil
}

// appAnswerWith answers every link of the open any-player offer in
// turn: `pay` says whether the link's chooser pays and with which cards.
// It returns the seats asked, in order, and settles the stack after.
func appAnswerWith(t *testing.T, g *game.Game, pay func(*game.PendingChoice) (bool, []uuid.UUID)) []uuid.UUID {
	t.Helper()
	var asked []uuid.UUID
	for i := 0; i < 8; i++ {
		link := appOpenLink(g)
		if link == nil {
			passPriorityAroundTable(t, g)
			return asked
		}
		asked = append(asked, link.Chooser)
		yes, cards := pay(link)
		if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, yes, cards); err != nil {
			t.Fatalf("answer link %d: %v", i, err)
		}
	}
	t.Fatal("the offer never ended")
	return nil
}

// appOnly pays as `who` (with `cards`, for a sacrifice or a discard)
// and declines as everyone else. uuid.Nil is nobody.
func appOnly(who uuid.UUID, cards ...uuid.UUID) func(*game.PendingChoice) (bool, []uuid.UUID) {
	return func(c *game.PendingChoice) (bool, []uuid.UUID) {
		if who != uuid.Nil && c.Chooser == who {
			return true, cards
		}
		return false, nil
	}
}

// appToUpkeepOf walks the turn to `seat`'s upkeep.
func appToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepUpkeep; i++ {
		if i > 40 {
			t.Fatalf("never reached seat %d's upkeep (at seat %d %v)", seat, g.Turn.ActiveSeat, g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func appInZone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == kind
}

// appCastAsSeat puts a card in `p`'s hand and casts it at instant speed
// (the caller is already in a main phase).
func appCastAsSeat(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// Dash Hopes: anyone paying 5 life counters it, and its target
// resolves; nobody paying lets it counter the target. A player short of
// 5 life can't pay (CR 119.4). The offer holds Dash Hopes on the stack.
func TestAnyPlayerPaysDashHopesIsCounteredByTheFirstToPayFiveLife(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := newCatalogGame(t)
		advanceToMainOf(t, g, 0)
		me, opp, short := g.Seats[0], g.Seats[1], g.Seats[2]
		short.Life = 4
		bolt := appCastAsSeat(t, g, opp, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		dash := appCastAsSeat(t, g, me, "Dash Hopes", "Instant", appDashHopes,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
		passPriorityAroundTable(t, g)
		link := appOpenLink(g)
		if link == nil || link.GuardsStackItem != dash || link.PayAction() == nil || link.PayAction().Kind != game.PayActionLife {
			t.Fatalf("paid=%v: the first link is %+v, want 5 life guarding Dash Hopes", paid, link)
		}
		if !appInZone(g, dash, game.ZoneStack) {
			t.Fatalf("paid=%v: Dash Hopes left the stack while the offer is open", paid)
		}
		before, oppBefore := me.Life, opp.Life
		pay := appOnly(uuid.Nil)
		if paid {
			// The short player says "pay" too: that is a decline.
			pay = func(c *game.PendingChoice) (bool, []uuid.UUID) {
				return c.Chooser == opp.ID || c.Chooser == short.ID, nil
			}
		}
		asked := appAnswerWith(t, g, pay)
		if paid {
			if !slices.Equal(asked, []uuid.UUID{me.ID, opp.ID}) {
				t.Errorf("paid: asked %v, want seat 0 then seat 1", asked)
			}
			if opp.Life != oppBefore-5 || !appInZone(g, dash, game.ZoneGraveyard) {
				t.Errorf("paid: payer at %d life, Dash Hopes in graveyard = %v", opp.Life, appInZone(g, dash, game.ZoneGraveyard))
			}
			if me.Life != before-3 {
				t.Errorf("paid: life %d, want the Bolt to have resolved (%d)", me.Life, before-3)
			}
			continue
		}
		if short.Life != 4 {
			t.Errorf("unpaid: the short player paid life")
		}
		if me.Life != before || !appInZone(g, bolt, game.ZoneGraveyard) {
			t.Errorf("unpaid: life %d (was %d), Bolt in graveyard = %v; want it countered", me.Life, before, appInZone(g, bolt, game.ZoneGraveyard))
		}
	}
}

// Temporal Extortion: each player is asked for half their own life,
// rounded up. A payer counters it; nobody paying takes the extra turn.
func TestAnyPlayerPaysTemporalExtortionAsksEachPlayerForHalfTheirLife(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		me.Life, opp.Life = 40, 21
		spell := castCatalogSpell(t, g, "Temporal Extortion", "Sorcery", appTemporalExtortion, nil)
		passPriorityAroundTable(t, g)
		var prices []string
		pay := func(c *game.PendingChoice) (bool, []uuid.UUID) {
			prices = append(prices, c.PayCost)
			return paid && c.Chooser == opp.ID, nil
		}
		appAnswerWith(t, g, pay)
		if len(prices) < 2 || prices[0] != "20 life" || prices[1] != "11 life" {
			t.Fatalf("paid=%v: prices %v, want 20 life then 11 life", paid, prices)
		}
		if paid {
			if opp.Life != 10 || !appInZone(g, spell, game.ZoneGraveyard) {
				t.Errorf("paid: payer at %d (want 10), spell in graveyard = %v", opp.Life, appInZone(g, spell, game.ZoneGraveyard))
			}
			endTurn(t, g)
			assertTurnOf(t, g, 1, false, "after a countered Temporal Extortion")
			continue
		}
		endTurn(t, g)
		assertTurnOf(t, g, 0, true, "after Temporal Extortion resolves")
	}
}

// Browbeat, Book Burning and Breaking Point: taking the damage stops the
// effect, and the spell deals it to the player who took it.
func TestAnyPlayerPaysDamageSorceriesHappenOnlyWhenNobodyTakesTheDamage(t *testing.T) {
	type check func(t *testing.T, g *game.Game, target *game.Player, bear uuid.UUID, paid bool)
	for _, tc := range []struct {
		name, oracle string
		damage       int
		targeted     bool
		check        check
	}{
		{"Browbeat", appBrowbeat, 5, true, nil}, // checked against the hand size below
		{"Book Burning", appBookBurning, 6, true, func(t *testing.T, _ *game.Game, target *game.Player, _ uuid.UUID, paid bool) {
			milled := target.Graveyard.Size() >= 6
			if milled == paid {
				t.Errorf("paid=%v: target milled = %v", paid, milled)
			}
		}},
		{"Breaking Point", appBreakingPoint, 6, false, func(t *testing.T, g *game.Game, _ *game.Player, bear uuid.UUID, paid bool) {
			if alive := appInZone(g, bear, game.ZoneBattlefield); alive != paid {
				t.Errorf("paid=%v: the creature is on the battlefield = %v", paid, alive)
			}
		}},
	} {
		for _, paid := range []bool{false, true} {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			bear := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
			var targets []game.TargetRef
			if tc.targeted {
				targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}
			}
			handBefore := me.Hand.Size()
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, targets)
			passPriorityAroundTable(t, g)
			link := appOpenLink(g)
			if link == nil || link.AcceptLabel == "" || link.PayAction() == nil || link.PayAction().Kind != game.PayActionDamage {
				t.Fatalf("%s paid=%v: the link is %+v, want a damage payment with the card's words", tc.name, paid, link)
			}
			who := uuid.Nil
			if paid {
				who = opp.ID
			}
			appAnswerWith(t, g, appOnly(who))
			if paid && opp.Life != 40-tc.damage {
				t.Errorf("%s paid: payer at %d life, want %d", tc.name, opp.Life, 40-tc.damage)
			}
			if tc.name == "Browbeat" {
				// The cast took one card out of the hand.
				drew := me.Hand.Size() - (handBefore)
				if (drew == 3) == paid {
					t.Errorf("Browbeat paid=%v: hand went %d → %d", paid, handBefore, me.Hand.Size())
				}
				continue
			}
			tc.check(t, g, me, bear, paid)
		}
	}
}

// Barbarian Bully: a player taking 4 damage from it stops the +2/+2.
func TestAnyPlayerPaysBarbarianBullyPumpsOnlyWhenNobodyTakesTheDamage(t *testing.T) {
	for _, tc := range []struct {
		paid bool
		pump int
	}{{false, 2}, {true, 0}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bully := pushCatalogPermanent(g, me.ID, "Barbarian Bully", "Creature — Human Barbarian", appBarbarianBully, false)
		base, _ := effectivePTOf(t, g, bully)
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, bully, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if me.Hand.Size() != hand-1 {
			t.Fatalf("the random discard was not paid: hand %d → %d", hand, me.Hand.Size())
		}
		passPriorityAroundTable(t, g)
		who := uuid.Nil
		if tc.paid {
			who = opp.ID
		}
		appAnswerWith(t, g, appOnly(who))
		if p, _ := effectivePTOf(t, g, bully); p != base+tc.pump {
			t.Errorf("paid=%v: power %d, want %d", tc.paid, p, base+tc.pump)
		}
		if tc.paid && opp.Life != 36 {
			t.Errorf("paid: payer at %d, want 36", opp.Life)
		}
		if err := g.ActivateCatalogAbility(me.ID, bully, 0, game.ActivateAbilityParams{}); err == nil {
			t.Errorf("paid=%v: a second activation this turn was allowed", tc.paid)
		}
	}
}

// Argothian Wurm and Shivan Wumpus: anyone sacrificing a land puts the
// creature on top of its owner's library; a player with no land can't
// pay.
func TestAnyPlayerPaysALandSacrificeTucksTheWurm(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Argothian Wurm", appArgothianWurm, "Creature — Wurm"},
		{"Shivan Wumpus", appShivanWumpus, "Creature — Beast"},
	} {
		for _, paid := range []bool{false, true} {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			land := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			wurm := castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			link := appOpenLink(g)
			if link == nil || link.PayCost != "sacrifice a land" {
				t.Fatalf("%s paid=%v: the link is %+v, want a land sacrifice", tc.name, paid, link)
			}
			pay := appOnly(uuid.Nil)
			if paid {
				pay = func(c *game.PendingChoice) (bool, []uuid.UUID) {
					if c.Chooser == opp.ID {
						return true, []uuid.UUID{land}
					}
					// Seat 0 has no land: its "pay" is a decline.
					return c.Chooser == me.ID, nil
				}
			}
			appAnswerWith(t, g, pay)
			if onTop := appInZone(g, wurm, game.ZoneLibrary); onTop != paid {
				t.Errorf("%s paid=%v: in a library = %v", tc.name, paid, onTop)
			}
			if gone := appInZone(g, land, game.ZoneGraveyard); gone != paid {
				t.Errorf("%s paid=%v: the land was sacrificed = %v", tc.name, paid, gone)
			}
			if top, err := me.Library.Top(); paid && (err != nil || top.InstanceID != wurm) {
				t.Errorf("%s paid: not on top of its owner's library", tc.name)
			}
		}
	}
}

// Brain Gorgers and Phantasmagorian: a creature sacrificed or three
// cards discarded counters the spell.
func TestAnyPlayerPaysCreatureCountersAreCounteredByASacrificeOrADiscard(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		discard      bool
	}{{"Brain Gorgers", appBrainGorgers, false}, {"Phantasmagorian", appPhantasmagorian, true}} {
		for _, paid := range []bool{false, true} {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			bear := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
			spell := castCatalogSpell(t, g, tc.name, "Creature — Horror", tc.oracle, nil)
			passPriorityAroundTable(t, g)
			link := appOpenLink(g)
			if link == nil || link.GuardsStackItem != spell {
				t.Fatalf("%s paid=%v: the link %+v does not guard the spell", tc.name, paid, link)
			}
			cards := []uuid.UUID{bear}
			if tc.discard {
				cards = []uuid.UUID{opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID, opp.Hand.Cards[2].InstanceID}
			}
			who := uuid.Nil
			if paid {
				who = opp.ID
			}
			appAnswerWith(t, g, appOnly(who, cards...))
			if countered := appInZone(g, spell, game.ZoneGraveyard); countered != paid {
				t.Errorf("%s paid=%v: countered = %v", tc.name, paid, countered)
			}
			for _, id := range cards {
				if gone := appInZone(g, id, game.ZoneGraveyard); gone != paid {
					t.Errorf("%s paid=%v: payment card in graveyard = %v", tc.name, paid, gone)
				}
			}
		}
	}
}

// Phantasmagorian's own ability returns it from the graveyard for
// three discards.
func TestPhantasmagorianReturnsFromTheGraveyardForThreeDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Phantasmagorian", TypeLine: "Creature — Horror", OracleID: appPhantasmagorian, Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	discards := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID, me.Hand.Cards[2].InstanceID}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: discards}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) || me.Hand.Size() != hand-3+1 {
		t.Errorf("in hand = %v, hand %d → %d; want it back for three cards", me.Hand.Contains(id), hand, me.Hand.Size())
	}
}

// Prowling Pangolin: two creatures sacrificed sacrifices it; a player
// with one creature can't pay.
func TestAnyPlayerPaysProwlingPangolinIsSacrificedForTwoCreatures(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, one := g.Seats[0], g.Seats[1], g.Seats[2]
		a := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		b := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		apaPush(g, one.ID, one.ID, game.Card{Name: "Lone Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		pangolin := castCatalogSpell(t, g, "Prowling Pangolin", "Creature — Pangolin Beast", appProwlingPangolin, nil)
		passPriorityAroundTable(t, g)
		pay := func(c *game.PendingChoice) (bool, []uuid.UUID) {
			switch {
			case paid && c.Chooser == opp.ID:
				return true, []uuid.UUID{a, b}
			case c.Chooser == one.ID:
				return true, nil // one creature: "pay" is a decline
			}
			return false, nil
		}
		asked := appAnswerWith(t, g, pay)
		if !paid && len(asked) != 4 {
			t.Errorf("unpaid: asked %v, want all four", asked)
		}
		if alive := appInZone(g, pangolin, game.ZoneBattlefield); alive == paid {
			t.Errorf("paid=%v: the Pangolin is on the battlefield = %v", paid, alive)
		}
		if paid && (appInZone(g, a, game.ZoneBattlefield) || appInZone(g, b, game.ZoneBattlefield)) {
			t.Error("paid: the payer's creatures survived")
		}
		_ = me
	}
}

// Innocent Traveler: only opponents are asked; a sacrifice stops the
// transform, and nobody paying transforms it into Malicious Invader,
// which gets +2/+0 while an opponent controls a Human.
func TestAnyPlayerPaysInnocentTravelerTransformsUnlessAnOpponentSacrifices(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := newCatalogGame(t)
		owner, opp := g.Seats[1], g.Seats[2]
		c := game.Card{
			InstanceID: uuid.New(), OracleID: innocentTravelerOracle, Layout: game.LayoutTransform,
			Owner: owner.ID, Controller: owner.ID,
			Faces: []game.Face{
				{Name: "Innocent Traveler", TypeLine: "Creature — Human", ManaCost: "{2}{B}{B}", Colors: []string{"B"}, Power: 1, Toughness: 3},
				{Name: "Malicious Invader", TypeLine: "Creature — Vampire", Colors: []string{"B"}, Power: 3, Toughness: 3},
			},
		}
		c.SetFace(0)
		traveler := pushBattlefieldCardWithTimestamp(g, c)
		victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Villager", TypeLine: "Creature — Human", Power: 1, Toughness: 1})
		appToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		who := uuid.Nil
		if paid {
			who = opp.ID
		}
		asked := appAnswerWith(t, g, appOnly(who, victim))
		if slices.Contains(asked, owner.ID) {
			t.Errorf("paid=%v: the controller was asked (%v)", paid, asked)
		}
		if paid {
			if !slices.Equal(asked, []uuid.UUID{opp.ID}) {
				t.Errorf("paid: asked %v, want seat 2 first", asked)
			}
			if rvBattlefieldCard(g, traveler).Name != "Innocent Traveler" {
				t.Error("paid: it transformed anyway")
			}
			continue
		}
		if len(asked) != 3 {
			t.Errorf("unpaid: asked %v, want the three opponents", asked)
		}
		if got := rvBattlefieldCard(g, traveler); got == nil || got.Name != "Malicious Invader" {
			t.Fatalf("unpaid: it did not transform: %+v", got)
		}
		// An opponent still controls a Human (the Villager).
		if p, _ := effectivePTOf(t, g, traveler); p != 5 {
			t.Errorf("unpaid: Malicious Invader's power %d, want 3 + 2", p)
		}
	}
}

// Distant Memories: only opponents are asked; a yes puts the exiled
// card into the caster's hand, and nobody saying yes draws three.
func TestAnyPlayerPaysDistantMemoriesHandsOverTheCardOrDrawsThree(t *testing.T) {
	for _, yes := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		needle := pushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Artifact"})
		castCatalogSpell(t, g, "Distant Memories", "Sorcery", appDistantMemories, nil)
		passPriorityAroundTable(t, g)
		search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
		if search == nil {
			t.Fatal("no search")
		}
		if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{needle}); err != nil {
			t.Fatalf("search: %v", err)
		}
		if !appInZone(g, needle, game.ZoneExile) {
			t.Fatal("the found card was not exiled")
		}
		link := appOpenLink(g)
		if link == nil || link.Chooser != opp.ID || link.PayAction() == nil || link.PayAction().Kind != game.PayActionNothing {
			t.Fatalf("yes=%v: the first link is %+v, want seat 1 and no payment", yes, link)
		}
		hand := me.Hand.Size()
		who := uuid.Nil
		if yes {
			who = opp.ID
		}
		asked := appAnswerWith(t, g, appOnly(who))
		if slices.Contains(asked, me.ID) {
			t.Errorf("yes=%v: the caster was asked", yes)
		}
		if yes {
			if !me.Hand.Contains(needle) || me.Hand.Size() != hand+1 {
				t.Errorf("yes: the card is not in hand, or something was drawn (hand %d → %d)", hand, me.Hand.Size())
			}
			continue
		}
		if !appInZone(g, needle, game.ZoneExile) || me.Hand.Size() != hand+3 {
			t.Errorf("no: card exiled = %v, hand %d → %d; want it left in exile and three drawn", appInZone(g, needle, game.ZoneExile), hand, me.Hand.Size())
		}
	}
}

// Aether Rift: a discarded creature card comes back unless someone pays
// 5 life; a noncreature card is not offered at all.
func TestAnyPlayerPaysAetherRiftReturnsTheCreatureUnlessSomeonePays(t *testing.T) {
	for _, tc := range []struct {
		name     string
		creature bool
		paid     bool
	}{{"creature, nobody pays", true, false}, {"creature, paid", true, true}, {"noncreature", false, false}} {
		g := newCatalogGame(t)
		owner, opp := g.Seats[1], g.Seats[2]
		pushCatalogPermanent(g, owner.ID, "Aether Rift", "Enchantment", appAetherRift, false)
		appToUpkeepOf(t, g, 1)
		owner.Hand.Cards = owner.Hand.Cards[:0]
		typeLine := "Artifact"
		if tc.creature {
			typeLine = "Creature — Beast"
		}
		card := uuid.New()
		owner.Hand.PushTop(game.Card{InstanceID: card, Name: "Discarded", TypeLine: typeLine, Power: 3, Toughness: 3, Owner: owner.ID, Controller: owner.ID})
		passPriorityAroundTable(t, g)
		link := appOpenLink(g)
		if !tc.creature {
			if link != nil {
				t.Errorf("%s: an offer was made", tc.name)
			}
			if !appInZone(g, card, game.ZoneGraveyard) {
				t.Errorf("%s: the card was not discarded", tc.name)
			}
			continue
		}
		if link == nil || link.Chooser != owner.ID || link.PayCost != "5 life" {
			t.Fatalf("%s: the first link is %+v, want the Rift's controller and 5 life", tc.name, link)
		}
		who := uuid.Nil
		if tc.paid {
			who = opp.ID
		}
		appAnswerWith(t, g, appOnly(who))
		if back := appInZone(g, card, game.ZoneBattlefield); back == tc.paid {
			t.Errorf("%s: the creature is on the battlefield = %v", tc.name, back)
		}
		if tc.paid && opp.Life != 35 {
			t.Errorf("%s: the payer is at %d life, want 35", tc.name, opp.Life)
		}
	}
}
