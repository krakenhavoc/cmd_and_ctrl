package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// smaug_unchained_b3_test.go — deck request #4724 (Smaug Unchained),
// batch 3: Rivaz of the Claw, Expedited Inheritance and Blazing Sunsteel.

const (
	sub3RivazOracle      = "ecff46b9-fcc9-4895-8b6f-c697fd15d01a"
	sub3InheritOracle    = "11e2f581-7c3c-4f2f-877f-e41540ab967e"
	sub3SunsteelOracle   = "46c77100-ab2e-43eb-b1cb-c22ed114213e"
	sub3GisaGeralfOracle = "20e94d0f-6887-45f8-a137-877011d786c6"
)

func TestSub3CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		sub3RivazOracle:    "Rivaz of the Claw",
		sub3InheritOracle:  "Expedited Inheritance",
		sub3SunsteelOracle: "Blazing Sunsteel",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// --- Rivaz of the Claw -------------------------------------------------------

// sub3RivazTable is the active seat in its precombat main with an untapped
// Rivaz of the Claw.
func sub3RivazTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	rivaz := pushCatalogPermanent(g, me.ID, "Rivaz of the Claw", "Legendary Creature — Lizard Warlock", sub3RivazOracle, false)
	return g, me, rivaz
}

// sub3GraveCard puts a free card into p's graveyard.
func sub3GraveCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 4, Toughness: 4, Owner: p.ID, Controller: p.ID})
	return id
}

// sub3Destroy destroys a permanent and lets every trigger resolve.
func sub3Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
}

func sub3Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
}

// sub3TriggersFrom counts the triggered items from one source waiting or
// on the stack.
func sub3TriggersFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			n++
		}
	}
	return n
}

func TestSub3RivazHasMenaceAndItsManaPaysOnlyForDragonCreatureSpells(t *testing.T) {
	g, me, rivaz := sub3RivazTable(t)
	if !hasKeywordOnBattlefield(t, g, rivaz, "menace") {
		t.Error("Rivaz has no menace")
	}
	if err := g.ActivateManaAbility(me.ID, rivaz, 0, game.ManaAbilityParams{Colors: []string{"B", "R"}}); err != nil {
		t.Fatalf("tap Rivaz: %v", err)
	}
	if got := poolColors(me); len(got) != 2 {
		t.Fatalf("pool = %v, want two mana", got)
	}
	cost, _ := game.ParseCost("{B}{R}")
	dragon := game.Card{Name: "Dragon", TypeLine: "Creature — Dragon", ManaCost: "{B}{R}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{B}{R}"}
	kindred := game.Card{Name: "Dragon Lore", TypeLine: "Kindred Sorcery — Dragon", ManaCost: "{B}{R}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(dragon)) {
		t.Error("refused a Dragon creature spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a creature spell that is not a Dragon")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(kindred)) {
		t.Error("paid for a Dragon spell that is not a creature")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(dragon)) {
		t.Error("paid for an ability of a Dragon")
	}
}

func TestSub3RivazCastsOneDragonCreatureFromTheGraveyardEachTurn(t *testing.T) {
	g, me, _ := sub3RivazTable(t)
	bear := sub3GraveCard(me, "Grave Bear", "Creature — Bear")
	first := sub3GraveCard(me, "Grave Dragon", "Creature — Dragon")
	second := sub3GraveCard(me, "Other Dragon", "Creature — Dragon")

	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Fatal("cast a creature that is not a Dragon from the graveyard")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast a Dragon from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first) {
		t.Fatal("the Dragon did not resolve")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Fatal("a second Dragon was cast from the graveyard in the same turn")
	}
}

