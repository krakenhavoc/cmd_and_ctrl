package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b341_test.go — Commander 2014 cards, S82 OLD batch 341 (#5553),
// played through the real catalog entries.

const (
	oldB341Afterlife    = "4c13e2b5-961a-4031-84b1-15bd19b94286"
	oldB341Requiem      = "59f10ddb-8287-4d2c-8c60-1080b108fa78"
	oldB341Tuktuk       = "754e227b-ff31-4b5c-b612-2221c2b44246"
	oldB341Vitriol      = "5b8b1f03-9e2e-40f2-97e0-c9fb8d780d56"
	oldB341Brave        = "9d73a51a-1c8e-40cd-8124-aeff406a0884"
	oldB341Distorting   = "e5387fc7-c3b3-4933-bffa-580ad793c7ab"
	oldB341Lashwrithe   = "066ba81e-b32d-4808-97ab-0b5f52d557e4"
	oldB341Scythe       = "a500b724-bf28-47a3-b434-6efdebbcc04c"
	oldB341Sifter       = "07a26388-773e-4614-a3de-f133c6694d59"
	oldB341ScrapMastery = "62dac700-49f7-4745-af67-37d2a81dcebc"
	oldB341Gnomes       = "54b5e429-7a44-480d-bea4-4f8eeb7449b5"
	oldB341Demon        = "dd2bbc90-0474-42b7-afe4-3655a120ab02"
)

func TestOldB341EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB341Afterlife: "Afterlife", oldB341Requiem: "Requiem Angel", oldB341Tuktuk: "Tuktuk the Explorer",
		oldB341Vitriol: "Wave of Vitriol", oldB341Brave: "Brave the Elements", oldB341Distorting: "Distorting Wake",
		oldB341Lashwrithe: "Lashwrithe", oldB341Scythe: "Strata Scythe", oldB341Sifter: "Grave Sifter",
		oldB341ScrapMastery: "Scrap Mastery", oldB341Gnomes: "Bottle Gnomes", oldB341Demon: "Xathrid Demon",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func oldB341Spirits(g *game.Game, who uuid.UUID) int {
	return effectstest.TokensNamedControlledBy(g, "Spirit", who)
}

func TestOldB341AfterlifeDestroysAndGivesTheControllerASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Afterlife", "Instant", oldB341Afterlife, effectstest.CardTarget(uuid.New())); err == nil {
		t.Fatal("a nonexistent target was accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Afterlife", "Instant", oldB341Afterlife, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature was not destroyed")
	}
	if got := oldB341Spirits(g, opp.ID); got != 1 {
		t.Errorf("the creature's controller has %d Spirits, want 1", got)
	}
	if got := oldB341Spirits(g, me.ID); got != 0 {
		t.Errorf("the caster got %d Spirits, want 0", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && (c.Power != 1 || c.Toughness != 1 || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "flying")) {
			t.Errorf("Spirit token = %+v, want a 1/1 flier", c)
		}
	}
}

func TestOldB341AfterlifeGivesTheSpiritEvenIfTheTargetIsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.MirB1Push(g, opp.ID, "Darksteel Bear", "Artifact Creature — Bear", "", 2, 2, "indestructible")
	effectstest.CastCatalogSpell(t, g, "Afterlife", "Instant", oldB341Afterlife, effectstest.CardTarget(golem))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Error("an indestructible creature was destroyed")
	}
	if got := oldB341Spirits(g, opp.ID); got != 1 {
		t.Errorf("Spirits = %d, want 1", got)
	}
}

func TestOldB341RequiemAngelMakesASpiritForOtherNonSpiritCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	angel := effectstest.MirB1Push(g, me.ID, "Requiem Angel", "Creature — Angel", oldB341Requiem, 5, 5, "flying")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spirit := effectstest.B12Creature(g, me.ID, "Ghost", "Creature — Spirit", 1, 1)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)

	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB341Spirits(g, me.ID); got != 1 {
		t.Fatalf("after a Bear died: %d Spirits, want 1", got)
	}
	effectstest.KillCreature(t, g, me.ID, spirit)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB341Spirits(g, me.ID); got != 1 {
		t.Errorf("a Spirit dying made a token: %d Spirits, want 1", got)
	}
	effectstest.KillCreature(t, g, opp.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB341Spirits(g, me.ID); got != 1 {
		t.Errorf("an opponent's creature dying made a token: %d Spirits, want 1", got)
	}
	effectstest.KillCreature(t, g, me.ID, angel)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB341Spirits(g, me.ID); got != 1 {
		t.Errorf("the Angel's own death made a token: %d Spirits, want 1", got)
	}
}

