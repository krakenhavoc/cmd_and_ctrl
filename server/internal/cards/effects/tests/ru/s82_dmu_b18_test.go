package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b18_test.go — Dominaria United Commander, batch 18 (#5487): token
// makers and tribal lords. Each test plays the real catalog entry and
// asserts what its text says, and what it refuses.

const (
	s82d18Baru        = "1cf27d78-d6ae-4fa2-9c8d-c277913f821c"
	s82d18Bladewing   = "cc4bc5d7-ed37-4a80-a5e7-eb68c255cbef"
	s82d18Greensleeve = "2ac2d483-2d2a-4912-8db4-04a5a2080831"
	s82d18Jedit       = "3d80fa33-9032-4c16-8355-2f18a71af67c"
	s82d18Rasputin    = "c16fd2b9-9147-43bc-aa08-6d1bddf1d3e8"
	s82d18Rohgahh     = "a14c692e-dce6-4121-84ec-7fa28ffe4d6b"
	s82d18Rosnakht    = "86f7d513-1f5f-4551-83e6-56d24da5ab39"
	s82d18Zeriam      = "fa1dc21d-f7f6-425f-a4ae-2b0f143091e3"
	s82d18Torsten     = "3b1c6c1e-e298-4f45-8ab7-e0bb5358e11f"
	s82d18Tobias      = "e4544369-910a-41b4-b835-eee341bb48c6"
)

func TestS82Dmu18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82d18Baru:        "Baru, Wurmspeaker",
		s82d18Bladewing:   "Bladewing, Deathless Tyrant",
		s82d18Greensleeve: "Greensleeves, Maro-Sorcerer",
		s82d18Jedit:       "Jedit Ojanen, Mercenary",
		s82d18Rasputin:    "Rasputin, the Oneiromancer",
		s82d18Rohgahh:     "Rohgahh, Kher Keep Overlord",
		s82d18Rosnakht:    "Rosnakht, Heir of Rohgahh",
		s82d18Zeriam:      "Zeriam, Golden Wind",
		s82d18Torsten:     "Torsten, Founder of Benalia",
		s82d18Tobias:      "Tobias, Doomed Conqueror",
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

// s82d18Count counts the permanents of a name a player controls.
func s82d18Count(g *game.Game, name string, controller uuid.UUID) int {
	return effectstest.CountOnBattlefield(g, name, controller)
}

// s82d18Grave puts a card in a player's graveyard.
func s82d18Grave(p *game.Player, name, typeLine string) {
	p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
}

// --- Greensleeves ----------------------------------------------------------

func TestS82Dmu18GreensleevesMakesABadgerPerLandAndCountsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gs := effectstest.PushCatalogPermanent(g, me.ID, "Greensleeves, Maro-Sorcerer", "Legendary Creature — Elemental Sorcerer", s82d18Greensleeve, false)
	effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Badger", me.ID); n != 1 {
		t.Fatalf("%d Badgers, want 1", n)
	}
	badger := effectstest.FindBattlefieldByName(g, "Badger")
	if p := effectstest.EffectivePower(t, g, badger); p != 3 {
		t.Errorf("Badger power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, gs); p != 3 {
		t.Errorf("Greensleeves power %d, want 3 (three lands)", p)
	}
	for _, kw := range []string{"protection from planeswalkers", "protection from Wizards"} {
		if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, gs), kw) {
			t.Errorf("Greensleeves lacks %q", kw)
		}
	}
}

func TestS82Dmu18GreensleevesIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Greensleeves, Maro-Sorcerer", "Legendary Creature — Elemental Sorcerer", s82d18Greensleeve, false)
	effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Badger", me.ID); n != 0 {
		t.Errorf("%d Badgers for an opponent's land", n)
	}
}

// --- Zeriam ----------------------------------------------------------------

