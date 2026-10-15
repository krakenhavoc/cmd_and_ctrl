package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Portal to Phyrexia — Artifact {9}:
//
//	"When this artifact enters, each opponent sacrifices three creatures
//	 of their choice.
//	 At the beginning of your upkeep, put target creature card from a
//	 graveyard onto the battlefield under your control. It's a Phyrexian
//	 in addition to its other types."
//
// The entry trigger is Ember Swallower's edict with opponents only:
// each opponent names three creatures, one opponent at a time in APNAP
// order, and every named creature is sacrificed together once the last
// has chosen (game.PlayersSacrificeThenForEffect, CR 101.4, #3664). A
// player with fewer than three creatures sacrifices what they have (CR
// 701.21a). Not targeted, so a hexproof creature is as sacrificeable as
// any other.
//
// The upkeep trigger targets a creature card in any graveyard as it
// goes on the stack, so with none it is removed (CR 603.3d). The card
// enters under the Portal's controller's control, whoever owns it.
// "It's a Phyrexian in addition to its other types" has no stated
// duration (CR 611.2a), so it is a layer-4 subtype added for as long as
// that permanent stays on the battlefield (CR 613.1d, CR 400.7), pinned
// to it the way the Enduring cycle's "It's an enchantment" is.
//
// The entry continuation adds the subtype after any copy or entry choice.
func init() {
	Register(Spec{
		OracleID:     "301d1d8e-d3fc-4010-8b29-4724fa0b31cd",
		Name:         "Portal to Phyrexia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Portal to Phyrexia — each opponent sacrifices three creatures", portalToPhyrexiaEdict)),
			Targeting(
				AtYourUpkeep("Portal to Phyrexia — put target creature card from a graveyard onto the battlefield under your control",
					portalToPhyrexiaReanimate),
				targetCreatureInAnyGraveyard(),
			),
		},
	})
}

// portalToPhyrexiaEdict asks each opponent, APNAP from the active
// player, for three creatures.
func portalToPhyrexiaEdict(g *game.Game, item *game.StackItem) error {
	var asks []uuid.UUID
	if n := len(g.Seats); n > 0 {
		start := g.Turn.ActiveSeat
		for i := 0; i < n; i++ {
			p := g.Seats[(start+i)%n]
			if p == nil || p.Eliminated || p.ID == item.Controller {
				continue
			}
			asks = append(asks, p.ID, p.ID, p.ID)
		}
	}
	return g.PlayersSacrificeThenForEffect(item.SourceCardID, asks,
		sacrificeSpec("a creature", Creature()), "Sacrifice a creature", nil)
}

// portalToPhyrexiaReanimate is the upkeep trigger's body.
func portalToPhyrexiaReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
			break
		}
	}
	if target == uuid.Nil {
		return nil
	}
	return g.PutOntoBattlefieldTogetherThenForEffect(
		[]game.BatchEntry{{CardID: target, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: item.Controller},
		func(g *game.Game, entered []uuid.UUID) error {
			if len(entered) == 0 {
				return nil
			}
			return (ScopedEffectFor{Target: entered[0], Mods: []game.Mod{game.AddSubtypesMod("Phyrexian")},
				Duration: g.PinnedTo(game.IndefiniteDuration(), entered[0]),
				Label:    "Portal to Phyrexia — it's a Phyrexian in addition to its other types"}).Apply(NewContext(g, item))
		})
}
