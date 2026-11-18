package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Workshop Elders — Creature — Human Artificer {6}{U}, 4/4:
//
//	"Artifact creatures you control have flying.
//	 At the beginning of combat on your turn, you may have target
//	 noncreature artifact you control become a 0/0 artifact creature. If
//	 you do, put four +1/+1 counters on it."
//
// The flying grant is a layer 6 keyword grant to the artifact creatures its
// controller has, the Elders included. The trigger is an optional targeted
// one (CR 603.5): as it resolves the controller says yes, the artifact
// becomes an artifact creature with base power and toughness 0/0 with no
// duration (CR 611.2a), pinned to that object (CR 400.7), and then gets the
// four counters. A target that is no longer a legal noncreature artifact
// does nothing (CR 608.2b), so no counters land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e9164be-3534-4445-bfad-00d16a50575c",
		Name:         "Workshop Elders",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				AtBeginningOfYourCombat("Workshop Elders — target noncreature artifact you control becomes a 0/0 artifact creature, then put four +1/+1 counters on it",
					workshopEldersAnimate),
				TargetPermanent("target noncreature artifact you control", And(Artifact(), Not(Creature()), YouControl()))),
				"Workshop Elders — have target noncreature artifact you control become a 0/0 artifact creature?"),
		},
	})
}

// workshopEldersAnimate makes the first still-legal target a 0/0 artifact
// creature, then puts four +1/+1 counters on it.
func workshopEldersAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || !onBattlefield(g, t.ID) {
			continue
		}
		mods := append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(0, 0)...)
		g.RegisterScopedEffectForEffect(ctx.Source(),
			[]game.AffectedObject{game.PinObject(t.ID, c.EnteredBattlefieldAt)}, mods,
			g.PinnedTo(game.IndefiniteDuration(), t.ID),
			"Workshop Elders — becomes a 0/0 artifact creature")
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 4}.Apply(ctx)
	}
	return nil
}