func TestS82Dmu18ZeriamMakesAGriffinWhenAnyYourGriffinConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	z := effectstest.PushCatalogPermanent(g, me.ID, "Zeriam, Golden Wind", "Legendary Creature — Griffin", s82d18Zeriam, false)
	other := effectstest.B12Creature(g, me.ID, "Other Griffin", "Creature — Griffin", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Griffin", me.ID); n != 0 {
		t.Fatalf("a Bear's damage made %d Griffins", n)
	}
	effectstest.DealCombatDamageToPlayer(g, z, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamageToPlayer(g, other, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Griffin", me.ID); n != 2 {
		t.Fatalf("%d Griffin tokens after two Griffins connected, want 2", n)
	}
	tok := effectstest.FindBattlefieldByName(g, "Griffin")
	if p := effectstest.EffectivePower(t, g, tok); p != 2 {
		t.Errorf("Griffin token power %d, want 2", p)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, tok), "flying") {
		t.Error("Griffin token has no flying")
	}
}

func TestS82Dmu18ZeriamIgnoresAnOpponentsGriffinAndCombatlessDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	z := effectstest.PushCatalogPermanent(g, me.ID, "Zeriam, Golden Wind", "Legendary Creature — Griffin", s82d18Zeriam, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Griffin", "Creature — Griffin", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: z, Target: opp.ID, Amount: 3})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Griffin", me.ID) + s82d18Count(g, "Griffin", opp.ID); n != 0 {
		t.Errorf("%d Griffin tokens, want none", n)
	}
}

// --- Bladewing -------------------------------------------------------------

func TestS82Dmu18BladewingMakesAKnightPerCreatureCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bw := effectstest.PushCatalogPermanent(g, me.ID, "Bladewing, Deathless Tyrant", "Legendary Creature — Dragon Skeleton", s82d18Bladewing, false)
	s82d18Grave(me, "Dead Bear", "Creature — Bear")
	s82d18Grave(me, "Dead Elf", "Creature — Elf")
	s82d18Grave(me, "Rock", "Artifact")
	s82d18Grave(opp, "Their Bear", "Creature — Bear")
	effectstest.DealCombatDamageToPlayer(g, bw, opp.ID, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Zombie Knight", me.ID); n != 2 {
		t.Fatalf("%d Zombie Knights, want 2", n)
	}
	k := effectstest.FindBattlefieldByName(g, "Zombie Knight")
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, k), "menace") {
		t.Error("Zombie Knight has no menace")
	}
}

func TestS82Dmu18BladewingHitsAPlaneswalkerButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bw := effectstest.PushCatalogPermanent(g, me.ID, "Bladewing, Deathless Tyrant", "Legendary Creature — Dragon Skeleton", s82d18Bladewing, false)
	s82d18Grave(me, "Dead Bear", "Creature — Bear")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	pw := effectstest.B12Push(g, opp.ID, "Some Walker", "Legendary Planeswalker — Test", "", 0, 0)

	dealCombatDamageToPlayerOrCard(g, bw, wall, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Zombie Knight", me.ID); n != 0 {
		t.Fatalf("%d Knights for damage to a creature", n)
	}
	dealCombatDamageToPlayerOrCard(g, bw, pw, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Zombie Knight", me.ID); n != 1 {
		t.Fatalf("%d Knights for damage to a planeswalker, want 1", n)
	}
}

// dealCombatDamageToPlayerOrCard is dealCombatDamageToPlayer for any target id.
func dealCombatDamageToPlayerOrCard(g *game.Game, src, target uuid.UUID, amount int) {
	effectstest.DealCombatDamageToPlayer(g, src, target, amount)
}

// --- Jedit Ojanen ----------------------------------------------------------

