package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 83 (S82, #5507): Commander Legends token makers. Every test plays
// the real catalog entry and asserts what its text says.

const (
	oldB083FirstResponse = "0fe66bc0-8eda-4625-a068-6a5de826466f"
	oldB083Araumi        = "1a62ba93-153c-4bed-9f6a-ff914df360c1"
	oldB083Transfusion   = "2eaee808-b72a-4950-ab4f-ab8e8d12f838"
	oldB083Nevinyrral    = "3ddde895-12aa-488a-94fb-27a1064bdb72"
	oldB083Blob          = "7fa8647a-7668-4db9-a77d-87e004b160e1"
	oldB083GorMuldrak    = "9abdd53a-aa62-447b-9525-02a2227cc44b"
	oldB083SzatsWill     = "f0d8bde1-76db-4bd2-ae90-3e3a3418de0b"
)

func oldB083MakeCommander(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
}

func TestOldB083EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB083FirstResponse: "First Response", oldB083Araumi: "Araumi of the Dead Tide",
		oldB083Transfusion: "Profane Transfusion", oldB083Nevinyrral: "Nevinyrral, Urborg Tyrant",
		oldB083Blob: "Biowaste Blob", oldB083GorMuldrak: "Gor Muldrak, Amphinologist", oldB083SzatsWill: "Szat's Will",
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

// --- First Response ------------------------------------------------------

func TestOldB083FirstResponseMakesASoldierWhenYouLostLifeLastTurn(t *testing.T) {
	for _, tc := range []struct {
		name              string
		oppLoses, meLoses bool
		want              int
	}{{"I lost life", false, true, 1}, {"only an opponent lost life", true, false, 0}, {"nobody lost life", false, false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "First Response", "Enchantment", oldB083FirstResponse, 0, 0)
			bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			if tc.oppLoses {
				effectstest.GainLifeNow(t, g, bear, opp, -2)
			}
			if tc.meLoses {
				effectstest.GainLifeNow(t, g, bear, me, -2)
			}
			effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
			effectstest.PassPriorityAroundTable(t, g)
			if got := len(effectstest.S46Tokens(g, me.ID, "Soldier")); got != tc.want {
				t.Errorf("%d Soldier tokens, want %d", got, tc.want)
			}
		})
	}
}

// --- Gor Muldrak, Amphinologist ------------------------------------------

func TestOldB083GorMuldrakProtectsYourPermanentsFromSalamanders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gor Muldrak, Amphinologist", "Legendary Creature — Human Scout", oldB083GorMuldrak, 3, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "protection from Salamanders") {
		t.Error("my creature lacks protection from Salamanders")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "protection from Salamanders") {
		t.Error("an opposing creature has protection from Salamanders")
	}
}

func TestOldB083GorMuldrakGivesEachPlayerWithTheFewestCreaturesASalamander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, p1, p2, p3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.B12Push(g, me.ID, "Gor Muldrak, Amphinologist", "Legendary Creature — Human Scout", oldB083GorMuldrak, 3, 2)
	effectstest.B12Creature(g, p2.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		p    *game.Player
		want int
	}{{me, 0}, {p1, 1}, {p2, 0}, {p3, 1}} {
		if got := len(effectstest.S46Tokens(g, tc.p.ID, "Salamander Warrior")); got != tc.want {
			t.Errorf("seat %v has %d Salamander Warriors, want %d", tc.p.ID, got, tc.want)
		}
	}
}

// --- Biowaste Blob -------------------------------------------------------

func TestOldB083BiowasteBlobPumpsOozesAndCopiesItselfWithACommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blob := effectstest.B12Push(g, me.ID, "Biowaste Blob", "Creature — Ooze", oldB083Blob, 0, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, blob); p != 1 || tough != 1 {
		t.Errorf("a lone Blob is %d/%d, want 1/1", p, tough)
	}
	ooze := effectstest.B12Creature(g, me.ID, "My Ooze", "Creature — Ooze", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ooze", "Creature — Ooze", 1, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, ooze); p != 2 {
		t.Errorf("my Ooze has power %d, want 2", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 1 {
		t.Errorf("their Ooze has power %d, want 1", p)
	}
	// No commander: no copy at the next upkeep of mine.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Biowaste Blob")); got != 0 {
		t.Fatalf("%d Blob tokens without a commander", got)
	}
	oldB083MakeCommander(g, ooze)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Biowaste Blob")); got != 1 {
		t.Errorf("%d Blob tokens with a commander, want 1", got)
	}
}

// --- Profane Transfusion -------------------------------------------------

