package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s46Lamentation   = "88e43e62-8bce-41b3-ac2d-47043fcc7077"
	s46MistDancer    = "bf76ba01-e06f-4451-a4b4-cf548b43eafc"
	s46Jubilation    = "2cf04076-adc4-4831-a3b8-9120d516f204"
	s46Impulsivity   = "02b6900b-219f-4832-8f93-6ef27ee76c0c"
	s46Belonging     = "0ec88e35-191b-411b-9eb0-2dfef7ae4174"
	s46Subterfuge    = "687cfdf7-8f35-464a-8d6e-9f2d769ed6b4"
	s46Scion         = "75fddffb-68f4-4b65-bf3b-4e2ada62a522"
	s46MassMysteries = "3ceffadd-c32d-4508-b307-e03bf3c8e7b8"
	s46Scurry        = "3beed054-a099-4887-b1bd-39cf9e514e94"
	s46EchoingAssalt = "d15dba3a-3b85-43fb-b5f3-c4d8d1e23b1d"
	s46Lightbringer  = "65ea7eb1-751b-4ecd-83cb-e9b6015ec194"
	s46KariZev       = "786baa29-afd4-40f0-95c8-920a972b9175"
)

func TestCmdrS46CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s46Lamentation: "Lamentation", s46MistDancer: "Mist Dancer", s46Jubilation: "Jubilation",
		s46Impulsivity: "Impulsivity", s46Belonging: "Belonging", s46Subterfuge: "Subterfuge",
		s46Scion: "Scion of Calamity", s46MassMysteries: "Mass of Mysteries", s46Scurry: "Scurry of Squirrels",
		s46EchoingAssalt: "Echoing Assault", s46Lightbringer: "Soaring Lightbringer", s46KariZev: "Kari Zev, Skyship Raider",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec.Name)
		}
	}
}

// s46Tokens lists the tokens `controller` has named `name`.
func s46Tokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller && IsToken(c) {
				out = append(out, c)
			}
		}
	})
	return out
}

// s46Encore activates encore from the graveyard and resolves it.
func s46Encore(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tgh int, mana string) (uuid.UUID, *game.Player) {
	t.Helper()
	id, me := activateFromGraveyard(t, g, name, typeLine, oracle, p, tgh, mana, game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatalf("%s: encore's cost exiles the card", name)
	}
	passPriorityAroundTable(t, g)
	return id, me
}

func TestCmdrS46LamentationDestroysAnOpponentsCreatureAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Lamentation", "Creature — Elemental Incarnation", s46Lamentation, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(mine) {
		t.Fatal("only the opponent's creature should be destroyed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestCmdrS46LamentationCannotTargetYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Lamentation", "Creature — Elemental Incarnation", s46Lamentation, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was offered with no opposing creature")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || me.Life != life {
		t.Error("the trigger did something with no legal target")
	}
}

func TestCmdrS46LamentationEncoreMakesATokenPerOpponentWhoDestroys(t *testing.T) {
	g := newCatalogGame(t)
	_, me := s46Encore(t, g, "Lamentation", "Creature — Elemental Incarnation", s46Lamentation, 5, 4, "{C}{C}{C}{C}{C}{C}{B}{B}")
	if n := len(s46Tokens(g, me.ID, "Lamentation")); n != 3 {
		t.Fatalf("%d tokens, want one per opponent", n)
	}
}

func TestCmdrS46MistDancerPumpsAndGivesFlyingToOtherMerfolkOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dancer := b12Push(g, me.ID, "Mist Dancer", "Creature — Merfolk Wizard", s46MistDancer, 3, 3)
	fish := pushTribalCreature(g, me.ID, "Fish", "Creature — Merfolk", 2, 2)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := pushTribalCreature(g, g.Seats[1].ID, "Their Fish", "Creature — Merfolk", 2, 2)
	if got := effectivePowerWithCounters(t, g, fish); got != 3 || !hasEffectiveKeyword(t, g, fish, "flying") {
		t.Errorf("my Merfolk power %d flying %v, want 3 and flying", got, hasEffectiveKeyword(t, g, fish, "flying"))
	}
	if got := effectivePowerWithCounters(t, g, dancer); got != 3 {
		t.Errorf("Mist Dancer itself is %d, want 3 (other Merfolk only)", got)
	}
	if got := effectivePowerWithCounters(t, g, bear); got != 2 || hasEffectiveKeyword(t, g, bear, "flying") {
		t.Errorf("a Bear was changed")
	}
	if got := effectivePowerWithCounters(t, g, theirs); got != 2 || hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Errorf("an opposing Merfolk was changed")
	}
}

