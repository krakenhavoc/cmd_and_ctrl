package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std_b55_cards_test.go — Wilds of Eldraine, Standard leftovers batches
// 55–58 (#4471). Each test plays the real catalog entry.

const (
	b55Graceful   = "b580c6d8-f501-4623-843a-5bb2b638a339"
	b55Hollow     = "4fad3816-41af-4fa4-85b4-ba37614ec848"
	b55Galefang   = "f12c2b65-4a04-4fab-ae77-973fe5cc823d"
	b55Huntsman   = "ae63fd5c-46ff-41fb-a1f6-6259336525c1"
	b55Redtooth   = "9e459e48-d2bb-4c4e-9a87-58efb56cddd1"
	b55Folly      = "f45542a0-0d49-4b56-9c3e-663ff8d0529a"
	b55Johann     = "8068f80f-88ee-4c0b-a0fb-3a6337f83f22"
	b55Looter     = "e04560ea-f2bd-4cb7-8f65-3aa0dd58fbdf"
	b55Rowan      = "333c15c6-c8fd-4cdd-89d2-06b921edeb8b"
	b55Sharae     = "c736935a-e006-4a67-b942-e8ccf1cd9abb"
	b56Talion     = "a019e34d-00c4-4687-9d41-fe550bf1dbdc"
	b56Beluna     = "06374e6d-62fa-4833-81cd-b4d0a35e7365"
	b56Callous    = "0e7405e3-e9e6-4213-a73f-e0c01a7d6721"
	b56Sugarmaw   = "c8bc28ea-5630-49d5-963a-5789fa83f3c7"
	b56Duelist    = "3ca07da2-29a1-40ce-a7f3-4838e95ca1d9"
	b56Ruiner     = "334fba11-e500-40fd-a142-41ff553642b5"
	b56Druid      = "555df32f-f8e9-49af-bc4a-1194ea325a97"
	b56Candy      = "7c8e4f29-53e0-4242-8ef8-7a9476dce178"
	b56Apple      = "b26a2875-c704-491a-b82d-07c316e53611"
	b56Inn        = "4dfd33e8-7e30-493b-8564-f7df5f0257aa"
	b57Bramble    = "68f36dfa-bbd2-4860-b09b-37e35136be72"
	b57Will       = "4ecb8653-db8c-468a-853d-420ff62a7a8a"
	b57Pixie      = "51bb6d3b-be6a-435d-8a80-0c0f0161892d"
	b57Confect    = "87f82d15-5b6b-400f-8a15-d45a2a735f7d"
	b57Hooligan   = "c808892a-83d5-4b09-b0da-118b0d6fe58e"
	b57Farsight   = "7de66ba3-6e97-4e0a-928e-7ce45dee278b"
	b57End        = "835c195a-8c35-4076-896d-8d2972cac270"
	b57Prank      = "052e8e1c-d006-4c9f-90da-deafbbf7fed3"
	b57Specter    = "caff40b1-28d1-4ef6-90d1-51a1dc06cbcd"
	b57Colony     = "3f01c327-ba50-4fc3-a930-fa1fb02e34b9"
	b58Goddric    = "35b2b172-5fbc-4e4d-a011-390ee1b6c848"
	b58Redcap     = "ada48353-4c86-44d6-83a5-e7e3c703b948"
	b58Feral      = "5a8196ef-24ee-40b6-97d1-8a2257180f86"
	b58Triplets   = "32de1b70-f993-4423-b0ef-0762c2b4d1b1"
	b58Sentinel   = "683775ac-481e-4c26-ab2f-5d2813af0268"
	b58Debut      = "07995092-af6c-43b8-b30b-17e020ead37c"
	b58Agatha     = "f398d222-f8e7-4983-9a77-267adef7f551"
	b58Hylda      = "1733c71b-fd28-48dd-8a3d-41afd4b22f0e"
	b58Yenna      = "0adede77-5181-4190-8066-4797cefe0189"
	b58HyldaCrown = "a8291d04-240e-456b-8c86-1fb4eb460476"
)

