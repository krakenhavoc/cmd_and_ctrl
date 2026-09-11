package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// targets.go — S20 sub-PR 1: the predicate library catalog cards
// compose their `Spec.Targets` from, plus constructors for the
// common shapes. A card file reads like its oracle text:
//
//	Targets: TargetCreature(NonBlack()),          // Doom Blade
//	Targets: TargetAny(),                         // Lightning Bolt
//	Targets: TargetSpell(Noncreature()),          // Negate
//	Targets: TargetCardInGraveyard(YourOwn()),    // Regrowth
//
// Predicates receive the live game (read-only — they run under
// g.mu), the caster, and the candidate. Compose with And / Or /
// Not. Keep them pure: no allocation-heavy scans per candidate; the
// legal-target walk calls them once per card on the battlefield /
// stack / graveyards on every snapshot.

// CardPredicate decides whether a candidate card is a legal target
// for the caster.
type CardPredicate func(g *game.Game, caster uuid.UUID, c game.Card) bool

// PlayerPredicate decides whether a candidate player is a legal
// target for the caster.
type PlayerPredicate func(g *game.Game, caster uuid.UUID, p *game.Player) bool

// --- composers -------------------------------------------------

// And passes when every predicate passes (empty list passes).
func And(preds ...CardPredicate) CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		for _, p := range preds {
			if !p(g, caster, c) {
				return false
			}
		}
		return true
	}
}

// Or passes when any predicate passes (empty list fails).
func Or(preds ...CardPredicate) CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		for _, p := range preds {
			if p(g, caster, c) {
				return true
			}
		}
		return false
	}
}

// Not inverts a predicate.
func Not(pred CardPredicate) CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return !pred(g, caster, c)
	}
}

// --- type predicates (post-layer, so type-changers compose) ----

func Creature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() }
}

func Artifact() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsArtifact() }
}

func Enchantment() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsEnchantment() }
}

func Land() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() }
}

func Instant() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsInstant() }
}

func Sorcery() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsSorcery() }
}

func Planeswalker() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsPlaneswalker() }
}

// Permanent passes for any permanent-type card (on the battlefield
// every card is one; on the stack it distinguishes permanent spells
// from instants / sorceries).
func Permanent() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsPermanent() }
}

// Nonland passes for anything that isn't a land.
func Nonland() CardPredicate { return Not(Land()) }

// Noncreature passes for anything that isn't a creature (Negate's
// "noncreature spell").
func Noncreature() CardPredicate { return Not(Creature()) }

// --- colour predicates -------------------------------------------

// OfColor passes when the card is the given colour letter.
func OfColor(color string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasColor(color) }
}

// NonBlack — Doom Blade. Likewise NonBlue etc. via Not(OfColor("U")).
func NonBlack() CardPredicate { return Not(OfColor("B")) }

// Colorless passes for cards with no colour.
func Colorless() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsColorless() }
}

// --- stat predicates ---------------------------------------------

// PowerLE passes when the creature's current power is ≤ n.
func PowerLE(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.CurrentPower() <= n }
}

// PowerGE passes when the creature's current power is ≥ n.
func PowerGE(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.CurrentPower() >= n }
}

// ManaValueLE passes when the card's mana value is ≤ n (Swan Song
// doesn't need it, but Counterspell variants and Abrupt Decay do).
func ManaValueLE(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		cost, err := game.ParseCost(c.ManaCost)
		if err != nil {
			return false
		}
		return cost.Generic+len(cost.Required) <= n
	}
}

// --- keyword predicates ------------------------------------------

// HasKeyword passes when the candidate has the named keyword,
// reading through game.HasKeyword so a keyword GRANTED by a Layer 6
// static (The Wandering Rescuer's hexproof, an Equipment's flying)
// counts exactly as a printed one does. kw must be one of the
// engine's canonical lowercase tokens — "flying", "reach",
// "first strike", "double strike", "deathtouch", "lifelink",
// "trample", "vigilance", "menace", "defender", "haste", "flash",
// "hexproof", "shroud". A token outside that set can never be true,
// because the deck importer filters Scryfall's array against the
// same table.
//
// This is for clauses that NAME a keyword — "target creature with
// flying", "destroy target creature without flying". It is NOT how
// hexproof and shroud are enforced: those are a rule, applied to
// every targeted clause by game.CanBeTargetedBy at the choke point,
// and a card does not opt in.
func HasKeyword(kw string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return game.HasKeyword(&c, kw)
	}
}

