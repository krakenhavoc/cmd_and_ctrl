package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b669_cards_test.go — Prophecy: Auras, batch 669 (#5617). Each test
// plays the real catalog entry and asserts what its text says.

const (
	b669Entangler  = "285cf707-df0c-4eb9-8f60-5d052ee8eb8c"
	b669Orders     = "2f3c66db-efc1-4571-a4db-371a3d532615"
	b669Flowering  = "40155cfb-078a-4359-9551-9c71266d8a1c"
	b669Barbed     = "45222123-f5e0-48ff-b80c-858727d155d7"
	b669Sunken     = "46569712-1f28-41cb-81be-fd938408c71f"
	b669Verdant    = "4d59fd2b-7fc6-49a7-8938-d3f08902a5a8"
	b669Jolrael    = "bf71cca7-be71-4429-9cb6-1eb34ff847eb"
	b669Noxious    = "c75ec3c6-2786-4386-b758-ac03610f2a99"
	b669Boon       = "e3e5b12c-2103-4b40-83d8-6d5449179b6f"
	b669Cloak      = "e52eb1a6-fff1-4a47-b434-31a74d76231c"
	b669Caress     = "e9e0b78e-07d5-4603-8e3b-27274148d1a1"
	b669FieldOwner = "Land"
)

func TestB669CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		b669Entangler: "Entangler", b669Orders: "Latulla's Orders", b669Flowering: "Flowering Field",
		b669Barbed: "Barbed Field", b669Sunken: "Sunken Field", b669Verdant: "Verdant Field",
		b669Jolrael: "Jolrael's Favor", b669Noxious: "Noxious Field", b669Boon: "Mageta's Boon",
		b669Cloak: "Alexi's Cloak", b669Caress: "Greel's Caress",
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

func b669Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func TestB669EntanglerLetsTheHostBlockAnyNumber(t *testing.T) {
	g, me, _ := b669Table(t)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	if got := effectstest.MbCapacity(t, g, host); got != 1 {
		t.Fatalf("capacity %d before the Aura, want 1", got)
	}
	effectstest.AuraCast(t, g, "Entangler", b669Entangler, host)
	if got := effectstest.MbCapacity(t, g, host); got != 0 {
		t.Errorf("capacity %d with the Aura, want 0 (any number)", got)
	}
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	if got := effectstest.MbCapacity(t, g, other); got != 1 {
		t.Errorf("an unenchanted creature has capacity %d", got)
	}
}

func TestB669MagetasBoonAlexisCloakAndGreelsCaress(t *testing.T) {
	g, me, _ := b669Table(t)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Mageta's Boon", b669Boon, host)
	if p, tt := effectstest.PtOf(t, g, host); p != 3 || tt != 4 {
		t.Errorf("Mageta's Boon: %d/%d, want 3/4", p, tt)
	}
	effectstest.AuraCast(t, g, "Greel's Caress", b669Caress, host)
	if p, tt := effectstest.PtOf(t, g, host); p != 0 || tt != 4 {
		t.Errorf("Greel's Caress: %d/%d, want 0/4", p, tt)
	}
	effectstest.AuraCast(t, g, "Alexi's Cloak", b669Cloak, host)
	if !effectstest.HasKeywordOnBattlefield(t, g, host, "shroud") {
		t.Error("the enchanted creature has no shroud")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2), "shroud") {
		t.Error("an unenchanted creature has shroud")
	}
}

func TestB669JolraelsFavorRegeneratesTheHost(t *testing.T) {
	g, me, _ := b669Table(t)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := effectstest.AuraCast(t, g, "Jolrael's Favor", b669Jolrael, host)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was activated with no mana")
	}
	effectstest.FloatForTest(g, me, "GG")
	effectstest.ApaActivate(t, g, me, aura, 0, game.ActivateAbilityParams{})
	effectstest.P1Destroy(t, g, host)
	if !effectstest.BattlefieldHas(g, host) {
		t.Error("the regeneration shield did not save the creature")
	}
}

