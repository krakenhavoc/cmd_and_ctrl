package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jumpstart Azorius deck request, batch 8 (#4894).

const (
	jmp8KeenGlidemaster    = "98deac9e-28ea-4d11-8d75-833bf6d66e6e"
	jmp8LawmagesBinding    = "10f0e797-3887-4c2b-9cff-80f110b952e1"
	jmp8LeaveInTheDust     = "7bb843b6-c55b-4796-b1a7-2fff7b05c276"
	jmp8LegionsJudgment    = "dce4d65f-cb91-48a2-a2a7-013e73dfb667"
	jmp8LoftyDenial        = "d799a2c4-628a-45da-b9ca-0337c46e3e14"
	jmp8LongRoadHome       = "a8f53018-c496-41f3-bc3f-3c4703b9e4e1"
	jmp8MomentOfHeroism    = "a167d593-cded-43e1-b9bb-2287dbd0cbec"
	jmp8PrescientChimera   = "af199b80-d5d8-417a-97d5-459a5bd48b22"
	jmp8RoamingGhostlight  = "0d5eb0b9-9b08-4d26-b141-a9d612a680a6"
	jmp8ArrestOracleForCmp = "81728b98-8cf9-4734-a318-69184bb4d15c"
)

func TestJmp8CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{
		jmp8KeenGlidemaster:   "Keen Glidemaster",
		jmp8LawmagesBinding:   "Lawmage's Binding",
		jmp8LeaveInTheDust:    "Leave in the Dust",
		jmp8LegionsJudgment:   "Legion's Judgment",
		jmp8LoftyDenial:       "Lofty Denial",
		jmp8LongRoadHome:      "Long Road Home",
		jmp8MomentOfHeroism:   "Moment of Heroism",
		jmp8PrescientChimera:  "Prescient Chimera",
		jmp8RoamingGhostlight: "Roaming Ghostlight",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full and none", name, spec.Completeness, spec.Caveats)
		}
	}
	for _, oracle := range []string{jmp8PrescientChimera, jmp8RoamingGhostlight} {
		spec, _ := Lookup(oracle)
		if !hasAbility(spec.PrintedKeywords, "flying") {
			t.Errorf("%s prints flying", spec.Name)
		}
	}
	if spec, _ := Lookup(jmp8LawmagesBinding); !hasAbility(spec.PrintedKeywords, "flash") {
		t.Error("Lawmage's Binding prints flash")
	}
}

func jmp8Live(g *game.Game, id uuid.UUID) *game.Card {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return findBattlefieldCardForTest(g, id)
}

func jmp8Flyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: owner, Controller: owner,
	})
}

// --- Keen Glidemaster ---

func TestJmp8KeenGlidemasterGrantsFlyingToTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	glide := pushCatalogPermanent(g, me.ID, "Keen Glidemaster", "Creature — Human Soldier", jmp8KeenGlidemaster, false)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	apaMana(me, "C", "C", "U")
	apaActivate(t, g, me, glide, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !game.HasKeyword(jmp8Live(g, bear), "flying") {
		t.Error("the target did not gain flying")
	}
	if game.HasKeyword(jmp8Live(g, glide), "flying") {
		t.Error("the Glidemaster itself gained flying")
	}
}

func TestJmp8KeenGlidemasterRefusesANoncreatureAndAnUnpaidCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	glide := pushCatalogPermanent(g, me.ID, "Keen Glidemaster", "Creature — Human Soldier", jmp8KeenGlidemaster, false)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	apaMana(me, "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, glide, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("an artifact was a legal target")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	glide2 := pushCatalogPermanent(g2, me2.ID, "Keen Glidemaster", "Creature — Human Soldier", jmp8KeenGlidemaster, false)
	apaMana(me2, "C", "C")
	if err := g2.ActivateCatalogAbility(me2.ID, glide2, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: glide2}}}); err == nil {
		t.Fatal("{2}{U} was paid with two colourless")
	}
}

// --- Lawmage's Binding ---

