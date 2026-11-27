package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b729_cards_test.go — Urza's Saga, batch 729 (#5624). Each test
// plays the real catalog entry and asserts what its text says.

const (
	b729RainOfSaltOracle    = "1219e330-01ac-405a-b75a-dd4298598167"
	b729OutmaneuverOracle   = "42b44108-4a03-4333-bf5a-e42493f4d5e4"
	b729CurfewOracle        = "6ff1e793-1569-4887-b6bb-d9855c5e8a26"
	b729PlanarBirthOracle   = "7a9d13b3-3823-463a-acd9-8b7a9d5e121f"
	b729AlterationOracle    = "c80935bf-c17f-4940-ad9f-3f4e3e8f71bc"
	b729HeadlongRushOracle  = "cc3abacc-6643-49c5-8df8-5b56ab20fba5"
	b729WaylayOracle        = "f0db78c7-09b5-4981-8ca8-6dbab597c5d9"
	b729SunderOracle        = "fddc93a2-912e-402a-9d0e-34f778e659ea"
	b729WildGrowthOracle    = "706ae742-1807-44b7-a4fa-f2e26f61519a"
	b729IndomitableOracle   = "069c3c87-81a8-4d2a-ad2f-8d3243d0e3db"
	b729WildGrowthLine      = "Enchantment — Aura"
	b729BasicForestTypeLine = "Basic Land — Forest"
)

func TestOldB729EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b729RainOfSaltOracle:   "Rain of Salt",
		b729OutmaneuverOracle:  "Outmaneuver",
		b729CurfewOracle:       "Curfew",
		b729PlanarBirthOracle:  "Planar Birth",
		b729AlterationOracle:   "Enchantment Alteration",
		b729HeadlongRushOracle: "Headlong Rush",
		b729WaylayOracle:       "Waylay",
		b729SunderOracle:       "Sunder",
	}
	for oracle, name := range want {
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

// b729Cast puts the spell in the active player's hand and casts it where the
// table stands (no walk to a main phase), strictly.
func b729Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, p game.CastSpellParams) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, p)
}

func b729Target(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func b729OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

func b729Creature(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", Power: p, Toughness: tough})
}

func TestOldB729RainOfSaltDestroysTwoTargetLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.AangPushLand(g, me.ID, "Forest", false)
	b := effectstest.AangPushLand(g, opp.ID, "Island", false)
	c := effectstest.AangPushLand(g, opp.ID, "Swamp", false)
	effectstest.CastCatalogSpell(t, g, "Rain of Salt", "Sorcery", b729RainOfSaltOracle, []game.TargetRef{b729Target(a), b729Target(b)})
	effectstest.PassPriorityAroundTable(t, g)
	if b729OnBattlefield(g, a) || b729OnBattlefield(g, b) {
		t.Error("a targeted land survived")
	}
	if !b729OnBattlefield(g, c) {
		t.Error("the untargeted land was destroyed")
	}
}

func TestOldB729RainOfSaltNeedsTwoLandTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.AangPushLand(g, me.ID, "Forest", false)
	bear := b729Creature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b729Cast(t, g, "Rain of Salt", "Sorcery", b729RainOfSaltOracle, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{b729Target(a)}}); err == nil {
		t.Error("one target land was accepted")
	}
	if err := b729Cast(t, g, "Rain of Salt", "Sorcery", b729RainOfSaltOracle, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{b729Target(a), b729Target(bear)}}); err == nil {
		t.Error("a creature was accepted as the second land")
	}
}

func TestOldB729SunderReturnsEveryLandToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.AangPushLand(g, me.ID, "Forest", false)
	b := effectstest.AangPushLand(g, opp.ID, "Island", true)
	bear := b729Creature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sunder", "Instant", b729SunderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if b729OnBattlefield(g, a) || b729OnBattlefield(g, b) {
		t.Error("a land stayed on the battlefield")
	}
	if !b729OnBattlefield(g, bear) {
		t.Error("a nonland permanent was returned")
	}
	if n := countHandNamed(opp, "Island"); n != 1 {
		t.Errorf("the opponent's hand holds %d Islands, want 1", n)
	}
	if n := countHandNamed(me, "Forest"); n != 1 {
		t.Errorf("your hand holds %d Forests, want 1", n)
	}
}

