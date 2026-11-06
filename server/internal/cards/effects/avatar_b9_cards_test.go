package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b9_cards_test.go — Avatar: The Last Airbender, batch 9 (#4459):
// every card cast, triggered or activated through the engine.

const (
	avB9JooDeeOracle        = "4601ed7b-716e-4c89-b733-1eeabe8bf5a2"
	avB9JuneOracle          = "1779869a-5e54-4ab5-86b8-d4d20e2e1ba4"
	avB9LoAndLiOracle       = "ad643992-eb9f-4a4a-9b74-a5aee2337f30"
	avB9MerchantOracle      = "23820b45-2f3a-4b6e-99ea-d84caccb7e37"
	avB9NorthernTempleOracl = "28a2ad5b-d637-4634-8d4b-889e1fc88319"
	avB9OzaisCrueltyOracle  = "aba5c046-2f8c-4749-b347-9a7d8cb1c865"
	avB9AirshipOracle       = "6ffe71c0-acb5-4ba9-8c2f-719330e1ea69"
	avB9PeddlersOracle      = "4771a871-576e-4a14-92f0-b5b41c63d615"
	avB9RavenEagleOracle    = "f75ac0a3-8795-48b4-91fd-4717900ef5b7"
	avB9SanctumOracle       = "132859dd-de66-45c6-8af4-ab5e202a17b0"
)

func TestAvatarB9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB9JooDeeOracle:        "Joo Dee, One of Many",
		avB9JuneOracle:          "June, Bounty Hunter",
		avB9LoAndLiOracle:       "Lo and Li, Twin Tutors",
		avB9MerchantOracle:      "Merchant of Many Hats",
		avB9NorthernTempleOracl: "Northern Air Temple",
		avB9OzaisCrueltyOracle:  "Ozai's Cruelty",
		avB9AirshipOracle:       "Phoenix Fleet Airship",
		avB9PeddlersOracle:      "Pirate Peddlers",
		avB9RavenEagleOracle:    "Raven Eagle",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func avB9AddMana(t *testing.T, g *game.Game, who uuid.UUID, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	})
}

func avB9Draw(t *testing.T, g *game.Game, who uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(who, n); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
}

func avB9Sacrifice(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
}

// avB9Live is a battlefield card with its layered characteristics read
// through a snapshot, so the layer pass has run.
func avB9Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				_ = c.Effective()
				out, found = c, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return out
}

