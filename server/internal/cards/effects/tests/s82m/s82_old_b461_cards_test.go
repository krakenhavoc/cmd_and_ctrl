package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b461_cards_test.go — New Phyrexia, batch 461 (S82, #5575).
// Each test plays the real catalog entry through the engine.

const (
	b461Forge     = "0aedf049-db6b-420e-80e7-591ed38d4fba"
	b461WingSp    = "7fbf6afc-8a16-447b-bdef-b7d0b640f76c"
	b461MaulSp    = "b0d0c534-8bd2-4801-aadd-dcc9c92d6ec3"
	b461VitalSp   = "e1899133-635d-46ff-b90d-6f8d46a1ffe1"
	b461Finale    = "0c1d5e6b-f6b6-4dc6-9bc1-3cc00b78da36"
	b461Zealot    = "30c7e304-ecfe-44c1-88d1-f507df63c080"
	b461Exclusion = "5e3fd5da-6724-4134-ab15-4616cb659f88"
	b461Invader   = "fe4c0dc1-f6a9-4828-9c19-4239c4f955a7"
	b461Respect   = "1c367fa2-844a-4083-9484-c6def1dee8a8"
	b461Mutation  = "652ef093-06e9-4bf1-b9cf-950fe043b319"
	b461Shrike    = "db98ee99-d4e5-4623-8446-5cce5d11dca0"
)

func TestB461EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b461Forge: "Chancellor of the Forge", b461WingSp: "Wing Splicer", b461MaulSp: "Maul Splicer",
		b461VitalSp: "Vital Splicer", b461Finale: "Life's Finale", b461Zealot: "Blind Zealot",
		b461Exclusion: "Exclusion Ritual", b461Invader: "Invader Parasite", b461Respect: "Due Respect",
		b461Mutation: "Argent Mutation", b461Shrike: "Impaler Shrike",
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

func b461Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func b461Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func b461MustCast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id, err := b461Cast(t, g, me, name, typeLine, oracle, targets)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func TestB461ChancellorOfTheForgeMakesGoblinsForEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b461Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b461MustCast(t, g, me, "Chancellor of the Forge", "Creature — Phyrexian Giant", b461Forge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Bears and the Chancellor are three creatures as the trigger resolves.
	if got := effectstest.B27CountNamed(g, me.ID, "Phyrexian Goblin"); got != 3 {
		t.Errorf("Goblins %d, want 3", got)
	}
}

func TestB461ChancellorOfTheForgeRevealedMakesAGoblinAtTheFirstUpkeep(t *testing.T) {
	g := effectstest.OpeningTable(t)
	starter := g.Seats[g.StartingSeat]
	effectstest.HandWith(g, g.StartingSeat,
		game.Card{Name: "Chancellor of the Forge", TypeLine: "Creature — Phyrexian Giant", ManaCost: "{4}{R}{R}{R}", OracleID: b461Forge, Power: 5, Toughness: 5},
		effectstest.FillerCard("Spare"))
	effectstest.KeepAll(t, g)
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceConfirm && c.Chooser == starter.ID {
			offer = c
		}
	}
	if offer == nil {
		t.Fatal("not offered the reveal")
	}
	if err := g.ResolveConfirm(offer.ID, starter.ID, true); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, starter.ID, "Phyrexian Goblin"); got != 1 {
		t.Errorf("Goblins %d, want 1", got)
	}
}

