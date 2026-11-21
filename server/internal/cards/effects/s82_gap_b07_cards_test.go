package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b07_cards_test.go — S82 GAP batch 7 (Innistrad Remastered artifacts,
// #5948).

const (
	gapB07AngelsTomb    = "de3ed52e-7cc2-49ea-96d9-4254aac46168"
	gapB07BlazingTorch  = "cd3d1e5d-db12-4878-92a2-23eb8d23eaf3"
	gapB07Cleaver       = "5cc5be7a-cbe6-4853-bb95-0390e25b4f2b"
	gapB07Chalice       = "28df85d6-e13a-4178-a361-7015afd316ca"
	gapB07CobbledWings  = "8d8682f3-9ef3-4aa7-9ea6-8a2ce09bff6f"
	gapB07Hauberk       = "dac85722-d741-4584-8b98-8ac220958a38"
	gapB07Helvault      = "fa3d8a6f-dc5b-4890-a806-f785a41660c7"
	gapB07Heirloom      = "1374498c-58fd-40b9-b63b-f87fedc42e58"
	gapB07SoulSeparator = "f1da52a4-08a0-4427-a461-a1dc6627b98e"
	gapB07StitchersGraf = "2b73a5ec-c52e-4074-a7cb-78d6eb2adef0"
	gapB07Amulet        = "48b561dd-f23d-409a-87c8-e49902b2477b"
)

func TestGapB07AllAreFull(t *testing.T) {
	for _, o := range []string{gapB07AngelsTomb, gapB07BlazingTorch, gapB07Cleaver, gapB07Chalice, gapB07Chalice + "#1",
		gapB07CobbledWings, gapB07Hauberk, gapB07Helvault, gapB07Heirloom, gapB07Heirloom + "#1", gapB07SoulSeparator,
		gapB07StitchersGraf, gapB07Amulet} {
		spec, ok := Lookup(o)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, completeness %v, caveats %v", o, ok, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB07Equipment(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return b21Push(g, owner, name, "Artifact — Equipment", oracle, 0, 0)
}

func gapB07PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := layeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func TestGapB07TravelersAmuletFetchesABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	amulet := pushCatalogPermanent(g, me.ID, "Traveler's Amulet", "Artifact", gapB07Amulet, false)
	seedSearchLibrary(me,
		searchTestLand("Bayou", "Land — Swamp Forest"),
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(amulet) {
		t.Error("the Amulet was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	prompt := searchChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, prompt, "Bayou") != uuid.Nil {
		t.Error("a nonbasic land is offered")
	}
	answerSearchNamed(t, g, me.ID, "Island")
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Island" {
			found = true
		}
	}
	if !found {
		t.Error("the Island is not in hand")
	}
}

func TestGapB07CobbledWingsGivesFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	wings := gapB07Equipment(g, me.ID, "Cobbled Wings", gapB07CobbledWings)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, wings, bear)
	assertKeywords(t, g, bear, "flying")
}

func TestGapB07ButchersCleaverGivesLifelinkOnlyToAHuman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	human := b16Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2, "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	cleaver := gapB07Equipment(g, me.ID, "Butcher's Cleaver", gapB07Cleaver)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, cleaver, bear)
	if p, th := gapB07PT(t, g, bear); p != 5 || th != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", p, th)
	}
	if got := effectiveAbilities(t, g, bear); gapB07Has(got, "lifelink") {
		t.Error("a non-Human has lifelink")
	}
	gapB07Untap(g, cleaver)
	equipTo(t, g, me.ID, cleaver, human)
	if p, th := gapB07PT(t, g, human); p != 5 || th != 2 {
		t.Errorf("the Human is %d/%d, want 5/2", p, th)
	}
	assertKeywords(t, g, human, "lifelink")
}

func gapB07Has(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func gapB07Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = false })
}

func TestGapB07DemonmailHauberkEquipsBySacrificingACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	fodder := b16Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1, "B")
	hauberk := gapB07Equipment(g, me.ID, "Demonmail Hauberk", gapB07Hauberk)
	advanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, hauberk, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("equipped without sacrificing a creature")
	}
	b16Activate(t, g, me.ID, hauberk, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{fodder}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still there")
	}
	if p, th := gapB07PT(t, g, bear); p != 6 || th != 4 {
		t.Errorf("the Bear is %d/%d, want 6/4", p, th)
	}
}

func TestGapB07BlazingTorchSacrificesToDealTwoAndStopsVampiresAndZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	torch := gapB07Equipment(g, me.ID, "Blazing Torch", gapB07BlazingTorch)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, torch, bear)
	idx, ref := grantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	life := opp.Life
	b16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if g.Battlefield.Contains(torch) {
		t.Error("the Torch was not sacrificed as a cost")
	}
	if !b16Tapped(t, g, bear) {
		t.Error("the creature did not tap to pay")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, opp.Life)
	}
}

