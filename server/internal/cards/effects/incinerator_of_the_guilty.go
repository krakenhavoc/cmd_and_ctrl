package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incinerator of the Guilty — Creature — Dragon {4}{R}{R}, 6/6 (Murders at
// Karlov Manor, #5223):
//
//	"Flying, trample
//	 Whenever this creature deals combat damage to a player, you may
//	 collect evidence X. When you do, this creature deals X damage to each
//	 creature and planeswalker that player controls. (To collect evidence
//	 X, exile cards with total mana value X or greater from your
//	 graveyard.)"
//
// Collect evidence X is the effect form with X chosen as the trigger
// resolves (RegisterCollectEvidenceX, CR 107.3f): the controller says
// whether to collect, picks any number of graveyard cards and then X, up to
// their total mana value. X = 0 with no card exiled is a collection too
// (the card's ruling), so "whenever you collect evidence" sees it. "When you
// do" is a reflexive trigger (CR 603.12) carrying X and the player who was
// dealt the damage; it has no target, and the damage to every creature and
// planeswalker that player controls as it resolves is dealt at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b7d037f-3269-4e3b-9744-a977acb69c67",
		Name:            "Incinerator of the Guilty",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			// A one-sided damage sweep of X (ADR 0126): the player dealt
			// combat damage is an opponent, and X is chosen as it resolves.
			TriggerWithPurpose(
				WheneverThisDealsCombatDamageToAPlayer("Incinerator of the Guilty — you may collect evidence X; when you do, it deals X damage to each creature and planeswalker that player controls",
					func(g *game.Game, item *game.StackItem) error {
						return incineratorEvidence.Ask(NewContext(g, item), "Incinerator of the Guilty — collect evidence X (exile cards from your graveyard, then choose X up to their total mana value) to deal X damage to each creature and planeswalker that player controls?")
					}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true}}),
		},
	})
}

// incineratorEvidence is "you may collect evidence X. When you do, …". Its
// keys are on-disk identities: never renamed, never reused.
var incineratorEvidence = RegisterCollectEvidenceX("incinerator-of-the-guilty", "Incinerator of the Guilty",
	func(ctx *Context, x int, _ []uuid.UUID) error {
		return ReflexiveTrigger{
			Label:   "Incinerator of the Guilty — deal " + strconv.Itoa(x) + " damage to each creature and planeswalker that player controls",
			Players: []uuid.UUID{ctx.Trigger().Event.Target},
			Body:    incineratorBlastBody,
			Params:  game.EffectParams{Amount: x},
		}.Apply(ctx)
	})

// incineratorBlastBody is "this creature deals X damage to each creature
// and planeswalker that player controls": X is Params.Amount, fixed when
// evidence was collected, and the player rides the payload. An on-disk key.
var incineratorBlastBody = game.ReflexiveBody("incinerator-of-the-guilty/damage",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		for _, ref := range ctx.Payload() {
			if ref.Kind == game.TargetPlayer {
				return damageEachMatching(ctx, And(Or(Creature(), Planeswalker()), ControlledBy(ref.ID)), p.Amount)
			}
		}
		return nil
	}, nil)
