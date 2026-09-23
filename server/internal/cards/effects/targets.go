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

// NonbasicLand — "nonbasic land" (CR 205.4c): a land without the
// basic SUPERTYPE. The land TYPE is not the test — a Sacred Foundry
// is a Mountain and still nonbasic, and a Snow-Covered Swamp is
// basic. Boseiju, Who Endures and the Wasteland family target with
// it.
func NonbasicLand() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsLand() && !c.HasSupertype("basic")
	}
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

// OfCreatureType passes when the candidate is a creature of the
// named type — "target Goblin creature", "each Elf you control".
// S26's type filter.
//
// Reads through Card.HasSubtype, so a changeling passes for every
// creature type (CR 702.73a) and a Layer-4 type grant counts. The
// IsCreature guard is not redundant: a Kindred (Tribal) card in a
// graveyard prints a creature type on a non-creature line, and
// "target Goblin creature" does not mean it.
func OfCreatureType(subtype string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.HasSubtype(subtype)
	}
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
// Read through game.(*Game).ManaValueForEffect: a {2/W}{2/W}{2/W}
// card is 6 (CR 202.3f), and {X} is zero everywhere except on the
// stack, where it is the value chosen for the spell (CR 202.3e). A
// card whose cost the engine can't read never passes, rather than
// passing as zero.
func ManaValueLE(n int) CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv <= n
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

// WithMorphAbility — "target creature with a morph ability"
// (CR 702.37a), the clause Backslide and Master of the Veil print.
//
// Not HasKeyword("morph"): morph is not a token in canonicalKeywords
// and could not be, because a bare token has nowhere to put the cost
// (morph.go). The card's declaration of the CR 708.4 face-down cast
// IS the ability, so the predicate asks for that.
//
// MEGAMORPH COUNTS AND DISGUISE DOES NOT. CR 702.37b: "a megamorph
// cost is a morph cost", and both declare FaceDownMorphed. Disguise
// is its own keyword with its own rule (CR 702.168), and a card that
// asks for "a morph ability" does not reach it — which is why the
// kind is compared rather than the presence of any face-down cast.
//
// A FACE-DOWN permanent answers false, because CatalogKey answers ""
// for one (CR 708.2a: it has no text to read the declaration off).
// That is also the only useful answer: CR 708.2b says turning a
// face-down permanent face down does nothing at all.
func WithMorphAbility() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		alt := game.FaceDownCastFor(game.CatalogKey(c))
		return alt != nil && alt.FaceDown != nil && alt.FaceDown.Kind == game.FaceDownMorphed
	}
}

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

// OtherThan passes for every card but the one named — the "another"
// in "return another target permanent card from your graveyard to
// your hand" (Eden, Seat of the Sanctum).
//
// It takes an instance ID rather than reading the source off the
// Context, because a target clause is built once and evaluated many
// times, in the enumerator and at resolution, against whichever game
// is in front of it. The ID is a scalar the clause's author captures
// when the clause is made, which is the same rule every other
// continuation in the catalog follows.
func OtherThan(id uuid.UUID) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != id }
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

// AnotherTarget is a triggered ability's "another target …" clause
// that excludes the trigger's OWN source by instance — for
// game.TriggeredAbility.TargetsFrom, which is handed the source, as a
// static Targets clause is not.
//
// `build` receives NotSelf(source) and returns the clause with it
// among its predicates:
//
//	TargetsFrom: AnotherTarget(func(other CardPredicate) *game.TargetSpec {
//		return TargetCreature("another target creature you control", YouControl(), other)
//	}),
//
// Exact where b03NotNamed is an approximation: a second permanent with
// the same name — a Clone or token copy of the source — stays a legal
// target, as printed, and the source itself never is.
func AnotherTarget(build func(other CardPredicate) *game.TargetSpec) func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec {
	return func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		self := uuid.Nil
		if source != nil {
			self = source.InstanceID
		}
		return build(NotSelf(self))
	}
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

// IsTokenPredicate matches a token — "target token you control"
// (Esika's Chariot). Named with the suffix because IsToken is
// already the plain card helper in helpers.go and the two are used
// side by side. Added in S27.
func IsTokenPredicate() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return IsToken(c)
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

// --- S23 mass-effect predicates ----------------------------------
//
// Board wipes are written as exclusions far more often than target
// clauses are — "all creatures except for Krakens", "all creatures
// that aren't Dragons", "all nonland permanents you don't control".
// These are the pieces those clauses are built from. They live here
// rather than in mass.go because a predicate is a predicate: Crux of
// Fate's "non-Dragon" and a hypothetical "target non-Dragon creature"
// must be the same function or they will drift.

