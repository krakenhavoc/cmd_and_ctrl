package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 5 (group A) — Secret Lair planeswalkers: Angrath, the
// Flame-Chained; Chandra, Flamecaller; Jaya, Venerated Firemage; Davriel,
// Rogue Shadowmage (#5462).

const (
	s82OldB005Angrath = "092ceffb-1ee4-452c-b15e-5b79267d42a6"
	s82OldB005Chandra = "0c2a9131-f3d7-4f71-8bcc-3c169574b2e3"
	s82OldB005Jaya    = "10a87c58-2d8b-4658-b7d6-8dcbb44a2d2f"
	s82OldB005Davriel = "1bd5e51f-f9a5-4eb7-bda2-957310602c20"
)

func TestS82OldB005ACardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB005Angrath: "Angrath, the Flame-Chained", s82OldB005Chandra: "Chandra, Flamecaller",
		s82OldB005Jaya: "Jaya, Venerated Firemage", s82OldB005Davriel: "Davriel, Rogue Shadowmage",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func s82OldB005EmptyHand(p *game.Player) { p.Hand.Cards = nil }

func s82OldB005Hand(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func TestS82OldB005AngrathPlusOneMakesEachOpponentDiscardAndLoseTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Angrath, the Flame-Chained", s82OldB005Angrath, 4)
	var opps []*game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opps = append(opps, s)
		}
	}
	s82OldB005EmptyHand(opps[0])
	s82OldB005Hand(opps[0], 2)
	s82OldB005EmptyHand(opps[1])
	s82OldB005EmptyHand(opps[2])
	lives := []int{opps[0].Life, opps[1].Life, opps[2].Life}
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, opps[0])
	effectstest.PassPriorityAroundTable(t, g)
	if opps[0].Hand.Size() != 1 {
		t.Errorf("opponent hand %d, want 1 after discarding one of two", opps[0].Hand.Size())
	}
	for i, o := range opps {
		if o.Life != lives[i]-2 {
			t.Errorf("opponent %d life %d, want %d", i, o.Life, lives[i]-2)
		}
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB005AngrathMinusThreeStealsAndSacrificesAtEndStepIfSmall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	walker := effectstest.S82Walker(g, me.ID, "Angrath, the Flame-Chained", s82OldB005Angrath, 4)
	small := effectstest.PushSizedCreatureForTest(g, opp.ID, "Small", "Creature — Goblin", "{1}{R}", []string{"R"}, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == small {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(small)})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, small)
	if !ok || c.Controller != me.ID || c.Tapped {
		t.Fatalf("stolen creature: present %v, controller %v, tapped %v", ok, c.Controller, c.Tapped)
	}
	if !effectstest.HasEffectiveKeyword(t, g, small, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a mana value 2 creature survived the end step")
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestS82OldB005AngrathMinusThreeSparesALargeCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	walker := effectstest.S82Walker(g, me.ID, "Angrath, the Flame-Chained", s82OldB005Angrath, 4)
	big := effectstest.PushSizedCreatureForTest(g, opp.ID, "Big", "Creature — Giant", "{3}{R}{R}", []string{"R"}, 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(big)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, big)
	if !ok || c.Controller != me.ID {
		t.Errorf("mana value 5 creature: present %v, controller %v; want kept by me until end of turn", ok, c.Controller)
	}
}

func TestS82OldB005AngrathMinusEightDrainsByGraveyardSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Angrath, the Flame-Chained", s82OldB005Angrath, 8)
	var opps []*game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opps = append(opps, s)
		}
	}
	for i, o := range opps {
		o.Graveyard.Cards = nil
		for j := 0; j < i+1; j++ {
			o.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Gone", TypeLine: "Sorcery", Owner: o.ID, Controller: o.ID})
		}
	}
	lives := []int{opps[0].Life, opps[1].Life, opps[2].Life}
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i, o := range opps {
		if o.Life != lives[i]-(i+1) {
			t.Errorf("opponent %d life %d, want %d", i, o.Life, lives[i]-(i+1))
		}
	}
}

func s82OldB005Elementals(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Elemental" {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

func TestS82OldB005ChandraPlusOneMakesTwoHasteElementalsExiledAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Chandra, Flamecaller", s82OldB005Chandra, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	els := s82OldB005Elementals(g, me.ID)
	if len(els) != 2 {
		t.Fatalf("%d Elementals, want 2", len(els))
	}
	for _, id := range els {
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.Power != 3 || c.Toughness != 1 || !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("Elemental is %d/%d", c.Power, c.Toughness)
		}
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(s82OldB005Elementals(g, me.ID)); n != 0 {
		t.Errorf("%d Elementals left after the end step", n)
	}
}

func TestS82OldB005ChandraZeroDiscardsTheHandAndDrawsOneMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Chandra, Flamecaller", s82OldB005Chandra, 4)
	effectstest.AdvanceToMain(t, g)
	s82OldB005EmptyHand(me)
	s82OldB005Hand(me, 3)
	lib := me.Library.Size()
	gy := me.Graveyard.Size()
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 4 || me.Graveyard.Size()-gy != 3 || lib-me.Library.Size() != 4 {
		t.Errorf("hand %d, graveyard +%d, library -%d; want 4, +3, -4", me.Hand.Size(), me.Graveyard.Size()-gy, lib-me.Library.Size())
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

func TestS82OldB005ChandraZeroWithAnEmptyHandDrawsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Chandra, Flamecaller", s82OldB005Chandra, 4)
	effectstest.AdvanceToMain(t, g)
	s82OldB005EmptyHand(me)
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d, want 1", me.Hand.Size())
	}
}