// b669Field puts the named Field Aura on a land of `owner`'s and returns the land and the
// index and ref of its granted ability.
func b669Field(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string) (uuid.UUID, int, string) {
	t.Helper()
	land := effectstest.GaPut(g, owner, "Plains", "Basic Land — Plains", "")
	effectstest.PushAuraOnLand(t, g, owner, name, oracle, land)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatalf("%s: the enchanted land has no granted ability", name)
	}
	return land, idx, ref
}

func TestB669BarbedFieldPingsFromTheLand(t *testing.T) {
	g, me, opp := b669Table(t)
	land, idx, ref := b669Field(t, g, me.ID, "Barbed Field", b669Barbed)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if !effectstest.PermanentTapped(g, land) {
		t.Error("the land did not tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a tapped land activated it again")
	}
}

func TestB669FloweringFieldPreventsOneDamage(t *testing.T) {
	g, me, opp := b669Table(t)
	land, idx, ref := b669Field(t, g, me.ID, "Flowering Field", b669Flowering)
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(opp.ID, me.ID, 3)
	})
	if me.Life != life-2 {
		t.Errorf("life %d after 3 damage with a 1-point shield, want %d", me.Life, life-2)
	}
}

func TestB669VerdantFieldPumpsUntilEndOfTurn(t *testing.T) {
	g, me, _ := b669Table(t)
	land, idx, ref := b669Field(t, g, me.ID, "Verdant Field", b669Verdant)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("%d/%d, want 3/3", p, tt)
	}
}

func TestB669NoxiousFieldDamagesEverything(t *testing.T) {
	g, me, opp := b669Table(t)
	land, idx, ref := b669Field(t, g, me.ID, "Noxious Field", b669Noxious)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 2, 3)
	myLife, theirLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if me.Life != myLife-1 || opp.Life != theirLife-1 {
		t.Errorf("life %d/%d, want each down 1", me.Life, opp.Life)
	}
	if effectstest.BattlefieldHas(g, small) {
		t.Error("the 2/1 survived")
	}
	if !effectstest.BattlefieldHas(g, big) {
		t.Error("the 2/3 died")
	}
}

func TestB669SunkenFieldCountersUnlessThePlayerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, victim := legB2CounterSetupB669(t, g)
		land, idx, ref := b669Field(t, g, opp.ID, "Sunken Field", b669Sunken)
		if err := g.ActivateCatalogAbility(opp.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		if !g.Stack.Contains(victim) {
			t.Fatal("the spell left the stack while the {1} was unanswered")
		}
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		countered := !g.Stack.Contains(victim) && me.Graveyard.Contains(victim)
		if countered == pay {
			t.Errorf("paid %v: countered %v", pay, countered)
		}
	}
}

func legB2CounterSetupB669(t *testing.T, g *game.Game) (*game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Spell", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	return me, opp, victim
}

func TestB669LatullasOrdersMayDestroyAnArtifactOfTheDamagedPlayer(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		other := g.Seats[2]
		host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		elsewhere := effectstest.B12Permanent(g, other.ID, "Other Rock", "Artifact")
		effectstest.AuraCast(t, g, "Latulla's Orders", b669Orders, host)
		effectstest.DealCombatDamageToPlayer(g, host, opp.ID, 2)
		effectstest.B04WaitForPick(t, g, me.ID)
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil || effectstest.HasID(p.PickTargetCards, elsewhere) || !effectstest.HasID(p.PickTargetCards, theirs) {
			t.Fatalf("pick offered %+v: only the damaged player's artifacts", p)
		}
		effectstest.PickCard(t, g, me.ID, theirs)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if gone := !effectstest.BattlefieldHas(g, theirs); gone != yes {
			t.Errorf("yes=%v: artifact gone=%v", yes, gone)
		}
		if !effectstest.BattlefieldHas(g, elsewhere) {
			t.Error("another player's artifact was destroyed")
		}
	}
}
