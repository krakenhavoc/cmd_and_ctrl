package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func avB9AddMana(t *testing.T, g *game.Game, who uuid.UUID, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	})
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Bin Me")
	joo := effectstest.B12Push(g, me.ID, "Joo Dee, One of Many", "Creature — Human Advisor", avB9JooDeeOracle, 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")

	avB9AddMana(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, joo, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if n := len(avB9Named(g, me.ID, "Joo Dee, One of Many")); n != 1 {
		t.Fatalf("%d Joo Dees before the surveil was answered, want 1 (the copy waits)", n)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if !g.Battlefield.Contains(joo) || !g.Battlefield.Contains(token) {
		t.Error("a Joo Dee was sacrificed instead of the chosen artifact")
	}
	if tok := effectstest.AvB9Live(t, g, token); tok.Tapped || !tok.IsToken() {
		t.Errorf("the copy should be an untapped token (tapped %v, token %v)", tok.Tapped, tok.IsToken())
	}
}

func TestAvatarB9JooDeeIsSorcerySpeedAndNeedsBlackMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	joo := effectstest.B12Push(g, me.ID, "Joo Dee, One of Many", "Creature — Human Advisor", avB9JooDeeOracle, 2, 2)
	avB9AddMana(t, g, me.ID, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, joo, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated in the draw step")
	}
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	joo2 := effectstest.B12Push(g2, me2.ID, "Joo Dee, One of Many", "Creature — Human Advisor", avB9JooDeeOracle, 2, 2)
	avB9AddMana(t, g2, me2.ID, "{R}")
	if err := g2.ActivateCatalogAbility(me2.ID, joo2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only red mana")
	}
}

// --- June, Bounty Hunter -----------------------------------------------------

func TestAvatarB9JuneIsUnblockableOnlyAfterTwoDraws(t *testing.T) {
	for _, drewTwo := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		june := effectstest.B12Push(g, me.ID, "June, Bounty Hunter", "Legendary Creature — Human Mercenary", avB9JuneOracle, 2, 2)
		wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
		effectstest.AdvanceToMain(t, g)
		drawn := g.TurnTallyFor(me.ID).CardsDrawn
		if drewTwo && drawn < 2 {
			effectstest.AvB9Draw(t, g, me.ID, 2-drawn)
		}
		if !drewTwo && drawn >= 2 {
			t.Fatalf("setup: already drew %d cards", drawn)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(june, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	june := effectstest.B12Push(g, me.ID, "June, Bounty Hunter", "Legendary Creature — Human Mercenary", avB9JuneOracle, 2, 2)

	avB9AddMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, june, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{june}}); err == nil {
		t.Fatal("June sacrificed herself")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, june, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if n := len(avB9Named(g, me.ID, "Clue")); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

func TestAvatarB9JuneActivatesOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	june := effectstest.B12Push(g, other.ID, "June, Bounty Hunter", "Legendary Creature — Human Mercenary", avB9JuneOracle, 2, 2)
	bear := effectstest.B12Creature(g, other.ID, "Bear", "Creature — Bear", 2, 2)
	avB9AddMana(t, g, other.ID, "{C}")
	if err := g.ActivateCatalogAbility(other.ID, june, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("June's ability was activated during an opponent's turn")
	}
}

// --- Lo and Li, Twin Tutors --------------------------------------------------

func TestAvatarB9LoAndLiTutorsALessonOrNoble(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibraryTop(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	effectstest.SeedLibraryTop(me, "Some Lesson", "Sorcery — Lesson", "{1}{B}")
	effectstest.SeedLibraryTop(me, "Some Noble", "Creature — Human Noble", "{2}{B}")

	effectstest.CastCatalogSpell(t, g, "Lo and Li, Twin Tutors", "Legendary Creature — Human Advisor", avB9LoAndLiOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Plain Bear") != uuid.Nil {
		t.Error("the search offered a card that is neither a Lesson nor a Noble")
	}
	if effectstest.SearchOptionNamed(g, c, "Some Noble") == uuid.Nil {
		t.Error("the search did not offer the Noble")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Some Lesson")
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Lo and Li, Twin Tutors", "Legendary Creature — Human Advisor", avB9LoAndLiOracle, 2, 2)
	noble := effectstest.B12Creature(g, me.ID, "My Noble", "Creature — Human Noble", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirNoble := effectstest.B12Creature(g, opp.ID, "Their Noble", "Creature — Human Noble", 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, noble, "lifelink") {
		t.Error("your Noble has no lifelink")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "lifelink") {
		t.Error("a non-Noble creature has lifelink")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirNoble, "lifelink") {
		t.Error("an opponent's Noble has lifelink")
	}

	// Ozai's Cruelty is a Lesson: its 2 damage gains its caster 2 life.
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Ozai's Cruelty", "Sorcery — Lesson", avB9OzaisCrueltyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("target life %d, want %d", opp.Life, oppLife-2)
	}
	if me.Life != life+2 {
		t.Errorf("caster life %d, want %d (the Lesson has lifelink)", me.Life, life+2)
	}
}

// --- Merchant of Many Hats ---------------------------------------------------

func TestAvatarB9MerchantReturnsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Fatal("the Merchant did not return to hand")
	}
}

// --- Northern Air Temple -----------------------------------------------------

func TestAvatarB9NorthernAirTempleDrainsPerShrineAndOnAnotherShrine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Permanent(g, me.ID, "Old Shrine", "Enchantment — Shrine")
	theirs := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Permanent(g, theirs.ID, "Their Shrine", "Enchantment — Shrine")
	opps := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps[p.ID] = p.Life
		}
	}
	life := me.Life

	effectstest.CastCatalogSpell(t, g, "Northern Air Temple", "Legendary Enchantment — Shrine", avB9NorthernTempleOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if want, ok := opps[p.ID]; ok && p.Life != want-2 {
			t.Errorf("opponent life %d, want %d (two Shrines you control)", p.Life, want-2)
		}
	}
	if me.Life != life+2 {
		t.Fatalf("your life %d, want %d", me.Life, life+2)
	}

	effectstest.CastCatalogSpell(t, g, "Sanctum of Fruitful Harvest", "Legendary Enchantment — Shrine", avB9SanctumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life, myLife, hand := opp.Life, me.Life, opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Ozai's Cruelty", "Sorcery — Lesson", avB9OzaisCrueltyOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("target life %d, want %d", opp.Life, life-2)
	}
	if me.Life != myLife {
		t.Errorf("caster life changed to %d without lifelink", me.Life)
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Fatalf("the target owes %d discards, want 2", n)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-2 {
		t.Errorf("target hand %d, want %d", opp.Hand.Size(), hand-2)
	}
}