func avB9Named(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Joo Dee, One of Many ----------------------------------------------------

func TestAvatarB9JooDeeSurveilsCopiesHerselfThenSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Bin Me")
	joo := b12Push(g, me.ID, "Joo Dee, One of Many", "Creature — Human Advisor", avB9JooDeeOracle, 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")

	avB9AddMana(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, joo, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if n := len(avB9Named(g, me.ID, "Joo Dee, One of Many")); n != 1 {
		t.Fatalf("%d Joo Dees before the surveil was answered, want 1 (the copy waits)", n)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	joos := avB9Named(g, me.ID, "Joo Dee, One of Many")
	if len(joos) != 2 {
		t.Fatalf("%d Joo Dees after the copy, want 2", len(joos))
	}
	var token uuid.UUID
	for _, id := range joos {
		if id != joo {
			token = id
		}
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[joo] || !offered[token] || !offered[rock] || offered[land] {
		t.Fatalf("sacrifice offered %v; want Joo Dee, the token and the artifact, not the land", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if !g.Battlefield.Contains(joo) || !g.Battlefield.Contains(token) {
		t.Error("a Joo Dee was sacrificed instead of the chosen artifact")
	}
	if tok := avB9Live(t, g, token); tok.Tapped || !tok.IsToken() {
		t.Errorf("the copy should be an untapped token (tapped %v, token %v)", tok.Tapped, tok.IsToken())
	}
}

func TestAvatarB9JooDeeIsSorcerySpeedAndNeedsBlackMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	joo := b12Push(g, me.ID, "Joo Dee, One of Many", "Creature — Human Advisor", avB9JooDeeOracle, 2, 2)
	avB9AddMana(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, joo, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated in the draw step")
	}
	g2 := newCatalogGame(t)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	joo2 := b12Push(g2, me2.ID, "Joo Dee, One of Many", "Creature — Human Advisor", avB9JooDeeOracle, 2, 2)
	avB9AddMana(t, g2, me2.ID, "{R}")
	if err := g2.ActivateCatalogAbility(me2.ID, joo2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only red mana")
	}
}

// --- June, Bounty Hunter -----------------------------------------------------

func TestAvatarB9JuneIsUnblockableOnlyAfterTwoDraws(t *testing.T) {
	for _, drewTwo := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		june := b12Push(g, me.ID, "June, Bounty Hunter", "Legendary Creature — Human Mercenary", avB9JuneOracle, 2, 2)
		wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
		advanceToMain(t, g)
		drawn := g.TurnTallyFor(me.ID).CardsDrawn
		if drewTwo && drawn < 2 {
			avB9Draw(t, g, me.ID, 2-drawn)
		}
		if !drewTwo && drawn >= 2 {
			t.Fatalf("setup: already drew %d cards", drawn)
		}
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(june, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(wall, june)
		if drewTwo && err == nil {
			t.Error("June was blocked after two cards were drawn")
		}
		if !drewTwo && err != nil {
			t.Errorf("June could not be blocked with fewer than two draws: %v", err)
		}
	}
}

func TestAvatarB9JuneSacrificesAnotherCreatureForAClue(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	june := b12Push(g, me.ID, "June, Bounty Hunter", "Legendary Creature — Human Mercenary", avB9JuneOracle, 2, 2)

	avB9AddMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, june, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{june}}); err == nil {
		t.Fatal("June sacrificed herself")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, june, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if n := len(avB9Named(g, me.ID, "Clue")); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

func TestAvatarB9JuneActivatesOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	june := b12Push(g, other.ID, "June, Bounty Hunter", "Legendary Creature — Human Mercenary", avB9JuneOracle, 2, 2)
	bear := b12Creature(g, other.ID, "Bear", "Creature — Bear", 2, 2)
	avB9AddMana(t, g, other.ID, "{C}")
	if err := g.ActivateCatalogAbility(other.ID, june, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("June's ability was activated during an opponent's turn")
	}
}

// --- Lo and Li, Twin Tutors --------------------------------------------------

func TestAvatarB9LoAndLiTutorsALessonOrNoble(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibraryTop(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	seedLibraryTop(me, "Some Lesson", "Sorcery — Lesson", "{1}{B}")
	seedLibraryTop(me, "Some Noble", "Creature — Human Noble", "{2}{B}")

	castCatalogSpell(t, g, "Lo and Li, Twin Tutors", "Legendary Creature — Human Advisor", avB9LoAndLiOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Plain Bear") != uuid.Nil {
		t.Error("the search offered a card that is neither a Lesson nor a Noble")
	}
	if searchOptionNamed(g, c, "Some Noble") == uuid.Nil {
		t.Error("the search did not offer the Noble")
	}
	answerSearchNamed(t, g, me.ID, "Some Lesson")
	passPriorityAroundTable(t, g)
	found := false
	for _, h := range me.Hand.Cards {
		if h.Name == "Some Lesson" {
			found = true
		}
	}
	if !found {
		t.Error("the Lesson did not reach the hand")
	}
}

func TestAvatarB9LoAndLiGivesYourNoblesAndLessonSpellsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Lo and Li, Twin Tutors", "Legendary Creature — Human Advisor", avB9LoAndLiOracle, 2, 2)
	noble := b12Creature(g, me.ID, "My Noble", "Creature — Human Noble", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirNoble := b12Creature(g, opp.ID, "Their Noble", "Creature — Human Noble", 2, 2)
	if !effectiveAbilitiesContain(t, g, noble, "lifelink") {
		t.Error("your Noble has no lifelink")
	}
	if effectiveAbilitiesContain(t, g, bear, "lifelink") {
		t.Error("a non-Noble creature has lifelink")
	}
	if effectiveAbilitiesContain(t, g, theirNoble, "lifelink") {
		t.Error("an opponent's Noble has lifelink")
	}

	// Ozai's Cruelty is a Lesson: its 2 damage gains its caster 2 life.
	life, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Ozai's Cruelty", "Sorcery — Lesson", avB9OzaisCrueltyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("target life %d, want %d", opp.Life, oppLife-2)
	}
	if me.Life != life+2 {
		t.Errorf("caster life %d, want %d (the Lesson has lifelink)", me.Life, life+2)
	}
}

// --- Merchant of Many Hats ---------------------------------------------------

func TestAvatarB9MerchantReturnsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Merchant of Many Hats", TypeLine: "Creature — Human Peasant Ally",
		OracleID: avB9MerchantOracle, Owner: me.ID, Controller: me.ID})
	avB9AddMana(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only two mana")
	}
	avB9AddMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Fatal("the Merchant did not return to hand")
	}
}

