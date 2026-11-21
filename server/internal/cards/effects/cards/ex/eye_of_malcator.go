package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eye of Malcator — Artifact {2}{U} (Phyrexia: All Will Be One):
//
//	"When this artifact enters, scry 2.
//	 Whenever another artifact you control enters, this artifact becomes a
//	 4/4 Phyrexian Eye artifact creature until end of turn."
//
// The animation adds the Creature type and the Phyrexian and Eye subtypes
// (layer 4) and sets base power and toughness 4/4 (layer 7b) until end of
// turn (CR 611.2a). It stays an artifact. Only while it is still the same
// object on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7270b5b6-a571-4e6d-91de-03f7cc035b3c",
		Name:         "Eye of Malcator",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eye of Malcator — scry 2", effects.Do(effects.Scry{N: 2})),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
			}, "Eye of Malcator — becomes a 4/4 Phyrexian Eye artifact creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if ctx.IsNewSourceObject(ctx.Source()) || !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					mods := append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Phyrexian", "Eye")...),
					}, game.SetBasePTMods(4, 4)...)
					return effects.ScopedEffectFor{
						Target:   ctx.Source(),
						Mods:     mods,
						Duration: g.UntilEndOfTurnDuration(),
						Label:    "Eye of Malcator — a 4/4 Phyrexian Eye artifact creature until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