// --- Phoenix Fleet Airship ---------------------------------------------------

func TestAvatarB9AirshipCopiesItselfOnlyAfterASacrifice(t *testing.T) {
	for _, sacrificed := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ship := effectstest.B12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4)
		effectstest.AdvanceToMain(t, g)
		if sacrificed {
			effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Junk", "Artifact"))
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		ships := avB9Named(g, me.ID, "Phoenix Fleet Airship")
		want := 1
		if sacrificed {
			want = 2
		}
		if len(ships) != want {
			t.Fatalf("sacrificed %v: %d Airships, want %d", sacrificed, len(ships), want)
		}
		if effectstest.AvB9Live(t, g, ship).IsCreature() {
			t.Error("the Airship is a creature with fewer than eight")
		}
		if !effectstest.EffectiveAbilitiesContain(t, g, ship, "flying") {
			t.Error("the Airship has no flying")
		}
	}
}

func TestAvatarB9AirshipIsACreatureWithEightNamedAirships(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var ships []uuid.UUID
	for i := 0; i < 7; i++ {
		ships = append(ships, effectstest.B12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4))
	}
	effectstest.B12Push(g, opp.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4)
	if effectstest.AvB9Live(t, g, ships[0]).IsCreature() {
		t.Fatal("seven Airships (and one an opponent controls) made a creature")
	}
	ships = append(ships, effectstest.B12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4))
	for _, id := range ships {
		c := effectstest.AvB9Live(t, g, id)
		if !c.IsCreature() || !c.IsArtifact() {
			t.Fatalf("with eight Airships one is not an artifact creature (creature %v, artifact %v)", c.IsCreature(), c.IsArtifact())
		}
	}
}

func TestAvatarB9AirshipCrewsForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := effectstest.B12Push(g, me.ID, "Phoenix Fleet Airship", "Artifact — Vehicle", avB9AirshipOracle, 4, 4)
	crewer := effectstest.B12Creature(g, me.ID, "Crewer", "Creature — Human", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB9Live(t, g, ship).IsCreature() {
		t.Error("the crewed Airship is not a creature")
	}
}

// --- Pirate Peddlers ---------------------------------------------------------

func TestAvatarB9PeddlersGrowsWhenYouSacrificeAnotherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	peddlers := effectstest.B12Push(g, me.ID, "Pirate Peddlers", "Creature — Human Pirate", avB9PeddlersOracle, 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, peddlers, "deathtouch") {
		t.Error("Pirate Peddlers has no deathtouch")
	}
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Junk", "Artifact"))
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Old Land", "Land"))
	if got := effectstest.PlusOnes(g, peddlers); got != 2 {
		t.Fatalf("Peddlers has %d counters, want 2", got)
	}
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, opp.ID, "Their Junk", "Artifact"))
	if got := effectstest.PlusOnes(g, peddlers); got != 2 {
		t.Errorf("an opponent's sacrifice grew Peddlers to %d", got)
	}
}

// --- Raven Eagle -------------------------------------------------------------

func TestAvatarB9RavenEagleExilesACreatureCardForAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})

	effectstest.CastCatalogSpell(t, g, "Raven Eagle", "Creature — Bird Assassin", avB9RavenEagleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the creature card was not exiled")
	}
	if n := len(avB9Named(g, me.ID, "Clue")); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

func TestAvatarB9RavenEagleAttackExilesANoncreatureCardForNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eagle := effectstest.B12Push(g, me.ID, "Raven Eagle", "Creature — Bird Assassin", avB9RavenEagleOracle, 2, 3)
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.DeclareAttack(t, g, opp.ID, eagle)
	effectstest.AnswerPickTarget(t, g, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the attack trigger did not exile the card")
	}
	if n := len(avB9Named(g, me.ID, "Clue")); n != 0 {
		t.Errorf("%d Clues for a noncreature card, want 0", n)
	}
}

func TestAvatarB9RavenEagleDrainsOnYourSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Raven Eagle", "Creature — Bird Assassin", avB9RavenEagleOracle, 2, 3)
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	if drawn > 1 {
		t.Fatalf("setup: already drew %d", drawn)
	}
	if drawn == 0 {
		effectstest.AvB9Draw(t, g, me.ID, 1)
	}
	life := me.Life
	effectstest.AvB9Draw(t, g, me.ID, 1)
	if me.Life != life+1 {
		t.Fatalf("life after the second draw %d, want %d", me.Life, life+1)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != game.StartingLife-1 {
			t.Errorf("opponent life %d, want %d", p.Life, game.StartingLife-1)
		}
	}
	effectstest.AvB9Draw(t, g, me.ID, 1)
	if me.Life != life+1 {
		t.Errorf("the third draw drained too: life %d", me.Life)
	}
}
