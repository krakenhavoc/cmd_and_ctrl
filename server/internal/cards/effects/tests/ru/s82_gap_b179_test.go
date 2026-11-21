package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b179_test.go — GAP batch 179 (#5948): Commander Starter Decks
// artifacts and creatures. Every test plays the real catalog entry.

const (
	gapB179Grimoire   = "922c6e1b-abcf-40e0-847e-9ebb7d0b6fd4"
	gapB179Atarka     = "eb287c04-97f5-4e47-beb1-66cc7d15e748"
	gapB179Debaser    = "db286eef-f0df-4f7c-8e06-957447166950"
	gapB179Specter    = "529ace49-c50e-4d3e-ac29-91e08af57fdc"
	gapB179Wretch     = "29e13ec1-a239-4217-a2bd-24ff90745ec0"
	gapB179Vela       = "185e34f0-60ae-4b07-94ac-96cef11b1604"
	gapB179FoeRazer   = "e7dd3e2c-34a5-4097-b6d1-2ccd097fe57f"
	gapB179Nullmage   = "4dcfd397-9d29-420c-8fd8-c7c2a99de122"
	gapB179CreatureTy = "Creature — Bear"
)

func TestGapB179EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		gapB179Grimoire, gapB179Atarka, gapB179Debaser, gapB179Specter,
		gapB179Wretch, gapB179Vela, gapB179FoeRazer, gapB179Nullmage,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- Grimoire of the Dead -------------------------------------------------

func TestGapB179GrimoireDiscardsForAStudyCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	book := effectstest.B12Push(g, me.ID, "Grimoire of the Dead", "Legendary Artifact — Book", gapB179Grimoire, 0, 0)
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err == nil {
		t.Fatal("activated without paying {1}")
	}
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard cost was not paid")
	}
	if got := effectstest.CountersOn(g, book, "study"); got != 1 {
		t.Errorf("study counters %d, want 1", got)
	}
}

func TestGapB179GrimoireNeedsThreeCountersToReturnEveryCreatureAsZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	book := effectstest.B12Push(g, me.ID, "Grimoire of the Dead", "Legendary Artifact — Book", gapB179Grimoire, 0, 0)
	effectstest.StdB9AddCounter(t, g, book, "study", 2)
	if err := g.ActivateCatalogAbility(me.ID, book, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only two study counters")
	}
	effectstest.StdB9AddCounter(t, g, book, "study", 1)
	mine := effectstest.PushGraveyardCardTyped(me, "My Bear", gapB179CreatureTy)
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Bear", gapB179CreatureTy)
	spell := effectstest.PushGraveyardCardTyped(opp, "A Spell", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, book, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(book) {
		t.Error("the Grimoire was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatal("a creature card stayed in its graveyard")
		}
		if c.Controller != me.ID {
			t.Error("the creature did not come back under my control")
		}
		view := effectstest.S58p6View(t, g, id)
		if !slices.Contains(view.Effective().Subtypes, "Zombie") || !slices.Contains(view.EffectiveColors(), "B") {
			t.Errorf("subtypes %v colours %v, want a black Zombie", view.Effective().Subtypes, view.EffectiveColors())
		}
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("a noncreature card was returned")
	}
}

// --- Atarka Monument ------------------------------------------------------

func TestGapB179AtarkaMonumentTapsForRedOrGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := effectstest.S58p6Push(g, me.ID, "Atarka Monument", "Artifact", gapB179Atarka, "{3}", 0, 0)
	for _, colour := range []string{"R", "G"} {
		me.ManaPool.EmptyPool()
		effectstest.FindBattlefieldCardByID(g, monument).Tapped = false
		if err := g.ActivateManaAbility(me.ID, monument, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("tap for %s: %v", colour, err)
		}
		if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != colour {
			t.Errorf("pool = %v, want one %s", me.ManaPool.Units(), colour)
		}
	}
}

