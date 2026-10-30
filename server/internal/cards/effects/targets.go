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

// FaceUp is "face-up" for a card in exile (Pull from Eternity, Riftsweeper):
// not exiled face down (foretell, hideaway, manifest-style exile). The
// engine's exile walk already refuses a face-down card its chooser has not
// seen, so this is the printed word for the cards the chooser HAS seen --
// a face-down card is not a "face-up exiled card" however well it is known.
func FaceUp() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.FaceDown }
}

// --- type predicates (post-layer, so type-changers compose) ----

func Creature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() }
}

// AttackingOrBlocking — "target attacking or blocking creature"
// (Helicarrier Strike, #1703). A creature is attacking when
// Card.AttackingTarget is set and blocking when Card.BlockingTarget
// is (attack_target.go / blockers.go); both are cleared at combat's
// end, so the clause is legal only during the combat that set them.
func AttackingOrBlocking() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil
	}
}

// BlockedCreature — "blocked creature" (CR 509.1h): an attacking
// creature that became blocked as blockers were declared, and stays
// blocked even if every creature blocking it has left combat ("Works
// even if the creature is 'blocked' by an effect rather than actual
// creatures", Benalish Missionary's ruling). Read through
// game.BlockedAttackerForEffect (#2026).
func BlockedCreature() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && g.BlockedAttackerForEffect(c.InstanceID)
	}
}

// BlockingCreature — "blocking creature" (CR 506.4): a creature declared
// as a blocker, from the declaration until combat is cleared as the end
// of combat step ends (CR 511.3), so outside that window nothing
// matches. Glyph of Destruction's "target blocking Wall you control".
func BlockingCreature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.BlockingTarget != uuid.Nil
	}
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

// Battle is "battle" (CR 310): Traumatic Revelation's "a creature or
// battle card". Reads the post-layer type like the rest.
func Battle() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsBattle() }
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
//
// The type is a WORD in the clause's text, so it is read through the
// judging object's changed words (HasSubtypeWord, CR 612.2): an
// Artificial Evolution on the source turns its "Goblin" into "Elf".
func OfCreatureType(subtype string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.HasSubtypeWord(subtype)
	}
}

// Nonland passes for anything that isn't a land.
func Nonland() CardPredicate { return Not(Land()) }

// Noncreature passes for anything that isn't a creature (Negate's
// "noncreature spell").
func Noncreature() CardPredicate { return Not(Creature()) }

// --- colour predicates -------------------------------------------

// OfColor passes when the card is the given colour letter.
//
// The letter is a colour WORD in the clause's text, so it is read
// through the judging object's changed words (HasColorWord, CR 612.2;
// ADR 0160's 2026-10-15 amendment): Doom Blade's "nonblack" under a
// Sleight of Mind that made it "nonblue" passes a black creature and
// refuses a blue one. With no lens lent the letter stands.
func OfColor(color string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasColorWord(color) }
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
//
// A creature SPELL is read with its own characteristic-defining ability
// applied (CR 604.3, CR 113.6a: it works on the stack), so a `*/*`
// creature spell such as Tarmogoyf counts what its text says
// (Spellbreaker Behemoth). Anything else reads its current power.
func PowerGE(n int) CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if g != nil && g.Stack != nil && g.Stack.Contains(c.InstanceID) {
			return g.PowerAnywhereForEffect(c) >= n
		}
		return c.CurrentPower() >= n
	}
}

// ToughnessGE passes when the creature's current toughness is ≥ n —
// Murdock's Crusade's "target creature with toughness 4 or greater"
// (#1703).
func ToughnessGE(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.CurrentToughness() >= n }
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

// FaceDown is the "face-down permanent" clause (CR 708.2): a battlefield
// permanent that is currently face down. Hauntwoods Shrieker's target
// (#2590).
//
// A card face down in EXILE (a foretold card, a Necropotence exile) is
// not a permanent and does not pass. The predicate reads the object's
// STATE and never its identity, so it tells the targeting player
// nothing about the card underneath.
func FaceDown() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.FaceDownIsPermanent()
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
// A TRIGGER'S "another" is also expressible with Another (a clause's
// ExcludeSource); this form remains for a clause built per source.
// Either is exact: a second permanent with the same name — a Clone or
// token copy of the source — stays a legal target, as printed, and
// the source itself never is.
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