func TestCmdrS46MistDancerEncoreTokenIsAMerfolkLord(t *testing.T) {
	g := newCatalogGame(t)
	_, me := s46Encore(t, g, "Mist Dancer", "Creature — Merfolk Wizard", s46MistDancer, 3, 3, "{C}{C}{C}{C}{C}{U}{U}")
	toks := s46Tokens(g, me.ID, "Mist Dancer")
	if len(toks) != 3 {
		t.Fatalf("%d tokens, want 3", len(toks))
	}
	if got := effectivePowerWithCounters(t, g, toks[0].InstanceID); got != 5 {
		t.Errorf("a token lord sees two others: power %d, want 5", got)
	}
}

func TestCmdrS46JubilationPumpsOnlyTheCreaturesYouHaveThen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Jubilation", "Creature — Elemental Incarnation", s46Jubilation, nil)
	passPriorityAroundTable(t, g)
	if got := effectivePowerWithCounters(t, g, bear); got != 4 || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Errorf("my Bear power %d trample %v, want 4 and trample", got, hasEffectiveKeyword(t, g, bear, "trample"))
	}
	if got := effectivePowerWithCounters(t, g, theirs); got != 2 || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Errorf("an opposing creature was pumped")
	}
	later := pushVanillaCreature(g, me.ID, "Later", 2, 2)
	if got := effectivePowerWithCounters(t, g, later); got != 2 {
		t.Errorf("a creature that entered afterwards is %d, want 2", got)
	}
	advanceTo(t, g, game.StepCleanup)
	if got := effectivePowerWithCounters(t, g, bear); got != 2 {
		t.Errorf("after the turn the Bear is %d, want 2", got)
	}
}

func TestCmdrS46JubilationEncoreTokensEachPump(t *testing.T) {
	g := newCatalogGame(t)
	_, me := s46Encore(t, g, "Jubilation", "Creature — Elemental Incarnation", s46Jubilation, 5, 5, "{C}{C}{C}{C}{C}{C}{C}{G}{G}")
	passPriorityAroundTable(t, g)
	toks := s46Tokens(g, me.ID, "Jubilation")
	if len(toks) != 3 {
		t.Fatalf("%d tokens, want 3", len(toks))
	}
	// Each token's enters trigger pumps every creature: +2/+2 three times.
	if got := effectivePowerWithCounters(t, g, toks[0].InstanceID); got != 11 {
		t.Errorf("a token's power %d, want 11 (5 + three +2s)", got)
	}
}

func TestCmdrS46ImpulsivityCastsAnyGravesyardSpellFreeAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Instant", TypeLine: "Instant",
		ManaCost: "{3}{U}{U}", Owner: opp.ID, Controller: opp.ID})
	creature := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Old Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Impulsivity", "Creature — Elemental Incarnation", s46Impulsivity, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, spell)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("free cast from the opponent's graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("the spell was not exiled instead of going to a graveyard")
	}
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("a creature card was cast from a graveyard")
	}
}

func TestCmdrS46BelongingMakesThreeChangelingShapeshifters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Belonging", "Creature — Elemental Incarnation", s46Belonging, nil)
	passPriorityAroundTable(t, g)
	toks := s46Tokens(g, me.ID, "Shapeshifter")
	if len(toks) != 3 {
		t.Fatalf("%d Shapeshifters, want 3", len(toks))
	}
	for _, c := range toks {
		if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 || !hasEffectiveKeyword(t, g, c.InstanceID, "changeling") {
			t.Errorf("Shapeshifter = %+v, want a colorless 1/1 changeling", c)
		}
	}
}

