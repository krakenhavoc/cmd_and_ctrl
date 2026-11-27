package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m10_5585_b499_cards_test.go — Magic 2010, batch 499 (#5585). Every card
// is registered Full and each one with rules text is played through the
// engine.

const (
	m10B499InfernoElementalOracle = "287cb63c-d52d-4871-83d1-792a41ae7adc"
	m10B499SphinxAmbassadorOracle = "69e1a166-02f7-43f8-b099-dbb01034370c"
	m10B499HypnoticSpecterOracle  = "759af941-f6a3-4726-91f2-9b1e4e55ea71"
	m10B499IgniteDisorderOracle   = "8f66317c-06fe-4320-9d98-eaedaef67c85"
	m10B499GuardianSeraphOracle   = "fe9106e3-c637-4995-9dc1-cae0be31f23b"
	m10B499YawningFissureOracle   = "634ec9d3-24c9-4090-a087-2624b6d5bf5b"
	m10B499BurstOfSpeedOracle     = "91e33674-b13a-41de-8998-4b57016ae42f"
	m10B499PanicAttackOracle      = "c1346677-d89e-49d0-9819-08805f00bb64"
	m10B499JumpOracle             = "f7518456-45ed-41d4-bd3c-5aacea28eb35"
	m10B499StoneGiantOracle       = "0b8e3f9b-a4da-49a3-8545-ce7a265e5856"
	m10B499WallOfBoneOracle       = "8ffe4986-9e09-4421-ad26-296a4c0df9e4"
	m10B499DrudgeSkeletonsOracle  = "c180ed02-07f4-4538-9bea-8c249234a8e2"

	m10B499BearType = "Creature — Bear"
)

func TestM10B499EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		m10B499InfernoElementalOracle, m10B499SphinxAmbassadorOracle, m10B499HypnoticSpecterOracle,
		m10B499IgniteDisorderOracle, m10B499GuardianSeraphOracle, m10B499YawningFissureOracle,
		m10B499BurstOfSpeedOracle, m10B499PanicAttackOracle, m10B499JumpOracle, m10B499StoneGiantOracle,
		m10B499WallOfBoneOracle, m10B499DrudgeSkeletonsOracle,
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

func TestM10B499KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Sphinx Ambassador", "Creature — Sphinx", m10B499SphinxAmbassadorOracle, "flying"},
		{"Hypnotic Specter", "Creature — Specter", m10B499HypnoticSpecterOracle, "flying"},
		{"Guardian Seraph", "Creature — Angel", m10B499GuardianSeraphOracle, "flying"},
		{"Wall of Bone", "Creature — Skeleton Wall", m10B499WallOfBoneOracle, "defender"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestM10B499InfernoElementalDamagesWhatBlocksOrIsBlockedByIt(t *testing.T) {
	// Becomes blocked.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elemental := effectstest.B12Push(g, me.ID, "Inferno Elemental", "Creature — Elemental", m10B499InfernoElementalOracle, 4, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", m10B499BearType, 1, 9)
	effectstest.BrAttack(t, g, elemental)
	if err := g.DeclareBlocker(blocker, elemental); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, blocker).DamageMarked; d < 3 {
		t.Errorf("the blocker has %d damage, want the trigger's 3", d)
	}

	// Blocks.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", m10B499BearType, 1, 9)
	wall := effectstest.B12Push(g, opp.ID, "Inferno Elemental", "Creature — Elemental", m10B499InfernoElementalOracle, 4, 4)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, attacker).DamageMarked; d < 3 {
		t.Errorf("the attacker has %d damage, want the trigger's 3", d)
	}
}

func TestM10B499HypnoticSpecterMakesTheOpponentDiscardAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	specter := effectstest.B12Push(g, me.ID, "Hypnotic Specter", "Creature — Specter", m10B499HypnoticSpecterOracle, 2, 2)
	effectstest.D2dSetHand(opp, 2)
	effectstest.DealCombatDamageToPlayer(g, specter, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, opp); got != 1 {
		t.Errorf("opponent hand %d, want 1", got)
	}
	// Damage to its own controller is not damage to an opponent.
	effectstest.D2dSetHand(me, 2)
	effectstest.DealCombatDamageToPlayer(g, specter, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != 2 {
		t.Errorf("own hand %d, want 2", got)
	}
}

