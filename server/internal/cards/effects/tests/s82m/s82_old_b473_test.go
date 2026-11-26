package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB473Geargrabber = "a483b1fd-751a-447e-8e4f-a54b1c194d2c"
	oldB473Gargantuan  = "c30ffe3f-ec1b-4af5-b817-7ef9a1c5d2f3"
	oldB473Arbalest    = "748fd75b-c61b-4e02-889d-b7dab23319ef"
	oldB473Architect   = "7af1d7f9-058a-4ccb-96b7-fc391b615c09"
	oldB473Acolyte     = "cf082161-47ff-46b9-926d-3fa5e5727861"
)

func TestOldB473CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB473Geargrabber: "Ogre Geargrabber",
		oldB473Gargantuan:  "Quicksilver Gargantuan",
		oldB473Arbalest:    "Heavy Arbalest",
		oldB473Architect:   "Grand Architect",
		oldB473Acolyte:     "Abuna Acolyte",
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

func TestOldB473GeargrabberStealsAndAttachesTheEquipmentUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Push(g, me.ID, "Ogre Geargrabber", "Creature — Ogre Warrior", oldB473Geargrabber, 4, 4)
	mine := effectstest.B12Permanent(g, me.ID, "My Blade", "Artifact — Equipment")
	blade := effectstest.B12Permanent(g, opp.ID, "Their Blade", "Artifact — Equipment")
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ogre, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{mine, rock} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatal("something other than an opponent's Equipment was accepted as the target")
		}
	}
	effectstest.AnswerPickTarget(t, g, blade)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ApaLive(g, blade).Controller; got != me.ID {
		t.Fatalf("the Equipment is controlled by %s, want its thief", got)
	}
	if !effectstest.AttachedToNow(t, g, blade, ogre) {
		t.Fatal("the stolen Equipment is not attached to the Ogre")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.ApaLive(g, blade)
	if live.Controller != opp.ID {
		t.Errorf("the Equipment is controlled by %s after the turn, want its owner", live.Controller)
	}
	if live.IsAttached() {
		t.Error("the Equipment stayed attached after control went back")
	}
}

func TestOldB473GargantuanEntersAsA7x7CopyOfAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.SeedCopyableCreature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	id := effectstest.CastWithCost(t, g, "Quicksilver Gargantuan", "Creature — Shapeshifter", "{5}{U}{U}", oldB473Gargantuan)
	effectstest.ResolveWithCopyChoice(t, g, theirs)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Hill Giant" || got.Controller != me.ID {
		t.Fatalf("entered as %q under %s, want my Hill Giant", got.Name, got.Controller)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 7 || tough != 7 {
		t.Errorf("the copy is %d/%d, want 7/7", p, tough)
	}
}

func TestOldB473GargantuanDeclinedIsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.SeedCopyableCreature(g, g.Seats[1].ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Quicksilver Gargantuan", TypeLine: "Creature — Shapeshifter", ManaCost: "{5}{U}{U}",
		OracleID: oldB473Gargantuan, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)
	if got := effectstest.CopyBattlefieldCard(t, g, id); got.Name != "Quicksilver Gargantuan" {
		t.Errorf("entered as %q, want itself", got.Name)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 7 || tough != 7 {
		t.Errorf("it is %d/%d, want 7/7", p, tough)
	}
}

func TestOldB473ArbalestGrantsA2DamagePingAndLocksTheUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bow := effectstest.B27Push(g, me.ID, "Heavy Arbalest", "Artifact — Equipment", oldB473Arbalest, "{3}", 0, 0)
	host := effectstest.B16Creature(g, me.ID, "Host", "Creature — Bear", 2, 5, "G")
	effectstest.AdvanceToMain(t, g)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, host); idx >= 0 {
		t.Fatal("the creature has the ability before it is equipped")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, bow, host)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d → %d, want 2 damage", before, opp.Life)
	}
	if !effectstest.ApaLive(g, host).Tapped {
		t.Fatal("the ability's {T} cost was not paid")
	}
	start := g.Turn.Seq
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !effectstest.ApaLive(g, host).Tapped {
		t.Error("the equipped creature untapped in its untap step")
	}
}

func TestOldB473GrandArchitectAnthemColourAndMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arch := effectstest.B27Push(g, me.ID, "Grand Architect", "Creature — Vedalken Artificer", oldB473Architect, "{1}{U}{U}", 1, 3, "U")
	blue := effectstest.B16Creature(g, me.ID, "Blue Bear", "Creature — Bear", 2, 2, "U")
	green := effectstest.B16Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Blue Bear", "Creature — Bear", 2, 2, "U")
	robot := effectstest.B27Push(g, me.ID, "Robot", "Artifact Creature — Construct", "", "{2}", 2, 2)
	for id, want := range map[uuid.UUID]int{arch: 1, blue: 3, green: 2, theirs: 2} {
		if p, _ := effectstest.PtOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
	effectstest.AdvanceToMain(t, g)
	// {U}: target artifact creature becomes blue.
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, arch, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(green)}); err == nil {
		t.Error("a nonartifact creature was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, arch, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(robot)})
	effectstest.PassPriorityAroundTable(t, g)
	if !slices.Contains(effectstest.ApaLive(g, robot).EffectiveColors(), "U") {
		t.Error("the artifact creature did not become blue")
	}
	if p, _ := effectstest.PtOf(t, g, robot); p != 3 {
		t.Errorf("the now-blue artifact creature has power %d, want 3 from the anthem", p)
	}
	// Tap an untapped blue creature: add {C}{C}, only for artifacts.
	if err := g.ActivateManaAbility(me.ID, arch, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{green}}); err == nil {
		t.Error("a green creature paid the cost")
	}
	me.ManaPool = nil
	if err := g.ActivateManaAbility(me.ID, arch, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{blue}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool %d, want 2", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "C" || len(tok.Restrictions) == 0 {
			t.Errorf("mana %+v, want restricted colourless", tok)
		}
	}
	if !effectstest.ApaLive(g, blue).Tapped {
		t.Error("the blue creature was not tapped")
	}
}

func TestOldB473AcolytePreventsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	acolyte := effectstest.B16Creature(g, me.ID, "Abuna Acolyte", "Creature — Cat Cleric", 1, 1, "W")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, acolyte)
		c.OracleID = oldB473Acolyte
	})
	robot := effectstest.B27Push(g, me.ID, "Robot", "Artifact Creature — Construct", "", "{2}", 2, 2)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("the second ability accepted a nonartifact creature")
	}
	effectstest.B16Activate(t, g, me.ID, acolyte, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(robot)})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, robot, 3) })
	live := effectstest.ApaLive(g, robot)
	if live.DamageMarked != 1 {
		t.Errorf("robot has %d damage marked, want 1 (2 of 3 prevented)", live.DamageMarked)
	}
}
