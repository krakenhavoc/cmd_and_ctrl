package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyrewild Shaman — Creature — Goblin Shaman {2}{R}, 3/1 (Ravnica: Clue
// Edition, #5929):
//
//	"Bloodrush — {1}{R}, Discard this card: Target attacking creature gets
//	 +3/+1 until end of turn.
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, if this card is in your graveyard, you may pay {3}. If you do,
//	 return this card to your hand."
//
// Bloodrush is an activated ability that works from the hand (Armus, Skin of
// Evil's shape): the discard is part of the cost, and a target that stopped
// attacking or left is skipped (CR 608.2b). The second ability is a
// graveyard trigger (CR 113.6) that fires once per damaged player in a
// combat damage step (CR 603.2c). Its intervening "if" (CR 603.4) is checked
// again as it resolves and again after the payment: the card must still be
// that same card in the graveyard (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1e71473-3eba-4d29-87e2-7f14acab2ccd",
		Name:         "Pyrewild Shaman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InGraveyard(WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Pyrewild Shaman — you may pay {3}: return this card to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if !pyrewildShamanInGraveyard(g, item) {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{3}",
						Question: "Pyrewild Shaman — pay {3} to return it from your graveyard to your hand?",
						Then:     pyrewildShamanPaid,
					}.Apply(NewContext(g, item))
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Bloodrush — {1}{R}, Discard this card: Target attacking creature gets +3/+1 until end of turn.",
			Cost:    game.AbilityCost{Mana: "{1}{R}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: withAnswers(ForTargets(BenefitUntilEOT(0, 3, 1)), game.AnswerPump),
			Targets: TargetCreature("target attacking creature", AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return BoostUntilEOT{Target: id, Power: 3, Toughness: 1, Label: "Pyrewild Shaman — bloodrush +3/+1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}

// pyrewildShamanInGraveyard is the intervening "if this card is in your
// graveyard" (CR 603.4): the card is in a graveyard and is not a new object
// since the trigger was put on the stack.
func pyrewildShamanInGraveyard(g *game.Game, item *game.StackItem) bool {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	return z != nil && z.Kind == game.ZoneGraveyard && !sourceIsNewObject(g, item)
}

// pyrewildShamanPaid is the paid half, keyed.
var pyrewildShamanPaid = PayUnlessBranches("pay-unless/pyrewild-shaman-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if !pyrewildShamanInGraveyard(ctx.Game, ctx.Item) {
			return nil
		}
		return ReturnFromGraveyard{Target: ctx.Source(), Dest: game.ZoneHand}.Apply(ctx)
	}, nil)
