package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boseiju, Who Endures — Legendary Land:
//
//	"{T}: Add {G}.
//	 Channel — {1}{G}, Discard this card: Destroy target artifact,
//	 enchantment, or nonbasic land an opponent controls. That player
//	 may search their library for a land card with a basic land type,
//	 put it onto the battlefield, then shuffle. This ability costs
//	 {1} less to activate for each legendary creature you control."
//
// The land that is never a dead draw, and the batch-01 triage filed
// it under "deferred combat keywords" — which channel is not. Channel
// (CR 702.142) is an ACTIVATED ability that functions from the hand,
// and #660 gave `ActivatedAbility.Zones` exactly that dimension: the
// hand zone, the `DiscardSelf` cost component, and one activation
// path that finds its source wherever it lives and treats the card's
// OWNER as "you" (CR 108.4).
//
// The discard is a COST, so it is paid at announce with the ability
// already on the stack (CR 602.2b) and is not refunded if the ability
// is countered — and a cost discard settles without asking, so a
// discard replacement does not get to redirect it.
//
// "That player MAY search" is the opponent's own optional search:
// `SearchLibrary` with Optional set, addressed to the permanent's
// controller as it was read BEFORE the destruction, because a card in
// a graveyard is a bad place to ask who used to control it. A basic
// land TYPE, not the basic supertype — a Sacred Foundry is a legal
// find, which is most of why the compensation is real.
//
// "This ability costs {1} less to activate for each legendary
// creature you control" is the channel ability's OWN cost clause
// (ActivatedAbility.CostModifiers, #1296), priced from the hand where
// channel is activated. It reduces GENERIC mana only (CR 601.2f), so
// with any number of legends out the channel still costs {G}.
func init() {
	Register(Spec{
		OracleID:     "bf1341dd-41a3-49f6-87ec-63170dde4324",
		Name:         "Boseiju, Who Endures",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:         "Channel — {1}{G}, Discard this card: Destroy target artifact, enchantment, or nonbasic land an opponent controls",
			Cost:          game.AbilityCost{Mana: "{1}{G}", DiscardSelf: true},
			Zones:         []game.ZoneKind{game.ZoneHand},
			CostModifiers: []game.CostModifier{ChannelDiscountPerLegendaryCreature()},
			Targets: TargetPermanent("target artifact, enchantment, or nonbasic land an opponent controls",
				Or(Artifact(), Enchantment(), NonbasicLand()), OpponentControls()),
			Effect: boseijuChannel,
		}},
	})
}

// boseijuChannel destroys the target and then offers its controller
// the basic-land-type search the card compensates them with.
//
// The search rides the destruction's CONTINUATION, not the next line:
// a commander among the targets pauses on the CR 903.9 window, and
// the search has to come after that settles. It is NOT gated on the
// `destroyed` list — "that player may search" is an unconditional
// clause about a PLAYER, so an indestructible artifact survives and
// its controller still gets the land, which is the printed ruling.
func boseijuChannel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var victim, owner uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		victim = t.ID
		if c, ok := g.LookupCardForEffect(victim); ok {
			owner = c.Controller
		}
		break
	}
	if victim == uuid.Nil || owner == uuid.Nil {
		return nil
	}
	source := item.SourceCardID
	return g.DestroyPermanentsThenForEffect([]uuid.UUID{victim},
		func(g *game.Game, _ []uuid.UUID) error {
			return SearchLibrary{
				Player:    owner,
				Source:    source,
				Predicate: hasABasicLandType,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Optional:  true,
				Shuffle:   true,
				Reason:    "Boseiju, Who Endures — you may search for a land card with a basic land type",
			}.Apply(NewContext(g, item))
		})
}

// hasABasicLandType is "a land card with a basic land type" — the
// land TYPE, not the basic supertype, so a shockland or a Triome
// qualifies and a Wastes does not.
func hasABasicLandType(c game.Card) bool {
	if !c.IsLand() {
		return false
	}
	for _, t := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		if c.HasSubtype(t) {
			return true
		}
	}
	return false
}
