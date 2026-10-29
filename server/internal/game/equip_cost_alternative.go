package game

import (
	"strings"

	"github.com/google/uuid"
)

// equip_cost_alternative.go is "you may pay {0} rather than pay the equip
// cost of the first equip ability you activate each turn" (Kíli the
// Resourceful, #4475): a CR 118.9 alternative cost for an ACTIVATED
// ability, which the S22 AlternativeCost machinery (a cast-time clause
// only spells reach) does not cover.
//
// THE CHOICE IS A ROW. Heart of Kiran's "rather than pay its crew cost"
// set the precedent: an alternative on an activated ability is a second
// row the player picks, not a "how do you want to pay" prompt. Here the
// rows are not the permanent's own text, they belong to whatever equip
// abilities the player's permanents have, so they are derived: while the
// alternative is open to a player, each equip ability of a permanent they
// may activate it on has a twin row whose cost is the alternative and
// nothing else (CR 118.9: the alternative replaces the whole cost; CR
// 601.2f, through 602.2b, still adds increases and takes reductions). The
// twin carries everything else of the row it mirrors, its target clause
// first among them, so "Equip legendary creature {3}" paid with {0} still
// equips only a legendary creature.
//
// A twin is addressed on the wire by an index and a ref of its own,
// AltEquipIndexBase + the mirrored row's index and AltEquipRefPrefix +
// the mirrored row's ref, so every reader that already sends a row's
// index and ref back (the client's menus, the legal enumerator, the bot)
// sends a twin's back unchanged and the activation path decodes it. The
// stack item names the MIRRORED row (ADR 0041 P9): the alternative is
// how the cost was paid, which the item does not need to rebuild itself.
//
// "THE FIRST EQUIP ABILITY YOU ACTIVATE EACH TURN" is a count of the
// player's equip activations this turn (PlayerTurnTally.EquipActivations),
// written at the announce of every equip activation, either row, whether
// or not it resolves. Activating a normal equip first spends the
// alternative for the turn, which is what "first" means.

// EquipCostAlternative is one such clause, contributed by a permanent on
// the battlefield to its controller. Declared on the catalog Spec
// (effects.Spec.EquipCostAlternatives) and read through
// CatalogEquipCostAlternatives, never stored on the Game.
type EquipCostAlternative struct {
	// Label is the clause as printed, for the tile's ability rows. The
	// twin row names the permanent the clause is on.
	Label string
	// Cost is the alternative, in brace notation: "{0}".
	Cost string
	// FirstEachTurn limits it to the first equip ability its controller
	// activates each turn.
	FirstEachTurn bool
	// Condition gates it on its controller ("as long as you have an
	// enduring story"). Nil is always. Caller holds g.mu.
	Condition func(g *Game, controller uuid.UUID) bool
}

// CatalogEquipCostAlternatives is the catalog's clauses for a catalog
// key, wired by the effects package. Nil in a game without a catalog.
var CatalogEquipCostAlternatives func(key string) []EquipCostAlternative

// AltEquipIndexBase is added to a mirrored row's index to address its
// twin. Far beyond any real row count.
const AltEquipIndexBase = 1 << 20

// AltEquipRefPrefix is prepended to a mirrored row's ref to name its twin.
const AltEquipRefPrefix = "alt-equip:"

// equipCostAlternativesOf is the clauses c contributes, as
// attackTaxesOf reads attack taxes.
func equipCostAlternativesOf(c *Card) []EquipCostAlternative {
	if CatalogEquipCostAlternatives == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return CatalogEquipCostAlternatives(key)
}

// EquipCostAlternativeLocked is the alternative open to `activator` for
// an equip ability right now, the name of the permanent offering it, and
// whether there is one: a clause of a permanent they control whose
// condition holds, and, for a first-each-turn clause, no equip ability
// activated by them yet this turn. The first in battlefield order when
// several are open; they are all "{0}" today.
//
// Caller must hold g.mu.
func (g *Game) EquipCostAlternativeLocked(activator uuid.UUID) (EquipCostAlternative, string, bool) {
	if g.Battlefield == nil || activator == uuid.Nil {
		return EquipCostAlternative{}, "", false
	}
	spent := g.TurnTallyFor(activator).EquipActivations > 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != activator {
			continue
		}
		for _, alt := range equipCostAlternativesOf(c) {
			if alt.FirstEachTurn && spent {
				continue
			}
			if alt.Condition != nil && !alt.Condition(g, activator) {
				continue
			}
			return alt, c.Name, true
		}
	}
	return EquipCostAlternative{}, "", false
}

// EquipTwin is one derived row: the shape to price, validate and pay, and
// the wire index and ref that address it.
type EquipTwin struct {
	Shape ActivatedAbilityShape
	// Of is the index of the mirrored row in the source's own list.
	Of    int
	Index int
	Ref   string
}

// EquipTwinsLocked is every twin row `activator` has on `source` right
// now: one per equip ability of the source, while an alternative is open
// to them. The legality of each (whose permanent, timing, targets) is the
// mirrored row's, asked by the caller as for any row.
//
// Caller must hold g.mu.
func (g *Game) EquipTwinsLocked(activator uuid.UUID, source Card, rows []ActivatedAbilityShape, origins AbilityOrigins) []EquipTwin {
	var out []EquipTwin
	var alt EquipCostAlternative
	var from string
	asked := false
	for i, ab := range rows {
		if !ab.Equip {
			continue
		}
		if !asked {
			var ok bool
			if alt, from, ok = g.EquipCostAlternativeLocked(activator); !ok {
				return nil
			}
			asked = true
		}
		out = append(out, EquipTwin{
			Shape: equipTwinShape(ab, alt, from),
			Of:    i,
			Index: AltEquipIndexBase + i,
			Ref:   AltEquipRefPrefix + origins.Ref(i),
		})
	}
	return out
}

// equipTwinShape is the mirrored row with the alternative as its whole
// cost and its label saying so, naming the permanent `from` offering it.
func equipTwinShape(ab ActivatedAbilityShape, alt EquipCostAlternative, from string) ActivatedAbilityShape {
	twin := ab
	twin.Cost = AbilityCost{Mana: alt.Cost}
	twin.Label = ab.Label + " — pay " + alt.Cost + " instead (" + from + ")"
	return twin
}

// decodeEquipTwin splits a wire index and ref into the mirrored row's and
// whether they named a twin. A twin's ref must name a twin too; anything
// else is a stale or confused announcement.
func decodeEquipTwin(index int, ref string) (int, string, bool, error) {
	if index < AltEquipIndexBase {
		if strings.HasPrefix(ref, AltEquipRefPrefix) {
			return 0, "", false, ErrStaleAbilityRef
		}
		return index, ref, false, nil
	}
	if ref != "" && !strings.HasPrefix(ref, AltEquipRefPrefix) {
		return 0, "", false, ErrStaleAbilityRef
	}
	return index - AltEquipIndexBase, strings.TrimPrefix(ref, AltEquipRefPrefix), true, nil
}

// withWireRef is a parked announcement's params with the ref the player
// sent, so the resume decodes the same twin again.
func withWireRef(p, wire ActivateAbilityParams) ActivateAbilityParams {
	p.Ref = wire.Ref
	return p
}
