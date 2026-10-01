package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commandeer — Instant {5}{U}{U}:
//
//	"You may exile two blue cards from your hand rather than pay this
//	 spell's mana cost.
//	 Gain control of target noncreature spell. You may choose new
//	 targets for it. (If that spell is an artifact, enchantment, or
//	 planeswalker, the permanent enters under your control.)"
//
// Aethersnatch's resolution over a noncreature clause, plus a pitch.
// The pitch is Force of Will's Pitch with no life and a hand spec of
// two: AlternativeCost.ExileFromHand reads its count off the spec's
// Min (#1745), so the caster names two blue cards, both are exiled as
// the cost is paid (CR 601.2h), and the client's picker and the bot's
// enumerator size themselves from the same number.
func init() {
	Register(Spec{
		OracleID:     "ad7d854b-d303-40eb-acd6-03a8023e05e7",
		Name:         "Commandeer",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Pitch("Exile two blue cards from your hand", 0,
				CardInYourHand("two blue cards from your hand", OfColor("U")).WithCount(2, 2),
				"two blue cards"),
		},
		Targets: TargetSpell("target noncreature spell", Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return GainControlOfSpell{Spell: t.ID, ChooseNewTargets: true,
				Label: "Commandeer — gain control of target noncreature spell"}.Apply(ctx)
		},
	})
}
