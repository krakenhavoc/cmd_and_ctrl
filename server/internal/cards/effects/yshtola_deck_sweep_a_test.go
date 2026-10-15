package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// yshtola_deck_sweep_a_test.go — the first slice of the cards the
// "Y'shtola, Night's Blessed" deck request asked for (#3657): Final
// Judgment, Hypnotic Sprite // Mesmeric Glare, Murderous Rider // Swift
// End, Torrential Gearhulk, Lyse Hext, Hraesvelgr of the First Brood,
// Thancred Waters, Ardbert, Papalymo Totolymo and Tome of Legends.

const (
	ysAFinalJudgmentOracle = "69ef9308-bc25-4f3c-a5fd-1ebaed6e4842"
	ysASpriteOracle        = "1f9efad1-9f51-47aa-a5a7-9c8103435a01"
	ysARiderOracle         = "1080c5b5-6651-4c6a-93e6-099fbe389e26"
	ysAGearhulkOracle      = "a6395447-677d-4c39-8eda-2d57e527c94e"
	ysALyseOracle          = "47bce63e-8edb-43c1-b273-4b86da5c616b"
	ysAHraesvelgrOracle    = "44cbf70c-61cf-473e-901d-039956ced642"
	ysAThancredOracle      = "dc36f81e-d1b0-44a4-bf82-4cd276bc0812"
	ysAArdbertOracle       = "a9243760-b9b1-4d43-8550-3528b5571dae"
	ysAPapalymoOracle      = "334e83cd-b513-4bd4-89a2-86875667e6c4"
	ysATomeOracle          = "cada2a1c-5db2-4702-9a57-cbe7da1bf208"
)

// ysAAdventure builds an Adventure card the way deck.toGameCard does.
func ysAAdventure(owner uuid.UUID, oracle string, creature, adventure game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutAdventure,
		Owner: owner, Controller: owner, Faces: []game.Face{creature, adventure},
	}
	c.SetFace(0)
	return c
}

func TestYsAFinalJudgmentExilesEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Final Judgment", "Sorcery", ysAFinalJudgmentOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		if onBattlefield(g, id) || !inExile(g, id) {
			t.Errorf("creature %s was not exiled", id)
		}
	}
	if !onBattlefield(g, rock) {
		t.Error("a noncreature permanent was exiled")
	}
}

func TestYsAMesmericGlareCountersOnlySmallSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := castWithCost(t, g, "Big Spell", "Sorcery", "{3}{U}", "")
	sprite := ysAAdventure(me.ID, ysASpriteOracle,
		game.Face{Name: "Hypnotic Sprite", TypeLine: "Creature — Faerie", ManaCost: "{U}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
		game.Face{Name: "Mesmeric Glare", TypeLine: "Instant — Adventure", ManaCost: "{2}{U}", Colors: []string{"U"}})
	glare := sprite.InstanceID
	me.Hand.PushTop(sprite)
	if err := g.CastSpell(me.ID, glare, game.CastSpellParams{Face: 1, Targets: b16TargetCard(big)}); err == nil {
		t.Fatal("a mana value 4 spell was accepted as the target of Mesmeric Glare")
	}
	passPriorityAroundTable(t, g)

	small := castWithCost(t, g, "Small Spell", "Sorcery", "{1}{U}", "")
	if err := g.CastSpell(me.ID, glare, game.CastSpellParams{Face: 1, Targets: b16TargetCard(small)}); err != nil {
		t.Fatalf("Mesmeric Glare on a mana value 2 spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(small) {
		t.Error("the small spell was not countered")
	}
	if !inExile(g, glare) {
		t.Error("Hypnotic Sprite did not go on an adventure in exile")
	}
}

func TestYsASwiftEndDestroysAndCostsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	before := me.Life
	riderCard := ysAAdventure(me.ID, ysARiderOracle,
		game.Face{Name: "Murderous Rider", TypeLine: "Creature — Zombie Knight", ManaCost: "{1}{B}{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
		game.Face{Name: "Swift End", TypeLine: "Instant — Adventure", ManaCost: "{1}{B}{B}", Colors: []string{"B"}})
	rider := riderCard.InstanceID
	me.Hand.PushTop(riderCard)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, rider, game.CastSpellParams{Face: 1, Targets: b16TargetCard(victim)}); err != nil {
		t.Fatalf("Swift End: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, victim) {
		t.Error("the target survived Swift End")
	}
	if me.Life != before-2 {
		t.Errorf("life %d, want %d", me.Life, before-2)
	}
}

func TestYsAMurderousRiderGoesToTheBottomWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rider := b12Push(g, me.ID, "Murderous Rider", "Creature — Zombie Knight", ysARiderOracle, 2, 3)
	castCatalogSpell(t, g, "Doom Blade", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rider) })
	passPriorityAroundTable(t, g)

	if onBattlefield(g, rider) || me.Graveyard.Contains(rider) {
		t.Fatal("the Rider should be neither on the battlefield nor in the graveyard")
	}
	bottom, err := me.Library.Bottom()
	if err != nil || bottom.InstanceID != rider {
		t.Errorf("bottom of the library = %v (err %v), want the Rider", bottom.Name, err)
	}
}

