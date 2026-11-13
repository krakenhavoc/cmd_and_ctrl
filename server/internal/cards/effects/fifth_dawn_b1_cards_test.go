package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fifth_dawn_b1_cards_test.go — Fifth Dawn, batch 1 (#5604, S82): Acquire
// through Beacon of Destruction.

const (
	dn5AcquireOracle             = "f1c3d749-5cdb-4af2-967f-6413d6cb04fe"
	dn5AdvancedHoverguardOracle  = "65dc61a4-9052-49d4-bce8-9c16f0339382"
	dn5AllSunsDawnOracle         = "a97d9a72-2e47-491b-b1db-9829a9abbbd7"
	dn5AnodetLurkerOracle        = "5c2dbe90-5784-4bc4-a715-76a61c1ecfbc"
	dn5ArachnoidOracle           = "7116db6d-82ea-4675-a0cd-73aaf94be43f"
	dn5ArmedResponseOracle       = "0258a606-1625-402e-b0c4-ffca06779b5e"
	dn5ArtificersIntuitionOracle = "be14b10a-7cac-47fc-a954-685091990291"
	dn5AuriokChampionOracle      = "633fd158-19d0-4d97-a076-2ad897bbed19"
	dn5AuriokSalvagersOracle     = "db2f3a06-f75e-4689-b147-9820aebb30e2"
	dn5AuriokWindwalkerOracle    = "f8a06210-de30-4e87-835e-80a6df5716ca"
	dn5AvariceTotemOracle        = "c4698dcc-baa0-44c5-a283-22f9dfe150f5"
	dn5BatteredGolemOracle       = "066dfefd-eef5-4c81-8e8a-310bac01674f"
	dn5BeaconOfCreationOracle    = "55f31c9c-a015-4965-a0c4-788b861dba21"
	dn5BeaconOfDestructionOracle = "15feb59b-bbbe-462a-8d38-d20735bbc2f3"
)

func TestDn5B1CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Acquire": dn5AcquireOracle, "Advanced Hoverguard": dn5AdvancedHoverguardOracle,
		"All Suns' Dawn": dn5AllSunsDawnOracle, "Anodet Lurker": dn5AnodetLurkerOracle,
		"Arachnoid": dn5ArachnoidOracle, "Armed Response": dn5ArmedResponseOracle,
		"Artificer's Intuition": dn5ArtificersIntuitionOracle, "Auriok Champion": dn5AuriokChampionOracle,
		"Auriok Salvagers": dn5AuriokSalvagersOracle, "Auriok Windwalker": dn5AuriokWindwalkerOracle,
		"Avarice Totem": dn5AvariceTotemOracle, "Battered Golem": dn5BatteredGolemOracle,
		"Beacon of Creation": dn5BeaconOfCreationOracle, "Beacon of Destruction": dn5BeaconOfDestructionOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered as %q, %v, caveats %v; want its name, Full and none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dn5GraveyardCard puts a card with a mana cost and colors into a
// player's graveyard.
func dn5GraveyardCard(p *game.Player, name, typeLine, manaCost string, colors ...string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, Colors: colors,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- Acquire -----------------------------------------------------------------

func TestDn5AcquireTakesAnArtifactFromAnOpponentsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := pushLibraryCardForTest(opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	rock := pushLibraryCardForTest(opp, game.Card{Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{2}"})

	castCatalogSpell(t, g, "Acquire", "Sorcery", dn5AcquireOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)

	if !g.Battlefield.Contains(rock) {
		t.Fatal("the artifact card should be on the battlefield")
	}
	if got := controllerOf(t, g, rock); got != me.ID {
		t.Errorf("the artifact is controlled by %v, want the caster", got)
	}
	if !opp.Library.Contains(bear) {
		t.Error("a creature card is not an artifact card and stays in the library")
	}
}

func TestDn5AcquireRefusesItsCastersOwnLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := castCatalogSpellErr(t, g, "Acquire", "Sorcery", dn5AcquireOracle, []game.TargetRef{playerTarget(me.ID)}); err == nil {
		t.Error("the caster was accepted as the target opponent")
	}
}

// --- Advanced Hoverguard -----------------------------------------------------

func TestDn5AdvancedHoverguardGainsShroud(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hover := mrd2Catalog(g, me.ID, "Advanced Hoverguard", "Creature — Drone", dn5AdvancedHoverguardOracle, "{3}{U}", 2, 2)
	if !hasKeywordOnBattlefield(t, g, hover, "flying") {
		t.Error("Advanced Hoverguard has flying")
	}
	if hasKeywordOnBattlefield(t, g, hover, "shroud") {
		t.Fatal("no shroud before the ability")
	}
	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, hover, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, hover, "shroud") {
		t.Error("the Hoverguard should have shroud until end of turn")
	}
}

// --- All Suns' Dawn ----------------------------------------------------------

func TestDn5AllSunsDawnReturnsOneCardPerColorAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	white := dn5GraveyardCard(me, "White Card", "Sorcery", "{W}", "W")
	azorius := dn5GraveyardCard(me, "White-Blue Card", "Instant", "{W}{U}", "W", "U")
	green := dn5GraveyardCard(me, "Green Card", "Creature — Bear", "{1}{G}", "G")
	black := dn5GraveyardCard(me, "Black Card", "Instant", "{B}", "B")

	spell := castCatalogSpell(t, g, "All Suns' Dawn", "Sorcery", dn5AllSunsDawnOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: white, Slot: 0},
		{Kind: game.TargetCard, ID: azorius, Slot: 1},
		{Kind: game.TargetCard, ID: green, Slot: 4},
	})
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"white": white, "white-blue": azorius, "green": green} {
		if !handHasID(me, id) {
			t.Errorf("the %s card should return to hand", name)
		}
	}
	if handHasID(me, black) {
		t.Error("the black card was not a target")
	}
	if !exileHas(g, spell) {
		t.Error("All Suns' Dawn should exile itself")
	}
}

