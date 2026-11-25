package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yannik, Scavenging Sentinel — Legendary Creature — Hyena Beast {2}{G}{W},
// 3/3 (Commander 2020, S82 OLD, #5513):
//
//	"Partner with Nikara, Lair Scavenger
//	 Vigilance
//	 When Yannik enters, exile another creature you control until Yannik
//	 leaves the battlefield. When you do, distribute X +1/+1 counters among
//	 any number of target creatures, where X is the exiled creature's power."
//
// The creature is chosen as the trigger resolves, not targeted, so a hexproof
// creature may be picked; with no other creature the ability does nothing. The
// exile is Banisher Priest's "until this leaves" record (CR 610.3): if Yannik
// has already left, nothing is exiled and the reflexive trigger never
// happens. X is the exiled creature's power as it last existed on the
// battlefield (CR 608.2h), read before the exile. "When you do" is a CR 603.12
// reflexive trigger whose targets (any number of creatures, at most X of
// them, each at least one counter) and split are announced as it goes on the
// stack (Numa, Joraga Chieftain's shape); an X of 0 or less distributes
// nothing. "Partner with" is the CR 702.124j entry search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "81ab8947-8046-4944-ae58-006a6eefa08a",
		Name:            "Yannik, Scavenging Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Yannik, Scavenging Sentinel", "Nikara, Lair Scavenger"),
			effects.WhenThisEnters("Yannik, Scavenging Sentinel — exile another creature you control until Yannik leaves the battlefield; when you do, distribute X +1/+1 counters among any number of target creatures",
				yannikChoose),
		},
	})
}

func yannikChoose(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := item.SourceCardID
	// Yannik's own object, taken now: the pick answers later, off the stack.
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return effects.ChoosePermanents{
		Question: "Yannik, Scavenging Sentinel — exile another creature you control until Yannik leaves the battlefield",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := effects.PermanentsControlledByMatching(g, of,
				effects.And(effects.Creature(), effects.NotSelf(src)))
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: yannikChosen,
		Carry: []uuid.UUID{ref.ID},
		Nums:  []int{ref.Epoch},
	}.Apply(ctx)
}

// yannikChosen exiles the chosen creature until Yannik leaves, remembering its
// power for the reflexive trigger. Keyed (#4290).
var yannikChosen = effects.RegisterPermanentsPick("prompt-run/yannik-scavenging-sentinel-chosen", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
		cards := picked.Cards()
		if len(cards) == 0 || len(r.Carry) == 0 || len(r.Nums) == 0 {
			return nil
		}
		id := cards[0]
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !effects.OnBattlefield(ctx.Game, id) {
			return nil
		}
		tail := ctx.ItemTail(yannikExiled)
		tail.N = c.PowerForComparison()
		return effects.ExileUntil{
			Target:      id,
			UntilLeaves: game.ObjectRef{ID: r.Carry[0], Epoch: r.Nums[0]},
			Label:       "Yannik, Scavenging Sentinel — the exiled creature returns when Yannik leaves the battlefield",
			ThenTail:    tail,
		}.Apply(ctx)
	},
})

// yannikExiled is "when you do": the counters, X being the exiled creature's
// power. An on-disk key.
var yannikExiled = effects.CardTail("yannik-scavenging-sentinel/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Yannik, Scavenging Sentinel — distribute X +1/+1 counters among any number of target creatures",
		Body:   yannikDistributeBody,
		Params: game.EffectParams{Amount: o.Tail.N},
	}.Apply(ctx)
})

// yannikDistributeBody is the reflexive trigger: any number of target
// creatures, at most X of them, with X divided among them. An on-disk key.
var yannikDistributeBody = game.ReflexiveBody("yannik-scavenging-sentinel/distribute",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		return effects.PutDividedCounters(effects.NewContext(g, item), game.CounterPlusOne)
	}),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("any number of target creatures").WithCount(0, p.Amount).Dividing(effects.Divide(p.Amount))
	})
