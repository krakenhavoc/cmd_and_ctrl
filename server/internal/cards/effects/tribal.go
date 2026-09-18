package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tribal.go — the shared builders for S26's creature-type cards
// (#78). Four shapes cover the whole sprint:
//
//	ChooseCreatureTypeAsEnters   "as this enters, choose a creature type"
//	TribeFilter               "other Goblin creatures you control"
//	TribalAnthem / TribalKeywordGrant   the lord statics
//	GrantAllCreatureTypesUntilEOT       "gains all creature types"
//
// They live here rather than in each card file because a tribal lord
// is the single most copy-pasted shape in the format: five of them
// ship in this sprint and the next twenty are the same two statics
// with a different noun. A builder means the self-exclusion bug
// ("other") and the controller-scope bug ("you control") each have
// one place to be wrong, and one place to be fixed.
//
// LANDWALK IS A KEYWORD GRANT LIKE ANY OTHER since #705. Until then
// the block check could not see the defending player's lands, so
// Goblin King and Elvish Champion shipped without their landwalk and
// Lord of Atlantis's islandwalk was inert. The landwalk tokens
// ("islandwalk", "forestwalk", "nonbasic landwalk", …) are canonical
// now and read by game/landwalk.go, so TribalKeywordGrant grants them
// with the same call it uses for haste.

// ChooseCreatureTypeAsEnters builds the `Spec.AsEnters` for a permanent
// whose text opens "As this permanent enters, choose a creature
// type" (CR 614.12) — Cavern of Souls, Door of Destinies,
// Vanquisher's Banner, Adaptive Automaton.
//
// `label` is the prompt header, normally the card's name.
//
// The prompt is queued rather than resolved: nothing happens until
// the controller answers, and until then the permanent's NamedTribe
// is empty and every static that reads it applies to nothing. See
// game/creature_type_choice.go for why this is an ETB hook and not a
// CR 614 replacement, and for what that costs.
func ChooseCreatureTypeAsEnters(label string) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueCreatureTypeChoiceForEffect(card.Controller, card.InstanceID, label)
		return nil
	}
}

// TribeFilter is the "which creatures does this lord pump" clause,
// spelled out field by field so a card file reads like its oracle
// text instead of like a predicate.
//
//	Lord of Atlantis   "Other Merfolk get +1/+1"
//	                   {Tribes: ["Merfolk"], Others: true}
//	Goblin Chieftain   "Other Goblin creatures you control get +1/+1"
//	                   {Tribes: ["Goblin"], Others: true, YoursOnly: true}
//	Death Baron        "Skeletons you control and other Zombies you
//	                    control get +1/+1"
//	                   {Tribes: ["Skeleton", "Zombie"], Others: true,
//	                    YoursOnly: true}
//	Vanquisher's Banner "Creatures you control of the chosen type"
//	                   {Chosen: true, YoursOnly: true}
//
// Two of those deserve a note.
//
// Death Baron's "other" qualifies only the Zombies, and this filter
// applies it to both halves. The two agree because Death Baron is a
// Zombie Wizard and not a Skeleton, so it is excluded either way. A
// future lord that IS one of the types it pumps unconditionally would
// need the split; there isn't one.
//
// Tribes is ANY-of, not all-of, and the bonus is granted once
// regardless of how many entries match. A changeling matches every
// entry and still gets +1/+1, which is right: the lord grants a
// bonus to a creature, not to a type.
type TribeFilter struct {
	// Tribes are the creature types this clause names. Empty with
	// Chosen unset matches every creature — which is Coat of Arms'
	// scope, not a lord's, so a lord always names something.
	Tribes []string

	// Chosen reads the tribe off the SOURCE permanent's NamedTribe
	// instead of a printed list (CR 614.12). A source with no type
	// named yet matches nothing.
	Chosen bool

	// Others excludes the source permanent itself — the "other" in
	// "other Merfolk".
	Others bool

	// YoursOnly restricts to the source's controller — the "you
	// control" in "other Goblin creatures you control". Lord of
	// Atlantis, Goblin King and Elvish Champion do NOT say it and
	// therefore pump the whole table's creatures of the type.
	YoursOnly bool
}

// Matches is the `StaticAbility.AppliesTo` body for this filter.
//
// Reads `target.IsCreature()` and `target.HasSubtype(...)`, both of
// which go through the post-layer effective view, so a lord sees a
// creature a Layer-4 effect animated and a changeling counts as its
// type. That is also why a lord must never read the printed type
// line: Maskwood Nexus makes every creature you control every type,
// and the printed line says nothing about it.
func (f TribeFilter) Matches(target *game.Card, _ *game.Game, source *game.Card) bool {
	if !target.IsCreature() {
		return false
	}
	if f.Others && target.InstanceID == source.InstanceID {
		return false
	}
	if f.YoursOnly && target.Controller != source.Controller {
		return false
	}
	if f.Chosen {
		if source.NamedTribe == "" {
			return false
		}
		return target.HasSubtype(source.NamedTribe)
	}
	if len(f.Tribes) == 0 {
		return true
	}
	for _, t := range f.Tribes {
		if target.HasSubtype(t) {
			return true
		}
	}
	return false
}