// WithoutKeyword is HasKeyword's negation, spelled out because
// "creature without flying" is how the clause reads on the card.
func WithoutKeyword(kw string) CardPredicate { return Not(HasKeyword(kw)) }

// --- controller / owner predicates ------------------------------

// YouControl passes for cards the caster controls.
func YouControl() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Controller == caster }
}

// OpponentControls passes for cards controlled by someone other
// than the caster.
func OpponentControls() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Controller != caster }
}

// YouOwn passes for cards the caster owns (graveyard targets:
// "return target card from your graveyard").
func YouOwn() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Owner == caster }
}

// --- stack-origin predicates -------------------------------------

// CastFromOwnersHand passes for a spell on the stack that was cast
// from its owner's hand — the ordinary case, and the one Wash Away's
// bracketed clause excludes with Not(CastFromOwnersHand()).
//
// "Its owner's" needs no separate check here: the engine only lets a
// player cast out of their OWN hand (castSourceZoneLocked resolves
// "hand" to the caster's), so a spell cast from a hand was cast from
// its owner's hand. Casts from the command zone and from an impulse
// exile grant are the ones this excludes, and a commander is exactly
// the case the printed clause is aimed at.
//
// A stack card with no announce record can't happen through the cast
// path, but if one appears it reads as an ordinary hand cast: the
// restrictive answer is the one that can't over-permit a narrow
// clause. Added in S22.
func CastFromOwnersHand() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return true
		}
		return item.CastFromZone == game.ZoneHand
	}
}

// NotSelf excludes the spell / source itself — for "another target
// creature" clauses and for stack targets that shouldn't be able to
// counter themselves.
func NotSelf(selfID uuid.UUID) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != selfID }
}

// Opponent passes for players other than the caster.
func Opponent() PlayerPredicate {
	return func(_ *game.Game, caster uuid.UUID, p *game.Player) bool { return p.ID != caster }
}

// --- spec constructors ------------------------------------------

// TargetAny — "any target" (CR 115.4): a player, creature,
// planeswalker or battle. Lightning Bolt, Shock, Lightning Helix.
func TargetAny() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "any",
		Label:   "any target",
		Players: true,
		Zones:   []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() || c.IsPlaneswalker() || c.IsBattle()
		},
		Min: 1, Max: 1,
	}
}

// TargetPlayer — "target player" / "target opponent" (pass
// Opponent()).
func TargetPlayer(label string, preds ...PlayerPredicate) *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "player",
		Label:   label,
		Players: true,
		PlayerOK: func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
			for _, pr := range preds {
				if !pr(g, caster, p) {
					return false
				}
			}
			return true
		},
		Min: 1, Max: 1,
	}
}

// TargetCreature — "target creature" narrowed by predicates
// (NonBlack(), OpponentControls(), PowerLE(2), …).
func TargetCreature(label string, preds ...CardPredicate) *game.TargetSpec {
	return battlefieldSpec("creature", label, And(append([]CardPredicate{Creature()}, preds...)...))
}

// TargetPermanent — "target permanent" narrowed by predicates
// ("target artifact or enchantment" is
// TargetPermanent("…", Or(Artifact(), Enchantment()))).
func TargetPermanent(label string, preds ...CardPredicate) *game.TargetSpec {
	return battlefieldSpec("permanent", label, And(preds...))
}

func battlefieldSpec(mode, label string, pred CardPredicate) *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  mode,
		Label: label,
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return pred(g, caster, c)
		},
		Min: 1, Max: 1,
	}
}

// TargetSpell — "target spell" on the stack, narrowed by predicates
// (Noncreature() for Negate, Creature() for Essence Scatter).
func TargetSpell(label string, preds ...CardPredicate) *game.TargetSpec {
	pred := And(preds...)
	return &game.TargetSpec{
		Mode:  "stack_spell",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneStack},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return pred(g, caster, c)
		},
		Min: 1, Max: 1,
	}
}

