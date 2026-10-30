package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hawkeyeMasterMarksmanOracle is the card's catalog key.
const hawkeyeMasterMarksmanOracle = "6b59bb1f-f2d4-4d47-bd57-23927d588381"

// Hawkeye, Master Marksman — Legendary Creature — Human Archer Hero
// {1}{R}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Reach, first strike
//	 Trick Arrows — Whenever Hawkeye becomes tapped, you may pay {1} up to
//	 three times. When you do, choose up to that many —
//	 • Net — Target creature can't block this turn.
//	 • Explosive — Hawkeye deals 2 damage to target player.
//	 • Boomerang — Discard a card, then draw a card."
//
// The trigger is one registered ability. Its body is MayPayX with Symbol
// "{1}" and Max 3 (#4579): the controller picks how many times to pay as
// it resolves, from 0 to the least of three and the mana they can produce
// right now (CR 608.2d), and then pays that much through the ordinary
// "you may pay" prompt, so the pool, the auto-tapper and Always / Never
// all apply. Paying nothing is the same as declining.
//
// Paying creates the reflexive trigger (CR 603.12, "when you do"), which
// goes on the stack above the tap trigger. Paying several times makes ONE
// trigger, not one per payment (CR 603.12a: "if a resolving spell or
// ability includes a choice to pay a cost multiple times and creates a
// triggered ability that triggers when that payment is made, paying that
// cost one or more times causes the reflexive triggered ability to trigger
// only once"). The tap trigger fires for an attacker tapped by attacking
// (EventAttack) as well as for every other tap (EventTapCard), as
// ThisBecameTapped says. Its count of bullets is the
// number of times {1} was paid, fixed when it is created (carried as
// Params.Amount and re-derived by hawkeyeTrickArrowsModes, so a restore
// point holds it). The bullets are chosen as it goes on the stack (CR
// 603.3c), "up to": none removes it (CR 700.2b), and a bullet whose target
// has no legal choice is not offered. Each bullet picks its target then
// (CR 603.3d) and runs in printed order (CR 608.2c). A bullet can't be
// chosen twice (CR 700.2d), so three payments buy at most the three
// printed bullets.
//
// The explosive damage is dealt by Hawkeye as it last existed, so a
// Hawkeye that left in response still deals it. The boomerang draws even
// from an empty hand.
//
// No simplification.
func init() {
	const label = "Hawkeye, Master Marksman — pay {1} up to three times; when you do, choose up to that many"
	Register(Spec{
		OracleID:        hawkeyeMasterMarksmanOracle,
		Name:            "Hawkeye, Master Marksman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "first strike"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped, label, hawkeyeTrickArrows),
		},
	})
}

// hawkeyeTrickArrows is the tap trigger's body: pay {1} up to three times.
func hawkeyeTrickArrows(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Symbol: "{1}",
		Max:    3,
		Label:  "Hawkeye, Master Marksman",
		Buys:   "choose up to X modes",
		Goal:   XAsHighAsYouCan,
		Unit:   game.PayAmountOther,
		Then:   hawkeyeTrickArrowsPaid,
	}.Apply(NewContext(g, item))
}

// hawkeyeTrickArrowsPaid is the paid half, keyed: the number of payments
// rides the prompt.
var hawkeyeTrickArrowsPaid = PayUnlessBranches("pay-unless/hawkeye-trick-arrows",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Hawkeye, Master Marksman — choose up to " + strconv.Itoa(r.Number),
			Body:   hawkeyeTrickArrowsBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// hawkeyeTrickArrowsBody is the reflexive trigger: its bullets are its
// mode clause, its count the payments made.
var hawkeyeTrickArrowsBody = game.ReflexiveModalBody("hawkeye/trick-arrows", hawkeyeTrickArrowsModes)

func hawkeyeTrickArrowsModes(_ uuid.UUID, p game.EffectParams) *game.ModeSpec {
	return ChooseN("Choose up to "+strconv.Itoa(p.Amount), 0, p.Amount,
		ModeDoing("Net — Target creature can't block this turn.", TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock,
					Label: "Hawkeye, Master Marksman — Net: can't block this turn"}.Apply(ctx)
			}),
		ModeDoing("Explosive — Hawkeye deals 2 damage to target player.", TargetPlayer("target player"),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
			}),
		ModeDoing("Boomerang — Discard a card, then draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player: item.Controller,
					Source: item.SourceCardID,
					N:      1,
				}, theRuinousWreckingCrewDraw, nil)
			}),
	)
}