// TargetSpellOrPermanent — "target spell or permanent": ONE pick that
// may be a spell on the stack or a permanent on the battlefield
// (Venser, Shaper Savant; Sink into Stupor). `spell` narrows the stack
// half and `permanent` the battlefield half; nil admits every spell /
// every permanent.
//
// Mode is "any" because the client's Mode string decides which
// SURFACES enter targeting and this clause needs both the stack and
// the battlefield (Aang, Swift Savior's shape); legality still comes
// from CardOK. Players stays false, so "any" never points at a player.
// Only a spell is a card on the stack — an ability is a StackMeta item
// with no card — so an ability is never offered.
func TargetSpellOrPermanent(label string, spell, permanent CardPredicate) *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  "any",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneStack, game.ZoneBattlefield},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, zone game.ZoneKind) bool {
			pred := permanent
			if zone == game.ZoneStack {
				pred = spell
			}
			return pred == nil || pred(g, caster, c)
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

// TargetCardInExile — "target card in exile", narrowed by predicates
// (Creature() for "creature card", YouOwn() for "a card you own in
// exile"). The exile sibling of TargetCardInGraveyard: the picker is
// the zone browser's exile tab, and the clause is re-judged at
// resolution (CR 608.2b), so a card that left exile in response is an
// illegal target. A face-down card the caster has not seen is never
// offered (the engine's walk drops it), so the legal set cannot leak
// what a hidden card is.
//
// "Exiled with this" is not a predicate here -- a CardPredicate is handed
// the caster, not the source. Wrap the clause in ExiledWithThis (#3259).
func TargetCardInExile(label string, preds ...CardPredicate) *game.TargetSpec {
	pred := And(preds...)
	return &game.TargetSpec{
		Mode:  "card_in_exile",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneExile},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return pred(g, caster, c)
		},
		Min: 1, Max: 1,
	}
}

// AmongTheCardsNamed narrows a reflexive trigger's clause to the cards
// its parent resolution named (CR 603.12): "choose up to one target
// creature card exiled this way" (Espers to Magicite, #3846). The set is
// the ReflexiveTrigger's Cards payload, stamped onto the clause by the
// engine when the trigger is created and again when a restored item
// re-derives it, so declare it on the body's ReflexiveBody registration
// like any reflexive clause. Pair it with the zone the cards are in:
//
//	AmongTheCardsNamed(TargetCardInExile("up to one target creature card exiled this way", Creature()).WithCount(0, 1))
func AmongTheCardsNamed(spec *game.TargetSpec) *game.TargetSpec {
	spec.AmongPayload = true
	return spec
}

// ExiledWithThis narrows a clause to the cards exiled with its own
// source (CR 607.2a): "target creature card exiled with The Darkness
// Crystal". It is data on the clause (game.TargetSpec.ExiledWithSource),
// judged in the one shared walk, so the legal set, the announce check
// (CR 601.2c), the resolution re-check (CR 608.2b), the client's picker
// and the bot's moves are one answer. The link is the OBJECT, so a
// second Crystal's exiles are not offered, and a Crystal that left and
// came back has no claim on what its earlier self exiled (CR 400.7).
// Wrap the whole clause, like RelativeToSource:
//
//	Targets: ExiledWithThis(TargetCardInExile("target creature card exiled with this", Creature()))
//
// Only a card whose exile stamped Card.ExiledWith carries the link:
// today that is a replacement exile (ReplacementEvent.ExiledWith, ADR
// 0066's 2026-10-07 amendment). A card exiled some other way -- by a
// resolving ability, until ADR 0155's stamp lands -- is never "exiled
// with" anything here, so the clause admits nothing for it.
func ExiledWithThis(spec *game.TargetSpec) *game.TargetSpec {
	spec.ExiledWithSource = true
	return spec
}

// ChosenAtRandom is "target … chosen at random" on a TRIGGERED
// ability's clause: Witch Hunt's "target opponent chosen at random",
// Scab-Clan Giant's "target creature an opponent controls chosen at
// random" (#1815). The engine draws the target uniformly from the
// clause's legal set as the trigger is put on the stack (CR 603.3d),
// through the keyed random stream, and nobody is asked; with no legal
// target the trigger is removed. It is game.TargetSpec.AtRandom.
//
// Since #4051 it is also an ACTIVATED ability's lone clause: Goblin
// Test Pilot's "any target chosen at random" is drawn as part of
// activating (CR 602.2b / 601.2c), the activator names nothing for it,
// and with no legal target the ability can't be activated.
//
// Register refuses it on a spell, on a mode, on an activated ability
// with any other clause and on a clause that picks anything but
// exactly one target: no printed card draws more than one.
func ChosenAtRandom(spec *game.TargetSpec) *game.TargetSpec {
	spec.AtRandom = true
	return spec
}