// --- Northern Air Temple -----------------------------------------------------

func TestAvatarB9NorthernAirTempleDrainsPerShrineAndOnAnotherShrine(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Permanent(g, me.ID, "Old Shrine", "Enchantment — Shrine")
	theirs := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Permanent(g, theirs.ID, "Their Shrine", "Enchantment — Shrine")
	opps := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps[p.ID] = p.Life
		}
	}
	life := me.Life

	castCatalogSpell(t, g, "Northern Air Temple", "Legendary Enchantment — Shrine", avB9NorthernTempleOracl, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if want, ok := opps[p.ID]; ok && p.Life != want-2 {
			t.Errorf("opponent life %d, want %d (two Shrines you control)", p.Life, want-2)
		}
	}
	if me.Life != life+2 {
		t.Fatalf("your life %d, want %d", me.Life, life+2)
	}

	castCatalogSpell(t, g, "Sanctum of Fruitful Harvest", "Legendary Enchantment — Shrine", avB9SanctumOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if want, ok := opps[p.ID]; ok && p.Life != want-3 {
			t.Errorf("after another Shrine, opponent life %d, want %d", p.Life, want-3)
		}
	}
	if me.Life != life+3 {
		t.Errorf("after another Shrine, your life %d, want %d", me.Life, life+3)
	}
}

// --- Ozai's Cruelty ----------------------------------------------------------

func TestAvatarB9OzaisCrueltyDamagesThenTheTargetDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life, myLife, hand := opp.Life, me.Life, opp.Hand.Size()
	castCatalogSpell(t, g, "Ozai's Cruelty", "Sorcery — Lesson", avB9OzaisCrueltyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("target life %d, want %d", opp.Life, life-2)
	}
	if me.Life != myLife {
		t.Errorf("caster life changed to %d without lifelink", me.Life)
	}
	if n := discardOwed(g, opp.ID); n != 2 {
		t.Fatalf("the target owes %d discards, want 2", n)
	}
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-2 {
		t.Errorf("target hand %d, want %d", opp.Hand.Size(), hand-2)
	}
}

// --- Phoenix Fleet Airship ---------------------------------------------------

func TestAvatarB9AirshipCopiesItselfOnlyAfterASacrifice(t *testing.T) {
	for _, sacrificed := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ship := b12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4)
		advanceToMain(t, g)
		if sacrificed {
			avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Junk", "Artifact"))
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		ships := avB9Named(g, me.ID, "Phoenix Fleet Airship")
		want := 1
		if sacrificed {
			want = 2
		}
		if len(ships) != want {
			t.Fatalf("sacrificed %v: %d Airships, want %d", sacrificed, len(ships), want)
		}
		if avB9Live(t, g, ship).IsCreature() {
			t.Error("the Airship is a creature with fewer than eight")
		}
		if !effectiveAbilitiesContain(t, g, ship, "flying") {
			t.Error("the Airship has no flying")
		}
	}
}