func TestGapB179AtarkaMonumentBecomesAFourFourFlyingDragonUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := effectstest.S58p6Push(g, me.ID, "Atarka Monument", "Artifact", gapB179Atarka, "{3}", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.ApaMana(me, "C", "C", "C", "C", "R", "G")
	effectstest.ApaActivate(t, g, me, monument, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, monument)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("creature=%v artifact=%v, want both", c.IsCreature(), c.IsArtifact())
	}
	if !slices.Contains(c.Effective().Subtypes, "Dragon") {
		t.Errorf("subtypes %v, want Dragon", c.Effective().Subtypes)
	}
	colours := c.EffectiveColors()
	if !slices.Contains(colours, "R") || !slices.Contains(colours, "G") || len(colours) != 2 {
		t.Errorf("colours %v, want red and green", colours)
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("%d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, monument, "flying") {
		t.Error("no flying")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	c = effectstest.S58p6View(t, g, monument)
	if c.IsCreature() || !c.IsArtifact() {
		t.Errorf("after the turn: creature=%v artifact=%v, want an artifact only", c.IsCreature(), c.IsArtifact())
	}
}

// --- Withered Wretch ------------------------------------------------------

func TestGapB179WitheredWretchExilesAnyCardFromAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	wretch := effectstest.B12Push(g, me.ID, "Withered Wretch", "Creature — Zombie Cleric", gapB179Wretch, 2, 2)
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Card", "Sorcery")
	mine := effectstest.PushGraveyardCardTyped(me, "My Card", "Instant")
	for _, id := range []uuid.UUID{theirs, mine} {
		effectstest.ApaMana(me, "C")
		effectstest.ApaActivate(t, g, me, wretch, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(id)})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.P1WantZone(t, g, id, game.ZoneExile, "the targeted graveyard card")
	}
}

func TestGapB179WitheredWretchCannotTargetABattlefieldCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	wretch := effectstest.B12Push(g, me.ID, "Withered Wretch", "Creature — Zombie Cleric", gapB179Wretch, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Bear", gapB179CreatureTy, 2, 2)
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, wretch, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(other)}); err == nil {
		t.Error("the Wretch targeted a permanent")
	}
}

// --- Nullmage Shepherd ----------------------------------------------------

func TestGapB179NullmageShepherdTapsFourCreaturesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	for _, typeLine := range []string{"Artifact", "Enchantment"} {
		t.Run(typeLine, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			shepherd := effectstest.B12Push(g, me.ID, "Nullmage Shepherd", "Creature — Elf Shaman", gapB179Nullmage, 2, 4)
			a := effectstest.B12Creature(g, me.ID, "A", gapB179CreatureTy, 1, 1)
			b := effectstest.B12Creature(g, me.ID, "B", gapB179CreatureTy, 1, 1)
			c := effectstest.B12Creature(g, me.ID, "C", gapB179CreatureTy, 1, 1)
			victim := effectstest.B12Permanent(g, opp.ID, "Victim", typeLine)
			effectstest.ApaActivate(t, g, me, shepherd, 0, game.ActivateAbilityParams{
				Targets: effectstest.CardTarget(victim), TapIDs: []uuid.UUID{shepherd, a, b, c},
			})
			effectstest.PassPriorityAroundTable(t, g)
			for _, id := range []uuid.UUID{shepherd, a, b, c} {
				if !effectstest.TappedOnBattlefield(t, g, id) {
					t.Error("a creature was not tapped for the cost")
				}
			}
			if g.Battlefield.Contains(victim) {
				t.Errorf("the %s survived", typeLine)
			}
		})
	}
}

func TestGapB179NullmageShepherdNeedsFourCreaturesAndAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	shepherd := effectstest.B12Push(g, me.ID, "Nullmage Shepherd", "Creature — Elf Shaman", gapB179Nullmage, 2, 4)
	a := effectstest.B12Creature(g, me.ID, "A", gapB179CreatureTy, 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", gapB179CreatureTy, 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, shepherd, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTarget(rock), TapIDs: []uuid.UUID{shepherd, a, b},
	}); err == nil {
		t.Error("activated with only three creatures tapped")
	}
	c := effectstest.B12Creature(g, me.ID, "C", gapB179CreatureTy, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", gapB179CreatureTy, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, shepherd, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTarget(bear), TapIDs: []uuid.UUID{shepherd, a, b, c},
	}); err == nil {
		t.Error("the Shepherd targeted a plain creature")
	}
}

// --- Vela the Night-Clad --------------------------------------------------

func TestGapB179VelaGivesOtherCreaturesIntimidate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	vela := effectstest.B12Push(g, me.ID, "Vela the Night-Clad", "Legendary Creature — Human Wizard", gapB179Vela, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gapB179CreatureTy, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", gapB179CreatureTy, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "intimidate") {
		t.Error("my other creature lacks intimidate")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "intimidate") {
		t.Error("an opponent's creature has intimidate")
	}
	_ = vela
}