func TestOldB083ProfaneTransfusionSwapsLifeAndMakesAHorrorOfTheDifference(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.StSetLife(t, g, me, 5)
	effectstest.StSetLife(t, g, opp, 40)
	id := effectstest.HandCardFull(me, "Profane Transfusion", "Sorcery", "{6}{B}{B}{B}", oldB083Transfusion, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("cast with one target player")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40 || opp.Life != 5 {
		t.Errorf("life totals %d and %d, want 40 and 5", me.Life, opp.Life)
	}
	horrors := effectstest.S46Tokens(g, me.ID, "Horror")
	if len(horrors) != 1 {
		t.Fatalf("%d Horror tokens, want 1", len(horrors))
	}
	if p, tough := effectstest.EffectivePTOf(t, g, horrors[0].InstanceID); p != 35 || tough != 35 {
		t.Errorf("Horror is %d/%d, want 35/35", p, tough)
	}
	if !horrors[0].HasCardType("artifact") || !horrors[0].IsCreature() || len(horrors[0].Colors) != 0 {
		t.Errorf("Horror is %q, colours %v: want a colorless artifact creature", horrors[0].TypeLine, horrors[0].Colors)
	}
}

func TestOldB083ProfaneTransfusionBetweenTwoOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	effectstest.StSetLife(t, g, a, 3)
	effectstest.StSetLife(t, g, b, 12)
	id := effectstest.HandCardFull(me, "Profane Transfusion", "Sorcery", "{6}{B}{B}{B}", oldB083Transfusion, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if a.Life != 12 || b.Life != 3 {
		t.Errorf("life totals %d and %d, want 12 and 3", a.Life, b.Life)
	}
	horrors := effectstest.S46Tokens(g, me.ID, "Horror")
	if len(horrors) != 1 {
		t.Fatalf("%d Horror tokens for the caster, want 1", len(horrors))
	}
	if p, _ := effectstest.EffectivePTOf(t, g, horrors[0].InstanceID); p != 9 {
		t.Errorf("Horror has power %d, want 9", p)
	}
}

// --- Nevinyrral, Urborg Tyrant -------------------------------------------

func TestOldB083NevinyrralHasHexproofFromArtifactsCreaturesAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nev := effectstest.B12Push(g, me.ID, "Nevinyrral, Urborg Tyrant", "Legendary Creature — Zombie Wizard", oldB083Nevinyrral, 3, 6)
	for _, kw := range []string{"hexproof from artifacts", "hexproof from creatures", "hexproof from enchantments"} {
		if !effectstest.HasEffectiveKeyword(t, g, nev, kw) {
			t.Errorf("Nevinyrral lacks %q", kw)
		}
		if _, ok := game.ParseHexproofFromQuality(kw); !ok {
			t.Errorf("%q is not a quality the engine reads", kw)
		}
	}
}

func TestOldB083NevinyrralMakesATappedZombiePerCreatureThatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.TrekB18BDestroy(t, g, a)
	effectstest.TrekB18BDestroy(t, g, b)
	effectstest.PassPriorityAroundTable(t, g)
	id := effectstest.HandCardFull(me, "Nevinyrral, Urborg Tyrant", "Legendary Creature — Zombie Wizard", "{3}{W}{U}{B}", oldB083Nevinyrral, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	zombies := effectstest.S46Tokens(g, me.ID, "Zombie")
	if len(zombies) != 2 {
		t.Fatalf("%d Zombies, want 2", len(zombies))
	}
	for _, z := range zombies {
		if !z.Tapped {
			t.Error("a Zombie entered untapped")
		}
	}
}

func TestOldB083NevinyrralDiesAndYouMayPayOneToDestroyEverything(t *testing.T) {
	for _, pay := range []bool{true, false} {
		name := "declined"
		if pay {
			name = "paid"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			nev := effectstest.B12Push(g, me.ID, "Nevinyrral, Urborg Tyrant", "Legendary Creature — Zombie Wizard", oldB083Nevinyrral, 3, 6)
			bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			relic := effectstest.PushTypedCard(g, opp.ID, "Relic", "Artifact", "{2}")
			effectstest.TrekB18BDestroy(t, g, nev)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.B06AddMana(me, "R")
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
			effectstest.PassPriorityAroundTable(t, g)
			gone := !g.Battlefield.Contains(bear) && !g.Battlefield.Contains(relic)
			if gone != pay {
				t.Errorf("everything destroyed = %v, want %v", gone, pay)
			}
		})
	}
}

// --- Szat's Will ---------------------------------------------------------

func TestOldB083SzatsWillIsChooseOneWithoutACommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Szat's Will", "Instant", "{4}{B}", oldB083SzatsWill, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Strict: true}); err == nil {
		t.Fatal("both modes were accepted without a commander")
	}
}

