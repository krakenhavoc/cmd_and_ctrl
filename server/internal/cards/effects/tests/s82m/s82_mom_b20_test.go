package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b20_test.go — S82 March of the Machine Commander batch 20: the
// team-up legends of combat, sacrifice and recursion.

const (
	s82ThaliaGitrogOracle = "422a76e1-97a2-4a90-b07d-f487a2d8c236"
	s82KroxaOracle        = "56679f06-982b-4f33-9e48-ecc64c802393"
	s82SlimefootOracle    = "5a6f4181-f5ad-4ecf-9566-e52870af14d0"
	s82KoglaOracle        = "fe8c2b7e-f80d-4acb-a976-edcf436a78a9"
	s82GhaltaOracle       = "756042c5-e444-45d7-8dc6-ae7c83c98438"
	s82RankleOracle       = "80d918b2-3a28-49ad-a485-496658bd7ac3"
	s82ZurgoOracle        = "941c9acd-2773-456a-88d5-0537735f93c6"
	s82GoroGoroOracle     = "854bdd4a-d549-41d5-af43-0e2154ed017e"
	s82MoiraOracle        = "445ad01c-c899-424d-8666-ee7bb2a6e535"
	s82BorborygmosOracle  = "22d9eef5-0f86-4d1a-bde0-65b990d4be75"
)

func s82Fill(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.PushCatalogGraveyardCard(p, "Filler", "Sorcery", "", 0, 0))
	}
	return ids
}

func TestS82ThaliaAndGitrogEntersOpponentsThingsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thalia := effectstest.B12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	abilities := effectstest.EffectiveAbilities(t, g, thalia)
	for _, kw := range []string{"first strike", "deathtouch"} {
		if !effectstest.ContainsString(abilities, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	mine := effectstest.B12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B13Tapped(t, g, mine) {
		t.Error("its controller's creatures are untouched")
	}
	if effectstest.B13Tapped(t, g, effectstest.B13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")) {
		t.Error("an opponent's basic land enters untapped")
	}
	if !effectstest.B13Tapped(t, g, effectstest.B13PlayAs(t, g, 2, "Ancient Tomb", "Land", "")) {
		t.Error("an opponent's nonbasic land enters tapped")
	}
	theirs := effectstest.B13PlayAs(t, g, 3, "Their Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
}

func TestS82ThaliaAndGitrogAllowsAnExtraLandPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	effectstest.AangAdvanceToMain(t, g, 0)
	for i := 0; i < 2; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("land play %d: %v", i+1, err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a third land play was allowed")
	}
}

func TestS82ThaliaAndGitrogAttackSacrificesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thalia := effectstest.B12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 1, 1)
	before := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, thalia)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, snack)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snack) {
		t.Error("the chosen creature survived")
	}
	if !g.Battlefield.Contains(other) || !g.Battlefield.Contains(thalia) {
		t.Error("an unchosen permanent was sacrificed")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d: one card drawn", got, before+1)
	}
}

func TestS82ThaliaAndGitrogMaySacrificeALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thalia := effectstest.B12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	forest := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	before := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, thalia)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the land was not sacrificed")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestS82KroxaExilesFiveToReturnACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	filler := s82Fill(me, 5)
	bear := effectstest.PushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastRingCreature(t, g, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, filler...)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return")
	}
	for _, id := range filler {
		if me.Graveyard.Contains(id) {
			t.Error("a chosen card stayed in the graveyard")
		}
	}
}

func TestS82KroxaDecliningOrTooFewDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s82Fill(me, 3)
	bear := effectstest.PushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastRingCreature(t, g, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked to exile five with only four cards")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("returned a creature without paying")
	}
	s82Fill(me, 2)
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	s82Fill(me2, 5)
	bear2 := effectstest.PushCatalogGraveyardCard(me2, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastRingCreature(t, g2, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerChooseCards(t, g2, me2.ID)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(bear2) || me2.Graveyard.Size() != 6 {
		t.Errorf("declining changed something: graveyard %d", me2.Graveyard.Size())
	}
}

func TestS82KroxaFiresOnAttackToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kroxa := effectstest.B12Push(g, me.ID, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	filler := s82Fill(me, 5)
	bear := effectstest.PushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, kroxa)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, filler...)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the attack trigger did not return the creature")
	}
}

func TestS82SlimefootMakesASaprolingAndReturnsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slime := effectstest.PushCatalogGraveyardCard(me, "Slimefoot and Squee", "Legendary Creature — Fungus Goblin", s82SlimefootOracle, 3, 3)
	bear := effectstest.PushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	land := effectstest.PushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	sap := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 green Saproling"))
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "B", "R", "G")
	// A noncreature card is no target; the ability is sorcery speed.
	if err := g.ActivateCatalogAbility(me.ID, slime, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{sap}, Targets: effectstest.B16TargetCard(land), Strict: true}); err == nil {
		t.Error("a land card was accepted as the other creature card")
	}
	effectstest.B16Activate(t, g, me.ID, slime, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sap}, Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(slime) || !g.Battlefield.Contains(bear) {
		t.Fatalf("both creatures should return: slime=%v bear=%v", g.Battlefield.Contains(slime), g.Battlefield.Contains(bear))
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Saproling" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Saprolings: the sacrificed one is gone and the returned card's enters trigger made one, want 1", n)
	}
}

func TestS82SlimefootCanReturnAloneAndNeedsASaproling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slime := effectstest.PushCatalogGraveyardCard(me, "Slimefoot and Squee", "Legendary Creature — Fungus Goblin", s82SlimefootOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "B", "R", "G")
	if err := g.ActivateCatalogAbility(me.ID, slime, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no Saproling to sacrifice")
	}
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	if err := g.ActivateCatalogAbility(me.ID, slime, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Error("a non-Saproling paid the cost")
	}
	sap := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 green Saproling"))
	effectstest.B16Activate(t, g, me.ID, slime, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sap}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(slime) {
		t.Error("it did not return alone")
	}
}

func TestS82KoglaFightsOrGainsTrampleAndHaste(t *testing.T) {
	t.Run("fight", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
		kogla := effectstest.CastRingCreature(t, g, "Kogla and Yidaro", "Legendary Creature — Ape Dinosaur Turtle", s82KoglaOracle, 7, 7)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		effectstest.AnswerPickTarget(t, g, theirs)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) {
			t.Error("the target survived 7 damage")
		}
		if got := effectstest.DamageOn(t, g, kogla); got != 3 {
			t.Errorf("Kogla has %d damage, want 3", got)
		}
	})
	t.Run("keywords", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		kogla := effectstest.CastRingCreature(t, g, "Kogla and Yidaro", "Legendary Creature — Ape Dinosaur Turtle", s82KoglaOracle, 7, 7)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		abilities := effectstest.EffectiveAbilities(t, g, kogla)
		if !effectstest.ContainsString(abilities, "trample") || !effectstest.ContainsString(abilities, "haste") {
			t.Errorf("abilities %v lack trample and haste", abilities)
		}
	})
}

func TestS82KoglaDiscardsToDestroyShuffleAndDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	kid := effectstest.PutInHand(me, game.Card{Name: "Kogla and Yidaro", TypeLine: "Legendary Creature — Ape Dinosaur Turtle",
		ManaCost: "{2}{R}{R}{G}{G}", OracleID: s82KoglaOracle, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	effectstest.B06AddMana(me, "C", "C", "R", "G")
	effectstest.B16Activate(t, g, me.ID, kid, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if me.Graveyard.Contains(kid) {
		t.Error("Kogla should have left the graveyard")
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library %d, want %d: Kogla shuffled in, one card drawn", me.Library.Size(), libBefore)
	}
	// Hand: Kogla left (-1), one card drawn (+1).
	if got := me.Hand.Size(); got != handBefore-1+1 {
		t.Errorf("hand %d, want %d", got, handBefore)
	}
}

func s82Treasures(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Treasure" {
			n++
		}
	}
	return n
}

