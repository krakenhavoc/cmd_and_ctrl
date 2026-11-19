package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BNS batch 8 (#5929): nine cards from The List.

const (
	bnsB08Bishop    = "e9398f3b-c457-4fff-9b09-377c612b6a48"
	bnsB08Burden    = "1ae88ba6-4973-45d1-a251-4d924ab4bc73"
	bnsB08Consign   = "8fe7436b-2467-486e-8a2d-7cbf21a65da9"
	bnsB08Death     = "680f5eca-b98e-4f73-a852-ab629d94297e"
	bnsB08Duskmantl = "67b2cd0c-ecc8-4129-b1ac-820c9924190c"
	bnsB08Gnaw      = "7b92ac54-3883-48cf-b01e-be9797dc1649"
	bnsB08Hard      = "0728b195-2dfe-4d73-91ae-e658a97b0b9a"
	bnsB08HighAlert = "5f657279-cc0f-4431-b517-47bd1c7109b6"
	bnsB08Ixidor    = "90b98aa7-c8d0-4853-887c-164f8160c3f3"
)

func TestBnsB08EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		bnsB08Bishop: "Bishop of the Bloodstained", bnsB08Burden: "Burden of Guilt",
		bnsB08Consign: "Consign", bnsB08Consign + "#1": "Oblivion", bnsB08Death: "Death Cloud",
		bnsB08Duskmantl: "Duskmantle, House of Shadow", bnsB08Gnaw: "Gnaw to the Bone",
		bnsB08Hard: "Hard Evidence", bnsB08HighAlert: "High Alert", bnsB08Ixidor: "Ixidor, Reality Sculptor",
	}
	for oracle, name := range want {
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

func TestBnsB08BishopDrainsOnePerVampireIncludingItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTestPermanent(g, me.ID, game.Card{Name: "Fang", TypeLine: "Creature — Vampire Rogue", Power: 1, Toughness: 1})
	pushTestPermanent(g, me.ID, game.Card{Name: "Bat", TypeLine: "Creature — Bat", Power: 1, Toughness: 1})
	pushTestPermanent(g, opp.ID, game.Card{Name: "Their Fang", TypeLine: "Creature — Vampire", Power: 1, Toughness: 1})
	before := opp.Life
	castCatalogSpell(t, g, "Bishop of the Bloodstained", "Creature — Vampire Cleric", bnsB08Bishop, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("opponent lost %d, want 2 (Fang and the Bishop)", got)
	}
}

func TestBnsB08BurdenOfGuiltTapsTheEnchantedCreatureForOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := auraBear(g, opp.ID)
	aura := auraCast(t, g, "Burden of Guilt", bnsB08Burden, theirs)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); !tapped {
		t.Error("the enchanted creature is untapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestBnsB08ConsignBouncesAndOblivionDiscardsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	land := pushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	advanceToMain(t, g)
	c := game.Card{
		InstanceID: uuid.New(), OracleID: bnsB08Consign, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Consign", TypeLine: "Instant", ManaCost: "{1}{U}", OracleText: "Return target nonland permanent to its owner's hand."},
			{Name: "Oblivion", TypeLine: "Sorcery", ManaCost: "{4}{B}", OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nTarget opponent discards two cards."},
		},
	}
	c.SettleImported()
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, Targets: cardTarget(land)}); err == nil {
		t.Error("Consign targeted a land")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: cardTarget(theirs)}); err != nil {
		t.Fatalf("cast Consign: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Fatal("the permanent was not returned to its owner's hand")
	}

	me.Graveyard.PushTop(c)
	handBefore := opp.Hand.Size()
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, FromZone: "graveyard",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast Oblivion: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, opp.ID); got != 2 {
		t.Fatalf("the opponent owes %d discards, want 2", got)
	}
	discardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != handBefore-2 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), handBefore-2)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("Oblivion is not exiled after resolving from the graveyard")
	}
}

func bnsB08DeathCloudTable(g *game.Game) (creatures, lands [][]uuid.UUID) {
	creatures = make([][]uuid.UUID, len(g.Seats))
	lands = make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		for j := 0; j < 3; j++ {
			creatures[i] = append(creatures[i], pushVanillaCreature(g, p.ID, "Fodder", 1, 1))
			lands[i] = append(lands[i], socB17Perm(g, p.ID, "Forest", "Basic Land — Forest", "", "", 0, 0))
		}
	}
	return creatures, lands
}

