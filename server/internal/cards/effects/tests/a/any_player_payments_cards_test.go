package testsa

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dash Hopes: anyone paying 5 life counters it, and its target
// resolves; nobody paying lets it counter the target. A player short of
// 5 life can't pay (CR 119.4). The offer holds Dash Hopes on the stack.
func TestAnyPlayerPaysDashHopesIsCounteredByTheFirstToPayFiveLife(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMainOf(t, g, 0)
		me, opp, short := g.Seats[0], g.Seats[1], g.Seats[2]
		short.Life = 4
		bolt := effectstest.AppCastAsSeat(t, g, opp, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		dash := effectstest.AppCastAsSeat(t, g, me, "Dash Hopes", "Instant", effectstest.AppDashHopes,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
		effectstest.PassPriorityAroundTable(t, g)
		link := effectstest.AppOpenLink(g)
		if link == nil || link.GuardsStackItem != dash || link.PayAction() == nil || link.PayAction().Kind != game.PayActionLife {
			t.Fatalf("paid=%v: the first link is %+v, want 5 life guarding Dash Hopes", paid, link)
		}
		if !effectstest.AppInZone(g, dash, game.ZoneStack) {
			t.Fatalf("paid=%v: Dash Hopes left the stack while the offer is open", paid)
		}
		before, oppBefore := me.Life, opp.Life
		pay := effectstest.AppOnly(uuid.Nil)
		if paid {
			// The short player says "pay" too: that is a decline.
			pay = func(c *game.PendingChoice) (bool, []uuid.UUID) {
				return c.Chooser == opp.ID || c.Chooser == short.ID, nil
			}
		}
		asked := effectstest.AppAnswerWith(t, g, pay)
		if paid {
			if !slices.Equal(asked, []uuid.UUID{me.ID, opp.ID}) {
				t.Errorf("paid: asked %v, want seat 0 then seat 1", asked)
			}
			if opp.Life != oppBefore-5 || !effectstest.AppInZone(g, dash, game.ZoneGraveyard) {
				t.Errorf("paid: payer at %d life, Dash Hopes in graveyard = %v", opp.Life, effectstest.AppInZone(g, dash, game.ZoneGraveyard))
			}
			if me.Life != before-3 {
				t.Errorf("paid: life %d, want the Bolt to have resolved (%d)", me.Life, before-3)
			}
			continue
		}
		if short.Life != 4 {
			t.Errorf("unpaid: the short player paid life")
		}
		if me.Life != before || !effectstest.AppInZone(g, bolt, game.ZoneGraveyard) {
			t.Errorf("unpaid: life %d (was %d), Bolt in graveyard = %v; want it countered", me.Life, before, effectstest.AppInZone(g, bolt, game.ZoneGraveyard))
		}
	}
}

// Temporal Extortion: each player is asked for half their own life,
// rounded up. A payer counters it; nobody paying takes the extra turn.
func TestAnyPlayerPaysTemporalExtortionAsksEachPlayerForHalfTheirLife(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		me.Life, opp.Life = 40, 21
		spell := effectstest.CastCatalogSpell(t, g, "Temporal Extortion", "Sorcery", effectstest.AppTemporalExtortion, nil)
		effectstest.PassPriorityAroundTable(t, g)
		var prices []string
		pay := func(c *game.PendingChoice) (bool, []uuid.UUID) {
			prices = append(prices, c.PayCost)
			return paid && c.Chooser == opp.ID, nil
		}
		effectstest.AppAnswerWith(t, g, pay)
		if len(prices) < 2 || prices[0] != "20 life" || prices[1] != "11 life" {
			t.Fatalf("paid=%v: prices %v, want 20 life then 11 life", paid, prices)
		}
		if paid {
			if opp.Life != 10 || !effectstest.AppInZone(g, spell, game.ZoneGraveyard) {
				t.Errorf("paid: payer at %d (want 10), spell in graveyard = %v", opp.Life, effectstest.AppInZone(g, spell, game.ZoneGraveyard))
			}
			effectstest.EndTurn(t, g)
			effectstest.AssertTurnOf(t, g, 1, false, "after a countered Temporal Extortion")
			continue
		}
		effectstest.EndTurn(t, g)
		effectstest.AssertTurnOf(t, g, 0, true, "after Temporal Extortion resolves")
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
		{"Browbeat", effectstest.AppBrowbeat, 5, true, nil}, // checked against the hand size below
		{"Book Burning", effectstest.AppBookBurning, 6, true, func(t *testing.T, _ *game.Game, target *game.Player, _ uuid.UUID, paid bool) {
			milled := target.Graveyard.Size() >= 6
			if milled == paid {
				t.Errorf("paid=%v: target milled = %v", paid, milled)
			}
		}},
		{"Breaking Point", effectstest.AppBreakingPoint, 6, false, func(t *testing.T, g *game.Game, _ *game.Player, bear uuid.UUID, paid bool) {
			if alive := effectstest.AppInZone(g, bear, game.ZoneBattlefield); alive != paid {
				t.Errorf("paid=%v: the creature is on the battlefield = %v", paid, alive)
			}
		}},
	} {
		for _, paid := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			bear := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
			var targets []game.TargetRef
			if tc.targeted {
				targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}
			}
			handBefore := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, targets)
			effectstest.PassPriorityAroundTable(t, g)
			link := effectstest.AppOpenLink(g)
			if link == nil || link.AcceptLabel == "" || link.PayAction() == nil || link.PayAction().Kind != game.PayActionDamage {
				t.Fatalf("%s paid=%v: the link is %+v, want a damage payment with the card's words", tc.name, paid, link)
			}
			who := uuid.Nil
			if paid {
				who = opp.ID
			}
			effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who))
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bully := effectstest.PushCatalogPermanent(g, me.ID, "Barbarian Bully", "Creature — Human Barbarian", effectstest.AppBarbarianBully, false)
		base, _ := effectstest.EffectivePTOf(t, g, bully)
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, bully, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if me.Hand.Size() != hand-1 {
			t.Fatalf("the random discard was not paid: hand %d → %d", hand, me.Hand.Size())
		}
		effectstest.PassPriorityAroundTable(t, g)
		who := uuid.Nil
		if tc.paid {
			who = opp.ID
		}
		effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who))
		if p, _ := effectstest.EffectivePTOf(t, g, bully); p != base+tc.pump {
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
		{"Argothian Wurm", effectstest.AppArgothianWurm, "Creature — Wurm"},
		{"Shivan Wumpus", effectstest.AppShivanWumpus, "Creature — Beast"},
	} {
		for _, paid := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			land := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			wurm := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			link := effectstest.AppOpenLink(g)
			if link == nil || link.PayCost != "sacrifice a land" {
				t.Fatalf("%s paid=%v: the link is %+v, want a land sacrifice", tc.name, paid, link)
			}
			pay := effectstest.AppOnly(uuid.Nil)
			if paid {
				pay = func(c *game.PendingChoice) (bool, []uuid.UUID) {
					if c.Chooser == opp.ID {
						return true, []uuid.UUID{land}
					}
					// Seat 0 has no land: its "pay" is a decline.
					return c.Chooser == me.ID, nil
				}
			}
			effectstest.AppAnswerWith(t, g, pay)
			if onTop := effectstest.AppInZone(g, wurm, game.ZoneLibrary); onTop != paid {
				t.Errorf("%s paid=%v: in a library = %v", tc.name, paid, onTop)
			}
			if gone := effectstest.AppInZone(g, land, game.ZoneGraveyard); gone != paid {
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
	}{{"Brain Gorgers", effectstest.AppBrainGorgers, false}, {"Phantasmagorian", effectstest.AppPhantasmagorian, true}} {
		for _, paid := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			bear := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
			spell := effectstest.CastCatalogSpell(t, g, tc.name, "Creature — Horror", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			link := effectstest.AppOpenLink(g)
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
			effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who, cards...))
			if countered := effectstest.AppInZone(g, spell, game.ZoneGraveyard); countered != paid {
				t.Errorf("%s paid=%v: countered = %v", tc.name, paid, countered)
			}
			for _, id := range cards {
				if gone := effectstest.AppInZone(g, id, game.ZoneGraveyard); gone != paid {
					t.Errorf("%s paid=%v: payment card in graveyard = %v", tc.name, paid, gone)
				}
			}
		}
	}
}