// The headline: a Dragon cast from the graveyard gains "When this creature
// dies, exile it." and keeps it after Rivaz has gone.
func TestSub3RivazDragonCastFromTheGraveyardIsExiledWhenItDies(t *testing.T) {
	g, me, rivaz := sub3RivazTable(t)
	dragon := sub3GraveCard(me, "Grave Dragon", "Creature — Dragon")
	if err := g.CastSpell(me.ID, dragon, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	if got := sub3TriggersFrom(g, rivaz); got != 1 {
		t.Fatalf("%d Rivaz triggers for a graveyard Dragon, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Fatal("the Dragon did not resolve")
	}
	// The grant is not tied to Rivaz.
	sub3Destroy(t, g, rivaz)
	sub3Destroy(t, g, dragon)
	if z := sub3Zone(g, dragon); z != game.ZoneExile {
		t.Fatalf("the Dragon that died is in %q, want exile", z)
	}
}

// A Dragon cast from hand gains nothing and stays in the graveyard.
func TestSub3RivazDragonCastFromHandIsNotExiled(t *testing.T) {
	g, _, rivaz := sub3RivazTable(t)
	dragon := castCatalogSpell(t, g, "Hand Dragon", "Creature — Dragon", "", nil)
	if got := sub3TriggersFrom(g, rivaz); got != 0 {
		t.Fatalf("%d Rivaz triggers for a Dragon cast from hand", got)
	}
	passPriorityAroundTable(t, g)
	sub3Destroy(t, g, dragon)
	if z := sub3Zone(g, dragon); z != game.ZoneGraveyard {
		t.Fatalf("the hand-cast Dragon is in %q, want the graveyard", z)
	}
}

// The trigger reads the cast, not Rivaz's permission: a Zombie Dragon cast
// through Gisa and Geralf once Rivaz's cast is spent still gains it, and a
// creature that is not a Dragon gains nothing.
func TestSub3RivazTriggersOnAnyGraveyardCastOfADragonCreature(t *testing.T) {
	g, me, rivaz := sub3RivazTable(t)
	pushCatalogPermanent(g, me.ID, "Gisa and Geralf", "Legendary Creature — Human Wizard", sub3GisaGeralfOracle, false)
	plain := sub3GraveCard(me, "Grave Dragon", "Creature — Dragon")
	zombieDragon := sub3GraveCard(me, "Zombie Dragon", "Creature — Zombie Dragon")

	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the Dragon through Rivaz: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, zombieDragon, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the Zombie Dragon through Gisa and Geralf: %v", err)
	}
	if got := sub3TriggersFrom(g, rivaz); got != 1 {
		t.Fatalf("%d Rivaz triggers for a Dragon cast through another permission, want 1", got)
	}
	passPriorityAroundTable(t, g)
	sub3Destroy(t, g, zombieDragon)
	if z := sub3Zone(g, zombieDragon); z != game.ZoneExile {
		t.Fatalf("the Zombie Dragon is in %q, want exile", z)
	}
}

func TestSub3RivazIgnoresANonDragonCastFromTheGraveyard(t *testing.T) {
	g, me, rivaz := sub3RivazTable(t)
	pushCatalogPermanent(g, me.ID, "Gisa and Geralf", "Legendary Creature — Human Wizard", sub3GisaGeralfOracle, false)
	zombie := sub3GraveCard(me, "Grave Zombie", "Creature — Zombie")
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the Zombie through Gisa and Geralf: %v", err)
	}
	if got := sub3TriggersFrom(g, rivaz); got != 0 {
		t.Fatalf("%d Rivaz triggers for a Zombie", got)
	}
	passPriorityAroundTable(t, g)
	sub3Destroy(t, g, zombie)
	if z := sub3Zone(g, zombie); z != game.ZoneGraveyard {
		t.Fatalf("the Zombie is in %q, want the graveyard", z)
	}
}

// --- Expedited Inheritance ---------------------------------------------------

// sub3InheritTable is a game with Expedited Inheritance under seat 0 and a
// three-card library for seat 1.
func sub3InheritTable(t *testing.T) (g *game.Game, me, opp *game.Player, lib []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Expedited Inheritance", "Enchantment", sub3InheritOracle, false)
	lib = seedSearchLibrary(opp,
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
		game.Card{Name: "Deeper Card", TypeLine: "Sorcery"},
	)
	return g, me, opp, lib
}