func TestDn5AllSunsDawnRefusesACardForTwoColorsOrTheWrongColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	azorius := dn5GraveyardCard(me, "White-Blue Card", "Instant", "{W}{U}", "W", "U")
	green := dn5GraveyardCard(me, "Green Card", "Creature — Bear", "{1}{G}", "G")
	if err := castCatalogSpellErr(t, g, "All Suns' Dawn", "Sorcery", dn5AllSunsDawnOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: azorius, Slot: 0},
		{Kind: game.TargetCard, ID: azorius, Slot: 1},
	}); err == nil {
		t.Error("one card was accepted for both the white and the blue slot")
	}
	// Stamped, as the client sends it: a lone pick at slot 0 is the white
	// slot's, not moved to the green one.
	spell := pushCatalogHandCard(me, "All Suns' Dawn", "Sorcery", dn5AllSunsDawnOracle)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: green, Slot: 0}}, TargetsStamped: true,
	}); err == nil {
		t.Error("a green card was accepted as the white card")
	}
}

// --- Anodet Lurker -----------------------------------------------------------

func TestDn5AnodetLurkerGainsThreeLifeWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	lurker := mrd2Catalog(g, me.ID, "Anodet Lurker", "Artifact Creature — Construct", dn5AnodetLurkerOracle, "{5}", 3, 3)
	life := me.Life
	p1Destroy(t, g, lurker)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Arachnoid ---------------------------------------------------------------

func TestDn5ArachnoidHasReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spider := mrd2Catalog(g, me.ID, "Arachnoid", "Artifact Creature — Spider", dn5ArachnoidOracle, "{6}", 2, 6)
	if !hasKeywordOnBattlefield(t, g, spider, "reach") {
		t.Error("Arachnoid has reach")
	}
}

// --- Armed Response ----------------------------------------------------------

