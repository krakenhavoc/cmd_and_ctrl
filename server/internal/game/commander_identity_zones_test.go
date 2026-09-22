package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// CR 903.4a: colour identity is established before the game begins and
// does not change with the commander's zone. commanderIdentityFor used
// to read only the command zone, so once the commander was cast — the
// normal mid-game state — every "any color" source fell back to printed
// WUBRG order and Command Tower offered all five colours.

const birdsOracleID = "d3a0b660-358c-41bd-9cd2-41fbf3491b1a"

func birdsAndTowerHook(oracleID string) []ManaAbilityShape {
	if s := birdsHook(oracleID); s != nil {
		return s
	}
	return commandTowerHook(oracleID)
}

// replaceCommanderForTest swaps the test deck's commander for a
// creature commander with the given printed cost and returns its ID.
func replaceCommanderForTest(g *Game, p *Player, cost string) uuid.UUID {
	cmdr := NewCard("Test Commander", p.ID)
	cmdr.TypeLine = "Legendary Creature — Elf"
	cmdr.ManaCost = cost
	cmdr.Power, cmdr.Toughness = 2, 2
	cmdr.IsCommander = true
	g.WithWriteLock(func() {
		p.Command.Cards = nil
		p.Command.PushTop(cmdr)
	})
	return cmdr.InstanceID
}

// castCommanderToBattlefieldForTest casts the commander from the
// command zone and resolves it, checking the identity on the stack on
// the way through.
func castCommanderToBattlefieldForTest(t *testing.T, g *Game, p *Player, id uuid.UUID, want []string) {
	t.Helper()
	p.ManaPool.AddMana(ManaToken{Color: "G"}, ManaToken{Color: "G"})
	if err := g.CastSpell(p.ID, id, CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("cast commander: %v", err)
	}
	if got := commanderIdentityFor(g, p).Colors; !reflect.DeepEqual(got, want) {
		t.Errorf("identity with the commander on the stack = %v, want %v", got, want)
	}
	resolveTop(t, g)
	if findBattlefieldCard(g, id) == nil {
		t.Fatalf("commander did not resolve onto the battlefield")
	}
	p.ManaPool = nil
}

func TestCommanderIdentityHoldsOnTheBattlefield(t *testing.T) {
	withCatalogHook(t, birdsAndTowerHook)

	setup := func(t *testing.T) (*Game, *Player) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		id := replaceCommanderForTest(g, p, "{1}{G}")
		castCommanderToBattlefieldForTest(t, g, p, id, []string{"G"})
		if got := commanderIdentityFor(g, p).Colors; !reflect.DeepEqual(got, []string{"G"}) {
			t.Fatalf("identity with the commander on the battlefield = %v, want [G]", got)
		}
		return g, p
	}

	t.Run("Birds offers all five, identity first", func(t *testing.T) {
		g, p := setup(t)
		birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", birdsOracleID)
		if err := g.ActivateManaAbility(p.ID, birds, 0, ManaAbilityParams{}); err != nil {
			t.Fatalf("activate Birds: %v", err)
		}
		c := choiceByKind(g, PendingChoiceMana)
		if c == nil {
			t.Fatalf("no mana choice queued")
		}
		if want := []string{"G", "W", "U", "B", "R"}; !reflect.DeepEqual(c.ColorOptions, want) {
			t.Errorf("Birds options = %v, want %v", c.ColorOptions, want)
		}
	})

	t.Run("Command Tower narrows to the identity", func(t *testing.T) {
		g, p := setup(t)
		tower := pushBattlefieldForTest(g, p.ID, "Command Tower", "Land", commandTowerOracleID)
		if err := g.ActivateManaAbility(p.ID, tower, 0, ManaAbilityParams{}); err != nil {
			t.Fatalf("activate Command Tower: %v", err)
		}
		c := choiceByKind(g, PendingChoiceMana)
		if c == nil {
			t.Fatalf("no mana choice queued")
		}
		if want := []string{"G"}; !reflect.DeepEqual(c.ColorOptions, want) {
			t.Errorf("Command Tower options = %v, want %v (printed: in your commander's color identity)", c.ColorOptions, want)
		}
	})

	t.Run("auto-tap for generic mints the identity colour", func(t *testing.T) {
		g, p := setup(t)
		birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", birdsOracleID)
		g.WithWriteLock(func() {
			g.materializePlanLocked(p, tapPlan{{CardID: birds}}, costFor(t, "{1}"))
		})
		if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "G" {
			t.Errorf("pool = %+v, want one {G}", p.ManaPool)
		}
	})

	t.Run("planner books the narrowed Tower", func(t *testing.T) {
		g, p := setup(t)
		pushBattlefieldForTest(g, p.ID, "Command Tower", "Land", commandTowerOracleID)
		var sources []tapSource
		g.WithWriteLock(func() { sources = gatherTapSources(g, p.ID, nil, 0) })
		if len(sources) != 1 || !reflect.DeepEqual(sources[0].Slots[0].Options, []string{"G"}) {
			t.Errorf("planned Tower slots = %+v, want one slot of [G]", sources)
		}
	})

	t.Run("effect-added mana orders by identity", func(t *testing.T) {
		g, p := setup(t)
		g.WithWriteLock(func() {
			if err := g.AddManaForEffect(p.ID, uuid.New(), "{W|U|B|R|G}"); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
		})
		c := choiceByKind(g, PendingChoiceMana)
		if c == nil || !reflect.DeepEqual(c.ColorOptions, []string{"G", "W", "U", "B", "R"}) {
			t.Errorf("effect options = %+v, want [G W U B R]", c)
		}
	})
}

