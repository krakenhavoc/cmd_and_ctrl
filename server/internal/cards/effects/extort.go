package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// extort.go — Extort (CR 702.101), first printed on both faces of Sorin
// of House Markov (#1117):
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do,
//	 each opponent loses 1 life and you gain that much life.)"
//
// Extort is a triggered ability (CR 702.101a), so it is a constructor
// on Spec.Triggered rather than a PrintedKeywords token, the way
// Cascade() and Storm() are: the keyword's rule lives in the trigger,
// and a bare string on the ability list would be a badge nothing
// reads. The trigger carries its name in TriggeredAbility.Keyword so
// cards/coverage can see it (#1258).
//
// Three things the printed text decides:
//
//   - "Whenever YOU cast a spell" is every spell its controller casts,
//     any type, from anywhere. The extort permanent must be on the
//     battlefield as the spell is cast, so a Sorin does not extort off
//     its own cast.
//   - The {W/B} is paid DURING RESOLUTION (MayPay), and a hybrid
//     symbol is paid with either colour. Declining is always an answer.
//   - "That much life" is the life actually lost, summed over the
//     opponents — b21DrainEachOpponentAndGainTheTotal reads it off the
//     drain's continuation, so an opponent whose loss was replaced
//     gives nothing (#793).
//
// Several instances each trigger and each ask separately (CR 702.101b).

// KeywordExtort is extort's trigger keyword.
const KeywordExtort = "extort"

// Extort is the whole keyword for the card named `name`: one trigger,
// declared once per printed instance.
func Extort(name string) game.TriggeredAbility {
	t := WheneverYouCast(nil, name+" — extort", extortPayToDrain(name))
	t.Keyword = KeywordExtort
	return t
}

// extortPayToDrain offers the {W/B} and, when it is paid, drains each
// opponent for 1 and gains the total.
func extortPayToDrain(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayPay{
			Chooser:  item.Controller,
			Cost:     "{W/B}",
			Question: name + " — extort: pay {W/B} to have each opponent lose 1 life and gain that much?",
			OnPay: func(ctx *Context) error {
				return b21DrainEachOpponentAndGainTheTotal(ctx.Game, ctx.Item, 1)
			},
		}.Apply(NewContext(g, item))
	}
}
