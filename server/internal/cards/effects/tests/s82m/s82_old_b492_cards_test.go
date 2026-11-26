package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 492 (Zendikar damage, #5583): Magma Rift, Seismic Shudder,
// Murasa Pyromancer, Runeflare Trap, Hellfire Mongrel, Kor Duelist, Lavaball
// Trap, Electropotence, Ruinous Minotaur and Spire Barrage.

const (
	oldB492Rift       = "0996d48a-f92a-4aaf-b8a3-f7ec8ff986ca"
	oldB492Shudder    = "10c0f705-4c23-4761-afa9-4a29cf90e8ae"
	oldB492Pyromancer = "1d6925f3-d132-4fa1-bde0-82d0128c319f"
	oldB492Runeflare  = "259efd92-416e-42d9-b501-ac1338fd35fc"
	oldB492Mongrel    = "633a0919-9ac3-4645-8f0b-3f00c198db79"
	oldB492Duelist    = "758dd2fd-d3ec-4782-9a60-12b0d79719da"
	oldB492Lavaball   = "a9cdfb6c-6087-4b0d-b8a5-175f3546b23a"
	oldB492Electro    = "c2459c41-82b4-4ca2-b634-11b55ddb63d9"
	oldB492Minotaur   = "ece9d3e5-79e5-4c56-9e8d-8abcd819c041"
	oldB492Barrage    = "ef84d5e2-5db7-40e7-af49-a2f6f494a8a4"
)

func TestOldB492EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB492Rift: "Magma Rift", oldB492Shudder: "Seismic Shudder", oldB492Pyromancer: "Murasa Pyromancer",
		oldB492Runeflare: "Runeflare Trap", oldB492Mongrel: "Hellfire Mongrel", oldB492Duelist: "Kor Duelist",
		oldB492Lavaball: "Lavaball Trap", oldB492Electro: "Electropotence", oldB492Minotaur: "Ruinous Minotaur",
		oldB492Barrage: "Spire Barrage",
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

func TestOldB492MagmaRiftSacrificesALandAndDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	big := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	id := effectstest.B13HandCard(me, "Magma Rift", "Sorcery", oldB492Rift)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(big)}); err == nil {
		t.Fatal("cast without sacrificing a land")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{land}, Targets: effectstest.CardTarget(big)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed as the spell was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("5 damage did not kill the 5/5")
	}
}

func TestOldB492SeismicShudderSparesFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Rat", "Creature — Rat", 1, 1)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	sturdy := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Seismic Shudder", "Instant", oldB492Shudder, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a 1/1 without flying survived")
	}
	if !g.Battlefield.Contains(flyer) {
		t.Error("the flyer died")
	}
	if !g.Battlefield.Contains(sturdy) {
		t.Error("the 2/2 died to 1 damage")
	}
}

func TestOldB492MurasaPyromancerDamagesByTheNumberOfAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Kor Scout", "Creature — Kor Scout Ally", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Murasa Pyromancer", TypeLine: "Creature — Human Shaman Ally", OracleID: oldB492Pyromancer, ManaCost: "{4}{R}{R}", Power: 3, Toughness: 2})
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		effectstest.PassPriorityUntilChoice(t, g)
		p = effectstest.LatestPickTarget(g, me.ID)
	}
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Allies (the Scout and the Pyromancer): 2 damage, not enough for a 3/3.
	if !g.Battlefield.Contains(victim) {
		t.Fatal("2 damage killed a 3/3")
	}
	if got := effectstest.DamageOn(t, g, victim); got != 2 {
		t.Errorf("damage marked %d, want 2 (two Allies)", got)
	}
}