// Phantasmagorian's own ability returns it from the graveyard for
// three discards.
func TestPhantasmagorianReturnsFromTheGraveyardForThreeDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Phantasmagorian", TypeLine: "Creature — Horror", OracleID: effectstest.AppPhantasmagorian, Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	discards := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID, me.Hand.Cards[2].InstanceID}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: discards}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) || me.Hand.Size() != hand-3+1 {
		t.Errorf("in hand = %v, hand %d → %d; want it back for three cards", me.Hand.Contains(id), hand, me.Hand.Size())
	}
}

// Prowling Pangolin: two creatures sacrificed sacrifices it; a player
// with one creature can't pay.
func TestAnyPlayerPaysProwlingPangolinIsSacrificedForTwoCreatures(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, one := g.Seats[0], g.Seats[1], g.Seats[2]
		a := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		b := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		effectstest.ApaPush(g, one.ID, one.ID, game.Card{Name: "Lone Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		pangolin := effectstest.CastCatalogSpell(t, g, "Prowling Pangolin", "Creature — Pangolin Beast", effectstest.AppProwlingPangolin, nil)
		effectstest.PassPriorityAroundTable(t, g)
		pay := func(c *game.PendingChoice) (bool, []uuid.UUID) {
			switch {
			case paid && c.Chooser == opp.ID:
				return true, []uuid.UUID{a, b}
			case c.Chooser == one.ID:
				return true, nil // one creature: "pay" is a decline
			}
			return false, nil
		}
		asked := effectstest.AppAnswerWith(t, g, pay)
		if !paid && len(asked) != 4 {
			t.Errorf("unpaid: asked %v, want all four", asked)
		}
		if alive := effectstest.AppInZone(g, pangolin, game.ZoneBattlefield); alive == paid {
			t.Errorf("paid=%v: the Pangolin is on the battlefield = %v", paid, alive)
		}
		if paid && (effectstest.AppInZone(g, a, game.ZoneBattlefield) || effectstest.AppInZone(g, b, game.ZoneBattlefield)) {
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
		g := effectstest.NewCatalogGame(t)
		owner, opp := g.Seats[1], g.Seats[2]
		c := game.Card{
			InstanceID: uuid.New(), OracleID: cardshu.InnocentTravelerOracle, Layout: game.LayoutTransform,
			Owner: owner.ID, Controller: owner.ID,
			Faces: []game.Face{
				{Name: "Innocent Traveler", TypeLine: "Creature — Human", ManaCost: "{2}{B}{B}", Colors: []string{"B"}, Power: 1, Toughness: 3},
				{Name: "Malicious Invader", TypeLine: "Creature — Vampire", Colors: []string{"B"}, Power: 3, Toughness: 3},
			},
		}
		c.SetFace(0)
		traveler := effectstest.PushBattlefieldCardWithTimestamp(g, c)
		victim := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Villager", TypeLine: "Creature — Human", Power: 1, Toughness: 1})
		effectstest.AppToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		who := uuid.Nil
		if paid {
			who = opp.ID
		}
		asked := effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who, victim))
		if slices.Contains(asked, owner.ID) {
			t.Errorf("paid=%v: the controller was asked (%v)", paid, asked)
		}
		if paid {
			if !slices.Equal(asked, []uuid.UUID{opp.ID}) {
				t.Errorf("paid: asked %v, want seat 2 first", asked)
			}
			if effectstest.RvBattlefieldCard(g, traveler).Name != "Innocent Traveler" {
				t.Error("paid: it transformed anyway")
			}
			continue
		}
		if len(asked) != 3 {
			t.Errorf("unpaid: asked %v, want the three opponents", asked)
		}
		if got := effectstest.RvBattlefieldCard(g, traveler); got == nil || got.Name != "Malicious Invader" {
			t.Fatalf("unpaid: it did not transform: %+v", got)
		}
		// An opponent still controls a Human (the Villager).
		if p, _ := effectstest.EffectivePTOf(t, g, traveler); p != 5 {
			t.Errorf("unpaid: Malicious Invader's power %d, want 3 + 2", p)
		}
	}
}

