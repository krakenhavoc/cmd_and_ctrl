package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_g1_cards_test.go — ADR 0108 PR 1, card group 1 (#1886,
// #1887): the "deals N damage … exile it instead" family, the shrink
// pair, and the "can't be regenerated this turn" cards.

type p1g1Spell struct {
	name, typ, oracle string
	dmg               int
}

// every "deals N damage to target creature [or planeswalker]. If that
// creature would die this turn, exile it instead." card.
var p1g1TargetSpells = []p1g1Spell{
	{"Bot Bashing Time", "Sorcery", "14079f05-fbc1-401c-9008-da678656b4c6", 6},
	{"Fanged Flames", "Sorcery", "46515251-2172-4b0a-81ac-4c0120b73360", 4},
	{"Feed the Flames", "Instant", "f474d244-d9be-4580-bf62-f97660e9c1a3", 5},
	{"Flame-Blessed Bolt", "Instant", "6a521eed-0965-4fce-8a11-190dc2863da8", 2},
	{"Magma Spray", "Instant", "fe16f1ab-58b4-4452-abe4-cbd9addd348f", 2},
	{"Obliterating Bolt", "Sorcery", "57fe941c-a830-4570-afe2-18f93c7a7b84", 4},
	{"Puncturing Blow", "Sorcery", "1128d2ab-0b6e-4912-8735-15521bc314e6", 5},
	{"Reduce to Ashes", "Sorcery", "9aab4b32-c5b3-4707-b359-9b5e3b63cd11", 5},
	{"Scorching Dragonfire", "Instant", "d14f313c-fea6-49c4-8197-5b74ee584a6b", 3},
	{"Scorchmark", "Instant", "8dc1148f-c6bc-469c-8d1a-7e3efd2de7e2", 2},
}

// every "deals N damage to any target. If a creature dealt damage this
// way would die this turn, exile it instead." card.
var p1g1AnyTargetSpells = []p1g1Spell{
	{"Annihilating Fire", "Instant", "762f891e-5d88-42b0-8abf-c69f3421011b", 3},
	{"Incendiary Flow", "Sorcery", "91b2ffe8-155d-4b9f-82dd-868cc895856b", 3},
	{"Pillar of Flame", "Sorcery", "468cfc88-a493-44dc-9d0a-63d9cc89c114", 2},
	{"Touch of the Void", "Sorcery", "6b530534-5c02-4874-9256-501102ef8a5f", 3},
	{"Yamabushi's Flame", "Instant", "0462e985-c99e-4404-b212-e9d8baecce72", 3},
}

// "that creature": the damage kills it, or it survives and dies later;
// either way it is exiled, even with the damage prevented.
func TestP1G1TargetCreatureSpellsExileWhatDiesThisTurn(t *testing.T) {
	for _, s := range p1g1TargetSpells {
		t.Run(s.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1].ID
			bear := p1Creature(g, opp, 1, s.dmg)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(bear))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, bear, game.ZoneExile, "the creature the damage killed")

			big := p1Creature(g, opp, 1, s.dmg+10)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(big))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, big, game.ZoneBattlefield, "the survivor")
			p1Destroy(t, g, big)
			p1WantZone(t, g, big, game.ZoneExile, "the survivor destroyed later this turn")

			shielded := p1Creature(g, opp, 1, s.dmg)
			p1ShieldCreature(g, shielded)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(shielded))
			passPriorityAroundTable(t, g)
			p1Destroy(t, g, shielded)
			p1WantZone(t, g, shielded, game.ZoneExile, "the shielded creature destroyed later this turn")
		})
	}
}

// The "or planeswalker" cards also take a planeswalker target.
func TestP1G1CreatureOrPlaneswalkerSpellsTakeAPlaneswalker(t *testing.T) {
	for _, name := range []string{"Fanged Flames", "Flame-Blessed Bolt", "Obliterating Bolt", "Scorching Dragonfire"} {
		var s p1g1Spell
		for _, c := range p1g1TargetSpells {
			if c.name == name {
				s = c
			}
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			pw := ccrLoyaltyWalker(g, g.Seats[1].ID, 5, 0)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(pw))
		})
	}
	// The creature-only cards refuse one.
	g := newCatalogGame(t)
	pw := ccrLoyaltyWalker(g, g.Seats[1].ID, 5, 0)
	if err := castCatalogSpellErr(t, g, "Magma Spray", "Instant", "fe16f1ab-58b4-4452-abe4-cbd9addd348f", pr6Card(pw)); err == nil {
		t.Error("Magma Spray accepted a planeswalker")
	}
}

