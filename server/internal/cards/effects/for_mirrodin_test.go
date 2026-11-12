package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// for_mirrodin_test.go pins the For Mirrodin! keyword (CR 702.163a, #5654)
// and the nine plain Equipment that carry it: the Rebel is made and the
// Equipment lands on it in one resolution, the token is the rule's 2/2 red
// Rebel, the creation's order prompt is a restore point, and each card's
// static and equip abilities do what it prints.

const (
	barbedBatterfistOracle  = "93c2b91d-7d4e-483c-a11f-4a525f881503"
	dragonwingGliderOracle  = "4cd4c6da-a84e-4cba-893b-799a7cb5c411"
	goldwardensHelmOracle   = "4c35f41a-55e4-49ca-b211-f57d7b1b90f6"
	hexgoldHalberdOracle    = "c43b176a-9001-4fe4-bd9b-809880134067"
	hexgoldHoverwingsOracle = "63018d1c-f7fb-44c8-bcd2-292e74ab5b54"
	kembasBannerOracle      = "bd31c248-8b3f-4fa5-a639-19b0e37b15c3"
	mirranBardicheOracle    = "1fd04d9e-cf89-49ba-a932-5d278f65bf77"
	sylvokBattleChairOracle = "2f6c21a1-d618-497c-bfde-c0f2e1014d99"
	vulshokSplitterOracle   = "6253f9d4-83ff-4a0b-b465-e9f9a565f664"
)

// castForMirrodin casts the named Equipment from the active seat's hand
// and settles the spell and the For Mirrodin! trigger. Returns the
// Equipment and the Rebel it made.
func castForMirrodin(t *testing.T, g *game.Game, name, oracle string) (equipment, rebel uuid.UUID) {
	t.Helper()
	equipment = castCatalogSpell(t, g, name, equipTypeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	rebel = findBattlefieldByName(g, "Rebel")
	if rebel == uuid.Nil {
		t.Fatalf("%s made no Rebel token", name)
	}
	if host := attachmentHostOf(t, g, equipment); host.Kind != game.TargetCard || host.ID != rebel {
		t.Fatalf("%s AttachedTo = %+v, want the Rebel %s", name, host, rebel)
	}
	return equipment, rebel
}

// forMirrodinKeywords are the keywords any of the nine cards grants; a
// card's test checks that it grants its own and none of the others.
var forMirrodinKeywords = []string{"flying", "haste", "vigilance", "trample", "first strike"}

func forMirrodinCheck(t *testing.T, g *game.Game, what string, id uuid.UUID, power, toughness int, keywords ...string) {
	t.Helper()
	if p, th := effectivePTOf(t, g, id); p != power || th != toughness {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, th, power, toughness)
	}
	for _, kw := range forMirrodinKeywords {
		if got, want := hasEffectiveKeyword(t, g, id, kw), containsString(keywords, kw); got != want {
			t.Errorf("%s has %s = %v, want %v", what, kw, got, want)
		}
	}
}

func TestForMirrodinMakesTheRulesRebelAndAttachesTheEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, rebel := castForMirrodin(t, g, "Vulshok Splitter", vulshokSplitterOracle)
	c, ok := g.LookupCardForEffect(rebel)
	if !ok {
		t.Fatal("the Rebel is not on the battlefield")
	}
	if !c.IsToken() || c.Name != "Rebel" || !c.IsCreature() || !c.HasSubtype("Rebel") {
		t.Errorf("token is %q %q (token %v), want a Rebel creature token", c.Name, c.TypeLine, c.IsToken())
	}
	if c.Power != 2 || c.Toughness != 2 {
		t.Errorf("token is printed %d/%d, want 2/2", c.Power, c.Toughness)
	}
	if len(c.Colors) != 1 || c.Colors[0] != "R" {
		t.Errorf("token colours %v, want red", c.Colors)
	}
	if c.Controller != me.ID {
		t.Error("the Rebel is not its Equipment's controller's")
	}
}

