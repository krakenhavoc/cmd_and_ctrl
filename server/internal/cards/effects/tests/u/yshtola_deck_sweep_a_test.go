package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestYsAFinalJudgmentExilesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Final Judgment", "Sorcery", effectstest.YsAFinalJudgmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		if effects.OnBattlefield(g, id) || !effectstest.InExile(g, id) {
			t.Errorf("creature %s was not exiled", id)
		}
	}
	if !effects.OnBattlefield(g, rock) {
		t.Error("a noncreature permanent was exiled")
	}
}

func TestYsAMesmericGlareCountersOnlySmallSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.CastWithCost(t, g, "Big Spell", "Sorcery", "{3}{U}", "")
	sprite := effectstest.YsAAdventure(me.ID, effectstest.YsASpriteOracle,
		game.Face{Name: "Hypnotic Sprite", TypeLine: "Creature — Faerie", ManaCost: "{U}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
		game.Face{Name: "Mesmeric Glare", TypeLine: "Instant — Adventure", ManaCost: "{2}{U}", Colors: []string{"U"}})
	glare := sprite.InstanceID
	me.Hand.PushTop(sprite)
	if err := g.CastSpell(me.ID, glare, game.CastSpellParams{Face: 1, Targets: effectstest.B16TargetCard(big)}); err == nil {
		t.Fatal("a mana value 4 spell was accepted as the target of Mesmeric Glare")
	}
	effectstest.PassPriorityAroundTable(t, g)

	small := effectstest.CastWithCost(t, g, "Small Spell", "Sorcery", "{1}{U}", "")
	if err := g.CastSpell(me.ID, glare, game.CastSpellParams{Face: 1, Targets: effectstest.B16TargetCard(small)}); err != nil {
		t.Fatalf("Mesmeric Glare on a mana value 2 spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(small) {
		t.Error("the small spell was not countered")
	}
	if !effectstest.InExile(g, glare) {
		t.Error("Hypnotic Sprite did not go on an adventure in exile")
	}
}

func TestYsASwiftEndDestroysAndCostsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	before := me.Life
	riderCard := effectstest.YsAAdventure(me.ID, effectstest.YsARiderOracle,
		game.Face{Name: "Murderous Rider", TypeLine: "Creature — Zombie Knight", ManaCost: "{1}{B}{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
		game.Face{Name: "Swift End", TypeLine: "Instant — Adventure", ManaCost: "{1}{B}{B}", Colors: []string{"B"}})
	rider := riderCard.InstanceID
	me.Hand.PushTop(riderCard)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, rider, game.CastSpellParams{Face: 1, Targets: effectstest.B16TargetCard(victim)}); err != nil {
		t.Fatalf("Swift End: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Error("the target survived Swift End")
	}
	if me.Life != before-2 {
		t.Errorf("life %d, want %d", me.Life, before-2)
	}
}

func TestYsAMurderousRiderGoesToTheBottomWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rider := effectstest.B12Push(g, me.ID, "Murderous Rider", "Creature — Zombie Knight", effectstest.YsARiderOracle, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rider) })
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, rider) || me.Graveyard.Contains(rider) {
		t.Fatal("the Rider should be neither on the battlefield nor in the graveyard")
	}
	bottom, err := me.Library.Bottom()
	if err != nil || bottom.InstanceID != rider {
		t.Errorf("bottom of the library = %v (err %v), want the Rider", bottom.Name, err)
	}
}

func TestYsATorrentialGearhulkCastsAnInstantForFreeAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Old Instant", TypeLine: "Instant", ManaCost: "{3}{U}{U}",
		Owner: me.ID, Controller: me.ID})
	sorcery := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Old Sorcery", TypeLine: "Sorcery", ManaCost: "{1}",
		Owner: me.ID, Controller: me.ID})

	effectstest.CastCatalogSpell(t, g, "Torrential Gearhulk", "Artifact Creature — Construct", effectstest.YsAGearhulkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("a sorcery was castable from the graveyard")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the instant for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) || !effectstest.InExile(g, bolt) {
		t.Error("the instant did not go to exile instead of the graveyard")
	}
}