func TestCmdrS46BelongingEncoreTokensMakeNineShapeshifters(t *testing.T) {
	g := newCatalogGame(t)
	_, me := s46Encore(t, g, "Belonging", "Creature — Elemental Incarnation", s46Belonging, 6, 6, "{C}{C}{C}{C}{C}{C}{W}{W}")
	passPriorityAroundTable(t, g)
	if n := len(s46Tokens(g, me.ID, "Shapeshifter")); n != 9 {
		t.Errorf("%d Shapeshifters, want 9 (three encore tokens, three each)", n)
	}
}

func TestCmdrS46SubterfugeGivesFlyingAndADrawTriggerUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := pushVanillaCreature(g, me.ID, "My Bear", 3, 3)
	castCatalogSpell(t, g, "Subterfuge", "Creature — Elemental Incarnation", s46Subterfuge, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the target did not gain flying")
	}
	hand := me.Hand.Size()
	life := opp.Life
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 || me.Hand.Size() != hand+3 {
		t.Errorf("life %d->%d hand %d->%d, want 3 damage and three cards", life, opp.Life, hand, me.Hand.Size())
	}
	advanceTo(t, g, game.StepCleanup)
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

func TestCmdrS46ScionOfCalamityMyriadsAndDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := g.Seats[0], g.Seats[1]
	scion := b12Push(g, me.ID, "Scion of Calamity", "Creature — Dinosaur", s46Scion, 5, 5)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp1.ID, Controller: opp1.ID})
	other := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Other Rock", TypeLine: "Artifact",
		Owner: g.Seats[2].ID, Controller: g.Seats[2].ID})
	declareAttack(t, g, opp1.ID, scion)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, false)
	copies := 0
	for _, c := range s46Tokens(g, me.ID, "Scion of Calamity") {
		if c.AttackingTarget == g.Seats[2].ID {
			copies++
		}
	}
	if copies != 1 {
		t.Fatalf("%d myriad copies, want the one accepted", copies)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact of the player the Scion hit survived")
	}
	_ = other
}

func TestCmdrS46MassOfMysteriesGrantsMyriadToAnotherElementalUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Mass of Mysteries", "Legendary Creature — Elemental", s46MassMysteries, 5, 5)
	elemental := pushTribalCreature(g, me.ID, "Test Elemental", "Creature — Elemental", 2, 2)
	bear := pushTribalCreature(g, me.ID, "Test Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, elemental, bear)
	passPriorityAroundTable(t, g)
	// Only the Elemental has myriad: two other opponents are asked.
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, true)
	if n := len(s46Tokens(g, me.ID, "Test Elemental")); n != 2 {
		t.Errorf("%d Elemental copies, want 2", n)
	}
	if n := len(s46Tokens(g, me.ID, "Test Bear")); n != 0 {
		t.Errorf("a non-Elemental got myriad: %d copies", n)
	}
}

func TestCmdrS46ScurryOfSquirrelsAsksEachOpponentTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scurry := b12Push(g, me.ID, "Scurry of Squirrels", "Creature — Squirrel Scout", s46Scurry, 2, 2)
	declareAttack(t, g, opp.ID, scurry)
	passPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		answerMayChoice(t, g, me.ID, true)
	}
	if n := len(s46Tokens(g, me.ID, "Scurry of Squirrels")); n != 4 {
		t.Fatalf("%d copies, want 4 (two for each of two other opponents)", n)
	}
	for _, c := range s46Tokens(g, me.ID, "Scurry of Squirrels") {
		if c.AttackingTarget == opp.ID {
			t.Error("a copy attacks the defending player")
		}
	}
}

func TestCmdrS46ScurryCombatDamagePutsACounterOnATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scurry := b12Push(g, me.ID, "Scurry of Squirrels", "Creature — Squirrel Scout", s46Scurry, 2, 2)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	declareAttack(t, g, opp.ID, scurry)
	passPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		answerMayChoice(t, g, me.ID, false)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if plusOneCounters(g, bear) != 1 {
		t.Errorf("counters on the Bear: %d, want 1", plusOneCounters(g, bear))
	}
}

