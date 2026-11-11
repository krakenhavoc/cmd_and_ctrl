package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Transcendent Dragon — Creature — Dragon {4}{U}{U}, 4/3 (Tarkir:
// Dragonstorm Commander):
//
//	"Flash
//	 Flying
//	 When this creature enters, if you cast it, counter target spell. If that
//	 spell is countered this way, exile it instead of putting it into its
//	 owner's graveyard, then you may cast it without paying its mana cost."
//
// "If you cast it" is an intervening if (CR 603.4) read off the permanent's
// cast provenance, Null Summoner's: a reanimated or blinked Dragon enters
// quietly, and being cast is a fact that does not change while the trigger is
// on the stack. The counter is Force of Negation's (CounterTarget with an
// exile destination), so a spell that can't be countered stays on the stack
// and is neither exiled nor offered; a commander whose owner chose the
// command zone isn't in exile and isn't offered either. The free cast is
// part of the resolution (CR 608.2g, game.CastExiledFreeKeepingForEffect):
// {0} with X as 0, no alternative cost, the card's own timing ignored, and
// declining leaves it in exile.
//
// Seam: game.CastExiledFreeKeepingForEffect.
func init() {
	Register(Spec{
		OracleID:        "e83a629e-2d74-48e2-ad4d-f390067cc51a",
		Name:            "Transcendent Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, nullSummonerWasCast),
					"Transcendent Dragon — counter target spell; if countered this way, exile it, then you may cast it without paying its mana cost",
					transcendentDragonCounter),
				TargetSpell("target spell")),
		},
	})
}

func transcendentDragonCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	stackID := item.Targets[0].ID
	if g.StackItemForEffect(stackID) == nil {
		return nil
	}
	if err := (CounterTarget{StackID: stackID, Dest: game.ZoneRef{Kind: game.ZoneExile}}).Apply(ctx); err != nil {
		return err
	}
	// Only a spell that really reached exile was "countered this way".
	if z := g.FindCardZoneForEffect(stackID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return g.CastExiledFreeKeepingForEffect(item.Controller, item.SourceCardID, stackID)
}
