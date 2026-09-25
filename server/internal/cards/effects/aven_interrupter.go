package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Interrupter — {1}{W}{W} Creature — Bird Rogue 2/2:
//
//	"Flash
//	 Flying
//	 When this creature enters, exile target spell. It becomes
//	 plotted. (Its owner may cast it as a sorcery on a later turn
//	 without paying its mana cost.)
//	 Spells your opponents cast from graveyards or from exile cost {2}
//	 more to cast."
//
// Waited on #1318. Its enters trigger is "exile target spell", and the
// only way to exile a spell before that issue was the plain exile
// route, which moved the card and left its stack record behind — the
// same wedge Aang, Swift Savior's airbend of a spell hit. The engine
// now retires the record on every stack exit, and "exile target spell"
// is its own primitive (ExileTargetSpell → Game.ExileSpellThenForEffect),
// which does not counter the spell: a spell that can't be countered is
// exiled all the same.
//
// "It becomes plotted" (CR 702.170c/d) is the effect half of plot,
// built alongside (game/plot.go): a free cast permission over the card
// in exile, in its owner's main phase with the stack empty, on any
// later turn, which no flash grant widens. It is granted from the
// exile's continuation, so a commander spell whose owner takes the
// command zone instead (CR 903.9) is not plotted — it is not in exile.
//
// The tax is an ordinary CR 601.2f cost increase on a cast whose
// source zone is a graveyard or exile, so it also taxes the plotted
// card this creature just made, when an opponent casts it — which is
// the printed interaction, not an accident.
func init() {
	Register(Spec{
		OracleID:        "d31fd12f-b4dd-4bc3-ace4-703dabd0f607",
		Name:            "Aven Interrupter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetSpell("target spell"),
			Key:       "Aven Interrupter — exile target spell; it becomes plotted",
			Effect:    avenInterrupterExileAndPlot,
		}},
		CostModifiers: []game.CostModifier{
			CostsMore(2, "Spells your opponents cast from graveyards or from exile cost {2} more to cast.",
				OpponentsSpell(), CastFromGraveyardOrExile()),
		},
	})
}

// avenInterrupterExileAndPlot is the trigger body: the first still-legal
// target (CR 608.2b — a spell that resolved or was countered in
// response is gone, and the trigger does nothing), exiled, then plotted
// if it landed there.
func avenInterrupterExileAndPlot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		spell := t.ID
		return ExileTargetSpell{StackID: spell, Then: func(ctx *Context, exiled bool) error {
			if !exiled {
				return nil
			}
			return PlotExiled{Card: spell}.Apply(ctx)
		}}.Apply(ctx)
	}
	return nil
}