func TestOldB083SzatsWillMakesEachOpponentSacrificeTheirBiggestCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 9, 9)
	id := effectstest.HandCardFull(me, "Szat's Will", "Instant", "{4}{B}", oldB083SzatsWill, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.S46Settle(t, g)
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Errorf("big on board %v, small %v, mine %v", g.Battlefield.Contains(big), g.Battlefield.Contains(small), g.Battlefield.Contains(mine))
	}
}

func TestOldB083SzatsWillExilesGraveyardsForThrullsEqualToTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.PushGraveyardTyped(opp, "Dead Giant", "Creature — Giant")
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == big {
				opp.Graveyard.Cards[i].Power, opp.Graveyard.Cards[i].Toughness = 4, 4
			}
		}
	})
	spell := effectstest.PushGraveyardTyped(opp, "Dead Spell", "Sorcery")
	mine := effectstest.PushGraveyardTyped(me, "My Dead Bear", "Creature — Bear")
	id := effectstest.HandCardFull(me, "Szat's Will", "Instant", "{4}{B}", oldB083SzatsWill, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.S46Settle(t, g)
	if !g.Exile.Contains(big) || !g.Exile.Contains(spell) {
		t.Error("an opponent's graveyard was not exiled")
	}
	if g.Exile.Contains(mine) {
		t.Error("my own graveyard was exiled")
	}
	if got := len(effectstest.S46Tokens(g, me.ID, "Thrull")); got != 4 {
		t.Errorf("%d Thrulls, want 4", got)
	}
}

func TestOldB083SzatsWillWithACommanderDoesBothAndCountsTheSacrificedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cmdr := effectstest.B12Creature(g, me.ID, "My Commander", "Creature — Bear", 2, 2)
	oldB083MakeCommander(g, cmdr)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 6, 6)
	id := effectstest.HandCardFull(me, "Szat's Will", "Instant", "{4}{B}", oldB083SzatsWill, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}}); err != nil {
		t.Fatalf("both modes with a commander: %v", err)
	}
	effectstest.S46Settle(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the opponent kept their biggest creature")
	}
	if !g.Exile.Contains(big) {
		t.Error("the sacrificed creature card was not exiled with the graveyard")
	}
	if got := len(effectstest.S46Tokens(g, me.ID, "Thrull")); got != 6 {
		t.Errorf("%d Thrulls, want 6", got)
	}
}

// --- Araumi of the Dead Tide ---------------------------------------------

func TestOldB083AraumiExilesOneCardPerOpponentAndGrantsEncoreForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	araumi := effectstest.B12Push(g, me.ID, "Araumi of the Dead Tide", "Legendary Creature — Merfolk Wizard", oldB083Araumi, 1, 4)
	effectstest.AdvanceToMain(t, g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Knight", TypeLine: "Creature — Human Knight", ManaCost: "{1}{W}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var fodder []uuid.UUID
	for i := 0; i < 3; i++ {
		fodder = append(fodder, effectstest.PushGraveyardTyped(me, "Fodder", "Instant"))
	}
	// Two cards is not enough with three opponents.
	if err := g.ActivateCatalogAbility(me.ID, araumi, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(dead), ExileIDs: fodder[:2]}); err == nil {
		t.Fatal("Araumi activated exiling fewer cards than I have opponents")
	}
	effectstest.B16Activate(t, g, me.ID, araumi, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(dead), ExileIDs: fodder})
	effectstest.PassPriorityAroundTable(t, g)
	for _, f := range fodder {
		if !g.Exile.Contains(f) {
			t.Error("a fodder card was not exiled as the cost")
		}
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("encore was activated for {C} when its cost is {1}{W}")
	}
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("encore for its mana cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Dead Knight")); got != 3 {
		t.Errorf("%d token copies, want one per opponent", got)
	}
}

func TestOldB083AraumiEncoreEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	araumi := effectstest.B12Push(g, me.ID, "Araumi of the Dead Tide", "Legendary Creature — Merfolk Wizard", oldB083Araumi, 1, 4)
	effectstest.AdvanceToMain(t, g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Knight", TypeLine: "Creature — Human Knight", ManaCost: "{1}{W}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var fodder []uuid.UUID
	for i := 0; i < 3; i++ {
		fodder = append(fodder, effectstest.PushGraveyardTyped(me, "Fodder", "Instant"))
	}
	effectstest.B16Activate(t, g, me.ID, araumi, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(dead), ExileIDs: fodder})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.FloatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the granted encore outlived the turn")
	}
}