// The trigger names the Equipment that entered (CR 400.7). Flickered in
// response, the one on the battlefield is a new object: nothing is
// attached, and the Rebel, a 2/2 and not a Germ, stays.
func TestForMirrodinDoesNotAttachAnEquipmentThatCameBackAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	helm := castCatalogSpell(t, g, "Goldwarden's Helm", equipTypeLine, goldwardensHelmOracle, nil)
	passUntilOnBattlefield(t, g, helm)
	if triggerOnStack(g, helm) == nil {
		t.Fatal("the For Mirrodin! trigger is not on the stack")
	}
	flickerInResponse(t, g, helm)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()

	if host := attachmentHostOf(t, g, helm); host.ID != uuid.Nil {
		t.Errorf("the new Helm was attached to %+v", host)
	}
	rebel := findBattlefieldByName(g, "Rebel")
	if rebel == uuid.Nil {
		t.Fatal("the Rebel was not made, or did not survive unequipped")
	}
	forMirrodinCheck(t, g, "the unequipped Rebel", rebel, 2, 2)
}

// Under a Doubling Season and a Parallel Lives the creation asks for the
// CR 616 order. That prompt is a restore point, and the restored answer
// attaches the Equipment to the first Rebel made, off the keyed tail.
func TestForMirrodinIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	chair := castCatalogSpell(t, g, "Sylvok Battle-Chair", equipTypeLine, sylvokBattleChairOracle, nil)
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	answerOrderBySource(t, g, ds, pl)
	passPriorityAroundTable(t, g)
	host := attachmentHostOf(t, g, chair)
	if host.Kind != game.TargetCard || !g.Battlefield.Contains(host.ID) {
		t.Fatalf("the Battle-Chair is attached to %+v, want a Rebel on the battlefield", host)
	}
	forMirrodinCheck(t, g, "the armed Rebel", host.ID, 6, 6, "trample")
	rebels := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rebel" {
				rebels++
			}
		}
	})
	if rebels != 4 {
		t.Errorf("%d Rebels, want 4 (doubled twice)", rebels)
	}
}

// The seven cards whose static is a fixed bonus: the Rebel arrives with
// it, a strict equip is refused without the mana and at instant speed,
// and a paid equip moves the bonus to the new host.
func TestForMirrodinEquipmentBonusesAndEquipCosts(t *testing.T) {
	cases := []struct {
		name, oracle, equip, pool string
		power, toughness          int
		keywords                  []string
	}{
		{"Barbed Batterfist", barbedBatterfistOracle, "{1}", "{R}", 1, -1, nil},
		{"Dragonwing Glider", dragonwingGliderOracle, "{3}{R}{R}", "{W}{W}{W}{R}{R}", 2, 2, []string{"flying", "haste"}},
		{"Goldwarden's Helm", goldwardensHelmOracle, "{1}{W}", "{R}{W}", 0, 1, nil},
		{"Hexgold Halberd", hexgoldHalberdOracle, "{2}{R}", "{W}{W}{R}", 0, 0, []string{"first strike", "trample"}},
		{"Mirran Bardiche", mirranBardicheOracle, "{3}{W}", "{R}{R}{R}{W}", 2, 1, []string{"vigilance"}},
		{"Sylvok Battle-Chair", sylvokBattleChairOracle, "{5}{G}{G}", "{R}{R}{R}{R}{R}{G}{G}", 4, 4, []string{"trample"}},
		{"Vulshok Splitter", vulshokSplitterOracle, "{2}{R}", "{W}{W}{R}", 2, 0, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := Lookup(tc.oracle)
			if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
				t.Fatalf("%s is not registered Full with no caveats", tc.name)
			}
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			equipment, rebel := castForMirrodin(t, g, tc.name, tc.oracle)
			forMirrodinCheck(t, g, "the Rebel", rebel, 2+tc.power, 2+tc.toughness, tc.keywords...)

			bear := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 3, 3)
			target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
			if err := g.ActivateCatalogAbility(me.ID, equipment, 0, game.ActivateAbilityParams{Targets: target, Strict: true}); err == nil {
				t.Fatalf("equipped without paying %s", tc.equip)
			}
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.pool); err != nil {
				t.Fatal(err)
			}
			if err := g.ActivateCatalogAbility(me.ID, equipment, 0, game.ActivateAbilityParams{Targets: target, Strict: true}); err != nil {
				t.Fatalf("equip for %s: %v", tc.equip, err)
			}
			passPriorityAroundTable(t, g)
			if host := attachmentHostOf(t, g, equipment); host.ID != bear {
				t.Fatalf("%s is attached to %+v, want the Bear", tc.name, host)
			}
			forMirrodinCheck(t, g, "the Bear", bear, 3+tc.power, 3+tc.toughness, tc.keywords...)
			forMirrodinCheck(t, g, "the Rebel left behind", rebel, 2, 2)
		})
	}
}