func TestBnsB08DeathCloudDoesFourThingsXTimesToEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	creatures, lands := bnsB08DeathCloudTable(g)
	lives := make([]int, len(g.Seats))
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i], hands[i] = p.Life, p.Hand.Size()
	}
	castXSpell(t, g, "Death Cloud", "Sorcery", bnsB08Death, "{X}{B}{B}{B}", 2, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != lives[i]-2 {
			t.Errorf("seat %d life %d, want %d", i, p.Life, lives[i]-2)
		}
	}
	for offset := range g.Seats {
		p := g.Seats[(active+offset)%len(g.Seats)]
		if got := discardOwed(g, p.ID); got != 2 {
			t.Fatalf("seat %s owes %d discards, want 2", p.Name, got)
		}
		for i := range g.Seats {
			for _, id := range append(append([]uuid.UUID{}, creatures[i]...), lands[i]...) {
				if !g.Battlefield.Contains(id) {
					t.Fatal("a permanent was sacrificed before the discards finished")
				}
			}
		}
		discardFromHand(t, g, p.ID)
	}
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]-2 {
			t.Errorf("seat %d hand %d, want %d", i, p.Hand.Size(), hands[i]-2)
		}
	}
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		prompt := sacrificeChoiceFor(g, p.ID)
		if prompt == nil {
			t.Fatalf("seat %s was not asked for creatures", p.Name)
		}
		answerSacrifice(t, g, p.ID, creatures[seat][0], creatures[seat][1])
	}
	for i := range g.Seats {
		if g.Battlefield.Contains(creatures[i][0]) || g.Battlefield.Contains(creatures[i][1]) || !g.Battlefield.Contains(creatures[i][2]) {
			t.Errorf("seat %d: the wrong creatures were sacrificed", i)
		}
		for _, id := range lands[i] {
			if !g.Battlefield.Contains(id) {
				t.Fatalf("seat %d: a land was sacrificed during the creature step", i)
			}
		}
	}
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		prompt := sacrificeChoiceFor(g, p.ID)
		if prompt == nil {
			t.Fatalf("seat %s was not asked for lands", p.Name)
		}
		answerSacrifice(t, g, p.ID, lands[seat][0], lands[seat][2])
	}
	passPriorityAroundTable(t, g)
	for i := range g.Seats {
		if g.Battlefield.Contains(lands[i][0]) || g.Battlefield.Contains(lands[i][2]) || !g.Battlefield.Contains(lands[i][1]) {
			t.Errorf("seat %d: the wrong lands were sacrificed", i)
		}
	}
}

func TestBnsB08DeathCloudWithXZeroDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	creatures, _ := bnsB08DeathCloudTable(g)
	life := g.Seats[0].Life
	castXSpell(t, g, "Death Cloud", "Sorcery", bnsB08Death, "{X}{B}{B}{B}", 0, nil)
	passPriorityAroundTable(t, g)
	if g.Seats[0].Life != life || !g.Battlefield.Contains(creatures[0][0]) {
		t.Error("X = 0 changed the table")
	}
}

func TestBnsB08DeathCloudTakesWhatAShortPlayerHas(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	creatures, lands := bnsB08DeathCloudTable(g)
	short := (active + 1) % len(g.Seats)
	g.Seats[short].Hand.Cards = g.Seats[short].Hand.Cards[:1]
	g.WithWriteLock(func() {
		g.Battlefield.Cards = socB17RemoveCards(g.Battlefield.Cards, creatures[short][1], creatures[short][2])
	})
	castXSpell(t, g, "Death Cloud", "Sorcery", bnsB08Death, "{X}{B}{B}{B}", 2, nil)
	passPriorityAroundTable(t, g)
	for offset := range g.Seats {
		p := g.Seats[(active+offset)%len(g.Seats)]
		if (active+offset)%len(g.Seats) == short {
			discardFromHand(t, g, p.ID)
			continue
		}
		discardFromHand(t, g, p.ID)
	}
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		if seat == short {
			continue
		}
		answerSacrifice(t, g, g.Seats[seat].ID, creatures[seat][0], creatures[seat][1])
	}
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		answerSacrifice(t, g, g.Seats[seat].ID, lands[seat][0], lands[seat][1])
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(creatures[short][0]) {
		t.Error("the short player's only creature was not sacrificed")
	}
	if g.Seats[short].Hand.Size() != 0 {
		t.Errorf("the short player kept %d cards", g.Seats[short].Hand.Size())
	}
}

func TestBnsB08DuskmantleMillsAPlayerForUBTap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := pushTestPermanent(g, me.ID, game.Card{Name: "Duskmantle, House of Shadow", OracleID: bnsB08Duskmantl, TypeLine: "Land"})
	lib := opp.Library.Size()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d", opp.Library.Size(), lib-1)
	}
	if tapped, _ := battlefieldCardTapped(g, land); !tapped {
		t.Error("the land did not tap as a cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a tapped land activated again")
	}
}

