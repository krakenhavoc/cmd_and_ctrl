package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Will of the Jeskai — Sorcery {3}{R} (Tarkir: Dragonstorm Commander):
//
//	"Choose one. If you control a commander as you cast this spell, you may
//	 choose both instead.
//	 • Each player may discard their hand and draw five cards.
//	 • Each instant and sorcery card in your graveyard gains flashback until
//	   end of turn. The flashback cost is equal to its mana cost."
//
// The mode count is Jeska's Will's (`OrUpToIf(2, YouControlACommander)`, read
// at announce and fixed, the rulings), and the bullets run in printed order
// (CR 608.2c). The first is the shared optional wheel (Ruin Grinder's: every
// player is asked before anything happens, in APNAP order, then the yes-sayers
// discard their hands and draw), here for five cards; a player may say yes with
// an empty hand. The second is Past in Flames' grant over the instant and
// sorcery cards in your graveyard as the bullet resolves, so with both bullets
// a card the wheel just discarded gains flashback too: the flashback cost is
// the card's mana cost, the card is exiled as it leaves the stack (CR
// 702.34a), and the grant ends with the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c3e6377-d09c-46b6-8a83-ada4a2930b14",
		Name:         "Will of the Jeskai",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Each player may discard their hand and draw five cards."),
			Mode("Each instant and sorcery card in your graveyard gains flashback until end of turn. The flashback cost is equal to its mana cost."),
		).OrUpToIf(2, YouControlACommander),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				return willOfTheJeskaiWheel.Start(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				return GrantCastFromYourGraveyard{
					Filter:            game.PermissionFilter{InstantOrSorceryOnly: true},
					AltCostKey:        "flashback",
					ExileOnResolution: true,
					Label:             "Flashback — its mana cost (Will of the Jeskai)",
				}.Apply(ctx)
			}),
		),
	})
}

// willOfTheJeskaiWheel is "each player may discard their hand and draw five
// cards". The key is on-disk: never renamed, never reused.
var willOfTheJeskaiWheel = newMayWheelDrawing("confirm/will-of-the-jeskai-may-wheel",
	"Will of the Jeskai — discard your hand and draw five cards?", 5,
	mayWheelStepsFor("then-steps/will-of-the-jeskai-may-wheel", 5))
