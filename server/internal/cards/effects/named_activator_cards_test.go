package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// named_activator_cards_test.go — ADR 0106 §1's 2026-10-07 amendment
// (#1947): "Only your opponents may activate this ability" (Clergy and
// Knight of the Holy Nimbus) and "Only this creature's owner may
// activate this ability" (Personal Incarnation).

const (
	naClergy      = "66566999-f70a-4f14-9bf0-23325295a977"
	naKnight      = "4e92c705-19c2-42df-ad12-808d272c505e"
	naIncarnation = "6e49a5b8-6bc4-4c7b-82c1-957f1fb0ca5f"
)

// The Clergy regenerates every destruction; its controller cannot pay to
// stop that, an opponent can, and then the next destruction sticks.
func TestClergyOnlyOpponentsMayStopItsRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	clergy := pushCatalogPermanent(g, me.ID, "Clergy of the Holy Nimbus", "Creature — Human Cleric", naClergy, false)

	p1Destroy(t, g, clergy)
	p1WantZone(t, g, clergy, game.ZoneBattlefield, "the Clergy regenerates")

	apaMana(me, "C")
	err := g.ActivateCatalogAbility(me.ID, clergy, 0, game.ActivateAbilityParams{Strict: true})
	if !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("the Clergy's controller activating its opponents-only row: err = %v, want ErrCardCallerMismatch", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("a refused activation spent mana: pool %d", len(me.ManaPool))
	}
	p1Destroy(t, g, clergy)
	p1WantZone(t, g, clergy, game.ZoneBattlefield, "the Clergy regenerates again: the refusal changed nothing")

	apaMana(opp, "C")
	apaActivate(t, g, opp, clergy, 0, game.ActivateAbilityParams{})
	p1Destroy(t, g, clergy)
	p1WantZone(t, g, clergy, game.ZoneGraveyard, "the Clergy that can't be regenerated")
}

// The Knight is the same at {2}, and its row is still an opponents-only
// row for the seat that does not control it.
func TestKnightOnlyOpponentsMayStopItsRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushCatalogPermanent(g, me.ID, "Knight of the Holy Nimbus", "Creature — Human Rebel Knight", naKnight, false)

	apaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("controller: err = %v, want ErrCardCallerMismatch", err)
	}
	apaMana(opp, "C")
	if err := g.ActivateCatalogAbility(opp.ID, knight, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("an opponent activated the Knight's {2} row with one mana")
	}
	apaMana(opp, "C")
	apaActivate(t, g, opp, knight, 0, game.ActivateAbilityParams{})
	p1Destroy(t, g, knight)
	p1WantZone(t, g, knight, game.ZoneGraveyard, "the Knight that can't be regenerated")
}

// Personal Incarnation, stolen: its controller cannot activate the row,
// its owner can, the damage goes to the OWNER, and when it dies the
// owner loses half their life, rounded up.
func TestPersonalIncarnationOnlyItsOwnerMayRedirect(t *testing.T) {
	g := newCatalogGame(t)
	thief, owner := g.Seats[0], g.Seats[1]
	inc := apaPush(g, owner.ID, thief.ID, game.Card{
		Name: "Personal Incarnation", OracleID: naIncarnation, TypeLine: "Creature — Avatar Incarnation", Power: 6, Toughness: 6,
	})
	foe := pr7Creature(g, thief.ID, "Pinger", 1, "R")

	if err := g.ActivateCatalogAbility(thief.ID, inc, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("the controller activating an owner-only row: err = %v, want ErrCardCallerMismatch", err)
	}
	pr7Activate(t, g, owner.ID, inc, 0, game.ActivateAbilityParams{Strict: true})
	life := owner.Life
	pr6Damage(t, g, foe, inc, 1)
	if owner.Life != life-1 || pr6Marked(g, inc) != 0 {
		t.Fatalf("owner life %d (want %d), Incarnation damage %d (want 0)", owner.Life, life-1, pr6Marked(g, inc))
	}
	pr6Damage(t, g, foe, inc, 1)
	if pr6Marked(g, inc) != 1 || owner.Life != life-1 {
		t.Fatalf("the charge is spent: Incarnation damage %d (want 1), owner life %d", pr6Marked(g, inc), owner.Life)
	}

	life = owner.Life
	p1Destroy(t, g, inc)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, inc, game.ZoneGraveyard, "the dead Incarnation")
	if want := life - (life+1)/2; owner.Life != want {
		t.Errorf("owner life %d after the dies trigger, want %d (half of %d, rounded up)", owner.Life, want, life)
	}
	if thief.Life != 40 {
		t.Errorf("the thief lost life: %d", thief.Life)
	}
}

// Register refuses a row that names two activators, and an opponents-only
// row that declares a purpose the bot would never read.
func TestNamedActivatorRegistrationGuards(t *testing.T) {
	for name, row := range map[string]ActivatedAbility{
		"two named activators": {Label: "x", Cost: ManaCost("{1}"), AnyPlayer: true, OpponentsOnly: true},
		"owner and opponents":  {Label: "x", Cost: ManaCost("{1}"), OwnerOnly: true, OpponentsOnly: true},
		"opponents with tap":   {Label: "x", Cost: TapCost(), OpponentsOnly: true},
		"opponents purpose":    {Label: "x", Cost: ManaCost("{1}"), OpponentsOnly: true, Purpose: game.Purpose{Draws: 1}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted it", name)
				}
			}()
			checkAnyPlayerAbility("Test", "ability 0", row)
		}()
	}
}