func TestYsALyseHextHasDoubleStrikeAfterTwoNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lyse := effectstest.B12Push(g, me.ID, "Lyse Hext", "Legendary Creature — Human Rebel Monk", effectstest.YsALyseOracle, 2, 2)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, lyse), "double strike") {
		t.Fatal("double strike before any spell")
	}
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, lyse), "double strike") {
		t.Error("double strike after one noncreature spell")
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, lyse), "double strike") {
		t.Error("no double strike after two noncreature spells")
	}
	effectstest.PassPriorityAroundTable(t, g)
	v := effectstest.S58p6View(t, g, lyse)
	if v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
		t.Errorf("Lyse is %d/%d after two prowess triggers, want 4/4", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestYsAHraesvelgrGivesATargetPlusOneAndUnblockableOnEntryAndCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushSizedCreature(g, me.ID, "Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Hraesvelgr of the First Brood", "Legendary Creature — Elder Dragon", effectstest.YsAHraesvelgrOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	powers := effectstest.S58p6View(t, g, bear).CurrentPower()
	dragon := effectstest.FindOnBattlefieldByOracle(g, effectstest.YsAHraesvelgrOracle)
	if dragon == uuid.Nil {
		t.Fatal("Hraesvelgr did not enter")
	}
	dv := effectstest.S58p6View(t, g, dragon)
	if powers+dv.CurrentPower() != 2+0+1 {
		t.Errorf("one creature should have +1/+0: bear+dragon power %d, want 3", powers+dv.CurrentPower())
	}

	effectstest.B24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	total := effectstest.S58p6View(t, g, bear).CurrentPower() + effectstest.S58p6View(t, g, dragon).CurrentPower()
	if total != 2+0+2 {
		t.Errorf("a noncreature cast should add another +1/+0: total %d, want 4", total)
	}
}

func TestYsAThancredGuardsALegendaryPermanentWhileYouControlHim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.B12Push(g, me.ID, "Ally", "Legendary Creature — Human", "", 2, 2)
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	thancred := effectstest.CastCatalogSpell(t, g, "Thancred Waters", "Legendary Creature — Human Warrior", effectstest.YsAThancredOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, ally), "indestructible") {
		t.Fatal("the other legendary permanent did not gain indestructible")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, thancred), "indestructible") {
		t.Error("Thancred is indestructible before a noncreature spell")
	}
	effectstest.B24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, thancred), "indestructible") {
		t.Error("Thancred did not gain indestructible from the spell")
	}
}

func TestYsAArdbertCountersLegendaryCreaturesByColourOfSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ardbert := effectstest.B12Push(g, me.ID, "Ardbert, Warrior of Darkness", "Legendary Creature — Spirit Warrior", effectstest.YsAArdbertOracle, 2, 2)
	other := effectstest.B12Push(g, me.ID, "Other Legend", "Legendary Creature — Human", "", 1, 1)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToMain(t, g)
	white := effectstest.HandCardFull(me, "White Spell", "Instant", "{W}", "", []string{"W"})
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ardbert, other} {
		if effectstest.CountersOn(g, id, game.CounterPlusOne) != 1 {
			t.Errorf("legendary creature %s: %d counters, want 1", id, effectstest.CountersOn(g, id, game.CounterPlusOne))
		}
		if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "vigilance") {
			t.Errorf("legendary creature %s did not gain vigilance", id)
		}
	}
	if effectstest.CountersOn(g, plain, game.CounterPlusOne) != 0 {
		t.Error("a nonlegendary creature got a counter")
	}

	gold := effectstest.HandCardFull(me, "Gold Spell", "Instant", "{W}{B}", "", []string{"W", "B"})
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityUntilTriggerOrderPrompt(t, g)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ardbert, game.CounterPlusOne); got != 3 {
		t.Errorf("a white-black spell triggers both abilities: %d counters, want 3", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, other), "menace") {
		t.Error("the black trigger did not grant menace")
	}
}

func TestYsAPapalymoDrainsOnNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Papalymo Totolymo", "Legendary Creature — Dwarf Wizard", effectstest.YsAPapalymoOracle, 1, 2)
	before := effectstest.B24Lives(g)
	effectstest.B24CastNoncreature(t, g, "Spell", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	after := effectstest.B24Lives(g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	papalymo := effectstest.B12Push(g, me.ID, "Papalymo Totolymo", "Legendary Creature — Dwarf Wizard", effectstest.YsAPapalymoOracle, 1, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	calm := effectstest.B12Creature(g, g.Seats[2].ID, "Calm", "Creature — Ogre", 5, 5)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })

	if err := g.ActivateCatalogAbility(me.ID, papalymo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, papalymo) {
		t.Error("Papalymo was not sacrificed")
	}
	if effects.OnBattlefield(g, big) || !effects.OnBattlefield(g, small) {
		t.Error("the opponent who lost life must sacrifice only their greatest-power creature")
	}
	if !effects.OnBattlefield(g, calm) {
		t.Error("an opponent who lost no life sacrificed a creature")
	}
}

func TestYsATomeOfLegendsCountsYourCommanderAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tome := effectstest.CastCatalogSpell(t, g, "Tome of Legends", "Artifact — Book", effectstest.YsATomeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, tome, "page"); got != 1 {
		t.Fatalf("page counters %d, want 1 as it enters", got)
	}
	cmdr := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cmdr, Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Owner: me.ID, Controller: me.ID, IsCommander: true})
	if err := g.CastSpell(me.ID, cmdr, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the commander: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, tome, "page"); got != 2 {
		t.Fatalf("page counters %d, want 2 after the commander entered", got)
	}

	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("draw ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || effectstest.CountersOn(g, tome, "page") != 1 {
		t.Errorf("hand %d→%d, page counters %d", hand, me.Hand.Size(), effectstest.CountersOn(g, tome, "page"))
	}
}