// sub3Confirm is the open yes/no for `who`.
func sub3Confirm(t *testing.T, g *game.Game, who uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == who {
			return c
		}
	}
	t.Fatalf("no yes/no prompt for %s: %+v", who, g.PendingChoices)
	return nil
}

// The opponent's creature is dealt damage, so the OPPONENT decides, exiles
// from their own library and may play the cards through their next turn.
func TestSub3ExpeditedInheritanceLetsTheDamagedCreaturesControllerImpulseThatMany(t *testing.T) {
	g, me, opp, lib := sub3InheritTable(t)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, theirs, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == me.ID {
			t.Fatal("the enchantment's controller was asked; the creature's controller decides")
		}
	}
	sub3Confirm(t, g, opp.ID)
	answerConfirmFor(t, g, opp, true)
	monSettle(t, g)

	for _, id := range lib[:2] {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not exiled", id)
		}
		perm := exiledPermission(g, id)
		if perm.Player != opp.ID || perm.CastOnly || !permissionLive(g, perm, opp.ID) {
			t.Errorf("%+v: the damaged creature's controller may PLAY it now", perm)
		}
		if permissionLive(g, perm, me.ID) {
			t.Error("the enchantment's controller may play the opponent's exiled card")
		}
		if perm.Duration.Kind != game.UntilEndOfTurn || perm.Duration.Player != opp.ID {
			t.Errorf("duration %+v, want the end of the opponent's next turn", perm.Duration)
		}
	}
	if !opp.Library.Contains(lib[2]) || !opp.Library.Contains(lib[3]) {
		t.Error("more than two cards left the library")
	}
}

// The window runs to the end of the creature controller's NEXT turn.
func TestSub3ExpeditedInheritanceWindowEndsAfterTheirNextTurn(t *testing.T) {
	g, _, opp, lib := sub3InheritTable(t)
	advanceToMainOf(t, g, 0)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, theirs, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
	answerConfirmFor(t, g, opp, true)
	monSettle(t, g)
	land := lib[0]
	if !g.Exile.Contains(land) {
		t.Fatal("the top card was not exiled")
	}
	advanceToMainOf(t, g, 1)
	if !permissionLive(g, exiledPermission(g, land), opp.ID) {
		t.Fatal("the window is closed during the opponent's next turn")
	}
	advanceToMainOf(t, g, 2)
	if exiledPermission(g, land).Granted() && permissionLive(g, exiledPermission(g, land), opp.ID) {
		t.Error("the window outlived the end of the opponent's next turn")
	}
}

func TestSub3ExpeditedInheritanceDeclineExilesNothing(t *testing.T) {
	g, _, opp, lib := sub3InheritTable(t)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, theirs, 3); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
	answerConfirmFor(t, g, opp, false)
	monSettle(t, g)
	for _, id := range lib {
		if !opp.Library.Contains(id) {
			t.Fatalf("a declined trigger exiled %s", id)
		}
	}
}

// Damage dealt at the same time is one trigger for the total, and a
// creature that died to it is still asked about through its last-known
// controller.
func TestSub3ExpeditedInheritanceReadsTheTotalAndALethallyDamagedCreature(t *testing.T) {
	g, me, opp, lib := sub3InheritTable(t)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	srcA := b12Creature(g, me.ID, "Source A", "Creature — Bear", 1, 1)
	srcB := b12Creature(g, me.ID, "Source B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(srcA, theirs, 1)
			return g.DealDamageToCreatureForEffect(srcB, theirs, 2)
		})
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the bear survived lethal damage")
	}
	c := sub3Confirm(t, g, opp.ID)
	if c.PickMatchArg != "3" {
		t.Errorf("asked to exile %q cards, want the total 3", c.PickMatchArg)
	}
	answerConfirmFor(t, g, opp, true)
	monSettle(t, g)
	for _, id := range lib[:3] {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not exiled", id)
		}
	}
	if !opp.Library.Contains(lib[3]) {
		t.Error("a fourth card was exiled")
	}
}