// "a creature dealt damage this way": the damaged creature is marked and
// one whose damage was all prevented is not.
func TestP1G1AnyTargetSpellsMarkOnlyACreatureTheyDealtDamage(t *testing.T) {
	for _, s := range p1g1AnyTargetSpells {
		t.Run(s.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			bear := p1Creature(g, opp.ID, 1, s.dmg)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(bear))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, bear, game.ZoneExile, "the creature the damage killed")

			big := p1Creature(g, opp.ID, 1, s.dmg+10)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(big))
			passPriorityAroundTable(t, g)
			p1Destroy(t, g, big)
			p1WantZone(t, g, big, game.ZoneExile, "the damaged survivor destroyed later")

			shielded := p1Creature(g, opp.ID, 1, s.dmg)
			p1ShieldCreature(g, shielded)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(shielded))
			passPriorityAroundTable(t, g)
			p1Destroy(t, g, shielded)
			p1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature dealt no damage")

			before := opp.Life
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
			passPriorityAroundTable(t, g)
			if opp.Life != before-s.dmg {
				t.Errorf("player life %d -> %d, want -%d", before, opp.Life, s.dmg)
			}
		})
	}
}

// Elspeth's Smite only targets an attacking or blocking creature.
func TestP1G1ElspethsSmiteNeedsAnAttackerOrBlocker(t *testing.T) {
	const oracle = "3f404fe4-4335-4dcc-ba90-78246c4b880b"
	g := newCatalogGame(t)
	bear := p1Creature(g, g.Seats[1].ID, 1, 3)
	if err := castCatalogSpellErr(t, g, "Elspeth's Smite", "Instant", oracle, pr6Card(bear)); err == nil {
		t.Fatal("Elspeth's Smite accepted a creature that is neither attacking nor blocking")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].AttackingTarget = g.Seats[0].ID
			}
		}
	})
	castCatalogSpell(t, g, "Elspeth's Smite", "Instant", oracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the attacker Smite killed")
}

// Carbonize: a regenerating creature is not regenerated and is exiled;
// both riders hold even when the damage is prevented; a player is only
// dealt damage.
func TestP1G1Carbonize(t *testing.T) {
	const oracle = "a32795a2-a965-4a85-9944-fd9eed464e65"
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := p1Creature(g, opp.ID, 1, 1)
	p1Regenerate(t, g, bear)
	castCatalogSpell(t, g, "Carbonize", "Instant", oracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the regenerating bear")

	shielded := p1Creature(g, opp.ID, 1, 2)
	p1ShieldCreature(g, shielded)
	p1Regenerate(t, g, shielded)
	castCatalogSpell(t, g, "Carbonize", "Instant", oracle, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneExile, "the shielded, regenerating creature destroyed later")

	before := opp.Life
	castCatalogSpell(t, g, "Carbonize", "Instant", oracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life %d -> %d, want -3", before, opp.Life)
	}
}

// Yamabushi's Storm: each creature dealt damage is marked, one whose
// damage was prevented is not.
func TestP1G1YamabushisStorm(t *testing.T) {
	const oracle = "3dfeb0c5-85d6-48fb-b924-d7b77f4b89d6"
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	small := p1Creature(g, opp, 1, 1)
	big := p1Creature(g, opp, 1, 5)
	shielded := p1Creature(g, opp, 1, 5)
	p1ShieldCreature(g, shielded)
	castCatalogSpell(t, g, "Yamabushi's Storm", "Sorcery", oracle, nil)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, small, game.ZoneExile, "the 1/1 the Storm killed")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the damaged 1/5 destroyed later")
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "the shielded 1/5")
}