func s82NamedOnBattlefield(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestS82GhaltaAndMavrenMakesADinosaurAsBigAsTheBiggestOtherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghalta := effectstest.B12Push(g, me.ID, "Ghalta and Mavren", "Legendary Creature — Dinosaur Vampire", s82GhaltaOracle, 12, 12)
	big := effectstest.PushVanillaCreature(g, me.ID, "Big", 5, 5)
	small := effectstest.PushVanillaCreature(g, me.ID, "Small", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, ghalta, big, small)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.TriggersOnStackFrom(g, ghalta); got != 1 {
		t.Fatalf("%d triggers for one declaration, want 1", got)
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Fatal("a second mode prompt: the declaration triggered it twice")
	}
	if err := func() error { return nil }(); err != nil {
		t.Fatal(err)
	}
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	dinos := s82NamedOnBattlefield(g, me.ID, "Dinosaur")
	if len(dinos) != 1 {
		t.Fatalf("%d Dinosaur tokens, want 1", len(dinos))
	}
	d := dinos[0]
	if p, tt := effectstest.PtOf(t, g, d.InstanceID); p != 5 || tt != 5 {
		t.Errorf("Dinosaur is %d/%d, want 5/5 (Ghalta's 12 is not counted)", p, tt)
	}
	if !d.Tapped || d.AttackingTarget != opp.ID {
		t.Errorf("Dinosaur tapped=%v attacking=%v, want tapped and attacking the opponent", d.Tapped, d.AttackingTarget)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, d.InstanceID), "trample") {
		t.Error("Dinosaur lacks trample")
	}
}

func TestS82GhaltaAndMavrenMakesAVampirePerOtherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghalta := effectstest.B12Push(g, me.ID, "Ghalta and Mavren", "Legendary Creature — Dinosaur Vampire", s82GhaltaOracle, 12, 12)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, ghalta, a, b)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	vamps := s82NamedOnBattlefield(g, me.ID, "Vampire")
	if len(vamps) != 2 {
		t.Fatalf("%d Vampires, want 2", len(vamps))
	}
	for _, v := range vamps {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, v.InstanceID), "lifelink") || v.AttackingTarget != uuid.Nil {
			t.Errorf("a Vampire lacks lifelink or is attacking: %s", v.Name)
		}
	}
}

func TestS82GhaltaAndMavrenTriggersWhenOthersAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ghalta and Mavren", "Legendary Creature — Dinosaur Vampire", s82GhaltaOracle, 12, 12)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) == nil {
		t.Fatal("no mode prompt when only another creature attacks")
	}
}