func TestDn5ArmedResponseDealsDamageEqualToEquipmentCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	attacker := blbB1Push(g, me.ID, "Attacking Bear", "Creature — Bear", "", 2, 2)
	idle := blbB1Push(g, me.ID, "Idle Bear", "Creature — Bear", "", 2, 2)
	seedEquipment(g, opp.ID, "Their Sword", "")
	seedEquipment(g, opp.ID, "Their Shield", "")
	seedEquipment(g, me.ID, "My Sword", "")
	declareAttack(t, g, opp.ID, attacker)

	response := pushCatalogHandCard(opp, "Armed Response", "Instant", dn5ArmedResponseOracle)
	if err := g.CastSpell(opp.ID, response, game.CastSpellParams{Targets: cardTarget(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	if err := g.CastSpell(opp.ID, response, game.CastSpellParams{Targets: cardTarget(attacker)}); err != nil {
		t.Fatalf("cast Armed Response: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("two damage, from the caster's two Equipment, should destroy the 2/2 attacker")
	}
}

// --- Artificer's Intuition ---------------------------------------------------

func TestDn5ArtificersIntuitionDiscardsAnArtifactForACheapOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	intuition := mrd2Catalog(g, me.ID, "Artificer's Intuition", "Enchantment", dn5ArtificersIntuitionOracle, "{1}{U}", 0, 0)
	pitchArtifact := pushCatalogHandCard(me, "Spare Rock", "Artifact", "")
	pitchSorcery := pushCatalogHandCard(me, "Spare Sorcery", "Sorcery", "")
	cheap := pushLibraryCardForTest(me, game.Card{Name: "Cheap Trinket", TypeLine: "Artifact", ManaCost: "{1}"})
	dear := pushLibraryCardForTest(me, game.Card{Name: "Dear Trinket", TypeLine: "Artifact", ManaCost: "{2}"})

	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, intuition, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitchSorcery}}); err == nil {
		t.Fatal("a sorcery card was accepted as the discarded artifact card")
	}
	if err := g.ActivateCatalogAbility(me.ID, intuition, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitchArtifact}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if handHasID(me, pitchArtifact) {
		t.Error("the artifact card is discarded as the cost")
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !handHasID(me, cheap) {
		t.Error("the mana value 1 artifact card should be put into the hand")
	}
	if !me.Library.Contains(dear) {
		t.Error("a mana value 2 artifact card is not offered")
	}
}

// --- Auriok Champion ---------------------------------------------------------

func TestDn5AuriokChampionMayGainLifeWhenAnotherCreatureEnters(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		champion := mrd2Catalog(g, me.ID, "Auriok Champion", "Creature — Human Cleric", dn5AuriokChampionOracle, "{W}{W}", 1, 1)
		for _, kw := range []string{game.ProtectionFromColor("B"), game.ProtectionFromColor("R")} {
			if !hasKeywordOnBattlefield(t, g, champion, kw) {
				t.Errorf("Auriok Champion has %s", kw)
			}
		}
		life := me.Life
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("3/3 green Beast"), 1) })
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := life
		if accept {
			want++
		}
		if me.Life != want {
			t.Errorf("accept=%v: life %d, want %d", accept, me.Life, want)
		}
	}
}

// --- Auriok Salvagers --------------------------------------------------------

func TestDn5AuriokSalvagersReturnsACheapArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	salvagers := mrd2Catalog(g, me.ID, "Auriok Salvagers", "Creature — Human Soldier", dn5AuriokSalvagersOracle, "{3}{W}", 2, 4)
	cheap := dn5GraveyardCard(me, "Cheap Trinket", "Artifact", "{1}")
	dear := dn5GraveyardCard(me, "Dear Trinket", "Artifact", "{2}")

	b06AddMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, salvagers, 0, game.ActivateAbilityParams{Targets: cardRefs(dear), Strict: true}); err == nil {
		t.Fatal("a mana value 2 artifact card was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, salvagers, 0, game.ActivateAbilityParams{Targets: cardRefs(cheap), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !handHasID(me, cheap) {
		t.Error("the mana value 1 artifact card should return to hand")
	}
}

// --- Auriok Windwalker -------------------------------------------------------

