package effects

import (
	"testing"

	"github.com/google/uuid"

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
		ids = append(ids, pushCatalogGraveyardCard(p, "Filler", "Sorcery", "", 0, 0))
	}
	return ids
}

func TestS82ThaliaAndGitrogEntersOpponentsThingsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thalia := b12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	abilities := effectiveAbilities(t, g, thalia)
	for _, kw := range []string{"first strike", "deathtouch"} {
		if !containsString(abilities, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	mine := b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, mine) {
		t.Error("its controller's creatures are untouched")
	}
	if b13Tapped(t, g, b13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")) {
		t.Error("an opponent's basic land enters untapped")
	}
	if !b13Tapped(t, g, b13PlayAs(t, g, 2, "Ancient Tomb", "Land", "")) {
		t.Error("an opponent's nonbasic land enters tapped")
	}
	theirs := b13PlayAs(t, g, 3, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
}

func TestS82ThaliaAndGitrogAllowsAnExtraLandPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	aangAdvanceToMain(t, g, 0)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thalia := b12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	snack := pushVanillaCreature(g, me.ID, "Snack", 1, 1)
	other := pushVanillaCreature(g, me.ID, "Other", 1, 1)
	before := me.Hand.Size()
	declareAttack(t, g, opp.ID, thalia)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, snack)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thalia := b12Push(g, me.ID, "Thalia and The Gitrog Monster", "Legendary Creature — Human Frog Horror", s82ThaliaGitrogOracle, 4, 4)
	forest := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	before := me.Hand.Size()
	declareAttack(t, g, opp.ID, thalia)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the land was not sacrificed")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestS82KroxaExilesFiveToReturnACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	filler := s82Fill(me, 5)
	bear := pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	castRingCreature(t, g, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, filler...)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82Fill(me, 3)
	bear := pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	castRingCreature(t, g, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked to exile five with only four cards")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("returned a creature without paying")
	}
	s82Fill(me, 2)
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	s82Fill(me2, 5)
	bear2 := pushCatalogGraveyardCard(me2, "Bear", "Creature — Bear", "", 2, 2)
	castRingCreature(t, g2, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	passPriorityAroundTable(t, g2)
	answerChooseCards(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(bear2) || me2.Graveyard.Size() != 6 {
		t.Errorf("declining changed something: graveyard %d", me2.Graveyard.Size())
	}
}

func TestS82KroxaFiresOnAttackToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kroxa := b12Push(g, me.ID, "Kroxa and Kunoros", "Legendary Creature — Elder Giant Dog", s82KroxaOracle, 6, 6)
	filler := s82Fill(me, 5)
	bear := pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, kroxa)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, filler...)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the attack trigger did not return the creature")
	}
}

func TestS82SlimefootMakesASaprolingAndReturnsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	slime := pushCatalogGraveyardCard(me, "Slimefoot and Squee", "Legendary Creature — Fungus Goblin", s82SlimefootOracle, 3, 3)
	bear := pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	land := pushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	sap := pushToken(g, me.ID, TokenCard("1/1 green Saproling"))
	advanceToMain(t, g)
	b06AddMana(me, "C", "B", "R", "G")
	// A noncreature card is no target; the ability is sorcery speed.
	if err := g.ActivateCatalogAbility(me.ID, slime, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{sap}, Targets: b16TargetCard(land), Strict: true}); err == nil {
		t.Error("a land card was accepted as the other creature card")
	}
	b16Activate(t, g, me.ID, slime, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sap}, Targets: b16TargetCard(bear)})
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	slime := pushCatalogGraveyardCard(me, "Slimefoot and Squee", "Legendary Creature — Fungus Goblin", s82SlimefootOracle, 3, 3)
	advanceToMain(t, g)
	b06AddMana(me, "C", "B", "R", "G")
	if err := g.ActivateCatalogAbility(me.ID, slime, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no Saproling to sacrifice")
	}
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	if err := g.ActivateCatalogAbility(me.ID, slime, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Error("a non-Saproling paid the cost")
	}
	sap := pushToken(g, me.ID, TokenCard("1/1 green Saproling"))
	b16Activate(t, g, me.ID, slime, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sap}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(slime) {
		t.Error("it did not return alone")
	}
}