func TestGapB07BlazingTorchCantBeBlockedByVampiresOrZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b16Creature(g, me.ID, "Attacker", "Creature — Human Soldier", 2, 2, "W")
	torch := gapB07Equipment(g, me.ID, "Blazing Torch", gapB07BlazingTorch)
	vampire := b16Creature(g, opp.ID, "Vampire", "Creature — Vampire", 2, 2, "B")
	zombie := b16Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2, "B")
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	equipTo(t, g, me.ID, torch, attacker)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(vampire, attacker); err == nil {
		t.Error("a Vampire blocked the equipped creature")
	}
	if err := g.DeclareBlocker(zombie, attacker); err == nil {
		t.Error("a Zombie blocked the equipped creature")
	}
	if err := g.DeclareBlocker(bear, attacker); err != nil {
		t.Errorf("a Bear could not block it: %v", err)
	}
}

func TestGapB07AngelsTombBecomesAnAngelWhenACreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tomb := pushCatalogPermanent(g, me.ID, "Angel's Tomb", "Artifact", gapB07AngelsTomb, false)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, tomb)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Angel") || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("the Tomb is %+v, want a 3/3 Angel artifact creature", c)
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("the Tomb has no flying")
	}
	if cs := c.EffectiveColors(); len(cs) != 1 || cs[0] != "W" {
		t.Errorf("colours %v, want white", cs)
	}
}

func TestGapB07AngelsTombStaysAnArtifactWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tomb := pushCatalogPermanent(g, me.ID, "Angel's Tomb", "Artifact", gapB07AngelsTomb, false)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, tomb); c.IsCreature() {
		t.Error("the declined Tomb became a creature")
	}
}

func TestGapB07StitchersGraftPumpsStopsTheUntapAndSacrificesTheOldHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	other := b16Creature(g, me.ID, "Other", "Creature — Bear", 2, 2, "G")
	graft := gapB07Equipment(g, me.ID, "Stitcher's Graft", gapB07StitchersGraf)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, graft, bear)
	if p, th := gapB07PT(t, g, bear); p != 5 || th != 5 {
		t.Fatalf("the host is %d/%d, want 5/5", p, th)
	}
	declareAttack(t, g, g.Seats[1].ID, bear)
	passPriorityAroundTable(t, g)
	// Four turns later it is still tapped: it skipped its controller's next untap step.
	for i := 0; i < 4; i++ {
		advanceToNextSeatsTurn(t, g)
	}
	advanceTo(t, g, game.StepUpkeep)
	if !b16Tapped(t, g, bear) {
		t.Error("the attacker untapped in its controller's next untap step")
	}
	advanceToMain(t, g)
	// Moving the Graft sacrifices the creature it left.
	equipTo(t, g, me.ID, graft, other)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature the Graft left was not sacrificed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the new host left")
	}
}

func TestGapB07HelvaultExilesAndReturnsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2, "G")
	vault := pushCatalogPermanent(g, me.ID, "Helvault", "Legendary Artifact", gapB07Helvault, false)
	advanceToMain(t, g)
	// {1},{T}: exile a creature you control; the second ability can't take mine.
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("exiled a creature I don't control with the first ability")
	}
	b16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) {
		t.Fatal("my creature was not exiled")
	}
	gapB07Untap(g, vault)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, vault, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("their creature was not exiled")
	}
	// Bounced, the Helvault returns nothing; it must go to the graveyard.
	b18Kill(t, g, vault)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(mine) || g.Exile.Contains(theirs) {
		t.Error("a card is still in exile after the Helvault died")
	}
	if n := len(gapB07Named(g, "Mine")); n != 1 {
		t.Errorf("%d copies of my creature on the battlefield, want 1", n)
	}
	if cs := gapB07Named(g, "Theirs"); len(cs) != 1 || cs[0].Controller != opp.ID {
		t.Errorf("their creature came back as %+v, want under its owner's control", cs)
	}
}

func gapB07Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestGapB07HelvaultStaysQuietWhenItIsBounced(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
	vault := pushCatalogPermanent(g, me.ID, "Helvault", "Legendary Artifact", gapB07Helvault, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(vault) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) {
		t.Error("a bounced Helvault returned its exiled card")
	}
}

func gapB07DFC(owner uuid.UUID, oracle string, face int, front, back game.Face) game.Card {
	c := game.Card{InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner, Faces: []game.Face{front, back}}
	c.SetFace(face)
	return c
}