func TestGapB179VelaDrainsEachOpponentWhenACreatureYouControlLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	vela := effectstest.B12Push(g, me.ID, "Vela the Night-Clad", "Legendary Creature — Human Wizard", gapB179Vela, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", gapB179CreatureTy, 2, 2)
	theirs := effectstest.B12Creature(g, effectstest.Top100Opponent(g).ID, "Theirs", gapB179CreatureTy, 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.StdB9Kill(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d after one of my creatures left", p.Name, p.Life, want)
		}
	}
	effectstest.StdB9Kill(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d: an opponent's creature leaving must not drain", p.Name, p.Life, want)
		}
	}
	effectstest.StdB9Kill(t, g, vela)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if p.Life != lives[p.ID]-2 {
			t.Errorf("%s life %d, want %d after Vela herself left", p.Name, p.Life, lives[p.ID]-2)
		}
	}
}

// --- Foe-Razer Regent -----------------------------------------------------

func TestGapB179FoeRazerRegentFightsOnEntryAndGrowsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", gapB179CreatureTy, 2, 2)
	effectstest.AdvanceToMain(t, g)
	regent := effectstest.CastHoldCreature(t, g, "Foe-Razer Regent", "Creature — Dragon", gapB179FoeRazer, 4, 5)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived a fight with a 4/5")
	}
	if got := effectstest.CountersOn(g, regent, game.CounterPlusOne); got != 0 {
		t.Fatalf("%d counters before the end step, want 0", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, regent, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters at the end step, want 2", got)
	}
}

// --- Rakshasa Debaser -----------------------------------------------------

func TestGapB179RakshasaDebaserTakesACreatureFromTheDefendersGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	debaser := effectstest.B12Push(g, me.ID, "Rakshasa Debaser", "Creature — Demon", gapB179Debaser, 6, 6)
	defenders := effectstest.PushGraveyardCardTyped(opp, "Defender's Bear", gapB179CreatureTy)
	elsewhere := effectstest.PushGraveyardCardTyped(other, "Other Bear", gapB179CreatureTy)
	spell := effectstest.PushGraveyardCardTyped(opp, "A Spell", "Sorcery")
	effectstest.DeclareAttack(t, g, opp.ID, debaser)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, elsewhere) || effectstest.HasID(p.PickTargetCards, spell) {
		t.Errorf("offered %v: only creature cards in the defending player's graveyard may be picked", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, defenders)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, defenders)
	if !ok || c.Controller != me.ID {
		t.Fatal("the creature card did not enter under my control")
	}
	if !other.Graveyard.Contains(elsewhere) {
		t.Error("another player's graveyard was touched")
	}
}

func TestGapB179RakshasaDebaserHasEncore(t *testing.T) {
	spec, _ := effects.Lookup(gapB179Debaser)
	found := false
	for _, a := range spec.Activated {
		if a.Label != "" && len(a.Label) >= 6 && a.Label[:6] == "Encore" {
			found = true
		}
	}
	if !found {
		t.Error("no Encore ability")
	}
}

// --- Scythe Specter -------------------------------------------------------

func TestGapB179ScytheSpecterDrainsTheGreatestDiscardedManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range []*game.Player{a, b, c} {
		p.Hand.Cards = p.Hand.Cards[:0]
	}
	five := uuid.New()
	a.Hand.PushTop(game.Card{InstanceID: five, Name: "Five", TypeLine: "Sorcery", ManaCost: "{4}{R}", Owner: a.ID, Controller: a.ID})
	alsoFive := uuid.New()
	b.Hand.PushTop(game.Card{InstanceID: alsoFive, Name: "Also Five", TypeLine: "Sorcery", ManaCost: "{3}{U}{U}", Owner: b.ID, Controller: b.ID})
	two := uuid.New()
	c.Hand.PushTop(game.Card{InstanceID: two, Name: "Two", TypeLine: "Sorcery", ManaCost: "{1}{G}", Owner: c.ID, Controller: c.ID})
	specter := effectstest.B12Push(g, me.ID, "Scythe Specter", "Creature — Specter", gapB179Specter, 4, 4)
	lives := map[uuid.UUID]int{a.ID: a.Life, b.ID: b.Life, c.ID: c.Life}
	effectstest.BrAttack(t, g, specter)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, a.ID, five)
	effectstest.AnswerDiscard(t, g, b.ID, alsoFive)
	effectstest.AnswerDiscard(t, g, c.ID, two)
	effectstest.PassPriorityAroundTable(t, g)
	// The attack hit seat 1 for 4, and every player that discarded a five loses 5.
	if want := lives[a.ID] - 4 - 5; a.Life != want {
		t.Errorf("the attacked player has %d life, want %d", a.Life, want)
	}
	if want := lives[b.ID] - 5; b.Life != want {
		t.Errorf("the second five-drop discarder has %d life, want %d", b.Life, want)
	}
	if c.Life != lives[c.ID] {
		t.Errorf("the player who discarded the cheaper card lost %d life", lives[c.ID]-c.Life)
	}
}