// Bleed Dry and Ob Nixilis's Cruelty: the shrink kills and exiles; a
// survivor is still exiled if it dies later this turn.
func TestP1G1ShrinkSpellsExileWhatDies(t *testing.T) {
	for _, s := range []p1g1Spell{
		{"Bleed Dry", "Instant", "6c3faf4f-83c1-4098-98b8-bae15d59b0de", 13},
		{"Ob Nixilis's Cruelty", "Instant", "c638957f-88bf-40c2-834c-2be39d73bf41", 5},
	} {
		t.Run(s.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1].ID
			bear := p1Creature(g, opp, s.dmg, s.dmg)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(bear))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, bear, game.ZoneExile, "the creature the shrink killed")

			big := p1Creature(g, opp, s.dmg+5, s.dmg+5)
			castCatalogSpell(t, g, s.name, s.typ, s.oracle, pr6Card(big))
			passPriorityAroundTable(t, g)
			p1WantZone(t, g, big, game.ZoneBattlefield, "the survivor")
			p1Destroy(t, g, big)
			p1WantZone(t, g, big, game.ZoneExile, "the survivor destroyed later")
		})
	}
}

// Engulfing Flames: the target can't be regenerated even though the
// damage is prevented; flashback works.
func TestP1G1EngulfingFlames(t *testing.T) {
	const oracle = "88f935ee-7dbb-4c27-8485-f56b4193fa11"
	g := newCatalogGame(t)
	opp := g.Seats[1].ID
	bear := p1Creature(g, opp, 1, 1)
	p1Regenerate(t, g, bear)
	castCatalogSpell(t, g, "Engulfing Flames", "Instant", oracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating 1/1")

	shielded := p1Creature(g, opp, 1, 3)
	p1ShieldCreature(g, shielded)
	p1Regenerate(t, g, shielded)
	castCatalogSpell(t, g, "Engulfing Flames", "Instant", oracle, pr6Card(shielded))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "the shielded creature, regeneration refused")

	flash := p1Creature(g, opp, 1, 1)
	id := seedGraveyardCard(t, g, "Engulfing Flames", "Instant", oracle)
	if err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: pr6Card(flash),
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, flash, game.ZoneGraveyard, "the creature the flashback killed")
	p1WantZone(t, g, id, game.ZoneExile, "the flashed-back card")
}

// Rage of Purphoros: 4 damage, no regeneration, then a scry.
func TestP1G1RageOfPurphoros(t *testing.T) {
	const oracle = "d6667a93-c706-451f-a83d-ffe5b9b9e53e"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	bear := p1Creature(g, g.Seats[1].ID, 1, 4)
	p1Regenerate(t, g, bear)
	castCatalogSpell(t, g, "Rage of Purphoros", "Sorcery", oracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("no scry 1 was asked")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating 1/4")
}

// Gravebind: the target can't be regenerated, and a card is drawn at the
// next turn's upkeep.
func TestP1G1Gravebind(t *testing.T) {
	const oracle = "0ce01e24-42db-42db-9ba2-38f653383991"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C", "D")
	bear := p1Creature(g, g.Seats[1].ID, 1, 1)
	p1Regenerate(t, g, bear)
	castCatalogSpell(t, g, "Gravebind", "Instant", oracle, pr6Card(bear))
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, bear)
	p1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear Gravebind targeted")

	before := me.Hand.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	advanceToUpkeepOf(t, g, next)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("Gravebind draws at the next upkeep: hand %d -> %d", before, me.Hand.Size())
	}
}

// Hurr Jackal and Furnace Brood: the activated ability marks its target.
func TestP1G1RegenerationBlockers(t *testing.T) {
	for _, c := range []struct{ name, typ, oracle string }{
		{"Hurr Jackal", "Creature — Jackal", "d17f5afa-a884-4b99-aa9e-89ddb3d43b22"},
		{"Furnace Brood", "Creature — Elemental", "cc87a268-253b-4377-bdc5-474940de8878"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			src := pushCatalogPermanent(g, me.ID, c.name, c.typ, c.oracle, false)
			bear := p1Creature(g, g.Seats[1].ID, 1, 1)
			p1Regenerate(t, g, bear)
			if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: pr6Card(bear)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			p1Destroy(t, g, bear)
			p1WantZone(t, g, bear, game.ZoneGraveyard, "the regenerating bear")
		})
	}
}
