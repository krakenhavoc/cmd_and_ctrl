package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sludge Monster — 5/5 Creature — Horror for {3}{U}{U} (EDHREC rank
// 3984):
//
//	"Whenever this creature enters or attacks, put a slime counter on
//	 up to one other target creature.
//	 Non-Horror creatures with slime counters on them lose all
//	 abilities and have base power and toughness 2/2."
//
// Repeatable, permanent removal that is not removal: the creature
// stays on the battlefield as a 2/2 vanilla, which beats destruction
// against indestructible, against recursion, and against anything
// whose value is its text box rather than its body. Point it at a
// commander and the commander stops doing anything at all.
//
// It is in the batch as the catalog's first ABILITY-REMOVING static
// that is not an attachment. Every other one rides an Aura or
// Equipment and applies to what the source is attached to; this one
// applies to a counter-marked set that can include permanents the
// Sludge Monster has never touched — a creature that picked up a
// slime counter from Toxrill, whose counters this reads too, because
// the card says "with slime counters on them" and names no source.
//
// # The two statics, and why the order works
//
// Losing all abilities is Layer 6; the base 2/2 is Layer 7b. The
// "Non-Horror" test is read post-layer, which means Layer 4 has
// already settled the types by the time either of these applies — so
// a changeling that a Maskwood Nexus made a Horror is exempt, and one
// that stops being a Horror becomes eligible.
//
// Ability removal is declared, not done by hand: emptying the keyword
// slice would take the KEYWORDS and leave every catalogued activated,
// triggered, mana, static and replacement ability working, because
// those are read through the catalog hooks at use time. The declared
// form is what tells the engine to suppress them.
//
// "Base power and toughness 2/2" is a SET at 7b, not a shrink, so
// +1/+1 counters (7d) and anthems (7c) still apply on top: a creature
// with three +1/+1 counters and a slime counter is a 5/5 with no
// abilities, exactly as printed.
//
// The Sludge Monster itself is a Horror, so it is immune to its own
// static even if something slimes it.
//
// # The trigger
//
// "Enters OR attacks" is one ability watching two events, so a
// Sludge Monster that enters and then attacks the same turn fires
// twice. "Up to one OTHER target creature" may name nothing, which is
// the right answer on an empty board and a legal one otherwise.
//
// "Other" is excluded by NAME (b03NotNamed), the catalog's standing
// posture, because a target clause is never handed the source. In a
// singleton format that is the same creature; it additionally
// excludes a token copy of the Sludge Monster, which is weaker than
// printed and never stronger — and in this particular case costs
// nothing at all, since a second Sludge Monster is a Horror and
// immune to the static anyway.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2802fc92-49a0-43c8-bc10-24882eeee3f1",
		Name:         "Sludge Monster",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:            game.Layer6Ability,
				RemovesAbilities: true,
				AppliesTo:        b38SlimedNonHorror,
				Apply:            func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
			},
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: b38SlimedNonHorror,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 2, 2
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrAttacks("Sludge Monster — put a slime counter on up to one other target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return AddCounter{Target: t.ID, Kind: "slime", N: 1}.Apply(ctx)
					}
					return nil
				}),
				TargetCreature("up to one other target creature", b03NotNamed("Sludge Monster")).WithCount(0, 1)),
		},
	})
}
