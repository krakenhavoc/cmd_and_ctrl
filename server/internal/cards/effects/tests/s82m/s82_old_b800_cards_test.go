package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 800 (Mirage, #5633): Telim'Tor's Edict, Jolt, Sapphire Charm,
// Reparations, Urborg Panther, Merfolk Seer, Infernal Contract and Zirilan of
// the Claw.

const (
	oldB800Panther  = "0678beaf-564c-4e98-b45f-a544925683a1"
	oldB800Edict    = "0e768d03-b2a3-4783-9cf5-7470f0203349"
	oldB800Jolt     = "13f69c70-dc93-4983-b649-d91b22633eb9"
	oldB800Seer     = "2351a9fa-17b5-41fc-93d0-34731d54f915"
	oldB800Charm    = "2ea2bf6c-a427-4d0c-b447-159959327697"
	oldB800Zirilan  = "3aeb28ac-8ed2-4a90-a631-d84361ac6acc"
	oldB800Repar    = "3cb86deb-f4e5-41c5-bfcd-614a24f3499c"
	oldB800Contract = "4372205f-45e6-49c7-b1fd-13a41e20d4de"
	oldB800Whirling = "6b37b0ef-7e89-447f-a189-b3ea7b42143f"
)

func TestOldB800EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB800Panther: "Urborg Panther", oldB800Edict: "Telim'Tor's Edict", oldB800Jolt: "Jolt",
		oldB800Seer: "Merfolk Seer", oldB800Charm: "Sapphire Charm", oldB800Zirilan: "Zirilan of the Claw",
		oldB800Repar: "Reparations", oldB800Contract: "Infernal Contract",
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

func oldB800Main(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// An empty hand keeps the turn's cleanup from asking for a discard.
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return me
}

// oldB800DrawsAtTheNextUpkeep asserts `p` draws one card at the next upkeep,
// and not before.
func oldB800DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Library.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOf(t, g, next)
	if p.Library.Size() != before {
		t.Fatalf("drew before the next upkeep's trigger resolved: library %d -> %d", before, p.Library.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p.Library.Size() != before-1 {
		t.Errorf("library %d -> %d at the next upkeep, want one card drawn", before, p.Library.Size())
	}
}

func TestOldB800TelimTorsEdictExilesYourPermanentAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB800Main(t, g)
	opp := g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	rock := effectstest.AcPermanent(g, me.ID, "Rock", "Artifact", 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Telim'Tor's Edict", "Instant", oldB800Edict, effectstest.LtCardTarget(theirs)); err == nil {
		t.Fatal("targeted a permanent the caster neither owns nor controls")
	}
	effectstest.CastCatalogSpell(t, g, "Telim'Tor's Edict", "Instant", oldB800Edict, effectstest.LtCardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !effectstest.InExile(g, rock) {
		t.Error("the Rock was not exiled")
	}
	oldB800DrawsAtTheNextUpkeep(t, g, me)
}

func TestOldB800TelimTorsEdictExilesAPermanentYouOwnButAnOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB800Main(t, g)
	opp := g.Seats[1]
	stolen := effectstest.AcPermanent(g, me.ID, "My Island", "Basic Land — Island", 0, 0)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, stolen).Controller = opp.ID })
	effectstest.CastCatalogSpell(t, g, "Telim'Tor's Edict", "Instant", oldB800Edict, effectstest.LtCardTarget(stolen))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, stolen) {
		t.Error("the stolen land the caster owns was not exiled")
	}
}

func TestOldB800JoltTapsUntapsOrLeavesAndDrawsAtTheNextUpkeep(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pick   int
		start  bool
		tapped bool
	}{{"tap", 0, false, true}, {"untap", 1, true, false}, {"leave", 2, true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := oldB800Main(t, g)
			opp := g.Seats[1]
			effectstest.SeedLibrary(me, "A", "B")
			land := effectstest.AcPermanent(g, opp.ID, "Island", "Basic Land — Island", 0, 0)
			if tc.start {
				g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Tapped = true })
			}
			effectstest.CastCatalogSpell(t, g, "Jolt", "Instant", oldB800Jolt, effectstest.LtCardTarget(land))
			effectstest.PassPriorityUntilChoice(t, g)
			effectstest.AnswerOptionPick(t, g, me.ID, tc.pick)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.PermanentTapped(g, land); got != tc.tapped {
				t.Errorf("tapped %v, want %v", got, tc.tapped)
			}
			oldB800DrawsAtTheNextUpkeep(t, g, me)
		})
	}
}

