package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 367 (S82, #5558): Journey into Nyx Auras — Market Festival,
// Armament of Nyx, Flamespeaker's Will, Cast into Darkness, Aspect of Gorgon,
// Nyx Infusion, Pin to the Earth and Mortal Obstinacy.

const (
	oldB367Festival  = "15c63f9b-d071-4522-b9f0-a1f441a61b81"
	oldB367Armament  = "1a4d16eb-343e-41ca-9363-90dc79da4769"
	oldB367Will      = "3abc09a9-3baa-45ff-b048-53e42d4fa9f2"
	oldB367Darkness  = "4160e621-3f13-486b-af17-728ce85f7ea5"
	oldB367Gorgon    = "4b65276e-7d03-4700-834c-48178fe98353"
	oldB367Infusion  = "52e3bcc0-0c1f-4846-bf52-4af66b7abd6f"
	oldB367Pin       = "605624ee-b7ca-4728-993f-e13b18cc57f3"
	oldB367Obstinacy = "ea29c489-bd04-4f36-8b59-bccaac153152"
)

func TestOldB367CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB367Festival:  "Market Festival",
		oldB367Armament:  "Armament of Nyx",
		oldB367Will:      "Flamespeaker's Will",
		oldB367Darkness:  "Cast into Darkness",
		oldB367Gorgon:    "Aspect of Gorgon",
		oldB367Infusion:  "Nyx Infusion",
		oldB367Pin:       "Pin to the Earth",
		oldB367Obstinacy: "Mortal Obstinacy",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB367Aura(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	return effectstest.PushAuraOnLand(t, g, owner, name, oracle, host)
}

func TestOldB367CastIntoDarknessShrinksPowerAndStopsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	host := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Host", 3, 3)
	effectstest.AuraCast(t, g, "Cast into Darkness", oldB367Darkness, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 1 || tough != 3 {
		t.Errorf("the host is %d/%d, want 1/3", p, tough)
	}
	if effectstest.AuraRestrictions(t, g, host)&game.CantBlock == 0 {
		t.Error("the host can still block")
	}
	_ = me
}

func TestOldB367PinToTheEarthTakesSixPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	host := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Host", 7, 2)
	effectstest.AuraCast(t, g, "Pin to the Earth", oldB367Pin, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 1 || tough != 2 {
		t.Errorf("the host is %d/%d, want 1/2", p, tough)
	}
}

func TestOldB367AspectOfGorgonGivesDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	host := effectstest.PushVanillaCreature(g, me.ID, "Host", 2, 2)
	effectstest.AuraCast(t, g, "Aspect of Gorgon", oldB367Gorgon, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 3 || tough != 5 {
		t.Errorf("the host is %d/%d, want 3/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, host, "deathtouch") {
		t.Error("the host has no deathtouch")
	}
}

func TestOldB367NyxInfusionPumpsAnEnchantmentCreatureAndShrinksOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := effectstest.B12Push(g, me.ID, "Plain", "Creature — Bear", "", 3, 3)
	god := effectstest.B12Push(g, me.ID, "God", "Enchantment Creature — God", "", 3, 3)
	oldB367Aura(t, g, me.ID, "Nyx Infusion", oldB367Infusion, plain)
	oldB367Aura(t, g, me.ID, "Nyx Infusion", oldB367Infusion, god)
	if p, tough := effectstest.PtOf(t, g, plain); p != 1 || tough != 1 {
		t.Errorf("the plain creature is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, god); p != 5 || tough != 5 {
		t.Errorf("the enchantment creature is %d/%d, want 5/5", p, tough)
	}
}

func TestOldB367ArmamentOfNyxDoubleStrikesOnAnEnchantmentAndPreventsOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := effectstest.B12Push(g, me.ID, "Plain", "Creature — Bear", "", 3, 3)
	god := effectstest.B12Push(g, me.ID, "God", "Enchantment Creature — God", "", 3, 3)
	oldB367Aura(t, g, me.ID, "Armament of Nyx", oldB367Armament, plain)
	oldB367Aura(t, g, me.ID, "Armament of Nyx", oldB367Armament, god)
	if !effectstest.HasEffectiveKeyword(t, g, god, "double strike") {
		t.Error("the enchantment creature has no double strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, plain, "double strike") {
		t.Error("the plain creature has double strike")
	}
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(plain, opp.ID, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if opp.Life != life {
		t.Errorf("the plain host dealt %d damage through the prevention", life-opp.Life)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(god, opp.ID, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if opp.Life != life-3 {
		t.Errorf("the enchantment host dealt %d damage, want 3", life-opp.Life)
	}
}

func TestOldB367MarketFestivalAddsTwoManaOfAnyColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	other := effectstest.B31Push(g, me.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	effectstest.PushAuraOnLand(t, g, me.ID, "Market Festival", oldB367Festival, forest)
	effectstest.TapForMana(t, g, me.ID, other)
	if got := len(effectstest.PoolColors(me)); got != 1 {
		t.Errorf("an unenchanted land gave %d mana, want 1", got)
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	for i := 0; i < 2; i++ {
		pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
		if pick == nil {
			t.Fatalf("mana pick %d missing", i+1)
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if got := effectstest.SortedPool(me); len(got) != 4 || got[0] != "G" || got[1] != "R" || got[2] != "R" || got[3] != "U" {
		t.Errorf("pool = %v, want the Island's {U}, the Forest's {G} and two chosen {R}", got)
	}
}

func TestOldB367SacrificeAurasPumpAndDestroyOnCombatDamage(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, victimName, victimType string
		spare                                string
	}{
		{"Flamespeaker's Will", oldB367Will, "Their Rock", "Artifact", "Creature — Wall"},
		{"Mortal Obstinacy", oldB367Obstinacy, "Their Idol", "Enchantment", "Creature — Wall"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, sacrifice := range []bool{true, false} {
				g := effectstest.NewCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				host := effectstest.B12Push(g, me.ID, "Host", "Creature — Bear", "", 2, 2)
				aura := oldB367Aura(t, g, me.ID, tc.name, tc.oracle, host)
				victim := effectstest.B12Push(g, opp.ID, tc.victimName, tc.victimType, "", 0, 0)
				if p, tough := effectstest.PtOf(t, g, host); p != 3 || tough != 3 {
					t.Fatalf("the host is %d/%d, want 3/3", p, tough)
				}
				effectstest.DealCombatDamageToPlayer(g, host, opp.ID, 3)
				effectstest.PassPriorityAroundTable(t, g)
				if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
					effectstest.PickTriggerTarget(t, g, me.ID, victim)
				}
				effectstest.PassPriorityAroundTable(t, g)
				effectstest.AnswerMayChoice(t, g, me.ID, sacrifice)
				effectstest.PassPriorityAroundTable(t, g)
				if sacrifice {
					if g.Battlefield.Contains(aura) {
						t.Error("the Aura was not sacrificed")
					}
					if g.Battlefield.Contains(victim) {
						t.Errorf("the %s was not destroyed", tc.victimType)
					}
				} else {
					if !g.Battlefield.Contains(aura) || !g.Battlefield.Contains(victim) {
						t.Error("declining still sacrificed the Aura or destroyed the target")
					}
				}
			}
		})
	}
}

func TestOldB367SacrificeAurasStayQuietWhenAnotherCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Push(g, me.ID, "Host", "Creature — Bear", "", 2, 2)
	other := effectstest.B12Push(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	oldB367Aura(t, g, me.ID, "Mortal Obstinacy", oldB367Obstinacy, host)
	effectstest.B12Push(g, opp.ID, "Their Idol", "Enchantment", "", 0, 0)
	effectstest.DealCombatDamageToPlayer(g, other, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a creature without the Aura set the trigger off")
	}
}
