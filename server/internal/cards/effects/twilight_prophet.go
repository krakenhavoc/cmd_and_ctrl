package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twilight Prophet — Creature — Vampire Cleric {2}{B}{B}, 2/4 (EDHREC
// rank 1138):
//
//	"Flying
//	 Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 At the beginning of your upkeep, if you have the city's blessing,
//	 reveal the top card of your library and put it into your hand. Each
//	 opponent loses X life and you gain X life, where X is that card's
//	 mana value."
//
// Darkstar Augur's flip with the price turned into a weapon, behind the
// blessing. The "if" is INTERVENING (CR 603.4): checked as the upkeep
// begins, so a controller without the blessing gets no trigger at all,
// and again as the trigger resolves.
//
// Reveal, then move, then drain, in the order Dark Confidant and the
// Augur use and for the same reasons (darkstar_augur.go): revealing is
// not drawing, the mana value is read before the card moves, and the
// life changes wait for the card to land in the hand through the
// library-to-hand continuation. A land, or a card with no mana cost, is
// X = 0 and changes nobody's life. "You gain X" is ONE gain whatever the
// number of opponents; each opponent loses X separately.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3afd8b9-d499-40f7-be41-f8fee6636a1d",
		Name:            "Twilight Prophet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, AllOf(ByYou, YouHaveTheCitysBlessingNow),
				twilightProphetLabel, twilightProphetFlip),
		},
	})
}

const twilightProphetLabel = "Twilight Prophet — reveal the top card of your library and put it into your hand; each opponent loses X life and you gain X life"

func twilightProphetFlip(g *game.Game, item *game.StackItem) error {
	if !YouHaveTheCitysBlessing(g, item.Controller) {
		return nil // CR 603.4: the intervening "if" is checked again
	}
	ctx := NewContext(g, item)
	player := item.Controller
	revealed := g.RevealTopOfLibraryForEffect(player, ctx.Source(), 1, twilightProphetLabel)
	if len(revealed) == 0 {
		return nil
	}
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		All:    true,
		Label:  twilightProphetLabel,
		Key:    twilightProphetTake,
	}.Apply(ctx)
}

// twilightProphetRecord is X, the revealed card's mana value.
const twilightProphetRecord = "twilight-prophet/x"

// twilightProphetDrain is "Each opponent loses X life and you gain X
// life".
var twilightProphetDrain = RegisterThenSteps("then-steps/twilight-prophet-drain",
	Instruction(func(ctx *Context) error {
		x := ctx.StepRecord(twilightProphetRecord).N()
		for _, p := range ctx.Game.Seats {
			if p == nil || p.Eliminated || p.ID == ctx.Controller() {
				continue
			}
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p.ID, -x); err != nil {
				return err
			}
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), ctx.StepRecord(twilightProphetRecord).N())
	}),
)

// twilightProphetTake is the flip's take, keyed (#4253): the card goes to
// hand, then X is its mana value, read off the card the take looked at.
var twilightProphetTake = RegisterTake("choose-cards/twilight-prophet-take", TakeDef{Then: twilightProphetPay})

func twilightProphetPay(g *game.Game, res TakeFromLibraryResult) error {
	x := 0
	if len(res.Looked) > 0 {
		if c, ok := g.LookupCardForEffect(res.Looked[0]); ok {
			x = c.ManaValue()
		}
	}
	if x == 0 {
		return nil
	}
	// The gain waits for a prompt the losses ask (a CR 616
	// order; #4162, CR 608.2c).
	ctx := keyedContext(g, res.Player, res.Source)
	ctx.RecordStep(twilightProphetRecord, nil, x)
	return twilightProphetDrain.Run(ctx)
}
