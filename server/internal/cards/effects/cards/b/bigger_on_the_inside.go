package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bigger on the Inside — Enchantment — Aura {3}{R}{G} (Doctor Who, S82 GAP2
// batch 13, #5948, #6101):
//
//	"Enchant artifact or land
//	 Enchanted permanent has "{T}: Target player adds two mana of any one
//	 color. The next spell they cast this turn has cascade.""
//
// A granted activated ability (ADR 0093, Squirrel Nest's shape): the
// enchanted permanent's controller activates it. It targets a player, so it
// is not a mana ability (CR 605.1a) and uses the stack.
//
// The TARGET player adds the mana and picks its colour, as The Warring
// Triad's and Stadium Vendors' do: the pick is queued for that seat and the
// two mana land in that seat's pool. "The next spell they cast this turn
// has cascade" is Dark Apostle's promise with "they" for "you": a delayed
// trigger controlled by the target player on their next cast this turn,
// which cascades from that spell's mana value as the spell is cast, above
// it on the stack (CR 702.85a), and ends with the turn.
//
// No simplification.
const biggerOnTheInsideGrant = "bigger-on-the-inside/mana-and-cascade"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f351961-69e3-4637-b010-d13f19daca39",
		Name:         "Bigger on the Inside",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Or(effects.Artifact(), effects.Land())),
		Grants: []effects.AbilityGrant{{
			Key: biggerOnTheInsideGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target player adds two mana of any one color. The next spell they cast this turn has cascade.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayer("target player"),
				Effect:  biggerOnTheInsideActivated,
			}},
			Text: "{T}: Target player adds two mana of any one color. The next spell they cast this turn has cascade.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(biggerOnTheInsideGrant)},
	})
}

// biggerOnTheInsideActivated gives the target player the mana and the
// cascade promise. A target gone by resolution gets neither (CR 608.2b).
func biggerOnTheInsideActivated(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (effects.AddMana{Player: t.ID, Produced: effects.OneColorOfAmount(2)}).Apply(ctx); err != nil {
			return err
		}
		next := effects.WhenYouNextCast("Bigger on the Inside — the next spell they cast this turn has cascade",
			game.CastFilter{}, effects.GapB194CascadeTheSpellBody)
		next.Controller = t.ID
		return next.Apply(ctx)
	}
	return nil
}