// Every zone a commander can be in keeps the identity, a stolen
// commander still belongs to its owner, and an opponent's commander
// never lends its identity.
func TestCommanderIdentityEveryZone(t *testing.T) {
	g := newActiveGame(t)
	p, opp := g.Seats[0], g.Seats[1]
	id := replaceCommanderForTest(g, p, "{1}{B}")
	replaceCommanderForTest(g, opp, "{R}")

	move := func(from, to *Zone) {
		t.Helper()
		g.WithWriteLock(func() {
			if _, err := MoveCard(from, to, id); err != nil {
				t.Fatalf("move commander: %v", err)
			}
		})
	}
	check := func(where string) {
		t.Helper()
		if got := commanderIdentityFor(g, p).Colors; !reflect.DeepEqual(got, []string{"B"}) {
			t.Errorf("identity with the commander in %s = %v, want [B]", where, got)
		}
	}
	check("the command zone")
	move(p.Command, g.Battlefield)
	check("the battlefield")
	findBattlefieldCard(g, id).Controller = opp.ID
	check("the battlefield under an opponent's control")
	if got := commanderIdentityFor(g, opp).Colors; !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("opponent's identity while controlling our commander = %v, want [R]", got)
	}
	move(g.Battlefield, p.Graveyard)
	check("the graveyard")
	move(p.Graveyard, g.Exile)
	check("exile")
	move(g.Exile, p.Hand)
	check("the hand")
	move(p.Hand, p.Library)
	check("the library")
}

// Partner commanders: the identity is the union, wherever each one is.
func TestCommanderIdentityUnionsPartners(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	replaceCommanderForTest(g, p, "{W}")
	second := NewCard("Second Partner", p.ID)
	second.TypeLine = "Legendary Creature — Human"
	second.ManaCost = "{2}{U}"
	second.IsCommander = true
	g.WithWriteLock(func() { g.Battlefield.PushTop(second) })
	if got := commanderIdentityFor(g, p).Colors; !reflect.DeepEqual(got, []string{"W", "U"}) {
		t.Errorf("partner identity = %v, want [W U]", got)
	}
}

// The executor walks the plan greedily (pickColorForSlot takes the most
// restrictive pending pip it can pay). Under a mono-W commander Birds
// lists W first, so a plan that pays {W}{B} from Birds and a Plains must
// still leave the {W} to the Plains and pay the off-identity {B} from
// Birds.
func TestAutoTapBirdsPaysOffIdentityPipBesideABasic(t *testing.T) {
	withCatalogHook(t, birdsHook)
	for _, cost := range []string{"{W}{B}", "{B}{W}"} {
		t.Run(cost, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			p := g.Seats[0]
			setCommanderCostForTest(t, p, "{1}{W}")
			pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", birdsOracleID)
			pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")
			id := pushTypedCardToHandWithCost(p, "Vindicate", "Sorcery", cost)
			if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
				t.Fatalf("auto-tap %s from Birds + Plains under a mono-W commander: %v", cost, err)
			}
			if len(p.ManaPool) != 0 {
				t.Errorf("post-spend pool = %+v, want empty", p.ManaPool)
			}
		})
	}
}
