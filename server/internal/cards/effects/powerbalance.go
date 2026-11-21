package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Powerbalance — Enchantment {R}{R} (Modern Horizons 3):
//
//	"Whenever an opponent casts a spell, you may reveal the top card of your
//	 library. If you do, you may cast that card without paying its mana cost
//	 if the two spells have the same mana value."
//
// A cast trigger on every opponent's spell (it resolves above that spell, so
// the spell is still on the stack and its mana value, {X} counted, is read
// then: CR 202.3e) asking the free yes/no for the reveal (MayChoice, the
// spell's mana value riding as the answer's Word). Revealing is public; the
// card is cast only when its mana value is the same, offered as a free cast
// from the library as the ability resolves (Aetherworks Marvel's offer, CR
// 608.2g), and a card that is not cast stays on top of the library. A land is
// revealed and never cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8be368dc-6212-4627-989e-077eea78f30d",
		Name:         "Powerbalance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnOpponentCast(nil),
				"Powerbalance — you may reveal the top card of your library; if the two spells have the same mana value, you may cast it without paying its mana cost",
				powerbalanceAsk),
		},
	})
}

func powerbalanceAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell, ok := g.LookupCardForEffect(ctx.Trigger().Event.CardID)
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(spell)
	return MayChoice{
		Question: "Powerbalance — reveal the top card of your library?",
		Then:     powerbalanceRevealed,
		Word:     strconv.Itoa(mv),
	}.Apply(ctx)
}

// powerbalanceRevealed is the yes branch: reveal, compare, offer the free
// cast. An on-disk key.
var powerbalanceRevealed = MayYes("confirm/gap-b56-powerbalance-reveal", func(ctx *Context, r game.ConfirmAnswer) error {
	g, you := ctx.Game, ctx.Controller()
	want := wordInt(r)
	revealed := g.RevealTopOfLibraryForEffect(you, ctx.Source(), 1, "Powerbalance — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if mv, ok := g.ManaValueForEffect(top); !ok || mv != want {
		return nil
	}
	perm := g.FreeCastFromLibraryPermission(you, ctx.Source(), "Powerbalance")
	perm.LapseOnPass = game.LapseStaysInHand
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ctx.Source(),
		Reason:       "Powerbalance",
		Cards:        revealed[:1],
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it on top",
	}) {
		return nil
	}
	g.GrantCastPermissionOverCardForEffect(revealed[0], perm)
	return nil
})
