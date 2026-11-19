package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel's Tomb — Artifact {3} (Innistrad Remastered, #5948):
//
//	"Whenever a creature you control enters, you may have this artifact become a
//	 3/3 white Angel artifact creature with flying until end of turn."
//
// The trigger is Dancing Sword's: a "you may" on a creature of yours entering
// (another one, since the Tomb itself is no creature while the trigger watches),
// and the change is one layer-aware effect on the Tomb for the turn: it gains
// the creature type and the Angel subtype, becomes white, flying, with base
// power and toughness 3/3, and is still an artifact (CR 205.1b). A Tomb that has
// left the battlefield by then does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de3ed52e-7cc2-49ea-96d9-4254aac46168",
		Name:         "Angel's Tomb",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				WheneverAnotherCreatureEntersUnderYourControl(
					"Angel's Tomb — you may have this artifact become a 3/3 white Angel artifact creature with flying until end of turn",
					angelsTombBecomeAnAngel),
				"Angel's Tomb — have this artifact become a 3/3 white Angel artifact creature with flying until end of turn?"),
		},
	})
}

func angelsTombBecomeAnAngel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if ctx.isNewSourceObjectAsThis(ctx.Source()) {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	mods := []game.Mod{
		game.AddTypesMod("Creature"),
		game.AddSubtypesMod(ctx.TextSubtype("Angel")),
		game.SetColorsMod(ctx.TextColor("W")),
		game.AddKeywordsMod("flying"),
	}
	mods = append(mods, game.SetBasePTMods(3, 3)...)
	return ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     mods,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Angel's Tomb — becomes a 3/3 white Angel artifact creature with flying until end of turn",
	}.Apply(ctx)
}