func countHandNamed(p *game.Player, name string) int {
	n := 0
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

func TestOldB729PlanarBirthReturnsEveryBasicLandCardTappedUnderItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardCardWithTypeLineID(me, "Forest", b729BasicForestTypeLine)
	theirs := effectstest.PushGraveyardCardWithTypeLineID(opp, "Forest", b729BasicForestTypeLine)
	nonbasic := effectstest.PushGraveyardCardWithTypeLineID(opp, "Tundra", "Land — Plains Island")
	bear := effectstest.PushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Planar Birth", "Sorcery", b729PlanarBirthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		id    uuid.UUID
		owner uuid.UUID
	}{{mine, me.ID}, {theirs, opp.ID}} {
		bf := effectstest.FindBattlefieldCardForTest(g, c.id)
		if bf == nil {
			t.Fatal("a basic land card stayed in a graveyard")
		}
		if bf.Controller != c.owner || !bf.Tapped {
			t.Errorf("land entered under %s tapped=%v, want its owner %s tapped", bf.Controller, bf.Tapped, c.owner)
		}
	}
	if b729OnBattlefield(g, nonbasic) || b729OnBattlefield(g, bear) {
		t.Error("a nonbasic land or a creature card came back")
	}
}

func TestOldB729CurfewEachPlayerReturnsACreatureTheyControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	only := b729Creature(g, opp.ID, "Lonely Bear", 2, 2)
	mineA := b729Creature(g, me.ID, "Mine A", 1, 1)
	mineB := b729Creature(g, me.ID, "Mine B", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Curfew", "Instant", b729CurfewOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The opponent has exactly one creature, so only the caster is asked.
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("the caster was not asked which creature to return")
	}
	if !b729OnBattlefield(g, only) {
		t.Fatal("nothing should have returned before the last choice")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{mineB}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if b729OnBattlefield(g, only) || b729OnBattlefield(g, mineB) {
		t.Error("a chosen creature stayed on the battlefield")
	}
	if !b729OnBattlefield(g, mineA) {
		t.Error("an unchosen creature returned")
	}
	if countHandNamed(opp, "Lonely Bear") != 1 || countHandNamed(me, "Mine B") != 1 {
		t.Error("the chosen creatures are not in their owners' hands")
	}
}

func TestOldB729CurfewWithNoCreaturesDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Curfew", "Instant", b729CurfewOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, g.Seats[0].ID); c != nil {
		t.Error("asked for a creature on an empty board")
	}
}

func TestOldB729WaylayMakesThreeKnightsAndExilesThemAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Waylay", "Instant", b729WaylayOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	count := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Knight" && c.IsToken() {
				n++
				if p, tough := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 2 || tough != 2 {
					t.Errorf("a Knight is %d/%d, want 2/2", p, tough)
				}
			}
		}
		return n
	}
	if n := count(); n != 3 {
		t.Fatalf("%d Knights, want 3", n)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := count(); n != 3 {
		t.Fatalf("%d Knights after the end step, want 3 (they leave at cleanup)", n)
	}
	me.Hand.Cards = nil // nothing to discard, so cleanup is not held for CR 514.1
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.PassPriorityAroundTable(t, g)
	if n := count(); n != 0 {
		t.Errorf("%d Knights survived the cleanup step", n)
	}
}

func TestOldB729HeadlongRushGivesAttackersFirstStrikeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	attacker := b729Creature(g, me.ID, "Attacker", 2, 2)
	idle := b729Creature(g, me.ID, "Idle", 2, 2)
	effectstest.Pr7bAttack(t, g, g.Seats[1].ID, attacker)
	if err := b729Cast(t, g, "Headlong Rush", "Instant", b729HeadlongRushOracle, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, attacker, "first strike") {
		t.Error("the attacker has no first strike")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, idle, "first strike") {
		t.Error("a creature that was not attacking gained first strike")
	}
}