func TestGapB07ChaliceOfLifeGainsLifeThenTurnsIntoDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chalice := pushBattlefieldCardWithTimestamp(g, gapB07DFC(me.ID, gapB07Chalice, 0,
		game.Face{Name: "Chalice of Life", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Face{Name: "Chalice of Death", TypeLine: "Artifact"}))
	advanceToMain(t, g)
	life := me.Life
	b16Activate(t, g, me.ID, chalice, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d → %d, want +1", life, me.Life)
	}
	if c, _ := battlefieldCard(g, chalice); c.Name != "Chalice of Life" {
		t.Fatalf("turned over at %d life, before the threshold: %s", me.Life, c.Name)
	}
	// At ten above the starting total it transforms.
	g.WithWriteLock(func() {
		me.Life = g.Settings.StartingLife + 9
		findBattlefieldCardForTest(g, chalice).Tapped = false
	})
	b16Activate(t, g, me.ID, chalice, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, chalice); c.Name != "Chalice of Death" {
		t.Fatalf("the Chalice is %q at %d life, want Chalice of Death", c.Name, me.Life)
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, chalice).Tapped = false })
	oppLife := opp.Life
	b16Activate(t, g, me.ID, chalice, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-5 {
		t.Errorf("opponent life %d → %d, want -5", oppLife, opp.Life)
	}
}

func TestGapB07NeglectedHeirloomTurnsIntoAshmouthBladeWhenItsCreatureTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushBattlefieldCardWithTimestamp(g, gapB07DFC(me.ID, "00000000-0000-0000-0000-00000000b007", 0,
		game.Face{Name: "Front", TypeLine: "Creature — Human", Power: 2, Toughness: 2},
		game.Face{Name: "Back", TypeLine: "Creature — Werewolf", Power: 3, Toughness: 3}))
	heirloom := pushBattlefieldCardWithTimestamp(g, gapB07DFC(me.ID, gapB07Heirloom, 0,
		game.Face{Name: "Neglected Heirloom", TypeLine: "Artifact — Equipment", ManaCost: "{1}"},
		game.Face{Name: "Ashmouth Blade", TypeLine: "Artifact — Equipment"}))
	advanceToMain(t, g)
	equipTo(t, g, me.ID, heirloom, host)
	if p, th := gapB07PT(t, g, host); p != 3 || th != 3 {
		t.Fatalf("the host is %d/%d with the Heirloom, want 3/3", p, th)
	}
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(host) })
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, heirloom); c.Name != "Ashmouth Blade" {
		t.Fatalf("the Equipment is %q after its creature transformed, want Ashmouth Blade", c.Name)
	}
	// The back face: +3/+3 and first strike, still attached.
	if p, th := gapB07PT(t, g, host); p != 6 || th != 6 {
		t.Errorf("the host is %d/%d with Ashmouth Blade, want 6/6 (3/3 back face +3/+3)", p, th)
	}
	assertKeywords(t, g, host, "first strike")
}

func TestGapB07SoulSeparatorMakesASpiritCopyAndASizedZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sep := pushCatalogPermanent(g, me.ID, "Soul Separator", "Artifact", gapB07SoulSeparator, false)
	dead := b17GraveyardCard(me, "Hill Giant", "Creature — Giant", "{3}{R}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == dead {
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 3, 3
			}
		}
	})
	other := b17GraveyardCard(me, "Dead Instant", "Instant", "{U}")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sep, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Fatal("targeted a noncreature card")
	}
	b16Activate(t, g, me.ID, sep, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sep) {
		t.Error("the Separator was not sacrificed")
	}
	if !g.Exile.Contains(dead) {
		t.Error("the card was not exiled")
	}
	spirits := gapB07Named(g, "Hill Giant")
	if len(spirits) != 1 {
		t.Fatalf("%d Hill Giant copies, want 1", len(spirits))
	}
	c := layeredCard(t, g, spirits[0].InstanceID)
	if !c.IsToken() || c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasSubtype("Spirit") || !c.HasSubtype("Giant") || !game.HasKeyword(&c, "flying") {
		t.Errorf("the copy is %+v, want a 1/1 flying Giant Spirit token", c)
	}
	zombies := gapB07Named(g, "Zombie")
	if len(zombies) != 1 {
		t.Fatalf("%d Zombies, want 1", len(zombies))
	}
	z := layeredCard(t, g, zombies[0].InstanceID)
	if !z.IsToken() || z.CurrentPower() != 3 || z.CurrentToughness() != 3 || len(z.Colors) != 1 || z.Colors[0] != "B" {
		t.Errorf("the Zombie is %+v, want a 3/3 black token", z)
	}
}