func TestB55EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b55Graceful: "Graceful Takedown", b55Hollow: "Hollow Scavenger", b55Hollow + "#1": "Bakery Raid",
		b55Galefang: "Howling Galefang", b55Huntsman: "The Huntsman's Redemption", b55Redtooth: "Redtooth Vanguard",
		b55Folly: "The Apprentice's Folly", b55Johann: "Johann, Apprentice Sorcerer", b55Looter: "Likeness Looter",
		b55Rowan: "Rowan, Scion of War", b55Sharae: "Sharae of Numbing Depths",
		b56Talion: "Talion, the Kindly Lord", b56Beluna: "Beluna Grandsquall", b56Beluna + "#1": "Seek Thrills",
		b56Callous: "Callous Sell-Sword", b56Callous + "#1": "Burn Together",
		b56Sugarmaw: "Devouring Sugarmaw", b56Sugarmaw + "#1": "Have for Dinner",
		b56Duelist: "Heartflame Duelist", b56Duelist + "#1": "Heartflame Slash",
		b56Ruiner: "Picnic Ruiner", b56Ruiner + "#1": "Stolen Goodies",
		b56Druid: "Questing Druid", b56Druid + "#1": "Seek the Beast",
		b56Candy: "Candy Trail", b56Apple: "Eriette's Tempting Apple", b56Inn: "Edgewall Inn",
		b57Bramble: "Bramble Familiar", b57Bramble + "#1": "Fetch Quest",
		b57Will: "Will, Scion of Peace", b57Pixie: "Storyteller Pixie", b57Confect: "Experimental Confectioner",
		b57Hooligan: "Charging Hooligan", b57Farsight: "Farsight Ritual", b57End: "The End",
		b57Prank: "Rankle's Prank", b57Specter: "Specter of Mortality", b57Colony: "Tangled Colony",
		b58Goddric: "Goddric, Cloaked Reveler", b58Redcap: "Redcap Gutter-Dweller", b58Feral: "Feral Encounter",
		b58Triplets: "Gruff Triplets", b58Sentinel: "Sentinel of Lost Lore", b58Debut: "Thunderous Debut",
		b58Agatha: "Agatha of the Vile Cauldron", b58Hylda: "Hylda of the Icy Crown",
		b58Yenna: "Yenna, Redtooth Regent", b58HyldaCrown: "Hylda's Crown of Winter",
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

// --- helpers ------------------------------------------------------------------

var (
	b55AdvFaceCreature = game.Face{Name: "Adv Creature", TypeLine: "Creature — Bear", ManaCost: "{1}", Colors: []string{"G"}, Power: 2, Toughness: 2}
	b55AdvFaceSpell    = game.Face{Name: "Adv Spell", TypeLine: "Sorcery — Adventure", ManaCost: "{G}", Colors: []string{"G"}}
)

// b55AdventurerCard is an adventure-layout card (creature face 0, Adventure
// face 1) owned by owner, not yet placed anywhere.
func b55AdventurerCard(owner uuid.UUID, name string) game.Card {
	creature, spell := b55AdvFaceCreature, b55AdvFaceSpell
	creature.Name, spell.Name = name, name+" Spell"
	c := game.Card{InstanceID: uuid.New(), OracleID: "adv-" + name, Name: name, Layout: game.LayoutAdventure,
		Owner: owner, Controller: owner, Faces: []game.Face{creature, spell}}
	c.SetFace(0)
	return c
}

// b55ExileAdventurer puts an adventurer card owned by owner into exile.
func b55ExileAdventurer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	c := b55AdventurerCard(owner, name)
	g.Exile.PushTop(c)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: c.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneExile})
	})
	return c.InstanceID
}

func b55Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return containsString(effectiveAbilities(t, g, id), kw)
}

func b55HandSize(g *game.Game, p *game.Player) int { return len(p.Hand.Cards) }

// --- Howling Galefang ---------------------------------------------------------

func TestB55HowlingGalefangHasHasteOnlyWhileYouOwnAnExiledAdventurer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	fang := b12Push(g, me.ID, "Howling Galefang", "Creature — Beast", b55Galefang, 3, 3)
	if !b55Has(t, g, fang, "vigilance") {
		t.Error("no vigilance")
	}
	if b55Has(t, g, fang, "haste") {
		t.Fatal("haste with nothing in exile")
	}
	b55ExileAdventurer(g, opp.ID, "Theirs")
	if b55Has(t, g, fang, "haste") {
		t.Error("an opponent's exiled Adventure card gave it haste")
	}
	b55ExileAdventurer(g, me.ID, "Mine")
	if !b55Has(t, g, fang, "haste") {
		t.Error("an exiled Adventure card you own did not give it haste")
	}
}

// --- Hollow Scavenger // Bakery Raid ------------------------------------------

func TestB55HollowScavengerPumpsOnceEachTurnSacrificingAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	hollow := b12Push(g, me.ID, "Hollow Scavenger", "Creature — Wolf", b55Hollow, 3, 2)
	b12Permanent(g, me.ID, "Food", "Token Artifact — Food")
	b12Permanent(g, me.ID, "Food", "Token Artifact — Food")
	advanceTo(t, g, game.StepPrecombatMain)
	foods := []uuid.UUID{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Food" {
			foods = append(foods, c.InstanceID)
		}
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hollow, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: foods[:1]}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, hollow); p != 5 || tough != 4 {
		t.Errorf("Scavenger is %d/%d, want 5/4 (foods left %d)", p, tough, b27CountNamed(g, me.ID, "Food"))
	}
	if err := g.ActivateCatalogAbility(me.ID, hollow, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: foods[1:2]}); err == nil {
		t.Error("a second activation in one turn was allowed")
	}
}

func TestB55BakeryRaidMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, b55Hollow,
		hobA6Face("Hollow Scavenger", "Creature — Wolf", "{2}{G}", 3, 2, "G"),
		hobA6Face("Bakery Raid", "Sorcery — Adventure", "{G}", 0, 0, "G"))
	hobA6MustCast(t, g, me, id, 1, nil)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Foods, want 1", got)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