// ControlledByThatPlayer narrows a TRIGGERED ability's clause to what
// "that player" / "that opponent" controls, where the triggering event
// names the player: the opponent the Dragons just dealt combat damage
// to (Parapet Thrasher's "Destroy target artifact that opponent
// controls", #3860). It is data on the clause
// (game.TargetSpec.ControlledByThatPlayer), judged in the one shared
// walk against the event the trigger carries, so the mode prompt (a
// bullet with no such artifact is not on offer, CR 603.3c), the legal
// set, the announce check (CR 601.2c), the resolution re-check (CR
// 608.2b), the client's picker and the bot's moves are one answer.
// Wrap the whole clause, like ExiledWithThis:
//
//	TargetPermanent("target artifact that opponent controls", Artifact())
//
// becomes ControlledByThatPlayer(TargetPermanent(…)). The player is the
// event's Target, so use it on a trigger whose event has one (combat
// damage to a player, a player being targeted, a spell cast at one);
// on an event that names none the clause admits nothing, and on a spell
// or activated ability it admits nothing too.
func ControlledByThatPlayer(spec *game.TargetSpec) *game.TargetSpec {
	spec.ControlledByThatPlayer = true
	return spec
}

// ControlledByThatEventActor narrows a TRIGGERED ability's clause to what
// the player who caused the triggering event controls (Black Bolt, Inhuman
// King's "destroy target nonland permanent that player controls" after
// "becomes the target of a spell or ability an opponent controls", #4707).
// ControlledByThatPlayer's twin: that one reads the event's Target (the
// player hit), this reads its Actor (on EventBecomesTarget, the controller
// of the targeting spell or ability). Data on the clause
// (game.TargetSpec.ControlledByEventActor), judged in the one shared walk;
// on a spell, an activation or an event with no Actor it admits nothing.
//
//	ControlledByThatEventActor(TargetPermanent("target nonland permanent that player controls", Nonland()))
func ControlledByThatEventActor(spec *game.TargetSpec) *game.TargetSpec {
	spec.ControlledByEventActor = true
	return spec
}

// OfAnOpponentsChoice is "target … of an opponent's choice" (#2055, ADR
// 0019's 2026-10-16 amendment): as the spell or ability is put on the
// stack, its controller chooses an opponent and THAT opponent chooses
// the clause's targets (CR 601.2c; the Nova Pentacle, Evangelize and
// Volcanic Offering rulings). The controller still targets, so the
// clause's predicates are written from the controller's side ("you
// don't control" is the controller's), and hexproof is judged for the
// controller. It is data on the clause (game.TargetSpec.
// ChosenByOpponent): the controller's payload leaves it out, the engine
// asks the controller which opponent and that opponent for the picks,
// and the client's walk and the bot's enumerator skip it. Wrap the
// whole clause:
//
//	OfAnOpponentsChoice(TargetCreature("target creature of an opponent's choice"))
//
// With "they control", wrap that in TheyControl. Not on a mode's clause,
// and never divided: effects.Register refuses both.
func OfAnOpponentsChoice(spec *game.TargetSpec) *game.TargetSpec {
	spec.ChosenByOpponent = true
	return spec
}

// OfThatPlayersChoice is "target … that player controls of their
// choice" on a triggered ability (#1997, ADR 0019 amendment 2026-10-23):
// the player the triggering event names (the player whose upkeep it is)
// chooses the clause's targets, though the ability's controller still
// controls it and still targets (the 2004 ruling on The Abyss). Legality
// is judged for the controller, so the clause's predicates are written
// from the controller's side. It is data on the clause
// (game.TargetSpec.ChosenByTriggerActor): the trigger's target walk sends
// the step's prompt to that player. "That player controls" is TheyControl
// around it. Triggers only; effects.Register refuses it elsewhere.
//
//	TheyControl(OfThatPlayersChoice(TargetCreature("target nonartifact creature that player controls of their choice", NonArtifact())))
func OfThatPlayersChoice(spec *game.TargetSpec) *game.TargetSpec {
	spec.ChosenByTriggerActor = true
	return spec
}