func TestOldB800JoltCannotTargetAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB800Main(t, g)
	walker := effectstest.AcPermanent(g, g.Seats[1].ID, "Jace", "Legendary Planeswalker — Jace", 0, 0)
	if err := effectstest.CastCatalogSpellErr(t, g, "Jolt", "Instant", oldB800Jolt, effectstest.LtCardTarget(walker)); err == nil {
		t.Error("targeted a planeswalker")
	}
}

func TestOldB800SapphireCharmModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB800Main(t, g)
	opp := g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C")
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.P1g3CastModal(t, g, "Sapphire Charm", "Instant", oldB800Charm, []int{1}, effectstest.P1g3ModeCard(mine, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
		t.Error("mode two did not give flying")
	}

	effectstest.P1g3CastModal(t, g, "Sapphire Charm", "Instant", oldB800Charm, []int{2}, effectstest.P1g3ModeCard(theirs, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PhasedOutInCatalogGame(g, theirs) {
		t.Error("mode three did not phase out their creature")
	}

	effectstest.P1g3CastModal(t, g, "Sapphire Charm", "Instant", oldB800Charm, []int{0},
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	oldB800DrawsAtTheNextUpkeep(t, g, opp)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
		return
	}
	t.Error("the flying outlasted the turn")
}

func TestOldB800SapphireCharmCannotPhaseOutYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB800Main(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	id := effectstest.HandCardOf(g, me, "Sapphire Charm", "Instant", "{U}", oldB800Charm)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{2}, Targets: []game.TargetRef{effectstest.P1g3ModeCard(mine, 0)}, Strict: true}); err == nil {
		t.Error("phased out a creature the caster controls")
	}
}

func TestOldB800ReparationsDrawsWhenAnOpponentTargetsYouOrYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB800Main(t, g)
	effectstest.SeedLibrary(opp, "A", "B", "C")
	effectstest.PushCatalogPermanent(g, opp.ID, "Reparations", "Enchantment", oldB800Repar, false)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// An opponent's (my) spell at the Reparations controller's creature: they may draw.
	before := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", oldB800Whirling, effectstest.LtCardTarget(theirBear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	// Declining draws nothing.
	before = opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", oldB800Whirling, effectstest.LtCardTarget(theirBear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 0 {
		t.Errorf("declined but drew %d", got)
	}
	// A spell aimed at someone else's creature does not trigger it.
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", oldB800Whirling, effectstest.LtCardTarget(myBear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, opp.ID) != nil {
		t.Error("a spell that targets another player's creature triggered Reparations")
	}
}

func TestOldB800ReparationsIgnoresYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB800Main(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Reparations", "Enchantment", oldB800Repar, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", oldB800Whirling, effectstest.LtCardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("the controller's own spell triggered Reparations")
	}
}

func TestOldB800UrborgPantherDestroysTheCreatureBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	panther := effectstest.B12Push(g, me.ID, "Urborg Panther", "Creature — Nightstalker Cat", oldB800Panther, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.CrtBlock(g, blocker, panther, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, panther, 0, effectstest.LtCardTarget(bystander), "B"); err == nil {
		t.Fatal("targeted a creature that is not blocking it")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, panther, 0, effectstest.LtCardTarget(blocker), "B"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(panther) {
		t.Error("the Panther was not sacrificed")
	}
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocker was not destroyed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that was not blocking died")
	}
}