func TestYsATorrentialGearhulkCastsAnInstantForFreeAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Old Instant", TypeLine: "Instant", ManaCost: "{3}{U}{U}",
		Owner: me.ID, Controller: me.ID})
	sorcery := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Old Sorcery", TypeLine: "Sorcery", ManaCost: "{1}",
		Owner: me.ID, Controller: me.ID})

	castCatalogSpell(t, g, "Torrential Gearhulk", "Artifact Creature — Construct", ysAGearhulkOracle, nil)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)

	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("a sorcery was castable from the graveyard")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the instant for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) || !inExile(g, bolt) {
		t.Error("the instant did not go to exile instead of the graveyard")
	}
}

func TestYsALyseHextHasDoubleStrikeAfterTwoNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lyse := b12Push(g, me.ID, "Lyse Hext", "Legendary Creature — Human Rebel Monk", ysALyseOracle, 2, 2)
	if hasAbility(effectiveAbilities(t, g, lyse), "double strike") {
		t.Fatal("double strike before any spell")
	}
	b24CastNoncreature(t, g, "First", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if hasAbility(effectiveAbilities(t, g, lyse), "double strike") {
		t.Error("double strike after one noncreature spell")
	}
	b24CastNoncreature(t, g, "Second", "Instant", "{U}")
	if !hasAbility(effectiveAbilities(t, g, lyse), "double strike") {
		t.Error("no double strike after two noncreature spells")
	}
	passPriorityAroundTable(t, g)
	v := s58p6View(t, g, lyse)
	if v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
		t.Errorf("Lyse is %d/%d after two prowess triggers, want 4/4", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestYsAHraesvelgrGivesATargetPlusOneAndUnblockableOnEntryAndCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushSizedCreature(g, me.ID, "Bear", 2, 2)

	castCatalogSpell(t, g, "Hraesvelgr of the First Brood", "Legendary Creature — Elder Dragon", ysAHraesvelgrOracle, nil)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)

	powers := s58p6View(t, g, bear).CurrentPower()
	dragon := findOnBattlefieldByOracle(g, ysAHraesvelgrOracle)
	if dragon == uuid.Nil {
		t.Fatal("Hraesvelgr did not enter")
	}
	dv := s58p6View(t, g, dragon)
	if powers+dv.CurrentPower() != 2+0+1 {
		t.Errorf("one creature should have +1/+0: bear+dragon power %d, want 3", powers+dv.CurrentPower())
	}

	b24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)
	total := s58p6View(t, g, bear).CurrentPower() + s58p6View(t, g, dragon).CurrentPower()
	if total != 2+0+2 {
		t.Errorf("a noncreature cast should add another +1/+0: total %d, want 4", total)
	}
}

func TestYsAThancredGuardsALegendaryPermanentWhileYouControlHim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := b12Push(g, me.ID, "Ally", "Legendary Creature — Human", "", 2, 2)
	b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	thancred := castCatalogSpell(t, g, "Thancred Waters", "Legendary Creature — Human Warrior", ysAThancredOracle, nil)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)

	if !hasAbility(effectiveAbilities(t, g, ally), "indestructible") {
		t.Fatal("the other legendary permanent did not gain indestructible")
	}
	if hasAbility(effectiveAbilities(t, g, thancred), "indestructible") {
		t.Error("Thancred is indestructible before a noncreature spell")
	}
	b24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, thancred), "indestructible") {
		t.Error("Thancred did not gain indestructible from the spell")
	}
}