func TestS82KoglaFightsOrGainsTrampleAndHaste(t *testing.T) {
	t.Run("fight", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
		kogla := castRingCreature(t, g, "Kogla and Yidaro", "Legendary Creature — Ape Dinosaur Turtle", s82KoglaOracle, 7, 7)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		answerPickTarget(t, g, theirs)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) {
			t.Error("the target survived 7 damage")
		}
		if got := damageOn(t, g, kogla); got != 3 {
			t.Errorf("Kogla has %d damage, want 3", got)
		}
	})
	t.Run("keywords", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		kogla := castRingCreature(t, g, "Kogla and Yidaro", "Legendary Creature — Ape Dinosaur Turtle", s82KoglaOracle, 7, 7)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		abilities := effectiveAbilities(t, g, kogla)
		if !containsString(abilities, "trample") || !containsString(abilities, "haste") {
			t.Errorf("abilities %v lack trample and haste", abilities)
		}
	})
}

func TestS82KoglaDiscardsToDestroyShuffleAndDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	advanceToMain(t, g)
	kid := putInHand(me, game.Card{Name: "Kogla and Yidaro", TypeLine: "Legendary Creature — Ape Dinosaur Turtle",
		ManaCost: "{2}{R}{R}{G}{G}", OracleID: s82KoglaOracle, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	b06AddMana(me, "C", "C", "R", "G")
	b16Activate(t, g, me.ID, kid, 0, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghalta := b12Push(g, me.ID, "Ghalta and Mavren", "Legendary Creature — Dinosaur Vampire", s82GhaltaOracle, 12, 12)
	big := pushVanillaCreature(g, me.ID, "Big", 5, 5)
	small := pushVanillaCreature(g, me.ID, "Small", 2, 2)
	declareAttack(t, g, opp.ID, ghalta, big, small)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatal(err)
	}
	if got := triggersOnStackFrom(g, ghalta); got != 1 {
		t.Fatalf("%d triggers for one declaration, want 1", got)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Fatal("a second mode prompt: the declaration triggered it twice")
	}
	if err := func() error { return nil }(); err != nil {
		t.Fatal(err)
	}
	passAndSettleAttackSplit(t, g, me.ID)
	dinos := s82NamedOnBattlefield(g, me.ID, "Dinosaur")
	if len(dinos) != 1 {
		t.Fatalf("%d Dinosaur tokens, want 1", len(dinos))
	}
	d := dinos[0]
	if p, tt := ptOf(t, g, d.InstanceID); p != 5 || tt != 5 {
		t.Errorf("Dinosaur is %d/%d, want 5/5 (Ghalta's 12 is not counted)", p, tt)
	}
	if !d.Tapped || d.AttackingTarget != opp.ID {
		t.Errorf("Dinosaur tapped=%v attacking=%v, want tapped and attacking the opponent", d.Tapped, d.AttackingTarget)
	}
	if !containsString(effectiveAbilities(t, g, d.InstanceID), "trample") {
		t.Error("Dinosaur lacks trample")
	}
}

func TestS82GhaltaAndMavrenMakesAVampirePerOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghalta := b12Push(g, me.ID, "Ghalta and Mavren", "Legendary Creature — Dinosaur Vampire", s82GhaltaOracle, 12, 12)
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	declareAttack(t, g, opp.ID, ghalta, a, b)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	vamps := s82NamedOnBattlefield(g, me.ID, "Vampire")
	if len(vamps) != 2 {
		t.Fatalf("%d Vampires, want 2", len(vamps))
	}
	for _, v := range vamps {
		if !containsString(effectiveAbilities(t, g, v.InstanceID), "lifelink") || v.AttackingTarget != uuid.Nil {
			t.Errorf("a Vampire lacks lifelink or is attacking: %s", v.Name)
		}
	}
}

func TestS82GhaltaAndMavrenTriggersWhenOthersAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ghalta and Mavren", "Legendary Creature — Dinosaur Vampire", s82GhaltaOracle, 12, 12)
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	declareAttack(t, g, opp.ID, a)
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) == nil {
		t.Fatal("no mode prompt when only another creature attacks")
	}
}

