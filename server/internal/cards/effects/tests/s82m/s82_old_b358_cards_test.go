package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b358_cards_test.go — S82 OLD batch 358 (#5556, Magic 2015, part 2):
// Resolute Archangel, Wall of Essence, Profane Memento, Stain the Mind,
// Chronostutter, Necrobite, Nissa, Worldwaker, Jace, the Living Guildpact,
// Chandra, Pyromaster, Restock, Soul of Shandalar and Soul of Ravnica. Each
// test plays the real catalog entry through the engine.

const (
	oldB358Archangel   = "0d6dc95d-55ce-4ef8-b065-d7203f5bf642"
	oldB358WallEssence = "1374ac18-01bd-452f-af72-a5f96285ed69"
	oldB358Memento     = "c81d1890-e741-49ec-a6ce-86c27fd4d502"
	oldB358StainMind   = "1994d563-8726-49c9-a18a-6ba2604a8cb7"
	oldB358Chrono      = "a7c46e5f-6e3a-405b-86fa-404d3b6971db"
	oldB358Necrobite   = "d2a3e219-08f3-4c05-b239-abf80a1c1c6b"
	oldB358Nissa       = "468f9b8e-4423-4131-8611-d34b0aa3d96f"
	oldB358Jace        = "55259de3-561d-4a9e-bb28-63c6a7ac5c68"
	oldB358Chandra     = "68e12748-9a7b-4e99-a6de-be097a269342"
	oldB358Restock     = "8b641cd3-bb08-4f2e-9dad-24f58a9c32e1"
	oldB358Shandalar   = "ac96d27c-dbae-4094-9209-3ef143f64713"
	oldB358Ravnica     = "eb014442-5935-4333-8078-ee4896dce653"
)

func TestOldB358EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB358Archangel: "Resolute Archangel", oldB358WallEssence: "Wall of Essence",
		oldB358Memento: "Profane Memento", oldB358StainMind: "Stain the Mind",
		oldB358Chrono: "Chronostutter", oldB358Necrobite: "Necrobite",
		oldB358Nissa: "Nissa, Worldwaker", oldB358Jace: "Jace, the Living Guildpact",
		oldB358Chandra: "Chandra, Pyromaster", oldB358Restock: "Restock",
		oldB358Shandalar: "Soul of Shandalar", oldB358Ravnica: "Soul of Ravnica",
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

func oldB358Walker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker — Test", OracleID: oracle,
		Counters: map[string]int{game.CounterLoyalty: loyalty}, Owner: owner, Controller: owner,
	})
}

func oldB358Activate(t *testing.T, g *game.Game, who, source uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, source, idx, params); err != nil {
		t.Fatalf("activate row %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Resolute Archangel --------------------------------------------------------

func TestOldB358ArchangelRestoresALowLifeTotalToTheStartingTotalOnly(t *testing.T) {
	for _, tc := range []struct {
		name        string
		life, want  int
		useStarting bool
	}{{"below", 7, 0, true}, {"at", 0, 0, true}, {"above", 99, 99, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		start := effects.StartingLifeOf(g)
		me.Life = tc.life
		if tc.useStarting {
			if tc.name == "at" {
				me.Life = start
			}
			tc.want = start
		}
		effectstest.CastCatalogSpell(t, g, "Resolute Archangel", "Creature — Angel", oldB358Archangel, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != tc.want {
			t.Errorf("%s: life = %d, want %d", tc.name, me.Life, tc.want)
		}
	}
}

// --- Wall of Essence -----------------------------------------------------------

func TestOldB358WallOfEssenceGainsLifeOnlyFromCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 3, 3)
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall of Essence", TypeLine: "Creature — Wall", OracleID: oldB358WallEssence,
		Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	life := opp.Life
	effectstest.Damage(t, g, attacker, wall, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("noncombat damage gained life: %d, want %d", opp.Life, life)
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+3 {
		t.Errorf("life = %d, want %d (3 combat damage gained)", opp.Life, life+3)
	}
}

// --- Profane Memento -----------------------------------------------------------

func TestOldB358MementoGainsOneLifeForAnOpponentsCreatureCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Profane Memento", "Artifact", oldB358Memento, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, land); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life = %d, want %d: my own creature or an opponent's land gained life", me.Life, life)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
}

// --- Chronostutter, Necrobite, Restock -----------------------------------------

func TestOldB358ChronostutterPutsTheCreatureSecondFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	top := effectstest.SeedLibrary(opp, "First", "Second", "Third")
	_ = me
	effectstest.CastCatalogSpell(t, g, "Chronostutter", "Instant", oldB358Chrono, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature is still on the battlefield")
	}
	if len(opp.Library.Cards) < 3 {
		t.Fatalf("library has %d cards", len(opp.Library.Cards))
	}
	n := len(opp.Library.Cards)
	if opp.Library.Cards[n-1].InstanceID != top[0] || opp.Library.Cards[n-2].InstanceID != bear {
		t.Errorf("the creature is not second from the top")
	}
}