func TestJmp8LawmagesBindingHasFlashAndStopsEveryActivation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	glide := pushCatalogPermanent(g, opp.ID, "Keen Glidemaster", "Creature — Human Soldier", jmp8KeenGlidemaster, false)
	elf := pushCatalogPermanent(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	// Beginning of combat: a sorcery-speed Aura is refused, a flash one is not.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.Step == game.StepPrecombatMain {
		t.Fatal("still in the main phase")
	}
	arrest := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: arrest, Name: "Arrest", TypeLine: "Enchantment — Aura",
		OracleID: jmp8ArrestOracleForCmp, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, arrest, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: glide}}}); err == nil {
		t.Fatal("Arrest, without flash, was cast outside a main phase")
	}
	cast := func(target uuid.UUID) {
		t.Helper()
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lawmage's Binding", TypeLine: "Enchantment — Aura",
			OracleID: jmp8LawmagesBinding, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
			t.Fatalf("Lawmage's Binding with flash outside a main phase: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	cast(glide)
	cast(elf)
	assertRestrictions(t, g, glide, game.CantAttackOrBlock|game.CantActivate|game.CantActivateMana)
	apaMana(opp, "C", "C", "U")
	if err := g.ActivateCatalogAbility(opp.ID, glide, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: glide}}}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("an enchanted creature activated an ability: %v", err)
	}
	if err := g.ActivateManaAbility(opp.ID, elf, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("an enchanted creature activated a mana ability")
	}
}

func TestJmp8LawmagesBindingEnchantsOnlyACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Lawmage's Binding", "Enchantment — Aura", jmp8LawmagesBinding,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("Lawmage's Binding was cast on an artifact")
	}
}

// --- Leave in the Dust ---

func TestJmp8LeaveInTheDustBouncesToTheOwnerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Leave in the Dust", "Instant", jmp8LeaveInTheDust,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	// castCatalogSpell added the spell to the hand and cast it from
	// there, so the hand is back to its size plus the one card drawn.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the spell, want %d (one drawn)", got, hand+1)
	}
}

func TestJmp8LeaveInTheDustRefusesALandAndFizzlesWithoutDrawing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := seedPermanentWithOracle(g, opp.ID, "Their Forest", "Basic Land — Forest", "")
	if err := castCatalogSpellErr(t, g, "Leave in the Dust", "Instant", jmp8LeaveInTheDust,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Fatal("a land was a legal target")
	}
	// A target gone in response: no bounce and no draw (CR 608.2b).
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castCatalogSpell(t, g, "Leave in the Dust", "Instant", jmp8LeaveInTheDust,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: a fizzled spell draws nothing", got, hand)
	}
}

// --- Legion's Judgment ---

func TestJmp8LegionsJudgmentDestroysOnlyPowerFourOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := pushVanillaCreature(g, opp.ID, "Small", 3, 5)
	big := pushVanillaCreature(g, opp.ID, "Big", 4, 1)
	if err := castCatalogSpellErr(t, g, "Legion's Judgment", "Sorcery", jmp8LegionsJudgment,
		[]game.TargetRef{{Kind: game.TargetCard, ID: small}}); err == nil {
		t.Fatal("a power-3 creature was a legal target")
	}
	castCatalogSpell(t, g, "Legion's Judgment", "Sorcery", jmp8LegionsJudgment,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the power-4 creature survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the power-3 creature was touched")
	}
}

// --- Lofty Denial ---

func jmp8CastVictim(t *testing.T, g *game.Game, opp *game.Player) uuid.UUID {
	t.Helper()
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: victim, Name: "Shock", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts an instant: %v", err)
	}
	return victim
}

