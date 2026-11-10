package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b72a_helpers.go — shared pieces of the Lorwyn Eclipsed transform
// legends built in standard batch 72 (Sygg, Grub; #4458).

// ThisEntersOrTransformsInto is "whenever this creature enters or
// transforms into <the face at index face>": one printed ability with two
// conditions (CR 701.27e, Huntmaster of the Fells). A permanent that enters
// on a face did not transform into it, so the entry half is the creature's
// own EventETB. Watch both game.EventETB and game.EventTransform.
func ThisEntersOrTransformsInto(face int) When {
	return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || ThisTransformedInto(face)(ev, source, lki, g)
	}
}

// PayToTransformAtYourFirstMain is the Lorwyn Eclipsed legends' "At the
// beginning of your first main phase, you may pay <cost>. If you do,
// transform <name>." on the face called `face`: a MayPay whose paid branch
// is Ashling's ashlingTransformPaid, so it does nothing if the permanent
// has turned over since the trigger went on the stack (CR 701.27f, the
// ruling).
func PayToTransformAtYourFirstMain(face, name, cost string) game.TriggeredAbility {
	return AtYourPrecombatMain(face+" — you may pay "+cost+" to transform "+name, func(g *game.Game, item *game.StackItem) error {
		return MayPay{
			Chooser:  item.Controller,
			Cost:     cost,
			Question: face + " — pay " + cost + " to transform " + name + "?",
			Then:     ashlingTransformPaid,
		}.Apply(NewContext(g, item))
	})
}