// --- Callous Sell-Sword // Burn Together --------------------------------------

func TestB56CallousSellSwordEntersWithACounterPerCreatureThatDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	mine1 := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	mine2 := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "C", "Creature — Bear", 1, 1)
	for _, id := range []uuid.UUID{mine1, mine2, theirs} {
		woeB1Destroy(t, g, id)
	}
	sword := castCatalogSpell(t, g, "Callous Sell-Sword", "Creature — Human Soldier", b56Callous, nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, sword, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 (the two that died under your control)", got)
	}
}

func TestB56BurnTogetherDealsPowerDamageThenSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	life := lifeOf(g, opp.ID)
	_, id := hobA6HandAdventure(t, g, b56Callous,
		hobA6Face("Callous Sell-Sword", "Creature — Human Soldier", "{1}{B}", 2, 2, "B"),
		hobA6Face("Burn Together", "Sorcery — Adventure", "{R}", 0, 0, "R"))
	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{hobA6Card(big), {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}})
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent at %d, want %d", got, life-4)
	}
	if g.Battlefield.Contains(big) {
		t.Error("the creature was not sacrificed")
	}
}

// --- Heartflame Duelist // Heartflame Slash -----------------------------------

func TestB56HeartflameSlashDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	life := lifeOf(g, opp.ID)
	_, id := hobA6HandAdventure(t, g, b56Duelist,
		hobA6Face("Heartflame Duelist", "Creature — Human Knight", "{1}{W}", 3, 1, "W"),
		hobA6Face("Heartflame Slash", "Instant — Adventure", "{2}{R}", 0, 0, "R"))
	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent at %d, want %d", got, life-3)
	}
}

func TestB56HeartflameDuelistGivesYourInstantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	b12Push(g, me.ID, "Heartflame Duelist", "Creature — Human Knight", b56Duelist, 3, 1)
	mine := lifeOf(g, me.ID)
	theirs := lifeOf(g, opp.ID)
	_, id := hobA6HandAdventure(t, g, "bolt-adv",
		hobA6Face("X", "Creature — Bear", "{1}", 1, 1, "R"),
		hobA6Face("Bolt", "Instant — Adventure", "{R}", 0, 0, "R"))
	_ = id
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != theirs-3 {
		t.Fatalf("opponent at %d, want %d", got, theirs-3)
	}
	if got := lifeOf(g, me.ID); got != mine+3 {
		t.Errorf("you are at %d, want %d (lifelink)", got, mine+3)
	}
}

// --- Picnic Ruiner // Stolen Goodies ------------------------------------------

func TestB56StolenGoodiesDistributesThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	_, id := hobA6HandAdventure(t, g, b56Ruiner,
		hobA6Face("Picnic Ruiner", "Creature — Goblin Rogue", "{1}{R}", 2, 2, "R"),
		hobA6Face("Stolen Goodies", "Sorcery — Adventure", "{3}{G}", 0, 0, "G"))
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1,
		Targets:      []game.TargetRef{hobA6Card(a), hobA6Card(b)},
		Distribution: map[uuid.UUID]int{a: 2, b: 1}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, a, game.CounterPlusOne) != 2 || countersOn(g, b, game.CounterPlusOne) != 1 {
		t.Errorf("counters %d/%d, want 2/1", countersOn(g, a, game.CounterPlusOne), countersOn(g, b, game.CounterPlusOne))
	}
}

func TestB56PicnicRuinerGainsDoubleStrikeWithAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	ruiner := b12Push(g, me.ID, "Picnic Ruiner", "Creature — Goblin Rogue", b56Ruiner, 2, 2)
	b12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ruiner, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !b55Has(t, g, ruiner, "double strike") {
		t.Error("no double strike with a power-4 creature")
	}
}

// --- Questing Druid // Seek the Beast -----------------------------------------

func TestB56QuestingDruidGrowsForWhiteBlueBlackOrRedSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	druid := b12Push(g, me.ID, "Questing Druid", "Creature — Human Druid", b56Druid, 1, 1)
	castCatalogSpell(t, g, "Green Spell", "Instant", "g-spell", nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, druid, game.CounterPlusOne); got != 0 {
		t.Fatalf("a colourless spell gave %d counters", got)
	}
}

func TestB56SeekTheBeastExilesTwoYouMayPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	lib := len(me.Library.Cards)
	_, id := hobA6HandAdventure(t, g, b56Druid,
		hobA6Face("Questing Druid", "Creature — Human Druid", "{1}{G}", 1, 1, "G"),
		hobA6Face("Seek the Beast", "Instant — Adventure", "{1}{R}", 0, 0, "R"))
	hobA6MustCast(t, g, me, id, 1, nil)
	if got := len(me.Library.Cards); got != lib-2 {
		t.Errorf("library %d -> %d, want two exiled", lib, got)
	}
}