func TestS82Dmu18JeditPaysGForACatWarriorOnEnteringAndForAnotherLegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	effectstest.CastAndResolveCreature(t, g, "Jedit Ojanen, Mercenary", "Legendary Creature — Cat Mercenary", s82d18Jedit)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Cat Warrior", me.ID); n != 1 {
		t.Fatalf("%d Cat Warriors after paying for Jedit's own entry, want 1", n)
	}
	cw := effectstest.FindBattlefieldByName(g, "Cat Warrior")
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, cw), "forestwalk") {
		t.Error("Cat Warrior has no forestwalk")
	}
	// Another legendary creature you control.
	effectstest.B06AddMana(me, "G")
	effectstest.CastAndResolveCreature(t, g, "Other Legend", "Legendary Creature — Human", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Cat Warrior", me.ID); n != 2 {
		t.Fatalf("%d Cat Warriors after a second legend, want 2", n)
	}
}

func TestS82Dmu18JeditDecliningOrANonlegendaryMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	effectstest.CastAndResolveCreature(t, g, "Jedit Ojanen, Mercenary", "Legendary Creature — Cat Mercenary", s82d18Jedit)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Cat Warrior", me.ID); n != 0 {
		t.Fatalf("%d Cat Warriors after declining", n)
	}
	effectstest.CastAndResolveCreature(t, g, "Plain Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a nonlegendary creature offered the payment")
	}
}

// --- Rohgahh ---------------------------------------------------------------

func TestS82Dmu18RohgahhPumpsOnlyYourOtherKobolds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	r := effectstest.PushCatalogPermanent(g, me.ID, "Rohgahh, Kher Keep Overlord", "Legendary Creature — Kobold Warrior", s82d18Rohgahh, false)
	mine := effectstest.B12Creature(g, me.ID, "Kobold A", "Creature — Kobold", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Kobold B", "Creature — Kobold", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("my Kobold power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Kobold power %d, want 1", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 1 {
		t.Errorf("a Bear power %d, want 1", p)
	}
	if p := effectstest.EffectivePower(t, g, r); p != 1 {
		t.Errorf("Rohgahh power %d, want printed (the pushed card's 1): he doesn't pump himself", p)
	}
}

func TestS82Dmu18RohgahhKoboldSpellPaysTwoForADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rohgahh, Kher Keep Overlord", "Legendary Creature — Kobold Warrior", s82d18Rohgahh, false)
	effectstest.B06AddMana(me, "R", "R")
	effectstest.CastAndResolveCreature(t, g, "Kobold Spell", "Creature — Kobold", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Dragon", me.ID); n != 1 {
		t.Fatalf("%d Dragons, want 1", n)
	}
	d := effectstest.FindBattlefieldByName(g, "Dragon")
	if p := effectstest.EffectivePower(t, g, d); p != 4 {
		t.Errorf("Dragon power %d, want 4", p)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, d), "flying") {
		t.Error("Dragon has no flying")
	}
}

func TestS82Dmu18RohgahhDecliningTheTwoMakesNoDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rohgahh, Kher Keep Overlord", "Legendary Creature — Kobold Warrior", s82d18Rohgahh, false)
	effectstest.B06AddMana(me, "R", "R")
	effectstest.CastAndResolveCreature(t, g, "Kobold Spell", "Creature — Kobold", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Dragon", me.ID); n != 0 {
		t.Fatalf("%d Dragons after declining", n)
	}
}

func TestS82Dmu18RohgahhDragonSpellMakesAKoboldAndAPlainSpellDoesNot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rohgahh, Kher Keep Overlord", "Legendary Creature — Kobold Warrior", s82d18Rohgahh, false)
	effectstest.CastAndResolveCreature(t, g, "Plain Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Kobolds of Kher Keep", me.ID); n != 0 {
		t.Fatalf("%d Kobold tokens for a Bear", n)
	}
	effectstest.CastAndResolveCreature(t, g, "Small Dragon", "Creature — Dragon", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Kobolds of Kher Keep", me.ID); n != 1 {
		t.Fatalf("%d Kobolds of Kher Keep for a Dragon spell, want 1", n)
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a Dragon spell offered the Kobold payment")
	}
}