func TestDn5AuriokWindwalkerAttachesAnEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := mrd2Catalog(g, me.ID, "Auriok Windwalker", "Creature — Human Wizard", dn5AuriokWindwalkerOracle, "{3}{W}", 2, 3)
	sword := seedEquipment(g, me.ID, "My Sword", "")
	bear := seedBear(g, me.ID)
	if !hasKeywordOnBattlefield(t, g, walker, "flying") {
		t.Error("Auriok Windwalker has flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sword}, {Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, sword); host.ID != bear {
		t.Errorf("the Equipment is attached to %+v, want the bear", host)
	}
	if !tappedForTest(t, g, walker) {
		t.Error("the Windwalker taps to pay the cost")
	}
}

// --- Avarice Totem -----------------------------------------------------------

func TestDn5AvariceTotemExchangesControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	totem := mrd2Catalog(g, me.ID, "Avarice Totem", "Artifact", dn5AvariceTotemOracle, "{1}", 0, 0)
	bear := seedBear(g, opp.ID)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")

	b06AddMana(me, "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, totem, 0, game.ActivateAbilityParams{Targets: cardRefs(land), Strict: true}); err == nil {
		t.Fatal("a land was accepted as the target nonland permanent")
	}
	if err := g.ActivateCatalogAbility(me.ID, totem, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, bear); got != me.ID {
		t.Errorf("the creature is controlled by %v, want the Totem's controller", got)
	}
	if got := controllerOf(t, g, totem); got != opp.ID {
		t.Errorf("the Totem is controlled by %v, want the creature's former controller", got)
	}
}

// --- Battered Golem ----------------------------------------------------------

func TestDn5BatteredGolemStaysTappedUntilAnArtifactEnters(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	golem := b12Push(g, owner.ID, "Battered Golem", "Artifact Creature — Golem", dn5BatteredGolemOracle, 3, 2)
	arnTap(t, g, golem)
	advanceToUpkeepOf(t, g, 1)
	if !arnTapped(t, g, golem) {
		t.Fatal("Battered Golem untapped in its controller's untap step")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(owner.ID, TokenCard("3/3 green Beast"), 1) })
	if b18TriggerPromptCount(g, owner.ID) != 0 {
		t.Error("a Beast is not an artifact")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(g.Seats[0].ID, TreasureToken(), 1) })
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	passPriorityAroundTable(t, g)
	if arnTapped(t, g, golem) {
		t.Error("the Golem may untap when an artifact enters, under anyone's control")
	}
}

// --- Beacon of Creation ------------------------------------------------------

func TestDn5BeaconOfCreationMakesAnInsectPerForestAndShufflesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")

	beacon := castCatalogSpell(t, g, "Beacon of Creation", "Sorcery", dn5BeaconOfCreationOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Insect", me.ID); got != 2 {
		t.Errorf("%d Insects, want one for each of my two Forests", got)
	}
	if !me.Library.Contains(beacon) {
		t.Error("the Beacon should be shuffled into its owner's library")
	}
	if me.Graveyard.Contains(beacon) {
		t.Error("the Beacon does not go to the graveyard")
	}
}

// --- Beacon of Destruction ---------------------------------------------------

func TestDn5BeaconOfDestructionDealsFiveAndShufflesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	beacon := castCatalogSpell(t, g, "Beacon of Destruction", "Instant", dn5BeaconOfDestructionOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("opponent's life %d, want %d", opp.Life, life-5)
	}
	if !me.Library.Contains(beacon) {
		t.Error("the Beacon should be shuffled into its owner's library")
	}
}

func TestDn5BeaconOfDestructionWithItsTargetGoneGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := seedBear(g, opp.ID)
	beacon := castCatalogSpell(t, g, "Beacon of Destruction", "Instant", dn5BeaconOfDestructionOracle, cardTarget(bear))
	p1Destroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(beacon) {
		t.Error("a Beacon that doesn't resolve is put into its owner's graveyard")
	}
}