// TribalAnthem is a lord's "+N/+N" half: Layer 7c, applied to every
// creature the filter matches.
func TribalAnthem(f TribeFilter, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: f.Matches,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// TribalScalingAnthem is the "+1/+1 for each X" half — Door of
// Destinies' charge counters, and the shape any "for each" lord
// wants. `per` is evaluated per recompute against the SOURCE, so it
// can read counters, a board count, or anything else that changes.
//
// Returning 0 from `per` is not a special case: the Apply adds zero
// and the card is inert until the count moves, which is exactly what
// a fresh Door of Destinies does.
func TribalScalingAnthem(f TribeFilter, per func(source *game.Card, g *game.Game) int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: f.Matches,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := per(source, g)
			c.Power += n
			c.Toughness += n
		},
	}
}

// TribalKeywordGrant is a lord's "and have <keyword>" half: Layer 6,
// idempotent, applied to every creature the filter matches.
//
// `keyword` must be one of the engine's canonical lowercase tokens
// (game/keywords.go). Granting anything else appends a string nothing
// reads.
func TribalKeywordGrant(f TribeFilter, keyword string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: f.Matches,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, k := range c.Abilities {
				if k == keyword {
					return
				}
			}
			c.Abilities = append(c.Abilities, keyword)
		},
	}
}

// AllCreatureTypesGrant is "these creatures are every creature type"
// as a battlefield static — Maskwood Nexus' first sentence.
//
// LAYER 4, not layer 6, even though what it writes is an ability
// string. The layer is the semantic claim and the storage is an
// implementation detail: this is a TYPE-changing effect (CR 613.1d),
// and "is every creature type" is carried as the changeling keyword
// only so that one map lookup answers what ~345 subtypes otherwise
// would (see HasAllCreatureTypes).
//
// Getting this wrong is observable and was, before there was a test
// for it. Declared in layer 6, the grant would be ordered against
// every LORD's keyword half by CR 613.7 timestamp, so a Goblin
// Chieftain that entered before the Nexus would grant haste to
// creatures that were not yet Goblins — while its +1/+1 (layer 7c,
// after all of layer 6) landed correctly. Half a working card, which
// is the hardest kind of bug to see.
func AllCreatureTypesGrant(f TribeFilter) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: f.Matches,
		Apply:     applyAllCreatureTypes,
	}
}

// applyAllCreatureTypes appends the changeling marker, idempotently.
// Shared by the static and the until-end-of-turn forms so the two can
// never disagree about what the marker is.
func applyAllCreatureTypes(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, k := range c.Abilities {
		if k == game.KeywordChangeling {
			return
		}
	}
	c.Abilities = append(c.Abilities, game.KeywordChangeling)
}

// GrantAllCreatureTypesUntilEOT is "gains all creature types until
// end of turn" (Shields of Velis Vel) — AllCreatureTypesGrant with a
// CR 514.2 duration instead of a source permanent.
//
// It is NOT GrantKeywordUntilEOT with a changeling argument, for the
// layer reason above: that primitive is hard-wired to layer 6, which
// is right for "gains trample" and wrong for a type change.
//
// The affected set is snapshotted at resolution (CR 611.2c), like
// every other until-end-of-turn primitive: a creature the targeted
// player casts afterwards is not affected.
type GrantAllCreatureTypesUntilEOT struct {
	// Match selects the affected permanents, evaluated ONCE.
	Match CardPredicate

	// Target pins the effect to one permanent. Ignored when Match is
	// set.
	Target uuid.UUID

	Label string
}

func (a GrantAllCreatureTypesUntilEOT) Apply(ctx *Context) error {
	set := eotSnapshot(ctx, a.Target, a.Match)
	if set == nil {
		return nil
	}
	ctx.Game.RegisterScopedStaticForEffect(game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: set.appliesTo(),
		Apply:     applyAllCreatureTypes,
	}, ctx.Source(), eotLabel(a.Label, "all creature types until end of turn"),
		ctx.Game.UntilEndOfTurnDuration())
	return nil
}

// ChosenTypeManaRestrictions builds the `RestrictionsFunc` for
// Cavern of Souls: "spend this mana only to cast a creature spell of
// the chosen type".
//
// The empty-tribe case is the important one. A Cavern whose
// controller has not answered the prompt yet returns a subtype tag
// naming nothing, and the matcher refuses an empty tag — so the mana
// is UNSPENDABLE rather than unrestricted. Returning nil here would
// hand the player free mana of any colour, which is the one direction
// a restriction must never fail in (#259).
func ChosenTypeManaRestrictions() func(*game.Game, uuid.UUID, uuid.UUID) []string {
	return func(g *game.Game, _ uuid.UUID, source uuid.UUID) []string {
		return []string{
			game.ManaRestrictCast,
			game.ManaRestrictType("Creature"),
			game.ManaRestrictSubtype(g.NamedTribeOf(source)),
		}
	}
}

// MatchCreatureSubtype is the creature-type filter for a scaled mana
// ability — Elvish Archdruid's "{T}: Add {G} for each Elf you
// control". Sibling of MatchLandSubtype in mana_derivation.go, and
// the reason it is separate: a subtype match that did not check
// IsCreature would count a Kindred artifact naming Elf.
func MatchCreatureSubtype(subtype string) func(game.Card) bool {
	return func(c game.Card) bool {
		return c.IsCreature() && c.HasSubtype(subtype)
	}
}