// --- Rosnakht --------------------------------------------------------------

func TestS82Dmu18RosnakhtHeroicMakesAKoboldOnlyWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	r := effectstest.PushCatalogPermanent(g, me.ID, "Rosnakht, Heir of Rohgahh", "Legendary Creature — Kobold Warrior", s82d18Rosnakht, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Kobolds of Kher Keep", me.ID); n != 0 {
		t.Fatalf("%d Kobold tokens for a spell at another creature", n)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Kobolds of Kher Keep", me.ID); n != 0 {
		t.Fatalf("%d Kobold tokens for a spell at a player", n)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: r}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Kobolds of Kher Keep", me.ID); n != 1 {
		t.Fatalf("%d Kobolds of Kher Keep after targeting Rosnakht, want 1", n)
	}
}

func TestS82Dmu18RosnakhtBattleCryPumpsTheOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := effectstest.PushCatalogPermanent(g, me.ID, "Rosnakht, Heir of Rohgahh", "Legendary Creature — Kobold Warrior", s82d18Rosnakht, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, r, bear)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("the other attacker has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, r); p != 1 {
		t.Errorf("Rosnakht has power %d, want unchanged", p)
	}
}

// --- Baru ------------------------------------------------------------------

func TestS82Dmu18BaruPumpsAndTramplesYourWurmsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Baru, Wurmspeaker", "Legendary Creature — Human Druid", s82d18Baru, false)
	mine := effectstest.B12Creature(g, me.ID, "Wurm", "Creature — Wurm", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 4, 4)
	if p := effectstest.EffectivePower(t, g, mine); p != 6 {
		t.Errorf("my Wurm power %d, want 6", p)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, mine), "trample") {
		t.Error("my Wurm has no trample")
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 4 {
		t.Errorf("their Wurm power %d, want 4", p)
	}
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, theirs), "trample") {
		t.Error("their Wurm has trample")
	}
}

func TestS82Dmu18BaruCostsLessByTheGreatestWurmPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	baru := effectstest.PushCatalogPermanent(g, me.ID, "Baru, Wurmspeaker", "Legendary Creature — Human Druid", s82d18Baru, false)
	// No Wurm: the full {7}{G}; two mana is refused.
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, baru, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {7}{G} with two mana and no Wurm")
	}
	// A 4/4 Wurm is 6/6 under Baru: {7} less 6 leaves {1}{G}.
	effectstest.B12Creature(g, me.ID, "Wurm", "Creature — Wurm", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, baru, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with a Wurm out: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Wurm", me.ID); n != 2 {
		t.Fatalf("%d Wurms, want 2 (one new token)", n)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool left %v, want it spent", effectstest.PoolColours(me))
	}
}

// --- Rasputin --------------------------------------------------------------

func TestS82Dmu18RasputinGetsACounterPerOpponentAndEachOpponentGetsAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// A 4/1 as printed: a 0/0 stand-in would die before the counters land.
	r := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: r, Name: "Rasputin, the Oneiromancer", TypeLine: "Legendary Creature — Human Wizard",
		OracleID: s82d18Rasputin, Power: 4, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, r, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, r, game.CounterDream); n != 3 {
		t.Errorf("%d dream counters, want 3 (three opponents)", n)
	}
	for _, opp := range g.Seats[1:] {
		if n := s82d18Count(g, "Goblin", opp.ID); n != 1 {
			t.Errorf("opponent %s controls %d Goblins, want 1", opp.Name, n)
		}
	}
	if n := s82d18Count(g, "Goblin", me.ID); n != 0 {
		t.Errorf("I control %d Goblins, want 0", n)
	}
}

