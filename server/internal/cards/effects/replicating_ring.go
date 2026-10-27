package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Replicating Ring — Snow Artifact {3} (EDHREC rank 1605):
//
//	"{T}: Add one mana of any color.
//	 At the beginning of your upkeep, put a night counter on this
//	 artifact. Then if it has eight or more night counters on it,
//	 remove all of them and create eight colorless snow artifact
//	 tokens named Replicated Ring with "{T}: Add one mana of any
//	 color.""
//
// A Manalith that pays out after eight upkeeps. The mana is the
// five-way pipe at its printed width — "any color", so not narrowed
// to the commander's identity — and the tokens carry the same
// ability on their template, the way Treasure does. The upkeep
// trigger is one ability: the counter goes on as it resolves, and
// the check that follows reads the total AFTER that placement, so
// the eighth upkeep is the one that pays out; a Ring that has been
// proliferated pays out sooner, as printed. The tokens are snow
// artifacts by type line, so anything reading the supertype sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ff00f5b-4bf9-4724-8bb5-6b9a9eb0ec7f",
		Name:         "Replicating Ring",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Replicating Ring — put a night counter on it; at eight, replicate", func(g *game.Game, item *game.StackItem) error {
				ring := item.SourceCardID
				if z := g.FindCardZoneForEffect(ring); z == nil || z.Kind != game.ZoneBattlefield || sourceIsNewObject(g, item) { // #1432
					return nil
				}
				// #1290: the payout check reads the count AFTER the
				// counter LANDS, via AddCounterByKeyedForEffect's
				// continuation, not on the next line — a Doubling
				// Season / Hardened Scales board pauses the placement
				// on a CR 616 prompt, and reading before it resumes
				// would see the pre-placement count.
				return g.AddCounterByKeyedForEffect(uuid.Nil, ring, "night", 1, NewContext(g, item).ItemTail(replicatingRingCheck))
			}),
		},
	})
}

// replicatingRingCheck is "if it has eight or more night counters on it,
// remove all of them and create eight tokens", read once the counter has
// landed. An on-disk key: never renamed, never reused.
var replicatingRingCheck = CardTail("replicating-ring/check", func(ctx *Context, _ game.TailOutcome) error {
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
	if !ok || c.Counters["night"] < 8 {
		return nil
	}
	// The tokens wait for a prompt the removal asks (#4162, CR 608.2c).
	return replicatingRingReplicate.Run(ctx)
})

// replicatingRingReplicate is "If there are eight or more night counters
// on it, remove them and create eight colorless artifact tokens named
// Replicated Ring …", once the count has been read.
var replicatingRingReplicate = RegisterThenSteps("then-steps/replicating-ring-replicate",
	Instruction(func(ctx *Context) error {
		ring := ctx.Item.SourceCardID
		c, ok := ctx.Game.LookupCardForEffect(ring)
		if !ok {
			return nil
		}
		return AddCounter{Target: ring, Kind: "night", N: -c.Counters["night"]}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		return CreateToken{Template: b14ReplicatedRingToken(), N: 8}.Apply(ctx)
	}),
)
