package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashcoat of the Shadow Swarm — Legendary Creature — Rat Warlock {3}{B},
// 3/4 (Bloomburrow Commander reprint, #5908):
//
//	"Whenever Ashcoat attacks or blocks, other Rats you control get +X/+X
//	 until end of turn, where X is the number of Rats you control.
//	 At the beginning of your end step, you may mill four cards. If you
//	 do, return up to two Rat creature cards from your graveyard to your
//	 hand."
//
// The attack-or-block trigger is Burning Sun Cavalry's. X counts every
// Rat you control, Ashcoat included, and the set pumped (the other Rats) is
// fixed as the trigger resolves (CR 611.2c). Rats are read as a type word
// (CR 612.2).
//
// The end-step mill is a yes/no asked as the trigger resolves and only when
// the library holds four cards (a player can't choose to mill more than
// their library holds); "if you do" is the yes. The offer is every Rat
// creature card in the graveyard as it is then, the just-milled ones
// included, with a floor of zero (Another Chance's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cc69ea5-42a2-4306-ac6d-cff5adb20bcb",
		Name:         "Ashcoat of the Shadow Swarm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
				},
				"Ashcoat of the Shadow Swarm — other Rats you control get +X/+X until end of turn, where X is the number of Rats you control",
				ashcoatPumpRats),
			AtYourEndStep("Ashcoat of the Shadow Swarm — you may mill four cards; if you do, return up to two Rat creature cards from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if p := g.PlayerByIDForEffect(ctx.Controller()); p == nil || p.Library == nil || p.Library.Size() < 4 {
						return nil
					}
					return MayChoice{
						Question: "Ashcoat of the Shadow Swarm — mill four cards?",
						Then:     ashcoatMill,
					}.Apply(ctx)
				}),
		},
	})
}

// ashcoatPumpRats gives the other Rats the controller controls +X/+X, X
// being the Rats they control as the ability resolves.
func ashcoatPumpRats(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := countYouControl(ctx, OfCreatureType("Rat"))
	return BoostUntilEOT{
		Match:     And(OfCreatureType("Rat"), YouControl(), OtherThan(item.SourceCardID)),
		Power:     x,
		Toughness: x,
		Label:     "Ashcoat of the Shadow Swarm — +X/+X",
	}.Apply(ctx)
}

// ashcoatMill is the answer to the mill question: on a yes, mill four, then
// return up to two Rat creature cards. An on-disk key.
var ashcoatMill = MayAnswer("confirm/ashcoat-mill-four", func(ctx *Context, r game.ConfirmAnswer) error {
	if !r.Accepted {
		return nil
	}
	if err := (MillCards{N: 4}).Apply(ctx); err != nil {
		return err
	}
	return ReturnChosenFromGraveyard{
		Player:   ctx.Controller(),
		Question: "Ashcoat of the Shadow Swarm — return up to two Rat creature cards from your graveyard to your hand",
		Match:    func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord("Rat") },
		Max:      2,
		ToHand:   true,
	}.Apply(ctx)
})