func TestS82OldB005ChandraMinusXDamagesEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	walker := effectstest.S82Walker(g, me.ID, "Chandra, Flamecaller", s82OldB005Chandra, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness 3 or less survived 3 damage")
	}
	if c, ok := effectstest.BattlefieldCard(g, big); !ok || c.DamageMarked != 3 {
		t.Errorf("big creature: present %v, damage %d, want 3", ok, c.DamageMarked)
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestS82OldB005ChandraMinusXRefusesMoreThanItsLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.S82Walker(g, me.ID, "Chandra, Flamecaller", s82OldB005Chandra, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, walker, 2, game.ActivateAbilityParams{Strict: true, XValue: 5}); err == nil {
		t.Error("X = 5 with 4 loyalty was allowed")
	}
}

func s82OldB005Damage(g *game.Game, src, target uuid.UUID, n int) {
	g.WithWriteLock(func() {
		if g.PlayerByIDForEffect(target) != nil {
			_ = g.DealDamageToPlayerForEffect(src, target, n)
		} else {
			_ = g.DealDamageToCreatureForEffect(src, target, n)
		}
	})
}

func TestS82OldB005JayaAddsOneToAnotherRedSourceYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Venerated Firemage", s82OldB005Jaya, 5)
	red := effectstest.PushSizedCreatureForTest(g, me.ID, "Red", "Creature — Goblin", "{R}", []string{"R"}, 1, 1)
	green := effectstest.PushSizedCreatureForTest(g, me.ID, "Green", "Creature — Elf", "{G}", []string{"G"}, 1, 1)
	theirRed := effectstest.PushSizedCreatureForTest(g, opp.ID, "Their Red", "Creature — Goblin", "{R}", []string{"R"}, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 9)
	life := opp.Life
	s82OldB005Damage(g, red, opp.ID, 2)
	if opp.Life != life-3 {
		t.Errorf("red source dealt %d to a player, want 3", life-opp.Life)
	}
	s82OldB005Damage(g, green, opp.ID, 2)
	if opp.Life != life-5 {
		t.Errorf("a green source was boosted: total %d, want 5", life-opp.Life)
	}
	meLife := me.Life
	s82OldB005Damage(g, theirRed, me.ID, 2)
	if me.Life != meLife-2 {
		t.Errorf("an opponent's red source dealt %d, want 2", meLife-me.Life)
	}
	s82OldB005Damage(g, red, victim, 2)
	if c, _ := effectstest.BattlefieldCard(g, victim); c.DamageMarked != 3 {
		t.Errorf("permanent took %d, want 3", c.DamageMarked)
	}
	if !g.Battlefield.Contains(jaya) {
		t.Fatal("Jaya left")
	}
}

func TestS82OldB005JayaDoesNotBoostItselfOrAnOpponentsRedSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Venerated Firemage", s82OldB005Jaya, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == jaya {
				g.Battlefield.Cards[i].Colors = []string{"R"}
				g.Battlefield.Cards[i].ManaCost = "{4}{R}"
			}
		}
	})
	theirRed := effectstest.PushSizedCreatureForTest(g, opp.ID, "Their Red", "Creature — Goblin", "{R}", []string{"R"}, 1, 1)
	meLife := me.Life
	s82OldB005Damage(g, theirRed, me.ID, 2)
	if me.Life != meLife-2 {
		t.Errorf("opponent's red source dealt %d to me, want 2", meLife-me.Life)
	}
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.S82Activate(t, g, me.ID, jaya, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("Jaya's −2 dealt %d, want exactly 2 (she is not 'another' source)", life-opp.Life)
	}
	if got := effectstest.LoyaltyOf(g, jaya); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestS82OldB005JayaStopsBoostingWhenSheLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	jaya := effectstest.S82Walker(g, me.ID, "Jaya, Venerated Firemage", s82OldB005Jaya, 5)
	red := effectstest.PushSizedCreatureForTest(g, me.ID, "Red", "Creature — Goblin", "{R}", []string{"R"}, 1, 1)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, jaya); err != nil {
		t.Fatal(err)
	}
	life := opp.Life
	s82OldB005Damage(g, red, opp.ID, 2)
	if opp.Life != life-2 {
		t.Errorf("dealt %d with Jaya gone, want 2", life-opp.Life)
	}
}

func TestS82OldB005DavrielDamagesAnOpponentWithOneOrFewerCardsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.S82Walker(g, me.ID, "Davriel, Rogue Shadowmage", s82OldB005Davriel, 3)
	opp := g.Seats[(seat+1)%4]
	s82OldB005EmptyHand(opp)
	s82OldB005Hand(opp, 1)
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%4)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

func TestS82OldB005DavrielIgnoresAnOpponentWithTwoCardsAndYourOwnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.S82Walker(g, me.ID, "Davriel, Rogue Shadowmage", s82OldB005Davriel, 3)
	opp := g.Seats[(seat+1)%4]
	s82OldB005EmptyHand(opp)
	s82OldB005Hand(opp, 2)
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%4)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent with two cards took %d", life-opp.Life)
	}
}

func TestS82OldB005DavrielMinusOneMakesATargetPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	walker := effectstest.S82Walker(g, me.ID, "Davriel, Rogue Shadowmage", s82OldB005Davriel, 3)
	s82OldB005EmptyHand(opp)
	s82OldB005Hand(opp, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 1 {
		t.Errorf("opponent hand %d, want 1", opp.Hand.Size())
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}
