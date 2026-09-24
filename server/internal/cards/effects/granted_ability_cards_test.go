package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_ability_cards_test.go — the first cards on ADR 0093's seam
// (the ADR's PR 2): static and attached grants of mana and activated
// abilities. Each test asserts behaviour — who has the ability, who
// may activate it, what it does — rather than a label.

const (
	gaCryptolithRiteOracle   = "043f869d-b11c-4c0d-9591-2bf0df7bde55"
	gaChromaticLanternOracle = "539f5396-d99a-417d-a84c-dff7930b5900"
	gaGemhideSliverOracle    = "2c09ca09-8e62-4fe3-9b3d-61573dd2ffbc"
	gaManaweftSliverOracle   = "bd47398d-da35-4a09-8754-771af91b14f4"
	gaNecroticSliverOracle   = "9655569d-bfa5-4665-9371-9f275b8d223e"
	gaRishkarOracle          = "761021ce-4559-464e-aa03-85c2fe78e267"
	gaJaheiraOracle          = "0aeeb0d7-15a3-4722-ad67-8fc7fe89220d"
	gaInsidiousRootsOracle   = "d75b2b8e-05c9-47da-b359-a867256d78ea"
	gaGreatDivideGuideOracle = "79e69a91-d580-47fb-be76-1e32c50d2fa0"
	gaTheWorldTreeOracle     = "3437d504-bf62-4c27-b15f-f6330182ff7e"
	gaParadiseMantleOracle   = "c1121b83-1ba2-473d-89c9-e3bbd4529072"
	gaSquirrelNestOracle     = "2d584333-b25e-4291-a2c9-80c6d9f8732a"
)

// gaPut seeds a permanent that has been under its controller's control
// since the turn began (so a {T} is not sick).
func gaPut(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner,
	})
}

// layeredCard is the live battlefield card with its layer cache fresh.
func gaLayered(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() { out, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("%s is not in the game", id)
	}
	return out
}

// grantedManaRowsOf counts a permanent's GRANTED mana ability rows.
func grantedManaRowsOf(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		_, origins := game.ManaAbilitiesWithOrigins(c)
		for _, o := range origins {
			if o.Granted() {
				n++
			}
		}
	})
	return n
}

// grantedActivatedIndex is the index of a permanent's first granted
// activated ability, or -1.
func grantedActivatedIndex(t *testing.T, g *game.Game, id uuid.UUID) (int, string) {
	t.Helper()
	_, origins := game.ActivatedAbilitiesWithOrigins(gaLayered(t, g, id))
	for i, o := range origins {
		if o.Granted() {
			return i, o.Ref
		}
	}
	return -1, ""
}

// answerTriggerTargets answers a multi-target trigger's pick_target
// prompt with the given cards.
func answerTriggerTargets(t *testing.T, g *game.Game, chooser uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	refs := make([]game.TargetRef, 0, len(cards))
	for _, c := range cards {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: c})
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// tapGrantedMana taps a permanent for its first granted mana ability,
// answering the colour pick with `color` when one is asked.
func tapGrantedMana(t *testing.T, g *game.Game, player, card uuid.UUID, color string) error {
	t.Helper()
	_, origins := game.ManaAbilitiesWithOrigins(gaLayered(t, g, card))
	for i, o := range origins {
		if !o.Granted() {
			continue
		}
		if err := g.ActivateManaAbility(player, card, i, game.ManaAbilityParams{Ref: o.Ref}); err != nil {
			return err
		}
		for _, ch := range g.PendingChoices {
			if ch != nil && ch.Kind == game.PendingChoiceMana && ch.Chooser == player {
				if err := g.ResolveManaChoice(ch.ID, player, color); err != nil {
					t.Fatalf("ResolveManaChoice: %v", err)
				}
			}
		}
		return nil
	}
	t.Fatalf("%s has no granted mana ability", card)
	return nil
}

