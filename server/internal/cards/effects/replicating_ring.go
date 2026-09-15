package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Replicating Ring — put a night counter on it; at eight, replicate", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ring := item.SourceCardID
				if z := g.FindCardZoneForEffect(ring); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				if err := (AddCounter{Target: ring, Kind: "night", N: 1}).Apply(ctx); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(ring)
				if !ok || c.Counters["night"] < 8 {
					return nil
				}
				if err := (AddCounter{Target: ring, Kind: "night", N: -c.Counters["night"]}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: b14ReplicatedRingToken(), N: 8}.Apply(ctx)
			}),
		},
	})
}