func TestOldB341TuktukLeavesAFiveFiveLegendaryGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tuktuk := effectstest.MirB1Push(g, me.ID, "Tuktuk the Explorer", "Legendary Creature — Goblin", oldB341Tuktuk, 1, 1, "haste")
	effectstest.KillCreature(t, g, me.ID, tuktuk)
	effectstest.PassPriorityAroundTable(t, g)
	var found *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Tuktuk the Returned" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("no Tuktuk the Returned token")
	}
	if found.Controller != me.ID || !effects.IsToken(*found) || found.Power != 5 || found.Toughness != 5 ||
		!found.HasSupertype("legendary") || !found.HasCardType("artifact") || !found.HasCardType("creature") ||
		!found.HasSubtype("Goblin") || !found.HasSubtype("Golem") || len(found.Colors) != 0 {
		t.Errorf("token = %+v", *found)
	}
}

func TestOldB341BottleGnomesSacrificesForThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gnomes := effectstest.MirB1Push(g, me.ID, "Bottle Gnomes", "Artifact Creature — Gnome", oldB341Gnomes, 1, 3)
	before := me.Life
	if err := effectstest.MirB1Activate(g, me.ID, gnomes, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(gnomes) {
		t.Error("the sacrifice was not paid as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life = %d, want %d", me.Life, before+3)
	}
}