func TestB461SplicersMakeGolemsAndGrantOnlyYourGolemsTheirKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, kw string
		golems                     int
	}{
		{"Wing Splicer", "Creature — Phyrexian Human Artificer", b461WingSp, "flying", 1},
		{"Maul Splicer", "Creature — Phyrexian Human Artificer", b461MaulSp, "trample", 2},
		{"Vital Splicer", "Creature — Phyrexian Human Artificer", b461VitalSp, "", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := b461Seats(g)
			effectstest.AdvanceToMain(t, g)
			theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			b461MustCast(t, g, me, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.B27CountNamed(g, me.ID, "Phyrexian Golem"); got != tc.golems {
				t.Fatalf("Golems %d, want %d", got, tc.golems)
			}
			if tc.kw == "" {
				return
			}
			var golem uuid.UUID
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.Name == "Phyrexian Golem" && c.Controller == me.ID {
						golem = c.InstanceID
					}
				}
			})
			if !effectstest.HasEffectiveKeyword(t, g, golem, tc.kw) {
				t.Errorf("your Golem lacks %s", tc.kw)
			}
			if effectstest.HasEffectiveKeyword(t, g, theirs, tc.kw) || effectstest.HasEffectiveKeyword(t, g, bear, tc.kw) {
				t.Errorf("a creature that is not your Golem has %s", tc.kw)
			}
		})
	}
}

func TestB461VitalSplicerRegeneratesOnlyYourGolems(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b461Seats(g)
	effectstest.AdvanceToMain(t, g)
	vs := effectstest.B12Push(g, me.ID, "Vital Splicer", "Creature — Phyrexian Human Artificer", b461VitalSp, 1, 1)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	for _, bad := range []uuid.UUID{bear, theirs} {
		if err := effectstest.RavB2Activate(t, g, me.ID, vs, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bad)}, "C"); err == nil {
			t.Fatal("it targeted something other than a Golem you control")
		}
	}
	effectstest.RavB2MustActivate(t, g, me.ID, vs, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(golem)}, "C")
	effectstest.P1Destroy(t, g, golem)
	if effectstest.FindBattlefieldCardForTest(g, golem) == nil {
		t.Error("the regenerated Golem died")
	}
}

func TestB461LifesFinaleDestroysAllThenBinsThreeCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b461Seats(g)
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	var cre []uuid.UUID
	for i := 0; i < 4; i++ {
		cre = append(cre, effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Library Beast", TypeLine: "Creature — Beast", Owner: opp.ID, Controller: opp.ID}))
	}
	land := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Library Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
	if _, err := b461Cast(t, g, me, "Life's Finale", "Sorcery", b461Finale, effectstest.PlayerRefs(me.ID)); err == nil {
		t.Fatal("it targeted the caster, not an opponent")
	}
	b461MustCast(t, g, me, "Life's Finale", "Sorcery", b461Finale, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.FindBattlefieldCardForTest(g, mine) != nil || effectstest.FindBattlefieldCardForTest(g, theirs) != nil {
		t.Error("a creature survived")
	}
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{cre[0], cre[1], cre[2], cre[3]}); err == nil {
		t.Error("four cards accepted, want at most three")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{cre[0], land}); err == nil {
		t.Error("a land was accepted")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{cre[0], cre[1], cre[2]}); err != nil {
		t.Fatalf("search: %v", err)
	}
	for _, id := range cre[:3] {
		if !effectstest.InGraveyard(opp, id) {
			t.Errorf("a found creature is not in their graveyard")
		}
	}
	if effectstest.InGraveyard(opp, cre[3]) {
		t.Error("a fourth creature was binned")
	}
}

func b461Hit(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	effectstest.BrAttack(t, g, id)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
}

func TestB461BlindZealotMaySacrificeToDestroyTheirCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		other := g.Seats[2]
		zealot := effectstest.B12Push(g, me.ID, "Blind Zealot", "Creature — Phyrexian Human Cleric", b461Zealot, 2, 2)
		victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
		elsewhere := effectstest.B12Creature(g, other.ID, "Elsewhere", "Creature — Ogre", 4, 4)
		b461Hit(t, g, me, zealot)
		effectstest.B04WaitForPick(t, g, me.ID)
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil || effectstest.HasID(p.PickTargetCards, elsewhere) || !effectstest.HasID(p.PickTargetCards, victim) {
			t.Fatalf("pick offered %+v: only the damaged player's creatures", p)
		}
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		gone := effectstest.FindBattlefieldCardForTest(g, victim) == nil
		if gone != yes {
			t.Errorf("yes=%v: victim gone=%v", yes, gone)
		}
		if (effectstest.FindBattlefieldCardForTest(g, zealot) == nil) != yes {
			t.Errorf("yes=%v: the Zealot's presence is wrong", yes)
		}
		if effectstest.FindBattlefieldCardForTest(g, elsewhere) == nil {
			t.Error("another player's creature was destroyed")
		}
	}
}

