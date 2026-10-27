package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// UntapUpToLands is "untap up to N lands" as printed: the resolving
// effect's controller chooses, as it resolves, up to N lands on the
// battlefield — ANY player's, since the clause says neither "target"
// nor "you control" — and those lands untap.
//
// It does not target (CR 115.1 — no "target"), so hexproof and shroud
// are irrelevant and nothing is chosen until resolution (CR 608.2).
// The pick is a choose_cards prompt over the battlefield (the shape
// Teferi Akosa's "tap any number of untapped creatures" uses), keyed
// (untapLandsChosen, #4253), so a table waiting on it is a restore point.
// A card with more printed after the untap registers its own keyed pick
// (Finale of Revelation's finaleUntapLands).
//
// Only TAPPED lands are offered. Choosing an untapped land is legal
// and does nothing, so leaving those out changes no outcome and
// keeps the prompt to the lands that matter. With no tapped land
// anywhere, nothing is asked.
//
// Every pick is re-read before it untaps: the prompt is asynchronous,
// and a land that left the battlefield between the question and the
// answer is skipped (the prompt's Zone re-check refuses it outright).
type UntapUpToLands struct {
	// N is the most lands that may be chosen.
	N int

	// Question is the prompt header, written the way the card is.
	Question string
}

func (u UntapUpToLands) Apply(ctx *Context) error {
	var tapped []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Tapped {
			tapped = append(tapped, c.InstanceID)
		}
	}
	if len(tapped) == 0 || u.N <= 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: u.Question,
		Cards:    tapped,
		Min:      0,
		Max:      u.N,
		ThenKey:  untapLandsChosen,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// untapLandsChosen untaps each land picked that is still on the battlefield:
// UntapUpToLands' keyed answer (#4253). The asking item rides the prompt as
// data.
var untapLandsChosen = askedPick("choose-cards/untap-up-to-lands", func(ctx *Context, r game.CardsChosen) error {
	next := ctx.asGroupMember()
	for _, id := range r.Picked {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (UntapTarget{Target: id}).Apply(next); err != nil {
			return err
		}
	}
	return nil
})