func TestM10B499IgniteDisorderDividesDamageOverWhiteAndBlueCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	white := effectstest.B16Creature(g, opp.ID, "White", m10B499BearType, 1, 6, "W")
	blue := effectstest.B16Creature(g, opp.ID, "Blue", m10B499BearType, 1, 6, "U")
	green := effectstest.B16Creature(g, opp.ID, "Green", m10B499BearType, 1, 6, "G")
	effectstest.AdvanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	bad := effectstest.B13HandCard(active, "Ignite Disorder", "Instant", m10B499IgniteDisorderOracle)
	if err := g.CastSpell(active.ID, bad, game.CastSpellParams{Targets: effectstest.CardRefs(green), Strict: true}); err == nil {
		t.Fatal("Ignite Disorder targeted a green creature")
	}
	effectstest.B12PlayFromHand(t, g, "Ignite Disorder", "Instant", m10B499IgniteDisorderOracle,
		game.CastSpellParams{Targets: effectstest.CardRefs(white, blue), Distribution: map[uuid.UUID]int{white: 2, blue: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, white).DamageMarked; d != 2 {
		t.Errorf("white creature took %d, want 2", d)
	}
	if d := effectstest.E2Card(t, g, blue).DamageMarked; d != 1 {
		t.Errorf("blue creature took %d, want 1", d)
	}
	if d := effectstest.E2Card(t, g, green).DamageMarked; d != 0 {
		t.Errorf("green creature took %d, want 0", d)
	}
}

func TestM10B499GuardianSeraphPreventsOneDamageFromOpponentSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Guardian Seraph", "Creature — Angel", m10B499GuardianSeraphOracle, 3, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", m10B499BearType, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", m10B499BearType, 3, 3)
	deal := func(src, victim uuid.UUID, n int) {
		t.Helper()
		if err := g.DealDamageToPlayerForEffect(src, victim, n); err != nil {
			t.Fatal(err)
		}
	}
	life := me.Life
	deal(theirs, me.ID, 3)
	if me.Life != life-2 {
		t.Errorf("after 3 from an opponent's creature: life %d, want %d", me.Life, life-2)
	}
	deal(theirs, me.ID, 1)
	if me.Life != life-2 {
		t.Errorf("after 1 from an opponent's creature: life %d, want %d (all prevented)", me.Life, life-2)
	}
	deal(mine, me.ID, 3)
	if me.Life != life-5 {
		t.Errorf("after 3 from my own creature: life %d, want %d", me.Life, life-5)
	}
	oppLife := opp.Life
	deal(theirs, opp.ID, 3)
	if opp.Life != oppLife-3 {
		t.Errorf("damage to the opponent: life %d, want %d", opp.Life, oppLife-3)
	}
}

func m10B499Land(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner,
	})
}

func TestM10B499YawningFissureMakesEachOpponentSacrificeALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := m10B499Land(g, me.ID, "My Forest")
	a, b := m10B499Land(g, opp.ID, "Forest A"), m10B499Land(g, opp.ID, "Forest B")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m10B499BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Yawning Fissure", "Sorcery", m10B499YawningFissureOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster is not an opponent")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(b) || !g.Battlefield.Contains(a) {
		t.Error("the chosen land was not the one sacrificed")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(bear) {
		t.Error("something other than an opponent's land was sacrificed")
	}
}

func TestM10B499BurstOfSpeedGivesYourCreaturesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m10B499BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m10B499BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Burst of Speed", "Sorcery", m10B499BurstOfSpeedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "haste") {
		t.Error("your creature has no haste")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "haste") {
		t.Error("an opponent's creature gained haste")
	}
	later := effectstest.B12Creature(g, me.ID, "Later", m10B499BearType, 2, 2)
	if effectstest.EffectiveAbilitiesContain(t, g, later, "haste") {
		t.Error("a creature that entered afterwards gained haste")
	}
}

func TestM10B499PanicAttackStopsUpToThreeCreaturesBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	attacker := effectstest.B12Creature(g, g.Seats[0].ID, "Attacker", m10B499BearType, 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "A", m10B499BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", m10B499BearType, 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", m10B499BearType, 2, 2)
	d := effectstest.B12Creature(g, opp.ID, "D", m10B499BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Panic Attack", "Sorcery", m10B499PanicAttackOracle, effectstest.CardRefs(a, b, c, d)); err == nil {
		t.Fatal("Panic Attack took four targets")
	}
	effectstest.CastCatalogSpell(t, g, "Panic Attack", "Sorcery", m10B499PanicAttackOracle, effectstest.CardRefs(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, attacker)
	for _, id := range []uuid.UUID{a, b, c} {
		if err := g.DeclareBlocker(id, attacker); err == nil {
			t.Errorf("a targeted creature blocked")
		}
	}
	if err := g.DeclareBlocker(d, attacker); err != nil {
		t.Errorf("the untargeted creature could not block: %v", err)
	}
}