func TestBnsB08GnawGainsTwoPerCreatureCardThenFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Elf", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.Seats[1].Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	life := me.Life
	castCatalogSpell(t, g, "Gnaw to the Bone", "Instant", bnsB08Gnaw, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d (two creature cards in your graveyard)", me.Life, life+4)
	}
	id := midB03Flash(t, g, "Gnaw to the Bone", "Instant", bnsB08Gnaw, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+8 {
		t.Errorf("after the flashback life %d, want %d", me.Life, life+8)
	}
	midB03Exiled(t, g, id, "Gnaw to the Bone")
}

func TestBnsB08GnawWithNoCreatureCardsGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	life := g.Seats[0].Life
	castCatalogSpell(t, g, "Gnaw to the Bone", "Instant", bnsB08Gnaw, nil)
	passPriorityAroundTable(t, g)
	if g.Seats[0].Life != life {
		t.Errorf("life %d, want %d", g.Seats[0].Life, life)
	}
}

func TestBnsB08HardEvidenceMakesACrabAndAClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Hard Evidence", "Sorcery", bnsB08Hard, nil)
	passPriorityAroundTable(t, g)
	crabs := midB03Tokens(g, me.ID, "Crab")
	if len(crabs) != 1 {
		t.Fatalf("%d Crabs, want 1", len(crabs))
	}
	if c := crabs[0]; c.Power != 0 || c.Toughness != 3 || len(c.Colors) != 1 || c.Colors[0] != "U" {
		t.Errorf("Crab is %d/%d %v, want 0/3 blue", c.Power, c.Toughness, c.Colors)
	}
	if n := countTokensNamed2(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

func TestBnsB08HighAlertAssignsByToughnessWaivesDefenderAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alert := pushTestPermanent(g, me.ID, game.Card{Name: "High Alert", OracleID: bnsB08HighAlert, TypeLine: "Enchantment"})
	wall := pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	glass := pushVanillaCreature(g, me.ID, "Glass Cannon", 4, 1)
	theirWall := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	advanceToMainOf(t, g, 0)
	before := g.Clone()
	advanceTo(t, before, game.StepDeclareAttackers)
	if err := declareResult(before, wall, opp.ID); err != nil {
		t.Errorf("your defender may attack: %v", err)
	}
	wantDefenderRefused(t, before, theirWall, me.ID)

	life := opp.Life
	attackWith(t, g, opp.ID, wall, glass)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 5 {
		t.Errorf("opponent lost %d, want 5 (4 from the Wall's toughness, 1 from the Cannon's)", got)
	}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{W}{U}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if tapped, _ := battlefieldCardTapped(g, wall); !tapped {
		t.Fatal("fixture: the attacking Wall should be tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, alert, 0, game.ActivateAbilityParams{Strict: true,
		Targets: cardTarget(wall)}); err != nil {
		t.Fatalf("untap: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, wall); tapped {
		t.Error("the Wall is still tapped")
	}
}

func TestBnsB08IxidorPumpsFaceDownCreaturesAndTurnsOneUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ixidor := pushTestPermanent(g, me.ID, game.Card{Name: "Ixidor, Reality Sculptor", OracleID: bnsB08Ixidor,
		TypeLine: "Legendary Creature — Human Wizard", Power: 3, Toughness: 4})
	faceUp := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	down := manifestTop(t, g, opp, bearCard())
	if p, tough := effectivePower(t, g, down), effectiveToughness(t, g, down); p != 3 || tough != 3 {
		t.Errorf("the face-down creature is %d/%d, want 3/3 (2/2 plus Ixidor's +1/+1)", p, tough)
	}
	if p := effectivePower(t, g, faceUp); p != 2 {
		t.Errorf("a face-up creature is %d power, want 2", p)
	}
	if p := effectivePower(t, g, ixidor); p != 3 {
		t.Errorf("Ixidor is %d power, want 3", p)
	}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ixidor, 0, game.ActivateAbilityParams{Strict: true,
		Targets: cardTarget(faceUp)}); err == nil {
		t.Error("a face-up creature is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, ixidor, 0, game.ActivateAbilityParams{Strict: true,
		Targets: cardTarget(down)}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, down)
	if c.FaceDown || c.Name != "Sleeping Bear" {
		t.Fatalf("the creature is still face down: face_down=%v name=%q", c.FaceDown, c.Name)
	}
	if p := effectivePower(t, g, down); p != 2 {
		t.Errorf("the face-up Bear is %d power, want 2 (no more anthem)", p)
	}
}
