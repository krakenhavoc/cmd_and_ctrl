package cardsth

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tidal Terror — Creature — Octopus {4}{U}{U}, 5/6 (March of the Machine):
//
//	"Whenever this creature attacks, you may tap two other untapped creatures
//	 you control. If you do, this creature can't be blocked this turn.
//	 Islandcycling {2}"
//
// Caparocti Sunborn's tap: a choice made as the trigger resolves, a card-set
// pick of exactly two (or none, the "you may" declined) over the OTHER untapped
// creatures its controller has. Each pick is re-read before it is tapped, and
// the evasion applies only if both were really tapped ("if you do"). It is
// pinned to the object that attacked: if Tidal Terror has gone, nothing is
// granted (CR 608.2b). Islandcycling is the shared Typecycling constructor
// (CR 702.29e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7866ce6-d507-4221-bb19-d3e6fd7b57d8",
		Name:         "Tidal Terror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Tidal Terror — tap two other untapped creatures to make it unblockable", tidalTerrorTapTwo),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Islandcycling", "{2}", "an Island card", effects.IsLandWithSubtype("Island")),
		},
	})
}

func tidalTerrorTapTwo(g *game.Game, item *game.StackItem) error {
	candidates := tidalTerrorOthers(g, item.Controller, item.SourceCardID)
	if len(candidates) < 2 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Tidal Terror — tap two other untapped creatures you control so it can't be blocked this turn?",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		ThenKey:  tidalTerrorTap,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var tidalTerrorTap = effects.AskedPickRule("choose-cards/tidal-terror-tap-two",
	func(picked []game.Card) bool { return len(picked) == 2 },
	func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Picked) != 2 {
			return nil
		}
		still := tidalTerrorOthers(ctx.Game, ctx.Controller(), r.Source)
		tapped := 0
		for _, id := range r.Picked {
			if !slices.Contains(still, id) {
				continue
			}
			if err := ctx.Game.TapTargetForEffect(id); err != nil {
				return err
			}
			tapped++
		}
		if tapped != 2 {
			return nil
		}
		return effects.RestrictUntilEOT{Target: r.Source, Restrictions: game.CantBeBlocked,
			Label: "Tidal Terror — can't be blocked this turn"}.Apply(ctx)
	})

// tidalTerrorOthers is every untapped creature `controller` controls other
// than `self`.
func tidalTerrorOthers(g *game.Game, controller, self uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && !c.Tapped && c.IsCreature() && c.InstanceID != self {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