func TestCmdrS46EchoingAssaultCopiesAnAttackerAs11AndSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Echoing Assault", TypeLine: "Enchantment",
		OracleID: s46EchoingAssalt, Owner: me.ID, Controller: me.ID})
	big := pushTribalCreature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	declareAttack(t, g, opp.ID, big)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)
	toks := s46Tokens(g, me.ID, "Big Bear")
	if len(toks) != 1 {
		t.Fatalf("%d copies, want 1", len(toks))
	}
	c := toks[0]
	if c.Power != 1 || c.Toughness != 1 || !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("copy = %dx%d tapped=%v attacking=%v, want a tapped 1/1 attacking the same player", c.Power, c.Toughness, c.Tapped, c.AttackingTarget)
	}
	if !hasEffectiveKeyword(t, g, c.InstanceID, "menace") {
		t.Error("a creature token you control has menace")
	}
	if hasEffectiveKeyword(t, g, big, "menace") {
		t.Error("a nontoken creature has menace")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c.InstanceID) {
		t.Error("the copy survived to the next end step")
	}
}

func TestCmdrS46EchoingAssaultIgnoresTokensAndPlaneswalkerAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Echoing Assault", TypeLine: "Enchantment",
		OracleID: s46EchoingAssalt, Owner: me.ID, Controller: me.ID})
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 5, Counters: map[string]int{"loyalty": 5}, Owner: opp.ID, Controller: opp.ID})
	big := pushTribalCreature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	declareAttack(t, g, walker, big)
	passPriorityAroundTable(t, g)
	if n := len(s46Tokens(g, me.ID, "Big Bear")); n != 0 {
		t.Errorf("%d copies after an attack on a planeswalker, want 0", n)
	}
}

func TestCmdrS46SoaringLightbringerMakesATappedAttackingGlimmerAndGrantsFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lb := b12Push(g, me.ID, "Soaring Lightbringer", "Enchantment Creature — Bird Glimmer", s46Lightbringer, 4, 5)
	ench := pushTribalCreature(g, me.ID, "Ench Bear", "Enchantment Creature — Bear", 2, 2)
	plain := pushTribalCreature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, ench, "flying") || hasEffectiveKeyword(t, g, plain, "flying") {
		t.Error("only other enchantment creatures gain flying")
	}
	declareAttack(t, g, opp.ID, lb)
	passPriorityAroundTable(t, g)
	toks := s46Tokens(g, me.ID, "Glimmer")
	if len(toks) != 1 {
		t.Fatalf("%d Glimmers, want 1", len(toks))
	}
	c := toks[0]
	if !c.Tapped || c.AttackingTarget != opp.ID || c.Power != 1 || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
		t.Errorf("Glimmer = %+v, want a tapped flying 1/1 attacking the attacked player", c)
	}
}

func TestCmdrS46SoaringLightbringerIgnoresAnAttackOnAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lb := b12Push(g, me.ID, "Soaring Lightbringer", "Enchantment Creature — Bird Glimmer", s46Lightbringer, 4, 5)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 5, Counters: map[string]int{"loyalty": 5}, Owner: opp.ID, Controller: opp.ID})
	declareAttack(t, g, walker, lb)
	passPriorityAroundTable(t, g)
	if n := len(s46Tokens(g, me.ID, "Glimmer")); n != 0 {
		t.Errorf("%d Glimmers after an attack on a planeswalker, want 0", n)
	}
}

func TestCmdrS46KariZevMakesRagavanAttackingAndExilesItAtEndOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kari := b12Push(g, me.ID, "Kari Zev, Skyship Raider", "Legendary Creature — Human Pirate", s46KariZev, 1, 3)
	declareAttack(t, g, opp.ID, kari)
	passPriorityAroundTable(t, g)
	passAndSettleAttackSplit(t, g, me.ID)
	toks := s46Tokens(g, me.ID, "Ragavan")
	if len(toks) != 1 {
		t.Fatalf("%d Ragavans, want 1", len(toks))
	}
	c := toks[0]
	if !c.Tapped || c.AttackingTarget == uuid.Nil || c.Power != 2 || c.Toughness != 1 || !c.IsLegendary() {
		t.Errorf("Ragavan = %+v, want a legendary tapped attacking 2/1", c)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c.InstanceID) {
		t.Error("Ragavan survived the end of combat")
	}
	if !g.Battlefield.Contains(kari) {
		t.Error("Kari Zev was exiled")
	}
}
