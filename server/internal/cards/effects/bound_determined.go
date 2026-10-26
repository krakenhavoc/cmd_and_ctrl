package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bound // Determined — split card, oracle df39fbee-6ddc-4733-a32e-cdaf369ecccf:
//
//	Bound — Instant {3}{B}{G}: "Sacrifice a creature. Return up to X
//	        cards from your graveyard to your hand, where X is the
//	        number of colors that creature was. Exile this card."
//	Determined — Instant {G}{U}: "Other spells you control can't be
//	        countered this turn. Draw a card."
//
// Both halves register as ADR 0034 faces (ADR 0103): Bound under the
// bare oracle ID, Determined under "#1". Neither has fuse.
//
// BOUND. The sacrifice is the caster's choice at resolution (it is not
// a cost), and X is the number of colours the creature was as it last
// existed on the battlefield (CR 608.2h), read off its last-known
// information, so a colour an effect gave it counts. The return is a
// second choice, up to X cards from the caster's graveyard — the
// creature just sacrificed among them. With no creature to sacrifice,
// X is 0 and nothing returns.
//
// "Exile this card" is done FIRST rather than last. The order is not
// observable: while Bound resolves it is on the stack, not in the
// graveyard, so it can never be one of the cards it returns, and its
// exile triggers nothing the two choices could see. Doing it first is
// what keeps that true here: the engine routes a resolving spell to its
// graveyard as soon as the resolution hands off to a prompt, so a
// card exiled last would sit in the graveyard, a candidate for its own
// return, while the caster chose.
//
// DETERMINED. "Other spells you control can't be countered this turn"
// is a "this turn" grant (ADR 0106 §4 decision 2, #1806) read at the
// counter gate, covering spells already on the stack and spells cast
// later this turn (CR 611.2c), ending at cleanup (CR 514.2). "Other" is
// Determined itself, excepted by object (CounterShieldGrant.Except).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df39fbee-6ddc-4733-a32e-cdaf369ecccf",
		Name:         "Bound",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := ctx.Game.ExileCardForEffect(ctx.Source()); err != nil {
				return err
			}
			controller := ctx.Controller()
			return ctx.SacrificeThen(controller,
				sacrificeSpec("a creature", Creature()),
				"Bound — sacrifice a creature",
				1, boundSacrificed)
		},
	})
	Register(Spec{
		OracleID:     "df39fbee-6ddc-4733-a32e-cdaf369ecccf#1",
		Name:         "Determined",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GrantCounterShield{From: "Determined", ExceptThis: true, Grant: SpellsCantBeCounteredThisTurn(
				"Other spells you control can't be countered this turn.",
				game.CounterShieldYouControl, game.PermissionFilter{})}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}

// boundSacrificed is Bound's X once the creature has really been
// sacrificed, keyed (#4290): the number of colors it was as it last existed
// (CR 608.2h), then the second choice.
var boundSacrificed = RunThen("prompt-run/bound-sacrificed", func(ctx *Context, r game.RunDone) error {
	controller := ctx.Controller()
	x := 0
	for _, id := range r.Sacrifices().By(controller) {
		if info, ok := ctx.Game.LastKnownPermanentForEffect(id); ok {
			x = boundColorCount(info.Characteristic.Colors)
		}
	}
	return boundReturnUpToX(ctx.Game, ctx.Source(), controller, x)
})

// boundReturnUpToX is Bound's second choice: up to x cards from the
// caster's graveyard, picked now, return to their owner's hand. A card
// that left the graveyard before the answer is skipped.
func boundReturnUpToX(g *game.Game, source, controller uuid.UUID, x int) error {
	p := g.PlayerByIDForEffect(controller)
	if x <= 0 || p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) == 0 {
		return nil
	}
	cards := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		cards = append(cards, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Bound — return up to " + strconv.Itoa(x) + " cards from your graveyard to your hand",
		Cards:    cards,
		Min:      0,
		Max:      x,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  boundReturned,
	})
	return nil
}

// boundReturned is Bound's second choice answered, keyed (#4290): each
// picked card still in the graveyard returns to its owner's hand.
var boundReturned = game.RegisterChooseCardsThen("choose-cards/bound-return", func(g *game.Game, r game.CardsChosen) error {
	ctx := keyedContext(g, r.Chooser, r.Source)
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// boundColorCount is "the number of colors that creature was": the
// distinct colours among W, U, B, R and G it had. A colorless creature
// is zero.
func boundColorCount(colors []string) int {
	n := 0
	for _, want := range game.AllColors {
		for _, c := range colors {
			if c == want {
				n++
				break
			}
		}
	}
	return n
}
