package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banon, the Returners' Leader — Legendary Creature — Human Cleric Rebel
// {R}{W}, 1/3:
//
//	"Pray — Once during each of your turns, you may cast a creature
//	 spell from among cards in your graveyard that were put there from
//	 anywhere other than the battlefield this turn.
//	 Whenever you attack, you may pay {1} and discard a card. If you do,
//	 draw a card."
//
// The cast is a STANDING graveyard permission (ADR 0066) derived off the
// battlefield, spent once per turn through Muldrotha's per-type budget
// (CastPermission.PerType holding "creature", the Gisa and Geralf shape),
// and gated to your own turn by TimingYourTurnOnly, which leaves the
// creature's own timing in force. What makes it Banon's is the filter:
// PermissionFilter.PutInGraveyardThisTurn reads the arrival record
// MoveCard stamps on every card that lands in a graveyard
// (Card.PutInGraveyardFrom / PutInGraveyardThisTurn, ADR 0066's
// 2026-10-15 amendment, #3769). A discarded, milled or countered
// creature card qualifies; one that died does not; one that left the
// graveyard and came back is a new object judged by its latest arrival.
//
// "Whenever you attack" is ONE trigger per attack declaration
// (OncePerBatch, Adeline's shape). The loot is a MayPay of {1} whose
// payment leads straight into a discard prompt with a floor of one, so
// the two halves of "pay {1} and discard a card" are both owed once the
// player says yes, and the draw is the discard's continuation: it runs
// only for a player who really discarded ("if you do"). With an empty
// hand the cost cannot be paid in full, so nothing is offered rather
// than taking the mana for nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b62e88f-d544-4c01-a5a2-1c185a3c9c1e",
		Name:         "Banon, the Returners' Leader",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter:   game.PermissionFilter{CreatureOnly: true, PutInGraveyardThisTurn: true},
			PerType:  []string{game.PermissionTypeCreature},
			Timing:   game.TimingYourTurnOnly,
			Label:    "Pray — once during each of your turns, cast a creature spell from your graveyard that was put there from anywhere other than the battlefield this turn (Banon)",
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Banon — pay {1} and discard a card to draw a card", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if b14HandSize(g, item.Controller) == 0 {
					return nil
				}
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Banon — pay {1} and discard a card to draw a card?",
					Then:     banonPaid,
				}.Apply(ctx)
			})),
		},
	})
}

// banonDraw is "then draw a card" once the discard has landed, keyed
// (#4264): a table waiting on the discard is a restore point.
var banonDraw = game.RegisterDiscardThen("choose-cards/banon-the-returners-leader-draw", func(g *game.Game, r game.DiscardsDone) error {
	if !r.Discarded.Discarded(r.Player) {
		return nil
	}
	return g.DrawNForEffect(r.Player, 1)
})

// banonPaid is the paid half of Banon's attack trigger, keyed (#4264): the
// discard it opens is keyed too (banonDraw), so a table waiting on either
// the payer or the discard is a restore point.
var banonPaid = PayUnlessBranches("pay-unless/banon-the-returners-leader-discard",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   ctx.Controller(),
			Source:   ctx.Source(),
			N:        1,
			Question: "Banon — discard a card, then draw a card",
		}, banonDraw, nil)
	}, nil)
