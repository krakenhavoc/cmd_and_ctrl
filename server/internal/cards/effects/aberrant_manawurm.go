package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aberrant Manawurm — Creature — Wurm {3}{G}, 2/5 (Secrets of
// Strixhaven, #4456):
//
//	"Trample
//	 Whenever you cast an instant or sorcery spell, this creature gets
//	 +X/+0 until end of turn, where X is the amount of mana spent to cast
//	 that spell."
//
// X is the mana actually spent, read off the spell's payment record as it
// is cast (Aetherflux Conduit's Build, as Manaform Hellkite reads it) and
// carried on the item, so a spell countered in response still pumps. A
// spell cast without paying its mana cost spent nothing: +0/+0. The
// pump is on the permanent that triggered: one that left and came back is
// a new object (CR 400.7).
//
// No simplification.
func init() {
	const label = "Aberrant Manawurm — +X/+0 until end of turn, X the mana spent to cast that spell"
	t := WheneverYouCast(Or(Instant(), Sorcery()), label, aberrantManawurmPump)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "a0657586-0404-4bb0-a92d-25f040d560c1",
		Name:            "Aberrant Manawurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{t},
	})
}

// aberrantManawurmPump gives the Manawurm +X/+0 until end of turn.
func aberrantManawurmPump(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return BoostUntilEOT{Target: item.SourceCardID, Power: max(item.Params.Amount, 0),
		Label: "Aberrant Manawurm — +X/+0 until end of turn"}.Apply(NewContext(g, item))
}
