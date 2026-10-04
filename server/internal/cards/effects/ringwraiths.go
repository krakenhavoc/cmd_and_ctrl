package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ringwraiths — Creature — Wraith Knight {4}{B}{B}, 5/5:
//
//	"When this creature enters, target creature an opponent controls
//	 gets -3/-3 until end of turn. If that creature is legendary, its
//	 controller loses 3 life.
//	 When the Ring tempts you, return this card from your graveyard to
//	 your hand."
//
// Whether the creature is legendary, and who controls it, are read as
// the ability resolves; with the target gone the ability does nothing
// and nobody loses life (2023-06-16 ruling). The second ability works
// only while the card is in your graveyard, on every temptation.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Ringwraiths — target creature an opponent controls gets -3/-3", ringwraithsShrink)
	enters.Targets = TargetCreature("target creature an opponent controls", OpponentControls())
	Register(Spec{
		OracleID:     "f09278dc-1e67-4cd8-977d-4b3b94430aac",
		Name:         "Ringwraiths",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			enters,
			InGraveyard(WheneverTheRingTemptsYou("Ringwraiths — return this card from your graveyard to your hand",
				returnThisCardFromYourGraveyardToHand)),
		},
	})
}

// ringwraithsShrink is the enters ability's body.
func ringwraithsShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: -3, Toughness: -3, Label: "Ringwraiths — -3/-3"}).Apply(ctx); err != nil {
			return err
		}
		if c.IsLegendary() {
			if err := g.ChangePlayerLifeForEffect(ctx.Source(), c.Controller, -3); err != nil {
				return err
			}
		}
	}
	return nil
}
