package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const endlessFootAssaultOracle = "77ccafd9-6d76-4658-81ab-bde0454fcc15"

// efaCast puts Endless Foot Assault in the active player's hand with its
// real cost, floats `mana` and casts it strictly, paying squad `squad`
// times.
func efaCast(t *testing.T, g *game.Game, me *game.Player, mana string, squad int) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Endless Foot Assault", OracleID: endlessFootAssaultOracle,
			TypeLine: "Enchantment", ManaCost: "{2}{W}", Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, mana)
	var optional []int
	for i := 0; i < squad; i++ {
		optional = append(optional, 0)
	}
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: optional})
}

func efaAssaults(g *game.Game, me uuid.UUID) (cards, tokens int) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Endless Foot Assault" && c.Controller == me {
			cards++
			if c.IsToken() {
				tokens++
			}
		}
	}
	return cards, tokens
}

func efaNinjas(t *testing.T, g *game.Game, me uuid.UUID) map[uuid.UUID]int {
	t.Helper()
	at := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Ninja" || c.Controller != me {
			continue
		}
		if !c.IsToken() || !c.Tapped || c.AttackingTarget == uuid.Nil {
			t.Errorf("a Ninja is not a tapped and attacking token: %+v", c)
		}
		if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" {
			t.Errorf("a Ninja is %d/%d %v, want a 1/1 black", c.Power, c.Toughness, c.Colors)
		}
		at[c.AttackingTarget]++
	}
	return at
}

func TestEndlessFootAssaultSquadMakesACopyForEachPayment(t *testing.T) {
	for _, tc := range []struct {
		name  string
		mana  string
		squad int
	}{
		{"not paid", "{W}{C}{C}", 0},
		{"paid once", "{W}{W}{C}{C}{C}", 1},
		{"paid three times", "{W}{W}{W}{W}{C}{C}{C}{C}{C}", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := exileCostTable(t)
			if _, err := efaCast(t, g, me, tc.mana, tc.squad); err != nil {
				t.Fatalf("cast paying squad %d times: %v", tc.squad, err)
			}
			if n := kxPoolSize(g, me); n != 0 {
				t.Fatalf("%d mana left in the pool: squad was not charged", n)
			}
			passPriorityAroundTable(t, g)
			cards, tokens := efaAssaults(g, me.ID)
			if cards != 1+tc.squad || tokens != tc.squad {
				t.Errorf("Endless Foot Assaults = %d (%d tokens), want %d (%d tokens)", cards, tokens, 1+tc.squad, tc.squad)
			}
			if len(g.PendingTriggers) != 0 {
				t.Errorf("%d triggers still waiting: a token copy's squad must not trigger", len(g.PendingTriggers))
			}
		})
	}
}

func TestEndlessFootAssaultSquadCostMustBePaidInFull(t *testing.T) {
	g, me, _ := exileCostTable(t)
	id, err := efaCast(t, g, me, "{W}{W}{C}{C}{C}", 2) // enough for one payment, not two
	if err == nil {
		t.Fatal("a cast paying squad twice with mana for once was accepted")
	}
	if !me.Hand.Contains(id) {
		t.Error("the refused cast moved the card")
	}
}

// The ruling: the copies are still made when the enchantment has left
// the battlefield before its squad ability resolves.
func TestEndlessFootAssaultSquadCopiesEvenWhenItLeftInResponse(t *testing.T) {
	g, me, _ := exileCostTable(t)
	id, err := efaCast(t, g, me, "{W}{W}{W}{C}{C}{C}{C}", 2)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 8 && triggersOnStackFrom(g, id) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if triggersOnStackFrom(g, id) != 1 {
		t.Fatal("the squad ability never went on the stack")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
		t.Fatalf("move to graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	cards, tokens := efaAssaults(g, me.ID)
	if cards != 2 || tokens != 2 {
		t.Errorf("Endless Foot Assaults = %d (%d tokens), want the two token copies", cards, tokens)
	}
}

// An Endless Foot Assault that was not cast (here put onto the
// battlefield directly) has no squad payments and makes no copies.
func TestEndlessFootAssaultNotCastMakesNoCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Endless Foot Assault", "Enchantment", endlessFootAssaultOracle, false)
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("%d triggers queued for an Assault that was not cast", len(g.PendingTriggers))
	}
}

func TestEndlessFootAssaultMakesANinjaAttackingEachOpponentOncePerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assault := pushCatalogPermanent(g, me.ID, "Endless Foot Assault", "Enchantment", endlessFootAssaultOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	// Two attackers at one opponent: one trigger, and a Ninja at EVERY
	// opponent, attacked or not.
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	if n := triggersOnStackFrom(g, assault); n != 1 {
		t.Fatalf("%d Endless Foot Assault triggers for one attack, want 1", n)
	}
	passPriorityAroundTable(t, g)

	at := efaNinjas(t, g, me.ID)
	if len(at) != len(g.Seats)-1 {
		t.Fatalf("Ninjas attack %v, want one at each of %d opponents", at, len(g.Seats)-1)
	}
	for _, p := range g.Seats[1:] {
		if at[p.ID] != 1 {
			t.Errorf("%d Ninjas attack %s, want 1", at[p.ID], p.Name)
		}
	}
	// The Ninjas entering attacking are not an attack: nothing more.
	if n := triggersOnStackFrom(g, assault); n != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("the Ninjas entering attacking triggered again (%d on the stack, %d pending)", n, len(g.PendingTriggers))
	}
}

func TestEndlessFootAssaultEachCopyMakesItsOwnNinjas(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Endless Foot Assault", "Enchantment", endlessFootAssaultOracle, false)
	pushCatalogPermanent(g, me.ID, "Endless Foot Assault", "Enchantment", endlessFootAssaultOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	settleOrderingTriggers(t, g) // two triggers of one controller: CR 603.3b
	total := 0
	for _, n := range efaNinjas(t, g, me.ID) {
		total += n
	}
	if want := 2 * (len(g.Seats) - 1); total != want {
		t.Errorf("%d Ninjas, want %d (one per opponent per Assault)", total, want)
	}
}

func TestEndlessFootAssaultIgnoresAnOpponentsAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Endless Foot Assault", "Enchantment", endlessFootAssaultOracle, false)
	advanceToNextSeatsTurn(t, g)
	them := g.Seats[g.Turn.ActiveSeat]
	theirs := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	declareAttack(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ninja" {
			t.Fatalf("an opponent's attack made a Ninja for %s", c.Controller)
		}
	}
}
