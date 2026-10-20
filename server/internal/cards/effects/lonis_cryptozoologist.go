package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lonis, Cryptozoologist — Legendary Creature — Snake Elf Scout {G}{U},
// 1/2:
//
//	"Whenever another nontoken creature you control enters, investigate.
//	 {T}, Sacrifice X Clues: Target opponent reveals the top X cards of
//	 their library. You may put a nonland permanent card with mana value
//	 X or less from among them onto the battlefield under your control.
//	 That player puts the rest on the bottom of their library in a
//	 random order."
//
// Two abilities.
//
// The investigate is Soul of the Harvest's trigger (another nontoken
// creature entering under your control) over the Clue token, the same
// one Tamiyo's Journal makes.
//
// The activated ability is Grim Hireling's "Sacrifice X Treasures" cost
// with Clues and {T} (SacrificeX, announced at activation, CR 602.2b),
// feeding Genesis Wave's reveal-then-put over an OPPONENT'S library:
//
//   - The reveal is the real one (RevealTopOfLibraryForEffect): every
//     seat is told which X cards were shown, because the clause says
//     "reveals", and the rest then go to the bottom, which clears what
//     the table knew about them — nobody learns the order afterwards.
//   - The put is PutFromLibraryOntoBattlefield, whose LibraryOwner is
//     the opponent (the pick is re-checked against THEIR library). The
//     permanent enters under the activator's control but is still its
//     owner's card (CR 108.3), so it dies to the opponent's graveyard.
//     Nonland and "mana value X or less" are the Match; a land, an
//     instant or sorcery, or a permanent that costs more is never
//     offered, and "you may" means the prompt can be answered with
//     nothing.
//   - "That player puts the rest on the bottom … in a random order" is
//     the bottom-in-random-order Then: every revealed card that was not
//     put onto the battlefield goes there, whether or not anything was
//     picked, and a library shorter than X reveals what it has.
//
// The ability targets the opponent, so it fizzles as a whole if they
// left the game (CR 608.2b), and the Clues are gone by then: the
// sacrifice is a cost. X of zero reveals nothing and does nothing.
//
// No simplification.
func init() {
	const label = "{T}, Sacrifice X Clues: Target opponent reveals the top X cards of their library. " +
		"You may put a nonland permanent card with mana value X or less from among them onto the battlefield " +
		"under your control. That player puts the rest on the bottom of their library in a random order."
	Register(Spec{
		OracleID:     "0ee06ed0-717a-4bfa-b634-d00e013c1f16",
		Name:         "Lonis, Cryptozoologist",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherNontokenCreatureYouControlEntered(ev, source, g)
			}, "Lonis, Cryptozoologist — investigate", Do(CreateToken{Template: ClueToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   label,
			Cost:    Plus(TapCost(), SacrificeX("X Clues", OfSubtype("Clue"))),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return lonisCryptozoologistDig(ctx, ctx.X())
			},
		}},
	})
}

// lonisCryptozoologistDig is the activated ability's body, given the X
// announced with the activation: reveal the
// top X of the target opponent's library, offer the controller the
// qualifying permanent cards, bottom the rest.
func lonisCryptozoologistDig(ctx *Context, x int) error {
	g := ctx.Game
	if x <= 0 {
		return nil
	}
	var victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	revealed := g.RevealTopOfLibraryForEffect(victim, ctx.Source(), x,
		"Lonis, Cryptozoologist — revealed from the top of the library")
	return PutFromLibraryOntoBattlefield{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Cards:        revealed,
		Match:        And(Nonland(), ManaValueLE(x)),
		Max:          1,
		Optional:     true,
		Label:        "Lonis, Cryptozoologist — you may put a nonland permanent card with mana value X or less onto the battlefield under your control",
		Then:         PutRestOnBottomInRandomOrder,
	}.Apply(ctx)
}