// Distant Memories: only opponents are asked; a yes puts the exiled
// card into the caster's hand, and nobody saying yes draws three.
func TestAnyPlayerPaysDistantMemoriesHandsOverTheCardOrDrawsThree(t *testing.T) {
	for _, yes := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Artifact"})
		effectstest.CastCatalogSpell(t, g, "Distant Memories", "Sorcery", effectstest.AppDistantMemories, nil)
		effectstest.PassPriorityAroundTable(t, g)
		search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
		if search == nil {
			t.Fatal("no search")
		}
		if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{needle}); err != nil {
			t.Fatalf("search: %v", err)
		}
		if !effectstest.AppInZone(g, needle, game.ZoneExile) {
			t.Fatal("the found card was not exiled")
		}
		link := effectstest.AppOpenLink(g)
		if link == nil || link.Chooser != opp.ID || link.PayAction() == nil || link.PayAction().Kind != game.PayActionNothing {
			t.Fatalf("yes=%v: the first link is %+v, want seat 1 and no payment", yes, link)
		}
		hand := me.Hand.Size()
		who := uuid.Nil
		if yes {
			who = opp.ID
		}
		asked := effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who))
		if slices.Contains(asked, me.ID) {
			t.Errorf("yes=%v: the caster was asked", yes)
		}
		if yes {
			if !me.Hand.Contains(needle) || me.Hand.Size() != hand+1 {
				t.Errorf("yes: the card is not in hand, or something was drawn (hand %d → %d)", hand, me.Hand.Size())
			}
			continue
		}
		if !effectstest.AppInZone(g, needle, game.ZoneExile) || me.Hand.Size() != hand+3 {
			t.Errorf("no: card exiled = %v, hand %d → %d; want it left in exile and three drawn", effectstest.AppInZone(g, needle, game.ZoneExile), hand, me.Hand.Size())
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
		g := effectstest.NewCatalogGame(t)
		owner, opp := g.Seats[1], g.Seats[2]
		effectstest.PushCatalogPermanent(g, owner.ID, "Aether Rift", "Enchantment", effectstest.AppAetherRift, false)
		effectstest.AppToUpkeepOf(t, g, 1)
		owner.Hand.Cards = owner.Hand.Cards[:0]
		typeLine := "Artifact"
		if tc.creature {
			typeLine = "Creature — Beast"
		}
		card := uuid.New()
		owner.Hand.PushTop(game.Card{InstanceID: card, Name: "Discarded", TypeLine: typeLine, Power: 3, Toughness: 3, Owner: owner.ID, Controller: owner.ID})
		effectstest.PassPriorityAroundTable(t, g)
		link := effectstest.AppOpenLink(g)
		if !tc.creature {
			if link != nil {
				t.Errorf("%s: an offer was made", tc.name)
			}
			if !effectstest.AppInZone(g, card, game.ZoneGraveyard) {
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
		effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who))
		if back := effectstest.AppInZone(g, card, game.ZoneBattlefield); back == tc.paid {
			t.Errorf("%s: the creature is on the battlefield = %v", tc.name, back)
		}
		if tc.paid && opp.Life != 35 {
			t.Errorf("%s: the payer is at %d life, want 35", tc.name, opp.Life)
		}
	}
}