func TestYsAArdbertCountersLegendaryCreaturesByColourOfSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ardbert := b12Push(g, me.ID, "Ardbert, Warrior of Darkness", "Legendary Creature — Spirit Warrior", ysAArdbertOracle, 2, 2)
	other := b12Push(g, me.ID, "Other Legend", "Legendary Creature — Human", "", 1, 1)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	advanceToMain(t, g)
	white := handCardFull(me, "White Spell", "Instant", "{W}", "", []string{"W"})
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ardbert, other} {
		if countersOn(g, id, game.CounterPlusOne) != 1 {
			t.Errorf("legendary creature %s: %d counters, want 1", id, countersOn(g, id, game.CounterPlusOne))
		}
		if !hasAbility(effectiveAbilities(t, g, id), "vigilance") {
			t.Errorf("legendary creature %s did not gain vigilance", id)
		}
	}
	if countersOn(g, plain, game.CounterPlusOne) != 0 {
		t.Error("a nonlegendary creature got a counter")
	}

	gold := handCardFull(me, "Gold Spell", "Instant", "{W}{B}", "", []string{"W", "B"})
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityUntilTriggerOrderPrompt(t, g)
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ardbert, game.CounterPlusOne); got != 3 {
		t.Errorf("a white-black spell triggers both abilities: %d counters, want 3", got)
	}
	if !hasAbility(effectiveAbilities(t, g, other), "menace") {
		t.Error("the black trigger did not grant menace")
	}
}

func TestYsAPapalymoDrainsOnNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Papalymo Totolymo", "Legendary Creature — Dwarf Wizard", ysAPapalymoOracle, 1, 2)
	before := b24Lives(g)
	b24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	after := b24Lives(g)
	if after[0] != before[0]+1 {
		t.Errorf("you gain 1: %d → %d", before[0], after[0])
	}
	for i := 1; i < len(after); i++ {
		if after[i] != before[i]-1 {
			t.Errorf("seat %d: %d → %d, want 1 damage", i, before[i], after[i])
		}
	}
}

func TestYsAPapalymoSacrificesTheGreatestPowerCreatureOfOpponentsWhoLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	papalymo := b12Push(g, me.ID, "Papalymo Totolymo", "Legendary Creature — Dwarf Wizard", ysAPapalymoOracle, 1, 2)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	calm := b12Creature(g, g.Seats[2].ID, "Calm", "Creature — Ogre", 5, 5)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })

	if err := g.ActivateCatalogAbility(me.ID, papalymo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)

	if onBattlefield(g, papalymo) {
		t.Error("Papalymo was not sacrificed")
	}
	if onBattlefield(g, big) || !onBattlefield(g, small) {
		t.Error("the opponent who lost life must sacrifice only their greatest-power creature")
	}
	if !onBattlefield(g, calm) {
		t.Error("an opponent who lost no life sacrificed a creature")
	}
}

func TestYsATomeOfLegendsCountsYourCommanderAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tome := castCatalogSpell(t, g, "Tome of Legends", "Artifact — Book", ysATomeOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, tome, "page"); got != 1 {
		t.Fatalf("page counters %d, want 1 as it enters", got)
	}
	cmdr := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cmdr, Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Owner: me.ID, Controller: me.ID, IsCommander: true})
	if err := g.CastSpell(me.ID, cmdr, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the commander: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, tome, "page"); got != 2 {
		t.Fatalf("page counters %d, want 2 after the commander entered", got)
	}

	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("draw ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || countersOn(g, tome, "page") != 1 {
		t.Errorf("hand %d→%d, page counters %d", hand, me.Hand.Size(), countersOn(g, tome, "page"))
	}
}

// findOnBattlefieldByOracle returns the first battlefield card with the
// oracle id, or uuid.Nil.
func findOnBattlefieldByOracle(g *game.Game, oracle string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oracle {
			return c.InstanceID
		}
	}
	return uuid.Nil
}