func TestCryptolithRiteGivesYourCreaturesAnyColourMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := gaPut(g, opp.ID, "Bear", "Creature — Bear", "")
	gaPut(g, me.ID, "Cryptolith Rite", "Enchantment", gaCryptolithRiteOracle)
	if n := grantedManaRowsOf(g, bear); n != 1 {
		t.Fatalf("your bear has %d granted mana rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's bear has %d granted rows, want 0", n)
	}
	if err := tapGrantedMana(t, g, me.ID, bear, "U"); err != nil {
		t.Fatalf("tap the bear: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
	if !gaLayered(t, g, bear).Tapped {
		t.Error("the granted {T} did not tap the bear")
	}
	// CR 302.6: a creature that entered this turn cannot use it.
	fresh := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fresh Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fresh {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	if err := tapGrantedMana(t, g, me.ID, fresh, "G"); err == nil {
		t.Error("a summoning-sick creature tapped for the Rite's mana")
	}
}

func TestChromaticLanternGivesYourLandsAnyColourMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	lantern := gaPut(g, me.ID, "Chromatic Lantern", "Artifact", gaChromaticLanternOracle)
	if n := grantedManaRowsOf(g, forest); n != 1 {
		t.Fatalf("the Forest has %d granted rows, want 1", n)
	}
	abs := game.ManaAbilitiesForCard(gaLayered(t, g, forest))
	if len(abs) != 2 || abs[0].Produced != "{G}" {
		t.Errorf("Forest mana = %+v, want its own {G} then the granted any colour", abs)
	}
	if n := grantedManaRowsOf(g, lantern); n != 0 {
		t.Errorf("the Lantern is not a land but has %d granted rows", n)
	}
	if err := tapGrantedMana(t, g, me.ID, forest, "R"); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestGemhideSliverGrantsEverySliverAndManaweftOnlyYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := gaPut(g, me.ID, "Muscle Sliver", "Creature — Sliver", "")
	theirs := gaPut(g, opp.ID, "Muscle Sliver", "Creature — Sliver", "")
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	gemhide := gaPut(g, me.ID, "Gemhide Sliver", "Creature — Sliver", gaGemhideSliverOracle)
	if grantedManaRowsOf(g, mine) != 1 || grantedManaRowsOf(g, theirs) != 1 || grantedManaRowsOf(g, gemhide) != 1 {
		t.Error("Gemhide gives ALL Slivers — yours, theirs and itself — the ability")
	}
	if grantedManaRowsOf(g, bear) != 0 {
		t.Error("a Bear is not a Sliver")
	}
	// The opponent's Sliver is theirs to tap, not yours (CR 602.2).
	if err := tapGrantedMana(t, g, me.ID, theirs, "G"); err == nil {
		t.Error("you tapped an opponent's Sliver for mana")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	mine2 := gaPut(g2, me2.ID, "Muscle Sliver", "Creature — Sliver", "")
	theirs2 := gaPut(g2, opp2.ID, "Muscle Sliver", "Creature — Sliver", "")
	gaPut(g2, me2.ID, "Manaweft Sliver", "Creature — Sliver", gaManaweftSliverOracle)
	if grantedManaRowsOf(g2, mine2) != 1 || grantedManaRowsOf(g2, theirs2) != 0 {
		t.Error("Manaweft gives only YOUR Sliver creatures the ability")
	}
}

func TestNecroticSliverGivesEverySliverADestroyAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := gaPut(g, me.ID, "Muscle Sliver", "Creature — Sliver", "")
	gaPut(g, me.ID, "Necrotic Sliver", "Creature — Sliver", gaNecroticSliverOracle)
	target := gaPut(g, opp.ID, "Sol Ring", "Artifact", "")
	idx, ref := grantedActivatedIndex(t, g, sliver)
	if idx < 0 {
		t.Fatal("the Sliver has no granted activated ability")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, sliver, idx, game.ActivateAbilityParams{
		Ref:     ref,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if g.Battlefield.Contains(sliver) {
		t.Error("the Sliver that activated the ability was not sacrificed")
	}
	if g.Battlefield.Contains(target) {
		t.Error("the target was not destroyed")
	}
}

func TestRishkarPutsCountersAndGrantsGreenToCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := gaPut(g, me.ID, "Bear A", "Creature — Bear", "")
	b := gaPut(g, me.ID, "Bear B", "Creature — Bear", "")
	plain := gaPut(g, me.ID, "Bear C", "Creature — Bear", "")
	castCatalogSpell(t, g, "Rishkar, Peema Renegade", "Legendary Creature — Elf Druid", gaRishkarOracle, nil)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, a).Counters[game.CounterPlusOne] != 1 || gaLayered(t, g, b).Counters[game.CounterPlusOne] != 1 {
		t.Fatal("the two targets should each have a +1/+1 counter")
	}
	if grantedManaRowsOf(g, a) != 1 || grantedManaRowsOf(g, b) != 1 {
		t.Error("a creature you control with a counter has {T}: Add {G}")
	}
	if grantedManaRowsOf(g, plain) != 0 {
		t.Error("a creature with no counter does not")
	}
	// Losing its last counter loses the ability.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(a, game.CounterPlusOne, -1) })
	if grantedManaRowsOf(g, a) != 0 {
		t.Error("a creature whose counter came off kept the ability")
	}
}

func TestJaheiraGivesEveryTokenYouControlGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	treasure := TokenCard("1/1 red Goblin")
	treasure.InstanceID, treasure.Owner, treasure.Controller = uuid.New(), me.ID, me.ID
	pushBattlefieldCardWithTimestamp(g, treasure)
	nontoken := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	gaPut(g, me.ID, "Jaheira, Friend of the Forest", "Legendary Creature — Human Elf Druid", gaJaheiraOracle)
	if grantedManaRowsOf(g, treasure.InstanceID) != 1 {
		t.Error("a token you control taps for {G}")
	}
	if grantedManaRowsOf(g, nontoken) != 0 {
		t.Error("a nontoken does not")
	}
	if spec, _ := Lookup(gaJaheiraOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the Background gap is declared")
	}
}

