package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whiplash, Vengeful Engineer — Legendary Creature — Human Artificer
// Villain {B}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Whiplash enters tapped.
//	 Whenever Whiplash attacks, if he's equipped, each opponent loses X
//	 life and you gain X life, where X is the number of Equipment
//	 attached to him."
//
// The tapped entry is the usual self-replacement. The attack trigger is an
// intervening "if" (CR 603.4): it is checked as the attack is declared and
// again as the trigger resolves, so an Equipment moved away in response
// makes it do nothing. X is counted as it resolves, from the Equipment
// attached to Whiplash then (CR 608.2h). Life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00ed7aca-b3c1-4d48-acd5-b1596daeb4c3",
		Name:         "Whiplash, Vengeful Engineer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "Whiplash is equipped",
				Test: func(ctx *Context) bool { return mshB11EquipmentAttachedToSource(ctx) > 0 },
			}, WheneverThisAttacks("Whiplash — each opponent loses X life and you gain X life, X the number of Equipment attached to him",
				mshB11WhiplashDrain)),
		},
	})
}

// mshB11EquipmentAttachedToSource is how many Equipment are attached to the
// item's source permanent, 0 when it has left the battlefield.
func mshB11EquipmentAttachedToSource(ctx *Context) int {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.HasSubtype("Equipment") && c.IsAttachedTo(ctx.Source()) {
			n++
		}
	}
	return n
}

// mshB11WhiplashDrain is "each opponent loses X life and you gain X life".
func mshB11WhiplashDrain(g *game.Game, item *game.StackItem) error {
	return b40DrainEachOpponent(mshB11EquipmentAttachedToSource(NewContext(g, item)))(g, item)
}
