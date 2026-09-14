package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rogue's Passage — Land (EDHREC rank 18):
//
//	"{T}: Add {C}.
//	 {4}, {T}: Target creature can't be blocked this turn."
//
// One of the most-played cards in the format, and it is played for
// exactly one reason: it is how a commander with twenty-one power
// worth of patience actually connects. The colourless mana is rent.
//
// It is here because it proves the half of the restriction
// vocabulary the Auras cannot. Pacifism's "can't attack or block"
// lives as long as the Aura is attached and is scoped by the
// attachment; this is the same bit with a DURATION — registered into
// the S32 until-end-of-turn registry, swept at cleanup (CR 514.2),
// with no permanent anywhere on the battlefield saying it is in
// effect. Both write the same field and the same two consumers read
// it, which is the whole argument for a restriction being a field
// rather than a keyword grant or a per-card hook.
//
// CR 611.2c: the creature is pinned at resolution, and the snapshot
// key carries its battlefield-entry stamp, so a creature that is
// flickered in response comes back a new object (CR 400.7) and is
// blockable again. That is correct and it is not special-cased here
// — it falls out of the shared eotSnapshot every until-end-of-turn
// primitive uses.
//
// The ability can target any creature, including an opponent's,
// which is the printed text and is occasionally the right play
// (unblocking a creature that is about to be chump-blocked to death
// by the player you want dead). It is instant-speed, so it can be
// activated after blockers would have been declared — at which point
// it does nothing, because restrictions are checked at DECLARATION
// (CR 509.1b) and the block already happened. That is the rule, not
// a gap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f29dc596-2121-4421-8463-15f6c2e8b9b3",
		Name:         "Rogue's Passage",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Target creature can't be blocked this turn",
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// CR 608.2b: a target that left in response is
				// skipped rather than errored — the ability does as
				// much as it can, which for one target is nothing.
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Rogue's Passage — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
