package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const trystansCommandOracle = "ecc4e230-a546-4cb4-987c-87300060bbe2"

func TestTrystansCommandIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(trystansCommandOracle)
	if !ok || spec.Name != "Trystan's Command" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// trystansCommandCast casts Trystan's Command strictly, paying {4}{B}{G}.
func trystansCommandCast(t *testing.T, g *game.Game, modes []int, targets ...game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Trystan's Command", TypeLine: "Kindred Sorcery — Elf",
		OracleID: trystansCommandOracle, ManaCost: "{4}{B}{G}", Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	b06AddMana(me, "B", "G", "C", "C", "C", "C")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: modes, Targets: targets})
}

// Copy + return: a token copy of my Elf, and two permanent cards back to
// my hand from my graveyard.
func TestTrystansCommandCopiesAnElfAndReturnsTwoPermanentCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := b12Creature(g, me.ID, "Elvish Warrior", "Creature — Elf Warrior", 2, 3)
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	land := pushGraveyardPermanent(me, "Old Forest", "Basic Land — Forest", "")
	if err := trystansCommandCast(t, g, []int{0, 1},
		modeRef(game.TargetCard, elf, 0, 0),
		modeRef(game.TargetCard, bear, 1, 0),
		modeRef(game.TargetCard, land, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := rfdCountNamed(g, me.ID, "Elvish Warrior"); n != 2 {
		t.Errorf("%d Elvish Warriors, want the original and a token copy", n)
	}
	if !me.Hand.Contains(bear) || !me.Hand.Contains(land) {
		t.Error("both permanent cards return to hand")
	}
}

// Destroy + pump: an enchantment is destroyed, and the target player's
// creatures get +3/+3 and untap; another player's tapped creature stays
// tapped and unpumped.
func TestTrystansCommandDestroysAnEnchantmentAndPumpsAndUntapsAPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == mine || id == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := trystansCommandCast(t, g, []int{2, 3},
		modeRef(game.TargetCard, aura, 0, 0),
		modeRef(game.TargetPlayer, me.ID, 1, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
	c, _ := battlefieldCard(g, mine)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 || c.Tapped {
		t.Errorf("my bear is %d/%d tapped %v, want an untapped 5/5", c.CurrentPower(), c.CurrentToughness(), c.Tapped)
	}
	o, _ := battlefieldCard(g, theirs)
	if o.CurrentPower() != 2 || !o.Tapped {
		t.Error("the other player's creature was pumped or untapped")
	}
}

// The return takes permanent cards only, the destroy a creature or
// enchantment only, and one bullet alone is not "choose two".
func TestTrystansCommandRefusesBadTargetsOrOneMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushGraveyardPermanent(me, "Old Sorcery", "Sorcery", "{1}{B}")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := trystansCommandCast(t, g, []int{1, 3},
		modeRef(game.TargetCard, spell, 0, 0),
		modeRef(game.TargetPlayer, me.ID, 1, 0)); err == nil {
		t.Error("returned a sorcery card")
	}
	if err := trystansCommandCast(t, g, []int{2, 3},
		modeRef(game.TargetCard, rock, 0, 0),
		modeRef(game.TargetPlayer, me.ID, 1, 0)); err == nil {
		t.Error("destroyed an artifact")
	}
	if err := trystansCommandCast(t, g, []int{3}, modeRef(game.TargetPlayer, me.ID, 0, 0)); err == nil {
		t.Error("cast with one mode")
	}
}