func TestAvatarB9AirshipIsACreatureWithEightNamedAirships(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var ships []uuid.UUID
	for i := 0; i < 7; i++ {
		ships = append(ships, b12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4))
	}
	b12Push(g, opp.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4)
	if avB9Live(t, g, ships[0]).IsCreature() {
		t.Fatal("seven Airships (and one an opponent controls) made a creature")
	}
	ships = append(ships, b12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4))
	for _, id := range ships {
		c := avB9Live(t, g, id)
		if !c.IsCreature() || !c.IsArtifact() {
			t.Fatalf("with eight Airships one is not an artifact creature (creature %v, artifact %v)", c.IsCreature(), c.IsArtifact())
		}
	}
}

func TestAvatarB9AirshipCrewsForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := b12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4)
	crewer := b12Creature(g, me.ID, "Crewer", "Creature — Human", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !avB9Live(t, g, ship).IsCreature() {
		t.Error("the crewed Airship is not a creature")
	}
}

// --- Pirate Peddlers ---------------------------------------------------------

func TestAvatarB9PeddlersGrowsWhenYouSacrificeAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	peddlers := b12Push(g, me.ID, "Pirate Peddlers", "Creature — Human Pirate", avB9PeddlersOracle, 2, 2)
	if !effectiveAbilitiesContain(t, g, peddlers, "deathtouch") {
		t.Error("Pirate Peddlers has no deathtouch")
	}
	avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Junk", "Artifact"))
	avB9Sacrifice(t, g, b12Permanent(g, me.ID, "Old Land", "Land"))
	if got := plusOnes(g, peddlers); got != 2 {
		t.Fatalf("Peddlers has %d counters, want 2", got)
	}
	avB9Sacrifice(t, g, b12Permanent(g, opp.ID, "Their Junk", "Artifact"))
	if got := plusOnes(g, peddlers); got != 2 {
		t.Errorf("an opponent's sacrifice grew Peddlers to %d", got)
	}
}

// --- Raven Eagle -------------------------------------------------------------

func TestAvatarB9RavenEagleExilesACreatureCardForAClue(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})

	castCatalogSpell(t, g, "Raven Eagle", "Creature — Bird Assassin", avB9RavenEagleOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, dead)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the creature card was not exiled")
	}
	if n := len(avB9Named(g, me.ID, "Clue")); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

func TestAvatarB9RavenEagleAttackExilesANoncreatureCardForNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eagle := b12Push(g, me.ID, "Raven Eagle", "Creature — Bird Assassin", avB9RavenEagleOracle, 2, 3)
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	declareAttack(t, g, opp.ID, eagle)
	answerPickTarget(t, g, spell)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the attack trigger did not exile the card")
	}
	if n := len(avB9Named(g, me.ID, "Clue")); n != 0 {
		t.Errorf("%d Clues for a noncreature card, want 0", n)
	}
}

func TestAvatarB9RavenEagleDrainsOnYourSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Raven Eagle", "Creature — Bird Assassin", avB9RavenEagleOracle, 2, 3)
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	if drawn > 1 {
		t.Fatalf("setup: already drew %d", drawn)
	}
	if drawn == 0 {
		avB9Draw(t, g, me.ID, 1)
	}
	life := me.Life
	avB9Draw(t, g, me.ID, 1)
	if me.Life != life+1 {
		t.Fatalf("life after the second draw %d, want %d", me.Life, life+1)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != game.StartingLife-1 {
			t.Errorf("opponent life %d, want %d", p.Life, game.StartingLife-1)
		}
	}
	avB9Draw(t, g, me.ID, 1)
	if me.Life != life+1 {
		t.Errorf("the third draw drained too: life %d", me.Life)
	}
}