// Harmful marks a clause whose effect hurts the permanent it picks
// (game.TargetSpec.Harmful, #4249): The Abyss's destroy. Nothing else on a
// pick_target prompt says which way the pick cuts, so a bot that has to
// pick one of its own permanents (the Abyss on its controller's own
// upkeep) gives up the least valuable instead of an arbitrary one. A
// clause that picks a permanent only; Register refuses it elsewhere.
//
//	Harmful(TheyControl(OfThatPlayersChoice(TargetCreature("target nonartifact creature that player controls of their choice", Not(Artifact())))))
func Harmful(spec *game.TargetSpec) *game.TargetSpec {
	spec.Harmful = true
	return spec
}

// TheyControl is the "they control" of "target creature of an opponent's
// choice they control" (Evangelize, Preacher, Arena): the candidate must
// be controlled by the opponent who chooses it (game.TargetSpec.
// ControlledByChooser), read at announce, at the CR 608.2b re-check and
// at a retarget against the opponent the announcement named. Only on an
// OfAnOpponentsChoice clause; effects.Register refuses it alone.
//
//	TheyControl(OfAnOpponentsChoice(TargetCreature("target creature of an opponent's choice they control")))
func TheyControl(spec *game.TargetSpec) *game.TargetSpec {
	spec.ControlledByChooser = true
	return spec
}

// SaddledThis narrows a clause to the creatures that saddled its own
// source this turn (CR 702.171c): Giant Beaver's "target creature that
// saddled it this turn". It is data on the clause
// (game.TargetSpec.PaidForSource), judged in the one shared walk, so the
// legal set, the announce check (CR 601.2c), the resolution re-check (CR
// 608.2b), the client's picker and the bot's moves are one answer. The
// payment is read by the source's OBJECT, so a creature that saddled a
// different Mount is not offered, a saddler that left the battlefield (or
// left and came back, CR 400.7) is not, and a Mount that left still has
// its saddlers. Wrap the whole clause, like ExiledWithThis:
//
//	Targets: SaddledThis(TargetCreature("target creature that saddled it this turn"))
func SaddledThis(spec *game.TargetSpec) *game.TargetSpec {
	spec.PaidForSource = game.PaidSaddled
	return spec
}

// CrewedThis is SaddledThis for a Vehicle: "target creature that crewed
// it this turn" (CR 702.122c).
func CrewedThis(spec *game.TargetSpec) *game.TargetSpec {
	spec.PaidForSource = game.PaidCrewed
	return spec
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
//
// A basic land type or a creature type is a word of the clause's text
// and reads through the judging object's changed words (CR 612.2).
func Subtype(name string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtypeWord(name) }
}

// AnySubtype passes when the card has ANY of the named subtypes —
// Whelming Wave's four-creature-type exclusion list, written once
// instead of Or'd four times at the call site.
func AnySubtype(names ...string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, n := range names {
			if c.HasSubtypeWord(n) {
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

// OwnedBy passes for cards the named player owns — "target creature
// card in that player's graveyard" (Sepulchral Primordial). ControlledBy
// is the battlefield sibling.
func OwnedBy(player uuid.UUID) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Owner == player }
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
		return c.HasSubtypeWord(subtype)
	}
}

// GreatestPowerYouControl matches a creature the caster controls
// whose power is not exceeded by any other creature they control
// (Triumph of Gerrard's "target creature you control with the
// greatest power").
//
// Ties all match — "the greatest" picks
// out a SET, and the player chooses among it — which is why this is a
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
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtypeWord(sub) }
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

// Snow passes for a permanent with the snow supertype (CR 205.4a), read
// off its effective type line: Icebreaker Kraken's "snow lands".
func Snow() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSupertype("snow") }
}

// YourCommander passes for a card that is the caster's own commander
// (Card.IsCommander, CR 903.4) currently under the caster's control —
// Commander's Plate's cheaper "equip commander" clause. Ownership
// alone is not enough: what this equip clause narrows is who the
// Equipment may move onto right now, which is a question about
// control (CR 702.6a's "creature you control"), same as every other
// equip target.
func YourCommander() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
		return c.IsCommander && c.Controller == caster
	}
}

// ACommander passes for any card that is a commander (Card.IsCommander,
// CR 903.4), whoever controls it: Timely Ward's "if it targets a
// commander". YourCommander is the narrower, controller-bound one.
func ACommander() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCommander }
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

// ToughnessLE passes when the creature's current toughness is ≤ n —
// Endangered Armodon's "a creature with toughness 2 or less" (ADR 0107
// §1). The sibling of ToughnessGE, read the same way.
func ToughnessLE(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.CurrentToughness() <= n }
}
