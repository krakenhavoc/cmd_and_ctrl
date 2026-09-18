package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Genesis Wave — Sorcery {X}{G}{G}{G}:
//
//	"Reveal the top X cards of your library. You may put any number of
//	 permanent cards with mana value X or less from among them onto the
//	 battlefield. Then put all cards revealed this way that weren't put
//	 onto the battlefield into your graveyard."
//
// Built on #745's PutFromLibraryOntoBattlefield: the X cards are
// revealed to the table, the caster picks any number of the permanent
// cards with mana value X or less in one choose_cards prompt, and the
// picks enter as ONE simultaneous entry — every entry replacement is
// evaluated against the board before any of them arrived, so a check
// land does not see the basic that came down beside it. The rest go to
// the graveyard as a plain move: this is not a mill (CR 701.17a), so
// no mill payoff sees it.
//
// Mana value is read from the printed cost. A card whose cost the
// engine cannot parse (a split or adventure card's joined cost) is
// never offered, which can only leave a legal pick out, never let an
// illegal one in. A monocoloured hybrid {2/W} counts as 2 (CR 202.3f)
// once ManaValueLE reads the engine's mana value (PR #774).
//
// "Any number" includes none, so declining is an answer; the revealed
// cards then all go to the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2487868-f386-438e-a73f-b494f6d35fac",
		Name:         "Genesis Wave",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), x,
				"Genesis Wave — revealed from the top of the library")
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    revealed,
				Match:    ManaValueLE(x),
				Optional: true,
				Label:    "Genesis Wave — put any number of permanent cards with mana value X or less onto the battlefield",
				Then:     PutRestIntoGraveyard,
			}.Apply(ctx)
		},
	})
}