func TestOldB729OutmaneuverBlockedCreatureAssignsAllItsDamageToThePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b729Creature(g, me.ID, "Attacker", 3, 3)
	wall := effectstest.MscWlBlocker(g, "Wall", 0, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: attacker})
	if err := b729Cast(t, g, "Outmaneuver", "Instant", b729OutmaneuverOracle, game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{b729Target(attacker)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7bDamageStep(t, g)
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceDamageAssignment {
			t.Fatal("the attacker was asked how to assign its damage; it is not a choice")
		}
	}
	if opp.Life != 37 {
		t.Errorf("the defending player has %d life, want 37 (all 3 damage assigned to them)", opp.Life)
	}
	if got := effectstest.Pr6Marked(g, wall); got != 0 {
		t.Errorf("the blocker was dealt %d damage, want none", got)
	}
}

func TestOldB729OutmaneuverOnlyHelpsTheCreaturesItTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helped := b729Creature(g, me.ID, "Helped", 3, 3)
	other := b729Creature(g, me.ID, "Other", 3, 3)
	wallA := effectstest.MscWlBlocker(g, "Wall A", 0, 5)
	wallB := effectstest.MscWlBlocker(g, "Wall B", 0, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, helped, other)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wallA: helped, wallB: other})
	if err := b729Cast(t, g, "Outmaneuver", "Instant", b729OutmaneuverOracle, game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{b729Target(helped)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != 37 {
		t.Errorf("the defending player has %d life, want 37 (only the targeted creature's 3)", opp.Life)
	}
	if got := effectstest.Pr6Marked(g, wallB); got != 3 {
		t.Errorf("the untargeted attacker's blocker took %d, want 3", got)
	}
}

func TestOldB729OutmaneuverCannotTargetAnUnblockedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b729Creature(g, me.ID, "Attacker", 3, 3)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{})
	if err := b729Cast(t, g, "Outmaneuver", "Instant", b729OutmaneuverOracle, game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{b729Target(attacker)}}); err == nil {
		t.Error("an unblocked creature was accepted")
	}
}

func b729Aura(g *game.Game, owner, host uuid.UUID, name, oracle string) uuid.UUID {
	return effectstest.Cmdr2B19Attach(g, owner, host, name, oracle)
}

func TestOldB729EnchantmentAlterationMovesAnAuraToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	from := b729Creature(g, opp.ID, "From", 2, 2)
	to := b729Creature(g, opp.ID, "To", 2, 2)
	land := effectstest.AangPushLand(g, opp.ID, "Island", false)
	aura := b729Aura(g, me.ID, from, "Indomitable Might", b729IndomitableOracle)
	effectstest.CastCatalogSpell(t, g, "Enchantment Alteration", "Instant", b729AlterationOracle, []game.TargetRef{b729Target(aura)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no choice of new host")
	}
	for _, id := range c.ChooseCards {
		if id == land || id == from {
			t.Errorf("offered %s as a host (a land or the current host)", id)
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{to}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	got := effectstest.FindBattlefieldCardForTest(g, aura)
	if got == nil || got.AttachedTo.ID != to {
		t.Errorf("the Aura is on %v, want the other creature", got)
	}
}

func TestOldB729EnchantmentAlterationKeepsALandAuraOnLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	from := effectstest.AangPushLand(g, opp.ID, "Forest", false)
	to := effectstest.AangPushLand(g, opp.ID, "Forest", false)
	bear := b729Creature(g, opp.ID, "Bear", 2, 2)
	aura := b729Aura(g, me.ID, from, "Wild Growth", b729WildGrowthOracle)
	effectstest.CastCatalogSpell(t, g, "Enchantment Alteration", "Instant", b729AlterationOracle, []game.TargetRef{b729Target(aura)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no choice of new host")
	}
	for _, id := range c.ChooseCards {
		if id == bear {
			t.Error("a creature was offered as a new host for a land Aura")
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{to}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := effectstest.FindBattlefieldCardForTest(g, aura); got == nil || got.AttachedTo.ID != to {
		t.Error("the land Aura did not move to the other land")
	}
}

func TestOldB729EnchantmentAlterationNeedsAnAuraOnACreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	loose := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := b729Cast(t, g, "Enchantment Alteration", "Instant", b729AlterationOracle, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{b729Target(loose)}}); err == nil {
		t.Error("a non-Aura was accepted")
	}
}