func TestOldB492RuneflareTrapCostNeedsAnOpponentWhoDrewThreeCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	trap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: trap, Name: "Runeflare Trap", TypeLine: "Instant — Trap", ManaCost: "{4}{R}{R}", OracleID: oldB492Runeflare, Owner: me.ID, Controller: me.ID})
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err == nil {
		t.Fatal("the trap cost was accepted before anyone drew")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3) })
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err == nil {
		t.Fatal("the trap cost was accepted when only the caster drew three")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err == nil {
		t.Fatal("the trap cost was accepted after two draws")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	hand := len(opp.Hand.Cards)
	life := opp.Life
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err != nil {
		t.Fatalf("the trap cost was refused after three draws: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != hand {
		t.Errorf("the opponent lost %d life, want %d (their hand size)", got, hand)
	}
}

func TestOldB492HellfireMongrelPunishesAnEmptyishHandAtTheirUpkeep(t *testing.T) {
	for _, hand := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		effectstest.HmlCreature(g, g.Seats[0].ID, "Hellfire Mongrel", oldB492Mongrel, 2, 2)
		for len(opp.Hand.Cards) > hand {
			opp.Hand.Cards = opp.Hand.Cards[1:]
		}
		for len(opp.Hand.Cards) < hand {
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		}
		life := opp.Life
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		// The draw step comes after the upkeep, so the hand is what it was.
		want := life
		if hand <= 2 {
			want = life - 2
		}
		if opp.Life != want {
			t.Errorf("hand %d: life %d → %d, want %d", hand, life, opp.Life, want)
		}
	}
}

func TestOldB492KorDuelistHasDoubleStrikeOnlyWhileEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	duelist := effectstest.HmlCreature(g, me.ID, "Kor Duelist", oldB492Duelist, 1, 1)
	if effectstest.HasEffectiveKeyword(t, g, duelist, "double strike") {
		t.Fatal("an unequipped Kor Duelist has double strike")
	}
	sword := effectstest.SeedEquipment(g, me.ID, "Sword", "")
	var err error
	g.WithWriteLock(func() { err = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: duelist}) })
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, duelist, "double strike") {
		t.Error("an equipped Kor Duelist lacks double strike")
	}
	other := effectstest.HmlCreature(g, me.ID, "Bear", "", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, other, "double strike") {
		t.Error("another creature got double strike")
	}
}

func TestOldB492LavaballTrapDestroysTwoLandsAndDamagesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	l1 := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	l2 := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	mine := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 2, 5)
	small := effectstest.B12Creature(g, opp.ID, "Rat", "Creature — Rat", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	trap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: trap, Name: "Lavaball Trap", TypeLine: "Instant — Trap", ManaCost: "{6}{R}{R}", OracleID: oldB492Lavaball, Owner: me.ID, Controller: me.ID})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: l1}, {Kind: game.TargetCard, ID: l2}}
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err == nil {
		t.Fatal("the trap cost was accepted with no opponent lands entered")
	}
	g.WithWriteLock(func() {
		if _, err := g.CreateTokensForEffect(opp.ID, game.Card{Name: "Land Token", TypeLine: "Token Land", Power: 0, Toughness: 0}, 2, game.TokenEntryOptions{}); err != nil {
			t.Errorf("CreateTokensForEffect: %v", err)
		}
	})
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets[:1]}); err == nil {
		t.Fatal("cast with one target land")
	}
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err != nil {
		t.Fatalf("the trap cost was refused once two lands entered: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(l1) || g.Battlefield.Contains(l2) {
		t.Error("a target land survived")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived 4 damage")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness 5 died to 4 damage")
	}
}

func TestOldB492ElectropotenceLetsAnEnteringCreatureDealItsPower(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Electropotence", "Enchantment", oldB492Electro, false)
		effectstest.AdvanceToMain(t, g)
		life := opp.Life
		effectstest.Msc17Cast(t, g, me, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3}, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatalf("pay=%v: no target prompt for the trigger: %+v", pay, g.PendingChoices)
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		if pay {
			effectstest.B10AddMana(me, "R", "R", "R")
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		got := life - opp.Life
		if pay && got != 3 {
			t.Errorf("paid: the opponent lost %d, want 3 (the Ogre's power)", got)
		}
		if !pay && got != 0 {
			t.Errorf("declined: the opponent lost %d, want 0", got)
		}
	}
}

func TestOldB492RuinousMinotaurSacrificesALandWhenItHurtsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minotaur := effectstest.HmlCreature(g, me.ID, "Ruinous Minotaur", oldB492Minotaur, 5, 2)
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.DealCombatDamageToPlayer(g, minotaur, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestChooseCardsFor(g, me.ID); pick != nil {
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
}

func TestOldB492RuinousMinotaurDoesNothingWithoutALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	minotaur := effectstest.HmlCreature(g, me.ID, "Ruinous Minotaur", oldB492Minotaur, 5, 2)
	effectstest.DealCombatDamageToPlayer(g, minotaur, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(minotaur) {
		t.Error("the Minotaur left the battlefield")
	}
}

func TestOldB492SpireBarrageCountsMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Spire Barrage", "Sorcery", oldB492Barrage, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d, want 3 (my three Mountains)", got)
	}
}