func TestS82RankleAndTorbranModes(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rankle := b12Push(g, me.ID, "Rankle and Torbran", "Legendary Creature — Faerie Dwarf", s82RankleOracle, 3, 4)
		declareAttack(t, g, opp.ID, rankle)
		advanceUntilModePick(t, g, me.ID)
		return g, me, opp, rankle
	}
	t.Run("treasure and damage", func(t *testing.T) {
		g, me, opp, _ := setup(t)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatal("no mode prompt after combat damage")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0, 2}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
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
		mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
		theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		for _, p := range g.Seats {
			if c := chooseCardsChoiceFor(g, p.ID); c != nil {
				switch p.ID {
				case me.ID:
					answerChooseCards(t, g, p.ID, mine)
				case opp.ID:
					answerChooseCards(t, g, p.ID, theirs)
				}
			}
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(mine) {
			t.Error("a chosen creature was not sacrificed")
		}
		if !g.Battlefield.Contains(rankle) {
			t.Error("Rankle and Torbran was sacrificed instead of the chosen creature")
		}
	})
	t.Run("choosing nothing does nothing", func(t *testing.T) {
		g, me, _, _ := setup(t)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if err := g.ResolveModePick(pick.ID, me.ID, []int{}); err != nil {
			t.Fatalf("any number includes none: %v", err)
		}
		passPriorityAroundTable(t, g)
		if s82Treasures(g, me.ID) != 0 {
			t.Error("a Treasure appeared")
		}
	})
}

func TestS82ZurgoAndOjutaiLooksTakesOneAndMayBounceTheDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zurgo := b12Push(g, me.ID, "Zurgo and Ojutai", "Legendary Creature — Orc Dragon", s82ZurgoOracle, 4, 4)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	attackWith(t, g, opp.ID, zurgo)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 3 {
		t.Fatalf("expected a choice among the top three, got %+v", c)
	}
	taken := c.ChooseCards[0]
	answerChooseCards(t, g, me.ID, taken)
	// The rest go to the bottom in an order the player names.
	if order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID); order != nil {
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
	back := chooseCardsChoiceFor(g, me.ID)
	if back == nil || len(back.ChooseCards) != 1 || back.ChooseCards[0] != zurgo {
		t.Fatalf("expected the Dragon offered back, got %+v", back)
	}
	answerChooseCards(t, g, me.ID, zurgo)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(zurgo) || !handHasID(me, zurgo) {
		t.Error("the Dragon was not returned to its owner's hand")
	}
}

func TestS82ZurgoAndOjutaiOnlyCountsDragonsAndHasHexproofWhenNew(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b12Push(g, me.ID, "Zurgo and Ojutai", "Legendary Creature — Orc Dragon", s82ZurgoOracle, 4, 4)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		t.Error("a non-Dragon dealing damage triggered it")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	z := castRingCreature(t, g2, "Zurgo and Ojutai", "Legendary Creature — Orc Dragon", s82ZurgoOracle, 4, 4)
	passPriorityAroundTable(t, g2)
	if !containsString(effectiveAbilities(t, g2, z), "hexproof") {
		t.Error("no hexproof the turn it entered")
	}
	_ = me2
}

func TestS82GoroGoroAndSatoruMakesADragonForNewCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Goro-Goro and Satoru", "Legendary Creature — Goblin Human", s82GoroGoroOracle, 3, 4)
	old := pushVanillaCreature(g, me.ID, "Old", 2, 2)
	attackWith(t, g, opp.ID, old)
	passPriorityAroundTable(t, g)
	if n := len(s82NamedOnBattlefield(g, me.ID, "Dragon Spirit")); n != 0 {
		t.Errorf("a creature that did not enter this turn made %d Dragons", n)
	}
}