func TestS82RankleAndTorbranModes(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rankle := effectstest.B12Push(g, me.ID, "Rankle and Torbran", "Legendary Creature — Faerie Dwarf", s82RankleOracle, 3, 4)
		effectstest.DeclareAttack(t, g, opp.ID, rankle)
		effectstest.AdvanceUntilModePick(t, g, me.ID)
		return g, me, opp, rankle
	}
	t.Run("treasure and damage", func(t *testing.T) {
		g, me, opp, _ := setup(t)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt after combat damage")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0, 2}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, p := range g.Seats {
			if got := s82Treasures(g, p.ID); got != 1 {
				t.Errorf("%s has %d Treasures, want 1", p.ID, got)
			}
		}
		before := opp.Life
		if err := func() (err error) {
			g.WithWriteLock(func() { err = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 1) })
			return
		}(); err != nil {
			t.Fatal(err)
		}
		if got := before - opp.Life; got != 3 {
			t.Errorf("1 damage to a player became %d, want 3", got)
		}
		beforeMe := me.Life
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, me.ID, 2) })
		if got := beforeMe - me.Life; got != 4 {
			t.Errorf("damage to its controller became %d, want 4: any source, any player", got)
		}
	})
	t.Run("each player sacrifices a creature", func(t *testing.T) {
		g, me, opp, rankle := setup(t)
		mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
		theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, p := range g.Seats {
			if c := effectstest.ChooseCardsChoiceFor(g, p.ID); c != nil {
				switch p.ID {
				case me.ID:
					effectstest.AnswerChooseCards(t, g, p.ID, mine)
				case opp.ID:
					effectstest.AnswerChooseCards(t, g, p.ID, theirs)
				}
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(mine) {
			t.Error("a chosen creature was not sacrificed")
		}
		if !g.Battlefield.Contains(rankle) {
			t.Error("Rankle and Torbran was sacrificed instead of the chosen creature")
		}
	})
	t.Run("choosing nothing does nothing", func(t *testing.T) {
		g, me, _, _ := setup(t)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if err := g.ResolveModePick(pick.ID, me.ID, []int{}); err != nil {
			t.Fatalf("any number includes none: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if s82Treasures(g, me.ID) != 0 {
			t.Error("a Treasure appeared")
		}
	})
}

func TestS82ZurgoAndOjutaiLooksTakesOneAndMayBounceTheDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zurgo := effectstest.B12Push(g, me.ID, "Zurgo and Ojutai", "Legendary Creature — Orc Dragon", s82ZurgoOracle, 4, 4)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	effectstest.AttackWith(t, g, opp.ID, zurgo)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 3 {
		t.Fatalf("expected a choice among the top three, got %+v", c)
	}
	taken := c.ChooseCards[0]
	effectstest.AnswerChooseCards(t, g, me.ID, taken)
	// The rest go to the bottom in an order the player names.
	if order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d", got, handBefore+1)
	}
	if got := me.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d, want %d", got, libBefore-1)
	}
	back := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if back == nil || len(back.ChooseCards) != 1 || back.ChooseCards[0] != zurgo {
		t.Fatalf("expected the Dragon offered back, got %+v", back)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, zurgo)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(zurgo) || !effectstest.HandHasID(me, zurgo) {
		t.Error("the Dragon was not returned to its owner's hand")
	}
}

func TestS82ZurgoAndOjutaiOnlyCountsDragonsAndHasHexproofWhenNew(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.B12Push(g, me.ID, "Zurgo and Ojutai", "Legendary Creature — Orc Dragon", s82ZurgoOracle, 4, 4)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		t.Error("a non-Dragon dealing damage triggered it")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	z := effectstest.CastRingCreature(t, g2, "Zurgo and Ojutai", "Legendary Creature — Orc Dragon", s82ZurgoOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g2, z), "hexproof") {
		t.Error("no hexproof the turn it entered")
	}
	_ = me2
}

func TestS82GoroGoroAndSatoruMakesADragonForNewCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Goro-Goro and Satoru", "Legendary Creature — Goblin Human", s82GoroGoroOracle, 3, 4)
	old := effectstest.PushVanillaCreature(g, me.ID, "Old", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, old)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(s82NamedOnBattlefield(g, me.ID, "Dragon Spirit")); n != 0 {
		t.Errorf("a creature that did not enter this turn made %d Dragons", n)
	}
}