// Except is the "all X except for Y" / "all X other than Y" /
// "all X that aren't Y" shape, spelled so the Go reads like the card:
//
//	Except(Creature(), Subtype("Kraken"), Subtype("Leviathan"))
//	  → "all creatures except for Krakens and Leviathans"
//
// Equivalent to And(base, Not(Or(exceptions...))), which is what it
// builds. With no exceptions it is just `base`, so a card that
// computes its exclusion list can pass an empty one without a
// special case.
func Except(base CardPredicate, exceptions ...CardPredicate) CardPredicate {
	if len(exceptions) == 0 {
		return base
	}
	return And(base, Not(Or(exceptions...)))
}

// Subtype passes when the card's effective subtypes include `name`
// ("Dragon", "Kraken", "Equipment", "Aura"). Reads through
// game.Card.HasSubtype, so a Layer-4 subtype grant counts exactly as
// a printed one does — the same reason the type predicates above go
// through HasCardType rather than through the printed type line.
//
// Case-insensitive, matching HasSubtype.
func Subtype(name string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtype(name) }
}

// AnySubtype passes when the card has ANY of the named subtypes —
// Whelming Wave's four-creature-type exclusion list, written once
// instead of Or'd four times at the call site.
func AnySubtype(names ...string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, n := range names {
			if c.HasSubtype(n) {
				return true
			}
		}
		return false
	}
}

// ControlledBy passes for cards a SPECIFIC player controls. Distinct
// from YouControl / OpponentControls, which are relative to the
// caster: River's Rebuke sweeps "all nonland permanents TARGET PLAYER
// controls", and the player in question is whoever the spell targeted,
// not the caster and not "an opponent".
func ControlledBy(player uuid.UUID) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Controller == player }
}

// Multicolored passes when the card has two or more colours — the
// "each player who controls a multicolored creature draws a card"
// half of Depopulate, and the predicate half of any gold-hoser.
func Multicolored() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return len(c.EffectiveColors()) >= 2 }
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

// CardsInYourGraveyard — "N other cards from your graveyard"
// (escape's exile cost, CR 702.138a). The multi-slot sibling of
// CardInYourHand, and the first cost component in the catalog that
// names more than one card.
//
// "Other" is NOT expressed here and deliberately so. The excluded
// card is the spell being cast, which the spec has no way to name —
// the engine excludes it because CR 601.2a has already moved it to
// the stack by the time the cost is paid, and the view drops it from
// the picker for the same reason. A predicate here would have to be
// handed the cast's instance ID, which a TargetSpec's CardOK does
// not receive.
//
// Ownership is baked in for the reason CardInYourHand gives: every
// printed clause of this shape says "your graveyard", and the cost
// of forgetting it is an escape that eats an opponent's yard.
func CardsInYourGraveyard(n int, label string, preds ...CardPredicate) *game.TargetSpec {
	pred := And(append([]CardPredicate{YouOwn()}, preds...)...)
	return &game.TargetSpec{
		Mode:  "card_in_graveyard",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return pred(g, caster, c)
		},
		Min: n, Max: n,
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

// Legendary passes for a permanent with the Legendary supertype —
// Blackblade Reforged's cheaper "equip legendary creature {3}", and
// the predicate half of every commander-flavoured attachment.
//
// Reads game.Card.IsLegendary, which parses the EFFECTIVE type line,
// so a permanent made legendary by a continuous effect counts and a
// token copy printed "except it isn't legendary" does not — the same
// answer the legend rule itself gets.
func Legendary() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLegendary() }
}

// Clauses builds a multi-clause target statement out of the ordinary
// single-clause constructors, in printed order (#764, ADR 0065 §1):
//
//	Targets: Clauses(
//		TargetCreature("target creature you control", YouControl()),
//		Distinct(TargetCreatureOrPlaneswalker(
//			"target creature or planeswalker you don't control", OpponentControls())),
//	),
//
// The result IS the first clause with the rest hung off it, which is
// why a one-clause card needs none of this and reads exactly as it
// did. Nil entries are skipped; a single argument is returned
// unchanged.
func Clauses(first *game.TargetSpec, then ...*game.TargetSpec) *game.TargetSpec {
	if first == nil {
		return nil
	}
	return first.Then(then...)
}

// Distinct marks a clause whose picks must differ from every EARLIER
// clause's picks in the same statement — "a SECOND target permanent
// you control". CR 601.2c lets one object fill two different
// instances of the word "target" unless the card says otherwise, so
// this is opt-in.
func Distinct(spec *game.TargetSpec) *game.TargetSpec {
	if spec != nil {
		spec.Distinct = true
	}
	return spec
}
