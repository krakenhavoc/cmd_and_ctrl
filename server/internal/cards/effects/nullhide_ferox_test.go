package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nullhide_ferox_test.go — ADR 0153 PR 2 (#3168, the card #1861 was filed
// for). The discard line is the Baloth replacement; the other three lines
// already existed and are pinned here as they compose.

const nullhideFeroxOracle = "5b04a337-3152-481e-973f-a11dbd615f93"

func TestNullhideFeroxEntersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	ferox := d.handCard(d.victim, "Nullhide Ferox", "Creature — Beast", "{2}{G}{G}", nullhideFeroxOracle, 6, 6)
	filler := d.filler(d.victim)

	d.theyDiscard(t, ferox, filler)

	c, ok := d.onBattlefield(ferox)
	if !ok {
		t.Fatalf("the Ferox is in %q, want the battlefield", d.zoneOf(ferox))
	}
	if c.Controller != d.victim.ID {
		t.Errorf("the Ferox is controlled by %v, want its owner", c.Controller)
	}
}

func TestNullhideFeroxStaysInTheGraveyardForYourOwnDiscard(t *testing.T) {
	d := newDBOTable(t)
	ferox := d.handCard(d.caster, "Nullhide Ferox", "Creature — Beast", "{2}{G}{G}", nullhideFeroxOracle, 6, 6)
	filler := d.filler(d.caster)

	d.iDiscard(t, ferox, filler)

	if !d.caster.Graveyard.Contains(ferox) {
		t.Fatalf("the Ferox is in %q, want the graveyard", d.zoneOf(ferox))
	}
}

// Its controller can't cast noncreature spells, and can cast creature
// spells.
func TestNullhideFeroxStopsItsControllerCastingNoncreatureSpells(t *testing.T) {
	d := newDBOTable(t)
	apaPush(d.g, d.caster.ID, d.caster.ID, apaCreature("Nullhide Ferox", nullhideFeroxOracle, 6, 6))

	err := castCatalogSpellErr(t, d.g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	if err == nil {
		t.Fatal("cast a noncreature spell with the Ferox out")
	}
	if err := castCatalogSpellErr(t, d.g, "Grizzly Bears", "Creature — Bear", "", nil); err != nil {
		t.Errorf("a creature spell is refused: %v", err)
	}
}

// Only its controller is restricted: an opponent's noncreature spells are
// fine.
func TestNullhideFeroxDoesNotRestrictItsControllersOpponents(t *testing.T) {
	d := newDBOTable(t)
	apaPush(d.g, d.victim.ID, d.victim.ID, apaCreature("Nullhide Ferox", nullhideFeroxOracle, 6, 6))

	err := castCatalogSpellErr(t, d.g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	if err != nil {
		t.Errorf("an opponent's noncreature spell is refused: %v", err)
	}
}

// {2}: any player may make it lose all abilities until end of turn. The
// opponent who pays is "you"; the Ferox is the one that loses its
// abilities, hexproof and the cast restriction included.
func TestNullhideFeroxAnyPlayerMayRemoveItsAbilities(t *testing.T) {
	d := newDBOTable(t)
	ferox := apaPush(d.g, d.victim.ID, d.victim.ID, apaCreature("Nullhide Ferox", nullhideFeroxOracle, 6, 6))
	if !game.HasKeyword(apaLive(d.g, ferox), "hexproof") {
		t.Fatal("the Ferox does not have hexproof to begin with")
	}
	rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: nullhideFeroxOracle})
	if len(rows) != 1 || !rows[0].AnyPlayer {
		t.Fatalf("rows = %+v, want one any-player row", rows)
	}

	apaMana(d.caster, "R", "R")
	apaActivate(t, d.g, d.caster, ferox, 0, game.ActivateAbilityParams{})

	if game.HasKeyword(apaLive(d.g, ferox), "hexproof") {
		t.Error("the Ferox kept hexproof")
	}
}