func TestOldB341DistortingWakeReturnsExactlyXNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Permanent(g, me.ID, "B", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	keep := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{U}{U}{U}{C}{C}")
	if err := g.CastSpell(me.ID, effectstest.HandCardFull(me, "Distorting Wake", "Sorcery", "{X}{U}{U}{U}", oldB341Distorting, []string{"U"}),
		game.CastSpellParams{Strict: true, XValue: 1, Targets: effectstest.CardTargets(land)}); err == nil {
		t.Fatal("a land was accepted as a target")
	}
	if err := g.CastSpell(me.ID, effectstest.HandCardFull(me, "Distorting Wake", "Sorcery", "{X}{U}{U}{U}", oldB341Distorting, []string{"U"}),
		game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardTargets(a)}); err == nil {
		t.Fatal("one target was accepted for X=2")
	}
	effectstest.P1g2Cast(t, g, "Distorting Wake", "Sorcery", oldB341Distorting, "{X}{U}{U}{U}",
		game.CastSpellParams{XValue: 2, Targets: effectstest.CardTargets(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !opp.Hand.Contains(a) {
		t.Error("A did not return to its owner's hand")
	}
	if g.Battlefield.Contains(b) || !me.Hand.Contains(b) {
		t.Error("B did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(land) {
		t.Error("an untargeted permanent was returned")
	}
}

func oldB341Named(g *game.Game, name string, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB341BraveTheElementsGivesWhiteCreaturesYouControlProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	w1 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Knight", TypeLine: "Creature — Knight",
		Power: 2, Toughness: 2, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Foe", TypeLine: "Creature — Knight",
		Power: 2, Toughness: 2, Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Brave the Elements", "Instant", oldB341Brave, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, w1, game.ProtectionFromColor("R")) {
		t.Error("the white creature did not gain protection from red")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, green, game.ProtectionFromColor("R")) {
		t.Error("a nonwhite creature gained protection")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, game.ProtectionFromColor("R")) {
		t.Error("an opponent's white creature gained protection")
	}
}

func TestOldB341LashwritheIsALivingWeaponCountingYourSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	effectstest.B12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	lash := effectstest.B12PlayFromHand(t, g, "Lashwrithe", "Artifact — Equipment", oldB341Lashwrithe, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	germs := oldB341Named(g, "Phyrexian Germ", me.ID)
	if len(germs) != 1 {
		t.Fatalf("Germs = %d, want 1", len(germs))
	}
	germ := germs[0].InstanceID
	if host := effectstest.AttachmentHostOf(t, g, lash); host.ID != germ {
		t.Fatalf("Lashwrithe is attached to %v, want the Germ", host)
	}
	if got := effectstest.EffectivePower(t, g, germ); got != 2 {
		t.Errorf("Germ power = %d, want 2 (two of my Swamps)", got)
	}
	effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	if got := effectstest.EffectiveToughness(t, g, germ); got != 3 {
		t.Errorf("Germ toughness = %d, want 3", got)
	}
}

func TestOldB341LashwritheEquipPaysBlackManaOrLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lash := effectstest.MirB1Push(g, me.ID, "Lashwrithe", "Artifact — Equipment", oldB341Lashwrithe, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.DcMana(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, lash, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("equip with {B}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40 {
		t.Errorf("life = %d, want 40 when paid in mana", me.Life)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("equipped power = %d, want 3", got)
	}
	// Moving it again with an empty pool pays 2 life for each {B/P}.
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, lash, 0, game.ActivateAbilityParams{Strict: true, PhyrexianLife: 2, Targets: effectstest.CardTarget(other)}); err != nil {
		t.Fatalf("equip paying life: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 36 {
		t.Errorf("life = %d, want 36 after two Phyrexian payments", me.Life)
	}
}

func TestOldB341StrataScytheImprintsALandAndCountsLandsWithItsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	forest := game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	effectstest.PushLibraryCardForTest(me, forest)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	scythe := effectstest.B12PlayFromHand(t, g, "Strata Scythe", "Artifact — Equipment", oldB341Scythe, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Name == "Forest" {
			exiled++
		}
	}
	if exiled != 1 {
		t.Fatalf("exiled Forests = %d, want 1", exiled)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, scythe, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power = %d, want 4 (two Forests on the battlefield, either controller)", got)
	}
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if got := effectstest.EffectiveToughness(t, g, bear); got != 5 {
		t.Errorf("equipped toughness = %d, want 5 after a third Forest", got)
	}
}

func TestOldB341GraveSifterEachPlayerReturnsCardsOfAChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	g1 := effectstest.PushGraveyardCardTyped(me, "Goblin A", "Creature — Goblin")
	g2 := effectstest.PushGraveyardCardTyped(me, "Goblin B", "Creature — Goblin")
	elf := effectstest.PushGraveyardCardTyped(me, "Elf", "Creature — Elf")
	oppGoblin := effectstest.PushGraveyardCardTyped(opp, "Their Goblin", "Creature — Goblin")
	oppElf := effectstest.PushGraveyardCardTyped(opp, "Their Elf", "Creature — Elf")
	effectstest.B12PlayFromHand(t, g, "Grave Sifter", "Creature — Elemental Beast", oldB341Sifter, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	effectstest.AnswerChooseCards(t, g, me.ID, g1)
	effectstest.AnswerCreatureType(t, g, opp.ID, "Elf")
	effectstest.AnswerChooseCards(t, g, opp.ID, oppElf)
	for _, seat := range g.Seats[2:] {
		effectstest.AnswerCreatureType(t, g, seat.ID, "Wizard")
	}
	if !me.Hand.Contains(g1) {
		t.Error("the chosen Goblin did not return")
	}
	if !me.Graveyard.Contains(g2) || !me.Graveyard.Contains(elf) {
		t.Error("a card that was not picked left the graveyard")
	}
	if !opp.Hand.Contains(oppElf) || !opp.Graveyard.Contains(oppGoblin) {
		t.Error("the opponent's Elf did not return alone")
	}
}

func TestOldB341ScrapMasteryReturnsExiledArtifactsAndSacrificesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mineBack := effectstest.PushGraveyardCardTyped(me, "Old Relic", "Artifact")
	theirBack := effectstest.PushGraveyardCardTyped(opp, "Their Golem", "Artifact Creature — Golem")
	myBoard := effectstest.B12Permanent(g, me.ID, "My Gadget", "Artifact")
	theirBoard := effectstest.B12Permanent(g, opp.ID, "Their Gadget", "Artifact")
	other := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.P1g2Cast(t, g, "Scrap Mastery", "Sorcery", oldB341ScrapMastery, "{3}{R}{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(oldB341Named(g, "Old Relic", me.ID)) != 1 || len(oldB341Named(g, "Their Golem", opp.ID)) != 1 {
		t.Error("the exiled artifact cards did not return to their owners' battlefields")
	}
	if g.Battlefield.Contains(myBoard) || g.Battlefield.Contains(theirBoard) {
		t.Error("an artifact that was on the battlefield was not sacrificed")
	}
	if !me.Graveyard.Contains(myBoard) || !opp.Graveyard.Contains(theirBoard) {
		t.Error("the sacrificed artifacts are not in their graveyards")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a nonartifact was sacrificed")
	}
	_, _ = mineBack, theirBack
}

func TestOldB341WaveOfVitriolSacrificesAndFetchesBasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	artifact := effectstest.B12Permanent(g, me.ID, "Gadget", "Artifact")
	ench := effectstest.B12Permanent(g, me.ID, "Aura", "Enchantment")
	myNonbasic := effectstest.B12Permanent(g, me.ID, "Strip Mine", "Land")
	myBasic := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	oppNonbasic := effectstest.B12Permanent(g, opp.ID, "Karakas", "Legendary Land")
	oppNonbasic2 := effectstest.B12Permanent(g, opp.ID, "Rishadan Port", "Land")
	creature := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	}
	effectstest.P1g2Cast(t, g, "Wave of Vitriol", "Sorcery", oldB341Vitriol, "{5}{G}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{artifact, ench, myNonbasic, oppNonbasic, oppNonbasic2} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", id)
		}
	}
	if !g.Battlefield.Contains(myBasic) || !g.Battlefield.Contains(creature) {
		t.Error("a basic land or a creature was sacrificed")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	prompt := effectstest.SearchChoiceFor(g, opp.ID)
	if prompt == nil || len(prompt.SearchCards) < 2 {
		t.Fatalf("no two-card search prompt for the opponent: %+v", prompt)
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, prompt.SearchCards[0], prompt.SearchCards[1])
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB341Named(g, "Forest", me.ID)); got != 2 {
		t.Errorf("my Forests = %d, want 2", got)
	}
	plains := oldB341Named(g, "Plains", opp.ID)
	if len(plains) != 2 {
		t.Fatalf("their Plains = %d, want 2", len(plains))
	}
	for _, c := range plains {
		if !c.Tapped {
			t.Error("a fetched land entered untapped")
		}
	}
}