func TestB461ImpalerShrikeMaySacrificeToDrawThree(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		shrike := effectstest.B12Push(g, me.ID, "Impaler Shrike", "Creature — Phyrexian Bird", b461Shrike, 3, 1)
		hand := len(me.Hand.Cards)
		b461Hit(t, g, me, shrike)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		want := hand
		if yes {
			want += 3
		}
		if got := len(me.Hand.Cards); got != want {
			t.Errorf("yes=%v: hand %d, want %d", yes, got, want)
		}
		if (effectstest.FindBattlefieldCardForTest(g, shrike) == nil) != yes {
			t.Errorf("yes=%v: the Shrike's presence is wrong", yes)
		}
	}
}

func TestB461ExclusionRitualBarsTheExiledCardsNameForEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b461Seats(g)
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Push(g, opp.ID, "Sol Ring", "Artifact", "", 0, 0)
	land := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	ritual := b461MustCast(t, g, me, "Exclusion Ritual", "Enchantment", b461Exclusion, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, land) {
		t.Error("a land was offered for 'target nonland permanent'")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, rock) {
		t.Fatal("the permanent was not exiled")
	}
	for _, p := range []*game.Player{me, opp} {
		if _, err := b461Cast(t, g, p, "Sol Ring", "Instant", "", nil); err == nil {
			t.Errorf("%s cast a spell named like the exiled card", p.Name)
		}
	}
	if _, err := b461Cast(t, g, me, "Mind Stone", "Instant", "", nil); err != nil {
		t.Errorf("an unrelated spell was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B14Kill(g, ritual)
	if _, err := b461Cast(t, g, me, "Sol Ring", "Instant", "", nil); err != nil {
		t.Errorf("the bar outlived the Ritual: %v", err)
	}
}

func TestB461InvaderParasiteBurnsAnOpponentForALandWithTheExiledName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b461Seats(g)
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	b461MustCast(t, g, me, "Invader Parasite", "Creature — Phyrexian Insect", b461Invader, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, forest) {
		t.Fatal("the land was not exiled")
	}
	life := opp.Life
	effectstest.B34LandEntersByEffect(t, g, opp, "Island", "Basic Land — Island")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a differently named land cost %d life", life-opp.Life)
	}
	effectstest.B34LandEntersByEffect(t, g, opp, "Forest", "Basic Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	effectstest.B34LandEntersByEffect(t, g, me, "Forest", "Basic Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != g.Settings.StartingLife {
		t.Error("the Parasite's own controller's land hurt them")
	}
}

func TestB461DueRespectEntersEverythingTappedAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b461Seats(g)
	effectstest.AdvanceToMain(t, g)
	hand := len(me.Hand.Cards)
	effectstest.ApaMana(me, "W", "C")
	b461MustCast(t, g, me, "Due Respect", "Instant", b461Respect, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	mine := effectstest.B34LandEntersByEffect(t, g, me, "Forest", "Basic Land — Forest")
	theirs := effectstest.B34LandEntersByEffect(t, g, opp, "Island", "Basic Land — Island")
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.TappedOnBattlefield(t, g, id) {
			t.Error("a permanent entered untapped")
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	later := effectstest.B34LandEntersByEffect(t, g, opp, "Forest", "Basic Land — Forest")
	if effectstest.TappedOnBattlefield(t, g, later) {
		t.Error("the effect lasted past the turn")
	}
}

func TestB461ArgentMutationMakesAnArtifactAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b461Seats(g)
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	hand := len(me.Hand.Cards)
	effectstest.ApaMana(me, "U", "C", "C")
	b461MustCast(t, g, me, "Argent Mutation", "Instant", b461Mutation, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, land)
	if !c.HasCardType("artifact") || !c.HasCardType("land") {
		t.Errorf("types %q, want an artifact land", c.TypeLine)
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}