func TestOldB358NecrobiteGivesDeathtouchAndARegenerationShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Bear", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Necrobite", "Instant", oldB358Necrobite, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "deathtouch") {
		t.Error("the creature did not gain deathtouch")
	}
	effectstest.Damage(t, g, src, bear, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
}

func TestOldB358RestockReturnsTwoCardsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushCatalogGraveyardCard(me, "Card A", "Sorcery", "", 0, 0)
	b := effectstest.PushCatalogGraveyardCard(me, "Card B", "Creature — Bear", "", 2, 2)
	spell := effectstest.CastCatalogSpell(t, g, "Restock", "Sorcery", oldB358Restock, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two targeted cards did not return to hand")
	}
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Restock is in %v, want exile", z)
	}
}

// --- Stain the Mind ------------------------------------------------------------

func TestOldB358StainTheMindExilesEveryNonlandCardWithTheNameFromThreeZones(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := func(name, typ string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typ, Owner: opp.ID, Controller: opp.ID}
	}
	var ids []uuid.UUID
	var other uuid.UUID
	g.WithWriteLock(func() {
		for _, z := range []*game.Zone{opp.Hand, opp.Graveyard, opp.Library} {
			c := mk("Lightning Bolt", "Instant")
			ids = append(ids, c.InstanceID)
			z.PushTop(c)
		}
		o := mk("Shock", "Instant")
		other = o.InstanceID
		opp.Hand.PushTop(o)
	})
	effectstest.CastCatalogSpell(t, g, "Stain the Mind", "Sorcery", oldB358StainMind, effectstest.P8Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Lightning Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a Lightning Bolt is not in exile (zone %v)", z)
		}
	}
	if z := g.FindCardZoneForEffect(other); z == nil || z.Kind != game.ZoneHand {
		t.Error("the Shock left the hand")
	}
}

func TestOldB358StainTheMindIgnoresALandName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushHandCardOfType(opp, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Stain the Mind", "Sorcery", oldB358StainMind, effectstest.P8Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Forest"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(land) {
		t.Error("a land card was exiled by a name Stain the Mind may not choose")
	}
}

// --- Souls ---------------------------------------------------------------------

func TestOldB358SoulOfShandalarDamagesThePlayerAndOneOfTheirCreatures(t *testing.T) {
	for _, fromGraveyard := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 5)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 5)
		var soul uuid.UUID
		if fromGraveyard {
			soul = effectstest.PushCatalogGraveyardCard(me, "Soul of Shandalar", "Creature — Avatar", oldB358Shandalar, 6, 6)
		} else {
			soul = effectstest.PushCatalogPermanent(g, me.ID, "Soul of Shandalar", "Creature — Avatar", oldB358Shandalar, false)
		}
		idx := 0
		if fromGraveyard {
			idx = 1
		}
		if err := g.ActivateCatalogAbility(me.ID, soul, idx, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: mine}}}); err == nil {
			t.Fatalf("graveyard=%v: the creature its own controller controls was accepted as the second target", fromGraveyard)
		}
		life := opp.Life
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}{R}{R}"); err != nil {
			t.Fatal(err)
		}
		if err := g.ActivateCatalogAbility(me.ID, soul, idx, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: theirs}}}); err != nil {
			t.Fatalf("graveyard=%v: activate: %v", fromGraveyard, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-3 {
			t.Errorf("graveyard=%v: life = %d, want %d", fromGraveyard, opp.Life, life-3)
		}
		if got := effectstest.DamageOn(t, g, theirs); got != 3 {
			t.Errorf("graveyard=%v: their creature has %d damage, want 3", fromGraveyard, got)
		}
		if fromGraveyard && me.Graveyard.Contains(soul) {
			t.Error("the Soul was not exiled from the graveyard as a cost")
		}
	}
}

func TestOldB358SoulOfRavnicaDrawsForEachColorAmongYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soul := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soul of Ravnica", TypeLine: "Creature — Avatar", OracleID: oldB358Ravnica,
		Power: 6, Toughness: 6, Colors: []string{"U"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.Pr7Creature(g, me.ID, "Red One", 1, "R")
	effectstest.Pr7Creature(g, me.ID, "Red Two", 1, "R")
	effectstest.Pr7Creature(g, me.ID, "Green One", 1, "G")
	effectstest.B12Permanent(g, me.ID, "Colorless Rock", "Artifact")
	before := me.Hand.Size()
	oldB358Activate(t, g, me.ID, soul, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d cards, want 3 (blue, red, green)", got)
	}
}

// --- Nissa, Worldwaker ---------------------------------------------------------

func TestOldB358NissaAnimatesALandAndUntapsForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	nissa := oldB358Walker(g, me.ID, "Nissa, Worldwaker", oldB358Nissa, 3)
	land := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	oldB358Activate(t, g, me.ID, nissa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)})
	c, _ := effectstest.BattlefieldCard(g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Fatalf("the land is %q, want a land Elemental creature", c.TypeLine)
	}
	if p, tt := effectstest.EffectivePower(t, g, land), effectstest.EffectiveToughness(t, g, land); p != 4 || tt != 4 {
		t.Errorf("the land is %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasKeywordOf(g, land, "trample") {
		t.Error("the land has no trample")
	}
}

func TestOldB358NissaUntapsUpToFourForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	nissa := oldB358Walker(g, me.ID, "Nissa, Worldwaker", oldB358Nissa, 3)
	var forests []uuid.UUID
	for i := 0; i < 4; i++ {
		f := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Tapped: true, Owner: me.ID, Controller: me.ID,
		})
		forests = append(forests, f)
	}
	oldB358Activate(t, g, me.ID, nissa, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(forests...)})
	for _, f := range forests {
		if effectstest.B16Tapped(t, g, f) {
			t.Error("a targeted Forest is still tapped")
		}
	}
}

func TestOldB358NissaUltimateFetchesAnyNumberOfBasicsAndAnimatesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	nissa := oldB358Walker(g, me.ID, "Nissa, Worldwaker", oldB358Nissa, 7)
	var lands [2]uuid.UUID
	g.WithWriteLock(func() {
		for i := range lands {
			c := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
			lands[i] = c.InstanceID
			me.Library.PushTop(c)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, nissa, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, lands[0], lands[1])
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range lands {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("a fetched Forest did not enter the battlefield")
		}
		if p := effectstest.EffectivePower(t, g, id); p != 4 {
			t.Errorf("a fetched Forest has power %d, want 4", p)
		}
	}
}

// --- Jace, the Living Guildpact ------------------------------------------------

func TestOldB358JacePlusOnePutsOneOfTheTopTwoIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jace := oldB358Walker(g, me.ID, "Jace, the Living Guildpact", oldB358Jace, 5)
	ids := effectstest.SeedLibrary(me, "Top", "Next")
	if err := g.ActivateCatalogAbility(me.ID, jace, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[1]) || me.Graveyard.Contains(ids[0]) {
		t.Error("the picked card is not in the graveyard alone")
	}
	if got := effectstest.LoyaltyOf(g, jace); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestOldB358JaceMinusThreeBouncesAnotherNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jace := oldB358Walker(g, me.ID, "Jace, the Living Guildpact", oldB358Jace, 5)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Error("a land was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(jace)}); err == nil {
		t.Error("Jace targeted himself")
	}
	oldB358Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestOldB358JaceUltimateWheelsTheTableAndYouDrawSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jace := oldB358Walker(g, me.ID, "Jace, the Living Guildpact", oldB358Jace, 8)
	oppCard := effectstest.PushHandCardOfType(opp, "Their Card", "Sorcery")
	oppYard := effectstest.PushCatalogGraveyardCard(opp, "Their Dead", "Sorcery", "", 0, 0)
	meYard := effectstest.PushCatalogGraveyardCard(me, "My Dead", "Sorcery", "", 0, 0)
	oppHand := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, jace, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(oppCard) || opp.Graveyard.Contains(oppYard) || me.Graveyard.Contains(meYard) {
		t.Error("a hand or graveyard card was not shuffled into its owner's library")
	}
	if me.Hand.Size() != 7 {
		t.Errorf("my hand = %d, want 7", me.Hand.Size())
	}
	if opp.Hand.Size() != 0 || oppHand == 0 {
		t.Errorf("the opponent's hand = %d (was %d), want 0: only you draw", opp.Hand.Size(), oppHand)
	}
}

// --- Chandra, Pyromaster -------------------------------------------------------

func TestOldB358ChandraPlusOnePingsAndStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	chandra := oldB358Walker(g, me.ID, "Chandra, Pyromaster", oldB358Chandra, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 3)
	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Error("a creature the first target's controller doesn't control was legal")
	}
	life := opp.Life
	oldB358Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: theirs}}})
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
	if got := effectstest.DamageOn(t, g, theirs); got != 1 {
		t.Errorf("their creature has %d damage, want 1", got)
	}
	effectstest.AssertRestrictions(t, g, theirs, game.CantBlock)
}

func TestOldB358ChandraZeroExilesTheTopCardAndLetsYouPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	chandra := oldB358Walker(g, me.ID, "Chandra, Pyromaster", oldB358Chandra, 4)
	ids := effectstest.SeedLibrary(me, "Top Card")
	oldB358Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	if z := g.FindCardZoneForEffect(ids[0]); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", z)
	}
	if got := effectstest.LoyaltyOf(g, chandra); got != 4 {
		t.Errorf("loyalty = %d, want 4", got)
	}
}

func TestOldB358ChandraUltimateCopiesAnExiledInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	chandra := oldB358Walker(g, me.ID, "Chandra, Pyromaster", oldB358Chandra, 7)
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for i := 0; i < 9; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		}
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	})
	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 0 {
		t.Errorf("library = %d, want 0 after exiling the top ten", me.Library.Size())
	}
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID); p == nil {
		t.Error("no offer to cast the copies is open")
	}
}