func TestOldB800UrborgPantherSummonsSpiritOfTheNightForThreeCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	panther := effectstest.B12Push(g, me.ID, "Urborg Panther", "Creature — Nightstalker Cat", oldB800Panther, 2, 2)
	shadow := effectstest.B12Creature(g, me.ID, "Feral Shadow", "Creature — Nightstalker", 1, 1)
	breath := effectstest.B12Creature(g, me.ID, "Breathstealer", "Creature — Nightstalker", 1, 1)
	shadow2 := effectstest.B12Creature(g, me.ID, "Feral Shadow", "Creature — Nightstalker", 1, 1)
	bystander := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spirit := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spirit of the Night", TypeLine: "Creature — Avatar", Power: 6, Toughness: 4})
	other := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Other", TypeLine: "Creature — Avatar"})
	params := func(ids ...uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, SacrificeIDs: ids}
	}
	if err := g.ActivateCatalogAbility(me.ID, panther, 1, params(shadow, shadow2)); err == nil {
		t.Fatal("paid with two Feral Shadows and no Breathstealer")
	}
	if err := g.ActivateCatalogAbility(me.ID, panther, 1, params(shadow, bystander)); err == nil {
		t.Fatal("paid with a creature of another name")
	}
	if err := g.ActivateCatalogAbility(me.ID, panther, 1, params(shadow)); err == nil {
		t.Fatal("paid with only one of the two named creatures")
	}
	if err := g.ActivateCatalogAbility(me.ID, panther, 1, params(shadow, breath)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for name, id := range map[string]uuid.UUID{"Panther": panther, "Feral Shadow": shadow, "Breathstealer": breath} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not sacrificed", name)
		}
	}
	if !g.Battlefield.Contains(shadow2) || !g.Battlefield.Contains(bystander) {
		t.Error("a creature that was not paid was sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, spirit)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spirit) {
		t.Error("Spirit of the Night did not enter the battlefield")
	}
	if g.Battlefield.Contains(other) {
		t.Error("another card was fetched")
	}
}

func TestOldB800MerfolkSeerDrawsOnlyIfYouPayWhenItDies(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		oldB800Main(t, g)
		effectstest.SeedLibrary(me, "A", "B", "C")
		seer := effectstest.B12Push(g, me.ID, "Merfolk Seer", "Creature — Merfolk Wizard", oldB800Seer, 2, 2)
		src := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
		before := me.Hand.Size()
		effectstest.DamageAndSettle(t, g, src, seer, 5)
		if g.Battlefield.Contains(seer) {
			t.Fatal("the Seer survived")
		}
		if pay {
			effectstest.DcMana(t, g, me.ID, "{U}{U}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if got := me.Hand.Size() - before; got != want {
			t.Errorf("pay %v: drew %d, want %d", pay, got, want)
		}
	}
}

func TestOldB800InfernalContractDrawsFourThenLosesHalfYourLifeRoundedUp(t *testing.T) {
	for _, tc := range []struct{ life, lost int }{{20, 10}, {21, 11}, {2, 1}} {
		g := effectstest.NewCatalogGame(t)
		me := oldB800Main(t, g)
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
		g.WithWriteLock(func() { me.Life = tc.life })
		before := me.Library.Size()
		effectstest.CastCatalogSpell(t, g, "Infernal Contract", "Sorcery", oldB800Contract, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := before - me.Library.Size(); got != 4 {
			t.Errorf("life %d: drew %d, want 4", tc.life, got)
		}
		if got := tc.life - me.Life; got != tc.lost {
			t.Errorf("life %d: lost %d, want %d", tc.life, got, tc.lost)
		}
	}
}

func TestOldB800ZirilanFetchesADragonWithHasteAndExilesItAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zir := effectstest.B12Push(g, me.ID, "Zirilan of the Claw", "Legendary Creature — Lizard Shaman", oldB800Zirilan, 3, 4)
	dragon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Test Dragon", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Keywords: []string{"flying"}})
	notDragon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := effectstest.CrtActivateRaw(t, g, me.ID, zir, 0, nil, "R"); err == nil {
		t.Fatal("activated without paying {1}{R}{R}")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, zir, 0, nil, "R", "R", "R"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if id := effectstest.SearchOptionNamed(g, c, "Test Bear"); id != uuid.Nil {
		t.Error("a non-Dragon was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Fatal("the Dragon did not enter the battlefield")
	}
	if g.Battlefield.Contains(notDragon) {
		t.Error("the Bear entered")
	}
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "haste") {
		t.Error("the Dragon has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) || !effectstest.InExile(g, dragon) {
		t.Error("the Dragon was not exiled at the beginning of the end step")
	}
}
