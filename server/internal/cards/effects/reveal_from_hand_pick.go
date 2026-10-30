package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_from_hand_pick.go — "Target player reveals N cards from their
// hand. You choose one of those cards. That player <exiles | discards>
// it." (Taster of Wares, Klaw, Sonic Subjugator; #4588).
//
// Unlike ChooseFromRevealedHand (ADR 0116), which reveals the WHOLE hand,
// the revealing player picks which N cards to show (a choose_cards prompt
// over their own hand; with N at least their hand size they reveal it all
// and nobody is asked) and only those are revealed, to every player (CR
// 701.20a). The effect's controller then picks one of the revealed cards
// (the reveal_pick prompt, which shows the chooser only the revealed cards
// of a pool that is not theirs), and onPick moves it. Both prompts carry
// the asking item as data and a registered key, so a table waiting on
// either is a restore point.

// RevealFromHandPick is one card's pair of keyed prompts. Build it once,
// as a package-level var in the card's file; the two keys are on-disk
// identities, ledgered in testdata/effect_keys.txt: never renamed, never
// reused.
type RevealFromHandPick struct {
	name, verb string
	reveal     game.ChooseCardsThen
	pick       game.ChooseCardsThen
}

// NewRevealFromHandPick registers the pair. name is the card's name, for
// the prompts' headers; verb is what the revealing player does with the
// chosen card as the header says it ("exiles", "discards"); revealKey and
// pickKey are the two "choose-cards/<card>-<what>" keys. onPick moves the
// card the effect's controller chose and runs against the asking item;
// owner is the player whose hand it was revealed from.
func NewRevealFromHandPick(name, verb, revealKey, pickKey string, onPick func(ctx *Context, owner, card uuid.UUID) error) *RevealFromHandPick {
	r := &RevealFromHandPick{name: name, verb: verb}
	r.pick = askedPick(pickKey, func(ctx *Context, c game.CardsChosen) error {
		if len(c.Picked) == 0 {
			return nil
		}
		return onPick(ctx, c.From, c.Picked[0])
	})
	r.reveal = askedPick(revealKey, func(ctx *Context, c game.CardsChosen) error {
		return r.revealAndPick(ctx, c.Chooser, c.Picked)
	})
	return r
}

// Ask is the whole instruction: opp reveals n cards from their hand, the
// controller chooses one, onPick moves it. n is the number as the effect
// resolves. An empty hand, or n of 0 or less, reveals nothing and asks
// nothing.
func (r *RevealFromHandPick) Ask(ctx *Context, opp uuid.UUID, n int) error {
	g := ctx.Game
	p := g.PlayerByIDForEffect(opp)
	if p == nil || p.Hand == nil {
		return nil
	}
	hand := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	if n <= 0 || len(hand) == 0 {
		return nil
	}
	if n >= len(hand) {
		return r.revealAndPick(ctx, opp, hand)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: opp,
		Source:     ctx.Source(),
		Question:   r.name + " — reveal cards from your hand",
		Cards:      hand,
		Min:        n,
		Max:        n,
		LiveZone:   game.ZoneHand,
		ThenKey:    r.reveal,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// revealAndPick reveals the cards and asks the effect's controller to
// choose one.
func (r *RevealFromHandPick) revealAndPick(ctx *Context, opp uuid.UUID, revealed []uuid.UUID) error {
	if len(revealed) == 0 {
		return nil
	}
	g := ctx.Game
	g.RevealForEffect(game.RevealSpec{
		Player: opp,
		Source: ctx.Source(),
		Reason: r.name + " — reveal cards from your hand",
		Cards:  revealed,
	})
	_, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
		Chooser:  ctx.Controller(),
		Owner:    opp,
		Source:   ctx.Source(),
		Question: r.name + " — choose one of those cards; that player " + r.verb + " it",
		Cards:    revealed,
		Min:      1,
		Max:      1,
		ThenKey:  r.pick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return err
}