func TestInsidiousRootsGrantsAndGrowsPlants(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	token := TokenCard("0/1 green Plant")
	token.InstanceID, token.Owner, token.Controller = uuid.New(), me.ID, me.ID
	pushBattlefieldCardWithTimestamp(g, token)
	gaPut(g, me.ID, "Insidious Roots", "Enchantment", gaInsidiousRootsOracle)
	if grantedManaRowsOf(g, token.InstanceID) != 1 {
		t.Fatal("a creature token you control taps for any colour")
	}
	// A creature card leaves your graveyard: one new Plant, and a
	// counter on each Plant — the old one and the new one.
	dead := game.NewCard("Dead Bear", me.ID)
	dead.TypeLine = "Creature — Bear"
	me.Graveyard.PushTop(dead)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, dead.InstanceID); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Plant"); n != 2 {
		t.Fatalf("%d Plants, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plant" && c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("a Plant has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
		}
	}
}

func TestGreatDivideGuideGrantsLandsAndAllies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	ally := gaPut(g, me.ID, "Ally", "Creature — Human Ally", "")
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	guide := gaPut(g, me.ID, "Great Divide Guide", "Creature — Human Scout Ally", gaGreatDivideGuideOracle)
	if grantedManaRowsOf(g, forest) != 1 || grantedManaRowsOf(g, ally) != 1 || grantedManaRowsOf(g, guide) != 1 {
		t.Error("each land and Ally you control, the Guide included, has the ability")
	}
	if grantedManaRowsOf(g, bear) != 0 {
		t.Error("a non-Ally creature does not")
	}
}

func TestTheWorldTreeFixesAtSixLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tree := gaPut(g, me.ID, "The World Tree", "Land", gaTheWorldTreeOracle)
	var forests []uuid.UUID
	for i := 0; i < 4; i++ {
		forests = append(forests, gaPut(g, me.ID, "Forest", "Basic Land — Forest", ""))
	}
	if grantedManaRowsOf(g, forests[0]) != 0 || grantedManaRowsOf(g, tree) != 0 {
		t.Fatal("five lands: no grant")
	}
	gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	if grantedManaRowsOf(g, forests[0]) != 1 || grantedManaRowsOf(g, tree) != 1 {
		t.Error("six lands: every land you control, the Tree included, taps for any colour")
	}
	own := game.ManaAbilitiesForCard(gaLayered(t, g, tree))
	if len(own) != 2 || own[0].Produced != "{G}" {
		t.Errorf("the Tree's mana = %+v, want its own {G} then the grant", own)
	}
}

func TestParadiseMantleGrantFollowsTheEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	other := gaPut(g, me.ID, "Other Bear", "Creature — Bear", "")
	mantle := gaPut(g, me.ID, "Paradise Mantle", "Artifact — Equipment", gaParadiseMantleOracle)
	if grantedManaRowsOf(g, bear) != 0 {
		t.Fatal("an unattached Mantle grants nothing")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, mantle, bear)
	if grantedManaRowsOf(g, bear) != 1 || grantedManaRowsOf(g, other) != 0 {
		t.Fatal("the equipped creature has the ability, the other does not")
	}
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, mantle, other)
	if grantedManaRowsOf(g, bear) != 0 || grantedManaRowsOf(g, other) != 1 {
		t.Error("the ability moves with the Mantle")
	}
}

func TestSquirrelNestLetsTheLandsControllerMakeSquirrels(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Nest on an OPPONENT's land feeds the opponent.
	theirLand := gaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	pushAuraOnLand(t, g, me.ID, "Squirrel Nest", gaSquirrelNestOracle, theirLand)
	idx, ref := grantedActivatedIndex(t, g, theirLand)
	if idx < 0 {
		t.Fatal("the enchanted land has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, theirLand, idx, game.ActivateAbilityParams{Ref: ref}); err == nil {
		t.Error("the Aura's controller activated an ability of a land they do not control")
	}
	g.Turn.PriorityHolder = 1
	g.Turn.ActiveSeat = 1
	if err := g.ActivateCatalogAbility(opp.ID, theirLand, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("the land's controller activates it: %v", err)
	}
	passPriorityAroundTable(t, g)
	squirrels := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" {
			squirrels++
			if c.Controller != opp.ID {
				t.Error("the Squirrel belongs to the land's controller")
			}
		}
	}
	if squirrels != 1 {
		t.Errorf("%d Squirrels, want 1", squirrels)
	}
}