func TestS82GoroGoroAndSatoruGrantsHasteAndMakesADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Goro-Goro and Satoru", "Legendary Creature — Goblin Human", s82GoroGoroOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	fresh := effectstest.CastRingCreature(t, g, "Fresh", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	goro := s82FindByName(g, me.ID, "Goro-Goro and Satoru")
	effectstest.B06AddMana(me, "C", "R")
	effectstest.B16Activate(t, g, me.ID, goro, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, fresh), "haste") {
		t.Fatal("the new creature did not gain haste")
	}
	effectstest.AttackWith(t, g, opp.ID, fresh)
	effectstest.PassPriorityAroundTable(t, g)
	dragons := s82NamedOnBattlefield(g, me.ID, "Dragon Spirit")
	if len(dragons) != 1 {
		t.Fatalf("%d Dragon Spirits, want 1", len(dragons))
	}
	if p, tt := effectstest.PtOf(t, g, dragons[0].InstanceID); p != 5 || tt != 5 {
		t.Errorf("Dragon Spirit is %d/%d, want 5/5", p, tt)
	}
}

func s82FindByName(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestS82MoiraAndTesharReturnsANonlandPermanentThatIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Moira and Teshar", "Legendary Creature — Phyrexian Spirit Bird", s82MoiraOracle, 4, 5)
	rock := effectstest.PushCatalogGraveyardCard(me, "Rock", "Artifact", "", 0, 0)
	land := effectstest.PushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	_ = land
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Historic Thing", "Artifact", "", nil)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil {
		t.Fatal("no target prompt for the return")
	}
	if err := func() error {
		p := effectstest.LatestPickTarget(g, me.ID)
		return g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land})
	}(); err == nil {
		t.Error("a land card was accepted")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("the permanent did not return")
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, rock), "haste") {
		t.Error("the returned permanent lacks haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the returned permanent was not exiled at the end step")
	}
	if me.Graveyard.Contains(rock) {
		t.Error("it went to the graveyard rather than exile")
	}
}

func TestS82MoiraAndTesharIgnoresNonHistoricSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Moira and Teshar", "Legendary Creature — Phyrexian Spirit Bird", s82MoiraOracle, 4, 5)
	effectstest.PushCatalogGraveyardCard(me, "Rock", "Artifact", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a nonhistoric spell triggered it")
	}
}

func TestS82BorborygmosDrawsDiscardsLandsAndDealsTwiceThatMuch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 5)
	effectstest.AdvanceToMain(t, g)
	l1 := effectstest.PutInHand(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	l2 := effectstest.PutInHand(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	spell := effectstest.PutInHand(me, game.Card{Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	handBefore := me.Hand.Size()
	effectstest.CastRingCreature(t, g, "Borborygmos and Fblthp", "Legendary Creature — Cyclops Homunculus", s82BorborygmosOracle, 6, 5)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after the draw")
	}
	for _, id := range c.ChooseCards {
		if id == spell {
			t.Error("a nonland card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, l1, l2)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, victim); got != 4 {
		t.Errorf("target has %d damage, want 4 (twice two lands)", got)
	}
	if me.Graveyard.Size() < 2 || !effectstest.HandHasID(me, spell) {
		t.Error("the lands were not discarded or the spell was")
	}
	// +1 drawn, -2 lands discarded; Borborygmos itself came and went.
	if got := me.Hand.Size(); got != handBefore-1 {
		t.Errorf("hand %d, want %d", got, handBefore-1)
	}
}

func TestS82BorborygmosDiscardingNothingDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.PutInHand(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.CastRingCreature(t, g, "Borborygmos and Fblthp", "Legendary Creature — Cyclops Homunculus", s82BorborygmosOracle, 6, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, victim); got != 0 {
		t.Errorf("damage %d with nothing discarded", got)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for with nothing discarded")
	}
}

func TestS82BorborygmosTucksItselfThirdFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bf := effectstest.B12Push(g, me.ID, "Borborygmos and Fblthp", "Legendary Creature — Cyclops Homunculus", s82BorborygmosOracle, 6, 5)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "U")
	effectstest.B16Activate(t, g, me.ID, bf, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bf) {
		t.Fatal("still on the battlefield")
	}
	if len(me.Library.Cards) < 3 || me.Library.Cards[len(me.Library.Cards)-3].InstanceID != bf {
		t.Errorf("not third from the top of the library")
	}
}
