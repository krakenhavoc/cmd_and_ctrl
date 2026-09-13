package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aura Shards — Enchantment {1}{G}{W}:
//
//	"Whenever a creature you control enters, you may destroy target
//	artifact or enchantment."
//
// One of the most oppressive enchantments in Commander: every
// creature you play is a Naturalize. Registered here because the
// machinery is already complete — this is Reclamation Sage's optional
// targeted trigger, moved from "when THIS enters" to "when any
// creature you control enters".
//
// The watcher deliberately does not exclude tokens: a token creature
// entering triggers it in paper, which is what makes Aura Shards and
// token generators such a punishing pair.
func init() {
	Register(Spec{
		OracleID:     "8d03d050-391c-4311-8c42-4ee632d40fdc",
		Name:         "Aura Shards",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				entered, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entered.IsCreature() && entered.Controller == source.Controller
			},
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return destroyChosenTargetTrigger(source, "Aura Shards — destroy target artifact or enchantment")
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Aura Shards — destroy an artifact or enchantment?",
			},
		}},
	})
}