func TestOldB341XathridDemonDrainsForTheSacrificedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	demon := effectstest.MirB1Push(g, me.ID, "Xathrid Demon", "Creature — Demon", oldB341Demon, 7, 7, "flying", "trample")
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 3, 3)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerSacrifice(t, g, me.ID, snack)
	if g.Battlefield.Contains(snack) || !g.Battlefield.Contains(demon) {
		t.Error("the Snack was not sacrificed or the Demon went with it")
	}
	for _, opp := range g.Seats[1:] {
		if opp.Life != 37 {
			t.Errorf("%s life = %d, want 37", opp.ID, opp.Life)
		}
	}
	if me.Life != 40 {
		t.Errorf("my life = %d, want 40", me.Life)
	}
}

func TestOldB341XathridDemonTapsAndHurtsYouWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	demon := effectstest.MirB1Push(g, me.ID, "Xathrid Demon", "Creature — Demon", oldB341Demon, 7, 7, "flying", "trample")
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != 33 {
		t.Errorf("life = %d, want 33", me.Life)
	}
	if !effectstest.TappedOnBattlefield(t, g, demon) {
		t.Error("the Demon did not tap")
	}
	for _, opp := range g.Seats[1:] {
		if opp.Life != 40 {
			t.Errorf("an opponent lost life: %d", opp.Life)
		}
	}
}