// TargetCardInGraveyard — "target card in a graveyard", narrowed by
// predicates (YouOwn() for "your graveyard", Creature() for
// "creature card").
func TargetCardInGraveyard(label string, preds ...CardPredicate) *game.TargetSpec {
	pred := And(preds...)
	return &game.TargetSpec{
		Mode:  "card_in_graveyard",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return pred(g, caster, c)
		},
		Min: 1, Max: 1,
	}
}

// --- S27 predicates --------------------------------------------

// OfSubtype matches a permanent whose EFFECTIVE subtypes include
// `subtype`, case-insensitively — "Knights you control" (History of
// Benalia), "Vehicle you control" (Heart of Kiran's crew clause).
//
// Effective, not printed: a Layer-4 type grant is exactly the sort of
// thing a tribal card is supposed to see, and Card.HasSubtype already
// reads the post-layer view on the battlefield and falls back to the
// printed type line everywhere else.
func OfSubtype(subtype string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.HasSubtype(subtype)
	}
}

// GreatestPowerYouControl matches a creature the caster controls
// whose power is not exceeded by any other creature they control
// (Triumph of Gerrard's "target creature you control with the
// greatest power").
//
// Ties all match, which is the rule (CR 700.3 — "the greatest" picks
// out a SET, and the player chooses among it); that is why this is a
// target predicate rather than a lookup that returns one card.
//
// Reads CurrentPower so counters and anthems count, and re-runs at
// resolution like every other target predicate, so a pump in response
// can legally take the target out of the set (CR 608.2b).
func GreatestPowerYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		if !c.IsCreature() || c.Controller != caster {
			return false
		}
		best := c.CurrentPower()
		for _, other := range g.BattlefieldCardsForEffect() {
			if other.Controller != caster || !other.IsCreature() {
				continue
			}
			if other.CurrentPower() > best {
				return false
			}
		}
		return true
	}
}

// --- cost clauses (S28) ------------------------------------------
//
// The three constructors below build specs for COSTS, not targets.
// The difference is load-bearing: a cost does not target (CR 601.2h),
// so hexproof, shroud and "can't be the target of spells" never
// narrow the set, and the engine matches these through
// SpecCandidatesForEffect rather than through the targeting gate.

// CardInYourHand — "a blue card from your hand" (Force of Will), "a
// white card from your hand" (Solitude's evoke cost).
//
// Nothing in the game TARGETS a card in a hand and nothing should: a
// hand is hidden information, and a legal-target list over one would
// leak an opponent's hand size and contents into a picker. The single
// consumer is the alternative-cost payment scan, computed per viewer
// over their own cards.
//
// Ownership is baked in rather than left to the caller. Every printed
// clause of this shape says "your hand", and the failure mode of
// forgetting YouOwn() is a picker offering an opponent's cards —
// both a rules bug and an information leak.
func CardInYourHand(label string, preds ...CardPredicate) *game.TargetSpec {
	pred := And(append([]CardPredicate{YouOwn()}, preds...)...)
	return &game.TargetSpec{
		Label: label,
		Zones: []game.ZoneKind{game.ZoneHand},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return pred(g, caster, c)
		},
		Min: 1, Max: 1,
	}
}

// PermanentYouControl — "an Island you control" (Daze's alternative
// cost). Same cost-not-target contract as CardInYourHand.
func PermanentYouControl(label string, preds ...CardPredicate) *game.TargetSpec {
	pred := And(append([]CardPredicate{YouControl()}, preds...)...)
	return &game.TargetSpec{
		Label: label,
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return pred(g, caster, c)
		},
		Min: 1, Max: 1,
	}
}

// HasSubtype passes when the card has the named subtype — "an Island
// you control" (Daze), "a Swamp" (Snuff Out). Reads EFFECTIVE
// subtypes, so a land something else turned into an Island counts,
// which is what the printed clause means.
func HasSubtype(sub string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hasSubtype(c, sub) }
}