func TestS82Dmu18RasputinTurnsCountersIntoColorlessOrAKnight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	r := effectstest.PushCatalogPermanent(g, me.ID, "Rasputin, the Oneiromancer", "Legendary Creature — Human Wizard", s82d18Rasputin, false)
	effectstest.AddCountersForTest(g, r, game.CounterDream, 3)

	effectstest.ActivateManaFor(t, g, me.ID, r, 0, game.ManaAbilityParams{CounterCounts: []int{2}})
	if got := effectstest.PoolColours(me); got["C"] != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool %v, want exactly {C}{C}", got)
	}
	if n := effectstest.CounterCount(g, r, game.CounterDream); n != 1 {
		t.Errorf("%d dream counters left, want 1", n)
	}
	if err := g.ActivateManaAbility(me.ID, r, 0, game.ManaAbilityParams{CounterCounts: []int{0}}); err == nil {
		t.Error("the mana ability took no counters")
	}

	// Untap and make the Knight with the last counter.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(r) })
	if err := g.ActivateCatalogAbility(me.ID, r, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Knight ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Knight", me.ID); n != 1 {
		t.Fatalf("%d Knights, want 1", n)
	}
	k := effectstest.FindBattlefieldByName(g, "Knight")
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, k), "protection from red") {
		t.Error("Knight has no protection from red")
	}
	if n := effectstest.CounterCount(g, r, game.CounterDream); n != 0 {
		t.Errorf("%d dream counters left, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(r) })
	if err := g.ActivateCatalogAbility(me.ID, r, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("made a Knight with no dream counter")
	}
}

// --- Torsten ---------------------------------------------------------------

func TestS82Dmu18TorstenTakesChosenCreaturesAndLandsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var creature, land, spell uuid.UUID
	push := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	spell = push("Top Spell", "Sorcery")
	land = push("Top Land", "Basic Land — Forest")
	creature = push("Top Creature", "Creature — Bear")
	effectstest.CastAndResolveCreature(t, g, "Torsten, Founder of Benalia", "Legendary Creature — Human Soldier", s82d18Torsten)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[creature] || !offered[land] || offered[spell] {
		t.Errorf("offered %v: want the creature and the land, not the sorcery", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, creature, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) || !me.Hand.Contains(land) {
		t.Error("the picked creature and land did not reach the hand")
	}
	if me.Hand.Contains(spell) || !me.Library.Contains(spell) {
		t.Error("the sorcery left the library")
	}
}

func TestS82Dmu18TorstenMakesSevenSoldiersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tor := effectstest.PushCatalogPermanent(g, me.ID, "Torsten, Founder of Benalia", "Legendary Creature — Human Soldier", s82d18Torsten, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tor, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Soldier", me.ID); n != 7 {
		t.Fatalf("%d Soldier tokens, want 7", n)
	}
}

// --- Tobias ----------------------------------------------------------------

func TestS82Dmu18TobiasMakesAZombiePerNontokenCreatureYouLostThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tob := effectstest.PushCatalogPermanent(g, me.ID, "Tobias, Doomed Conqueror", "Legendary Creature — Human Soldier", s82d18Tobias, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	tok := effectstest.B12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b, tok, theirs} {
		id := id
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Zombie", me.ID); n != 0 {
		t.Fatalf("%d Zombies before Tobias died", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tob, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	// Bear A, Bear B and Tobias himself: the token and the opponent's creature don't count.
	if n := s82d18Count(g, "Zombie", me.ID); n != 3 {
		t.Fatalf("%d Zombies, want 3", n)
	}
	z := effectstest.FindBattlefieldByName(g, "Zombie")
	if p := effectstest.EffectivePower(t, g, z); p != 2 {
		t.Errorf("Zombie power %d, want 2", p)
	}
}

func TestS82Dmu18TobiasAloneMakesOneZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tob := effectstest.PushCatalogPermanent(g, me.ID, "Tobias, Doomed Conqueror", "Legendary Creature — Human Soldier", s82d18Tobias, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tob, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := s82d18Count(g, "Zombie", me.ID); n != 1 {
		t.Fatalf("%d Zombies, want 1", n)
	}
}