func TestM10B499JumpGivesFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", m10B499BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Jump", "Instant", m10B499JumpOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Fatal("the target has no flying")
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

func TestM10B499StoneGiantFliesThenDiesAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	giant := effectstest.B12Push(g, me.ID, "Stone Giant", "Creature — Giant", m10B499StoneGiantOracle, 3, 4)
	small := effectstest.B12Creature(g, me.ID, "Small", m10B499BearType, 5, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", m10B499BearType, 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m10B499BearType, 1, 1)
	for _, bad := range []uuid.UUID{big, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bad), Strict: true}); err == nil {
			t.Fatal("Stone Giant targeted a creature that isn't yours with toughness below its power")
		}
	}
	effectstest.ApaActivate(t, g, me, giant, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(small)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, small, "flying") {
		t.Error("the target has no flying")
	}
	if !g.Battlefield.Contains(small) {
		t.Fatal("the target died before the end step")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the target survived the end step")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(giant) {
		t.Error("something else was destroyed")
	}
}

func TestM10B499SkeletonsRegenerate(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Wall of Bone", "Creature — Skeleton Wall", m10B499WallOfBoneOracle},
		{"Drudge Skeletons", "Creature — Skeleton", m10B499DrudgeSkeletonsOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			effectstest.ApaMana(me, "B")
			effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if n := effectstest.RegenShieldsOn(t, g, id); n != 1 {
				t.Fatalf("shields = %d, want 1", n)
			}
			effectstest.DestroyIt(t, g, id)
			if !g.Battlefield.Contains(id) {
				t.Error("the shield did not replace the destruction")
			}
		})
	}
}

func m10B499SphinxGame(t *testing.T, libraryCard game.Card) (g *game.Game, me, opp *game.Player, sphinx, found uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	sphinx = effectstest.B12Push(g, me.ID, "Sphinx Ambassador", "Creature — Sphinx", m10B499SphinxAmbassadorOracle, 5, 5)
	libraryCard.Owner, libraryCard.Controller = opp.ID, opp.ID
	found = effectstest.PushLibraryCard(opp, libraryCard)
	effectstest.DealCombatDamageToPlayer(g, sphinx, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, found)
	return
}

func TestM10B499SphinxAmbassadorTakesACreatureTheOpponentDidNotName(t *testing.T) {
	g, me, opp, _, found := m10B499SphinxGame(t, game.Card{InstanceID: uuid.New(), Name: "Prize Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	c := effectstest.PendingOfKind(g, game.PendingChoiceCardName)
	if c == nil || c.Chooser != opp.ID {
		t.Fatalf("no card-name prompt for the searched player: %+v", g.PendingChoices)
	}
	if err := g.ResolveCardNameChoice(c.ID, opp.ID, "Some Other Card"); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil || pick.Chooser != me.ID {
		t.Fatalf("no may-put prompt for the searcher: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{found}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card := effectstest.E2Card(t, g, found)
	if !g.Battlefield.Contains(found) || card.Controller != me.ID || card.Owner != opp.ID {
		t.Errorf("the creature is on the battlefield=%v under %v, want under the Sphinx's controller", g.Battlefield.Contains(found), card.Controller)
	}
}

func TestM10B499SphinxAmbassadorLeavesTheCardWhenItsNameIsGuessed(t *testing.T) {
	g, me, opp, _, found := m10B499SphinxGame(t, game.Card{InstanceID: uuid.New(), Name: "Prize Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	c := effectstest.PendingOfKind(g, game.PendingChoiceCardName)
	if c == nil {
		t.Fatal("no card-name prompt")
	}
	if err := g.ResolveCardNameChoice(c.ID, opp.ID, "Prize Bear"); err != nil {
		t.Fatal(err)
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceChooseCards) != nil {
		t.Error("the searcher was offered a card whose name was named")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(found) || !opp.Library.Contains(found) {
		t.Error("the named card left the library")
	}
	_ = me
}

func TestM10B499SphinxAmbassadorLeavesANoncreatureCard(t *testing.T) {
	g, _, opp, _, found := m10B499SphinxGame(t, game.Card{InstanceID: uuid.New(), Name: "Plain Rock", TypeLine: "Artifact"})
	c := effectstest.PendingOfKind(g, game.PendingChoiceCardName)
	if c == nil {
		t.Fatal("no card-name prompt")
	}
	if err := g.ResolveCardNameChoice(c.ID, opp.ID, "Something Else"); err != nil {
		t.Fatal(err)
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceChooseCards) != nil {
		t.Error("a noncreature card was offered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(found) || !opp.Library.Contains(found) {
		t.Error("the noncreature card left the library")
	}
}