// Your own creature: you are the one asked.
func TestSub3ExpeditedInheritanceAsksYouForYourOwnCreature(t *testing.T) {
	g, me, _, _ := sub3InheritTable(t)
	mine := seedSearchLibrary(me, game.Card{Name: "My Top", TypeLine: "Instant"})
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, bear, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
	answerConfirmFor(t, g, me, true)
	monSettle(t, g)
	if !g.Exile.Contains(mine[0]) || exiledPermission(g, mine[0]).Player != me.ID {
		t.Fatal("your own creature's damage did not let you exile your top card")
	}
}

// --- Blazing Sunsteel --------------------------------------------------------

// sub3SunsteelTable is seat 0's main phase with a 2/4 equipped with
// Blazing Sunsteel.
func sub3SunsteelTable(t *testing.T) (g *game.Game, me, opp *game.Player, sunsteel, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	sunsteel = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blazing Sunsteel", TypeLine: "Artifact — Equipment",
		OracleID: sub3SunsteelOracle, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	equipTo(t, g, me.ID, sunsteel, bear)
	if host := attachmentHostOf(t, g, sunsteel); host.ID != bear {
		t.Fatalf("Sunsteel attached to %+v, want the bear", host)
	}
	return g, me, opp, sunsteel, bear
}

func TestSub3BlazingSunsteelPumpsPerOpponentStillInTheGame(t *testing.T) {
	g, _, opp, _, bear := sub3SunsteelTable(t)
	if got := effectivePower(t, g, bear); got != 2+3 {
		t.Fatalf("power %d with three opponents, want 5", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness %d, want 4 (+1/+0)", got)
	}
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if got := effectivePower(t, g, bear); got != 2+2 {
		t.Errorf("power %d after an opponent left, want 4", got)
	}
}

// The equipped creature is dealt damage and deals that much to the target,
// as the source of the new damage.
func TestSub3BlazingSunsteelEquippedCreatureReflectsTheDamage(t *testing.T) {
	g, _, opp, _, bear := sub3SunsteelTable(t)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, bear, 3); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	monSettle(t, g)
	if opp.Life != life-3 {
		t.Fatalf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	found := false
	for _, ev := range dealDamageEventsTo(g, opp.ID) {
		if ev.Source == bear && ev.Amount == 3 {
			found = true
		}
	}
	if !found {
		t.Error("the equipped creature is not the source of the reflected damage")
	}
}

// Two sources at once are one trigger for the total, and a creature that
// died to the damage still deals it.
func TestSub3BlazingSunsteelReflectsTheTotalEvenWhenTheCreatureDies(t *testing.T) {
	g, _, opp, sunsteel, bear := sub3SunsteelTable(t)
	srcA := b12Creature(g, opp.ID, "Source A", "Creature — Bear", 2, 2)
	srcB := b12Creature(g, opp.ID, "Source B", "Creature — Bear", 3, 3)
	life := opp.Life
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(srcA, bear, 2)
			return g.DealDamageToCreatureForEffect(srcB, bear, 3)
		})
	})
	g.RunStateChecksForTest()
	if got := sub3TriggersFrom(g, sunsteel); got > 1 {
		t.Fatalf("%d Sunsteel triggers for one damage instance, want 1", got)
	}
	monSettle(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	monSettle(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the 2/4 survived 5 damage")
	}
	if opp.Life != life-5 {
		t.Fatalf("opponent life %d -> %d, want -5 (the total)", life, opp.Life)
	}
}

// A creature the Sunsteel is not attached to reflects nothing.
func TestSub3BlazingSunsteelIgnoresAnUnequippedCreature(t *testing.T) {
	g, me, _, sunsteel, _ := sub3SunsteelTable(t)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, other, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if got := sub3TriggersFrom(g, sunsteel); got != 0 {
		t.Fatalf("%d Sunsteel triggers for an unequipped creature", got)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("an unequipped creature's damage asked for a target")
		}
	}
}
