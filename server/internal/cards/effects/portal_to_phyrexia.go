package effects

import (
	"errors"

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
// one prompted run, each opponent asked three times in APNAP order
// (game.PlayersSacrificeThenForEffect), and a player with fewer than
// three creatures sacrifices what they have (CR 701.21a). Not targeted,
// so a hexproof creature is as sacrificeable as any other. Like every
// "each player sacrifices" in the engine, the choices are made and
// carried out one prompt at a time.
//
// The upkeep trigger targets a creature card in any graveyard as it
// goes on the stack, so with none it is removed (CR 603.3d). The card
// enters under the Portal's controller's control, whoever owns it.
// "It's a Phyrexian in addition to its other types" has no stated
// duration (CR 611.2a), so it is a layer-4 subtype added for as long as
// that permanent stays on the battlefield (CR 613.1d, CR 400.7), pinned
// to it the way the Enduring cycle's "It's an enchantment" is.
//
// One simplification, weaker than printed: a creature whose entry stops
// to ask its controller something (a Clone's copy choice, an "as this
// enters, choose" card) enters without becoming a Phyrexian, because the
// type is added only once the creature is on the battlefield.
func init() {
	Register(Spec{
		OracleID:     "301d1d8e-d3fc-4010-8b29-4724fa0b31cd",
		Name:         "Portal to Phyrexia",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A creature that asks a question as it enters, such as a Clone, doesn't become a Phyrexian when the Portal returns it.",
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Portal to Phyrexia — each opponent sacrifices three creatures", portalToPhyrexiaEdict),
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
	entered, err := g.ReturnFromGraveyardWithCountersForEffect(target, item.Controller, false, nil)
	if errors.Is(err, game.ErrCardNotFound) {
		return nil // CR 608.2b: it left the graveyard in response
	}
	if err != nil || entered == uuid.Nil {
		return err
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod("Phyrexian")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Portal to Phyrexia — it's a Phyrexian in addition to its other types",
	}.Apply(ctx)
}
