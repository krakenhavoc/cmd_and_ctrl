package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnsire, Sunstar Dreadnought — Legendary Artifact — Spacecraft {5},
// 20/20:
//
//	"Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 20+.)
//	 10+ | Whenever you attack, Dawnsire deals 100 damage to up to one
//	 target creature or planeswalker.
//	 20+ | Flying"
//
// Uthros Research Craft's gated trigger with The Seriema's bars: the
// attack trigger does not exist below ten charge counters
// (TriggeredAbility.ActiveWhen), and from twenty the Spacecraft is a
// 20/20 artifact creature with flying. "Whenever you attack" is
// Adeline's one-trigger-per-declaration shape, and "up to one" is a
// target clause with a minimum of zero, so the trigger may be put on
// the stack with no target and then does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afc9436b-8cad-4916-929d-ff33a37b42d5",
		Name:         "Dawnsire, Sunstar Dreadnought",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Triggered:    []game.TriggeredAbility{dawnsireAttackTrigger()},
		Static: append(
			SpacecraftAt(20, 20, 20),
			ThresholdKeywords(20, "flying"),
		),
	})
}

// dawnsireAttackTrigger is the "10+" line.
func dawnsireAttackTrigger() game.TriggeredAbility {
	spec := TargetPermanent("up to one target creature or planeswalker", Or(Creature(), Planeswalker()))
	spec.Min = 0
	t := Targeting(
		OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclaredByYou(ev, source.Controller)
		}, "Dawnsire — 100 damage to up to one target creature or planeswalker",
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, tgt := range ctx.LegalTargets() {
					return DealDamage{Source: item.SourceCardID, Target: tgt.ID, Amount: 100}.Apply(ctx)
				}
				return nil
			})),
		spec)
	t.ActiveWhen = AtChargeCounters(10)
	return t
}