func TestS82GoroGoroAndSatoruGrantsHasteAndMakesADragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Goro-Goro and Satoru", "Legendary Creature — Goblin Human", s82GoroGoroOracle, 3, 4)
	advanceToMain(t, g)
	fresh := castRingCreature(t, g, "Fresh", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	goro := s82FindByName(g, me.ID, "Goro-Goro and Satoru")
	b06AddMana(me, "C", "R")
	b16Activate(t, g, me.ID, goro, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !containsString(effectiveAbilities(t, g, fresh), "haste") {
		t.Fatal("the new creature did not gain haste")
	}
	attackWith(t, g, opp.ID, fresh)
	passPriorityAroundTable(t, g)
	dragons := s82NamedOnBattlefield(g, me.ID, "Dragon Spirit")
	if len(dragons) != 1 {
		t.Fatalf("%d Dragon Spirits, want 1", len(dragons))
	}
	if p, tt := ptOf(t, g, dragons[0].InstanceID); p != 5 || tt != 5 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Moira and Teshar", "Legendary Creature — Phyrexian Spirit Bird", s82MoiraOracle, 4, 5)
	rock := pushCatalogGraveyardCard(me, "Rock", "Artifact", "", 0, 0)
	land := pushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	_ = land
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Historic Thing", "Artifact", "", nil)
	if p := latestPickTarget(g, me.ID); p == nil {
		t.Fatal("no target prompt for the return")
	}
	if err := func() error {
		p := latestPickTarget(g, me.ID)
		return g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land})
	}(); err == nil {
		t.Error("a land card was accepted")
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("the permanent did not return")
	}
	if !containsString(effectiveAbilities(t, g, rock), "haste") {
		t.Error("the returned permanent lacks haste")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the returned permanent was not exiled at the end step")
	}
	if me.Graveyard.Contains(rock) {
		t.Error("it went to the graveyard rather than exile")
	}
}

func TestS82MoiraAndTesharIgnoresNonHistoricSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Moira and Teshar", "Legendary Creature — Phyrexian Spirit Bird", s82MoiraOracle, 4, 5)
	pushCatalogGraveyardCard(me, "Rock", "Artifact", "", 0, 0)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a nonhistoric spell triggered it")
	}
}

func TestS82BorborygmosDrawsDiscardsLandsAndDealsTwiceThatMuch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 5)
	advanceToMain(t, g)
	l1 := putInHand(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	l2 := putInHand(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	spell := putInHand(me, game.Card{Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	handBefore := me.Hand.Size()
	castRingCreature(t, g, "Borborygmos and Fblthp", "Legendary Creature — Cyclops Homunculus", s82BorborygmosOracle, 6, 5)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after the draw")
	}
	for _, id := range c.ChooseCards {
		if id == spell {
			t.Error("a nonland card was offered")
		}
	}
	answerChooseCards(t, g, me.ID, l1, l2)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, victim); got != 4 {
		t.Errorf("target has %d damage, want 4 (twice two lands)", got)
	}
	if me.Graveyard.Size() < 2 || !handHasID(me, spell) {
		t.Error("the lands were not discarded or the spell was")
	}
	// +1 drawn, -2 lands discarded; Borborygmos itself came and went.
	if got := me.Hand.Size(); got != handBefore-1 {
		t.Errorf("hand %d, want %d", got, handBefore-1)
	}
}

func TestS82BorborygmosDiscardingNothingDealsNoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 5)
	advanceToMain(t, g)
	putInHand(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	castRingCreature(t, g, "Borborygmos and Fblthp", "Legendary Creature — Cyclops Homunculus", s82BorborygmosOracle, 6, 5)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, victim); got != 0 {
		t.Errorf("damage %d with nothing discarded", got)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for with nothing discarded")
	}
}

func TestS82BorborygmosTucksItselfThirdFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bf := b12Push(g, me.ID, "Borborygmos and Fblthp", "Legendary Creature — Cyclops Homunculus", s82BorborygmosOracle, 6, 5)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "U")
	b16Activate(t, g, me.ID, bf, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bf) {
		t.Fatal("still on the battlefield")
	}
	if len(me.Library.Cards) < 3 || me.Library.Cards[len(me.Library.Cards)-3].InstanceID != bf {
		t.Errorf("not third from the top of the library")
	}
}