// Equip is sorcery speed (CR 702.6a): refused on another player's turn
// even with the mana in the pool.
func TestForMirrodinEquipIsRefusedOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fist, _ := castForMirrodin(t, g, "Barbed Batterfist", barbedBatterfistOracle)
	bear := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 3, 3)
	advanceToNextSeatsTurn(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, fist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true,
	}); err == nil {
		t.Fatal("equipped on another player's turn")
	}
}

// Barbed Batterfist's -1 toughness is real: on a 1-toughness creature it
// is a 0-toughness creature, and that dies (CR 704.5f).
func TestBarbedBatterfistKillsAOneToughnessHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fist, _ := castForMirrodin(t, g, "Barbed Batterfist", barbedBatterfistOracle)
	elf := b7Put(g, me.ID, "Elf", "Creature — Elf", "test-elf", 1, 1)
	equipTo(t, g, me.ID, fist, elf)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("a 2/0 Elf survived")
	}
	if !g.Battlefield.Contains(fist) {
		t.Error("the Batterfist left with its host")
	}
}

// Hexgold Halberd's keywords are there during its controller's turn only.
func TestHexgoldHalberdGrantsOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	_, rebel := castForMirrodin(t, g, "Hexgold Halberd", hexgoldHalberdOracle)
	forMirrodinCheck(t, g, "on your turn the Rebel", rebel, 2, 2, "first strike", "trample")
	advanceToNextSeatsTurn(t, g)
	forMirrodinCheck(t, g, "on another turn the Rebel", rebel, 2, 2)
}

// Hexgold Hoverwings: flying is for its own host; the +1/+0 is for every
// equipped creature its controller controls, and for nothing else.
func TestHexgoldHoverwingsPumpsEveryEquippedCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wings, rebel := castForMirrodin(t, g, "Hexgold Hoverwings", hexgoldHoverwingsOracle)
	forMirrodinCheck(t, g, "the Rebel", rebel, 3, 2, "flying")

	bare := b7Put(g, me.ID, "Bare Bear", "Creature — Bear", "test-bare", 2, 2)
	armed := b7Put(g, me.ID, "Armed Bear", "Creature — Bear", "test-armed", 2, 2)
	theirs := b7Put(g, opp.ID, "Their Bear", "Creature — Bear", "test-theirs", 2, 2)
	aura := b7Put(g, me.ID, "Enchanted Bear", "Creature — Bear", "test-enchanted", 2, 2)
	attach := func(name, typeLine string, owner, host uuid.UUID) {
		id := b7Put(g, owner, name, typeLine, "test-"+name, 0, 0)
		g.WithWriteLock(func() {
			if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
				t.Fatalf("attach %s: %v", name, err)
			}
		})
	}
	attach("Plain Sword", equipTypeLine, me.ID, armed)
	attach("Their Sword", equipTypeLine, opp.ID, theirs)
	attach("Plain Aura", "Enchantment — Aura", me.ID, aura)

	forMirrodinCheck(t, g, "an unequipped creature", bare, 2, 2)
	forMirrodinCheck(t, g, "another equipped creature you control", armed, 3, 2)
	forMirrodinCheck(t, g, "an opponent's equipped creature", theirs, 2, 2)
	forMirrodinCheck(t, g, "an enchanted creature", aura, 2, 2)

	// Moved onto the bare Bear, the Rebel is no longer equipped.
	equipTo(t, g, me.ID, wings, bare)
	forMirrodinCheck(t, g, "the Rebel left behind", rebel, 2, 2)
	forMirrodinCheck(t, g, "the new host", bare, 3, 2, "flying")
}

// Kemba's Banner counts its controller's creatures on every pass, the
// equipped creature included.
func TestKembasBannerCountsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	banner, rebel := castForMirrodin(t, g, "Kemba's Banner", kembasBannerOracle)
	forMirrodinCheck(t, g, "the lone Rebel", rebel, 3, 3)

	bear := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 2, 2)
	b7Put(g, me.ID, "Rock", "Artifact", "test-rock", 0, 0)
	b7Put(g, opp.ID, "Their Bear", "Creature — Bear", "test-theirs", 2, 2)
	forMirrodinCheck(t, g, "the Rebel beside a Bear", rebel, 4, 4)

	equipTo(t, g, me.ID, banner, bear)
	forMirrodinCheck(t, g, "the Bear", bear, 4, 4)
	forMirrodinCheck(t, g, "the Rebel left behind", rebel, 2, 2)
}
