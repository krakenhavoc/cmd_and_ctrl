package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Admiral Brass, Unsinkable — Legendary Creature — Human Pirate {2}{U}{B}{R}, 3/3:
//
//	"When Admiral Brass enters, mill four cards.
//	 At the beginning of combat on your turn, you may return target Pirate
//	 creature card from your graveyard to the battlefield with a finality
//	 counter on it. It has base power and toughness 4/4. It gains haste
//	 until end of turn. (If a creature with a finality counter on it would
//	 die, exile it instead.)"
//
// The return enters the Pirate with the finality counter (CR 122.1; the
// engine owns the exile replacement, game/finality_counters.go). Base 4/4
// is a layer 7b effect that lasts as long as that object stays on the
// battlefield (CR 400.7), so counters and anthems still stack on it; haste
// lasts until end of turn. The "may" is asked as the trigger resolves, and
// a target that has left the graveyard by then is not returned (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "4df179b4-7c6b-4e63-9d7d-62f4df377989",
		Name:         "Admiral Brass, Unsinkable",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Admiral Brass — mill four cards", Do(MillCards{N: 4})),
			Optional(Targeting(
				AtBeginningOfYourCombat("Admiral Brass — return target Pirate creature card from your graveyard to the battlefield with a finality counter; it has base power and toughness 4/4 and gains haste until end of turn",
					cmdr2B05AdmiralBrassReturn),
				TargetCardInGraveyard("target Pirate creature card from your graveyard", YouOwn(), Creature(), IsPirateCard())),
				"Admiral Brass — return a Pirate creature card from your graveyard?"),
		},
	})
}

func cmdr2B05AdmiralBrassReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false,
		map[string]int{game.CounterFinality: 1}, ctx.ItemTail(cmdr2B05AdmiralBrassEntered))
}

var cmdr2B05AdmiralBrassEntered = EnteredTail("cmdr2-b05/admiral-brass-entered", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   entered,
		Mods:     game.SetBasePTMods(4, 4),
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Admiral Brass — base power and toughness 4/4",
	}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: entered, Keywords: []string{"haste"}, Label: "Admiral Brass — haste until end of turn"}.Apply(ctx)
})