func TestJmp8LoftyDenialTaxesFourOnlyWithAFlyerYouControl(t *testing.T) {
	for _, row := range []struct {
		name  string
		setup func(g *game.Game, me, opp *game.Player)
		cost  string
	}{
		{"no flyer", func(*game.Game, *game.Player, *game.Player) {}, "{1}"},
		{"a flyer you control", func(g *game.Game, me, _ *game.Player) { jmp8Flyer(g, me.ID, "My Bird") }, "{4}"},
		{"only an opponent's flyer", func(g *game.Game, _, opp *game.Player) { jmp8Flyer(g, opp.ID, "Their Bird") }, "{1}"},
		{"a non-flyer you control", func(g *game.Game, me, _ *game.Player) { pushVanillaCreature(g, me.ID, "Bear", 2, 2) }, "{1}"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		row.setup(g, me, opp)
		advanceToMain(t, g)
		victim := jmp8CastVictim(t, g, opp)
		castCatalogSpell(t, g, "Lofty Denial", "Instant", jmp8LoftyDenial,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		passPriorityAroundTable(t, g)
		ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
		if ask == nil {
			t.Fatalf("%s: the spell's controller is not asked to pay", row.name)
		}
		if ask.Chooser != opp.ID {
			t.Errorf("%s: the prompt goes to the spell's controller", row.name)
		}
		if ask.PayCost != row.cost {
			t.Errorf("%s: tax is %s, want %s", row.name, ask.PayCost, row.cost)
		}
	}
}

func TestJmp8LoftyDenialCountersWhenDeclinedAndNotWhenPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		_, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		victim := jmp8CastVictim(t, g, opp)
		castCatalogSpell(t, g, "Lofty Denial", "Instant", jmp8LoftyDenial,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		passPriorityAroundTable(t, g)
		if pay {
			apaMana(opp, "C")
		}
		answerPayUnless(t, g, opp.ID, pay)
		passPriorityAroundTable(t, g)
		countered := false
		for _, c := range opp.Graveyard.Cards {
			if c.InstanceID == victim {
				countered = true
			}
		}
		// A resolved instant also goes to the graveyard; the difference is
		// whether the pay prompt was honoured, read off the event log.
		if !countered {
			t.Errorf("pay=%v: the victim is not in its owner's graveyard", pay)
		}
		if got := counteredInLog(g, victim); got == pay {
			t.Errorf("pay=%v: countered=%v", pay, got)
		}
	}
}

// --- Long Road Home ---

func TestJmp8LongRoadHomeReturnsUnderItsOwnerWithACounter(t *testing.T) {
	for _, stolen := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		// A creature an opponent owns and I control comes back under
		// its owner.
		owner := me.ID
		if stolen {
			owner = opp.ID
		}
		bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
			TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: owner, Controller: me.ID})
		castCatalogSpell(t, g, "Long Road Home", "Instant", jmp8LongRoadHome,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) || b16CountNamed(g, "Bear") != 0 {
			t.Fatalf("stolen=%v: the Bear was not exiled until the end step", stolen)
		}
		advanceToEndStepOf(t, g, 0)
		passPriorityAroundTable(t, g)
		back := semestersEndCard(t, g, "Bear")
		if back.InstanceID == bear {
			t.Errorf("stolen=%v: the Bear returned as the same object", stolen)
		}
		if back.Controller != owner {
			t.Errorf("stolen=%v: the Bear returned under %s, want its owner %s", stolen, back.Controller, owner)
		}
		if back.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("stolen=%v: counters %v, want one +1/+1", stolen, back.Counters)
		}
	}
}

func TestJmp8LongRoadHomeRefusesANoncreatureAndLeavesATokenInExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Long Road Home", "Instant", jmp8LongRoadHome,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("an artifact was a legal target")
	}
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier Token",
		TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Long Road Home", "Instant", jmp8LongRoadHome,
		[]game.TargetRef{{Kind: game.TargetCard, ID: tok}})
	passPriorityAroundTable(t, g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Soldier Token"); n != 0 {
		t.Errorf("a token came back (%d)", n)
	}
}

// --- Moment of Heroism ---

func TestJmp8MomentOfHeroismPumpsAndGrantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Moment of Heroism", "Instant", jmp8MomentOfHeroism,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	c := jmp8Live(g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(c, "lifelink") {
		t.Error("Bear did not gain lifelink")
	}
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Moment of Heroism", "Instant", jmp8MomentOfHeroism,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Error("an artifact was a legal target")
	}
}

// --- Prescient Chimera ---

func TestJmp8PrescientChimeraScriesOnYourInstantOrSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Prescient Chimera", "Creature — Chimera", jmp8PrescientChimera, false)
	advanceToMain(t, g)

	// Opponent's instant: no scry for anyone.
	jmp8CastVictim(t, g, opp)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil || scryChoiceFor(g, opp.ID) != nil {
		t.Fatal("an opponent's instant triggered the Chimera")
	}

	// Your creature spell: no scry.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("a creature spell triggered the Chimera")
	}

	// Your sorcery: scry 1.
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("your sorcery did not trigger a scry")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scry %d, want 1", len(c.ScryCards))
	}
}

// --- Roaming Ghostlight ---

func TestJmp8RoamingGhostlightBouncesANonSpiritCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	spirit := b12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castHoldCreature(t, g, "Roaming Ghostlight", "Creature — Spirit", jmp8RoamingGhostlight, 3, 2)
	ch := pickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{spirit, rock} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was a legal target", bad)
		}
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(spirit) {
		t.Error("the Spirit moved")
	}
}

func TestJmp8RoamingGhostlightMayChooseNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Roaming Ghostlight", "Creature — Spirit", jmp8RoamingGhostlight, 3, 2)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("pick no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was bounced with no target chosen")
	}
}
